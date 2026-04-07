package c4model

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
	case strings.Contains(sp, "system context") && strings.Contains(sp, "level 1"):
		if m.pass1Err != nil {
			return nil, m.pass1Err
		}
		return m.pass1Response, nil
	case strings.Contains(sp, "component") && strings.Contains(sp, "level 3"):
		if m.pass2Err != nil {
			return nil, m.pass2Err
		}
		return m.pass2Response, nil
	case strings.Contains(sp, "code") && strings.Contains(sp, "level 4"):
		if m.pass3Err != nil {
			return nil, m.pass3Err
		}
		return m.pass3Response, nil
	default:
		return nil, fmt.Errorf("unexpected system prompt: %.100s", req.SystemPrompt)
	}
}

func multipassTestConfig() config.ClaudeConfig {
	return config.ClaudeConfig{
		APIKey:         "test-key",
		Model:          "claude-sonnet-4-6",
		MaxTokens:      4096,
		TokenBudget:    80000,
		TimeoutSecs:    30,
		MaxConcurrency: 2,
	}
}

func makePass1JSON(systems []System, containers []Container, rels []Relationship) string {
	result := Pass1Result{Systems: systems, Containers: containers, Relationships: rels}
	data, _ := json.Marshal(result)
	return string(data)
}

func makePass2JSON(components []Component, rels []Relationship) string {
	result := Pass2Result{Components: components, Relationships: rels}
	data, _ := json.Marshal(result)
	return string(data)
}

func makePass3JSON(codeElements []CodeElement) string {
	result := Pass3Result{CodeElements: codeElements}
	data, _ := json.Marshal(result)
	return string(data)
}

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
		projectDir, modulePath,
		[]string{modulePath + "/api", modulePath + "/db"},
	)

	systems := []System{{ID: "system-main", Name: "test", Description: "Main system", ModulePath: modulePath}}
	containers := []Container{
		{ID: "container-api", Name: "API", PackagePath: modulePath + "/api", SystemID: "system-main", Technology: "Go"},
		{ID: "container-db", Name: "DB", PackagePath: modulePath + "/db", SystemID: "system-main", Technology: "Go"},
	}
	pass1Rels := []Relationship{{SourceID: "container-api", TargetID: "container-db", Description: "uses", Level: "container"}}

	apiComponents := []Component{
		{ID: "component-handler", Name: "Handler", ContainerID: "container-api", PackagePath: modulePath + "/api", Technology: "Go"},
		{ID: "component-service", Name: "Service", ContainerID: "container-api", PackagePath: modulePath + "/api", Technology: "Go"},
	}
	pass2Rels := []Relationship{{SourceID: "component-handler", TargetID: "component-service", Description: "calls", Level: "component"}}

	codeElements := []CodeElement{{ID: "code-handler-get", Name: "Get", ComponentID: "component-handler", Type: "method"}}

	mock := &pipelineMock{
		pass1Response: &llm.Response{Content: makePass1JSON(systems, containers, pass1Rels), Usage: llm.Usage{InputTokens: 1000, OutputTokens: 500}},
		pass2Response: &llm.Response{Content: makePass2JSON(apiComponents, pass2Rels), Usage: llm.Usage{InputTokens: 500, OutputTokens: 250}},
		pass3Response: &llm.Response{Content: makePass3JSON(codeElements), Usage: llm.Usage{InputTokens: 200, OutputTokens: 100}},
	}

	builder := NewMultiPassBuilder(mock, multipassTestConfig(), NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(result)

	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Len(t, model.Systems, 1)
	assert.Equal(t, "system-main", model.Systems[0].ID)
	assert.Len(t, model.Containers, 2)

	// Mock returns same pass2Response for both containers → 4 components total
	assert.GreaterOrEqual(t, len(model.Components), 4)
	assert.GreaterOrEqual(t, len(model.CodeElements), 1)

	// Verify parent-child links
	sys := model.FindSystem("system-main")
	require.NotNil(t, sys)
	assert.Contains(t, sys.ContainerIDs, "container-api")
	assert.Contains(t, sys.ContainerIDs, "container-db")

	apiContainer := model.FindContainer("container-api")
	require.NotNil(t, apiContainer)
	assert.NotEmpty(t, apiContainer.ComponentIDs)

	assert.NotEmpty(t, model.Relationships)
}

func TestMultiPassBuilderPass1FailureFallback(t *testing.T) {
	modulePath := "github.com/example/test"
	projectDir := setupTestRepo(t, modulePath, []string{"api"})

	result := multipassAnalysisResult(projectDir, modulePath, []string{modulePath + "/api"})

	mock := &pipelineMock{pass1Err: fmt.Errorf("API rate limited")}

	builder := NewMultiPassBuilder(mock, multipassTestConfig(), NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(result)

	require.NoError(t, err)
	require.NotNil(t, model)
	assert.NotEmpty(t, model.Systems, "fallback should produce systems")
}

func TestMultiPassBuilderPass2PartialFailure(t *testing.T) {
	modulePath := "github.com/example/test"
	projectDir := setupTestRepo(t, modulePath, []string{"api", "db"})

	result := multipassAnalysisResult(projectDir, modulePath, []string{modulePath + "/api", modulePath + "/db"})

	systems := []System{{ID: "system-main", Name: "test", Description: "Main system", ModulePath: modulePath}}
	containers := []Container{
		{ID: "container-api", Name: "API", PackagePath: modulePath + "/api", SystemID: "system-main"},
		{ID: "container-db", Name: "DB", PackagePath: modulePath + "/db", SystemID: "system-main"},
	}

	mock := &pipelineMock{
		pass1Response: &llm.Response{Content: makePass1JSON(systems, containers, nil), Usage: llm.Usage{InputTokens: 1000, OutputTokens: 500}},
		pass2Err:      fmt.Errorf("pass2 failure"),
		pass3Err:      fmt.Errorf("pass3 failure"),
	}

	builder := NewMultiPassBuilder(mock, multipassTestConfig(), NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(result)

	require.NoError(t, err)
	require.NotNil(t, model)
	assert.Len(t, model.Systems, 1)
	assert.Len(t, model.Containers, 2)
	assert.Empty(t, model.Components)
}

func TestMultiPassBuilderNilResult(t *testing.T) {
	mock := &pipelineMock{}
	builder := NewMultiPassBuilder(mock, multipassTestConfig(), NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(nil)

	require.Error(t, err)
	assert.Nil(t, model)
}

func TestMultiPassBuilderNoContainers(t *testing.T) {
	modulePath := "github.com/example/test"
	projectDir := setupTestRepo(t, modulePath, nil)

	result := multipassAnalysisResult(projectDir, modulePath, nil)

	systems := []System{{ID: "system-main", Name: "test", Description: "Standalone system", ModulePath: modulePath}}

	mock := &pipelineMock{
		pass1Response: &llm.Response{Content: makePass1JSON(systems, nil, nil), Usage: llm.Usage{InputTokens: 500, OutputTokens: 200}},
	}

	builder := NewMultiPassBuilder(mock, multipassTestConfig(), NewSmartModelBuilder())
	model, err := builder.BuildFromAnalysis(result)

	require.NoError(t, err)
	require.NotNil(t, model)
	assert.Len(t, model.Systems, 1)
	assert.Empty(t, model.Containers)
	assert.Empty(t, model.Components)

	mock.mu.Lock()
	defer mock.mu.Unlock()
	assert.Len(t, mock.calls, 1, "only pass1 call should have been made")
}

func TestMergeResultsRebuildIDs(t *testing.T) {
	pass1 := &Pass1Result{
		Systems:    []System{{ID: "sys-1", Name: "System"}},
		Containers: []Container{{ID: "cont-1", Name: "C1", SystemID: "sys-1"}, {ID: "cont-2", Name: "C2", SystemID: "sys-1"}},
	}
	pass2s := []Pass2Result{
		{Components: []Component{{ID: "comp-a", Name: "A", ContainerID: "cont-1"}, {ID: "comp-b", Name: "B", ContainerID: "cont-1"}}},
		{Components: []Component{{ID: "comp-c", Name: "C", ContainerID: "cont-2"}}},
	}
	pass3s := []Pass3Result{
		{CodeElements: []CodeElement{{ID: "code-1", Name: "F1", ComponentID: "comp-a"}, {ID: "code-2", Name: "F2", ComponentID: "comp-a"}}},
		{CodeElements: []CodeElement{{ID: "code-3", Name: "F3", ComponentID: "comp-c"}}},
	}

	model := MergeResults(pass1, pass2s, pass3s)
	require.NotNil(t, model)

	sys := model.FindSystem("sys-1")
	require.NotNil(t, sys)
	assert.ElementsMatch(t, []string{"cont-1", "cont-2"}, sys.ContainerIDs)

	cont1 := model.FindContainer("cont-1")
	require.NotNil(t, cont1)
	assert.ElementsMatch(t, []string{"comp-a", "comp-b"}, cont1.ComponentIDs)

	cont2 := model.FindContainer("cont-2")
	require.NotNil(t, cont2)
	assert.ElementsMatch(t, []string{"comp-c"}, cont2.ComponentIDs)

	compA := model.FindComponent("comp-a")
	require.NotNil(t, compA)
	assert.ElementsMatch(t, []string{"code-1", "code-2"}, compA.CodeElements)

	compC := model.FindComponent("comp-c")
	require.NotNil(t, compC)
	assert.ElementsMatch(t, []string{"code-3"}, compC.CodeElements)
}

func TestDeduplicateRelationships(t *testing.T) {
	rels := []Relationship{
		{SourceID: "a", TargetID: "b", Description: "uses", Level: "container"},
		{SourceID: "a", TargetID: "b", Description: "uses (dup)", Level: "container"},
		{SourceID: "a", TargetID: "c", Description: "calls", Level: "component"},
		{SourceID: "a", TargetID: "b", Description: "calls", Level: "component"}, // different level
	}

	deduped := DeduplicateRelationships(rels)

	assert.Len(t, deduped, 3)
	keys := make(map[string]bool)
	for _, r := range deduped {
		keys[r.SourceID+"->"+r.TargetID+"@"+r.Level] = true
	}
	assert.True(t, keys["a->b@container"])
	assert.True(t, keys["a->c@component"])
	assert.True(t, keys["a->b@component"])
}
