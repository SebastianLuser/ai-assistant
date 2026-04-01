package clients

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"jarvis/pkg/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey = "sk-test-key"
	testModel  = "claude-test"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *ClaudeClient) {
	srv := httptest.NewServer(handler)
	client := NewClaudeClientWithBaseURL(testAPIKey, testModel, srv.URL)
	return srv, client
}

func successHandler(text string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]string{{"type": "text", "text": text}},
			"usage":   map[string]int{"input_tokens": 10, "output_tokens": 5},
		})
	}
}

func TestComplete_ReturnsText(t *testing.T) {
	srv, client := newTestServer(successHandler("hello world"))
	defer srv.Close()

	result, err := client.Complete("system", "user message")

	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestComplete_SendsCorrectHeaders(t *testing.T) {
	var receivedHeaders http.Header
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		successHandler("ok")(w, r)
	})
	defer srv.Close()

	_, err := client.Complete("", "test")

	require.NoError(t, err)
	assert.Equal(t, testAPIKey, receivedHeaders.Get("x-api-key"))
	assert.Equal(t, "2023-06-01", receivedHeaders.Get("anthropic-version"))
	assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
}

func TestComplete_SendsCorrectBody(t *testing.T) {
	var receivedReq claudeRequest
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedReq)
		successHandler("ok")(w, r)
	})
	defer srv.Close()

	_, err := client.Complete("system prompt", "user msg")

	require.NoError(t, err)
	assert.Equal(t, testModel, receivedReq.Model)
	assert.NotNil(t, receivedReq.System)
	assert.Equal(t, 2048, receivedReq.MaxTokens)
	require.Len(t, receivedReq.Messages, 1)
	assert.Equal(t, "user", receivedReq.Messages[0].Role)
}

func TestComplete_APIError_ReturnsError(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"type": "invalid_request_error", "message": "bad request"},
		})
	})
	defer srv.Close()

	_, err := client.Complete("", "test")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrClaudeAPI))
}

func TestComplete_EmptyResponse_ReturnsError(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"content": []any{},
		})
	})
	defer srv.Close()

	_, err := client.Complete("", "test")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrClaudeEmpty))
}

func TestCompleteJSON_UnmarshalsResponse(t *testing.T) {
	srv, client := newTestServer(successHandler(`{"amount":5000,"category":"Transporte"}`))
	defer srv.Close()

	var target struct {
		Amount   float64 `json:"amount"`
		Category string  `json:"category"`
	}
	err := client.CompleteJSON("", "test", &target)

	require.NoError(t, err)
	assert.Equal(t, float64(5000), target.Amount)
	assert.Equal(t, "Transporte", target.Category)
}

func TestCompleteJSON_InvalidJSON_ReturnsError(t *testing.T) {
	srv, client := newTestServer(successHandler("not valid json"))
	defer srv.Close()

	var target map[string]any
	err := client.CompleteJSON("", "test", &target)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrClaudeJSON))
}

func TestCompleteMessages_MultipleMessages(t *testing.T) {
	var receivedReq claudeRequest
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedReq)
		successHandler("response")(w, r)
	})
	defer srv.Close()

	msgs := []domain.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi"},
		{Role: "user", Content: "how are you"},
	}
	result, err := client.CompleteMessages("sys", msgs)

	require.NoError(t, err)
	assert.Equal(t, "response", result)
	assert.Len(t, receivedReq.Messages, 3)
}

func TestNewClaudeClientWithBaseURL_OverridesURL(t *testing.T) {
	client := NewClaudeClientWithBaseURL("key", "model", "http://localhost:9999")

	assert.Equal(t, "http://localhost:9999", client.baseURL)
}

func TestNewClaudeClient_UsesDefaultURL(t *testing.T) {
	client := NewClaudeClient("key", "model")

	assert.Equal(t, claudeDefaultBaseURL, client.baseURL)
}

func TestClaudeClient_ImplementsAIProvider(t *testing.T) {
	var _ domain.AIProvider = (*ClaudeClient)(nil)
}
