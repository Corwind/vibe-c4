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

func TestParsePass1ResponseValid(t *testing.T) {
	input := `{
		"systems": [{"id": "sys1", "name": "My System", "description": "Main system", "external": false}],
		"containers": [{"id": "cont1", "name": "API", "description": "REST API", "technology": "Go", "package_path": "cmd/api", "system_id": "sys1"}],
		"relationships": [{"source_id": "sys1", "target_id": "cont1", "description": "contains"}]
	}`

	result, err := ParsePass1Response(input)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Len(t, result.Systems, 1)
	assert.Equal(t, "sys1", result.Systems[0].ID)
	assert.Equal(t, "My System", result.Systems[0].Name)
	assert.Len(t, result.Containers, 1)
	assert.Equal(t, "cont1", result.Containers[0].ID)
	assert.Len(t, result.Relationships, 1)
}

func TestParsePass1ResponseNoSystems(t *testing.T) {
	input := `{"systems": [], "containers": [], "relationships": []}`

	result, err := ParsePass1Response(input)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 1 system")
}

func TestParsePass1ResponseInvalidJSON(t *testing.T) {
	result, err := ParsePass1Response(`{not valid json`)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestParsePass1ResponseCodeFence(t *testing.T) {
	input := "Here is the result:\n```json\n" + `{
		"systems": [{"id": "sys1", "name": "System", "description": "A system", "external": false}],
		"containers": [],
		"relationships": []
	}` + "\n```\nDone."

	result, err := ParsePass1Response(input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Systems, 1)
}

func TestParsePass2ResponseValid(t *testing.T) {
	input := `{
		"components": [
			{"id": "comp1", "name": "Handler", "type": "struct", "container_id": "c1"},
			{"id": "comp2", "name": "Service", "type": "interface", "container_id": "c1"}
		],
		"relationships": [{"source_id": "comp1", "target_id": "comp2", "description": "uses"}]
	}`

	result, err := ParsePass2Response(input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Components, 2)
	assert.Len(t, result.Relationships, 1)
}

func TestParsePass2ResponseEmpty(t *testing.T) {
	input := `{"components": [], "relationships": []}`

	result, err := ParsePass2Response(input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Components)
}

func TestParsePass3ResponseValid(t *testing.T) {
	input := `{
		"code_elements": [
			{"id": "ce1", "name": "HandleRequest", "type": "method", "component_id": "comp1", "file_path": "handler.go", "line": 42},
			{"id": "ce2", "name": "Name", "type": "field", "component_id": "comp1"}
		]
	}`

	result, err := ParsePass3Response(input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.CodeElements, 2)
	assert.Equal(t, "ce1", result.CodeElements[0].ID)
	assert.Equal(t, 42, result.CodeElements[0].Line)
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
