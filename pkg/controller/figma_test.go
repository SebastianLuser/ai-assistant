package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jarvis/clients"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
)

func newMockFigmaServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/images/"):
			json.NewEncoder(w).Encode(map[string]any{
				"images": map[string]string{"node1": "https://img.example.com/1.png"},
			})
		case strings.HasSuffix(r.URL.Path, "/comments"):
			json.NewEncoder(w).Encode(map[string]any{
				"comments": []map[string]any{
					{"id": "c1", "message": "Nice!", "created_at": "2026-03-10", "user": map[string]string{"handle": "user1"}},
				},
			})
		case strings.HasPrefix(r.URL.Path, "/projects/"):
			json.NewEncoder(w).Encode(map[string]any{
				"files": []map[string]string{
					{"key": "f1", "name": "Design", "thumbnail_url": "https://thumb.example.com", "last_modified": "2026-03-10"},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/components"):
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{
					"components": []map[string]string{
						{"key": "comp1", "name": "Button", "description": "A button"},
					},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/nodes"):
			json.NewEncoder(w).Encode(map[string]any{
				"nodes": map[string]any{
					"node1": map[string]any{
						"document":   map[string]string{"id": "node1"},
						"components": map[string]any{},
					},
				},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"name": "Test File", "lastModified": "2026-03-10", "thumbnailUrl": "https://thumb.example.com", "version": "1",
			})
		}
	}))
}

func TestFigmaController_GetFile_MissingFileKey(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetFile(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetFile_Success(t *testing.T) {
	srv := newMockFigmaServer()
	defer srv.Close()
	client := clients.NewFigmaClientWithBaseURL("token", srv.URL)
	ctrl := NewFigmaController(client)
	req := test.NewMockRequest().WithParam("file_key", "abc123")

	resp := ctrl.GetFile(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFigmaController_GetNodes_MissingIDs(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest().WithParam("file_key", "abc123")

	resp := ctrl.GetNodes(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetNodes_Success(t *testing.T) {
	srv := newMockFigmaServer()
	defer srv.Close()
	client := clients.NewFigmaClientWithBaseURL("token", srv.URL)
	ctrl := NewFigmaController(client)
	req := test.NewMockRequest().WithParam("file_key", "abc123").WithQuery("ids", "node1")

	resp := ctrl.GetNodes(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFigmaController_GetImages_MissingFileKey(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetImages(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetImages_MissingIDs(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest().WithParam("file_key", "abc123")

	resp := ctrl.GetImages(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetImages_Success(t *testing.T) {
	srv := newMockFigmaServer()
	defer srv.Close()
	client := clients.NewFigmaClientWithBaseURL("token", srv.URL)
	ctrl := NewFigmaController(client)
	req := test.NewMockRequest().WithParam("file_key", "abc123").WithQuery("ids", "node1")

	resp := ctrl.GetImages(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFigmaController_GetImages_WithScaleAndFormat(t *testing.T) {
	srv := newMockFigmaServer()
	defer srv.Close()
	client := clients.NewFigmaClientWithBaseURL("token", srv.URL)
	ctrl := NewFigmaController(client)
	req := test.NewMockRequest().WithParam("file_key", "abc123").WithQuery("ids", "node1").WithQuery("format", "png").WithQuery("scale", "3.0")

	resp := ctrl.GetImages(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFigmaController_GetComments_MissingFileKey(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetComments(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetComments_Success(t *testing.T) {
	srv := newMockFigmaServer()
	defer srv.Close()
	client := clients.NewFigmaClientWithBaseURL("token", srv.URL)
	ctrl := NewFigmaController(client)
	req := test.NewMockRequest().WithParam("file_key", "abc123")

	resp := ctrl.GetComments(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFigmaController_GetProjectFiles_MissingProjectID(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetProjectFiles(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetProjectFiles_Success(t *testing.T) {
	srv := newMockFigmaServer()
	defer srv.Close()
	client := clients.NewFigmaClientWithBaseURL("token", srv.URL)
	ctrl := NewFigmaController(client)
	req := test.NewMockRequest().WithParam("project_id", "proj1")

	resp := ctrl.GetProjectFiles(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFigmaController_GetComponents_MissingFileKey(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetComponents(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetComponents_Success(t *testing.T) {
	srv := newMockFigmaServer()
	defer srv.Close()
	client := clients.NewFigmaClientWithBaseURL("token", srv.URL)
	ctrl := NewFigmaController(client)
	req := test.NewMockRequest().WithParam("file_key", "abc123")

	resp := ctrl.GetComponents(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
