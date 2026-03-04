package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeAdapter_SendsCorrectRequest(t *testing.T) {
	var capturedBody []byte
	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		capturedBody, _ = io.ReadAll(r.Body)

		respData, _ := os.ReadFile("testdata/valid_response.json")
		w.Header().Set("Content-Type", "application/json")
		w.Write(respData)
	}))
	defer server.Close()

	adapter := NewClaudeAdapter(ClaudeConfig{
		APIKey:  "test-key-123",
		Model:   "claude-sonnet-4-20250514",
		BaseURL: server.URL,
	})

	input := &InterpretationInput{ProjectName: "test-project"}
	_, err := adapter.Interpret(context.Background(), input)
	require.NoError(t, err)

	// Verify headers
	assert.Equal(t, "test-key-123", capturedHeaders.Get("x-api-key"))
	assert.Equal(t, "2023-06-01", capturedHeaders.Get("anthropic-version"))
	assert.Equal(t, "application/json", capturedHeaders.Get("content-type"))

	// Verify body structure
	var reqBody claudeRequest
	err = json.Unmarshal(capturedBody, &reqBody)
	require.NoError(t, err)

	assert.Equal(t, "claude-sonnet-4-20250514", reqBody.Model)
	assert.Equal(t, 8192, reqBody.MaxTokens)
	assert.NotEmpty(t, reqBody.System)
	assert.Len(t, reqBody.Messages, 1)
	assert.Equal(t, "user", reqBody.Messages[0].Role)
	assert.Contains(t, reqBody.Messages[0].Content, "test-project")
}

func TestClaudeAdapter_ParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respData, _ := os.ReadFile("testdata/valid_response.json")
		w.Header().Set("Content-Type", "application/json")
		w.Write(respData)
	}))
	defer server.Close()

	adapter := NewClaudeAdapter(ClaudeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	result, err := adapter.Interpret(context.Background(), &InterpretationInput{ProjectName: "test"})
	require.NoError(t, err)

	assert.Equal(t, "TestApp", result.SystemContext.Name)
	assert.Equal(t, "A test application", result.SystemContext.Description)
	require.Len(t, result.SystemContext.Actors, 1)
	assert.Equal(t, "User", result.SystemContext.Actors[0].Name)
	require.Len(t, result.SystemContext.ExternalSystems, 1)
	assert.Equal(t, "PostgreSQL", result.SystemContext.ExternalSystems[0].Name)
	require.Len(t, result.Containers, 1)
	assert.Equal(t, "test-app", result.Containers[0].ID)
	require.Len(t, result.Components, 1)
	assert.Equal(t, "user-controller", result.Components[0].ID)
	require.Len(t, result.Relationships, 1)
	assert.Equal(t, "component", result.Relationships[0].Level)
}

func TestClaudeAdapter_StripsCodeFences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respData, _ := os.ReadFile("testdata/codefenced_response.json")
		w.Header().Set("Content-Type", "application/json")
		w.Write(respData)
	}))
	defer server.Close()

	adapter := NewClaudeAdapter(ClaudeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	result, err := adapter.Interpret(context.Background(), &InterpretationInput{ProjectName: "test"})
	require.NoError(t, err)

	assert.Equal(t, "TestApp", result.SystemContext.Name)
	require.Len(t, result.Containers, 1)
	assert.Equal(t, "test-app", result.Containers[0].ID)
}

func TestClaudeAdapter_HandlesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"type":"invalid_request","message":"bad request"}}`))
	}))
	defer server.Close()

	adapter := NewClaudeAdapter(ClaudeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	_, err := adapter.Interpret(context.Background(), &InterpretationInput{ProjectName: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestClaudeAdapter_HandlesEmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"content":[]}`))
	}))
	defer server.Close()

	adapter := NewClaudeAdapter(ClaudeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	_, err := adapter.Interpret(context.Background(), &InterpretationInput{ProjectName: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty response content")
}

func TestNewClaudeAdapter_AppliesDefaults(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		respData, _ := os.ReadFile("testdata/valid_response.json")
		w.Header().Set("Content-Type", "application/json")
		w.Write(respData)
	}))
	defer server.Close()

	adapter := NewClaudeAdapter(ClaudeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	_, err := adapter.Interpret(context.Background(), &InterpretationInput{ProjectName: "test"})
	require.NoError(t, err)

	var reqBody claudeRequest
	err = json.Unmarshal(capturedBody, &reqBody)
	require.NoError(t, err)

	assert.Equal(t, "claude-sonnet-4-20250514", reqBody.Model)
	assert.Equal(t, 8192, reqBody.MaxTokens)
}
