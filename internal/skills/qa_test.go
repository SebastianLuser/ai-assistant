package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSkill_ValidSkill(t *testing.T) {
	rubric := DefaultRubric()
	skill := Skill{
		Name:        "finance",
		Description: "Registra gastos en lenguaje natural",
		Tags:        []string{"finance", "sheets"},
		Content:     "# Finanzas\nContenido del skill",
	}

	result := ValidateSkill(skill, rubric)

	assert.True(t, result.Valid)
	assert.Equal(t, 1.0, result.Score)
	assert.Empty(t, result.Issues)
}

func TestValidateSkill_MissingName(t *testing.T) {
	rubric := DefaultRubric()
	skill := Skill{
		Description: "A skill",
		Tags:        []string{"test"},
		Content:     "Content",
	}

	result := ValidateSkill(skill, rubric)

	assert.False(t, result.Valid)
	assert.Less(t, result.Score, 1.0)

	hasNameError := false
	for _, issue := range result.Issues {
		if issue.Rule == "required_field" && issue.Message == "name is required" {
			hasNameError = true
		}
	}
	assert.True(t, hasNameError)
}

func TestValidateSkill_MissingTags(t *testing.T) {
	rubric := DefaultRubric()
	skill := Skill{
		Name:        "test",
		Description: "A test skill with enough length",
		Tags:        []string{},
		Content:     "Content",
	}

	result := ValidateSkill(skill, rubric)

	assert.False(t, result.Valid)
}

func TestValidateSkill_EmptyContent(t *testing.T) {
	rubric := DefaultRubric()
	skill := Skill{
		Name:        "test",
		Description: "A test skill with enough length",
		Tags:        []string{"test"},
		Content:     "",
	}

	result := ValidateSkill(skill, rubric)

	assert.False(t, result.Valid)
}

func TestValidateSkill_ShortDescription(t *testing.T) {
	rubric := DefaultRubric()
	skill := Skill{
		Name:        "test",
		Description: "Short",
		Tags:        []string{"test"},
		Content:     "Content",
	}

	result := ValidateSkill(skill, rubric)

	assert.True(t, result.Valid) // warning, not error
	assert.Less(t, result.Score, 1.0)
}

func TestValidateSkill_TodoMarkers(t *testing.T) {
	rubric := DefaultRubric()
	skill := Skill{
		Name:        "test",
		Description: "A test skill with enough length",
		Tags:        []string{"test"},
		Content:     "Some content with TODO markers",
	}

	result := ValidateSkill(skill, rubric)

	assert.True(t, result.Valid) // warning only
	hasWIP := false
	for _, issue := range result.Issues {
		if issue.Rule == "wip_markers" {
			hasWIP = true
		}
	}
	assert.True(t, hasWIP)
}

func TestValidateAll_DetectsDuplicates(t *testing.T) {
	rubric := DefaultRubric()
	skills := []Skill{
		{Name: "finance", Description: "First finance skill", Tags: []string{"finance"}, Content: "Content 1"},
		{Name: "finance", Description: "Second finance skill", Tags: []string{"finance"}, Content: "Content 2"},
	}

	results := ValidateAll(skills, rubric)

	assert.Len(t, results, 2)
	assert.False(t, results[1].Valid)

	hasDuplicate := false
	for _, issue := range results[1].Issues {
		if issue.Rule == "duplicate_name" {
			hasDuplicate = true
		}
	}
	assert.True(t, hasDuplicate)
}

func TestValidateAll_WithRealSkills(t *testing.T) {
	loader := NewLoader("../../skills")
	skills, err := loader.LoadAll()
	if err != nil {
		t.Skip("skills directory not accessible")
	}

	rubric := DefaultRubric()
	results := ValidateAll(skills, rubric)

	for _, r := range results {
		assert.True(t, r.Valid, "skill %s should be valid: %v", r.Skill, r.Issues)
	}
}

func TestDefaultRubric(t *testing.T) {
	r := DefaultRubric()
	assert.Equal(t, 10, r.MinDescriptionLen)
	assert.Equal(t, 1, r.MinTags)
	assert.Greater(t, r.MaxContentLen, 0)
}
