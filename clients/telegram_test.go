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

func newTelegramTestClient(srvURL string) *TelegramClient {
	c := NewTelegramClient("test-bot-token")
	// Replace the base URL so all requests go to the test server.
	// We use a custom transport that redirects.
	c.httpClient = &http.Client{
		Transport: newRedirectTransport(srvURL),
	}
	return c
}

func TestTelegram_SendTextMessage_Success(t *testing.T) {
	var receivedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	err := c.SendTextMessage("12345", "hello from test")

	require.NoError(t, err)
	assert.Equal(t, "12345", receivedBody["chat_id"])
	assert.Equal(t, "hello from test", receivedBody["text"])
	assert.Equal(t, "Markdown", receivedBody["parse_mode"])
}

func TestTelegram_SendTextMessage_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"ok":false,"description":"bot was blocked"}`))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	err := c.SendTextMessage("12345", "hello")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "telegram: api error 403")
	assert.Contains(t, err.Error(), "bot was blocked")
}

func TestTelegram_SendTextMessage_LongMessage_Splits(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)
	longText := strings.Repeat("x", 5000)

	err := c.SendTextMessage("12345", longText)

	require.NoError(t, err)
	assert.Equal(t, 2, callCount, "should split into 2 messages")
}

func TestTelegram_SendChatAction_Success(t *testing.T) {
	var receivedBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	err := c.SendChatAction("12345", "typing")

	require.NoError(t, err)
	assert.Equal(t, "12345", receivedBody["chat_id"])
	assert.Equal(t, "typing", receivedBody["action"])
}

func TestTelegram_SendChatAction_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	err := c.SendChatAction("12345", "typing")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "telegram: action error 500")
}

func TestTelegram_SendTyping_CallsSendChatAction(t *testing.T) {
	var receivedBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	err := c.SendTyping("67890")

	require.NoError(t, err)
	assert.Equal(t, "typing", receivedBody["action"])
}

func TestTelegram_DownloadMedia_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getFile") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"result": map[string]string{
					"file_path": "voice/file_42.ogg",
				},
			})
			return
		}
		// File download endpoint
		w.Write([]byte("ogg-binary-data"))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	data, mime, err := c.DownloadMedia("file-id-123")

	require.NoError(t, err)
	assert.Equal(t, "audio/ogg", mime)
	assert.Equal(t, []byte("ogg-binary-data"), data)
}

func TestTelegram_DownloadMedia_MP3Extension(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getFile") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"result": map[string]string{
					"file_path": "music/song.mp3",
				},
			})
			return
		}
		w.Write([]byte("mp3-data"))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	_, mime, err := c.DownloadMedia("file-id-mp3")

	require.NoError(t, err)
	assert.Equal(t, "audio/mpeg", mime)
}

func TestTelegram_DownloadMedia_M4AExtension(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getFile") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"result": map[string]string{
					"file_path": "audio/clip.m4a",
				},
			})
			return
		}
		w.Write([]byte("m4a-data"))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	_, mime, err := c.DownloadMedia("file-id-m4a")

	require.NoError(t, err)
	assert.Equal(t, "audio/mp4", mime)
}

func TestTelegram_DownloadMedia_GetFileError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("file not found"))
	}))
	defer srv.Close()

	c := newTelegramTestClient(srv.URL)

	_, _, err := c.DownloadMedia("bad-file-id")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "telegram: get file error 404")
}

func TestTelegram_Name(t *testing.T) {
	c := NewTelegramClient("token")

	assert.Equal(t, "telegram", c.Name())
}

func TestTelegram_AckMessage_NoOp(t *testing.T) {
	c := NewTelegramClient("token")

	err := c.AckMessage("msg-123")

	assert.NoError(t, err)
}
