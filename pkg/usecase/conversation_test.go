package usecase

import (
	"context"
	"errors"
	"testing"

	"jarvis/internal/hooks"
	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testSessionID = "test-session"
)

var (
	errStoreFailure = errors.New("store failure")
)

func buildMessages(n int) []domain.ConversationMessage {
	msgs := make([]domain.ConversationMessage, n)
	for i := 0; i < n; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		msgs[i] = domain.ConversationMessage{Role: role, Content: "message"}
	}
	return msgs
}

func newTestConversationUseCase(repo *test.MockMemoryService, claudeText string) (*ConversationUseCase, func()) {
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: claudeText})
	hooksRegistry := hooks.NewRegistry()
	uc := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)
	return uc, srv.Close
}

func TestConversationUseCase_Ingest_SavesMessage(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", testSessionID, "user", "hello").Return(nil)
	uc, cleanup := newTestConversationUseCase(repo, "")
	defer cleanup()

	err := uc.Ingest(testSessionID, "user", "hello")

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestConversationUseCase_Ingest_PropagatesStoreError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SaveConversation", testSessionID, "user", "hello").Return(errStoreFailure)
	uc, cleanup := newTestConversationUseCase(repo, "")
	defer cleanup()

	err := uc.Ingest(testSessionID, "user", "hello")

	require.Error(t, err)
	assert.True(t, errors.Is(err, errStoreFailure))
}

func TestConversationUseCase_Assemble_ReturnsMessages(t *testing.T) {
	repo := new(test.MockMemoryService)
	storedMsgs := []domain.ConversationMessage{
		{Role: "user", Content: "hola"},
		{Role: "assistant", Content: "hola, como estas?"},
	}
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).Return(storedMsgs, nil)
	uc, cleanup := newTestConversationUseCase(repo, "")
	defer cleanup()

	msgs, err := uc.Assemble(testSessionID)

	require.NoError(t, err)
	require.Len(t, msgs, 2)
	assert.Equal(t, "user", msgs[0].Role)
	assert.Equal(t, "hola", msgs[0].Content)
	assert.Equal(t, "assistant", msgs[1].Role)
}

func TestConversationUseCase_Assemble_LoadError_ReturnsError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).
		Return([]domain.ConversationMessage(nil), errStoreFailure)
	uc, cleanup := newTestConversationUseCase(repo, "")
	defer cleanup()

	_, err := uc.Assemble(testSessionID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrConversationLoad))
}

func TestConversationUseCase_Assemble_TriggersCompactWhenOverThreshold(t *testing.T) {
	repo := new(test.MockMemoryService)
	longHistory := buildMessages(domain.CompactThreshold)
	compactedHistory := []domain.ConversationMessage{
		{Role: "assistant", Content: "Resumen de conversacion anterior: summary"},
	}

	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).
		Return(longHistory, nil).Once()
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).
		Return(longHistory, nil).Once()
	repo.On("ReplaceConversation", testSessionID, mock.Anything).Return(nil)
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).
		Return(compactedHistory, nil).Once()

	uc, cleanup := newTestConversationUseCase(repo, "summary")
	defer cleanup()

	msgs, err := uc.Assemble(testSessionID)

	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, "Resumen de conversacion anterior: summary", msgs[0].Content)
	repo.AssertExpectations(t)
}

func TestConversationUseCase_Compact_SkipsShortConversations(t *testing.T) {
	repo := new(test.MockMemoryService)
	shortHistory := buildMessages(3)
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).Return(shortHistory, nil)
	uc, cleanup := newTestConversationUseCase(repo, "")
	defer cleanup()

	err := uc.Compact(testSessionID)

	require.NoError(t, err)
	repo.AssertNotCalled(t, "ReplaceConversation")
}

func TestConversationUseCase_Compact_ReplacesWithSummary(t *testing.T) {
	repo := new(test.MockMemoryService)
	history := buildMessages(8)
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).Return(history, nil)
	repo.On("ReplaceConversation", testSessionID, mock.MatchedBy(func(msgs []domain.ConversationMessage) bool {
		return len(msgs) == 1 && msgs[0].Role == "assistant"
	})).Return(nil)

	uc, cleanup := newTestConversationUseCase(repo, "resumen compactado")
	defer cleanup()

	err := uc.Compact(testSessionID)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestConversationUseCase_Compact_EmitsAfterCompactionHook(t *testing.T) {
	repo := new(test.MockMemoryService)
	history := buildMessages(8)
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).Return(history, nil)
	repo.On("ReplaceConversation", testSessionID, mock.Anything).Return(nil)

	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "resumen"})
	defer srv.Close()
	hooksRegistry := hooks.NewRegistry()
	hookCalled := false
	hooksRegistry.Register(hooks.AfterCompaction, func(_ context.Context, event hooks.Event) error {
		hookCalled = true
		return nil
	})
	uc := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	uc.Compact(testSessionID)

	assert.True(t, hookCalled)
}

// --- chunkMessages tests ---

func TestChunkMessages_ExactChunks(t *testing.T) {
	msgs := buildMessages(10)

	chunks := chunkMessages(msgs, 5)

	assert.Len(t, chunks, 2)
	assert.Len(t, chunks[0], 5)
	assert.Len(t, chunks[1], 5)
}

func TestChunkMessages_UnevenChunks(t *testing.T) {
	msgs := buildMessages(7)

	chunks := chunkMessages(msgs, 3)

	assert.Len(t, chunks, 3)
	assert.Len(t, chunks[0], 3)
	assert.Len(t, chunks[1], 3)
	assert.Len(t, chunks[2], 1)
}

func TestChunkMessages_SingleChunk(t *testing.T) {
	msgs := buildMessages(3)

	chunks := chunkMessages(msgs, 10)

	assert.Len(t, chunks, 1)
	assert.Len(t, chunks[0], 3)
}

func TestChunkMessages_EmptyMessages(t *testing.T) {
	chunks := chunkMessages(nil, 5)

	assert.Empty(t, chunks)
}

// --- compactMultiStage tests ---

func TestConversationUseCase_compactMultiStage_SmallInput_DelegatesToSinglePass(t *testing.T) {
	repo := new(test.MockMemoryService)
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "single pass summary"})
	defer srv.Close()
	hooksRegistry := hooks.NewRegistry()
	uc := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	msgs := buildMessages(domain.ChunkSize)

	summary, err := uc.compactMultiStage(msgs)

	require.NoError(t, err)
	assert.Equal(t, "single pass summary", summary)
}

func TestConversationUseCase_compactMultiStage_MultiChunk_MergesSummaries(t *testing.T) {
	repo := new(test.MockMemoryService)
	// Responses: chunk1 summary, chunk2 summary, merged summary
	srv, ai := test.NewMockClaudeServer(
		test.ClaudeResponse{Text: "chunk1 summary"},
		test.ClaudeResponse{Text: "chunk2 summary"},
		test.ClaudeResponse{Text: "merged summary"},
	)
	defer srv.Close()
	hooksRegistry := hooks.NewRegistry()
	uc := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	msgs := buildMessages(domain.ChunkSize * 2)

	summary, err := uc.compactMultiStage(msgs)

	require.NoError(t, err)
	assert.Equal(t, "merged summary", summary)
}

func TestConversationUseCase_compactMultiStage_SingleChunkResult(t *testing.T) {
	repo := new(test.MockMemoryService)
	// Only one chunk so only one summary, no merge needed
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "only chunk"})
	defer srv.Close()
	hooksRegistry := hooks.NewRegistry()
	uc := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	msgs := buildMessages(domain.ChunkSize + 1)

	summary, err := uc.compactMultiStage(msgs)

	require.NoError(t, err)
	assert.Equal(t, "only chunk", summary)
}

// --- summarizeChunks tests ---

func TestConversationUseCase_summarizeChunks_AIError(t *testing.T) {
	repo := new(test.MockMemoryService)
	srv, ai := test.NewMockClaudeServerError(500, "server_error", "internal")
	defer srv.Close()
	hooksRegistry := hooks.NewRegistry()
	uc := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	chunks := [][]domain.ConversationMessage{buildMessages(3)}

	_, err := uc.summarizeChunks(chunks)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrCompactChunk))
}

// --- mergeSummaries tests ---

func TestConversationUseCase_mergeSummaries_AIError(t *testing.T) {
	repo := new(test.MockMemoryService)
	srv, ai := test.NewMockClaudeServerError(500, "server_error", "internal")
	defer srv.Close()
	hooksRegistry := hooks.NewRegistry()
	uc := NewConversationUseCase(repo, ai, hooksRegistry, 0, 0)

	_, err := uc.mergeSummaries([]string{"summary1", "summary2"})

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrCompactMerge))
}

// --- Compact edge cases ---

func TestConversationUseCase_Compact_LoadError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).
		Return([]domain.ConversationMessage(nil), errStoreFailure)
	uc, cleanup := newTestConversationUseCase(repo, "")
	defer cleanup()

	err := uc.Compact(testSessionID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errStoreFailure))
}

func TestConversationUseCase_Compact_ReplaceError(t *testing.T) {
	repo := new(test.MockMemoryService)
	history := buildMessages(8)
	repo.On("LoadConversation", testSessionID, domain.MaxHistoryMessages).Return(history, nil)
	repo.On("ReplaceConversation", testSessionID, mock.Anything).Return(errStoreFailure)

	uc, cleanup := newTestConversationUseCase(repo, "summary")
	defer cleanup()

	err := uc.Compact(testSessionID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errStoreFailure))
}

func TestFormatMessages_FormatsCorrectly(t *testing.T) {
	msgs := []domain.ConversationMessage{
		{Role: "user", Content: "hola"},
		{Role: "assistant", Content: "chau"},
	}

	result := FormatMessages(msgs)

	expected := "[user]: hola\n[assistant]: chau\n"
	assert.Equal(t, expected, result)
}
