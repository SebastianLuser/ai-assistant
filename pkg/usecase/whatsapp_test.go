package usecase

import (
	"testing"

	"jarvis/internal/hooks"
	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hel...", truncate("hello world", 3))
}

func TestTruncate_ExactLength(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 5))
}

func TestGeneratePairingCode(t *testing.T) {
	code := generatePairingCode()

	assert.Len(t, code, 8)
	assert.NotEqual(t, "00000000", code)
}

func TestGeneratePairingCode_Unique(t *testing.T) {
	code1 := generatePairingCode()
	code2 := generatePairingCode()

	// Probabilistically they should differ
	assert.Len(t, code1, 8)
	assert.Len(t, code2, 8)
}

// --- isAuthorized tests ---

func TestMessageRouter_isAuthorized_AllowedFrom(t *testing.T) {
	router := &MessageRouter{
		allowedFrom:   "1234567890",
		pairedSenders: make(map[string]bool),
	}

	assert.True(t, router.isAuthorized("1234567890"))
	assert.False(t, router.isAuthorized("9999999999"))
}

func TestMessageRouter_isAuthorized_PairedSender(t *testing.T) {
	router := &MessageRouter{
		allowedFrom:   "",
		pairedSenders: map[string]bool{"1234567890": true},
	}

	assert.True(t, router.isAuthorized("1234567890"))
	assert.False(t, router.isAuthorized("9999999999"))
}

func TestMessageRouter_isAuthorized_NobodyAllowed(t *testing.T) {
	router := &MessageRouter{
		allowedFrom:   "",
		pairedSenders: make(map[string]bool),
	}

	assert.False(t, router.isAuthorized("1234567890"))
}

// --- handleUnauthorized tests ---

func TestMessageRouter_handleUnauthorized_WrongCode(t *testing.T) {
	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("SendMessage", "123", mock.Anything).Return(nil)

	router := &MessageRouter{
		pairingCode:   "abcd1234",
		pairedSenders: make(map[string]bool),
	}

	router.handleUnauthorized(ch, "123", "msg-id", "wrong-code")

	assert.False(t, router.pairedSenders["123"])
	ch.AssertCalled(t, "SendMessage", "123", mock.Anything)
}

func TestMessageRouter_handleUnauthorized_CorrectCode(t *testing.T) {
	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("SendMessage", "123", mock.Anything).Return(nil)

	router := &MessageRouter{
		pairingCode:   "abcd1234",
		pairedSenders: make(map[string]bool),
	}

	router.handleUnauthorized(ch, "123", "msg-id", "abcd1234")

	assert.True(t, router.pairedSenders["123"])
	// Code should be regenerated
	assert.NotEqual(t, "abcd1234", router.pairingCode)
	assert.Len(t, router.pairingCode, 8)
}

func TestMessageRouter_handleUnauthorized_CorrectCodeWithWhitespace(t *testing.T) {
	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("SendMessage", "123", mock.Anything).Return(nil)

	router := &MessageRouter{
		pairingCode:   "abcd1234",
		pairedSenders: make(map[string]bool),
	}

	router.handleUnauthorized(ch, "123", "msg-id", "  abcd1234  ")

	assert.True(t, router.pairedSenders["123"])
}

// --- GetPairingCode tests ---

func TestMessageRouter_GetPairingCode(t *testing.T) {
	router := &MessageRouter{
		pairingCode:   "test1234",
		pairedSenders: make(map[string]bool),
	}

	code := router.GetPairingCode()

	assert.Equal(t, "test1234", code)
}

// --- buildSystemPrompt tests ---

func TestMessageRouter_buildSystemPrompt_NoSkills(t *testing.T) {
	router := &MessageRouter{
		skills: nil,
	}

	prompt := router.buildSystemPrompt("hello", "whatsapp")

	assert.Contains(t, prompt, domain.DefaultSystemPrompt)
	assert.Contains(t, prompt, "whatsapp")
	assert.Contains(t, prompt, "Fecha:")
	assert.Contains(t, prompt, "Hora:")
}

func TestMessageRouter_buildSystemPrompt_ContainsDayPeriod(t *testing.T) {
	router := &MessageRouter{
		skills: nil,
	}

	prompt := router.buildSystemPrompt("hello", "telegram")

	// Should contain one of the day periods
	hasPeriod := false
	for _, p := range []string{"madrugada", "mañana", "tarde", "noche"} {
		if assert.ObjectsAreEqual(true, len(prompt) > 0) {
			if contains(prompt, p) {
				hasPeriod = true
				break
			}
		}
	}
	assert.True(t, hasPeriod, "prompt should contain a day period")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- ProcessMessage tests ---

func newTestRouter(t *testing.T, allowedFrom string) (*MessageRouter, func()) {
	t.Helper()
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("LoadConversation", mock.Anything, mock.Anything).Return([]domain.ConversationMessage{}, nil)

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "respuesta"})

	hooksRegistry := hooks.NewRegistry()
	convUC := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	router := NewMessageRouter(convUC, ai, nil, nil, nil, nil, hooksRegistry, nil, allowedFrom)

	return router, srv.Close
}

func TestMessageRouter_ProcessMessage_Authorized(t *testing.T) {
	router, cleanup := newTestRouter(t, "1234567890")
	defer cleanup()

	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("AckMessage", "msg-1").Return(nil)
	ch.On("SendMessage", "1234567890", mock.Anything).Return(nil)

	router.ProcessMessage(ch, "1234567890", "msg-1", "hola")

	ch.AssertCalled(t, "AckMessage", "msg-1")
	ch.AssertCalled(t, "SendMessage", "1234567890", mock.Anything)
}

func TestMessageRouter_ProcessMessage_Unauthorized(t *testing.T) {
	router, cleanup := newTestRouter(t, "1234567890")
	defer cleanup()

	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("SendMessage", "9999999999", mock.Anything).Return(nil)

	router.ProcessMessage(ch, "9999999999", "msg-1", "hola")

	ch.AssertCalled(t, "SendMessage", "9999999999", mock.Anything)
	// Should NOT have called AckMessage since unauthorized
	ch.AssertNotCalled(t, "AckMessage", mock.Anything)
}

func TestMessageRouter_ProcessMessage_GroupChat_NotMentioned(t *testing.T) {
	router, cleanup := newTestRouter(t, "1234567890")
	defer cleanup()

	ch := &test.MockChannel{ChannelName: "test"}

	meta := domain.MessageMeta{IsGroup: true, BotName: "jarvis"}
	router.ProcessMessage(ch, "1234567890", "msg-1", "hola amigos", meta)

	// Should not have been processed (no mention)
	ch.AssertNotCalled(t, "AckMessage", mock.Anything)
	ch.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything)
}

func TestMessageRouter_ProcessMessage_GroupChat_Mentioned(t *testing.T) {
	router, cleanup := newTestRouter(t, "1234567890")
	defer cleanup()

	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("AckMessage", "msg-1").Return(nil)
	ch.On("SendMessage", "1234567890", mock.Anything).Return(nil)

	meta := domain.MessageMeta{IsGroup: true, BotName: "jarvis"}
	router.ProcessMessage(ch, "1234567890", "msg-1", "hola @asistente", meta)

	ch.AssertCalled(t, "AckMessage", "msg-1")
	ch.AssertCalled(t, "SendMessage", "1234567890", mock.Anything)
}

// --- handleMessage tests ---

func TestMessageRouter_handleMessage_WithAgent(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("LoadConversation", mock.Anything, mock.Anything).Return([]domain.ConversationMessage{}, nil)

	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "agent response"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "ai response"})
	defer srv.Close()

	hooksRegistry := hooks.NewRegistry()
	convUC := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)
	agentUC := NewAgentUseCase(provider, NewToolRegistry())

	router := NewMessageRouter(convUC, ai, agentUC, nil, nil, nil, hooksRegistry, nil, "123")

	response, err := router.handleMessage("123", "hola", "test")

	require.NoError(t, err)
	assert.Equal(t, "agent response", response)
}

func TestMessageRouter_handleMessage_WithUsageTracker(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("LoadConversation", mock.Anything, mock.Anything).Return([]domain.ConversationMessage{}, nil)

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "response"})
	defer srv.Close()

	hooksRegistry := hooks.NewRegistry()
	convUC := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)
	tracker := NewUsageTracker()

	router := NewMessageRouter(convUC, ai, nil, nil, nil, nil, hooksRegistry, tracker, "123")

	_, err := router.handleMessage("123", "hola", "test")

	require.NoError(t, err)
	session := tracker.GetSession("test-123")
	assert.NotNil(t, session)
	assert.Equal(t, 1, session.Requests)
}

func TestMessageRouter_handleMessage_IngestError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", mock.Anything, mock.Anything, mock.Anything).Return(errStoreFailure)

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "response"})
	defer srv.Close()

	hooksRegistry := hooks.NewRegistry()
	convUC := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	router := NewMessageRouter(convUC, ai, nil, nil, nil, nil, hooksRegistry, nil, "123")

	_, err := router.handleMessage("123", "hola", "test")

	require.Error(t, err)
}

// --- NewMessageRouter tests ---

func TestNewMessageRouter_SetsFields(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, nil, nil, hooksRegistry, nil, "123")

	assert.Equal(t, "123", router.allowedFrom)
	assert.NotEmpty(t, router.pairingCode)
	assert.NotNil(t, router.pairedSenders)
}

func TestNewMessageRouter_EmptyAllowedFrom_GeneratesPairingCode(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, nil, nil, hooksRegistry, nil, "")

	assert.NotEmpty(t, router.pairingCode)
	assert.Len(t, router.pairingCode, 8)
}

// --- ProcessAudioMessage tests ---

func TestMessageRouter_ProcessAudioMessage_Unauthorized(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, nil, nil, hooksRegistry, nil, "allowed")

	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("SendMessage", "unauthorized", mock.Anything).Return(nil)

	router.ProcessAudioMessage(ch, "unauthorized", "msg-1", "media-1")

	ch.AssertCalled(t, "SendMessage", "unauthorized", mock.Anything)
	ch.AssertNotCalled(t, "AckMessage", mock.Anything)
}

func TestMessageRouter_ProcessAudioMessage_NoMediaDownloader(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, nil, nil, hooksRegistry, nil, "123")

	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("AckMessage", "msg-1").Return(nil)
	ch.On("SendMessage", "123", mock.Anything).Return(nil)

	router.ProcessAudioMessage(ch, "123", "msg-1", "media-1")

	ch.AssertCalled(t, "AckMessage", "msg-1")
	ch.AssertCalled(t, "SendMessage", "123", "No puedo procesar audios desde este canal.")
}

func TestMessageRouter_ProcessAudioMessage_NoTranscriber(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, nil, nil, hooksRegistry, nil, "123")
	router.transcriber = nil

	mch := &test.MockMediaChannel{}
	mch.ChannelName = "test"
	mch.On("AckMessage", "msg-1").Return(nil)
	mch.On("SendMessage", "123", mock.Anything).Return(nil)

	router.ProcessAudioMessage(mch, "123", "msg-1", "media-1")

	mch.AssertCalled(t, "SendMessage", "123", "La transcripción de audio no está configurada. Necesito OPENAI_API_KEY.")
}

func TestMessageRouter_ProcessAudioMessage_DownloadError(t *testing.T) {
	transcriber := new(test.MockTranscriber)
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, transcriber, nil, hooksRegistry, nil, "123")

	mch := &test.MockMediaChannel{}
	mch.ChannelName = "test"
	mch.On("AckMessage", "msg-1").Return(nil)
	mch.On("DownloadMedia", "media-1").Return([]byte(nil), "", assert.AnError)
	mch.On("SendMessage", "123", mock.Anything).Return(nil)

	router.ProcessAudioMessage(mch, "123", "msg-1", "media-1")

	mch.AssertCalled(t, "SendMessage", "123", "No pude descargar el audio.")
}

func TestMessageRouter_ProcessAudioMessage_TranscriptionError(t *testing.T) {
	transcriber := new(test.MockTranscriber)
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, transcriber, nil, hooksRegistry, nil, "123")

	mch := &test.MockMediaChannel{}
	mch.ChannelName = "test"
	mch.On("AckMessage", "msg-1").Return(nil)
	mch.On("DownloadMedia", "media-1").Return([]byte("audio"), "audio/ogg", nil)
	mch.On("SendMessage", "123", mock.Anything).Return(nil)
	transcriber.On("Transcribe", []byte("audio"), "audio/ogg").Return("", assert.AnError)

	router.ProcessAudioMessage(mch, "123", "msg-1", "media-1")

	mch.AssertCalled(t, "SendMessage", "123", "No pude transcribir el audio.")
}

// --- ProcessImageMessage tests ---

func TestMessageRouter_ProcessImageMessage_Unauthorized(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, nil, nil, hooksRegistry, nil, "allowed")

	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("SendMessage", "unauthorized", mock.Anything).Return(nil)

	router.ProcessImageMessage(ch, "unauthorized", "msg-1", "media-1", "caption")

	ch.AssertCalled(t, "SendMessage", "unauthorized", mock.Anything)
}

func TestMessageRouter_ProcessImageMessage_NoMediaDownloader(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, nil, nil, hooksRegistry, nil, "123")

	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("AckMessage", "msg-1").Return(nil)
	ch.On("SendMessage", "123", mock.Anything).Return(nil)

	router.ProcessImageMessage(ch, "123", "msg-1", "media-1", "")

	ch.AssertCalled(t, "SendMessage", "123", "No puedo procesar imágenes desde este canal.")
}

func TestMessageRouter_ProcessImageMessage_DownloadError(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	router := NewMessageRouter(nil, nil, nil, nil, nil, nil, hooksRegistry, nil, "123")

	mch := &test.MockMediaChannel{}
	mch.ChannelName = "test"
	mch.On("AckMessage", "msg-1").Return(nil)
	mch.On("DownloadMedia", "media-1").Return([]byte(nil), "", assert.AnError)
	mch.On("SendMessage", "123", mock.Anything).Return(nil)

	router.ProcessImageMessage(mch, "123", "msg-1", "media-1", "")

	mch.AssertCalled(t, "SendMessage", "123", "No pude descargar la imagen.")
}

func TestMessageRouter_ProcessImageMessage_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("LoadConversation", mock.Anything, mock.Anything).Return([]domain.ConversationMessage{}, nil)

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "veo una imagen"})
	defer srv.Close()

	hooksRegistry := hooks.NewRegistry()
	convUC := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)
	router := NewMessageRouter(convUC, ai, nil, nil, nil, nil, hooksRegistry, nil, "123")

	mch := &test.MockMediaChannel{}
	mch.ChannelName = "test"
	mch.On("AckMessage", "msg-1").Return(nil)
	mch.On("DownloadMedia", "media-1").Return([]byte("imagedata"), "image/jpeg", nil)
	mch.On("SendMessage", "123", mock.Anything).Return(nil)

	router.ProcessImageMessage(mch, "123", "msg-1", "media-1", "que es esto?")

	mch.AssertCalled(t, "SendMessage", "123", mock.Anything)
}

func TestMessageRouter_ProcessImageMessage_NoCaption_DefaultPrompt(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("LoadConversation", mock.Anything, mock.Anything).Return([]domain.ConversationMessage{}, nil)

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "veo algo"})
	defer srv.Close()

	hooksRegistry := hooks.NewRegistry()
	convUC := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)
	router := NewMessageRouter(convUC, ai, nil, nil, nil, nil, hooksRegistry, nil, "123")

	mch := &test.MockMediaChannel{}
	mch.ChannelName = "test"
	mch.On("AckMessage", "msg-1").Return(nil)
	mch.On("DownloadMedia", "media-1").Return([]byte("img"), "image/png", nil)
	mch.On("SendMessage", "123", mock.Anything).Return(nil)

	router.ProcessImageMessage(mch, "123", "msg-1", "media-1", "")

	mch.AssertCalled(t, "SendMessage", "123", mock.Anything)
}

// --- handleMessage with orchestrator ---

func TestMessageRouter_handleMessage_WithOrchestrator(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("LoadConversation", mock.Anything, mock.Anything).Return([]domain.ConversationMessage{}, nil)

	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "orchestrator response"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "ai response"})
	defer srv.Close()

	hooksRegistry := hooks.NewRegistry()
	convUC := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)
	orch := NewAgentOrchestrator(provider, NewToolRegistry(), nil)

	router := NewMessageRouter(convUC, ai, nil, orch, nil, nil, hooksRegistry, nil, "123")

	response, err := router.handleMessage("123", "hola", "test")

	require.NoError(t, err)
	assert.Equal(t, "orchestrator response", response)
}

// --- ProcessMessage with error in handleMessage ---

func TestMessageRouter_ProcessMessage_HandleMessageError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", mock.Anything, mock.Anything, mock.Anything).Return(errStoreFailure)

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "response"})
	defer srv.Close()

	hooksRegistry := hooks.NewRegistry()
	convUC := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	router := NewMessageRouter(convUC, ai, nil, nil, nil, nil, hooksRegistry, nil, "123")

	ch := &test.MockChannel{ChannelName: "test"}
	ch.On("AckMessage", "msg-1").Return(nil)
	ch.On("SendMessage", "123", mock.Anything).Return(nil)

	router.ProcessMessage(ch, "123", "msg-1", "hola")

	// Should send error message
	ch.AssertCalled(t, "SendMessage", "123", "Perdón, hubo un error procesando tu mensaje. Intentá de nuevo.")
}
