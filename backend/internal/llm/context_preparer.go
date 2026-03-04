package llm

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

// ContextPreparer transforms raw AnalysisResult into a compact InterpretationInput
// suitable for LLM consumption.
type ContextPreparer struct{}

// Prepare compresses an AnalysisResult into an InterpretationInput by stripping
// file paths, line numbers, and deduplicating facts.
func (cp *ContextPreparer) Prepare(result *analyzer.AnalysisResult, projectPath string) *InterpretationInput {
	modulePath := result.Module.ModulePath

	input := &InterpretationInput{
		ProjectName: path.Base(modulePath),
		ModulePath:  modulePath,
	}

	input.PackageSummaries = cp.compressPackages(result.Packages, modulePath)
	input.ExternalInteractions = cp.deduplicateExternalInteractions(result.ExternalInteractions)
	input.Entrypoints = cp.mapEntrypoints(result.Entrypoints)
	input.CallGraphEdges = cp.filterCallGraph(result.CallGraph)
	input.InterfaceImpls = cp.mapInterfaceImpls(result.InterfaceImpls)
	input.DirectDeps = cp.collectDirectDeps(result.Module.DirectDeps)
	input.DeploymentHints = cp.scanDeploymentHints(projectPath)

	return input
}

func (cp *ContextPreparer) compressPackages(packages []analyzer.PackageInfo, modulePath string) []PackageSummary {
	summaries := make([]PackageSummary, 0, len(packages))
	for _, pkg := range packages {
		summary := PackageSummary{
			Name:       pkg.Name,
			ImportPath: pkg.ImportPath,
		}

		for _, s := range pkg.Structs {
			summary.Structs = append(summary.Structs, s.Name)
		}
		for _, i := range pkg.Interfaces {
			summary.Interfaces = append(summary.Interfaces, i.Name)
		}
		for _, f := range pkg.Functions {
			summary.Functions = append(summary.Functions, f.Name)
		}

		for _, imp := range pkg.Imports {
			if strings.HasPrefix(imp, modulePath) {
				summary.InternalImports = append(summary.InternalImports, imp)
			} else {
				summary.ExternalImports = append(summary.ExternalImports, imp)
			}
		}

		summaries = append(summaries, summary)
	}
	return summaries
}

func (cp *ContextPreparer) deduplicateExternalInteractions(interactions []analyzer.ExternalInteraction) []ExternalInteractionSummary {
	seen := make(map[string]struct{})
	var result []ExternalInteractionSummary

	for _, ei := range interactions {
		key := string(ei.Kind) + "|" + ei.Technology
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, ExternalInteractionSummary{
			Kind:       string(ei.Kind),
			Technology: ei.Technology,
		})
	}
	return result
}

func (cp *ContextPreparer) mapEntrypoints(entrypoints []analyzer.Entrypoint) []EntrypointSummary {
	summaries := make([]EntrypointSummary, 0, len(entrypoints))
	for _, ep := range entrypoints {
		summaries = append(summaries, EntrypointSummary{
			Kind:    ep.Kind,
			Package: ep.PkgPath,
			Route:   ep.Route,
		})
	}
	return summaries
}

func (cp *ContextPreparer) filterCallGraph(calls []analyzer.FunctionCall) []CallEdgeSummary {
	seen := make(map[string]struct{})
	var edges []CallEdgeSummary

	for _, call := range calls {
		if call.CallerType == "" || call.CalleeType == "" {
			continue
		}
		if call.CallerPkg == call.CalleePkg && call.CallerType == call.CalleeType {
			continue
		}

		key := call.CallerPkg + "|" + call.CallerType + "|" + call.CalleePkg + "|" + call.CalleeType
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		edges = append(edges, CallEdgeSummary{
			CallerPkg:  call.CallerPkg,
			CallerType: call.CallerType,
			CalleePkg:  call.CalleePkg,
			CalleeType: call.CalleeType,
		})
	}
	return edges
}

func (cp *ContextPreparer) mapInterfaceImpls(impls []analyzer.InterfaceImpl) []InterfaceImplSummary {
	summaries := make([]InterfaceImplSummary, 0, len(impls))
	for _, impl := range impls {
		summaries = append(summaries, InterfaceImplSummary{
			StructPkg:     impl.StructPkg,
			StructName:    impl.StructName,
			InterfacePkg:  impl.InterfacePkg,
			InterfaceName: impl.InterfaceName,
		})
	}
	return summaries
}

func (cp *ContextPreparer) collectDirectDeps(deps []analyzer.Dependency) []string {
	paths := make([]string, 0, len(deps))
	for _, dep := range deps {
		paths = append(paths, dep.Path)
	}
	return paths
}

func (cp *ContextPreparer) scanDeploymentHints(projectPath string) []string {
	var hints []string

	candidates := []struct {
		file string
		hint string
	}{
		{"Dockerfile", "Dockerfile present"},
		{"docker-compose.yml", "docker-compose.yml present"},
		{"docker-compose.yaml", "docker-compose.yaml present"},
	}

	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(projectPath, c.file)); err == nil {
			hints = append(hints, c.hint)
		}
	}

	// Check for k8s manifests directory
	k8sDirs := []string{"k8s", "kubernetes", "deploy", "manifests"}
	for _, dir := range k8sDirs {
		dirPath := filepath.Join(projectPath, dir)
		info, err := os.Stat(dirPath)
		if err == nil && info.IsDir() {
			hints = append(hints, dir+"/ directory present (likely Kubernetes manifests)")
			break
		}
	}

	return hints
}
