package analyzer_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(filename), "testdata", "sample-project")
}

func TestAnalyzeProject_ModuleInfo(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	assert.Equal(t, "github.com/example/sample-project", result.Module.ModulePath)
	assert.Equal(t, "1.22", result.Module.GoVersion)
}

func TestAnalyzeProject_DirectDependencies(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	assert.Len(t, result.Module.DirectDeps, 2)

	depPaths := make([]string, len(result.Module.DirectDeps))
	for i, d := range result.Module.DirectDeps {
		depPaths[i] = d.Path
	}
	assert.Contains(t, depPaths, "github.com/go-chi/chi/v5")
	assert.Contains(t, depPaths, "github.com/lib/pq")
}

func TestAnalyzeProject_IndirectDependencies(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	assert.Len(t, result.Module.IndirectDeps, 1)
	assert.Equal(t, "github.com/stretchr/testify", result.Module.IndirectDeps[0].Path)
	assert.Equal(t, "v1.8.4", result.Module.IndirectDeps[0].Version)
}

func TestAnalyzeProject_PackageDiscovery(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	require.Len(t, result.Packages, 4, "should find 4 packages: main, handler, service, utils")

	pkgNames := make(map[string]bool)
	for _, p := range result.Packages {
		pkgNames[p.Name] = true
	}
	assert.True(t, pkgNames["main"])
	assert.True(t, pkgNames["handler"])
	assert.True(t, pkgNames["service"])
	assert.True(t, pkgNames["utils"])
}

func TestAnalyzeProject_PackageRoles(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	roleMap := make(map[string]string)
	for _, p := range result.Packages {
		roleMap[p.Name] = p.Role
	}
	assert.Equal(t, "cmd", roleMap["main"])
	assert.Equal(t, "internal", roleMap["handler"])
	assert.Equal(t, "internal", roleMap["service"])
	assert.Equal(t, "pkg", roleMap["utils"])
}

func TestAnalyzeProject_PackageImportPaths(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	importPathMap := make(map[string]string)
	for _, p := range result.Packages {
		importPathMap[p.Name] = p.ImportPath
	}
	assert.Equal(t, "github.com/example/sample-project/cmd/server", importPathMap["main"])
	assert.Equal(t, "github.com/example/sample-project/internal/handler", importPathMap["handler"])
	assert.Equal(t, "github.com/example/sample-project/internal/service", importPathMap["service"])
	assert.Equal(t, "github.com/example/sample-project/pkg/utils", importPathMap["utils"])
}

func TestAnalyzeProject_PackageGoFiles(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	for _, p := range result.Packages {
		assert.NotEmpty(t, p.GoFiles, "package %s should have Go files", p.Name)
	}
}

func TestAnalyzeProject_PackageImports(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	importsMap := make(map[string][]string)
	for _, p := range result.Packages {
		importsMap[p.Name] = p.Imports
	}

	// main imports handler and service
	assert.Contains(t, importsMap["main"], "github.com/example/sample-project/internal/handler")
	assert.Contains(t, importsMap["main"], "github.com/example/sample-project/internal/service")

	// handler imports service and utils
	assert.Contains(t, importsMap["handler"], "github.com/example/sample-project/internal/service")
	assert.Contains(t, importsMap["handler"], "github.com/example/sample-project/pkg/utils")

	// service imports utils
	assert.Contains(t, importsMap["service"], "github.com/example/sample-project/pkg/utils")
}

func TestAnalyzeProject_ImportGraph(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	require.NotNil(t, result.ImportGraph)

	// main -> handler, service
	mainImports := result.ImportGraph["github.com/example/sample-project/cmd/server"]
	assert.Contains(t, mainImports, "github.com/example/sample-project/internal/handler")
	assert.Contains(t, mainImports, "github.com/example/sample-project/internal/service")

	// handler -> service, utils
	handlerImports := result.ImportGraph["github.com/example/sample-project/internal/handler"]
	assert.Contains(t, handlerImports, "github.com/example/sample-project/internal/service")
	assert.Contains(t, handlerImports, "github.com/example/sample-project/pkg/utils")

	// service -> utils
	serviceImports := result.ImportGraph["github.com/example/sample-project/internal/service"]
	assert.Contains(t, serviceImports, "github.com/example/sample-project/pkg/utils")
}

func TestAnalyzeProject_NonExistentPath(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	_, err := a.AnalyzeProject(context.Background(), "/nonexistent/path")
	assert.Error(t, err)
}

func TestAnalyzeProject_NoGoMod(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	// Use a path that exists but has no go.mod
	_, err := a.AnalyzeProject(context.Background(), "/tmp")
	assert.Error(t, err)
}
