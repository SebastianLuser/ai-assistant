package agents

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadAll(t *testing.T) {
	loader := NewLoader("../../agents")
	agents, err := loader.LoadAll()

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(agents), 5)

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

func TestLoader_LoadAll_FinanceAgent(t *testing.T) {
	loader := NewLoader("../../agents")
	agents, err := loader.LoadAll()
	require.NoError(t, err)

	var finance *struct {
		found bool
		agent interface{ AllowedTools() []string }
	}

	for _, a := range agents {
		if a.ID == "finance" {
			assert.Equal(t, "Finance Agent", a.Name)
			assert.Contains(t, a.AllowedTools, "save_expense")
			assert.Contains(t, a.DeniedTools, "delegate_to_agent")
			assert.Contains(t, a.DelegatesTo, "planner")
			assert.NotEmpty(t, a.VetoRules)
			assert.Equal(t, 10, a.Priority)
			assert.Equal(t, 2000, a.MaxTokenBudget)
			assert.Contains(t, a.SystemPrompt, "Personalidad")
			return
		}
	}
	_ = finance
	t.Fatal("finance agent not found")
}

func TestParse_WithFrontmatter(t *testing.T) {
	raw := `---
id: test
name: Test Agent
description: A test agent
allowed_tools: [tool1, tool2]
veto_rules:
  - "Rule 1"
priority: 5
---
## Content
Agent persona content here.`

	agent, err := parse(raw)

	require.NoError(t, err)
	assert.Equal(t, "test", agent.ID)
	assert.Equal(t, "Test Agent", agent.Name)
	assert.Equal(t, []string{"tool1", "tool2"}, agent.AllowedTools)
	assert.Equal(t, []string{"Rule 1"}, agent.VetoRules)
	assert.Equal(t, 5, agent.Priority)
	assert.Contains(t, agent.SystemPrompt, "Content")
}

func TestParse_WithoutFrontmatter(t *testing.T) {
	raw := "Just plain content"

	agent, err := parse(raw)

	require.NoError(t, err)
	assert.Equal(t, "", agent.ID)
	assert.Equal(t, "Just plain content", agent.SystemPrompt)
}
