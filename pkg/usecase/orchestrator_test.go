package usecase

import (
	"testing"

	"jarvis/pkg/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentOrchestrator_BuildOrchestratorPrompt(t *testing.T) {
	agents := []domain.AgentDefinition{
		{ID: "finance", Name: "Finance", Description: "Handles money"},
		{ID: "dev", Name: "Dev", Description: "Handles code"},
	}

	o := NewAgentOrchestrator(nil, NewToolRegistry(), agents)
	prompt := o.buildOrchestratorPrompt("Base prompt\n")

	assert.Contains(t, prompt, "Base prompt")
	assert.Contains(t, prompt, "finance")
	assert.Contains(t, prompt, "Handles money")
	assert.Contains(t, prompt, "dev")
	assert.Contains(t, prompt, "delegate_to_agent")
}

func TestAgentOrchestrator_FilterTools_Allowed(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "save_expense"}, func(input map[string]any) (string, error) { return "ok", nil })
	reg.Register(domain.ToolDefinition{Name: "list_github_issues"}, func(input map[string]any) (string, error) { return "ok", nil })
	reg.Register(domain.ToolDefinition{Name: "save_note"}, func(input map[string]any) (string, error) { return "ok", nil })

	o := NewAgentOrchestrator(nil, reg, nil)
	agent := domain.AgentDefinition{
		ID:           "finance",
		AllowedTools: []string{"save_expense", "save_note"},
	}

	filtered := o.filterTools(agent)

	assert.Len(t, filtered.Definitions(), 2)

	_, err := filtered.Execute("save_expense", nil)
	assert.NoError(t, err)

	_, err = filtered.Execute("list_github_issues", nil)
	assert.Error(t, err)
}

func TestAgentOrchestrator_FilterTools_Denied(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "save_expense"}, func(input map[string]any) (string, error) { return "ok", nil })
	reg.Register(domain.ToolDefinition{Name: "dangerous_tool"}, func(input map[string]any) (string, error) { return "ok", nil })

	o := NewAgentOrchestrator(nil, reg, nil)
	agent := domain.AgentDefinition{
		ID:          "safe",
		DeniedTools: []string{"dangerous_tool"},
	}

	filtered := o.filterTools(agent)

	assert.Len(t, filtered.Definitions(), 1)
	assert.Equal(t, "save_expense", filtered.Definitions()[0].Name)
}

func TestAgentOrchestrator_FilterTools_NoDelegation(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "delegate_to_agent"}, func(input map[string]any) (string, error) { return "ok", nil })
	reg.Register(domain.ToolDefinition{Name: "save_note"}, func(input map[string]any) (string, error) { return "ok", nil })

	o := NewAgentOrchestrator(nil, reg, nil)
	agent := domain.AgentDefinition{ID: "sub"}

	filtered := o.filterTools(agent)

	// delegate_to_agent should be filtered out from sub-agents
	assert.Len(t, filtered.Definitions(), 1)
	assert.Equal(t, "save_note", filtered.Definitions()[0].Name)
}

func TestDefaultAgents_HasExpectedAgents(t *testing.T) {
	agents := DefaultAgents()

	assert.Len(t, agents, 5)

	ids := make(map[string]bool)
	for _, a := range agents {
		ids[a.ID] = true
		assert.NotEmpty(t, a.Name)
		assert.NotEmpty(t, a.Description)
		assert.NotEmpty(t, a.SystemPrompt)
	}

	assert.True(t, ids["finance"])
	assert.True(t, ids["dev"])
	assert.True(t, ids["study"])
	assert.True(t, ids["planner"])
	assert.True(t, ids["assistant"])
}

// --- agentIDs tests ---

func TestAgentOrchestrator_agentIDs(t *testing.T) {
	agents := []domain.AgentDefinition{
		{ID: "finance", Name: "Finance Agent"},
		{ID: "dev", Name: "Dev Agent"},
		{ID: "study", Name: "Study Agent"},
	}

	orch := NewAgentOrchestrator(nil, NewToolRegistry(), agents)

	ids := orch.agentIDs()

	assert.Equal(t, []string{"finance", "dev", "study"}, ids)
}

func TestAgentOrchestrator_agentIDs_Empty(t *testing.T) {
	orch := NewAgentOrchestrator(nil, NewToolRegistry(), nil)

	ids := orch.agentIDs()

	assert.Empty(t, ids)
}

// --- runSubAgent tests ---

func TestAgentOrchestrator_runSubAgent_NotFound(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{},
	}
	orch := NewAgentOrchestrator(provider, NewToolRegistry(), nil)

	_, err := orch.runSubAgent("nonexistent", "do something")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
}

func TestAgentOrchestrator_runSubAgent_Success(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "Sub-agent result"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	agents := []domain.AgentDefinition{
		{
			ID:           "helper",
			Name:         "Helper",
			SystemPrompt: "You are a helper.",
		},
	}

	orch := NewAgentOrchestrator(provider, NewToolRegistry(), agents)

	result, err := orch.runSubAgent("helper", "help me")

	require.NoError(t, err)
	assert.Equal(t, "Sub-agent result", result)
}

func TestAgentOrchestrator_runSubAgent_AIError_ReturnsErrorMessage(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks:     nil,
				stopReason: "",
				err:        assert.AnError,
			},
		},
	}

	agents := []domain.AgentDefinition{
		{
			ID:           "failing",
			Name:         "Failing Agent",
			SystemPrompt: "You fail.",
		},
	}

	orch := NewAgentOrchestrator(provider, NewToolRegistry(), agents)

	result, err := orch.runSubAgent("failing", "do it")

	require.NoError(t, err)
	assert.Contains(t, result, "Error del agente Failing Agent")
}

func TestAgentOrchestrator_runSubAgent_WithFilteredTools(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "Done with limited tools"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	tools := NewToolRegistry()
	tools.Register(domain.ToolDefinition{Name: "search"}, func(input map[string]any) (string, error) { return "ok", nil })
	tools.Register(domain.ToolDefinition{Name: "delete"}, func(input map[string]any) (string, error) { return "ok", nil })

	agents := []domain.AgentDefinition{
		{
			ID:           "limited",
			Name:         "Limited",
			SystemPrompt: "Limited agent.",
			AllowedTools: []string{"search"},
		},
	}

	orch := NewAgentOrchestrator(provider, tools, agents)

	result, err := orch.runSubAgent("limited", "search for something")

	require.NoError(t, err)
	assert.Equal(t, "Done with limited tools", result)
}

// --- buildOrchestratorTools tests ---

func TestAgentOrchestrator_buildOrchestratorTools_IncludesDelegation(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "save_note"}, func(input map[string]any) (string, error) { return "ok", nil })

	agents := []domain.AgentDefinition{
		{ID: "helper", Name: "Helper", Description: "Helps"},
	}

	o := NewAgentOrchestrator(nil, reg, agents)
	orchTools := o.buildOrchestratorTools()

	defs := orchTools.Definitions()

	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}

	assert.True(t, names["save_note"])
	assert.True(t, names["delegate_to_agent"])
}

func TestAgentOrchestrator_buildOrchestratorTools_CopiesExistingTools(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "tool_a"}, func(input map[string]any) (string, error) { return "a", nil })
	reg.Register(domain.ToolDefinition{Name: "tool_b"}, func(input map[string]any) (string, error) { return "b", nil })

	o := NewAgentOrchestrator(nil, reg, nil)
	orchTools := o.buildOrchestratorTools()

	// Should have tool_a, tool_b, and delegate_to_agent
	assert.Len(t, orchTools.Definitions(), 3)

	resultA, err := orchTools.Execute("tool_a", nil)
	assert.NoError(t, err)
	assert.Equal(t, "a", resultA)
}

// --- Run with delegation tests ---

func TestAgentOrchestrator_Run_WithDelegation(t *testing.T) {
	// Responses are consumed sequentially:
	// 1. Orchestrator gets tool_use (delegate_to_agent)
	// 2. Sub-agent gets text response (sub-agent result)
	// 3. Orchestrator gets final text response
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks: []domain.ContentBlock{
					{Type: "tool_use", ID: "t1", Name: "delegate_to_agent", Input: map[string]any{"agent_id": "helper", "task": "do it"}},
				},
				stopReason: domain.StopReasonToolUse,
			},
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "Sub-agent did it"}},
				stopReason: domain.StopReasonEndTurn,
			},
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "Final answer from orchestrator"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	agents := []domain.AgentDefinition{
		{
			ID:           "helper",
			Name:         "Helper",
			SystemPrompt: "You help.",
		},
	}

	orch := NewAgentOrchestrator(provider, NewToolRegistry(), agents)

	result, err := orch.Run("system", []domain.Message{{Role: "user", Content: "delegate this"}})

	require.NoError(t, err)
	assert.Equal(t, "Final answer from orchestrator", result)
	assert.Equal(t, 3, provider.calls)
}

func TestAgentOrchestrator_Run_SimpleText(t *testing.T) {
	provider := &mockToolUseProvider{
		responses: []mockToolResponse{
			{
				blocks:     []domain.ContentBlock{{Type: "text", Text: "Direct answer"}},
				stopReason: domain.StopReasonEndTurn,
			},
		},
	}

	orch := NewAgentOrchestrator(provider, NewToolRegistry(), nil)

	result, err := orch.Run("system", []domain.Message{{Role: "user", Content: "hi"}})

	require.NoError(t, err)
	assert.Equal(t, "Direct answer", result)
}
