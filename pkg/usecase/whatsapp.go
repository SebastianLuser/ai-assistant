package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"
	"sync"

	"asistente/internal/hooks"
	"asistente/internal/skills"
	"asistente/pkg/domain"
)

// MessageRouter handles incoming messages from any channel (WhatsApp, Telegram, CLI, etc).
// Uses Claude's tool use to autonomously decide which actions to execute.
type MessageRouter struct {
	conversation *ConversationUseCase
	ai           domain.AIProvider
	agent        *AgentUseCase
	transcriber  domain.Transcriber
	skills       skills.SkillProvider
	hooks        *hooks.Registry
	allowedFrom  string

	pairingMu     sync.RWMutex
	pairingCode   string
	pairedSenders map[string]bool
}

func NewMessageRouter(
	conversation *ConversationUseCase,
	ai domain.AIProvider,
	agent *AgentUseCase,
	transcriber domain.Transcriber,
	skillsProvider skills.SkillProvider,
	hooksRegistry *hooks.Registry,
	allowedFrom string,
) *MessageRouter {
	code := generatePairingCode()
	if allowedFrom == "" {
		log.Printf("message-router: pairing code for new senders: %s", code)
	}

	return &MessageRouter{
		conversation:  conversation,
		ai:            ai,
		agent:         agent,
		transcriber:   transcriber,
		skills:        skillsProvider,
		hooks:         hooksRegistry,
		allowedFrom:   allowedFrom,
		pairingCode:   code,
		pairedSenders: make(map[string]bool),
	}
}

// GetPairingCode returns the current pairing code for authorizing new senders.
func (r *MessageRouter) GetPairingCode() string {
	r.pairingMu.RLock()
	defer r.pairingMu.RUnlock()
	return r.pairingCode
}

func generatePairingCode() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "0000"
	}
	return hex.EncodeToString(b)
}

// ProcessMessage handles an incoming text message from any channel.
func (r *MessageRouter) ProcessMessage(ch domain.Channel, from, messageID, text string) {
	if !r.isAuthorized(from) {
		r.handleUnauthorized(ch, from, messageID, text)
		return
	}

	_ = ch.AckMessage(messageID)
	log.Printf("%s: from=%s msg=%q", ch.Name(), from, truncate(text, 80))

	response, err := r.handleMessage(from, text, ch.Name())
	if err != nil {
		log.Printf("%s: error: %v", ch.Name(), err)
		response = "Perdón, hubo un error procesando tu mensaje. Intentá de nuevo."
	}

	if err := ch.SendMessage(from, response); err != nil {
		log.Printf("%s: failed to send reply to %s: %v", ch.Name(), from, err)
	}

	r.hooks.Emit(context.Background(), hooks.MessageProcessed, map[string]string{
		"channel": ch.Name(), "from": from, "message": text,
	})
}

// ProcessAudioMessage downloads audio, transcribes it, then processes as text.
func (r *MessageRouter) ProcessAudioMessage(ch domain.Channel, from, messageID, mediaID string) {
	if !r.isAuthorized(from) {
		r.handleUnauthorized(ch, from, messageID, mediaID)
		return
	}

	_ = ch.AckMessage(messageID)

	// Download audio
	downloader, ok := ch.(domain.MediaDownloader)
	if !ok {
		log.Printf("%s: channel does not support media download", ch.Name())
		_ = ch.SendMessage(from, "No puedo procesar audios desde este canal.")
		return
	}

	if r.transcriber == nil {
		_ = ch.SendMessage(from, "La transcripción de audio no está configurada. Necesito OPENAI_API_KEY.")
		return
	}

	audioData, mimeType, err := downloader.DownloadMedia(mediaID)
	if err != nil {
		log.Printf("%s: failed to download audio: %v", ch.Name(), err)
		_ = ch.SendMessage(from, "No pude descargar el audio.")
		return
	}

	log.Printf("%s: transcribing audio from %s (%d bytes, %s)", ch.Name(), from, len(audioData), mimeType)

	text, err := r.transcriber.Transcribe(audioData, mimeType)
	if err != nil {
		log.Printf("%s: transcription failed: %v", ch.Name(), err)
		_ = ch.SendMessage(from, "No pude transcribir el audio.")
		return
	}

	log.Printf("%s: transcribed: %q", ch.Name(), truncate(text, 100))

	// Process the transcribed text as a normal message
	r.ProcessMessage(ch, from, messageID, text)
}

func (r *MessageRouter) handleMessage(from, text, channelName string) (string, error) {
	sessionID := channelName + "-" + from

	if err := r.conversation.Ingest(sessionID, domain.RoleUser, text); err != nil {
		return "", err
	}

	messages, err := r.conversation.Assemble(sessionID)
	if err != nil {
		return "", err
	}

	systemPrompt := r.buildSystemPrompt(text, channelName)

	// Use agent with tools if available, otherwise fall back to plain chat.
	var response string
	if r.agent != nil {
		response, err = r.agent.Run(systemPrompt, messages)
	} else {
		response, err = r.ai.CompleteMessages(systemPrompt, messages)
	}
	if err != nil {
		return "", err
	}

	_ = r.conversation.Ingest(sessionID, domain.RoleAssistant, response)

	return response, nil
}

func (r *MessageRouter) buildSystemPrompt(message, channelName string) string {
	var sb strings.Builder
	sb.WriteString(domain.DefaultSystemPrompt)
	sb.WriteString("El usuario te habla por " + channelName + ". Sé conciso.\n")
	sb.WriteString("Usá las herramientas disponibles cuando sea apropiado para ejecutar acciones.\n\n")

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

func (r *MessageRouter) isAuthorized(from string) bool {
	if r.allowedFrom != "" {
		return from == r.allowedFrom
	}
	r.pairingMu.RLock()
	defer r.pairingMu.RUnlock()
	return r.pairedSenders[from]
}

func (r *MessageRouter) handleUnauthorized(ch domain.Channel, from, messageID, text string) {
	trimmed := strings.TrimSpace(text)
	r.pairingMu.Lock()
	defer r.pairingMu.Unlock()

	if trimmed == r.pairingCode {
		r.pairedSenders[from] = true
		r.pairingCode = generatePairingCode()
		log.Printf("%s: sender %s paired successfully, new code: %s", ch.Name(), from, r.pairingCode)
		_ = ch.SendMessage(from, "Paired! Ya estás autorizado para hablarme.")
		return
	}

	log.Printf("%s: unauthorized sender %s (send pairing code to connect)", ch.Name(), from)
	_ = ch.SendMessage(from, "No te conozco. Enviame el código de vinculación para conectarte.")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
