package controller

import (
	"context"
	"net/http"
	"strings"

	"jarvis/internal/hooks"
	"jarvis/internal/skills"
	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/web"
)

type ConversationController struct {
	usecase *usecase.ConversationUseCase
	ai      domain.AIProvider
	skills  skills.SkillProvider
	hooks   *hooks.Registry
}

func NewConversationController(uc *usecase.ConversationUseCase, ai domain.AIProvider, skillsProvider skills.SkillProvider, hooksRegistry *hooks.Registry) *ConversationController {
	return &ConversationController{
		usecase: uc,
		ai:      ai,
		skills:  skillsProvider,
		hooks:   hooksRegistry,
	}
}

func (c *ConversationController) PostChat(req web.Request) web.Response {
	var payload domain.ChatRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ChatResponse{Error: "invalid body"})
	}

	if err := payload.Validate(); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ChatResponse{Error: err.Error()})
	}

	if payload.SessionID == "" {
		payload.SessionID = payload.Sender
	}

	c.hooks.Emit(context.Background(), hooks.MessageReceived, map[string]string{
		"sender": payload.Sender, "message": payload.Message, "session_id": payload.SessionID,
	})

	if err := c.usecase.Ingest(payload.SessionID, domain.RoleUser, payload.Message); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ChatResponse{Error: "failed to save message"})
	}

	messages, err := c.usecase.Assemble(payload.SessionID)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ChatResponse{Error: "failed to assemble context"})
	}

	systemPrompt := c.buildSystemPrompt(payload.Sender, payload.Message)

	response, err := c.ai.CompleteMessages(systemPrompt, messages)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ChatResponse{Error: err.Error()})
	}

	c.hooks.Emit(context.Background(), hooks.BeforeResponse, map[string]string{
		"session_id": payload.SessionID, "response": response,
	})

	_ = c.usecase.Ingest(payload.SessionID, domain.RoleAssistant, response)

	return web.NewJSONResponse(http.StatusOK, domain.ChatResponse{Success: true, Response: response})
}

func (c *ConversationController) buildSystemPrompt(sender, message string) string {
	var sb strings.Builder
	sb.WriteString(domain.DefaultSystemPrompt)
	sb.WriteString("El usuario se llama " + sender + ".\n\n")

	loaded, err := c.skills.LoadEnabled()
	if err != nil || len(loaded) == 0 {
		return sb.String()
	}

	relevant := selectRelevantSkills(loaded, message)

	sb.WriteString(domain.SkillsSectionHeader)
	sb.WriteString(skills.FormatForPrompt(relevant))

	return sb.String()
}

func selectRelevantSkills(allSkills []skills.Skill, message string) []skills.Skill {
	tags := skills.ClassifyMessage(message)
	if len(tags) == 0 {
		return allSkills
	}
	return skills.FilterByTags(allSkills, tags...)
}
