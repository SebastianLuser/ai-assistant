package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	spotifyBaseURL        = "https://api.spotify.com"
	spotifyDefaultTimeout = 15 * time.Second

	spotifyPathCurrentlyPlaying = "/v1/me/player/currently-playing"
	spotifyPathPlay             = "/v1/me/player/play"
	spotifyPathPause            = "/v1/me/player/pause"
	spotifyPathNext             = "/v1/me/player/next"
)

// SpotifyTrack represents a Spotify track.
type SpotifyTrack struct {
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	URL       string `json:"url"`
	IsPlaying bool   `json:"is_playing"`
}

// SpotifyClient is the Spotify API client.
type SpotifyClient struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

// NewSpotifyClient creates a new Spotify API client.
func NewSpotifyClient(accessToken string) *SpotifyClient {
	return &SpotifyClient{
		accessToken: accessToken,
		baseURL:     spotifyBaseURL,
		httpClient:  &http.Client{Timeout: spotifyDefaultTimeout},
	}
}

// NewSpotifyClientWithBaseURL creates a Spotify client pointing at a custom base URL (for testing).
func NewSpotifyClientWithBaseURL(accessToken, baseURL string) *SpotifyClient {
	c := NewSpotifyClient(accessToken)
	c.baseURL = baseURL
	return c
}

// GetCurrentlyPlaying returns the currently playing track.
func (c *SpotifyClient) GetCurrentlyPlaying() (*SpotifyTrack, error) {
	resp, err := c.doRequest(http.MethodGet, spotifyPathCurrentlyPlaying)
	if err != nil {
		return nil, err
	}

	// 204 means nothing is playing
	if resp == nil {
		return nil, nil
	}

	var raw struct {
		IsPlaying bool `json:"is_playing"`
		Item      struct {
			Name    string `json:"name"`
			Album   struct {
				Name string `json:"name"`
			} `json:"album"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			ExternalURLs struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
		} `json:"item"`
	}

	if err := json.Unmarshal(resp, &raw); err != nil {
		return nil, fmt.Errorf("spotify: parse currently playing: %w", err)
	}

	artist := ""
	if len(raw.Item.Artists) > 0 {
		artist = raw.Item.Artists[0].Name
	}

	return &SpotifyTrack{
		Name:      raw.Item.Name,
		Artist:    artist,
		Album:     raw.Item.Album.Name,
		URL:       raw.Item.ExternalURLs.Spotify,
		IsPlaying: raw.IsPlaying,
	}, nil
}

// Play resumes playback.
func (c *SpotifyClient) Play() error {
	_, err := c.doRequest(http.MethodPut, spotifyPathPlay)
	return err
}

// Pause pauses playback.
func (c *SpotifyClient) Pause() error {
	_, err := c.doRequest(http.MethodPut, spotifyPathPause)
	return err
}

// Next skips to the next track.
func (c *SpotifyClient) Next() error {
	_, err := c.doRequest(http.MethodPost, spotifyPathNext)
	return err
}

func (c *SpotifyClient) doRequest(method, path string) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("spotify: create request: %w", err)
	}

	req.Header.Set(headerAuthorization, "Bearer "+c.accessToken)
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("spotify: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("spotify: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("spotify: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
