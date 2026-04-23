package profiles

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	loader := NewLoader("../../config/profiles")
	p, err := loader.Load("work")

	require.NoError(t, err)
	assert.Equal(t, "work", p.Name)
	assert.Contains(t, p.AllowedSkills, "github")
	assert.Contains(t, p.AllowedTools, "list_github_issues")
	assert.Contains(t, p.AllowedAgents, "dev")
	assert.NotEmpty(t, p.ExtraPrompt)
}

func TestLoader_Load_NotFound(t *testing.T) {
	loader := NewLoader("../../config/profiles")
	_, err := loader.Load("nonexistent")

	assert.Error(t, err)
}

func TestLoader_LoadAll(t *testing.T) {
	loader := NewLoader("../../config/profiles")
	profiles, err := loader.LoadAll()

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(profiles), 3)

	names := make(map[string]bool)
	for _, p := range profiles {
		names[p.Name] = true
	}
	assert.True(t, names["work"])
	assert.True(t, names["personal"])
	assert.True(t, names["study"])
}

func TestProfile_AllowsSkillTag_EmptyAllowsAll(t *testing.T) {
	p := Profile{AllowedSkills: []string{}}
	assert.True(t, p.AllowsSkillTag("anything"))
}

func TestProfile_AllowsSkillTag_Filtered(t *testing.T) {
	p := Profile{AllowedSkills: []string{"finance", "sheets"}}

	assert.True(t, p.AllowsSkillTag("finance"))
	assert.False(t, p.AllowsSkillTag("github"))
}

func TestProfile_AllowsTool(t *testing.T) {
	p := Profile{AllowedTools: []string{"save_expense"}}

	assert.True(t, p.AllowsTool("save_expense"))
	assert.False(t, p.AllowsTool("list_github_issues"))
}

func TestProfile_AllowsTool_EmptyAllowsAll(t *testing.T) {
	p := Profile{AllowedTools: []string{}}
	assert.True(t, p.AllowsTool("anything"))
}

func TestProfile_AllowsAgent(t *testing.T) {
	p := Profile{AllowedAgents: []string{"finance", "planner"}}

	assert.True(t, p.AllowsAgent("finance"))
	assert.False(t, p.AllowsAgent("dev"))
}

func TestProfile_AllowsRule(t *testing.T) {
	p := Profile{AllowedRules: []string{"finance-rules"}}

	assert.True(t, p.AllowsRule("finance-rules"))
	assert.False(t, p.AllowsRule("dev-rules"))
}
