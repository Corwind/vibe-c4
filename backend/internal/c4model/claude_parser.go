package c4model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Pass1Result holds the output of the first pipeline pass: systems, containers,
// and top-level relationships.
type Pass1Result struct {
	Systems       []System       `json:"systems"`
	Containers    []Container    `json:"containers"`
	Relationships []Relationship `json:"relationships"`
}

// Pass2Result holds the output of the second pipeline pass: components and
// their relationships within a container.
type Pass2Result struct {
	Components    []Component    `json:"components"`
	Relationships []Relationship `json:"relationships"`
}

// Pass3Result holds the output of the third pipeline pass: code elements for
// individual components.
type Pass3Result struct {
	CodeElements []CodeElement `json:"code_elements"`
}

// ParsePass1Response parses a Claude response into a Pass1Result. It extracts
// JSON from possible markdown fences, unmarshals, and validates that at least
// one system is present.
func ParsePass1Response(response string) (*Pass1Result, error) {
	raw := ExtractJSON(response)

	var result Pass1Result
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("failed to parse pass 1 response: %w", err)
	}

	if len(result.Systems) == 0 {
		return nil, fmt.Errorf("pass 1 result must contain at least 1 system")
	}

	return &result, nil
}

// ParsePass2Response parses a Claude response into a Pass2Result. Components
// may be empty for an empty container, so no strict validation is applied.
func ParsePass2Response(response string) (*Pass2Result, error) {
	raw := ExtractJSON(response)

	var result Pass2Result
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("failed to parse pass 2 response: %w", err)
	}

	return &result, nil
}

// ParsePass3Response parses a Claude response into a Pass3Result. Code elements
// may be empty, so no strict validation is applied.
func ParsePass3Response(response string) (*Pass3Result, error) {
	raw := ExtractJSON(response)

	var result Pass3Result
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("failed to parse pass 3 response: %w", err)
	}

	return &result, nil
}

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

	// Validate code element ComponentID references (must reference an actual component)
	componentIDs := make(map[string]bool)
	for _, c := range model.Components {
		componentIDs[c.ID] = true
	}
	for _, ce := range model.CodeElements {
		if !componentIDs[ce.ComponentID] {
			return fmt.Errorf("code element %q references non-existent component ID %q", ce.ID, ce.ComponentID)
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
