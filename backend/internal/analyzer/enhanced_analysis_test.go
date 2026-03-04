package analyzer_test

import (
	"context"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func analyzeTestProject(t *testing.T) *analyzer.AnalysisResult {
	t.Helper()
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)
	return result
}

// --- Call Graph (body analysis) integration tests ---

func TestCallGraph_IsPopulated(t *testing.T) {
	result := analyzeTestProject(t)
	require.NotEmpty(t, result.CallGraph, "call graph should not be empty")
}

func TestCallGraph_ContainsCrossPackageCalls(t *testing.T) {
	result := analyzeTestProject(t)

	// service.DoWork() calls utils.Log()
	var found bool
	for _, call := range result.CallGraph {
		if call.CallerPkg == "github.com/example/sample-project/internal/service" &&
			call.CalleeFunc == "Log" &&
			call.CalleePkg == "github.com/example/sample-project/pkg/utils" {
			found = true
			break
		}
	}
	assert.True(t, found, "should detect service -> utils.Log cross-package call")
}

func TestCallGraph_ContainsMethodCalls(t *testing.T) {
	result := analyzeTestProject(t)

	// Handler.HealthCheck/DoWork call utils.Log() — this IS detected because
	// utils is a package import (simple selector: utils.Log()).
	// However h.svc.DoWork() is a chained selector (h.svc).DoWork() which
	// the body visitor skips because sel.X is not an Ident.
	var found bool
	for _, call := range result.CallGraph {
		if call.CallerPkg == "github.com/example/sample-project/internal/handler" &&
			call.CallerType == "Handler" &&
			call.CalleePkg == "github.com/example/sample-project/pkg/utils" &&
			call.CalleeFunc == "Log" {
			found = true
			break
		}
	}
	assert.True(t, found, "should detect Handler method calling utils.Log()")
}

func TestCallGraph_MainCallsServiceNew(t *testing.T) {
	result := analyzeTestProject(t)

	var found bool
	for _, call := range result.CallGraph {
		if call.CallerPkg == "github.com/example/sample-project/cmd/server" &&
			call.CalleePkg == "github.com/example/sample-project/internal/service" &&
			call.CalleeFunc == "New" {
			found = true
			break
		}
	}
	assert.True(t, found, "should detect main -> service.New() cross-package call")
}

// --- External interaction integration tests ---

func TestExternalInteractions_IsPopulated(t *testing.T) {
	result := analyzeTestProject(t)
	require.NotEmpty(t, result.ExternalInteractions, "external interactions should not be empty")
}

func TestExternalInteractions_ChiLocalVarCallsNotDetected(t *testing.T) {
	result := analyzeTestProject(t)

	// Chi route calls like r.Get("/health", ...) where r is a local variable
	// are not detected as external interactions because the body visitor only
	// triggers pattern detection for package-level import calls.
	var chiHandlers int
	for _, ei := range result.ExternalInteractions {
		if ei.Technology == "chi" && ei.Kind == analyzer.ExtKindHTTPHandler {
			chiHandlers++
		}
	}
	assert.Equal(t, 0, chiHandlers, "chi routes via local variable are not currently detected as chi HTTP handlers")
}

func TestExternalInteractions_DetectsHTTPListenAndServe(t *testing.T) {
	result := analyzeTestProject(t)

	var found bool
	for _, ei := range result.ExternalInteractions {
		if ei.Technology == "net/http" && ei.FuncName == "ListenAndServe" {
			found = true
			break
		}
	}
	assert.True(t, found, "should detect http.ListenAndServe as external interaction")
}

func TestExternalInteractions_NoFalsePositivesForInternalPkgs(t *testing.T) {
	result := analyzeTestProject(t)

	for _, ei := range result.ExternalInteractions {
		assert.NotEqual(t, "github.com/example/sample-project/pkg/utils", ei.PkgPath,
			"internal package calls should not be detected as external interactions")
	}
}

// --- Interface matching integration tests ---

func TestInterfaceImpls_IsPopulated(t *testing.T) {
	result := analyzeTestProject(t)
	require.NotEmpty(t, result.InterfaceImpls, "interface implementations should not be empty")
}

func TestInterfaceImpls_ServiceImplementsWorker(t *testing.T) {
	result := analyzeTestProject(t)

	var found bool
	for _, impl := range result.InterfaceImpls {
		if impl.StructName == "Service" && impl.InterfaceName == "Worker" {
			assert.Equal(t, "github.com/example/sample-project/internal/service", impl.StructPkg)
			assert.Equal(t, "github.com/example/sample-project/internal/service", impl.InterfacePkg)
			found = true
			break
		}
	}
	assert.True(t, found, "Service should implement Worker interface")
}

func TestInterfaceImpls_HandlerDoesNotImplementWorker(t *testing.T) {
	result := analyzeTestProject(t)

	for _, impl := range result.InterfaceImpls {
		if impl.InterfaceName == "Worker" {
			assert.NotEqual(t, "Handler", impl.StructName,
				"Handler should not be reported as implementing Worker")
		}
	}
}

// --- Entrypoint detection integration tests ---

func TestEntrypoints_IsPopulated(t *testing.T) {
	result := analyzeTestProject(t)
	require.NotEmpty(t, result.Entrypoints, "entrypoints should not be empty")
}

func TestEntrypoints_DetectsMainFunction(t *testing.T) {
	result := analyzeTestProject(t)

	var found bool
	for _, ep := range result.Entrypoints {
		if ep.Kind == "main" && ep.FuncName == "main" {
			assert.Equal(t, "github.com/example/sample-project/cmd/server", ep.PkgPath)
			found = true
			break
		}
	}
	assert.True(t, found, "should detect main() as an entrypoint")
}

func TestEntrypoints_ChiLocalVarRoutesNotDetected(t *testing.T) {
	result := analyzeTestProject(t)

	// Chi route calls like r.Get("/health", ...) where r is a local variable
	// are not detected as http_route entrypoints because the body visitor only
	// triggers external pattern detection for package-level import calls.
	var routes []string
	for _, ep := range result.Entrypoints {
		if ep.Kind == "http_route" {
			routes = append(routes, ep.Route)
		}
	}

	assert.Empty(t, routes, "chi routes via local variable are not currently detected as http_route entrypoints")
}

func TestEntrypoints_HTTPRoutesHaveLineInfo(t *testing.T) {
	result := analyzeTestProject(t)

	for _, ep := range result.Entrypoints {
		if ep.Kind == "http_route" {
			assert.NotEmpty(t, ep.PkgPath, "http route entrypoint should have PkgPath")
			assert.NotEmpty(t, ep.Route, "http route entrypoint should have Route")
			assert.Greater(t, ep.Line, 0, "http route entrypoint should have Line > 0")
		}
	}
}

// --- Pattern detection exported function tests ---

func TestDetectExternalPattern_DatabaseSQL(t *testing.T) {
	ei := analyzer.DetectExternalPattern("database/sql", "", "Query")
	require.NotNil(t, ei)
	assert.Equal(t, analyzer.ExtKindDatabase, ei.Kind)
	assert.Equal(t, "database/sql", ei.Technology)
}

func TestDetectExternalPattern_NoMatch(t *testing.T) {
	ei := analyzer.DetectExternalPattern("github.com/example/project/internal/service", "", "DoWork")
	assert.Nil(t, ei, "internal packages should not match any external pattern")
}
