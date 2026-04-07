package llm

import "context"

// Client defines the interface for LLM completion calls.
type Client interface {
	Complete(ctx context.Context, req Request) (*Response, error)
}

// Request represents an LLM completion request.
type Request struct {
	SystemPrompt string    `json:"system_prompt"`
	Messages     []Message `json:"messages"`
	MaxTokens    int       `json:"max_tokens"`
}

// Message represents a single message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents the LLM completion response.
type Response struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
}

// Usage holds token usage statistics for a completion call.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
