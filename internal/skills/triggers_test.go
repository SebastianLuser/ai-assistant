package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindTriggeredSkills_Match(t *testing.T) {
	allSkills := []Skill{
		{Name: "budget-check", AfterTool: "save_expense", Content: "Check budget"},
		{Name: "other", AfterTool: "create_event", Content: "Other"},
		{Name: "no-trigger", Content: "No trigger"},
	}

	triggered := FindTriggeredSkills(allSkills, "save_expense")

	assert.Len(t, triggered, 1)
	assert.Equal(t, "budget-check", triggered[0].Name)
}

func TestFindTriggeredSkills_NoMatch(t *testing.T) {
	allSkills := []Skill{
		{Name: "budget-check", AfterTool: "save_expense", Content: "Check budget"},
	}

	triggered := FindTriggeredSkills(allSkills, "create_task")

	assert.Empty(t, triggered)
}

func TestFindTriggeredSkills_DisabledSkillSkipped(t *testing.T) {
	disabled := false
	allSkills := []Skill{
		{Name: "disabled", AfterTool: "save_expense", Enabled: &disabled, Content: "Disabled"},
	}

	triggered := FindTriggeredSkills(allSkills, "save_expense")

	assert.Empty(t, triggered)
}

func TestFindTriggeredSkills_EmptySkills(t *testing.T) {
	triggered := FindTriggeredSkills(nil, "save_expense")

	assert.Empty(t, triggered)
}
