package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Rule struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    Triggers `yaml:"triggers"`
	Content     string   `yaml:"-"`
}

type Triggers struct {
	Tags      []string `yaml:"tags"`
	TimeRange string   `yaml:"time_range"`
	Channel   string   `yaml:"channel"`
	DayOfWeek []string `yaml:"day_of_week"`
}

type Loader struct {
	dir string
}

func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

func (l *Loader) LoadAll() ([]Rule, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("rules: read dir: %w", err)
	}

	var result []Rule
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(l.dir, entry.Name()))
		if err != nil {
			continue
		}

		rule, err := parse(string(raw))
		if err != nil {
			continue
		}

		if rule.Name == "" {
			rule.Name = strings.TrimSuffix(entry.Name(), ".md")
		}

		result = append(result, rule)
	}

	return result, nil
}

func MatchRules(rules []Rule, tags []string, now time.Time, channel string) []Rule {
	tagSet := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		tagSet[t] = struct{}{}
	}

	var matched []Rule
	for _, r := range rules {
		if matchesRule(r, tagSet, now, channel) {
			matched = append(matched, r)
		}
	}
	return matched
}

func matchesRule(r Rule, tagSet map[string]struct{}, now time.Time, channel string) bool {
	hasCondition := false

	if len(r.Triggers.Tags) > 0 {
		hasCondition = true
		if !matchesTags(r.Triggers.Tags, tagSet) {
			return false
		}
	}

	if r.Triggers.TimeRange != "" {
		hasCondition = true
		if !matchesTimeRange(r.Triggers.TimeRange, now) {
			return false
		}
	}

	if r.Triggers.Channel != "" {
		hasCondition = true
		if !strings.EqualFold(r.Triggers.Channel, channel) {
			return false
		}
	}

	if len(r.Triggers.DayOfWeek) > 0 {
		hasCondition = true
		if !matchesDayOfWeek(r.Triggers.DayOfWeek, now) {
			return false
		}
	}

	// A rule with no conditions never matches automatically.
	return hasCondition
}

func matchesTags(triggerTags []string, tagSet map[string]struct{}) bool {
	for _, t := range triggerTags {
		if _, ok := tagSet[t]; ok {
			return true
		}
	}
	return false
}

func matchesTimeRange(timeRange string, now time.Time) bool {
	parts := strings.SplitN(timeRange, "-", 2)
	if len(parts) != 2 {
		return false
	}

	start, err := time.Parse("15:04", strings.TrimSpace(parts[0]))
	if err != nil {
		return false
	}
	end, err := time.Parse("15:04", strings.TrimSpace(parts[1]))
	if err != nil {
		return false
	}

	current := now.Hour()*60 + now.Minute()
	startMin := start.Hour()*60 + start.Minute()
	endMin := end.Hour()*60 + end.Minute()

	if startMin <= endMin {
		return current >= startMin && current < endMin
	}
	// Wraps midnight (e.g., "22:00-06:00").
	return current >= startMin || current < endMin
}

func matchesDayOfWeek(days []string, now time.Time) bool {
	today := strings.ToLower(now.Weekday().String())
	for _, d := range days {
		if strings.EqualFold(d, today) {
			return true
		}
	}
	return false
}

func FormatForPrompt(rules []Rule) string {
	if len(rules) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Reglas activas\n\n")
	for _, r := range rules {
		sb.WriteString(r.Content)
		sb.WriteString("\n\n")
	}
	return sb.String()
}

func parse(raw string) (Rule, error) {
	var rule Rule

	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "---") {
		rule.Content = trimmed
		return rule, nil
	}

	rest := trimmed[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		rule.Content = trimmed
		return rule, nil
	}

	frontmatter := strings.TrimSpace(rest[:idx])
	body := strings.TrimSpace(rest[idx+4:])

	if err := yaml.Unmarshal([]byte(frontmatter), &rule); err != nil {
		return rule, fmt.Errorf("rules: parse frontmatter: %w", err)
	}

	rule.Content = body
	return rule, nil
}
