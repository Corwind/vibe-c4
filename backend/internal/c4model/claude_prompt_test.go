package c4model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSystemPrompt(t *testing.T) {
	prompt := BuildSystemPrompt()

	// Should contain C4 level descriptions
	assert.Contains(t, prompt, "Level 1 - System Context")
	assert.Contains(t, prompt, "Level 2 - Container")
	assert.Contains(t, prompt, "Level 3 - Component")
	assert.Contains(t, prompt, "Level 4 - Code")

	// Should contain JSON schema references
	assert.Contains(t, prompt, "systems")
	assert.Contains(t, prompt, "containers")
	assert.Contains(t, prompt, "components")
	assert.Contains(t, prompt, "code_elements")
	assert.Contains(t, prompt, "relationships")

	// Should contain ID generation rules
	assert.Contains(t, prompt, "system-")
	assert.Contains(t, prompt, "container-")
	assert.Contains(t, prompt, "component-")
	assert.Contains(t, prompt, "code-")
	assert.Contains(t, prompt, "sanitized")

	// Should instruct to return only valid JSON
	assert.Contains(t, prompt, "ONLY valid JSON")

	// Should contain relationship rules
	assert.Contains(t, prompt, "source_id")
	assert.Contains(t, prompt, "target_id")

	// Should contain quality guidelines
	assert.Contains(t, prompt, "PURPOSE")
	assert.Contains(t, prompt, "RESPONSIBILITY")
}

func TestBuildUserPrompt(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/example/myapp",
			GoVersion:  "1.21",
		},
		Packages: []analyzer.PackageInfo{
			{
				Name:       "main",
				ImportPath: "github.com/example/myapp",
				GoFiles:    []string{"main.go"},
			},
		},
	}

	repoContents := map[string]string{
		"main.go":             "package main\n\nfunc main() {}\n",
		"internal/handler.go": "package internal\n",
	}

	prompt := BuildUserPrompt(result, repoContents)

	// Section 1: Static Analysis Results
	assert.Contains(t, prompt, "## Static Analysis Results")
	assert.Contains(t, prompt, "github.com/example/myapp")

	// Section 2: Source Code
	assert.Contains(t, prompt, "## Source Code")
	assert.Contains(t, prompt, "### main.go")
	assert.Contains(t, prompt, "```go")
	assert.Contains(t, prompt, "func main()")
	assert.Contains(t, prompt, "### internal/handler.go")

	// Section 3: Instructions
	assert.Contains(t, prompt, "## Instructions")
	assert.Contains(t, prompt, "generate a complete C4 model JSON")
	assert.Contains(t, prompt, "all 4 levels")
}

func TestReadFullRepository(t *testing.T) {
	// Create a temp directory with various files
	tmpDir := t.TempDir()

	// Create go.mod
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0644))

	// Create main.go
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644))

	// Create a subdirectory with a Go file
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "internal", "service.go"), []byte("package internal\n\ntype Service struct{}\n"), 0644))

	// Create SQL file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "schema.sql"), []byte("CREATE TABLE users (id INT);"), 0644))

	// Create YAML config
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("port: 8080"), 0644))

	// Create Markdown doc
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Project"), 0644))

	// Large budget: should read all files
	result, err := ReadFullRepository(tmpDir, 100000)
	require.NoError(t, err)

	assert.Contains(t, result, "go.mod")
	assert.Contains(t, result, "main.go")
	assert.Contains(t, result, filepath.Join("internal", "service.go"))
	assert.Contains(t, result, "schema.sql")
	assert.Contains(t, result, "config.yaml")
	assert.Contains(t, result, "README.md")
	assert.Len(t, result, 6)

	// Very small budget: should prioritize tier 1 files (go.mod, main.go)
	result, err = ReadFullRepository(tmpDir, 20)
	require.NoError(t, err)

	// With a budget of 20 tokens (~80 chars), go.mod and main.go (tier 1) should be included
	// but not all files
	assert.Contains(t, result, "go.mod")
	assert.Contains(t, result, "main.go")
	// The budget may or may not fit the other files
	assert.LessOrEqual(t, len(result), 6)
}

func TestReadFullRepositoryExcludesVendorAndTests(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0644))

	// Create a normal Go file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644))

	// Create vendor directory with a Go file
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "vendor", "lib"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "vendor", "lib", "lib.go"), []byte("package lib\n"), 0644))

	// Create a test file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte("package main\n\nimport \"testing\"\n"), 0644))

	// Create testdata directory
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "testdata"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "testdata", "fixture.go"), []byte("package testdata\n"), 0644))

	// Create go.sum (should be excluded)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.sum"), []byte("hash123"), 0644))

	// Create image file (should be excluded)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "logo.png"), []byte("PNG"), 0644))

	// Create node_modules dir (should be excluded)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "node_modules", "lib"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "node_modules", "lib", "index.js"), []byte("module.exports = {}"), 0644))

	// Create package-lock.json (should be excluded)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte("{}"), 0644))

	result, err := ReadFullRepository(tmpDir, 100000)
	require.NoError(t, err)

	// Should include normal files
	assert.Contains(t, result, "go.mod")
	assert.Contains(t, result, "main.go")

	// Should exclude vendor, test files, testdata, binary files, lock files, and node_modules
	for path := range result {
		assert.False(t, strings.HasPrefix(path, "vendor"), "should not include vendor files: %s", path)
		assert.False(t, strings.HasSuffix(path, "_test.go"), "should not include test files: %s", path)
		assert.False(t, strings.HasPrefix(path, "testdata"), "should not include testdata files: %s", path)
		assert.False(t, strings.HasPrefix(path, "node_modules"), "should not include node_modules files: %s", path)
		assert.False(t, strings.HasSuffix(path, ".png"), "should not include image files: %s", path)
		assert.NotEqual(t, "go.sum", path, "should not include go.sum")
		assert.NotEqual(t, "package-lock.json", path, "should not include package-lock.json")
	}
}

func TestClassifyFileTier(t *testing.T) {
	tests := []struct {
		path    string
		content string
		tier    int
	}{
		{"main.go", "package main", 1},
		{"Dockerfile", "FROM golang:1.24", 1},
		{"Makefile", "build:", 1},
		{"docker-compose.yml", "services:", 1},
		{"internal/types.go", "type Foo interface {}", 1},
		{"schema.sql", "CREATE TABLE", 2},
		{"api.proto", "service API {}", 2},
		{"internal/handler/handler.go", "package handler", 2},
		{"config.yaml", "port: 8080", 3},
		{"README.md", "# Project", 3},
		{"internal/model/user.go", "package model", 3},
		{"utils/helper.go", "package utils", 4},
		{"scripts/deploy.sh", "#!/bin/bash", 4},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.tier, classifyFileTier(tt.path, tt.content))
		})
	}
}

func TestCodeFenceLanguage(t *testing.T) {
	assert.Equal(t, "go", codeFenceLanguage("main.go"))
	assert.Equal(t, "sql", codeFenceLanguage("schema.sql"))
	assert.Equal(t, "yaml", codeFenceLanguage("config.yaml"))
	assert.Equal(t, "dockerfile", codeFenceLanguage("Dockerfile"))
	assert.Equal(t, "makefile", codeFenceLanguage("Makefile"))
	assert.Equal(t, "protobuf", codeFenceLanguage("api.proto"))
	assert.Equal(t, "", codeFenceLanguage("data.txt"))
}

func TestBuildContextPrompt(t *testing.T) {
	prompt := BuildContextPrompt()

	assert.Contains(t, prompt, "System Context")
	assert.Contains(t, prompt, "Container")
	assert.Contains(t, prompt, "systems")
	assert.Contains(t, prompt, "containers")
	assert.Contains(t, prompt, "relationships")
	assert.NotContains(t, prompt, "code_elements")
	assert.NotContains(t, prompt, `"level": "component"`)
}

func TestBuildComponentPrompt(t *testing.T) {
	container := Container{
		ID:          "container-internal-handler",
		Name:        "internal/handler",
		Description: "HTTP handler package",
		Technology:  "Go",
		PackagePath: "github.com/ex/app/internal/handler",
		SystemID:    "system-app",
	}
	siblings := []Container{
		{ID: "container-internal-service", Name: "internal/service"},
		{ID: "container-pkg-utils", Name: "pkg/utils"},
	}

	prompt := BuildComponentPrompt(container, siblings)

	assert.Contains(t, prompt, "internal/handler")
	assert.Contains(t, prompt, "container-internal-handler")
	assert.Contains(t, prompt, "internal/service")
	assert.Contains(t, prompt, "pkg/utils")
	assert.Contains(t, prompt, "components")
	assert.Contains(t, prompt, "container_id")
	assert.NotContains(t, prompt, "code_elements")
}

func TestBuildCodePrompt(t *testing.T) {
	component := Component{
		ID:          "component-handler-handler",
		Name:        "Handler",
		Type:        ComponentTypeStruct,
		ContainerID: "container-internal-handler",
		PackagePath: "github.com/ex/app/internal/handler",
		Methods:     []string{"ServeHTTP", "Handle"},
		Fields:      []string{"svc *service.Service", "logger *log.Logger"},
	}

	prompt := BuildCodePrompt(component)

	assert.Contains(t, prompt, "Handler")
	assert.Contains(t, prompt, "component-handler-handler")
	assert.Contains(t, prompt, "ServeHTTP")
	assert.Contains(t, prompt, "Handle")
	assert.Contains(t, prompt, "svc *service.Service")
	assert.Contains(t, prompt, "code_elements")
	assert.Contains(t, prompt, "component_id")
}

func TestBuildContextUserPrompt(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/example/myapp",
			GoVersion:  "1.21",
		},
		Packages: []analyzer.PackageInfo{
			{Name: "main", ImportPath: "github.com/example/myapp", GoFiles: []string{"main.go"}},
		},
	}
	repoContents := map[string]string{
		"main.go": "package main\n\nfunc main() {}\n",
	}

	prompt := BuildContextUserPrompt(result, repoContents)

	assert.Contains(t, prompt, "Static Analysis Results")
	assert.Contains(t, prompt, "main.go")
	assert.Contains(t, prompt, "Do NOT generate components")
}

func TestScopeFilesToPackage(t *testing.T) {
	repoContents := map[string]string{
		"internal/handler/handler.go": "package handler",
		"internal/handler/routes.go":  "package handler",
		"internal/service/service.go": "package service",
		"pkg/utils/logger.go":         "package utils",
		"go.mod":                      "module github.com/ex/app",
		"main.go":                     "package main",
	}
	modulePath := "github.com/ex/app"

	t.Run("scopes to specific package directory", func(t *testing.T) {
		scoped := ScopeFilesToPackage(repoContents, "github.com/ex/app/internal/handler", modulePath)
		assert.Len(t, scoped, 2)
		assert.Contains(t, scoped, "internal/handler/handler.go")
		assert.Contains(t, scoped, "internal/handler/routes.go")
		assert.NotContains(t, scoped, "internal/service/service.go")
	})

	t.Run("root package returns all files", func(t *testing.T) {
		scoped := ScopeFilesToPackage(repoContents, modulePath, modulePath)
		assert.Len(t, scoped, len(repoContents))
	})

	t.Run("non-matching package returns empty map", func(t *testing.T) {
		scoped := ScopeFilesToPackage(repoContents, "github.com/other/pkg", modulePath)
		assert.Empty(t, scoped)
	})
}

func TestScopeAnalysis(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/ex/app"},
		Packages: []analyzer.PackageInfo{
			{Name: "handler", ImportPath: "github.com/ex/app/internal/handler"},
			{Name: "service", ImportPath: "github.com/ex/app/internal/service"},
			{Name: "utils", ImportPath: "github.com/ex/app/pkg/utils"},
		},
		ImportGraph: map[string][]string{
			"github.com/ex/app/internal/handler": {"github.com/ex/app/internal/service"},
			"github.com/ex/app/internal/service": {"github.com/ex/app/pkg/utils"},
		},
		CallGraph: []analyzer.FunctionCall{
			{CallerPkg: "github.com/ex/app/internal/handler", CallerFunc: "Handle", CalleePkg: "github.com/ex/app/internal/service"},
			{CallerPkg: "github.com/ex/app/internal/service", CallerFunc: "DoWork", CalleePkg: "github.com/ex/app/pkg/utils"},
		},
		ExternalInteractions: []analyzer.ExternalInteraction{
			{Kind: analyzer.ExtKindHTTPHandler, PkgPath: "github.com/ex/app/internal/handler"},
			{Kind: analyzer.ExtKindDatabase, PkgPath: "github.com/ex/app/internal/service"},
		},
		InterfaceImpls: []analyzer.InterfaceImpl{
			{StructPkg: "github.com/ex/app/internal/handler", StructName: "Handler"},
		},
		Entrypoints: []analyzer.Entrypoint{
			{Kind: "http_handler", PkgPath: "github.com/ex/app/internal/handler"},
		},
	}

	scoped := ScopeAnalysis(result, []string{"github.com/ex/app/internal/handler"})

	assert.Equal(t, result.Module, scoped.Module)
	require.Len(t, scoped.Packages, 1)
	assert.Equal(t, "github.com/ex/app/internal/handler", scoped.Packages[0].ImportPath)
	require.Len(t, scoped.ImportGraph, 1)
	assert.Contains(t, scoped.ImportGraph, "github.com/ex/app/internal/handler")
	require.Len(t, scoped.CallGraph, 1)
	assert.Equal(t, "github.com/ex/app/internal/handler", scoped.CallGraph[0].CallerPkg)
	require.Len(t, scoped.ExternalInteractions, 1)
	require.Len(t, scoped.Entrypoints, 1)
	require.Len(t, scoped.InterfaceImpls, 1)
}

func TestEstimateTokens(t *testing.T) {
	assert.Equal(t, 0, EstimateTokens(""))
	assert.Equal(t, 1, EstimateTokens("abc"))
	assert.Equal(t, 2, EstimateTokens("abcdef"))
	assert.Equal(t, 33, EstimateTokens(strings.Repeat("a", 100)))
}
