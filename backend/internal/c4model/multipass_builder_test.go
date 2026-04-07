package c4model_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/Corwind/vibe-c4/backend/internal/config"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pipelineMock dispatches responses based on the system prompt content.
type pipelineMock struct {
	mu            sync.Mutex
	calls         []llm.Request
	pass1Response *llm.Response
	pass2Response *llm.Response
	pass3Response *llm.Response
	pass1Err      error
	pass2Err      error
	pass3Err      error
}

func (m *pipelineMock) Complete(_ context.Context, req llm.Request) (*llm.Response, error) {
	m.mu.Lock()
	m.calls = append(m.calls, req)
	m.mu.Unlock()

	sp := strings.ToLower(req.SystemPrompt)

	switch {
	case strings.Contains(sp, "system context") || strings.Contains(sp, "level 1"):
		if m.pass1Err != nil {
			return nil, m.pass1Err
		}
		return m.pass1Response, nil
	case (strings.Contains(sp, "component") && strings.Contains(sp, "level 3")) || strings.Contains(sp, "container"):
		if m.pass2Err != nil {
			return nil, m.pass2Err
		}
		return m.pass2Response, nil
	case (strings.Contains(sp, "code") && strings.Contains(sp, "level 4")) || strings.Contains(sp, "component"):
		if m.pass3Err != nil {
			return nil, m.pass3Err
		}
		return m.pass3Response, nil
	default:
		return nil, fmt.Errorf("unexpected system prompt: %s", req.SystemPrompt)
	}
}

func multipassTestConfig() config.ClaudeConfig {
	return config.ClaudeConfig{
		APIKey:         "test-key",
		Model:          "claude-sonnet-4-6-20250514",
		MaxTokens:      4096,
		TokenBudget:    80000,
		TimeoutSecs:    30,
		MaxConcurrency: 2,
	}
}

func makePass1JSON(systems []c4model.System, containers []c4model.Container, rels []c4model.Relationship) string {
	result := c4model.Pass1Result{
		Systems:       systems,
		Containers:    containers,
		Relationships: rels,
	}
	data, _ := json.Marshal(result)
	return string(data)
}

func makePass2JSON(components []c4model.Component, rels []c4model.Relationship) string {
	result := c4model.Pass2Result{
		Components:    components,
		Relationships: rels,
	}
	data, _ := json.Marshal(result)
	return string(data)
}

func makePass3JSON(codeElements []c4model.CodeElement) string {
	result := c4model.Pass3Result{
		CodeElements: codeElements,
	}
	data, _ := json.Marshal(result)
	return string(data)
}

// setupTestRepo creates a temporary directory with go.mod and subdirectories
// matching the provided package paths.
func setupTestRepo(t *testing.T, modulePath string, packageDirs []string) string {
	t.Helper()
	dir := t.TempDir()

	goModContent := fmt.Sprintf("module %s\n\ngo 1.22\n", modulePath)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0644))

	for _, pkgDir := range packageDirs {
		pkgPath := filepath.Join(dir, pkgDir)
		require.NoError(t, os.MkdirAll(pkgPath, 0755))
		goFile := filepath.Join(pkgPath, "main.go")
		require.NoError(t, os.WriteFile(goFile, []byte(fmt.Sprintf("package %s\n", filepath.Base(pkgDir))), 0644))
	}

	return dir
}

func multipassAnalysisResult(projectPath, modulePath string, pkgPaths []string) *analyzer.AnalysisResult {
	packages := make([]analyzer.PackageInfo, len(pkgPaths))
	for i, p := range pkgPaths {
		packages[i] = analyzer.PackageInfo{
			Name:       filepath.Base(p),
			ImportPath: p,
			Dir:        strings.TrimPrefix(p, modulePath+"/"),
			Role:       "internal",
		}
	}
	return &analyzer.AnalysisResult{
		ProjectPath: projectPath,
		Module: analyzer.ModuleInfo{
			ModulePath: modulePath,
			GoVersion:  "1.22",
		},
		Packages: packages,
	}
}

func TestMultiPassBuilderHappyPath(t *testing.T) {
	modulePath := "github.com/example/test"
	projectDir := setupTestRepo(t, modulePath, []string{"api", "db"})

	result := multipassAnalysisResult(
		projectDir,
		modulePath,
		[]string{modulePath + "/api", modulePath + "/db"},
	)

	systems := []c4model.System{
		{ID: "system-main", Name: "test", Description: "Main system", ModulePath: modulePath},
	}
	containers := []c4model.Container{
		{ID: "container-api", Name: "API", PackagePath: modulePath + "/api", SystemID: "system-main", Technology: "Go"},
		{ID: "container-db", Name: "DB", PackagePath: modulePath + "/db", SystemID: "system-main", Technology: "Go"},
	}
	pass1Rels := []c4model.Relationship{
		{SourceID: "container-api", TargetID: "container-db", Description: "uses", Level: "container"},
	}

	apiComponents := []c4model.Component{
		{ID: "component-handler", Name: "Handler", ContainerID: "container-api", PackagePath: modulePath + "/api", Technology: "Go"},
		{ID: "component-service", Name: "Service", ContainerID: "container-api", PackagePath: modulePath + "/api", Technology: "Go"},
	}
	dbComponents := []c4model.Component{
		{ID: "component-repo", Name: "Repository", ContainerID: "container-db", PackagePath: modulePath + "/db", Technology: "Go"},
		{ID: "component-model", Name: "Model", ContainerID: "container-db", PackagePath: modulePath + "/db", Technology: "Go"},
	}
	pass2Rels := []c4model.Relationship{
		{SourceID: "component-handler", TargetID: "component-service", Description: "calls", Level: "component"},
	}

	codeElements := []c4model.CodeElement{
		{ID: "code-handler-get", Name: "Get", ComponentID: "component-handler", Type: "method"},
	}

	mock := &pipelineMock{
		pass1Response: &llm.Response{
			Content: makePass1JSON(systems, containers, pass1Rels),
			Usage:   llm.Usage{InputTokens: 1000, OutputTokens: 500},
		},
		pass2Response: &llm.Response{
			Content: makePass2JSON(apiComponents, pass2Rels),
			Usage:   llm.Usage{InputTokens: 500, OutputTokens: 250},
		},
		pass3Response: &llm.Response{
			Content: makePass3JSON(codeElements),
			Usage:   llm.Usage{InputTokens: 200, OutputTokens: 100},
		},
	}

	builder := c4model.NewMultiPassBuilder(mock, multipassTestConfig(), c4model.NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(result)

	require.NoError(t, err)
	require.NotNil(t, model)

	// Verify systems
	assert.Len(t, model.Systems, 1)
	assert.Equal(t, "system-main", model.Systems[0].ID)

	// Verify containers
	assert.Len(t, model.Containers, 2)

	// Verify components: the mock returns the same pass2Response for both containers
	// (apiComponents for both calls), so we get 4 components total (2 per container call).
	// Note: dbComponents would be used if mock dispatched differently per container.
	_ = dbComponents
	assert.GreaterOrEqual(t, len(model.Components), 4)

	// Verify code elements exist
	assert.GreaterOrEqual(t, len(model.CodeElements), 1)

	// Verify parent-child: system should have container IDs
	sys := model.FindSystem("system-main")
	require.NotNil(t, sys)
	assert.Contains(t, sys.ContainerIDs, "container-api")
	assert.Contains(t, sys.ContainerIDs, "container-db")

	// Verify containers have component IDs
	apiContainer := model.FindContainer("container-api")
	require.NotNil(t, apiContainer)
	assert.NotEmpty(t, apiContainer.ComponentIDs)

	// Verify relationships include pass1 and pass2 relationships
	assert.NotEmpty(t, model.Relationships)
}

func TestMultiPassBuilderPass1FailureFallback(t *testing.T) {
	modulePath := "github.com/example/test"
	projectDir := setupTestRepo(t, modulePath, []string{"api"})

	result := multipassAnalysisResult(
		projectDir,
		modulePath,
		[]string{modulePath + "/api"},
	)

	mock := &pipelineMock{
		pass1Err: fmt.Errorf("API rate limited"),
	}

	fallback := c4model.NewSmartModelBuilder()
	builder := c4model.NewMultiPassBuilder(mock, multipassTestConfig(), fallback)
	model, err := builder.BuildFromAnalysis(result)

	// Should fall back to SmartModelBuilder, not return error to caller
	require.NoError(t, err)
	require.NotNil(t, model)
	assert.NotEmpty(t, model.Systems, "fallback should produce systems")
}

func TestMultiPassBuilderPass2PartialFailure(t *testing.T) {
	modulePath := "github.com/example/test"
	projectDir := setupTestRepo(t, modulePath, []string{"api", "db"})

	result := multipassAnalysisResult(
		projectDir,
		modulePath,
		[]string{modulePath + "/api", modulePath + "/db"},
	)

	systems := []c4model.System{
		{ID: "system-main", Name: "test", Description: "Main system", ModulePath: modulePath},
	}
	containers := []c4model.Container{
		{ID: "container-api", Name: "API", PackagePath: modulePath + "/api", SystemID: "system-main"},
		{ID: "container-db", Name: "DB", PackagePath: modulePath + "/db", SystemID: "system-main"},
	}

	mock := &pipelineMock{
		pass1Response: &llm.Response{
			Content: makePass1JSON(systems, containers, nil),
			Usage:   llm.Usage{InputTokens: 1000, OutputTokens: 500},
		},
		pass2Err: fmt.Errorf("pass2 failure"),
		pass3Err: fmt.Errorf("pass3 failure"),
	}

	builder := c4model.NewMultiPassBuilder(mock, multipassTestConfig(), c4model.NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(result)

	require.NoError(t, err)
	require.NotNil(t, model)

	// Should have systems and containers from pass 1
	assert.Len(t, model.Systems, 1)
	assert.Len(t, model.Containers, 2)
	// No components because all pass 2 calls failed
	assert.Empty(t, model.Components)
}

func TestMultiPassBuilderNilResult(t *testing.T) {
	mock := &pipelineMock{}
	builder := c4model.NewMultiPassBuilder(mock, multipassTestConfig(), c4model.NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(nil)

	require.Error(t, err)
	assert.Nil(t, model)
}

func TestMultiPassBuilderNoContainers(t *testing.T) {
	modulePath := "github.com/example/test"
	projectDir := setupTestRepo(t, modulePath, nil)

	result := multipassAnalysisResult(projectDir, modulePath, nil)

	systems := []c4model.System{
		{ID: "system-main", Name: "test", Description: "Standalone system", ModulePath: modulePath},
	}

	mock := &pipelineMock{
		pass1Response: &llm.Response{
			Content: makePass1JSON(systems, nil, nil),
			Usage:   llm.Usage{InputTokens: 500, OutputTokens: 200},
		},
	}

	builder := c4model.NewMultiPassBuilder(mock, multipassTestConfig(), c4model.NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(result)

	require.NoError(t, err)
	require.NotNil(t, model)
	assert.Len(t, model.Systems, 1)
	assert.Empty(t, model.Containers)
	assert.Empty(t, model.Components)

	// Verify no pass2/pass3 calls were made
	mock.mu.Lock()
	defer mock.mu.Unlock()
	assert.Len(t, mock.calls, 1, "only pass1 call should have been made")
}

func TestMergeResultsRebuildIDs(t *testing.T) {
	pass1 := &c4model.Pass1Result{
		Systems: []c4model.System{
			{ID: "sys-1", Name: "System"},
		},
		Containers: []c4model.Container{
			{ID: "cont-1", Name: "Container1", SystemID: "sys-1"},
			{ID: "cont-2", Name: "Container2", SystemID: "sys-1"},
		},
	}
	pass2s := []c4model.Pass2Result{
		{
			Components: []c4model.Component{
				{ID: "comp-a", Name: "CompA", ContainerID: "cont-1"},
				{ID: "comp-b", Name: "CompB", ContainerID: "cont-1"},
			},
		},
		{
			Components: []c4model.Component{
				{ID: "comp-c", Name: "CompC", ContainerID: "cont-2"},
			},
		},
	}
	pass3s := []c4model.Pass3Result{
		{
			CodeElements: []c4model.CodeElement{
				{ID: "code-1", Name: "Func1", ComponentID: "comp-a"},
				{ID: "code-2", Name: "Func2", ComponentID: "comp-a"},
			},
		},
		{
			CodeElements: []c4model.CodeElement{
				{ID: "code-3", Name: "Func3", ComponentID: "comp-c"},
			},
		},
	}

	model := c4model.MergeResults(pass1, pass2s, pass3s)
	require.NotNil(t, model)

	// System should have both container IDs
	sys := model.FindSystem("sys-1")
	require.NotNil(t, sys)
	assert.ElementsMatch(t, []string{"cont-1", "cont-2"}, sys.ContainerIDs)

	// Container 1 should have comp-a and comp-b
	cont1 := model.FindContainer("cont-1")
	require.NotNil(t, cont1)
	assert.ElementsMatch(t, []string{"comp-a", "comp-b"}, cont1.ComponentIDs)

	// Container 2 should have comp-c
	cont2 := model.FindContainer("cont-2")
	require.NotNil(t, cont2)
	assert.ElementsMatch(t, []string{"comp-c"}, cont2.ComponentIDs)

	// Component comp-a should have code-1 and code-2
	compA := model.FindComponent("comp-a")
	require.NotNil(t, compA)
	assert.ElementsMatch(t, []string{"code-1", "code-2"}, compA.CodeElements)

	// Component comp-c should have code-3
	compC := model.FindComponent("comp-c")
	require.NotNil(t, compC)
	assert.ElementsMatch(t, []string{"code-3"}, compC.CodeElements)
}

func TestDeduplicateRelationships(t *testing.T) {
	rels := []c4model.Relationship{
		{SourceID: "a", TargetID: "b", Description: "uses", Level: "container"},
		{SourceID: "a", TargetID: "b", Description: "uses (duplicate)", Level: "container"},
		{SourceID: "a", TargetID: "c", Description: "calls", Level: "component"},
		{SourceID: "a", TargetID: "b", Description: "calls", Level: "component"}, // different level, not a duplicate
	}

	deduped := c4model.DeduplicateRelationships(rels)

	assert.Len(t, deduped, 3)

	// Verify the unique combinations exist
	keys := make(map[string]bool)
	for _, r := range deduped {
		key := r.SourceID + "->" + r.TargetID + "@" + r.Level
		keys[key] = true
	}
	assert.True(t, keys["a->b@container"])
	assert.True(t, keys["a->c@component"])
	assert.True(t, keys["a->b@component"])
}
