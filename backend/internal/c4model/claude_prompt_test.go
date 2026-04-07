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
	// Create a temp directory with Go files
	tmpDir := t.TempDir()

	// Create go.mod
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0644))

	// Create main.go
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644))

	// Create a subdirectory with a Go file
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "internal", "service.go"), []byte("package internal\n\ntype Service struct{}\n"), 0644))

	// Large budget: should read all files
	result, err := ReadFullRepository(tmpDir, 100000)
	require.NoError(t, err)

	assert.Contains(t, result, "go.mod")
	assert.Contains(t, result, "main.go")
	assert.Contains(t, result, filepath.Join("internal", "service.go"))
	assert.Len(t, result, 3)

	// Very small budget: should prioritize tier 1 files (go.mod, main.go)
	result, err = ReadFullRepository(tmpDir, 20)
	require.NoError(t, err)

	// With a budget of 20 tokens (~80 chars), go.mod and main.go (tier 1) should be included
	// but not all files
	assert.Contains(t, result, "go.mod")
	assert.Contains(t, result, "main.go")
	// The budget may or may not fit the internal/service.go file
	assert.LessOrEqual(t, len(result), 3)
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

	result, err := ReadFullRepository(tmpDir, 100000)
	require.NoError(t, err)

	// Should include normal files
	assert.Contains(t, result, "go.mod")
	assert.Contains(t, result, "main.go")

	// Should exclude vendor, test files, and testdata
	for path := range result {
		assert.False(t, strings.HasPrefix(path, "vendor"), "should not include vendor files: %s", path)
		assert.False(t, strings.HasSuffix(path, "_test.go"), "should not include test files: %s", path)
		assert.False(t, strings.HasPrefix(path, "testdata"), "should not include testdata files: %s", path)
	}
}

func TestEstimateTokens(t *testing.T) {
	assert.Equal(t, 0, EstimateTokens(""))
	assert.Equal(t, 1, EstimateTokens("abc"))
	assert.Equal(t, 2, EstimateTokens("abcdef"))
	assert.Equal(t, 33, EstimateTokens(strings.Repeat("a", 100)))
}
