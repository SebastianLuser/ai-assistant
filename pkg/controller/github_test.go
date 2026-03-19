package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

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
