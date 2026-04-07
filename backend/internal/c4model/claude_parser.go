package c4model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Pass1Result holds parsed output from the L1+L2 (System Context + Containers) pass.
type Pass1Result struct {
	Systems       []System       `json:"systems"`
	Containers    []Container    `json:"containers"`
	Relationships []Relationship `json:"relationships"`
}

// Pass2Result holds parsed output from an L3 (Components) pass.
type Pass2Result struct {
	Components    []Component    `json:"components"`
	Relationships []Relationship `json:"relationships"`
}

// Pass3Result holds parsed output from an L4 (Code Elements) pass.
type Pass3Result struct {
	CodeElements []CodeElement `json:"code_elements"`
}

// ParsePass1Response parses the Claude response for Pass 1 (systems + containers).
func ParsePass1Response(response string) (*Pass1Result, error) {
	cleaned := extractJSON(response)
	var result Pass1Result
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("parse pass1 response: %w", err)
	}
	return &result, nil
}

// ParsePass2Response parses the Claude response for Pass 2 (components).
func ParsePass2Response(response string) (*Pass2Result, error) {
	cleaned := extractJSON(response)
	var result Pass2Result
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("parse pass2 response: %w", err)
	}
	return &result, nil
}

// ParsePass3Response parses the Claude response for Pass 3 (code elements).
func ParsePass3Response(response string) (*Pass3Result, error) {
	cleaned := extractJSON(response)
	var result Pass3Result
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("parse pass3 response: %w", err)
	}
	return &result, nil
}

// ValidateC4Model checks that a C4Model has basic structural integrity.
func ValidateC4Model(model *C4Model) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}
	if len(model.Systems) == 0 {
		return fmt.Errorf("model has no systems")
	}
	return nil
}

// extractJSON attempts to extract a JSON object from a response that may
// contain markdown fences or other surrounding text.
func extractJSON(response string) string {
	// Try to find JSON inside markdown code fences
	if idx := strings.Index(response, "```json"); idx != -1 {
		start := idx + len("```json")
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}
	if idx := strings.Index(response, "```"); idx != -1 {
		start := idx + len("```")
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}
	// Try to find raw JSON object
	if idx := strings.Index(response, "{"); idx != -1 {
		return response[idx:]
	}
	return response
}
