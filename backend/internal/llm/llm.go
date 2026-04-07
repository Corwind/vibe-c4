package llm

import "context"

// Message represents a conversation message.
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// Request represents an LLM completion request.
type Request struct {
	SystemPrompt string
	Messages     []Message
	MaxTokens    int
}

// Usage contains token usage statistics.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Response represents an LLM completion response.
type Response struct {
	Content string
	Usage   Usage
}

// Client defines the interface for LLM interactions.
type Client interface {
	Complete(ctx context.Context, req Request) (*Response, error)
}
