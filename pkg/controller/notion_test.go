package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"asistente/clients"
	"asistente/test"

	"github.com/stretchr/testify/assert"
)

const (
	validNotionBody   = `{"title":"My Page","content":"Hello"}`
	emptyTitleBody    = `{"title":"","content":"Hello"}`
	invalidNotionJSON = `{bad`
)

func newMockNotionServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]string{"id": "page-123"})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"id":               "page-123",
				"last_edited_time": "2026-03-10T10:00:00Z",
				"properties": map[string]any{
					"title": map[string]any{
						"title": []map[string]string{
							{"plain_text": "My Page"},
						},
					},
				},
			})
		}
	}))
}

func TestNotionController_CreatePage_InvalidJSON(t *testing.T) {
	ctrl := NewNotionController(nil, "parent-id")
	req := test.NewMockRequest().WithBody(invalidNotionJSON)

	resp := ctrl.CreatePage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestNotionController_CreatePage_EmptyTitle(t *testing.T) {
	ctrl := NewNotionController(nil, "parent-id")
	req := test.NewMockRequest().WithBody(emptyTitleBody)

	resp := ctrl.CreatePage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "title is required", errorFromBody(t, resp.Body))
}

func TestNotionController_CreatePage_Success(t *testing.T) {
	srv := newMockNotionServer()
	defer srv.Close()
	client := clients.NewNotionClientWithBaseURL("key", srv.URL)
	ctrl := NewNotionController(client, "parent-id")
	req := test.NewMockRequest().WithBody(validNotionBody)

	resp := ctrl.CreatePage(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}

func TestNotionController_GetPage_MissingID(t *testing.T) {
	ctrl := NewNotionController(nil, "parent-id")
	req := test.NewMockRequest()

	resp := ctrl.GetPage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "id is required", errorFromBody(t, resp.Body))
}

func TestNotionController_GetPage_Success(t *testing.T) {
	srv := newMockNotionServer()
	defer srv.Close()
	client := clients.NewNotionClientWithBaseURL("key", srv.URL)
	ctrl := NewNotionController(client, "parent-id")
	req := test.NewMockRequest().WithParam("id", "page-123")

	resp := ctrl.GetPage(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
