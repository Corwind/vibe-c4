package analyzer

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// GoAnalyzer implements the Analyzer interface for Go projects.
type GoAnalyzer struct{}

// NewGoAnalyzer creates a new GoAnalyzer instance.
func NewGoAnalyzer() *GoAnalyzer {
	return &GoAnalyzer{}
}

// AnalyzeProject parses a Go project at the given path and returns analysis results.
func (a *GoAnalyzer) AnalyzeProject(ctx context.Context, projectPath string) (*AnalysisResult, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolving project path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path does not exist: %s", absPath)
	}

	moduleInfo, err := a.parseGoMod(absPath)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}

	packages, err := a.scanPackages(ctx, absPath, moduleInfo.ModulePath)
	if err != nil {
		return nil, fmt.Errorf("scanning packages: %w", err)
	}

	importGraph := a.buildImportGraph(packages, moduleInfo.ModulePath)

	callGraph, externals, entrypoints := a.analyzeAllBodies(ctx, absPath, moduleInfo.ModulePath, packages)
	mainEntrypoints := a.detectMainEntrypoints(packages)
	entrypoints = append(entrypoints, mainEntrypoints...)
	interfaceImpls := matchInterfaceImpls(packages)

	return &AnalysisResult{
		Module:               *moduleInfo,
		Packages:             packages,
		ImportGraph:          importGraph,
		CallGraph:            callGraph,
		ExternalInteractions: externals,
		InterfaceImpls:       interfaceImpls,
		Entrypoints:          entrypoints,
	}, nil
}

func (a *GoAnalyzer) parseGoMod(projectPath string) (*ModuleInfo, error) {
	goModPath := filepath.Join(projectPath, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}

	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}

	info := &ModuleInfo{
		ModulePath: f.Module.Mod.Path,
	}

	if f.Go != nil {
		info.GoVersion = f.Go.Version
	}

	for _, req := range f.Require {
		dep := Dependency{
			Path:    req.Mod.Path,
			Version: req.Mod.Version,
		}
		if req.Indirect {
			info.IndirectDeps = append(info.IndirectDeps, dep)
		} else {
			info.DirectDeps = append(info.DirectDeps, dep)
		}
	}

	return info, nil
}

func (a *GoAnalyzer) scanPackages(ctx context.Context, projectPath string, modulePath string) ([]PackageInfo, error) {
	var packages []PackageInfo
	fset := token.NewFileSet()

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !info.IsDir() {
			return nil
		}

		// Skip hidden directories and vendor
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || base == "vendor" || base == "testdata" {
			return filepath.SkipDir
		}

		// Parse Go files with full AST (not imports-only) for type extraction
		pkgs, err := parser.ParseDir(fset, path, func(fi os.FileInfo) bool {
			return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
		}, parser.ParseComments)
		if err != nil {
			return nil // skip directories that can't be parsed
		}

		if len(pkgs) == 0 {
			return nil
		}

		for _, pkg := range pkgs {
			relDir, err := filepath.Rel(projectPath, path)
			if err != nil {
				continue
			}

			importPath := modulePath
			if relDir != "." {
				importPath = modulePath + "/" + filepath.ToSlash(relDir)
			}

			pkgInfo := PackageInfo{
				Name:       pkg.Name,
				ImportPath: importPath,
				Dir:        relDir,
				Role:       classifyPackageRole(relDir),
			}

			var allImports []string
			importSet := make(map[string]bool)

			for filename, file := range pkg.Files {
				pkgInfo.GoFiles = append(pkgInfo.GoFiles, filepath.Base(filename))
				for _, imp := range file.Imports {
					impPath := strings.Trim(imp.Path.Value, `"`)
					if !importSet[impPath] {
						importSet[impPath] = true
						allImports = append(allImports, impPath)
					}
				}

				relFile := filepath.Join(relDir, filepath.Base(filename))
				extractTypesFromFile(fset, file, relFile, &pkgInfo)
			}

			pkgInfo.Imports = allImports

			// Attach methods to their receiver structs
			attachMethodsToStructs(&pkgInfo)

			packages = append(packages, pkgInfo)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return packages, nil
}

func classifyPackageRole(relDir string) string {
	parts := strings.Split(filepath.ToSlash(relDir), "/")
	if len(parts) == 0 {
		return "root"
	}

	switch parts[0] {
	case "cmd":
		return "cmd"
	case "internal":
		return "internal"
	case "pkg":
		return "pkg"
	case "api":
		return "api"
	case ".":
		return "root"
	default:
		return "other"
	}
}

func (a *GoAnalyzer) analyzeAllBodies(ctx context.Context, projectPath string, modulePath string, packages []PackageInfo) ([]FunctionCall, []ExternalInteraction, []Entrypoint) {
	var allCalls []FunctionCall
	var allExternals []ExternalInteraction
	var allEntrypoints []Entrypoint

	fset := token.NewFileSet()

	for _, pkg := range packages {
		if ctx.Err() != nil {
			break
		}

		pkgDir := filepath.Join(projectPath, pkg.Dir)
		pkgs, err := parser.ParseDir(fset, pkgDir, func(fi os.FileInfo) bool {
			return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if err != nil {
			continue
		}

		for _, astPkg := range pkgs {
			for _, file := range astPkg.Files {
				importAliasMap := buildImportAliasMap(file)

				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok {
						continue
					}

					callerType := ""
					if fn.Recv != nil {
						callerType = receiverTypeName(fn.Recv)
					}

					calls, externals, entrypoints := analyzeBody(fset, fn, pkg.ImportPath, callerType, modulePath, importAliasMap)
					allCalls = append(allCalls, calls...)
					allExternals = append(allExternals, externals...)
					allEntrypoints = append(allEntrypoints, entrypoints...)
				}
			}
		}
	}

	return allCalls, allExternals, allEntrypoints
}

func (a *GoAnalyzer) detectMainEntrypoints(packages []PackageInfo) []Entrypoint {
	var entrypoints []Entrypoint
	for _, pkg := range packages {
		if pkg.Name != "main" {
			continue
		}
		for _, fn := range pkg.Functions {
			if fn.Name == "main" {
				entrypoints = append(entrypoints, Entrypoint{
					Kind:     "main",
					PkgPath:  pkg.ImportPath,
					FuncName: "main",
					FilePath: fn.FilePath,
					Line:     fn.Line,
				})
			}
		}
	}
	return entrypoints
}

func (a *GoAnalyzer) buildImportGraph(packages []PackageInfo, modulePath string) map[string][]string {
	graph := make(map[string][]string)
	for _, pkg := range packages {
		var internalImports []string
		for _, imp := range pkg.Imports {
			if strings.HasPrefix(imp, modulePath) {
				internalImports = append(internalImports, imp)
			}
		}
		if len(internalImports) > 0 {
			graph[pkg.ImportPath] = internalImports
		}
	}
	return graph
}

