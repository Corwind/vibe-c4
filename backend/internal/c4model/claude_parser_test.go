package c4model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	assert.Equal(t, "Go", result.Containers[0].Technology)

	assert.Len(t, result.Relationships, 1)
	assert.Equal(t, "sys1", result.Relationships[0].SourceID)
}

func TestParsePass1ResponseNoSystems(t *testing.T) {
	input := `{
		"systems": [],
		"containers": [],
		"relationships": []
	}`

	result, err := ParsePass1Response(input)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 1 system")
}

func TestParsePass1ResponseInvalidJSON(t *testing.T) {
	input := `{not valid json`

	result, err := ParsePass1Response(input)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestParsePass1ResponseCodeFence(t *testing.T) {
	input := "Here is the result:\n```json\n" + `{
		"systems": [{"id": "sys1", "name": "System", "description": "A system", "external": false}],
		"containers": [{"id": "c1", "name": "Container", "package_path": "pkg", "system_id": "sys1"}],
		"relationships": []
	}` + "\n```\nDone."

	result, err := ParsePass1Response(input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Systems, 1)
	assert.Equal(t, "sys1", result.Systems[0].ID)
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
	assert.Equal(t, "comp1", result.Components[0].ID)
	assert.Equal(t, ComponentType("struct"), result.Components[0].Type)
	assert.Equal(t, "comp2", result.Components[1].ID)

	assert.Len(t, result.Relationships, 1)
	assert.Equal(t, "comp1", result.Relationships[0].SourceID)
}

func TestParsePass2ResponseEmpty(t *testing.T) {
	input := `{
		"components": [],
		"relationships": []
	}`

	result, err := ParsePass2Response(input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Components)
}

func TestParsePass3ResponseValid(t *testing.T) {
	input := `{
		"code_elements": [
			{"id": "ce1", "name": "HandleRequest", "type": "method", "component_id": "comp1", "file_path": "handler.go", "line": 42},
			{"id": "ce2", "name": "Name", "type": "field", "component_id": "comp1", "file_path": "handler.go", "line": 10}
		]
	}`

	result, err := ParsePass3Response(input)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Len(t, result.CodeElements, 2)
	assert.Equal(t, "ce1", result.CodeElements[0].ID)
	assert.Equal(t, CodeElementType("method"), result.CodeElements[0].Type)
	assert.Equal(t, 42, result.CodeElements[0].Line)
	assert.Equal(t, "ce2", result.CodeElements[1].ID)
}
