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
	todoistBaseURL        = "https://api.todoist.com"
	todoistDefaultTimeout = 15 * time.Second
	todoistPathTasks      = "/rest/v2/tasks"
)

// TodoistTask represents a Todoist task.
type TodoistTask struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	DueDate   string `json:"due_date,omitempty"`
	Priority  int    `json:"priority"`
	Completed bool   `json:"is_completed"`
	URL       string `json:"url"`
}

// TodoistClient is the Todoist API client.
type TodoistClient struct {
	apiToken   string
	baseURL    string
	httpClient *http.Client
}

// NewTodoistClient creates a new Todoist API client.
func NewTodoistClient(apiToken string) *TodoistClient {
	return &TodoistClient{
		apiToken:   apiToken,
		baseURL:    todoistBaseURL,
		httpClient: &http.Client{Timeout: todoistDefaultTimeout},
	}
}

// NewTodoistClientWithBaseURL creates a Todoist client pointing at a custom base URL (for testing).
func NewTodoistClientWithBaseURL(apiToken, baseURL string) *TodoistClient {
	c := NewTodoistClient(apiToken)
	c.baseURL = baseURL
	return c
}

// GetTasks returns all active tasks.
func (c *TodoistClient) GetTasks() ([]TodoistTask, error) {
	resp, err := c.doRequest(http.MethodGet, todoistPathTasks, nil)
	if err != nil {
		return nil, err
	}

	var rawTasks []struct {
		ID        string `json:"id"`
		Content   string `json:"content"`
		Priority  int    `json:"priority"`
		Completed bool   `json:"is_completed"`
		URL       string `json:"url"`
		Due       *struct {
			Date string `json:"date"`
		} `json:"due"`
	}

	if err := json.Unmarshal(resp, &rawTasks); err != nil {
		return nil, fmt.Errorf("todoist: parse tasks: %w", err)
	}

	tasks := make([]TodoistTask, len(rawTasks))
	for i, raw := range rawTasks {
		tasks[i] = TodoistTask{
			ID:        raw.ID,
			Content:   raw.Content,
			Priority:  raw.Priority,
			Completed: raw.Completed,
			URL:       raw.URL,
		}
		if raw.Due != nil {
			tasks[i].DueDate = raw.Due.Date
		}
	}

	return tasks, nil
}

// CreateTask creates a new task.
func (c *TodoistClient) CreateTask(content string, dueDate *string) (TodoistTask, error) {
	body := map[string]any{
		"content": content,
	}
	if dueDate != nil {
		body["due_date"] = *dueDate
	}

	resp, err := c.doRequest(http.MethodPost, todoistPathTasks, body)
	if err != nil {
		return TodoistTask{}, err
	}

	var raw struct {
		ID        string `json:"id"`
		Content   string `json:"content"`
		Priority  int    `json:"priority"`
		Completed bool   `json:"is_completed"`
		URL       string `json:"url"`
		Due       *struct {
			Date string `json:"date"`
		} `json:"due"`
	}

	if err := json.Unmarshal(resp, &raw); err != nil {
		return TodoistTask{}, fmt.Errorf("todoist: parse create response: %w", err)
	}

	task := TodoistTask{
		ID:        raw.ID,
		Content:   raw.Content,
		Priority:  raw.Priority,
		Completed: raw.Completed,
		URL:       raw.URL,
	}
	if raw.Due != nil {
		task.DueDate = raw.Due.Date
	}

	return task, nil
}

// CompleteTask marks a task as completed.
func (c *TodoistClient) CompleteTask(taskID string) error {
	_, err := c.doRequest(http.MethodPost, todoistPathTasks+"/"+taskID+"/close", nil)
	return err
}

func (c *TodoistClient) doRequest(method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("todoist: marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("todoist: create request: %w", err)
	}

	req.Header.Set(headerAuthorization, "Bearer "+c.apiToken)
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("todoist: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("todoist: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("todoist: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
