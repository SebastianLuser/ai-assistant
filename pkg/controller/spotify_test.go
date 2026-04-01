package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jarvis/clients"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
)

func newMockSpotifyServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/me/player/currently-playing":
			json.NewEncoder(w).Encode(map[string]any{
				"is_playing": true,
				"item": map[string]any{
					"name":          "Song",
					"album":         map[string]string{"name": "Album"},
					"artists":       []map[string]string{{"name": "Artist"}},
					"external_urls": map[string]string{"spotify": "https://open.spotify.com/track/1"},
				},
			})
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestSpotifyController_New(t *testing.T) {
	ctrl := NewSpotifyController(nil)

	assert.NotNil(t, ctrl)
}

func TestSpotifyController_GetCurrentlyPlaying_Success(t *testing.T) {
	srv := newMockSpotifyServer()
	defer srv.Close()
	client := clients.NewSpotifyClientWithBaseURL("token", srv.URL)
	ctrl := NewSpotifyController(client)
	req := test.NewMockRequest()

	resp := ctrl.GetCurrentlyPlaying(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestSpotifyController_Play_Success(t *testing.T) {
	srv := newMockSpotifyServer()
	defer srv.Close()
	client := clients.NewSpotifyClientWithBaseURL("token", srv.URL)
	ctrl := NewSpotifyController(client)
	req := test.NewMockRequest()

	resp := ctrl.Play(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestSpotifyController_Pause_Success(t *testing.T) {
	srv := newMockSpotifyServer()
	defer srv.Close()
	client := clients.NewSpotifyClientWithBaseURL("token", srv.URL)
	ctrl := NewSpotifyController(client)
	req := test.NewMockRequest()

	resp := ctrl.Pause(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestSpotifyController_Next_Success(t *testing.T) {
	srv := newMockSpotifyServer()
	defer srv.Close()
	client := clients.NewSpotifyClientWithBaseURL("token", srv.URL)
	ctrl := NewSpotifyController(client)
	req := test.NewMockRequest()

	resp := ctrl.Next(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
