package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

func TestJiraController_GetIssue_MissingKey(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_CreateIssue_InvalidJSON(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest().WithBody(`{invalid`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_CreateIssue_MissingProjectKey(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest().WithBody(`{"project_key":"","summary":"bug","issue_type":"Bug"}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_CreateIssue_MissingSummary(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest().WithBody(`{"project_key":"PROJ","summary":"","issue_type":"Bug"}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestJiraController_CreateIssue_MissingIssueType(t *testing.T) {
	ctrl := NewJiraController(nil)
	req := test.NewMockRequest().WithBody(`{"project_key":"PROJ","summary":"bug","issue_type":""}`)

	resp := ctrl.CreateIssue(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}
