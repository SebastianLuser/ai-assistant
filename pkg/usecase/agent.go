package usecase

import (
	"log"
	"strings"

	"asistente/pkg/domain"
)

const agentMaxTurns = 10

// AgentUseCase runs an agentic loop where Claude decides which tools to call.
type AgentUseCase struct {
	ai       domain.ToolUseProvider
	tools    *ToolRegistry
	maxTurns int
}

// NewAgentUseCase creates an agent that can use tools via Claude's tool use API.
func NewAgentUseCase(ai domain.ToolUseProvider, tools *ToolRegistry) *AgentUseCase {
	return &AgentUseCase{
		ai:       ai,
		tools:    tools,
		maxTurns: agentMaxTurns,
	}
}

// Run executes the agentic loop: Claude decides tools → we execute → Claude continues.
// Returns the final text response.
func (a *AgentUseCase) Run(system string, messages []domain.Message) (string, error) {
	msgs := make([]domain.Message, len(messages))
	copy(msgs, messages)

	toolDefs := a.tools.Definitions()

	for turn := 0; turn < a.maxTurns; turn++ {
		blocks, stopReason, err := a.ai.CompleteWithTools(system, msgs, toolDefs)
		if err != nil {
			return "", err
		}

		var textParts []string
		var toolCalls []domain.ToolCall

		for _, b := range blocks {
			switch b.Type {
			case "text":
				if b.Text != "" {
					textParts = append(textParts, b.Text)
				}
			case "tool_use":
				toolCalls = append(toolCalls, domain.ToolCall{
					ID: b.ID, Name: b.Name, Input: b.Input,
				})
			}
		}

		// If Claude is done or no tool calls, return the text.
		if stopReason == domain.StopReasonEndTurn || len(toolCalls) == 0 {
			return strings.Join(textParts, "\n"), nil
		}

		// Append assistant message with the raw blocks.
		assistantBlocks := make([]domain.ContentBlock, len(blocks))
		copy(assistantBlocks, blocks)
		msgs = append(msgs, domain.Message{
			Role:          domain.RoleAssistant,
			ContentBlocks: assistantBlocks,
		})

		// Execute each tool and collect results.
		var resultBlocks []domain.ContentBlock
		for _, tc := range toolCalls {
			log.Printf("agent: turn=%d executing tool %s (id=%s)", turn, tc.Name, tc.ID)

			result, err := a.tools.Execute(tc.Name, tc.Input)
			if err != nil {
				log.Printf("agent: tool %s failed: %v", tc.Name, err)
				result = err.Error()
			}

			// Claude API format: tool_result uses ID (mapped to tool_use_id)
			// and Text (mapped to content) by domainToClaudeMessages.
			resultBlocks = append(resultBlocks, domain.ContentBlock{
				Type: "tool_result",
				ID:   tc.ID,
				Text: result,
			})
		}

		// Append tool results as user message.
		msgs = append(msgs, domain.Message{
			Role:          domain.RoleUser,
			ContentBlocks: resultBlocks,
		})
	}

	return "", domain.ErrAgentMaxTurns
}
