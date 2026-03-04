package c4model

import (
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIBuilder_BuildsSystemContext(t *testing.T) {
	interpretation := &llm.InterpretationResult{
		SystemContext: llm.SystemContextInterpretation{
			Name:        "MyApp",
			Description: "A sample application",
			Actors: []llm.ActorInterpretation{
				{Name: "WebUser", Description: "End user", Type: "person"},
			},
			ExternalSystems: []llm.ExternalSystemInterpretation{
				{Name: "PostgreSQL", Description: "Primary database", Technology: "SQL", Kind: "database"},
			},
		},
	}

	builder := NewAIModelBuilder(interpretation)
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/myapp"},
	}

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)

	// Main system + actor + external = 3 systems
	require.Len(t, model.Systems, 3)

	// Main system
	assert.Equal(t, "MyApp", model.Systems[0].Name)
	assert.False(t, model.Systems[0].External)

	// Actor
	assert.Equal(t, "WebUser", model.Systems[1].Name)
	assert.True(t, model.Systems[1].External)
	assert.Equal(t, "actor", model.Systems[1].SystemKind)

	// External system
	assert.Equal(t, "PostgreSQL", model.Systems[2].Name)
	assert.True(t, model.Systems[2].External)
	assert.Equal(t, "database", model.Systems[2].SystemKind)

	// System-level relationships: actor->main, main->external
	systemRels := filterRelsByLevel(model.Relationships, "system")
	require.Len(t, systemRels, 2)
	assert.Equal(t, model.Systems[1].ID, systemRels[0].SourceID) // actor -> main
	assert.Equal(t, model.Systems[0].ID, systemRels[0].TargetID)
	assert.Equal(t, model.Systems[0].ID, systemRels[1].SourceID) // main -> external
	assert.Equal(t, model.Systems[2].ID, systemRels[1].TargetID)
	assert.Equal(t, "SQL", systemRels[1].Technology)
}

func TestAIBuilder_BuildsContainers(t *testing.T) {
	interpretation := &llm.InterpretationResult{
		SystemContext: llm.SystemContextInterpretation{
			Name: "MyApp",
		},
		Containers: []llm.ContainerInterpretation{
			{
				ID:           "container-api",
				Name:         "API Service",
				Description:  "REST API",
				Technology:   "Go",
				Type:         "service",
				PackagePaths: []string{"internal/api", "internal/handler"},
			},
			{
				ID:          "container-db",
				Name:        "Database",
				Description: "PostgreSQL store",
				Technology:  "PostgreSQL",
				Type:        "database",
			},
		},
	}

	builder := NewAIModelBuilder(interpretation)
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/myapp"},
	}

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)
	require.Len(t, model.Containers, 2)

	// First container
	assert.Equal(t, "container-api", model.Containers[0].ID)
	assert.Equal(t, "API Service", model.Containers[0].Name)
	assert.Equal(t, "REST API", model.Containers[0].Description)
	assert.Equal(t, "Go", model.Containers[0].Technology)
	assert.Equal(t, "internal/api,internal/handler", model.Containers[0].PackagePath)
	assert.Equal(t, model.Systems[0].ID, model.Containers[0].SystemID)

	// Second container
	assert.Equal(t, "container-db", model.Containers[1].ID)
	assert.Equal(t, "Database", model.Containers[1].Name)
	assert.Empty(t, model.Containers[1].PackagePath)

	// Main system tracks container IDs
	assert.Contains(t, model.Systems[0].ContainerIDs, "container-api")
	assert.Contains(t, model.Systems[0].ContainerIDs, "container-db")
}

func TestAIBuilder_BuildsComponents(t *testing.T) {
	interpretation := &llm.InterpretationResult{
		SystemContext: llm.SystemContextInterpretation{Name: "MyApp"},
		Containers: []llm.ContainerInterpretation{
			{ID: "container-api", Name: "API"},
		},
		Components: []llm.ComponentInterpretation{
			{
				ID:          "comp-user-handler",
				Name:        "UserHandler",
				Description: "Handles user HTTP requests",
				Role:        "controller",
				ContainerID: "container-api",
			},
			{
				ID:          "comp-user-service",
				Name:        "UserService",
				Description: "User business logic",
				Role:        "service",
				ContainerID: "container-api",
			},
		},
	}

	builder := NewAIModelBuilder(interpretation)
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/myapp"},
	}

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)
	require.Len(t, model.Components, 2)

	assert.Equal(t, "comp-user-handler", model.Components[0].ID)
	assert.Equal(t, "UserHandler", model.Components[0].Name)
	assert.Equal(t, "Handles user HTTP requests", model.Components[0].Description)
	assert.Equal(t, ComponentTypeStruct, model.Components[0].Type)
	assert.Equal(t, "container-api", model.Components[0].ContainerID)

	assert.Equal(t, "comp-user-service", model.Components[1].ID)
	assert.Equal(t, ComponentTypeStruct, model.Components[1].Type)

	// Container tracks component IDs
	assert.Contains(t, model.Containers[0].ComponentIDs, "comp-user-handler")
	assert.Contains(t, model.Containers[0].ComponentIDs, "comp-user-service")
}

func TestAIBuilder_BuildsRelationships(t *testing.T) {
	interpretation := &llm.InterpretationResult{
		SystemContext: llm.SystemContextInterpretation{Name: "MyApp"},
		Relationships: []llm.RelationshipInterpretation{
			{SourceID: "container-api", TargetID: "container-db", Description: "reads/writes", Level: "container"},
			{SourceID: "comp-handler", TargetID: "comp-service", Description: "delegates to", Level: "component"},
		},
	}

	builder := NewAIModelBuilder(interpretation)
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/myapp"},
	}

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)

	// AI relationships are appended (no system-level actors/externals here)
	require.Len(t, model.Relationships, 2)
	assert.Equal(t, "container-api", model.Relationships[0].SourceID)
	assert.Equal(t, "container", model.Relationships[0].Level)
	assert.Equal(t, "comp-handler", model.Relationships[1].SourceID)
	assert.Equal(t, "component", model.Relationships[1].Level)
}

func TestAIBuilder_BuildsCodeElements(t *testing.T) {
	interpretation := &llm.InterpretationResult{
		SystemContext: llm.SystemContextInterpretation{Name: "MyApp"},
		Components: []llm.ComponentInterpretation{
			{
				ID:          "comp-user-svc",
				Name:        "UserService",
				ContainerID: "container-api",
				StructNames: []string{"UserService"},
			},
		},
	}

	builder := NewAIModelBuilder(interpretation)
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/myapp"},
		Packages: []analyzer.PackageInfo{
			{
				ImportPath: "github.com/example/myapp/internal/service",
				Structs: []analyzer.StructInfo{
					{
						Name: "UserService",
						Methods: []analyzer.MethodInfo{
							{Name: "CreateUser", FilePath: "service.go", Line: 10},
						},
						Fields: []analyzer.FieldInfo{
							{Name: "repo", Type: "Repository"},
							{Name: "Base", Type: "BaseService", Embedded: true},
						},
					},
				},
			},
		},
	}

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)

	// Method + non-embedded field = 2 code elements
	require.Len(t, model.CodeElements, 2)

	// Method linked to AI component ID
	method := model.CodeElements[0]
	assert.Equal(t, "CreateUser", method.Name)
	assert.Equal(t, CodeElementTypeMethod, method.Type)
	assert.Equal(t, "comp-user-svc", method.ComponentID) // AI ID, not sanitizeID fallback

	// Non-embedded field linked to AI component ID
	field := model.CodeElements[1]
	assert.Equal(t, "repo", field.Name)
	assert.Equal(t, CodeElementTypeField, field.Type)
	assert.Equal(t, "comp-user-svc", field.ComponentID)
}

func TestAIBuilder_NilResult(t *testing.T) {
	interpretation := &llm.InterpretationResult{
		SystemContext: llm.SystemContextInterpretation{Name: "MyApp"},
	}
	builder := NewAIModelBuilder(interpretation)

	model, err := builder.BuildFromAnalysis(nil)
	assert.Nil(t, model)
	assert.EqualError(t, err, "analysis result is nil")
}

func TestAIBuilder_NilInterpretation(t *testing.T) {
	builder := NewAIModelBuilder(nil)

	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/myapp"},
	}

	model, err := builder.BuildFromAnalysis(result)
	assert.Nil(t, model)
	assert.EqualError(t, err, "interpretation result is nil")
}

// filterRelsByLevel is a test helper that filters relationships by level.
func filterRelsByLevel(rels []Relationship, level string) []Relationship {
	var out []Relationship
	for _, r := range rels {
		if r.Level == level {
			out = append(out, r)
		}
	}
	return out
}
