package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	clickupBaseURL        = "https://api.clickup.com"
	clickupDefaultTimeout = 15 * time.Second

	clickupPathTeamTask = "/api/v2/team/%s/task"
	clickupPathListTask = "/api/v2/list/%s/task"
	clickupPathTask     = "/api/v2/task/%s"
	clickupPathUser     = "/api/v2/user"
)

// ClickUpTask represents a ClickUp task.
type ClickUpTask struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Assignees []string `json:"assignees"`
	URL       string   `json:"url"`
	DueDate   string   `json:"due_date,omitempty"`
}

// ClickUpClient is the ClickUp API client.
type ClickUpClient struct {
	apiToken   string
	teamID     string
	baseURL    string
	httpClient *http.Client
}

// NewClickUpClient creates a new ClickUp API client.
func NewClickUpClient(apiToken, teamID string) *ClickUpClient {
	return &ClickUpClient{
		apiToken:   apiToken,
		teamID:     teamID,
		baseURL:    clickupBaseURL,
		httpClient: &http.Client{Timeout: clickupDefaultTimeout},
	}
}

// NewClickUpClientWithBaseURL creates a ClickUp client pointing at a custom base URL (for testing).
func NewClickUpClientWithBaseURL(apiToken, teamID, baseURL string) *ClickUpClient {
	c := NewClickUpClient(apiToken, teamID)
	c.baseURL = baseURL
	return c
}

// getMyUserID returns the authenticated user's ClickUp ID.
func (c *ClickUpClient) getMyUserID() (int64, error) {
	resp, err := c.doRequest(http.MethodGet, clickupPathUser, nil)
	if err != nil {
		return 0, err
	}

	var result struct {
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, fmt.Errorf("clickup: parse user: %w", err)
	}

	return result.User.ID, nil
}

// GetMyTasks returns tasks assigned to the current user in the configured team.
func (c *ClickUpClient) GetMyTasks() ([]ClickUpTask, error) {
	userID, err := c.getMyUserID()
	if err != nil {
		return nil, fmt.Errorf("clickup: get user id: %w", err)
	}

	path := fmt.Sprintf(clickupPathTeamTask, c.teamID) + fmt.Sprintf("?assignees[]=%d", userID)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tasks []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			URL    string `json:"url"`
			Status struct {
				Status string `json:"status"`
			} `json:"status"`
			Assignees []struct {
				Username string `json:"username"`
			} `json:"assignees"`
			DueDate string `json:"due_date"`
		} `json:"tasks"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("clickup: parse tasks: %w", err)
	}

	tasks := make([]ClickUpTask, len(result.Tasks))
	for i, raw := range result.Tasks {
		assignees := make([]string, len(raw.Assignees))
		for j, a := range raw.Assignees {
			assignees[j] = a.Username
		}
		tasks[i] = ClickUpTask{
			ID:        raw.ID,
			Name:      raw.Name,
			Status:    raw.Status.Status,
			Assignees: assignees,
			URL:       raw.URL,
			DueDate:   raw.DueDate,
		}
	}

	return tasks, nil
}

// GetTask returns a single task by ID.
func (c *ClickUpClient) GetTask(taskID string) (ClickUpTask, error) {
	path := fmt.Sprintf(clickupPathTask, taskID)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return ClickUpTask{}, err
	}

	var raw struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		Status struct {
			Status string `json:"status"`
		} `json:"status"`
		Assignees []struct {
			Username string `json:"username"`
		} `json:"assignees"`
		DueDate string `json:"due_date"`
	}

	if err := json.Unmarshal(resp, &raw); err != nil {
		return ClickUpTask{}, fmt.Errorf("clickup: parse task: %w", err)
	}

	assignees := make([]string, len(raw.Assignees))
	for j, a := range raw.Assignees {
		assignees[j] = a.Username
	}

	return ClickUpTask{
		ID:        raw.ID,
		Name:      raw.Name,
		Status:    raw.Status.Status,
		Assignees: assignees,
		URL:       raw.URL,
		DueDate:   raw.DueDate,
	}, nil
}

// CreateTask creates a new task in the given list.
func (c *ClickUpClient) CreateTask(listID, name, description string) (ClickUpTask, error) {
	path := fmt.Sprintf(clickupPathListTask, listID)
	body := map[string]string{
		"name":        name,
		"description": description,
	}

	resp, err := c.doRequest(http.MethodPost, path, body)
	if err != nil {
		return ClickUpTask{}, err
	}

	var raw struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		Status struct {
			Status string `json:"status"`
		} `json:"status"`
		Assignees []struct {
			Username string `json:"username"`
		} `json:"assignees"`
		DueDate string `json:"due_date"`
	}

	if err := json.Unmarshal(resp, &raw); err != nil {
		return ClickUpTask{}, fmt.Errorf("clickup: parse create response: %w", err)
	}

	assignees := make([]string, len(raw.Assignees))
	for i, a := range raw.Assignees {
		assignees[i] = a.Username
	}

	return ClickUpTask{
		ID:        raw.ID,
		Name:      raw.Name,
		Status:    raw.Status.Status,
		Assignees: assignees,
		URL:       raw.URL,
		DueDate:   raw.DueDate,
	}, nil
}

// UpdateTaskStatus updates the status of a task.
func (c *ClickUpClient) UpdateTaskStatus(taskID, status string) error {
	path := fmt.Sprintf(clickupPathTask, taskID)
	body := map[string]any{
		"status": status,
	}

	_, err := c.doRequest(http.MethodPut, path, body)
	return err
}

func (c *ClickUpClient) doRequest(method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("clickup: marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("clickup: create request: %w", err)
	}

	// ClickUp uses token without Bearer prefix
	req.Header.Set(headerAuthorization, c.apiToken)
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("clickup: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("clickup: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("clickup: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
