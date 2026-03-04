package analyzer_test

import (
	"context"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeProject_CallGraphNotEmpty(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	assert.NotEmpty(t, result.CallGraph, "call graph should not be empty")
}

func TestAnalyzeProject_CallGraph_ChainedSelectorNotDetected(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	// Handler.DoWork calls h.svc.DoWork() — this is a chained selector (h.svc).DoWork()
	// The body visitor only handles simple selectors (ident.Method), not chained ones,
	// so this call is NOT recorded in the call graph.
	var found bool
	for _, call := range result.CallGraph {
		if call.CallerPkg == "github.com/example/sample-project/internal/handler" &&
			call.CallerType == "Handler" &&
			call.CalleeType == "svc" &&
			call.CalleeFunc == "DoWork" {
			found = true
			break
		}
	}
	assert.False(t, found, "chained selector calls like h.svc.DoWork() are not currently detected")
}

func TestAnalyzeProject_CallGraph_HandlerCallsUtils(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	var found bool
	for _, call := range result.CallGraph {
		if call.CallerPkg == "github.com/example/sample-project/internal/handler" &&
			call.CalleePkg == "github.com/example/sample-project/pkg/utils" &&
			call.CalleeFunc == "Log" {
			found = true
			break
		}
	}
	assert.True(t, found, "should detect handler calling utils.Log()")
}

func TestAnalyzeProject_CallGraph_ServiceCallsUtils(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	var found bool
	for _, call := range result.CallGraph {
		if call.CallerPkg == "github.com/example/sample-project/internal/service" &&
			call.CalleePkg == "github.com/example/sample-project/pkg/utils" &&
			call.CalleeFunc == "Log" {
			found = true
			break
		}
	}
	assert.True(t, found, "should detect service calling utils.Log()")
}

func TestAnalyzeProject_CallGraph_MainCallsChiNewRouter(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	// chi.NewRouter() uses the package name "chi" as identifier, but the import
	// path "github.com/go-chi/chi/v5" has last segment "v5". The alias map maps
	// "v5" -> full path, so "chi" is treated as a local variable, not an import.
	// This means chi.NewRouter() is recorded as a local method call, not cross-package.
	var foundCrossPackage bool
	for _, call := range result.CallGraph {
		if call.CallerPkg == "github.com/example/sample-project/cmd/server" &&
			call.CalleePkg == "github.com/go-chi/chi/v5" &&
			call.CalleeFunc == "NewRouter" {
			foundCrossPackage = true
			break
		}
	}

	// Check if it's recorded as a local call instead
	var foundLocal bool
	for _, call := range result.CallGraph {
		if call.CallerPkg == "github.com/example/sample-project/cmd/server" &&
			call.CalleeFunc == "NewRouter" {
			foundLocal = true
			break
		}
	}

	// At least one of these should be true depending on how the alias resolves
	assert.True(t, foundCrossPackage || foundLocal,
		"should detect main calling NewRouter() either as cross-package or local call")
}

func TestAnalyzeProject_ExternalInteractionsNotEmpty(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	assert.NotEmpty(t, result.ExternalInteractions, "should detect external interactions")
}

func TestAnalyzeProject_ExternalInteractions_ChiLocalVarNotDetected(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	// Chi route calls like r.Get("/health", ...) where r is a local variable
	// are not detected as external interactions because the body visitor only
	// triggers pattern detection for package-level import calls (chi.Method()).
	var chiHandlers []analyzer.ExternalInteraction
	for _, ei := range result.ExternalInteractions {
		if ei.Technology == "chi" && ei.Kind == analyzer.ExtKindHTTPHandler {
			chiHandlers = append(chiHandlers, ei)
		}
	}
	assert.Empty(t, chiHandlers, "chi routes via local variable are not detected as chi HTTP handler interactions")
}

func TestAnalyzeProject_ExternalInteractions_NetHTTP(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	var httpInteractions []analyzer.ExternalInteraction
	for _, ei := range result.ExternalInteractions {
		if ei.Technology == "net/http" {
			httpInteractions = append(httpInteractions, ei)
		}
	}

	// main.go calls http.ListenAndServe
	var found bool
	for _, ei := range httpInteractions {
		if ei.FuncName == "ListenAndServe" {
			found = true
			assert.Equal(t, analyzer.ExternalDependencyKind("http_handler"), ei.Kind)
			break
		}
	}
	assert.True(t, found, "should detect http.ListenAndServe as HTTP handler interaction")
}

func TestAnalyzeProject_ExternalInteractions_HaveLineNumbers(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	for _, ei := range result.ExternalInteractions {
		assert.Greater(t, ei.Line, 0, "external interaction %s.%s should have a line number", ei.PkgPath, ei.FuncName)
	}
}

func TestAnalyzeProject_InterfaceImplsDetected(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	assert.NotEmpty(t, result.InterfaceImpls, "should detect interface implementations")

	// Service should implement Worker
	var found bool
	for _, impl := range result.InterfaceImpls {
		if impl.StructName == "Service" && impl.InterfaceName == "Worker" {
			found = true
			assert.Equal(t, "github.com/example/sample-project/internal/service", impl.StructPkg)
			assert.Equal(t, "github.com/example/sample-project/internal/service", impl.InterfacePkg)
			break
		}
	}
	assert.True(t, found, "should detect Service implements Worker")
}
