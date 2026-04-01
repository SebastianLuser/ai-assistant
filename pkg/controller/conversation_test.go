package controller

import (
	"net/http"
	"testing"

	"jarvis/internal/hooks"
	"jarvis/internal/skills"
	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validChatBody   = `{"message":"hola","sender":"Sebas","session_id":"s1"}`
	emptyChatBody   = `{"message":"","sender":"Sebas"}`
	invalidChatBody = `{nope`
	noSessionBody   = `{"message":"hola","sender":"Sebas"}`
)

func TestConversationController_PostChat_InvalidJSON(t *testing.T) {
	ctrl := NewConversationController(nil, nil, nil, nil)
	req := test.NewMockRequest().WithBody(invalidChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestConversationController_PostChat_EmptyMessage(t *testing.T) {
	ctrl := NewConversationController(nil, nil, nil, nil)
	req := test.NewMockRequest().WithBody(emptyChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

var emptySkillProvider = &mockSkillProvider{}

func TestConversationController_PostChat_SessionIDFallsBackToSender(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SaveConversation", "Sebas", domain.RoleUser, "hola").Return(nil)
	store.On("LoadConversation", "Sebas", mock.Anything).Return([]domain.ConversationMessage{
		{Role: domain.RoleUser, Content: "hola"},
	}, nil)
	store.On("SaveConversation", "Sebas", domain.RoleAssistant, mock.Anything).Return(nil)

	ai := new(test.MockAIProvider)
	ai.On("CompleteMessages", mock.Anything, mock.Anything).Return("todo bien", nil)

	hooksReg := hooks.NewRegistry()
	uc := usecase.NewConversationUseCase(store, ai, hooksReg, 50, 40)
	ctrl := NewConversationController(uc, ai, emptySkillProvider, hooksReg)
	req := test.NewMockRequest().WithBody(noSessionBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertCalled(t, "SaveConversation", "Sebas", domain.RoleUser, "hola")
}

func TestConversationController_PostChat_Success(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SaveConversation", "s1", domain.RoleUser, "hola").Return(nil)
	store.On("LoadConversation", "s1", mock.Anything).Return([]domain.ConversationMessage{
		{Role: domain.RoleUser, Content: "hola"},
	}, nil)
	store.On("SaveConversation", "s1", domain.RoleAssistant, "respuesta").Return(nil)

	ai := new(test.MockAIProvider)
	ai.On("CompleteMessages", mock.Anything, mock.Anything).Return("respuesta", nil)

	hooksReg := hooks.NewRegistry()
	uc := usecase.NewConversationUseCase(store, ai, hooksReg, 50, 40)
	ctrl := NewConversationController(uc, ai, emptySkillProvider, hooksReg)
	req := test.NewMockRequest().WithBody(validChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
	ai.AssertExpectations(t)
}

func TestConversationController_PostChat_IngestError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SaveConversation", "s1", domain.RoleUser, "hola").Return(domain.ErrStoreOpen)

	hooksReg := hooks.NewRegistry()
	uc := usecase.NewConversationUseCase(store, nil, hooksReg, 50, 40)
	ctrl := NewConversationController(uc, nil, emptySkillProvider, hooksReg)
	req := test.NewMockRequest().WithBody(validChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestConversationController_PostChat_AssembleError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SaveConversation", "s1", domain.RoleUser, "hola").Return(nil)
	store.On("LoadConversation", "s1", mock.Anything).Return([]domain.ConversationMessage(nil), domain.ErrStoreOpen)

	hooksReg := hooks.NewRegistry()
	uc := usecase.NewConversationUseCase(store, nil, hooksReg, 50, 40)
	ctrl := NewConversationController(uc, nil, emptySkillProvider, hooksReg)
	req := test.NewMockRequest().WithBody(validChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestConversationController_PostChat_AIError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SaveConversation", "s1", domain.RoleUser, "hola").Return(nil)
	store.On("LoadConversation", "s1", mock.Anything).Return([]domain.ConversationMessage{
		{Role: domain.RoleUser, Content: "hola"},
	}, nil)

	ai := new(test.MockAIProvider)
	ai.On("CompleteMessages", mock.Anything, mock.Anything).Return("", domain.ErrClaudeAPI)

	hooksReg := hooks.NewRegistry()
	uc := usecase.NewConversationUseCase(store, ai, hooksReg, 50, 40)
	ctrl := NewConversationController(uc, ai, emptySkillProvider, hooksReg)
	req := test.NewMockRequest().WithBody(validChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

// mockSkillProvider implements skills.SkillProvider for testing.
type mockSkillProvider struct {
	skills []skills.Skill
	err    error
}

func (m *mockSkillProvider) LoadEnabled() ([]skills.Skill, error) {
	return m.skills, m.err
}

var _ skills.SkillProvider = (*mockSkillProvider)(nil)

func TestConversationController_PostChat_WithSkills(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SaveConversation", "s1", domain.RoleUser, "hola").Return(nil)
	store.On("LoadConversation", "s1", mock.Anything).Return([]domain.ConversationMessage{
		{Role: domain.RoleUser, Content: "hola"},
	}, nil)
	store.On("SaveConversation", "s1", domain.RoleAssistant, "respuesta con skills").Return(nil)

	ai := new(test.MockAIProvider)
	ai.On("CompleteMessages", mock.Anything, mock.Anything).Return("respuesta con skills", nil)

	enabled := true
	sp := &mockSkillProvider{
		skills: []skills.Skill{
			{Name: "greeting", Description: "Greets user", Content: "Say hi", Enabled: &enabled, Tags: []string{"general"}},
		},
	}

	hooksReg := hooks.NewRegistry()
	uc := usecase.NewConversationUseCase(store, ai, hooksReg, 50, 40)
	ctrl := NewConversationController(uc, ai, sp, hooksReg)
	req := test.NewMockRequest().WithBody(validChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestConversationController_PostChat_SkillProviderError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SaveConversation", "s1", domain.RoleUser, "hola").Return(nil)
	store.On("LoadConversation", "s1", mock.Anything).Return([]domain.ConversationMessage{
		{Role: domain.RoleUser, Content: "hola"},
	}, nil)
	store.On("SaveConversation", "s1", domain.RoleAssistant, mock.Anything).Return(nil)

	ai := new(test.MockAIProvider)
	ai.On("CompleteMessages", mock.Anything, mock.Anything).Return("ok", nil)

	sp := &mockSkillProvider{err: domain.ErrStoreOpen}

	hooksReg := hooks.NewRegistry()
	uc := usecase.NewConversationUseCase(store, ai, hooksReg, 50, 40)
	ctrl := NewConversationController(uc, ai, sp, hooksReg)
	req := test.NewMockRequest().WithBody(validChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
