package rules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_WithFrontmatter(t *testing.T) {
	raw := `---
name: test-rule
description: A test rule
triggers:
  tags: [finance]
  time_range: "06:00-12:00"
  channel: whatsapp
  day_of_week: [monday, friday]
---
## Rule content
Some instructions here.`

	rule, err := parse(raw)

	require.NoError(t, err)
	assert.Equal(t, "test-rule", rule.Name)
	assert.Equal(t, "A test rule", rule.Description)
	assert.Equal(t, []string{"finance"}, rule.Triggers.Tags)
	assert.Equal(t, "06:00-12:00", rule.Triggers.TimeRange)
	assert.Equal(t, "whatsapp", rule.Triggers.Channel)
	assert.Equal(t, []string{"monday", "friday"}, rule.Triggers.DayOfWeek)
	assert.Contains(t, rule.Content, "Rule content")
}

func TestParse_WithoutFrontmatter(t *testing.T) {
	raw := "Just plain content"

	rule, err := parse(raw)

	require.NoError(t, err)
	assert.Equal(t, "", rule.Name)
	assert.Equal(t, "Just plain content", rule.Content)
}

func TestMatchRules_ByTags(t *testing.T) {
	rules := []Rule{
		{Name: "finance", Triggers: Triggers{Tags: []string{"finance", "sheets"}}},
		{Name: "dev", Triggers: Triggers{Tags: []string{"github", "code"}}},
		{Name: "habits", Triggers: Triggers{Tags: []string{"habits"}}},
	}

	matched := MatchRules(rules, []string{"finance"}, time.Now(), "whatsapp")

	require.Len(t, matched, 1)
	assert.Equal(t, "finance", matched[0].Name)
}

func TestMatchRules_ByMultipleTags(t *testing.T) {
	rules := []Rule{
		{Name: "finance", Triggers: Triggers{Tags: []string{"finance"}}},
		{Name: "dev", Triggers: Triggers{Tags: []string{"github"}}},
	}

	matched := MatchRules(rules, []string{"finance", "github"}, time.Now(), "")

	assert.Len(t, matched, 2)
}

func TestMatchRules_ByTimeRange(t *testing.T) {
	rules := []Rule{
		{Name: "morning", Triggers: Triggers{TimeRange: "06:00-12:00"}},
	}

	morning := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	afternoon := time.Date(2026, 4, 20, 15, 0, 0, 0, time.UTC)

	assert.Len(t, MatchRules(rules, nil, morning, ""), 1)
	assert.Len(t, MatchRules(rules, nil, afternoon, ""), 0)
}

func TestMatchRules_ByTimeRange_WrapsMidnight(t *testing.T) {
	rules := []Rule{
		{Name: "night", Triggers: Triggers{TimeRange: "22:00-06:00"}},
	}

	lateNight := time.Date(2026, 4, 20, 23, 30, 0, 0, time.UTC)
	earlyMorning := time.Date(2026, 4, 20, 3, 0, 0, 0, time.UTC)
	noon := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

	assert.Len(t, MatchRules(rules, nil, lateNight, ""), 1)
	assert.Len(t, MatchRules(rules, nil, earlyMorning, ""), 1)
	assert.Len(t, MatchRules(rules, nil, noon, ""), 0)
}

func TestMatchRules_ByChannel(t *testing.T) {
	rules := []Rule{
		{Name: "wa-only", Triggers: Triggers{Channel: "whatsapp"}},
	}

	assert.Len(t, MatchRules(rules, nil, time.Now(), "whatsapp"), 1)
	assert.Len(t, MatchRules(rules, nil, time.Now(), "telegram"), 0)
}

func TestMatchRules_ByDayOfWeek(t *testing.T) {
	rules := []Rule{
		{Name: "weekday", Triggers: Triggers{DayOfWeek: []string{"monday", "tuesday", "wednesday", "thursday", "friday"}}},
	}

	monday := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC) // Monday
	saturday := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

	assert.Len(t, MatchRules(rules, nil, monday, ""), 1)
	assert.Len(t, MatchRules(rules, nil, saturday, ""), 0)
}

func TestMatchRules_CombinedConditions(t *testing.T) {
	rules := []Rule{
		{
			Name: "finance-morning-wa",
			Triggers: Triggers{
				Tags:      []string{"finance"},
				TimeRange: "06:00-12:00",
				Channel:   "whatsapp",
			},
		},
	}

	morning := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	afternoon := time.Date(2026, 4, 20, 15, 0, 0, 0, time.UTC)

	// All conditions met.
	assert.Len(t, MatchRules(rules, []string{"finance"}, morning, "whatsapp"), 1)

	// Tag matches but time doesn't.
	assert.Len(t, MatchRules(rules, []string{"finance"}, afternoon, "whatsapp"), 0)

	// Tag and time match but wrong channel.
	assert.Len(t, MatchRules(rules, []string{"finance"}, morning, "telegram"), 0)
}

func TestMatchRules_NoConditions(t *testing.T) {
	rules := []Rule{
		{Name: "empty", Triggers: Triggers{}},
	}

	assert.Len(t, MatchRules(rules, []string{"finance"}, time.Now(), "whatsapp"), 0)
}

func TestMatchRules_EmptyTags(t *testing.T) {
	rules := []Rule{
		{Name: "time-only", Triggers: Triggers{TimeRange: "00:00-23:59"}},
	}

	assert.Len(t, MatchRules(rules, nil, time.Now(), ""), 1)
}

func TestFormatForPrompt_Empty(t *testing.T) {
	assert.Equal(t, "", FormatForPrompt(nil))
}

func TestFormatForPrompt_WithRules(t *testing.T) {
	rules := []Rule{
		{Content: "Rule 1 content"},
		{Content: "Rule 2 content"},
	}

	result := FormatForPrompt(rules)

	assert.Contains(t, result, "Reglas activas")
	assert.Contains(t, result, "Rule 1 content")
	assert.Contains(t, result, "Rule 2 content")
}

func TestLoader_LoadAll(t *testing.T) {
	loader := NewLoader("../../rules")
	rules, err := loader.LoadAll()

	require.NoError(t, err)
	assert.Greater(t, len(rules), 0)

	names := make(map[string]bool)
	for _, r := range rules {
		names[r.Name] = true
		assert.NotEmpty(t, r.Content)
	}

	assert.True(t, names["finance-rules"])
	assert.True(t, names["dev-rules"])
}

func TestMatchesTimeRange_InvalidFormat(t *testing.T) {
	assert.False(t, matchesTimeRange("invalid", time.Now()))
	assert.False(t, matchesTimeRange("abc-def", time.Now()))
	assert.False(t, matchesTimeRange("", time.Now()))
}
