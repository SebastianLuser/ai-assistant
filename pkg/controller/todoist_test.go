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

const (
	validTaskBody   = `{"content":"Buy milk"}`
	emptyTaskBody   = `{"content":""}`
	taskWithDate    = `{"content":"Buy milk","due_date":"2026-03-15"}`
	taskBadDate     = `{"content":"Buy milk","due_date":"15/03/2026"}`
	invalidTaskJSON = `{nah`
)

func newMockTodoistServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rest/v2/tasks":
			json.NewEncoder(w).Encode([]map[string]any{
				{"id": "1", "content": "Buy milk", "priority": 1, "is_completed": false, "url": "https://todoist.com/1"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/rest/v2/tasks":
			json.NewEncoder(w).Encode(map[string]any{
				"id": "2", "content": "Buy milk", "priority": 1, "is_completed": false, "url": "https://todoist.com/2",
			})
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestTodoistController_CreateTask_InvalidJSON(t *testing.T) {
	ctrl := NewTodoistController(nil)
	req := test.NewMockRequest().WithBody(invalidTaskJSON)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestTodoistController_CreateTask_EmptyContent(t *testing.T) {
	ctrl := NewTodoistController(nil)
	req := test.NewMockRequest().WithBody(emptyTaskBody)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: content is required", errorFromBody(t, resp.Body))
}

func TestTodoistController_CreateTask_InvalidDateFormat(t *testing.T) {
	ctrl := NewTodoistController(nil)
	req := test.NewMockRequest().WithBody(taskBadDate)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: due_date must be in YYYY-MM-DD format", errorFromBody(t, resp.Body))
}

func TestTodoistController_CompleteTask_MissingID(t *testing.T) {
	ctrl := NewTodoistController(nil)
	req := test.NewMockRequest()

	resp := ctrl.CompleteTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "id is required", errorFromBody(t, resp.Body))
}

func TestTodoistController_CompleteTask_EmptyID(t *testing.T) {
	ctrl := NewTodoistController(nil)
	req := test.NewMockRequest().WithParam("id", "")

	resp := ctrl.CompleteTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestTodoistController_GetTasks_Success(t *testing.T) {
	srv := newMockTodoistServer()
	defer srv.Close()
	client := clients.NewTodoistClientWithBaseURL("token", srv.URL)
	ctrl := NewTodoistController(client)
	req := test.NewMockRequest()

	resp := ctrl.GetTasks(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestTodoistController_CreateTask_Success(t *testing.T) {
	srv := newMockTodoistServer()
	defer srv.Close()
	client := clients.NewTodoistClientWithBaseURL("token", srv.URL)
	ctrl := NewTodoistController(client)
	req := test.NewMockRequest().WithBody(validTaskBody)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}

func TestTodoistController_CompleteTask_Success(t *testing.T) {
	srv := newMockTodoistServer()
	defer srv.Close()
	client := clients.NewTodoistClientWithBaseURL("token", srv.URL)
	ctrl := NewTodoistController(client)
	req := test.NewMockRequest().WithParam("id", "1")

	resp := ctrl.CompleteTask(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
