package usecase

import (
	"context"
	"log"
	"strings"

	"asistente/internal/hooks"
	"asistente/internal/skills"
	"asistente/pkg/domain"
	"asistente/pkg/service"
)

// MessageRouter handles incoming messages from any channel (WhatsApp, Telegram, CLI, etc).
// It detects intent, delegates to the appropriate usecase, and sends a reply via the channel.
type MessageRouter struct {
	conversation *ConversationUseCase
	finance      *FinanceUseCase
	memorySvc    service.MemoryService
	embedder     service.Embedder
	ai           domain.AIProvider
	skills       skills.SkillProvider
	hooks        *hooks.Registry
	allowedFrom  string
}

func NewMessageRouter(
	conversation *ConversationUseCase,
	finance *FinanceUseCase,
	memorySvc service.MemoryService,
	embedder service.Embedder,
	ai domain.AIProvider,
	skillsProvider skills.SkillProvider,
	hooksRegistry *hooks.Registry,
	allowedFrom string,
) *MessageRouter {
	return &MessageRouter{
		conversation: conversation,
		finance:      finance,
		memorySvc:    memorySvc,
		embedder:     embedder,
		ai:           ai,
		skills:       skillsProvider,
		hooks:        hooksRegistry,
		allowedFrom:  allowedFrom,
	}
}

// ProcessMessage handles an incoming text message from any channel.
// It detects intent, delegates to the appropriate usecase, and sends a reply.
func (r *MessageRouter) ProcessMessage(ch domain.Channel, from, messageID, text string) {
	if r.allowedFrom != "" && from != r.allowedFrom {
		log.Printf("%s: ignoring message from unauthorized sender %s", ch.Name(), from)
		return
	}

	_ = ch.AckMessage(messageID)

	intent := detectIntent(text)
	log.Printf("%s: from=%s intent=%s msg=%q", ch.Name(), from, intent, truncate(text, 80))

	response, err := r.handleIntent(intent, from, text, ch.Name())
	if err != nil {
		log.Printf("%s: error handling intent %s: %v", ch.Name(), intent, err)
		response = "Perdón, hubo un error procesando tu mensaje. Intentá de nuevo."
	}

	if err := ch.SendMessage(from, response); err != nil {
		log.Printf("%s: failed to send reply to %s: %v", ch.Name(), from, err)
	}

	r.hooks.Emit(context.Background(), hooks.MessageProcessed, map[string]string{
		"channel": ch.Name(), "from": from, "intent": intent, "message": text,
	})
}

type intentType = string

const (
	intentExpense intentType = "expense"
	intentNote    intentType = "note"
	intentChat    intentType = "chat"
)

// detectIntent classifies the message using simple prefix/keyword matching.
// This avoids burning an AI call for every incoming message.
func detectIntent(text string) intentType {
	lower := strings.ToLower(strings.TrimSpace(text))

	expensePrefixes := []string{"gaste ", "gasté ", "gastamos ", "pague ", "pagué ", "pagamos ", "compre ", "compré ", "compramos "}
	for _, p := range expensePrefixes {
		if strings.HasPrefix(lower, p) {
			return intentExpense
		}
	}

	expenseKeywords := []string{" lucas", " luquitas", " pesos", " dolares", " dólares", " usd"}
	for _, k := range expenseKeywords {
		if strings.Contains(lower, k) {
			return intentExpense
		}
	}

	notePrefixes := []string{"nota ", "nota: ", "recordá ", "recorda ", "recordame ", "acordate "}
	for _, p := range notePrefixes {
		if strings.HasPrefix(lower, p) {
			return intentNote
		}
	}

	return intentChat
}

func (r *MessageRouter) handleIntent(intent, from, text, channelName string) (string, error) {
	switch intent {
	case intentExpense:
		return r.handleExpense(text)
	case intentNote:
		return r.handleNote(text, channelName)
	default:
		return r.handleChat(from, text, channelName)
	}
}

func (r *MessageRouter) handleExpense(text string) (string, error) {
	if r.finance == nil {
		return "El módulo de finanzas no está configurado.", nil
	}
	return r.finance.ProcessExpense(text, "Sebas")
}

func (r *MessageRouter) handleNote(text, channelName string) (string, error) {
	if r.memorySvc == nil {
		return "El módulo de notas no está configurado.", nil
	}

	content := stripNotePrefix(text)

	var embedding []float64
	if r.embedder != nil {
		emb, err := r.embedder.Embed(content)
		if err != nil {
			log.Printf("%s: embedding failed, saving without: %v", channelName, err)
		} else {
			embedding = emb
		}
	}

	_, err := r.memorySvc.Save(content, []string{channelName}, embedding)
	if err != nil {
		return "", domain.Wrapf(domain.ErrStoreSave, err)
	}

	return "Anotado!", nil
}

func (r *MessageRouter) handleChat(from, text, channelName string) (string, error) {
	sessionID := channelName + "-" + from

	if err := r.conversation.Ingest(sessionID, domain.RoleUser, text); err != nil {
		return "", err
	}

	messages, err := r.conversation.Assemble(sessionID)
	if err != nil {
		return "", err
	}

	systemPrompt := r.buildSystemPrompt(text, channelName)

	response, err := r.ai.CompleteMessages(systemPrompt, messages)
	if err != nil {
		return "", err
	}

	_ = r.conversation.Ingest(sessionID, domain.RoleAssistant, response)

	return response, nil
}

func (r *MessageRouter) buildSystemPrompt(message, channelName string) string {
	var sb strings.Builder
	sb.WriteString(domain.DefaultSystemPrompt)
	sb.WriteString("El usuario te habla por " + channelName + ". Sé conciso.\n\n")

	if r.skills == nil {
		return sb.String()
	}

	loaded, err := r.skills.LoadEnabled()
	if err != nil || len(loaded) == 0 {
		return sb.String()
	}

	tags := skills.ClassifyMessage(message)
	var relevant []skills.Skill
	if len(tags) == 0 {
		relevant = loaded
	} else {
		relevant = skills.FilterByTags(loaded, tags...)
	}

	sb.WriteString(domain.SkillsSectionHeader)
	sb.WriteString(skills.FormatForPrompt(relevant))

	return sb.String()
}

func stripNotePrefix(text string) string {
	lower := strings.ToLower(text)
	prefixes := []string{"nota: ", "nota ", "recordá ", "recorda ", "recordame ", "acordate "}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return strings.TrimSpace(text[len(p):])
		}
	}
	return text
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
