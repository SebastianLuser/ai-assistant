package controller

import (
	"net/http"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

// JiraController handles Jira API endpoints.
type JiraController struct {
	client *clients.JiraClient
}

// NewJiraController creates a new JiraController.
func NewJiraController(client *clients.JiraClient) *JiraController {
	return &JiraController{client: client}
}

// GetMyIssues returns issues assigned to the current user.
func (c *JiraController) GetMyIssues(req web.Request) web.Response {
	issues, err := c.client.GetMyIssues()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.JiraIssueListResponse{Error: err.Error()})
	}

	domainIssues := make([]domain.JiraIssue, len(issues))
	for i, iss := range issues {
		domainIssues[i] = domain.JiraIssue{
			Key: iss.Key, Summary: iss.Summary, Status: iss.Status,
			Assignee: iss.Assignee, URL: iss.URL, Type: iss.Type,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.JiraIssueListResponse{
		Success: true, Issues: domainIssues,
	})
}

// GetIssue returns a single issue by key.
func (c *JiraController) GetIssue(req web.Request) web.Response {
	key, ok := req.Param(domain.PathParamKey)
	if !ok || key == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.JiraIssueResponse{Error: "key is required"})
	}

	issue, err := c.client.GetIssue(key)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.JiraIssueResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.JiraIssueResponse{
		Success: true,
		Issue: &domain.JiraIssue{
			Key: issue.Key, Summary: issue.Summary, Status: issue.Status,
			Assignee: issue.Assignee, URL: issue.URL, Type: issue.Type,
		},
	})
}

// CreateIssue creates a new Jira issue.
func (c *JiraController) CreateIssue(req web.Request) web.Response {
	var payload domain.JiraCreateIssueRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.JiraIssueResponse{Error: "invalid body"})
	}

	if payload.ProjectKey == "" || payload.Summary == "" || payload.IssueType == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.JiraIssueResponse{Error: "project_key, summary, and issue_type are required"})
	}

	issue, err := c.client.CreateIssue(payload.ProjectKey, payload.Summary, payload.Description, payload.IssueType)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.JiraIssueResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.JiraIssueResponse{
		Success: true,
		Issue: &domain.JiraIssue{
			Key: issue.Key, Summary: issue.Summary, Status: issue.Status,
			Assignee: issue.Assignee, URL: issue.URL, Type: issue.Type,
		},
	})
}
