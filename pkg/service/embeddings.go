package service

import (
	"encoding/json"

	"jarvis/pkg/domain"
)

type AIEmbedder struct {
	provider domain.AIProvider
}

func NewAIEmbedder(provider domain.AIProvider) *AIEmbedder {
	return &AIEmbedder{provider: provider}
}

// NewClaudeEmbedder is a backward-compatible alias for NewAIEmbedder.
func NewClaudeEmbedder(provider domain.AIProvider) *AIEmbedder {
	return NewAIEmbedder(provider)
}

const embeddingPrompt = `Genera un vector de embedding simplificado para el siguiente texto.
El vector debe tener exactamente 64 dimensiones, valores entre -1 y 1.
Responde SOLO con un array JSON de 64 numeros float, sin explicacion.
Ejemplo: [0.1, -0.3, 0.5, ...]`

func (e *AIEmbedder) Embed(text string) ([]float64, error) {
	resp, err := e.provider.Complete(embeddingPrompt, text)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrEmbedGenerate, err)
	}

	var vector []float64
	if err := json.Unmarshal([]byte(resp), &vector); err != nil {
		return nil, domain.Wrapf(domain.ErrEmbedParse, err)
	}

	return vector, nil
}

// MemoryService defines the contract for memory and conversation storage.
// Implementations: PGMemoryService.
type MemoryService interface {
	Save(content string, tags []string, embedding []float64) (int64, error)
	Search(queryEmbedding []float64, limit int) ([]domain.Memory, error)
	SearchFTS(query string, limit int) ([]domain.Memory, error)
	SearchHybrid(query string, queryEmbedding []float64, limit int, vecWeight, ftsWeight float64) ([]domain.Memory, error)
	Delete(id int64) error

	SaveConversation(sessionID, role, content string) error
	LoadConversation(sessionID string, limit int) ([]domain.ConversationMessage, error)
	ClearConversation(sessionID string) error
	ReplaceConversation(sessionID string, msgs []domain.ConversationMessage) error

	LogHabit(name string) error
	GetHabitStreak(name string) (int, int, error)
	ListHabitsToday() ([]string, error)

	ListExpenses(from, to string) ([]domain.Expense, error)

	PruneSessions(olderThanDays int) (int64, error)

	Close() error
}

// Embedder generates vector embeddings from text.
type Embedder interface {
	Embed(text string) ([]float64, error)
}

// FinanceService defines the contract for finance data access.
type FinanceService interface {
	SaveExpense(expense domain.ParsedExpense) error
}
