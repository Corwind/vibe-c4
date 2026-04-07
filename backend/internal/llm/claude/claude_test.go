package claude_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Corwind/vibe-c4/backend/internal/llm"
	"github.com/Corwind/vibe-c4/backend/internal/llm/claude"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockServer(handler http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages", handler)
	return httptest.NewServer(mux)
}

func newTestClient(serverURL string) *claude.Client {
	return claude.New(claude.Config{
		APIKey:      "test-key",
		Model:       "claude-sonnet-4-20250514",
		MaxTokens:   1024,
		TimeoutSecs: 5,
		BaseURL:     serverURL,
	})
}

func TestCompleteHappyPath(t *testing.T) {
	server := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := map[string]interface{}{
			"id":   "msg_test",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{"type": "text", "text": "Hello from Claude!"},
			},
			"model":          "claude-sonnet-4-20250514",
			"stop_reason":    "end_turn",
			"stop_sequence":  nil,
			"usage":          map[string]interface{}{"input_tokens": 100, "output_tokens": 50},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := newTestClient(server.URL)

	resp, err := client.Complete(context.Background(), llm.Request{
		SystemPrompt: "You are a helpful assistant.",
		Messages: []llm.Message{
			{Role: "user", Content: "Say hello"},
		},
		MaxTokens: 512,
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello from Claude!", resp.Content)
	assert.Equal(t, 100, resp.Usage.InputTokens)
	assert.Equal(t, 50, resp.Usage.OutputTokens)
}

func TestCompleteAPIError(t *testing.T) {
	server := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type": "error",
			"error": map[string]interface{}{
				"type":    "authentication_error",
				"message": "invalid x-api-key",
			},
		})
	})
	defer server.Close()

	client := newTestClient(server.URL)

	resp, err := client.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "user", Content: "hello"},
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCompleteTimeout(t *testing.T) {
	server := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]interface{}{{"type": "text", "text": "too late"}},
			"model":   "claude-sonnet-4-20250514",
			"usage":   map[string]interface{}{"input_tokens": 10, "output_tokens": 5},
		})
	})
	defer server.Close()

	client := claude.New(claude.Config{
		APIKey:      "test-key",
		Model:       "claude-sonnet-4-20250514",
		MaxTokens:   1024,
		TimeoutSecs: 1,
		BaseURL:     server.URL,
	})

	resp, err := client.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "user", Content: "hello"},
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCompleteEmptyResponse(t *testing.T) {
	server := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"id":            "msg_test",
			"type":          "message",
			"role":          "assistant",
			"content":       []map[string]interface{}{},
			"model":         "claude-sonnet-4-20250514",
			"stop_reason":   "end_turn",
			"stop_sequence": nil,
			"usage":         map[string]interface{}{"input_tokens": 10, "output_tokens": 0},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	client := newTestClient(server.URL)

	resp, err := client.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "user", Content: "hello"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "", resp.Content)
	assert.Equal(t, 10, resp.Usage.InputTokens)
	assert.Equal(t, 0, resp.Usage.OutputTokens)
}
