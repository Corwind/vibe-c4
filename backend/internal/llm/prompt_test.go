package llm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPromptMessages_SystemPromptContainsC4Rules(t *testing.T) {
	input := &InterpretationInput{ProjectName: "test"}
	systemPrompt, _ := BuildPromptMessages(input)

	requiredPhrases := []string{
		"System Context",
		"Containers",
		"Components",
		"deployable unit",
		"Do NOT create a component for every struct",
		"business terms",
	}

	for _, phrase := range requiredPhrases {
		assert.True(t, strings.Contains(systemPrompt, phrase),
			"system prompt should contain %q", phrase)
	}
}

func TestBuildPromptMessages_UserMessageContainsJSON(t *testing.T) {
	input := &InterpretationInput{
		ProjectName: "myproject",
		ModulePath:  "github.com/example/myproject",
		DirectDeps:  []string{"github.com/go-chi/chi/v5"},
	}

	_, userMessage := BuildPromptMessages(input)

	assert.Contains(t, userMessage, `"project_name":"myproject"`)
	assert.Contains(t, userMessage, `"module_path":"github.com/example/myproject"`)
	assert.Contains(t, userMessage, `"direct_deps":["github.com/go-chi/chi/v5"]`)
}

func TestBuildPromptMessages_UserMessageInstructsJSONOnly(t *testing.T) {
	input := &InterpretationInput{ProjectName: "test"}
	_, userMessage := BuildPromptMessages(input)

	assert.Contains(t, userMessage, "ONLY")
	assert.Contains(t, userMessage, "JSON")
}
