package c4model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseClaudeResponse parses Claude's response text into a C4Model.
// It handles responses that may be wrapped in markdown code fences.
func ParseClaudeResponse(response string) (*C4Model, error) {
	jsonStr := ExtractJSON(response)

	var model C4Model
	if err := json.Unmarshal([]byte(jsonStr), &model); err != nil {
		return nil, fmt.Errorf("failed to parse C4 model JSON: %w", err)
	}

	return &model, nil
}

// ValidateC4Model checks that the model is structurally valid:
// - model is not nil
// - at least one system exists
// - all containers reference valid system IDs
// - all components reference valid container IDs
// - all relationships reference valid source and target IDs
func ValidateC4Model(model *C4Model) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	if len(model.Systems) == 0 {
		return fmt.Errorf("model must contain at least one system")
	}

	// Build typed ID sets for proper cross-level validation
	systemIDs := make(map[string]bool)
	containerIDs := make(map[string]bool)
	allIDs := make(map[string]bool)

	for _, s := range model.Systems {
		systemIDs[s.ID] = true
		allIDs[s.ID] = true
	}
	for _, c := range model.Containers {
		containerIDs[c.ID] = true
		allIDs[c.ID] = true
	}
	for _, c := range model.Components {
		allIDs[c.ID] = true
	}
	for _, ce := range model.CodeElements {
		allIDs[ce.ID] = true
	}

	// Validate container SystemID references (must reference an actual system)
	for _, c := range model.Containers {
		if !systemIDs[c.SystemID] {
			return fmt.Errorf("container %q references non-existent system ID %q", c.ID, c.SystemID)
		}
	}

	// Validate component ContainerID references (must reference an actual container)
	for _, c := range model.Components {
		if !containerIDs[c.ContainerID] {
			return fmt.Errorf("component %q references non-existent container ID %q", c.ID, c.ContainerID)
		}
	}

	// Validate relationship references (can reference any entity)
	for _, r := range model.Relationships {
		if !allIDs[r.SourceID] {
			return fmt.Errorf("relationship references non-existent source ID %q", r.SourceID)
		}
		if !allIDs[r.TargetID] {
			return fmt.Errorf("relationship references non-existent target ID %q", r.TargetID)
		}
	}

	return nil
}

// ExtractJSON extracts JSON content from text that may be wrapped in markdown code fences.
// It handles: raw JSON, ```json ... ```, and ``` ... ``` formats.
func ExtractJSON(text string) string {
	trimmed := strings.TrimSpace(text)

	// If it already starts with '{', return as-is
	if strings.HasPrefix(trimmed, "{") {
		return trimmed
	}

	// Try ```json\n...\n```
	if idx := strings.Index(trimmed, "```json"); idx != -1 {
		start := idx + len("```json")
		// Skip to next newline
		if nlIdx := strings.Index(trimmed[start:], "\n"); nlIdx != -1 {
			start += nlIdx + 1
		}
		if end := strings.Index(trimmed[start:], "```"); end != -1 {
			return strings.TrimSpace(trimmed[start : start+end])
		}
	}

	// Try ```\n...\n```
	if idx := strings.Index(trimmed, "```"); idx != -1 {
		start := idx + len("```")
		// Skip to next newline
		if nlIdx := strings.Index(trimmed[start:], "\n"); nlIdx != -1 {
			start += nlIdx + 1
		}
		remaining := trimmed[start:]
		if end := strings.Index(remaining, "```"); end != -1 {
			return strings.TrimSpace(remaining[:end])
		}
	}

	return trimmed
}
