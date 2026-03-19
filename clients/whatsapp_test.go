package clients

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWhatsAppTestClient(srv *httptest.Server) *WhatsAppClient {
	c := NewWhatsAppClient("phone-123", "test-token")
	c.httpClient = srv.Client()
	return c
}

func TestWhatsApp_SendTextMessage_Success(t *testing.T) {
	var receivedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)

		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newWhatsAppTestClient(srv)
	// Override the sendMessage URL by patching the client to point at our test server
	// Since sendMessage builds the URL internally, we need a different approach.
	// We'll use a custom transport that redirects all requests to the test server.
	c.httpClient = &http.Client{
		Transport: newRedirectTransport(srv.URL),
	}

	err := c.SendTextMessage("+5491100001111", "hello world")

	require.NoError(t, err)
	assert.Equal(t, "whatsapp", receivedBody["messaging_product"])
	assert.Equal(t, "text", receivedBody["type"])
}

// redirectTransport redirects all requests to a target URL (test server).
type redirectTransport struct {
	target    string
	transport http.RoundTripper
}

func newRedirectTransport(target string) *redirectTransport {
	return &redirectTransport{
		target:    target,
		transport: &http.Transport{},
	}
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.target, "http://")
	return t.transport.RoundTrip(req)
}

func TestWhatsApp_SendTextMessage_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid phone"}`))
	}))
	defer srv.Close()

	c := NewWhatsAppClient("phone-123", "test-token")
	c.httpClient = &http.Client{
		Transport: newRedirectTransport(srv.URL),
	}

	err := c.SendTextMessage("+5491100001111", "hello")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "whatsapp: api error 400")
	assert.Contains(t, err.Error(), "invalid phone")
}

func TestWhatsApp_MarkAsRead_Success(t *testing.T) {
	var receivedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewWhatsAppClient("phone-123", "test-token")
	c.httpClient = &http.Client{
		Transport: newRedirectTransport(srv.URL),
	}

	err := c.MarkAsRead("wamid.abc123")

	require.NoError(t, err)
	assert.Equal(t, "read", receivedBody["status"])
	assert.Equal(t, "wamid.abc123", receivedBody["message_id"])
}

func TestSplitMessage_ShortText(t *testing.T) {
	chunks := splitMessage("hello", 100)

	require.Len(t, chunks, 1)
	assert.Equal(t, "hello", chunks[0])
}

func TestSplitMessage_SplitsAtNewline(t *testing.T) {
	text := strings.Repeat("a", 60) + "\n" + strings.Repeat("b", 30)
	chunks := splitMessage(text, 80)

	require.Len(t, chunks, 2)
	assert.Equal(t, strings.Repeat("a", 60)+"\n", chunks[0])
	assert.Equal(t, strings.Repeat("b", 30), chunks[1])
}

func TestSplitMessage_SplitsAtSpace(t *testing.T) {
	text := strings.Repeat("a", 60) + " " + strings.Repeat("b", 30)
	chunks := splitMessage(text, 80)

	require.Len(t, chunks, 2)
	assert.Equal(t, strings.Repeat("a", 60)+" ", chunks[0])
	assert.Equal(t, strings.Repeat("b", 30), chunks[1])
}

func TestSplitMessage_HardCut(t *testing.T) {
	text := strings.Repeat("a", 200)
	chunks := splitMessage(text, 100)

	require.Len(t, chunks, 2)
	assert.Len(t, chunks[0], 100)
	assert.Len(t, chunks[1], 100)
}

func TestWhatsApp_DownloadMedia_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "media-download-url") {
			// Step 2: return binary data
			w.Write([]byte("audio-binary-data"))
			return
		}
		// Step 1: return media info
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"url":       "http://" + r.Host + "/media-download-url",
			"mime_type": "audio/ogg",
		})
	}))
	defer srv.Close()

	c := NewWhatsAppClient("phone-123", "test-token")
	c.httpClient = &http.Client{
		Transport: newRedirectTransport(srv.URL),
	}

	data, mime, err := c.DownloadMedia("media-123")

	require.NoError(t, err)
	assert.Equal(t, "audio/ogg", mime)
	assert.Equal(t, []byte("audio-binary-data"), data)
}

func TestWhatsApp_DownloadMedia_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer srv.Close()

	c := NewWhatsAppClient("phone-123", "test-token")
	c.httpClient = &http.Client{
		Transport: newRedirectTransport(srv.URL),
	}

	_, _, err := c.DownloadMedia("bad-id")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "whatsapp: media info error 404")
}

func TestWhatsApp_SendTyping_NoOp(t *testing.T) {
	c := NewWhatsAppClient("phone-123", "test-token")

	err := c.SendTyping("+5491100001111")

	assert.NoError(t, err)
}

func TestWhatsApp_Name(t *testing.T) {
	c := NewWhatsAppClient("phone-123", "test-token")

	assert.Equal(t, "whatsapp", c.Name())
}

func TestWhatsApp_AckMessage_CallsMarkAsRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewWhatsAppClient("phone-123", "test-token")
	c.httpClient = &http.Client{
		Transport: newRedirectTransport(srv.URL),
	}

	err := c.AckMessage("wamid.abc123")

	assert.NoError(t, err)
}
