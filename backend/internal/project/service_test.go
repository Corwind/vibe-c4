package project

import (
	"context"
	"fmt"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
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

func TestService_AnalyzeFromPath(t *testing.T) {
	analysisResult := minimalAnalysisResult()
	expectedModel := &c4model.C4Model{
		Systems: []c4model.System{{ID: "system-main", Name: "myapp"}},
	}

	svc := NewService(
		&mockAnalyzer{result: analysisResult},
		&mockBuilder{model: expectedModel},
		nil,
	)

	proj, err := svc.AnalyzeFromPath(context.Background(), "/tmp/myapp", "myapp")
	require.NoError(t, err)

	assert.Equal(t, StatusComplete, proj.Status)
	assert.NotNil(t, proj.Model)
	assert.Equal(t, "system-main", proj.Model.Systems[0].ID)
	assert.Equal(t, proj.ID, proj.Model.ProjectID)
}

func TestService_AnalyzeFromPath_AnalyzerFailure(t *testing.T) {
	svc := NewService(
		&mockAnalyzer{err: fmt.Errorf("cannot parse go.mod")},
		&mockBuilder{},
		nil,
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
		nil,
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
		nil,
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

func TestService_AnalyzeFromPathWithMode_AI_Success(t *testing.T) {
	analysisResult := minimalAnalysisResult()
	aiModel := &c4model.C4Model{
		Systems: []c4model.System{{ID: "system-ai", Name: "AI Generated"}},
	}

	svc := NewService(
		&mockAnalyzer{result: analysisResult},
		&mockBuilder{model: &c4model.C4Model{}},
		&mockBuilder{model: aiModel},
	)

	proj, err := svc.AnalyzeFromPathWithMode(context.Background(), "/tmp/myapp", "myapp", "ai")
	require.NoError(t, err)

	assert.Equal(t, StatusComplete, proj.Status)
	assert.Equal(t, "system-ai", proj.Model.Systems[0].ID)
}

func TestService_AnalyzeFromPathWithMode_AI_NoKey(t *testing.T) {
	svc := NewService(
		&mockAnalyzer{result: minimalAnalysisResult()},
		&mockBuilder{},
		nil,
	)

	proj, err := svc.AnalyzeFromPathWithMode(context.Background(), "/tmp/myapp", "myapp", "ai")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no API key configured")
	assert.Equal(t, StatusFailed, proj.Status)
}

func TestService_AnalyzeFromPathWithMode_StaticDefault(t *testing.T) {
	analysisResult := minimalAnalysisResult()
	staticModel := &c4model.C4Model{
		Systems: []c4model.System{{ID: "system-static", Name: "Static"}},
	}

	svc := NewService(
		&mockAnalyzer{result: analysisResult},
		&mockBuilder{model: staticModel},
		&mockBuilder{model: &c4model.C4Model{}},
	)

	proj, err := svc.AnalyzeFromPathWithMode(context.Background(), "/tmp/myapp", "myapp", "static")
	require.NoError(t, err)

	assert.Equal(t, StatusComplete, proj.Status)
	assert.Equal(t, "system-static", proj.Model.Systems[0].ID)
}
