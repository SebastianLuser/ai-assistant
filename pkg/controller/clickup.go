package controller

import (
	"net/http"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

// ClickUpController handles ClickUp API endpoints.
type ClickUpController struct {
	client *clients.ClickUpClient
}

// NewClickUpController creates a new ClickUpController.
func NewClickUpController(client *clients.ClickUpClient) *ClickUpController {
	return &ClickUpController{client: client}
}

// GetMyTasks returns tasks assigned to the current user.
func (c *ClickUpController) GetMyTasks(req web.Request) web.Response {
	tasks, err := c.client.GetMyTasks()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ClickUpTaskListResponse{Error: err.Error()})
	}

	domainTasks := make([]domain.ClickUpTask, len(tasks))
	for i, t := range tasks {
		domainTasks[i] = domain.ClickUpTask{
			ID: t.ID, Name: t.Name, Status: t.Status,
			Assignees: t.Assignees, URL: t.URL, DueDate: t.DueDate,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.ClickUpTaskListResponse{
		Success: true, Tasks: domainTasks,
	})
}

// GetTask returns a single task by ID.
func (c *ClickUpController) GetTask(req web.Request) web.Response {
	taskID, ok := req.Param(domain.PathParamID)
	if !ok || taskID == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ClickUpTaskResponse{Error: "id is required"})
	}

	task, err := c.client.GetTask(taskID)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ClickUpTaskResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.ClickUpTaskResponse{
		Success: true,
		Task: &domain.ClickUpTask{
			ID: task.ID, Name: task.Name, Status: task.Status,
			Assignees: task.Assignees, URL: task.URL, DueDate: task.DueDate,
		},
	})
}

// CreateTask creates a new task.
func (c *ClickUpController) CreateTask(req web.Request) web.Response {
	var payload domain.ClickUpCreateTaskRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ClickUpTaskResponse{Error: "invalid body"})
	}

	if err := payload.Validate(); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ClickUpTaskResponse{Error: err.Error()})
	}

	task, err := c.client.CreateTask(payload.ListID, payload.Name, payload.Description)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ClickUpTaskResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.ClickUpTaskResponse{
		Success: true,
		Task: &domain.ClickUpTask{
			ID: task.ID, Name: task.Name, Status: task.Status,
			Assignees: task.Assignees, URL: task.URL, DueDate: task.DueDate,
		},
	})
}

// UpdateTaskStatus updates the status of a task.
func (c *ClickUpController) UpdateTaskStatus(req web.Request) web.Response {
	taskID, ok := req.Param(domain.PathParamID)
	if !ok || taskID == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ClickUpActionResponse{Error: "id is required"})
	}

	var payload domain.ClickUpUpdateStatusRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ClickUpActionResponse{Error: "invalid body"})
	}

	if err := payload.Validate(); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ClickUpActionResponse{Error: err.Error()})
	}

	if err := c.client.UpdateTaskStatus(taskID, payload.Status); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ClickUpActionResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.ClickUpActionResponse{Success: true})
}
