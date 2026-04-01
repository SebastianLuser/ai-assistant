package controller

import (
	"net/http"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

// SpotifyController handles Spotify API endpoints.
type SpotifyController struct {
	client *clients.SpotifyClient
}

// NewSpotifyController creates a new SpotifyController.
func NewSpotifyController(client *clients.SpotifyClient) *SpotifyController {
	return &SpotifyController{client: client}
}

// GetCurrentlyPlaying returns the currently playing track.
func (c *SpotifyController) GetCurrentlyPlaying(req web.Request) web.Response {
	track, err := c.client.GetCurrentlyPlaying()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.SpotifyTrackResponse{Error: err.Error()})
	}

	if track == nil {
		return web.NewJSONResponse(http.StatusOK, domain.SpotifyTrackResponse{
			Success: true, Track: nil,
		})
	}

	return web.NewJSONResponse(http.StatusOK, domain.SpotifyTrackResponse{
		Success: true,
		Track: &domain.SpotifyTrack{
			Name: track.Name, Artist: track.Artist, Album: track.Album,
			URL: track.URL, IsPlaying: track.IsPlaying,
		},
	})
}

// Play resumes playback.
func (c *SpotifyController) Play(req web.Request) web.Response {
	if err := c.client.Play(); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.SpotifyActionResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.SpotifyActionResponse{Success: true})
}

// Pause pauses playback.
func (c *SpotifyController) Pause(req web.Request) web.Response {
	if err := c.client.Pause(); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.SpotifyActionResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.SpotifyActionResponse{Success: true})
}

// Next skips to the next track.
func (c *SpotifyController) Next(req web.Request) web.Response {
	if err := c.client.Next(); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.SpotifyActionResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.SpotifyActionResponse{Success: true})
}
