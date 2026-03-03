package c4model_test

import (
	"encoding/json"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiagramLevel_String(t *testing.T) {
	tests := []struct {
		level    c4model.DiagramLevel
		expected string
	}{
		{c4model.LevelContext, "Context"},
		{c4model.LevelContainer, "Container"},
		{c4model.LevelComponent, "Component"},
		{c4model.LevelCode, "Code"},
		{c4model.DiagramLevel(0), "Unknown"},
		{c4model.DiagramLevel(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestDiagramLevel_Values(t *testing.T) {
	assert.Equal(t, c4model.DiagramLevel(1), c4model.LevelContext)
	assert.Equal(t, c4model.DiagramLevel(2), c4model.LevelContainer)
	assert.Equal(t, c4model.DiagramLevel(3), c4model.LevelComponent)
	assert.Equal(t, c4model.DiagramLevel(4), c4model.LevelCode)
}

func TestComponentType_Constants(t *testing.T) {
	assert.Equal(t, c4model.ComponentType("struct"), c4model.ComponentTypeStruct)
	assert.Equal(t, c4model.ComponentType("interface"), c4model.ComponentTypeInterface)
	assert.Equal(t, c4model.ComponentType("function"), c4model.ComponentTypeFunction)
}

func TestCodeElementType_Constants(t *testing.T) {
	assert.Equal(t, c4model.CodeElementType("function"), c4model.CodeElementTypeFunction)
	assert.Equal(t, c4model.CodeElementType("method"), c4model.CodeElementTypeMethod)
	assert.Equal(t, c4model.CodeElementType("field"), c4model.CodeElementTypeField)
}

func newTestModel() *c4model.C4Model {
	return &c4model.C4Model{
		ProjectID: "test-project",
		Systems: []c4model.System{
			{ID: "sys-1", Name: "My System", ModulePath: "github.com/example/app", ContainerIDs: []string{"cont-1", "cont-2"}},
			{ID: "sys-ext", Name: "External DB", External: true},
		},
		Containers: []c4model.Container{
			{ID: "cont-1", Name: "api", PackagePath: "internal/api", SystemID: "sys-1", ComponentIDs: []string{"comp-1"}},
			{ID: "cont-2", Name: "core", PackagePath: "internal/core", SystemID: "sys-1", ComponentIDs: []string{"comp-2", "comp-3"}},
		},
		Components: []c4model.Component{
			{ID: "comp-1", Name: "Handler", Type: c4model.ComponentTypeStruct, ContainerID: "cont-1"},
			{ID: "comp-2", Name: "Service", Type: c4model.ComponentTypeInterface, ContainerID: "cont-2"},
			{ID: "comp-3", Name: "ServiceImpl", Type: c4model.ComponentTypeStruct, ContainerID: "cont-2"},
		},
		CodeElements: []c4model.CodeElement{
			{ID: "code-1", Name: "ServeHTTP", Type: c4model.CodeElementTypeMethod, ComponentID: "comp-1", FilePath: "internal/api/handler.go", Line: 15},
		},
		Relationships: []c4model.Relationship{
			{SourceID: "comp-1", TargetID: "comp-2", Description: "uses", Technology: "Go interface"},
			{SourceID: "comp-3", TargetID: "comp-2", Description: "implements"},
			{SourceID: "cont-1", TargetID: "cont-2", Description: "imports"},
		},
	}
}

func TestC4Model_FindSystem(t *testing.T) {
	model := newTestModel()

	t.Run("existing system", func(t *testing.T) {
		sys := model.FindSystem("sys-1")
		require.NotNil(t, sys)
		assert.Equal(t, "My System", sys.Name)
		assert.False(t, sys.External)
	})

	t.Run("external system", func(t *testing.T) {
		sys := model.FindSystem("sys-ext")
		require.NotNil(t, sys)
		assert.True(t, sys.External)
	})

	t.Run("non-existing system", func(t *testing.T) {
		sys := model.FindSystem("does-not-exist")
		assert.Nil(t, sys)
	})
}

func TestC4Model_FindContainer(t *testing.T) {
	model := newTestModel()

	t.Run("existing container", func(t *testing.T) {
		cont := model.FindContainer("cont-1")
		require.NotNil(t, cont)
		assert.Equal(t, "api", cont.Name)
		assert.Equal(t, "sys-1", cont.SystemID)
	})

	t.Run("non-existing container", func(t *testing.T) {
		cont := model.FindContainer("does-not-exist")
		assert.Nil(t, cont)
	})
}

func TestC4Model_FindComponent(t *testing.T) {
	model := newTestModel()

	t.Run("existing component", func(t *testing.T) {
		comp := model.FindComponent("comp-2")
		require.NotNil(t, comp)
		assert.Equal(t, "Service", comp.Name)
		assert.Equal(t, c4model.ComponentTypeInterface, comp.Type)
	})

	t.Run("non-existing component", func(t *testing.T) {
		comp := model.FindComponent("does-not-exist")
		assert.Nil(t, comp)
	})
}

func TestC4Model_ContainersForSystem(t *testing.T) {
	model := newTestModel()

	t.Run("system with containers", func(t *testing.T) {
		containers := model.ContainersForSystem("sys-1")
		assert.Len(t, containers, 2)
		assert.Equal(t, "cont-1", containers[0].ID)
		assert.Equal(t, "cont-2", containers[1].ID)
	})

	t.Run("system without containers", func(t *testing.T) {
		containers := model.ContainersForSystem("sys-ext")
		assert.Empty(t, containers)
	})

	t.Run("non-existing system", func(t *testing.T) {
		containers := model.ContainersForSystem("does-not-exist")
		assert.Empty(t, containers)
	})
}

func TestC4Model_ComponentsForContainer(t *testing.T) {
	model := newTestModel()

	t.Run("container with multiple components", func(t *testing.T) {
		components := model.ComponentsForContainer("cont-2")
		assert.Len(t, components, 2)
		assert.Equal(t, "comp-2", components[0].ID)
		assert.Equal(t, "comp-3", components[1].ID)
	})

	t.Run("container with single component", func(t *testing.T) {
		components := model.ComponentsForContainer("cont-1")
		assert.Len(t, components, 1)
		assert.Equal(t, "comp-1", components[0].ID)
	})

	t.Run("non-existing container", func(t *testing.T) {
		components := model.ComponentsForContainer("does-not-exist")
		assert.Empty(t, components)
	})
}

func TestC4Model_RelationshipsFrom(t *testing.T) {
	model := newTestModel()

	t.Run("source with relationships", func(t *testing.T) {
		rels := model.RelationshipsFrom("comp-1")
		assert.Len(t, rels, 1)
		assert.Equal(t, "comp-2", rels[0].TargetID)
		assert.Equal(t, "uses", rels[0].Description)
	})

	t.Run("source without relationships", func(t *testing.T) {
		rels := model.RelationshipsFrom("does-not-exist")
		assert.Empty(t, rels)
	})
}

func TestC4Model_RelationshipsTo(t *testing.T) {
	model := newTestModel()

	t.Run("target with multiple relationships", func(t *testing.T) {
		rels := model.RelationshipsTo("comp-2")
		assert.Len(t, rels, 2)
	})

	t.Run("target without relationships", func(t *testing.T) {
		rels := model.RelationshipsTo("does-not-exist")
		assert.Empty(t, rels)
	})
}

func TestSystem_JSONSerialization(t *testing.T) {
	sys := c4model.System{
		ID:           "sys-1",
		Name:         "Test System",
		Description:  "A test system",
		External:     false,
		ModulePath:   "github.com/example/app",
		ContainerIDs: []string{"c1", "c2"},
	}

	data, err := json.Marshal(sys)
	require.NoError(t, err)

	var decoded c4model.System
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, sys, decoded)
}

func TestRelationship_JSONSerialization(t *testing.T) {
	rel := c4model.Relationship{
		SourceID:    "a",
		TargetID:    "b",
		Description: "calls",
		Technology:  "gRPC",
	}

	data, err := json.Marshal(rel)
	require.NoError(t, err)

	var decoded c4model.Relationship
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, rel, decoded)
}

func TestRelationship_JSONOmitsEmptyTechnology(t *testing.T) {
	rel := c4model.Relationship{
		SourceID:    "a",
		TargetID:    "b",
		Description: "uses",
	}

	data, err := json.Marshal(rel)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	_, hasTech := raw["technology"]
	assert.False(t, hasTech, "technology should be omitted when empty")
}
