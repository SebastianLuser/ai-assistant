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

func newMockGitHubServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/user/repos":
			json.NewEncoder(w).Encode([]clients.GitHubRepo{
				{ID: 1, Name: "repo1", FullName: "user/repo1", URL: "https://github.com/user/repo1"},
			})
		case r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(clients.GitHubIssue{ID: 10, Number: 1, Title: "new issue", State: "open"})
		case r.URL.Path == "/repos/user/repo1/pulls":
			json.NewEncoder(w).Encode([]clients.GitHubPullRequest{
				{ID: 1, Number: 1, Title: "PR 1", State: "open"},
			})
		default:
			json.NewEncoder(w).Encode([]clients.GitHubIssue{
				{ID: 1, Number: 1, Title: "Issue 1", State: "open"},
			})
		}
	}))
}

func TestGitHubController_ListRepos_Success(t *testing.T) {
	srv := newMockGitHubServer()
	defer srv.Close()
	client := clients.NewGitHubClientWithBaseURL("token", srv.URL)
	ctrl := NewGitHubController(client)
	req := test.NewMockRequest()

	resp := ctrl.ListRepos(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestGitHubController_ListIssues_MissingOwner(t *testing.T) {
	ctrl := NewGitHubController(nil)
	req := test.NewMockRequest()

	resp := ctrl.ListIssues(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGitHubController_ListIssues_MissingRepo(t *testing.T) {
	ctrl := NewGitHubController(nil)
	req := test.NewMockRequest().WithParam("owner", "user")

	resp := ctrl.ListIssues(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGitHubController_ListIssues_Success(t *testing.T) {
	srv := newMockGitHubServer()
	defer srv.Close()
	client := clients.NewGitHubClientWithBaseURL("token", srv.URL)
	ctrl := NewGitHubController(client)
	req := test.NewMockRequest().WithParam("owner", "user").WithParam("repo", "repo1")

	resp := ctrl.ListIssues(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestGitHubController_CreateIssue_MissingOwner(t *testing.T) {
	ctrl := NewGitHubController(nil)
	req := test.NewMockRequest()

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGitHubController_CreateIssue_MissingRepo(t *testing.T) {
	ctrl := NewGitHubController(nil)
	req := test.NewMockRequest().WithParam("owner", "user")

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGitHubController_CreateIssue_InvalidJSON(t *testing.T) {
	ctrl := NewGitHubController(nil)
	req := test.NewMockRequest().WithParam("owner", "user").WithParam("repo", "myrepo").WithBody(`{invalid`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGitHubController_CreateIssue_EmptyTitle(t *testing.T) {
	ctrl := NewGitHubController(nil)
	req := test.NewMockRequest().WithParam("owner", "user").WithParam("repo", "myrepo").WithBody(`{"title":"","body":"some body"}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGitHubController_CreateIssue_Success(t *testing.T) {
	srv := newMockGitHubServer()
	defer srv.Close()
	client := clients.NewGitHubClientWithBaseURL("token", srv.URL)
	ctrl := NewGitHubController(client)
	req := test.NewMockRequest().WithParam("owner", "user").WithParam("repo", "repo1").WithBody(`{"title":"Bug fix","body":"details"}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}

func TestGitHubController_ListPRs_MissingOwner(t *testing.T) {
	ctrl := NewGitHubController(nil)
	req := test.NewMockRequest()

	resp := ctrl.ListPRs(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGitHubController_ListPRs_MissingRepo(t *testing.T) {
	ctrl := NewGitHubController(nil)
	req := test.NewMockRequest().WithParam("owner", "user")

	resp := ctrl.ListPRs(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGitHubController_ListPRs_Success(t *testing.T) {
	srv := newMockGitHubServer()
	defer srv.Close()
	client := clients.NewGitHubClientWithBaseURL("token", srv.URL)
	ctrl := NewGitHubController(client)
	req := test.NewMockRequest().WithParam("owner", "user").WithParam("repo", "repo1")

	resp := ctrl.ListPRs(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
