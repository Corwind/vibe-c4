package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ClaudeConfig holds configuration for the Claude API adapter.
type ClaudeConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
	BaseURL   string
	Timeout   time.Duration
}

// ClaudeAdapter implements Interpreter by calling the Anthropic Messages API.
type ClaudeAdapter struct {
	config ClaudeConfig
	client *http.Client
}

// NewClaudeAdapter creates a ClaudeAdapter with defaults applied for zero-value fields.
func NewClaudeAdapter(config ClaudeConfig) *ClaudeAdapter {
	if config.Model == "" {
		config.Model = "claude-sonnet-4-20250514"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 8192
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}

	return &ClaudeAdapter{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

type claudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	System    string           `json:"system"`
	Messages  []claudeMessage  `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []claudeContentBlock `json:"content"`
	Error   *claudeError         `json:"error,omitempty"`
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type claudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Interpret sends analysis facts to Claude and returns the C4 interpretation.
func (a *ClaudeAdapter) Interpret(ctx context.Context, input *InterpretationInput) (*InterpretationResult, error) {
	systemPrompt, userMessage := BuildPromptMessages(input)

	reqBody := claudeRequest{
		Model:     a.config.Model,
		MaxTokens: a.config.MaxTokens,
		System:    systemPrompt,
		Messages: []claudeMessage{
			{Role: "user", Content: userMessage},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("llm: failed to marshal request: %w", err)
	}

	url := a.config.BaseURL + "/v1/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("llm: failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("llm: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm: API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBytes, &claudeResp); err != nil {
		return nil, fmt.Errorf("llm: failed to parse response: %w", err)
	}

	if claudeResp.Error != nil {
		return nil, fmt.Errorf("llm: API error (%s): %s", claudeResp.Error.Type, claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("llm: empty response content")
	}

	text := claudeResp.Content[0].Text
	text = stripCodeFences(text)

	var result InterpretationResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("llm: failed to parse interpretation JSON: %w", err)
	}

	return &result, nil
}

// stripCodeFences removes markdown code fences (```json ... ```) if present.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Remove opening fence line
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		// Remove closing fence
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
