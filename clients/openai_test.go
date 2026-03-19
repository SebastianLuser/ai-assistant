package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAI_Transcribe_WithRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"text": "Hola mundo transcrito",
		})
	}))
	defer srv.Close()

	c := NewOpenAIClient("test-openai-key", "gpt-4")
	// Use redirect transport on the main httpClient - but Transcribe creates its own.
	// We'll test at the transport level by replacing http.DefaultTransport temporarily.
	origTransport := http.DefaultTransport
	http.DefaultTransport = newRedirectTransport(srv.URL)
	defer func() { http.DefaultTransport = origTransport }()

	text, err := c.Transcribe([]byte("fake-audio"), "audio/ogg")

	require.NoError(t, err)
	assert.Equal(t, "Hola mundo transcrito", text)
}

func TestOpenAI_Transcribe_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer srv.Close()

	c := NewOpenAIClient("test-openai-key", "gpt-4")
	origTransport := http.DefaultTransport
	http.DefaultTransport = newRedirectTransport(srv.URL)
	defer func() { http.DefaultTransport = origTransport }()

	_, err := c.Transcribe([]byte("audio"), "audio/mpeg")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "whisper: api error 429")
}

func TestMimeToExt_KnownTypes(t *testing.T) {
	tests := []struct {
		mime     string
		expected string
	}{
		{"audio/ogg", ".ogg"},
		{"audio/ogg; codecs=opus", ".ogg"},
		{"audio/mpeg", ".mp3"},
		{"audio/mp4", ".m4a"},
		{"audio/wav", ".wav"},
		{"audio/webm", ".webm"},
	}

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			assert.Equal(t, tt.expected, mimeToExt(tt.mime))
		})
	}
}

func TestMimeToExt_UnknownDefaultsToOgg(t *testing.T) {
	assert.Equal(t, ".ogg", mimeToExt("audio/unknown"))
	assert.Equal(t, ".ogg", mimeToExt(""))
}

func TestOpenAI_Complete_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openaiChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "hello from openai"}},
			},
		})
	}))
	defer srv.Close()

	c := NewOpenAIClientWithBaseURL("test-key", "gpt-4", srv.URL)

	result, err := c.Complete("system", "user message")

	require.NoError(t, err)
	assert.Equal(t, "hello from openai", result)
}

func TestOpenAI_Complete_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openaiChatResponse{
			Error: &struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			}{
				Message: "quota exceeded",
				Type:    "insufficient_quota",
			},
		})
	}))
	defer srv.Close()

	c := NewOpenAIClientWithBaseURL("test-key", "gpt-4", srv.URL)

	_, err := c.Complete("system", "test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "quota exceeded")
}

func TestOpenAI_Complete_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openaiChatResponse{})
	}))
	defer srv.Close()

	c := NewOpenAIClientWithBaseURL("test-key", "gpt-4", srv.URL)

	_, err := c.Complete("system", "test")

	require.Error(t, err)
}

func TestOpenAI_ImplementsAIProvider(t *testing.T) {
	var _ interface {
		Complete(string, string) (string, error)
	}
	// Compile-time check is already in openai.go, just verify constructor works
	c := NewOpenAIClient("key", "model")
	assert.NotNil(t, c)
}

func TestNewOpenAIClient_DefaultURL(t *testing.T) {
	c := NewOpenAIClient("key", "model")

	assert.Equal(t, openaiDefaultBaseURL, c.baseURL)
}

func TestNewOpenAIClientWithBaseURL_OverridesURL(t *testing.T) {
	c := NewOpenAIClientWithBaseURL("key", "model", "http://localhost:8080")

	assert.Equal(t, "http://localhost:8080", c.baseURL)
}
