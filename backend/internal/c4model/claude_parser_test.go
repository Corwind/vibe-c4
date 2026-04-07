package c4model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validC4ModelJSON() string {
	return `{
  "project_id": "test-project",
  "systems": [
    {
      "id": "system-myapp",
      "name": "MyApp",
      "description": "Main application system",
      "external": false
    }
  ],
  "containers": [
    {
      "id": "container-api",
      "name": "API",
      "description": "REST API service",
      "technology": "Go",
      "package_path": "internal/api",
      "system_id": "system-myapp"
    }
  ],
  "components": [
    {
      "id": "component-handler",
      "name": "Handler",
      "description": "HTTP request handler",
      "technology": "Go struct",
      "type": "struct",
      "container_id": "container-api"
    }
  ],
  "code_elements": [
    {
      "id": "code-handler-serve",
      "name": "ServeHTTP",
      "description": "Handles incoming HTTP requests",
      "type": "method",
      "component_id": "component-handler"
    }
  ],
  "relationships": [
    {
      "source_id": "system-myapp",
      "target_id": "container-api",
      "description": "contains",
      "level": "context"
    }
  ]
}`
}

func TestParseClaudeResponseValidJSON(t *testing.T) {
	model, err := ParseClaudeResponse(validC4ModelJSON())
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, "test-project", model.ProjectID)
	assert.Len(t, model.Systems, 1)
	assert.Equal(t, "system-myapp", model.Systems[0].ID)
	assert.Len(t, model.Containers, 1)
	assert.Equal(t, "container-api", model.Containers[0].ID)
	assert.Len(t, model.Components, 1)
	assert.Equal(t, "component-handler", model.Components[0].ID)
	assert.Len(t, model.CodeElements, 1)
	assert.Equal(t, "code-handler-serve", model.CodeElements[0].ID)
	assert.Len(t, model.Relationships, 1)
}

func TestParseClaudeResponseWithCodeFence(t *testing.T) {
	wrapped := "Here is the C4 model:\n\n```json\n" + validC4ModelJSON() + "\n```\n\nHope this helps!"

	model, err := ParseClaudeResponse(wrapped)
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, "test-project", model.ProjectID)
	assert.Len(t, model.Systems, 1)
}

func TestParseClaudeResponseInvalidJSON(t *testing.T) {
	model, err := ParseClaudeResponse("this is not json at all")
	assert.Error(t, err)
	assert.Nil(t, model)
	assert.Contains(t, err.Error(), "failed to parse C4 model JSON")
}

func TestValidateC4ModelValid(t *testing.T) {
	model := &C4Model{
		Systems: []System{
			{ID: "system-main", Name: "Main"},
		},
		Containers: []Container{
			{ID: "container-api", Name: "API", SystemID: "system-main"},
		},
		Components: []Component{
			{ID: "component-handler", Name: "Handler", ContainerID: "container-api"},
		},
		CodeElements: []CodeElement{
			{ID: "code-serve", Name: "ServeHTTP", ComponentID: "component-handler"},
		},
		Relationships: []Relationship{
			{SourceID: "system-main", TargetID: "container-api", Description: "contains"},
		},
	}

	err := ValidateC4Model(model)
	assert.NoError(t, err)
}

func TestValidateC4ModelNoSystems(t *testing.T) {
	model := &C4Model{
		Systems: []System{},
	}

	err := ValidateC4Model(model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one system")
}

func TestValidateC4ModelNil(t *testing.T) {
	err := ValidateC4Model(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestValidateC4ModelBrokenRelationship(t *testing.T) {
	model := &C4Model{
		Systems: []System{
			{ID: "system-main", Name: "Main"},
		},
		Relationships: []Relationship{
			{SourceID: "system-main", TargetID: "nonexistent-id", Description: "broken"},
		},
	}

	err := ValidateC4Model(model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-existent")
	assert.Contains(t, err.Error(), "nonexistent-id")
}

func TestValidateC4ModelInvalidContainerSystemID(t *testing.T) {
	model := &C4Model{
		Systems: []System{
			{ID: "system-main", Name: "Main"},
		},
		Containers: []Container{
			{ID: "container-api", Name: "API", SystemID: "system-does-not-exist"},
		},
	}

	err := ValidateC4Model(model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-existent system ID")
	assert.Contains(t, err.Error(), "system-does-not-exist")
}

func TestValidateC4ModelInvalidComponentContainerID(t *testing.T) {
	model := &C4Model{
		Systems: []System{
			{ID: "system-main", Name: "Main"},
		},
		Containers: []Container{
			{ID: "container-api", Name: "API", SystemID: "system-main"},
		},
		Components: []Component{
			{ID: "component-handler", Name: "Handler", ContainerID: "container-missing"},
		},
	}

	err := ValidateC4Model(model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-existent container ID")
}

func TestValidateC4ModelInvalidCodeElementComponentID(t *testing.T) {
	model := &C4Model{
		Systems: []System{
			{ID: "system-main", Name: "Main"},
		},
		Containers: []Container{
			{ID: "container-api", Name: "API", SystemID: "system-main"},
		},
		Components: []Component{
			{ID: "component-handler", Name: "Handler", ContainerID: "container-api"},
		},
		CodeElements: []CodeElement{
			{ID: "code-foo", Name: "Foo", ComponentID: "component-nonexistent"},
		},
	}

	err := ValidateC4Model(model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-existent component ID")
	assert.Contains(t, err.Error(), "component-nonexistent")
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "raw JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with leading whitespace",
			input:    "  \n  {\"key\": \"value\"}",
			expected: `{"key": "value"}`,
		},
		{
			name:     "code fence with json tag",
			input:    "Here is the result:\n\n```json\n{\"key\": \"value\"}\n```\n\nDone!",
			expected: `{"key": "value"}`,
		},
		{
			name:     "code fence without json tag",
			input:    "Result:\n\n```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "no JSON at all",
			input:    "just plain text",
			expected: "just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
