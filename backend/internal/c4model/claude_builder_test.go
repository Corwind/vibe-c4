package c4model

import (
	"context"
	"fmt"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/config"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLLMClient implements llm.Client for testing.
type mockLLMClient struct {
	response *llm.Response
	err      error
}

func (m *mockLLMClient) Complete(_ context.Context, _ llm.Request) (*llm.Response, error) {
	return m.response, m.err
}

func testConfig() config.ClaudeConfig {
	return config.ClaudeConfig{
		APIKey:      "test-key",
		Model:       "claude-sonnet-4-20250514",
		MaxTokens:   4096,
		TokenBudget: 80000,
		TimeoutSecs: 30,
	}
}

func testAnalysisResult(t *testing.T) *analyzer.AnalysisResult {
	t.Helper()
	return &analyzer.AnalysisResult{
		ProjectPath: t.TempDir(),
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/example/test",
			GoVersion:  "1.21",
		},
		Packages: []analyzer.PackageInfo{
			{
				Name:       "main",
				ImportPath: "github.com/example/test",
				GoFiles:    []string{"main.go"},
			},
		},
	}
}

func TestClaudeModelBuilderHappyPath(t *testing.T) {
	mock := &mockLLMClient{
		response: &llm.Response{
			Content: validC4ModelJSON(),
			Usage:   llm.Usage{InputTokens: 500, OutputTokens: 200},
		},
	}

	builder := NewClaudeModelBuilder(mock, testConfig(), NewSmartModelBuilder())
	result := testAnalysisResult(t)

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, "test-project", model.ProjectID)
	assert.Len(t, model.Systems, 1)
	assert.Equal(t, "system-myapp", model.Systems[0].ID)
	assert.Len(t, model.Containers, 1)
	assert.Len(t, model.Components, 1)
	assert.Len(t, model.CodeElements, 1)
	assert.Len(t, model.Relationships, 1)
}

func TestClaudeModelBuilderAPIErrorFallsBack(t *testing.T) {
	mock := &mockLLMClient{
		err: fmt.Errorf("API connection refused"),
	}

	fallback := NewSmartModelBuilder()
	builder := NewClaudeModelBuilder(mock, testConfig(), fallback)
	result := testAnalysisResult(t)

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)
	require.NotNil(t, model)

	// Should have gotten a model from the SmartModelBuilder fallback
	assert.Greater(t, len(model.Systems), 0)
}

func TestClaudeModelBuilderInvalidJSONFallsBack(t *testing.T) {
	mock := &mockLLMClient{
		response: &llm.Response{
			Content: "this is not valid json at all",
			Usage:   llm.Usage{InputTokens: 100, OutputTokens: 10},
		},
	}

	fallback := NewSmartModelBuilder()
	builder := NewClaudeModelBuilder(mock, testConfig(), fallback)
	result := testAnalysisResult(t)

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)
	require.NotNil(t, model)

	// Should have gotten a model from the SmartModelBuilder fallback
	assert.Greater(t, len(model.Systems), 0)
}

func TestClaudeModelBuilderInvalidModelFallsBack(t *testing.T) {
	// Valid JSON but invalid C4 model (no systems)
	invalidModel := `{"project_id": "test", "systems": [], "containers": [], "components": [], "code_elements": [], "relationships": []}`

	mock := &mockLLMClient{
		response: &llm.Response{
			Content: invalidModel,
			Usage:   llm.Usage{InputTokens: 100, OutputTokens: 10},
		},
	}

	fallback := NewSmartModelBuilder()
	builder := NewClaudeModelBuilder(mock, testConfig(), fallback)
	result := testAnalysisResult(t)

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)
	require.NotNil(t, model)

	// Should have gotten a model from the SmartModelBuilder fallback
	assert.Greater(t, len(model.Systems), 0)
}

func TestClaudeModelBuilderNilResult(t *testing.T) {
	mock := &mockLLMClient{}
	builder := NewClaudeModelBuilder(mock, testConfig(), NewSmartModelBuilder())

	model, err := builder.BuildFromAnalysis(nil)
	assert.Error(t, err)
	assert.Nil(t, model)
}
