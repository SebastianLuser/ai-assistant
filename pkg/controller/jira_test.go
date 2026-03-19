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

func newMockJiraServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/rest/api/3/search":
			json.NewEncoder(w).Encode(map[string]any{
				"issues": []map[string]any{
					{
						"key": "PROJ-1",
						"fields": map[string]any{
							"summary":   "Fix bug",
							"status":    map[string]string{"name": "Open"},
							"assignee":  map[string]string{"displayName": "Sebas"},
							"issuetype": map[string]string{"name": "Bug"},
						},
					},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue":
			json.NewEncoder(w).Encode(map[string]string{"key": "PROJ-2"})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"key": "PROJ-1",
				"fields": map[string]any{
					"summary":   "Fix bug",
					"status":    map[string]string{"name": "Open"},
					"assignee":  map[string]string{"displayName": "Sebas"},
					"issuetype": map[string]string{"name": "Bug"},
				},
			})
		}
	}))
}

func TestJiraController_GetIssue_MissingKey(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_GetIssue_Success(t *testing.T) {
	srv := newMockJiraServer()
	defer srv.Close()
	client := clients.NewJiraClient(srv.URL, "test@test.com", "token")
	ctrl := NewJiraController(client)
	req := test.NewMockRequest().WithParam("key", "PROJ-1")

	resp := ctrl.GetIssue(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestJiraController_GetMyIssues_Success(t *testing.T) {
	srv := newMockJiraServer()
	defer srv.Close()
	client := clients.NewJiraClient(srv.URL, "test@test.com", "token")
	ctrl := NewJiraController(client)
	req := test.NewMockRequest()

	resp := ctrl.GetMyIssues(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestJiraController_CreateIssue_InvalidJSON(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest().WithBody(`{invalid`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_CreateIssue_MissingProjectKey(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest().WithBody(`{"project_key":"","summary":"bug","issue_type":"Bug"}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_CreateIssue_MissingSummary(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest().WithBody(`{"project_key":"PROJ","summary":"","issue_type":"Bug"}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_CreateIssue_MissingIssueType(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest().WithBody(`{"project_key":"PROJ","summary":"bug","issue_type":""}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_CreateIssue_Success(t *testing.T) {
	srv := newMockJiraServer()
	defer srv.Close()
	client := clients.NewJiraClient(srv.URL, "test@test.com", "token")
	ctrl := NewJiraController(client)
	req := test.NewMockRequest().WithBody(`{"project_key":"PROJ","summary":"new bug","issue_type":"Bug"}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}
