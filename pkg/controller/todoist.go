package controller

import (
	"net/http"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

// TodoistController handles Todoist API endpoints.
type TodoistController struct {
	client *clients.TodoistClient
}

// NewTodoistController creates a new TodoistController.
func NewTodoistController(client *clients.TodoistClient) *TodoistController {
	return &TodoistController{client: client}
}

// GetTasks returns all active tasks.
func (c *TodoistController) GetTasks(req web.Request) web.Response {
	tasks, err := c.client.GetTasks()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.TodoistTaskListResponse{Error: err.Error()})
	}

	domainTasks := make([]domain.TodoistTask, len(tasks))
	for i, t := range tasks {
		domainTasks[i] = domain.TodoistTask{
			ID: t.ID, Content: t.Content, DueDate: t.DueDate,
			Priority: t.Priority, Completed: t.Completed, URL: t.URL,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.TodoistTaskListResponse{
		Success: true, Tasks: domainTasks,
	})
}

// CreateTask creates a new task.
func (c *TodoistController) CreateTask(req web.Request) web.Response {
	var payload domain.TodoistCreateTaskRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.TodoistTaskResponse{Error: "invalid body"})
	}

	if err := payload.Validate(); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.TodoistTaskResponse{Error: err.Error()})
	}

	task, err := c.client.CreateTask(payload.Content, payload.DueDate)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.TodoistTaskResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.TodoistTaskResponse{
		Success: true,
		Task: &domain.TodoistTask{
			ID: task.ID, Content: task.Content, DueDate: task.DueDate,
			Priority: task.Priority, Completed: task.Completed, URL: task.URL,
		},
	})
}

// CompleteTask marks a task as completed.
func (c *TodoistController) CompleteTask(req web.Request) web.Response {
	taskID, ok := req.Param(domain.PathParamID)
	if !ok || taskID == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.TodoistActionResponse{Error: "id is required"})
	}

	if err := c.client.CompleteTask(taskID); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.TodoistActionResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.TodoistActionResponse{Success: true})
}
