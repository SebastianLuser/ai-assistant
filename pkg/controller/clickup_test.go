package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"asistente/clients"
	"asistente/test"

	"github.com/stretchr/testify/assert"
)

func newMockClickUpServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v2/user":
			json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 123},
			})
		case strings.Contains(r.URL.Path, "/team/") && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]any{
				"tasks": []map[string]any{
					{
						"id": "t1", "name": "Task 1", "url": "https://clickup.com/t1",
						"status":    map[string]string{"status": "open"},
						"assignees": []map[string]string{{"username": "user1"}},
					},
				},
			})
		case strings.Contains(r.URL.Path, "/list/") && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]any{
				"id": "t2", "name": "New Task", "url": "https://clickup.com/t2",
				"status":    map[string]string{"status": "open"},
				"assignees": []map[string]string{},
			})
		case r.Method == http.MethodPut:
			json.NewEncoder(w).Encode(map[string]string{"id": "t1"})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"id": "t1", "name": "Task 1", "url": "https://clickup.com/t1",
				"status":    map[string]string{"status": "open"},
				"assignees": []map[string]string{{"username": "user1"}},
			})
		}
	}))
}

func TestClickUpController_CreateTask_InvalidJSON(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithBody(`{bad`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestClickUpController_CreateTask_MissingListID(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithBody(`{"name":"Task 1"}`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: list_id is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_CreateTask_MissingName(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithBody(`{"list_id":"123"}`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: name is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_CreateTask_NameTooLong(t *testing.T) {
	ctrl := NewClickUpController(nil)
	long := strings.Repeat("a", 501)
	req := test.NewMockRequest().WithBody(`{"list_id":"123","name":"` + long + `"}`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: name exceeds 500 characters", errorFromBody(t, resp.Body))
}

func TestClickUpController_GetTask_MissingID(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "id is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_UpdateTaskStatus_MissingID(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest()

	resp := ctrl.UpdateTaskStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "id is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_UpdateTaskStatus_InvalidJSON(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithParam("id", "abc123").WithBody(`{bad`)

	resp := ctrl.UpdateTaskStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestClickUpController_UpdateTaskStatus_EmptyStatus(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithParam("id", "abc123").WithBody(`{"status":""}`)

	resp := ctrl.UpdateTaskStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: status is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_GetMyTasks_Success(t *testing.T) {
	srv := newMockClickUpServer()
	defer srv.Close()
	client := clients.NewClickUpClientWithBaseURL("token", "team1", srv.URL)
	ctrl := NewClickUpController(client)
	req := test.NewMockRequest()

	resp := ctrl.GetMyTasks(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestClickUpController_GetTask_Success(t *testing.T) {
	srv := newMockClickUpServer()
	defer srv.Close()
	client := clients.NewClickUpClientWithBaseURL("token", "team1", srv.URL)
	ctrl := NewClickUpController(client)
	req := test.NewMockRequest().WithParam("id", "t1")

	resp := ctrl.GetTask(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestClickUpController_CreateTask_Success(t *testing.T) {
	srv := newMockClickUpServer()
	defer srv.Close()
	client := clients.NewClickUpClientWithBaseURL("token", "team1", srv.URL)
	ctrl := NewClickUpController(client)
	req := test.NewMockRequest().WithBody(`{"list_id":"list1","name":"New Task","description":"desc"}`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}

func TestClickUpController_UpdateTaskStatus_Success(t *testing.T) {
	srv := newMockClickUpServer()
	defer srv.Close()
	client := clients.NewClickUpClientWithBaseURL("token", "team1", srv.URL)
	ctrl := NewClickUpController(client)
	req := test.NewMockRequest().WithParam("id", "t1").WithBody(`{"status":"done"}`)

	resp := ctrl.UpdateTaskStatus(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
