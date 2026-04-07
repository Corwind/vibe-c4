package c4model_test

import (
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func promptAnalysisResult() *analyzer.AnalysisResult {
	return &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/ex/app",
			GoVersion:  "1.22",
			DirectDeps: []analyzer.Dependency{
				{Path: "github.com/lib/pq", Version: "v1.10.0"},
			},
		},
		Packages: []analyzer.PackageInfo{
			{
				Name:       "handler",
				ImportPath: "github.com/ex/app/internal/handler",
				Dir:        "internal/handler",
				Role:       "internal",
				GoFiles:    []string{"handler.go"},
				Imports:    []string{"github.com/ex/app/internal/service"},
				Structs: []analyzer.StructInfo{
					{Name: "Handler", FilePath: "internal/handler/handler.go", Line: 5},
				},
			},
			{
				Name:       "service",
				ImportPath: "github.com/ex/app/internal/service",
				Dir:        "internal/service",
				Role:       "internal",
				GoFiles:    []string{"service.go"},
				Imports:    []string{"github.com/ex/app/pkg/utils"},
				Structs: []analyzer.StructInfo{
					{Name: "Service", FilePath: "internal/service/service.go", Line: 10},
				},
				Interfaces: []analyzer.InterfaceInfo{
					{Name: "Worker", FilePath: "internal/service/service.go", Line: 3},
				},
			},
			{
				Name:       "utils",
				ImportPath: "github.com/ex/app/pkg/utils",
				Dir:        "pkg/utils",
				Role:       "pkg",
				GoFiles:    []string{"logger.go"},
				Functions: []analyzer.FunctionInfo{
					{Name: "Log", Params: []string{"string"}, FilePath: "pkg/utils/logger.go", Line: 5},
				},
			},
		},
		ImportGraph: map[string][]string{
			"github.com/ex/app/internal/handler": {"github.com/ex/app/internal/service"},
			"github.com/ex/app/internal/service": {"github.com/ex/app/pkg/utils"},
			"github.com/ex/app/pkg/utils":        {},
		},
		CallGraph: []analyzer.FunctionCall{
			{CallerPkg: "github.com/ex/app/internal/handler", CallerFunc: "Handle", CalleePkg: "github.com/ex/app/internal/service", CalleeFunc: "DoWork"},
			{CallerPkg: "github.com/ex/app/internal/service", CallerFunc: "DoWork", CalleePkg: "github.com/ex/app/pkg/utils", CalleeFunc: "Log"},
			{CallerPkg: "github.com/ex/app/pkg/utils", CallerFunc: "Log", CalleePkg: "fmt", CalleeFunc: "Println"},
		},
		ExternalInteractions: []analyzer.ExternalInteraction{
			{Kind: analyzer.ExtKindHTTPHandler, PkgPath: "github.com/ex/app/internal/handler", FuncName: "Handle"},
			{Kind: analyzer.ExtKindDatabase, PkgPath: "github.com/ex/app/internal/service", FuncName: "Query"},
		},
		InterfaceImpls: []analyzer.InterfaceImpl{
			{StructPkg: "github.com/ex/app/internal/service", StructName: "Service", InterfacePkg: "github.com/ex/app/internal/service", InterfaceName: "Worker"},
			{StructPkg: "github.com/ex/app/internal/handler", StructName: "Handler", InterfacePkg: "net/http", InterfaceName: "Handler"},
		},
		Entrypoints: []analyzer.Entrypoint{
			{Kind: "http_handler", PkgPath: "github.com/ex/app/internal/handler", FuncName: "Handle", Route: "/api/v1"},
			{Kind: "main", PkgPath: "github.com/ex/app/pkg/utils", FuncName: "Init"},
		},
	}
}

func sampleRepoContents() map[string]string {
	return map[string]string{
		"internal/handler/handler.go":  "package handler\n\ntype Handler struct{}\n",
		"internal/handler/routes.go":   "package handler\n\nfunc Routes() {}\n",
		"internal/service/service.go":  "package service\n\ntype Service struct{}\n",
		"pkg/utils/logger.go":          "package utils\n\nfunc Log(msg string) {}\n",
		"go.mod":                       "module github.com/ex/app\n",
		"main.go":                      "package main\n\nfunc main() {}\n",
	}
}

// --- Test BuildContextPrompt ---

func TestBuildContextPrompt(t *testing.T) {
	prompt := c4model.BuildContextPrompt()

	// Should mention C4 levels 1 and 2
	assert.Contains(t, prompt, "System Context")
	assert.Contains(t, prompt, "Container")

	// Should contain JSON schema keys for systems and containers
	assert.Contains(t, prompt, "systems")
	assert.Contains(t, prompt, "containers")
	assert.Contains(t, prompt, "relationships")

	// Should NOT contain component-level or code-element schema references
	assert.NotContains(t, prompt, "code_elements")
	assert.NotContains(t, prompt, `"level": "component"`)
}

// --- Test BuildComponentPrompt ---

func TestBuildComponentPrompt(t *testing.T) {
	container := c4model.Container{
		ID:          "container-internal-handler",
		Name:        "internal/handler",
		Description: "HTTP handler package",
		Technology:  "Go",
		PackagePath: "github.com/ex/app/internal/handler",
		SystemID:    "system-app",
	}
	siblings := []c4model.Container{
		{
			ID:          "container-internal-service",
			Name:        "internal/service",
			PackagePath: "github.com/ex/app/internal/service",
		},
		{
			ID:          "container-pkg-utils",
			Name:        "pkg/utils",
			PackagePath: "github.com/ex/app/pkg/utils",
		},
	}

	prompt := c4model.BuildComponentPrompt(container, siblings)

	// Should mention the container being analyzed
	assert.Contains(t, prompt, "internal/handler")
	assert.Contains(t, prompt, "container-internal-handler")

	// Should list sibling container names
	assert.Contains(t, prompt, "internal/service")
	assert.Contains(t, prompt, "pkg/utils")

	// Should contain component schema keys
	assert.Contains(t, prompt, "components")
	assert.Contains(t, prompt, "relationships")
	assert.Contains(t, prompt, "container_id")

	// Should NOT contain systems or code_elements
	assert.NotContains(t, prompt, "code_elements")
}

// --- Test BuildCodePrompt ---

func TestBuildCodePrompt(t *testing.T) {
	component := c4model.Component{
		ID:          "component-handler-handler",
		Name:        "Handler",
		Type:        c4model.ComponentTypeStruct,
		ContainerID: "container-internal-handler",
		PackagePath: "github.com/ex/app/internal/handler",
		Methods:     []string{"ServeHTTP", "Handle"},
		Fields:      []string{"svc *service.Service", "logger *log.Logger"},
	}

	prompt := c4model.BuildCodePrompt(component)

	// Should mention the component being analyzed
	assert.Contains(t, prompt, "Handler")
	assert.Contains(t, prompt, "component-handler-handler")

	// Should list methods and fields
	assert.Contains(t, prompt, "ServeHTTP")
	assert.Contains(t, prompt, "Handle")
	assert.Contains(t, prompt, "svc *service.Service")
	assert.Contains(t, prompt, "logger *log.Logger")

	// Should contain code_elements schema
	assert.Contains(t, prompt, "code_elements")
	assert.Contains(t, prompt, "component_id")
}

// --- Test BuildContextUserPrompt ---

func TestBuildContextUserPrompt(t *testing.T) {
	result := promptAnalysisResult()
	repoContents := sampleRepoContents()

	prompt := c4model.BuildContextUserPrompt(result, repoContents)

	// Should contain analysis section
	assert.Contains(t, prompt, "Static Analysis Results")

	// Should contain source code section
	assert.Contains(t, prompt, "handler.go")

	// Should contain level instructions
	assert.Contains(t, prompt, "Level 1")
	assert.Contains(t, prompt, "Level 2")

	// Should instruct NOT to generate components
	assert.Contains(t, prompt, "Do NOT generate components")
}

// --- Test BuildComponentUserPrompt ---

func TestBuildComponentUserPrompt(t *testing.T) {
	result := promptAnalysisResult()
	container := c4model.Container{
		ID:   "container-internal-handler",
		Name: "internal/handler",
	}
	scopedFiles := map[string]string{
		"internal/handler/handler.go": "package handler\n\ntype Handler struct{}\n",
	}

	prompt := c4model.BuildComponentUserPrompt(result, container, scopedFiles)

	assert.Contains(t, prompt, "container-internal-handler")
	assert.Contains(t, prompt, "internal/handler")
}

// --- Test BuildCodeUserPrompt ---

func TestBuildCodeUserPrompt(t *testing.T) {
	result := promptAnalysisResult()
	component := c4model.Component{
		ID:          "component-handler",
		Name:        "Handler",
		ContainerID: "container-internal-handler",
	}
	scopedFiles := map[string]string{
		"internal/handler/handler.go": "package handler\n\ntype Handler struct{}\n",
	}

	prompt := c4model.BuildCodeUserPrompt(result, component, scopedFiles)

	assert.Contains(t, prompt, "component-handler")
	assert.Contains(t, prompt, "Handler")
}

// --- Test ScopeFilesToPackage ---

func TestScopeFilesToPackage(t *testing.T) {
	repoContents := sampleRepoContents()
	modulePath := "github.com/ex/app"

	t.Run("scopes to specific package directory", func(t *testing.T) {
		scoped := c4model.ScopeFilesToPackage(repoContents, "github.com/ex/app/internal/handler", modulePath)

		assert.Len(t, scoped, 2)
		assert.Contains(t, scoped, "internal/handler/handler.go")
		assert.Contains(t, scoped, "internal/handler/routes.go")
		assert.NotContains(t, scoped, "internal/service/service.go")
	})

	t.Run("root package returns all files", func(t *testing.T) {
		scoped := c4model.ScopeFilesToPackage(repoContents, modulePath, modulePath)

		assert.Len(t, scoped, len(repoContents))
	})

	t.Run("non-matching package returns empty map", func(t *testing.T) {
		scoped := c4model.ScopeFilesToPackage(repoContents, "github.com/other/pkg", modulePath)

		assert.Empty(t, scoped)
	})
}

// --- Test ScopeAnalysis ---

func TestScopeAnalysis(t *testing.T) {
	result := promptAnalysisResult()

	scoped := c4model.ScopeAnalysis(result, []string{"github.com/ex/app/internal/handler"})

	// Module and ProjectPath should be copied as-is
	assert.Equal(t, result.Module, scoped.Module)

	// Should only have the handler package
	require.Len(t, scoped.Packages, 1)
	assert.Equal(t, "github.com/ex/app/internal/handler", scoped.Packages[0].ImportPath)

	// ImportGraph should only have handler entry
	require.Len(t, scoped.ImportGraph, 1)
	assert.Contains(t, scoped.ImportGraph, "github.com/ex/app/internal/handler")

	// CallGraph should only have handler calls
	require.Len(t, scoped.CallGraph, 1)
	assert.Equal(t, "github.com/ex/app/internal/handler", scoped.CallGraph[0].CallerPkg)

	// ExternalInteractions should only have handler interactions
	require.Len(t, scoped.ExternalInteractions, 1)
	assert.Equal(t, "github.com/ex/app/internal/handler", scoped.ExternalInteractions[0].PkgPath)

	// Entrypoints should only have handler entrypoints
	require.Len(t, scoped.Entrypoints, 1)
	assert.Equal(t, "github.com/ex/app/internal/handler", scoped.Entrypoints[0].PkgPath)

	// InterfaceImpls should only have handler impls
	require.Len(t, scoped.InterfaceImpls, 1)
	assert.Equal(t, "github.com/ex/app/internal/handler", scoped.InterfaceImpls[0].StructPkg)
}
