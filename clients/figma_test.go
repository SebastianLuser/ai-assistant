package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFigmaTestServer(handler http.HandlerFunc) (*httptest.Server, *FigmaClient) {
	srv := httptest.NewServer(handler)
	c := NewFigmaClient("figma-test-token")
	c.httpClient = srv.Client()
	return srv, c
}

// newFigmaClient creates a FigmaClient that points doRequest at the test server.
func newFigmaClientWithURL(token, baseURL string) *FigmaClient {
	c := NewFigmaClient(token)
	c.httpClient = &http.Client{
		Transport: newRedirectTransport(baseURL),
	}
	return c
}

func TestFigma_GetFile_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "figma-test-token", r.Header.Get("X-Figma-Token"))
		assert.Contains(t, r.URL.Path, "/files/abc123")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FigmaFile{
			Name:         "My Design",
			LastModified: "2025-01-15",
			Version:      "v42",
		})
	}))
	defer srv.Close()

	c := newFigmaClientWithURL("figma-test-token", srv.URL)

	file, err := c.GetFile("abc123")

	require.NoError(t, err)
	assert.Equal(t, "My Design", file.Name)
	assert.Equal(t, "2025-01-15", file.LastModified)
	assert.Equal(t, "v42", file.Version)
}

func TestFigma_GetFile_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"status":404,"err":"Not found"}`))
	}))
	defer srv.Close()

	c := newFigmaClientWithURL("token", srv.URL)

	_, err := c.GetFile("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "figma: api error 404")
}

func TestFigma_GetImages_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/images/file1")
		assert.Contains(t, r.URL.RawQuery, "ids=1:2")
		assert.Contains(t, r.URL.RawQuery, "format=png")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"images": map[string]string{
				"1:2": "https://cdn.figma.com/image1.png",
			},
		})
	}))
	defer srv.Close()

	c := newFigmaClientWithURL("token", srv.URL)

	images, err := c.GetImages("file1", []string{"1:2"}, "png", 2)

	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, "1:2", images[0].NodeID)
	assert.Equal(t, "https://cdn.figma.com/image1.png", images[0].ImageURL)
}

func TestFigma_GetImages_DefaultFormatAndScale(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "format=png")
		assert.Contains(t, r.URL.RawQuery, "scale=2")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"images": map[string]string{},
		})
	}))
	defer srv.Close()

	c := newFigmaClientWithURL("token", srv.URL)

	_, err := c.GetImages("file1", []string{"1:1"}, "", 0)

	require.NoError(t, err)
}

func TestFigma_GetImages_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	}))
	defer srv.Close()

	c := newFigmaClientWithURL("token", srv.URL)

	_, err := c.GetImages("file1", []string{"1:1"}, "svg", 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "figma: api error 403")
}

func TestFigma_GetComments_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/files/design1/comments")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"comments": []map[string]any{
				{
					"id":         "c1",
					"message":    "Fix the alignment",
					"created_at": "2025-03-01T10:00:00Z",
					"user":       map[string]string{"handle": "alice"},
				},
				{
					"id":         "c2",
					"message":    "Looks good",
					"created_at": "2025-03-02T14:00:00Z",
					"user":       map[string]string{"handle": "bob"},
				},
			},
		})
	}))
	defer srv.Close()

	c := newFigmaClientWithURL("token", srv.URL)

	comments, err := c.GetComments("design1")

	require.NoError(t, err)
	require.Len(t, comments, 2)
	assert.Equal(t, "Fix the alignment", comments[0].Message)
	assert.Equal(t, "alice", comments[0].User.Handle)
	assert.Equal(t, "Looks good", comments[1].Message)
}

func TestFigma_GetComments_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid token"))
	}))
	defer srv.Close()

	c := newFigmaClientWithURL("bad-token", srv.URL)

	_, err := c.GetComments("design1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "figma: api error 401")
}

func TestFigma_GetComments_EmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"comments": []any{},
		})
	}))
	defer srv.Close()

	c := newFigmaClientWithURL("token", srv.URL)

	comments, err := c.GetComments("empty-file")

	require.NoError(t, err)
	assert.Empty(t, comments)
}
