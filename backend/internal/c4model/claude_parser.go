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

// ExtractJSON extracts a JSON string from text that may be wrapped in markdown
// code fences.
func ExtractJSON(text string) string {
	// Look for ```json ... ``` fences first.
	if start := strings.Index(text, "```json"); start != -1 {
		jsonStart := start + len("```json")
		if end := strings.Index(text[jsonStart:], "```"); end != -1 {
			return strings.TrimSpace(text[jsonStart : jsonStart+end])
		}
	}
	// Fall back to generic ``` ... ``` fences.
	if start := strings.Index(text, "```"); start != -1 {
		jsonStart := start + len("```")
		if end := strings.Index(text[jsonStart:], "```"); end != -1 {
			return strings.TrimSpace(text[jsonStart : jsonStart+end])
		}
	}
	return strings.TrimSpace(text)
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
