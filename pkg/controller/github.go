package controller

import (
	"net/http"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

// GitHubController handles GitHub API endpoints.
type GitHubController struct {
	client *clients.GitHubClient
}

// NewGitHubController creates a new GitHubController.
func NewGitHubController(client *clients.GitHubClient) *GitHubController {
	return &GitHubController{client: client}
}

// ListRepos returns the authenticated user's repositories.
func (c *GitHubController) ListRepos(req web.Request) web.Response {
	repos, err := c.client.ListRepos()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.GitHubRepoListResponse{Error: err.Error()})
	}

	domainRepos := make([]domain.GitHubRepo, len(repos))
	for i, r := range repos {
		domainRepos[i] = domain.GitHubRepo{
			ID: r.ID, Name: r.Name, FullName: r.FullName, URL: r.URL,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.GitHubRepoListResponse{
		Success: true, Repos: domainRepos,
	})
}

// ListIssues returns issues for the given repository.
func (c *GitHubController) ListIssues(req web.Request) web.Response {
	owner, ok := req.Param(domain.PathParamOwner)
	if !ok || owner == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GitHubIssueListResponse{Error: "owner is required"})
	}

	repo, ok := req.Param(domain.PathParamRepo)
	if !ok || repo == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GitHubIssueListResponse{Error: "repo is required"})
	}

	issues, err := c.client.ListIssues(owner, repo)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.GitHubIssueListResponse{Error: err.Error()})
	}

	domainIssues := make([]domain.GitHubIssue, len(issues))
	for i, iss := range issues {
		domainIssues[i] = domain.GitHubIssue{
			ID: iss.ID, Number: iss.Number, Title: iss.Title, State: iss.State, URL: iss.URL,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.GitHubIssueListResponse{
		Success: true, Issues: domainIssues,
	})
}

// CreateIssue creates a new issue in the given repository.
func (c *GitHubController) CreateIssue(req web.Request) web.Response {
	owner, ok := req.Param(domain.PathParamOwner)
	if !ok || owner == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GitHubIssueResponse{Error: "owner is required"})
	}

	repo, ok := req.Param(domain.PathParamRepo)
	if !ok || repo == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GitHubIssueResponse{Error: "repo is required"})
	}

	var payload domain.GitHubCreateIssueRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GitHubIssueResponse{Error: "invalid body"})
	}

	if payload.Title == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GitHubIssueResponse{Error: "title is required"})
	}

	issue, err := c.client.CreateIssue(owner, repo, payload.Title, payload.Body)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.GitHubIssueResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.GitHubIssueResponse{
		Success: true,
		Issue: &domain.GitHubIssue{
			ID: issue.ID, Number: issue.Number, Title: issue.Title, State: issue.State, URL: issue.URL,
		},
	})
}

// ListPRs returns pull requests for the given repository.
func (c *GitHubController) ListPRs(req web.Request) web.Response {
	owner, ok := req.Param(domain.PathParamOwner)
	if !ok || owner == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GitHubPRListResponse{Error: "owner is required"})
	}

	repo, ok := req.Param(domain.PathParamRepo)
	if !ok || repo == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GitHubPRListResponse{Error: "repo is required"})
	}

	prs, err := c.client.ListPRs(owner, repo)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.GitHubPRListResponse{Error: err.Error()})
	}

	domainPRs := make([]domain.GitHubPullRequest, len(prs))
	for i, pr := range prs {
		domainPRs[i] = domain.GitHubPullRequest{
			ID: pr.ID, Number: pr.Number, Title: pr.Title, State: pr.State, URL: pr.URL, Draft: pr.Draft,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.GitHubPRListResponse{
		Success: true, PullRequests: domainPRs,
	})
}
