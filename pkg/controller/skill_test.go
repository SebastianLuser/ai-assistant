package controller

import (
	"net/http"
	"testing"

	"asistente/internal/skills"
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

// mockSkillWriter implements skills.SkillWriter for testing.
type mockSkillWriter struct {
	loaded []skills.Skill
	err    error
	saveErr error
}

func (m *mockSkillWriter) LoadEnabled() ([]skills.Skill, error) {
	return m.loaded, m.err
}

func (m *mockSkillWriter) Save(_ skills.Skill) error {
	return m.saveErr
}

var _ skills.SkillWriter = (*mockSkillWriter)(nil)

func TestSkillController_CreateSkill_Success(t *testing.T) {
	writer := &mockSkillWriter{}
	ctrl := NewSkillController(writer)
	req := test.NewMockRequest().WithBody(`{"name":"test-skill","content":"some content","description":"A test skill","tags":["test"]}`)

	resp := ctrl.CreateSkill(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}

func TestSkillController_CreateSkill_SaveError(t *testing.T) {
	writer := &mockSkillWriter{saveErr: assert.AnError}
	ctrl := NewSkillController(writer)
	req := test.NewMockRequest().WithBody(`{"name":"test-skill","content":"some content"}`)

	resp := ctrl.CreateSkill(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestSkillController_ListSkills_Success(t *testing.T) {
	enabled := true
	writer := &mockSkillWriter{
		loaded: []skills.Skill{
			{Name: "skill1", Description: "desc", Tags: []string{"tag1"}, Enabled: &enabled},
		},
	}
	ctrl := NewSkillController(writer)
	req := test.NewMockRequest()

	resp := ctrl.ListSkills(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestSkillController_ListSkills_Empty(t *testing.T) {
	writer := &mockSkillWriter{loaded: []skills.Skill{}}
	ctrl := NewSkillController(writer)
	req := test.NewMockRequest()

	resp := ctrl.ListSkills(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestSkillController_ListSkills_Error(t *testing.T) {
	writer := &mockSkillWriter{err: assert.AnError}
	ctrl := NewSkillController(writer)
	req := test.NewMockRequest()

	resp := ctrl.ListSkills(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}
