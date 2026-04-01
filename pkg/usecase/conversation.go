package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"

	"jarvis/internal/hooks"
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

const chunkSummaryPrompt = `Resumi este fragmento de conversacion en 2-3 oraciones.
Preserva: hechos, decisiones, datos concretos (numeros, fechas, nombres).
Responde SOLO con el resumen, sin explicacion.`

const mergeSummaryPrompt = `Tenes varios resumenes parciales de una conversacion.
Mergealos en un unico resumen coherente y conciso (1 parrafo).
Prioriza el contexto mas reciente.
Preserva: hechos, decisiones, datos concretos (numeros, fechas, nombres), UUIDs, URLs.
Responde SOLO con el resumen final.`

const singlePassPrompt = `Resumi esta conversacion en un parrafo conciso.
Preserva:
- Hechos importantes mencionados
- Decisiones tomadas
- Datos concretos (numeros, fechas, nombres)
- Contexto necesario para continuar la conversacion

Se breve pero no pierdas informacion clave.`

type ConversationUseCase struct {
	store            service.MemoryService
	ai               domain.AIProvider
	hooks            *hooks.Registry
	maxHistory       int
	compactThreshold int
}

func NewConversationUseCase(store service.MemoryService, ai domain.AIProvider, hooksRegistry *hooks.Registry, maxHistory, compactThreshold int) *ConversationUseCase {
	if maxHistory <= 0 {
		maxHistory = domain.MaxHistoryMessages
	}
	if compactThreshold <= 0 {
		compactThreshold = domain.CompactThreshold
	}
	return &ConversationUseCase{
		store:            store,
		ai:               ai,
		hooks:            hooksRegistry,
		maxHistory:       maxHistory,
		compactThreshold: compactThreshold,
	}
}

func (uc *ConversationUseCase) Ingest(sessionID, role, content string) error {
	return uc.store.SaveConversation(sessionID, role, content)
}

func (uc *ConversationUseCase) Assemble(sessionID string) ([]domain.Message, error) {
	msgs, err := uc.store.LoadConversation(sessionID, uc.maxHistory)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrConversationLoad, err)
	}

	if len(msgs) >= uc.compactThreshold {
		if err := uc.Compact(sessionID); err != nil {
			return nil, domain.Wrapf(domain.ErrCompactFailed, err)
		}
		msgs, err = uc.store.LoadConversation(sessionID, uc.maxHistory)
		if err != nil {
			return nil, domain.Wrapf(domain.ErrConversationLoad, err)
		}
	}

	result := make([]domain.Message, len(msgs))
	for i, m := range msgs {
		result[i] = domain.Message{Role: m.Role, Content: m.Content}
	}

	return result, nil
}

func (uc *ConversationUseCase) Compact(sessionID string) error {
	msgs, err := uc.store.LoadConversation(sessionID, uc.maxHistory)
	if err != nil {
		return err
	}

	if len(msgs) < domain.MinMessagesToCompact {
		return nil
	}

	// Emit before compaction so hooks can save important data.
	uc.hooks.Emit(context.Background(), hooks.BeforeCompaction, map[string]string{
		"session_id":    sessionID,
		"message_count": fmt.Sprintf("%d", len(msgs)),
	})

	summary, err := uc.compactMultiStage(msgs)
	if err != nil {
		log.Printf("context: multi-stage compact failed, falling back to single-pass: %v", err)
		summary, err = uc.compactSinglePass(msgs)
		if err != nil {
			return domain.Wrapf(domain.ErrCompactFailed, err)
		}
	}

	compacted := []domain.ConversationMessage{
		{Role: domain.RoleAssistant, Content: domain.CompactSummaryPrefix + summary},
	}

	if err := uc.store.ReplaceConversation(sessionID, compacted); err != nil {
		return err
	}

	uc.hooks.Emit(context.Background(), hooks.AfterCompaction, map[string]string{
		"session_id": sessionID, "summary": summary,
	})

	return nil
}

func (uc *ConversationUseCase) compactMultiStage(msgs []domain.ConversationMessage) (string, error) {
	if len(msgs) <= domain.ChunkSize {
		return uc.compactSinglePass(msgs)
	}

	chunks := chunkMessages(msgs, domain.ChunkSize)

	summaries, err := uc.summarizeChunks(chunks)
	if err != nil {
		return "", err
	}

	if len(summaries) == 1 {
		return summaries[0], nil
	}

	return uc.mergeSummaries(summaries)
}

func chunkMessages(msgs []domain.ConversationMessage, size int) [][]domain.ConversationMessage {
	var chunks [][]domain.ConversationMessage
	for i := 0; i < len(msgs); i += size {
		end := i + size
		if end > len(msgs) {
			end = len(msgs)
		}
		chunks = append(chunks, msgs[i:end])
	}
	return chunks
}

func (uc *ConversationUseCase) summarizeChunks(chunks [][]domain.ConversationMessage) ([]string, error) {
	var summaries []string
	for _, chunk := range chunks {
		summary, err := uc.ai.Complete(chunkSummaryPrompt, FormatMessages(chunk))
		if err != nil {
			return nil, domain.Wrapf(domain.ErrCompactChunk, err)
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (uc *ConversationUseCase) mergeSummaries(summaries []string) (string, error) {
	all := strings.Join(summaries, "\n\n---\n\n")
	merged, err := uc.ai.Complete(mergeSummaryPrompt, all)
	if err != nil {
		return "", domain.Wrapf(domain.ErrCompactMerge, err)
	}
	return merged, nil
}

func (uc *ConversationUseCase) compactSinglePass(msgs []domain.ConversationMessage) (string, error) {
	return uc.ai.Complete(singlePassPrompt, FormatMessages(msgs))
}

// FormatMessages formats conversation messages into a readable string for AI prompts.
func FormatMessages(msgs []domain.ConversationMessage) string {
	var sb strings.Builder
	for _, m := range msgs {
		sb.WriteString("[" + m.Role + "]: " + m.Content + "\n")
	}
	return sb.String()
}
