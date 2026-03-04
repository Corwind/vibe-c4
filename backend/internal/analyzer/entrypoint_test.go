package analyzer_test

import (
	"context"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeProject_DetectsMainEntrypoint(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	var mainEntrypoints []analyzer.Entrypoint
	for _, ep := range result.Entrypoints {
		if ep.Kind == "main" {
			mainEntrypoints = append(mainEntrypoints, ep)
		}
	}

	require.Len(t, mainEntrypoints, 1, "should detect exactly one main entrypoint")
	assert.Equal(t, "github.com/example/sample-project/cmd/server", mainEntrypoints[0].PkgPath)
	assert.Equal(t, "main", mainEntrypoints[0].FuncName)
}

func TestAnalyzeProject_RouteEntrypointsFromLocalVar_NotDetected(t *testing.T) {
	// Chi route calls like r.Get("/health", ...) where r is a local variable
	// are not currently detected as entrypoints because the body visitor only
	// triggers external pattern detection for package-level import calls.
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	var routeEntrypoints []analyzer.Entrypoint
	for _, ep := range result.Entrypoints {
		if ep.Kind == "http_route" {
			routeEntrypoints = append(routeEntrypoints, ep)
		}
	}

	// This documents the current behavior: local variable chi calls are not detected
	assert.Empty(t, routeEntrypoints, "chi routes via local variable are not currently detected as entrypoints")
}

func TestAnalyzeProject_EntrypointsNotEmpty(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	assert.NotEmpty(t, result.Entrypoints, "should detect at least one entrypoint")
}

func TestAnalyzeProject_MainEntrypointHasFilePath(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	for _, ep := range result.Entrypoints {
		if ep.Kind == "main" {
			assert.NotEmpty(t, ep.FilePath, "main entrypoint should have a file path")
			assert.Greater(t, ep.Line, 0, "main entrypoint should have a line number")
		}
	}
}
