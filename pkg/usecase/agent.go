package usecase

import (
	"context"
	"strings"

	"jarvis/internal/skills"
	"jarvis/internal/tracing"
	"jarvis/pkg/domain"
)

const agentMaxTurns = 10

// AgentUseCase runs an agentic loop where Claude decides which tools to call.
type AgentUseCase struct {
	ai          domain.ToolUseProvider
	tools       *ToolRegistry
	maxTurns    int
	skillsCache []skills.Skill
}

// NewAgentUseCase creates an agent that can use tools via Claude's tool use API.
func NewAgentUseCase(ai domain.ToolUseProvider, tools *ToolRegistry) *AgentUseCase {
	return &AgentUseCase{
		ai:       ai,
		tools:    tools,
		maxTurns: agentMaxTurns,
	}
}

// SetSkills provides the loaded skills for trigger-based composition.
func (a *AgentUseCase) SetSkills(s []skills.Skill) {
	a.skillsCache = s
}

// Run executes the agentic loop: Claude decides tools → we execute → Claude continues.
// Returns the final text response.
func (a *AgentUseCase) Run(ctx context.Context, system string, messages []domain.Message) (string, error) {
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
		log := tracing.Logger(ctx)
		var resultBlocks []domain.ContentBlock
		var triggeredExtra []skills.Skill
		for _, tc := range toolCalls {
			confirmed := inputBool(tc.Input, "confirmed")
			useDryRun := a.tools.IsDryRun(tc.Name) && a.tools.HasPreview(tc.Name) && !confirmed

			var result string
			var err error
			if useDryRun {
				log.Info("agent: preview tool", "turn", turn, "tool", tc.Name, "id", tc.ID)
				result, err = a.tools.Preview(ctx, tc.Name, tc.Input)
			} else {
				log.Info("agent: executing tool", "turn", turn, "tool", tc.Name, "id", tc.ID)
				result, err = a.tools.Execute(ctx, tc.Name, tc.Input)
			}
			if err != nil {
				log.Warn("agent: tool failed", "tool", tc.Name, "err", err)
				result = err.Error()
			} else if len(a.skillsCache) > 0 {
				triggered := skills.FindTriggeredSkills(a.skillsCache, tc.Name)
				triggeredExtra = append(triggeredExtra, triggered...)
			}

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

		if len(triggeredExtra) > 0 {
			extra := skills.FormatForPrompt(triggeredExtra)
			log.Info("agent: triggered skills", "count", len(triggeredExtra))
			msgs = append(msgs, domain.Message{
				Role:    domain.RoleUser,
				Content: "[Sistema: ejecuta tambien esto]\n" + extra,
			})
		}
	}

	return "", domain.ErrAgentMaxTurns
}
