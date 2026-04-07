package claude

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Corwind/vibe-c4/backend/internal/llm"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Config holds the configuration for the Claude client.
type Config struct {
	APIKey      string
	Model       string
	MaxTokens   int
	TimeoutSecs int
	BaseURL     string // For testing only
}

// Client implements llm.Client using the Anthropic Claude API.
type Client struct {
	client  anthropic.Client
	model   string
	max     int
	timeout time.Duration
}

// New creates a new Claude client from the given configuration.
func New(cfg Config) *Client {
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	timeout := 30 * time.Second
	if cfg.TimeoutSecs > 0 {
		timeout = time.Duration(cfg.TimeoutSecs) * time.Second
	}

	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	return &Client{
		client:  client,
		model:   cfg.Model,
		max:     maxTokens,
		timeout: timeout,
	}
}

// Complete sends a completion request to the Claude API and returns the response.
func (c *Client) Complete(ctx context.Context, req llm.Request) (*llm.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	for _, m := range req.Messages {
		switch m.Role {
		case "user":
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case "assistant":
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		default:
			return nil, fmt.Errorf("unsupported message role: %s", m.Role)
		}
	}

	maxTokens := int64(c.max)
	if req.MaxTokens > 0 {
		maxTokens = int64(req.MaxTokens)
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: maxTokens,
		Messages:  messages,
	}

	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		}
	}

	message, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("claude API call failed: %w", err)
	}

	var parts []string
	for _, block := range message.Content {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}

	return &llm.Response{
		Content: strings.Join(parts, ""),
		Usage: llm.Usage{
			InputTokens:  int(message.Usage.InputTokens),
			OutputTokens: int(message.Usage.OutputTokens),
		},
	}, nil
}
