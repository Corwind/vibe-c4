package c4model_test

import (
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleAnalysisResult() *analyzer.AnalysisResult {
	return &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/example/sample-project",
			GoVersion:  "1.22",
			DirectDeps: []analyzer.Dependency{
				{Path: "github.com/go-chi/chi/v5", Version: "v5.0.10"},
				{Path: "github.com/lib/pq", Version: "v1.10.9"},
			},
			IndirectDeps: []analyzer.Dependency{
				{Path: "github.com/stretchr/testify", Version: "v1.8.4"},
			},
		},
		Packages: []analyzer.PackageInfo{
			{
				Name:       "main",
				ImportPath: "github.com/example/sample-project/cmd/server",
				Dir:        "cmd/server",
				Role:       "cmd",
				GoFiles:    []string{"main.go"},
				Imports:    []string{"github.com/example/sample-project/internal/handler", "github.com/example/sample-project/internal/service"},
			},
			{
				Name:       "handler",
				ImportPath: "github.com/example/sample-project/internal/handler",
				Dir:        "internal/handler",
				Role:       "internal",
				GoFiles:    []string{"handler.go"},
				Imports:    []string{"github.com/example/sample-project/internal/service", "github.com/example/sample-project/pkg/utils"},
				Structs: []analyzer.StructInfo{
					{
						Name:     "Handler",
						FilePath: "internal/handler/handler.go",
						Line:     8,
						Fields: []analyzer.FieldInfo{
							{Name: "svc", Type: "*service.Service"},
						},
						Methods: []analyzer.MethodInfo{
							{Name: "ServeHTTP", FilePath: "internal/handler/handler.go", Line: 16},
						},
					},
				},
				Functions: []analyzer.FunctionInfo{
					{Name: "New", Returns: []string{"*Handler"}, FilePath: "internal/handler/handler.go", Line: 12},
				},
			},
			{
				Name:       "service",
				ImportPath: "github.com/example/sample-project/internal/service",
				Dir:        "internal/service",
				Role:       "internal",
				GoFiles:    []string{"service.go"},
				Imports:    []string{"github.com/example/sample-project/pkg/utils"},
				Structs: []analyzer.StructInfo{
					{
						Name:     "BaseService",
						FilePath: "internal/service/service.go",
						Line:     13,
						Fields:   []analyzer.FieldInfo{{Name: "Name", Type: "string"}},
					},
					{
						Name:     "Service",
						FilePath: "internal/service/service.go",
						Line:     18,
						Fields: []analyzer.FieldInfo{
							{Name: "BaseService", Type: "BaseService", Embedded: true},
							{Name: "running", Type: "bool"},
						},
						Methods: []analyzer.MethodInfo{
							{Name: "DoWork", Returns: []string{"string"}, FilePath: "internal/service/service.go", Line: 28},
							{Name: "Stop", Returns: []string{"error"}, FilePath: "internal/service/service.go", Line: 33},
						},
					},
				},
				Interfaces: []analyzer.InterfaceInfo{
					{
						Name:     "Worker",
						FilePath: "internal/service/service.go",
						Line:     6,
						Methods: []analyzer.MethodInfo{
							{Name: "DoWork", Returns: []string{"string"}},
							{Name: "Stop", Returns: []string{"error"}},
						},
					},
				},
				Functions: []analyzer.FunctionInfo{
					{Name: "New", Returns: []string{"*Service"}, FilePath: "internal/service/service.go", Line: 23},
					{Name: "HelperFunc", Params: []string{"string"}, Returns: []string{"string"}, FilePath: "internal/service/service.go", Line: 38},
				},
			},
			{
				Name:       "utils",
				ImportPath: "github.com/example/sample-project/pkg/utils",
				Dir:        "pkg/utils",
				Role:       "pkg",
				GoFiles:    []string{"logger.go"},
				Functions: []analyzer.FunctionInfo{
					{Name: "Log", Params: []string{"string"}, FilePath: "pkg/utils/logger.go", Line: 5},
				},
			},
		},
		ImportGraph: map[string][]string{
			"github.com/example/sample-project/cmd/server":       {"github.com/example/sample-project/internal/handler", "github.com/example/sample-project/internal/service"},
			"github.com/example/sample-project/internal/handler": {"github.com/example/sample-project/internal/service", "github.com/example/sample-project/pkg/utils"},
			"github.com/example/sample-project/internal/service": {"github.com/example/sample-project/pkg/utils"},
		},
	}
}

func TestBuilder_CreatesMainSystem(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(sampleAnalysisResult())
	require.NoError(t, err)

	// Should have the main system
	var mainSystem *c4model.System
	for i, s := range model.Systems {
		if !s.External {
			mainSystem = &model.Systems[i]
			break
		}
	}
	require.NotNil(t, mainSystem, "should have a non-external main system")
	assert.Equal(t, "sample-project", mainSystem.Name)
	assert.Equal(t, "github.com/example/sample-project", mainSystem.ModulePath)
}

func TestBuilder_CreatesExternalSystems(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(sampleAnalysisResult())
	require.NoError(t, err)

	var externalSystems []c4model.System
	for _, s := range model.Systems {
		if s.External {
			externalSystems = append(externalSystems, s)
		}
	}
	// Should have external systems for direct deps (chi, pq)
	assert.Len(t, externalSystems, 2)

	names := make(map[string]bool)
	for _, s := range externalSystems {
		names[s.Name] = true
	}
	assert.True(t, names["chi/v5"])
	assert.True(t, names["pq"])
}

func TestBuilder_CreatesContainers(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(sampleAnalysisResult())
	require.NoError(t, err)

	assert.Len(t, model.Containers, 4, "should create container for each package")

	containerNames := make(map[string]bool)
	for _, c := range model.Containers {
		containerNames[c.Name] = true
	}
	assert.True(t, containerNames["cmd/server"])
	assert.True(t, containerNames["internal/handler"])
	assert.True(t, containerNames["internal/service"])
	assert.True(t, containerNames["pkg/utils"])
}

func TestBuilder_ContainersBelongToMainSystem(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(sampleAnalysisResult())
	require.NoError(t, err)

	var mainSystemID string
	for _, s := range model.Systems {
		if !s.External {
			mainSystemID = s.ID
			break
		}
	}
	require.NotEmpty(t, mainSystemID)

	for _, c := range model.Containers {
		assert.Equal(t, mainSystemID, c.SystemID, "container %s should belong to main system", c.Name)
	}
}

func TestBuilder_CreatesComponents(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(sampleAnalysisResult())
	require.NoError(t, err)

	// Should have components: Handler (struct), BaseService (struct), Service (struct),
	// Worker (interface), plus standalone functions: New (handler), New (service), HelperFunc, Log
	assert.GreaterOrEqual(t, len(model.Components), 4, "should have at least struct/interface components")

	componentNames := make(map[string]bool)
	for _, c := range model.Components {
		componentNames[c.Name] = true
	}
	assert.True(t, componentNames["Handler"])
	assert.True(t, componentNames["Service"])
	assert.True(t, componentNames["Worker"])
	assert.True(t, componentNames["BaseService"])
}

func TestBuilder_ComponentTypes(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(sampleAnalysisResult())
	require.NoError(t, err)

	typeMap := make(map[string]c4model.ComponentType)
	for _, c := range model.Components {
		typeMap[c.Name] = c.Type
	}
	assert.Equal(t, c4model.ComponentTypeStruct, typeMap["Handler"])
	assert.Equal(t, c4model.ComponentTypeStruct, typeMap["Service"])
	assert.Equal(t, c4model.ComponentTypeInterface, typeMap["Worker"])
}

func TestBuilder_CreatesContainerRelationships(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(sampleAnalysisResult())
	require.NoError(t, err)

	// Should have relationships from import graph
	assert.NotEmpty(t, model.Relationships, "should have relationships")

	// Find relationships where description is "imports"
	var importRels []c4model.Relationship
	for _, r := range model.Relationships {
		if r.Description == "imports" {
			importRels = append(importRels, r)
		}
	}
	assert.NotEmpty(t, importRels, "should have import relationships")
}

func TestBuilder_CreatesCodeElements(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(sampleAnalysisResult())
	require.NoError(t, err)

	assert.NotEmpty(t, model.CodeElements, "should have code elements")

	// Check that methods are represented as code elements
	var methodElements []c4model.CodeElement
	for _, ce := range model.CodeElements {
		if ce.Type == c4model.CodeElementTypeMethod {
			methodElements = append(methodElements, ce)
		}
	}
	assert.NotEmpty(t, methodElements, "should have method code elements")
}

func TestBuilder_NilAnalysisResult(t *testing.T) {
	builder := c4model.NewModelBuilder()
	_, err := builder.BuildFromAnalysis(nil)
	assert.Error(t, err)
}
