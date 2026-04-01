package usecase

import (
	"testing"

	"jarvis/pkg/domain"

	"github.com/stretchr/testify/assert"
)

type mockToolUseProvider struct {
	calls      int
	responses  []mockToolResponse
}

type mockToolResponse struct {
	blocks     []domain.ContentBlock
	stopReason string
	err        error
}

func (m *mockToolUseProvider) CompleteWithTools(system string, messages []domain.Message, tools []domain.ToolDefinition, opts ...domain.CompletionOption) ([]domain.ContentBlock, string, error) {
	if m.calls >= len(m.responses) {
		return []domain.ContentBlock{{Type: "text", Text: "done"}}, domain.StopReasonEndTurn, nil
	}
	resp := m.responses[m.calls]
	m.calls++
	return resp.blocks, resp.stopReason, resp.err
}

func TestAgentUseCase_Run_SimpleText(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "Hola!"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	agent := NewAgentUseCase(provider, NewToolRegistry())
	result, err := agent.Run("system", []domain.Message{{Role: "user", Content: "hola"}})

	assert.NoError(t, err)
	assert.Equal(t, "Hola!", result)
}

func TestAgentUseCase_Run_ToolCall(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks: []domain.ContentBlock{
					{Type: "text", Text: "Voy a buscar..."},
					{Type: "tool_use", ID: "t1", Name: "test_tool", Input: map[string]any{"q": "test"}},
				},
				stopReason: domain.StopReasonToolUse,
			},
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "Encontré: result"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "test_tool"}, func(input map[string]any) (string, error) {
		return "result for " + inputString(input, "q"), nil
	})

	agent := NewAgentUseCase(provider, reg)
	result, err := agent.Run("system", []domain.Message{{Role: "user", Content: "buscar test"}})

	assert.NoError(t, err)
	assert.Equal(t, "Encontré: result", result)
	assert.Equal(t, 2, provider.calls)
}

func TestAgentUseCase_Run_MaxTurnsExceeded(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: make([]mockToolResponse, 20),
	}
	for i := range provider.responses {
		provider.responses[i] = mockToolResponse{
			blocks: []domain.ContentBlock{
				{Type: "tool_use", ID: "t1", Name: "loop_tool", Input: map[string]any{}},
			},
			stopReason: domain.StopReasonToolUse,
		}
	}

	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "loop_tool"}, func(input map[string]any) (string, error) {
		return "ok", nil
	})

	agent := &AgentUseCase{ai: provider, tools: reg, maxTurns: 3}
	_, err := agent.Run("system", []domain.Message{{Role: "user", Content: "loop"}})

	assert.ErrorIs(t, err, domain.ErrAgentMaxTurns)
}

func TestAgentUseCase_Run_ToolError(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks: []domain.ContentBlock{
					{Type: "tool_use", ID: "t1", Name: "fail_tool", Input: map[string]any{}},
				},
				stopReason: domain.StopReasonToolUse,
			},
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "Tool failed, sorry"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "fail_tool"}, func(input map[string]any) (string, error) {
		return "", assert.AnError
	})

	agent := NewAgentUseCase(provider, reg)
	result, err := agent.Run("system", []domain.Message{{Role: "user", Content: "do it"}})

	assert.NoError(t, err)
	assert.Equal(t, "Tool failed, sorry", result)
}
