package analyzer

import (
	"go/ast"
	"go/token"
	"strings"
)

type bodyVisitor struct {
	fset           *token.FileSet
	callerPkg      string
	callerType     string
	callerFunc     string
	importAliasMap map[string]string
	modulePath     string

	calls       []FunctionCall
	externals   []ExternalInteraction
	entrypoints []Entrypoint
}

func (v *bodyVisitor) Visit(node ast.Node) ast.Visitor {
	callExpr, ok := node.(*ast.CallExpr)
	if !ok {
		return v
	}

	switch fn := callExpr.Fun.(type) {
	case *ast.SelectorExpr:
		v.visitSelectorCall(callExpr, fn)
	case *ast.Ident:
		v.visitIdentCall(fn)
	}

	return v
}

func (v *bodyVisitor) visitSelectorCall(callExpr *ast.CallExpr, sel *ast.SelectorExpr) {
	funcName := sel.Sel.Name
	pos := v.fset.Position(callExpr.Pos())

	xIdent, isIdent := sel.X.(*ast.Ident)
	if !isIdent {
		return
	}

	alias := xIdent.Name
	fullPkg, isImport := v.importAliasMap[alias]

	if isImport {
		// Cross-package call: pkg.Func()
		fc := FunctionCall{
			CallerPkg:      v.callerPkg,
			CallerType:     v.callerType,
			CallerFunc:     v.callerFunc,
			CalleePkg:      fullPkg,
			CalleeFunc:     funcName,
			CalleePkgAlias: alias,
			Line:           pos.Line,
		}
		v.calls = append(v.calls, fc)

		// Check for external pattern
		if ei := DetectExternalPattern(fullPkg, "", funcName); ei != nil {
			ei.FilePath = pos.Filename
			ei.Line = pos.Line

			// For chi router methods, extract route pattern from first string argument
			if ei.Technology == "chi" {
				if route := extractFirstStringArg(callExpr); route != "" {
					ei.Detail = route
					v.entrypoints = append(v.entrypoints, Entrypoint{
						Kind:     "http_route",
						PkgPath:  v.callerPkg,
						FuncName: v.callerFunc,
						Route:    strings.ToUpper(funcName) + " " + route,
						Line:     pos.Line,
					})
				}
			}

			v.externals = append(v.externals, *ei)
		}
	} else {
		// Method call on a local variable: obj.Method()
		// We record it as a same-package call with the alias as the type name
		fc := FunctionCall{
			CallerPkg:  v.callerPkg,
			CallerType: v.callerType,
			CallerFunc: v.callerFunc,
			CalleePkg:  v.callerPkg,
			CalleeType: alias,
			CalleeFunc: funcName,
			Line:       pos.Line,
		}
		v.calls = append(v.calls, fc)
	}
}

func (v *bodyVisitor) visitIdentCall(ident *ast.Ident) {
	fc := FunctionCall{
		CallerPkg:  v.callerPkg,
		CallerType: v.callerType,
		CallerFunc: v.callerFunc,
		CalleePkg:  v.callerPkg,
		CalleeFunc: ident.Name,
	}
	v.calls = append(v.calls, fc)
}

// extractFirstStringArg returns the value of the first string literal argument in a call, if any.
func extractFirstStringArg(call *ast.CallExpr) string {
	if len(call.Args) == 0 {
		return ""
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	return strings.Trim(lit.Value, `"`)
}

// buildImportAliasMap builds a mapping from import alias (or last path element) to full import path.
func buildImportAliasMap(file *ast.File) map[string]string {
	m := make(map[string]string)
	for _, imp := range file.Imports {
		fullPath := strings.Trim(imp.Path.Value, `"`)
		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(fullPath, "/")
			alias = parts[len(parts)-1]
		}
		if alias == "_" || alias == "." {
			continue
		}
		m[alias] = fullPath
	}
	return m
}

// analyzeBody walks a function body and extracts calls, external interactions, and entrypoints.
func analyzeBody(fset *token.FileSet, fn *ast.FuncDecl, callerPkg, callerType, modulePath string, importAliasMap map[string]string) ([]FunctionCall, []ExternalInteraction, []Entrypoint) {
	if fn.Body == nil {
		return nil, nil, nil
	}

	v := &bodyVisitor{
		fset:           fset,
		callerPkg:      callerPkg,
		callerType:     callerType,
		callerFunc:     fn.Name.Name,
		importAliasMap: importAliasMap,
		modulePath:     modulePath,
	}

	ast.Walk(v, fn.Body)

	return v.calls, v.externals, v.entrypoints
}
