package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

func TestSkillController_CreateSkill_InvalidJSON(t *testing.T) {
	ctrl := NewSkillController(nil)
	req := test.NewMockRequest().WithBody(`{invalid`)

	resp := ctrl.CreateSkill(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestSkillController_CreateSkill_MissingName(t *testing.T) {
	ctrl := NewSkillController(nil)
	req := test.NewMockRequest().WithBody(`{"name":"","content":"some content"}`)

	resp := ctrl.CreateSkill(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestSkillController_CreateSkill_MissingContent(t *testing.T) {
	ctrl := NewSkillController(nil)
	req := test.NewMockRequest().WithBody(`{"name":"test-skill","content":""}`)

	resp := ctrl.CreateSkill(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}
