package project

import (
	"context"
	"fmt"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAnalyzer is a test double for the analyzer.Analyzer interface.
type mockAnalyzer struct {
	result *analyzer.AnalysisResult
	err    error
}

func (m *mockAnalyzer) AnalyzeProject(_ context.Context, _ string) (*analyzer.AnalysisResult, error) {
	return m.result, m.err
}

// mockBuilder is a test double for the c4model.ModelBuilder interface.
type mockBuilder struct {
	model *c4model.C4Model
	err   error
}

func (m *mockBuilder) BuildFromAnalysis(_ *analyzer.AnalysisResult) (*c4model.C4Model, error) {
	return m.model, m.err
}

func minimalAnalysisResult() *analyzer.AnalysisResult {
	return &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/example/myapp",
		},
		Packages: []analyzer.PackageInfo{
			{
				Name:       "main",
				ImportPath: "github.com/example/myapp",
				Dir:        ".",
			},
		},
		ImportGraph: map[string][]string{},
	}
}

func minimalInterpretation() *llm.InterpretationResult {
	return &llm.InterpretationResult{
		SystemContext: llm.SystemContextInterpretation{
			Name:        "MyApp",
			Description: "A sample application",
		},
		Containers: []llm.ContainerInterpretation{
			{ID: "container-api", Name: "API", Technology: "Go", Type: "service"},
		},
		Components: []llm.ComponentInterpretation{
			{ID: "comp-handler", Name: "Handler", Role: "controller", ContainerID: "container-api"},
		},
	}
}

func TestService_AnalyzeFromPath_StaticOnly(t *testing.T) {
	analysisResult := minimalAnalysisResult()
	expectedModel := &c4model.C4Model{
		Systems: []c4model.System{{ID: "system-main", Name: "myapp"}},
	}

	svc := NewService(
		&mockAnalyzer{result: analysisResult},
		&mockBuilder{model: expectedModel},
	)

	proj, err := svc.AnalyzeFromPath(context.Background(), "/tmp/myapp", "myapp")
	require.NoError(t, err)

	assert.Equal(t, StatusComplete, proj.Status)
	assert.NotNil(t, proj.Model)
	assert.Equal(t, "system-main", proj.Model.Systems[0].ID)
	assert.Equal(t, proj.ID, proj.Model.ProjectID)
}

func TestService_AnalyzeFromPath_WithInterpreter(t *testing.T) {
	analysisResult := minimalAnalysisResult()
	interp := &llm.MockInterpreter{
		Result: minimalInterpretation(),
	}

	svc := NewService(
		&mockAnalyzer{result: analysisResult},
		&mockBuilder{}, // fallback builder, should not be used
		WithInterpreter(interp),
	)

	proj, err := svc.AnalyzeFromPath(context.Background(), "/tmp/myapp", "myapp")
	require.NoError(t, err)

	// Interpreter was called
	require.NotNil(t, interp.CalledWith)
	assert.Equal(t, "github.com/example/myapp", interp.CalledWith.ModulePath)

	// Model was built via AIModelBuilder (not the mock fallback builder)
	assert.Equal(t, StatusComplete, proj.Status)
	require.NotNil(t, proj.Model)
	assert.Equal(t, "MyApp", proj.Model.Systems[0].Name)
}

func TestService_AnalyzeFromPath_AIFailureFallback(t *testing.T) {
	analysisResult := minimalAnalysisResult()
	fallbackModel := &c4model.C4Model{
		Systems: []c4model.System{{ID: "static-system", Name: "fallback"}},
	}

	interp := &llm.MockInterpreter{
		Err: fmt.Errorf("LLM unavailable"),
	}

	svc := NewService(
		&mockAnalyzer{result: analysisResult},
		&mockBuilder{model: fallbackModel},
		WithInterpreter(interp),
	)

	proj, err := svc.AnalyzeFromPath(context.Background(), "/tmp/myapp", "myapp")
	require.NoError(t, err)

	// Falls back to the static builder
	assert.Equal(t, StatusComplete, proj.Status)
	require.NotNil(t, proj.Model)
	assert.Equal(t, "static-system", proj.Model.Systems[0].ID)
}

func TestService_AnalyzeFromPath_AnalyzerFailure(t *testing.T) {
	svc := NewService(
		&mockAnalyzer{err: fmt.Errorf("cannot parse go.mod")},
		&mockBuilder{},
	)

	proj, err := svc.AnalyzeFromPath(context.Background(), "/tmp/bad", "bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "analysis failed")
	assert.Equal(t, StatusFailed, proj.Status)
	assert.Equal(t, "cannot parse go.mod", proj.Error)
}

func TestService_Get(t *testing.T) {
	analysisResult := minimalAnalysisResult()
	model := &c4model.C4Model{}

	svc := NewService(
		&mockAnalyzer{result: analysisResult},
		&mockBuilder{model: model},
	)

	proj, err := svc.AnalyzeFromPath(context.Background(), "/tmp/myapp", "myapp")
	require.NoError(t, err)

	found, ok := svc.Get(proj.ID)
	assert.True(t, ok)
	assert.Equal(t, proj.ID, found.ID)
	assert.Equal(t, "myapp", found.Name)

	_, ok = svc.Get("nonexistent")
	assert.False(t, ok)
}

func TestService_List(t *testing.T) {
	model := &c4model.C4Model{}

	svc := NewService(
		&mockAnalyzer{result: minimalAnalysisResult()},
		&mockBuilder{model: model},
	)

	_, err := svc.AnalyzeFromPath(context.Background(), "/tmp/a", "projA")
	require.NoError(t, err)
	_, err = svc.AnalyzeFromPath(context.Background(), "/tmp/b", "projB")
	require.NoError(t, err)

	projects := svc.List()
	assert.Len(t, projects, 2)

	names := map[string]bool{}
	for _, p := range projects {
		names[p.Name] = true
	}
	assert.True(t, names["projA"])
	assert.True(t, names["projB"])
}

func TestService_WithInterpreter_Option(t *testing.T) {
	interp := &llm.MockInterpreter{}

	svc := NewService(
		&mockAnalyzer{},
		&mockBuilder{},
		WithInterpreter(interp),
	)

	// The interpreter field is unexported, so we verify it indirectly:
	// it should be non-nil by checking that builderForResult tries AI path
	assert.NotNil(t, svc.interpreter)
}
