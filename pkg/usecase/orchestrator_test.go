package usecase

import (
	"testing"

	"asistente/pkg/domain"

	"github.com/stretchr/testify/assert"
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
