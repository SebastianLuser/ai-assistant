package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Skill struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Enabled     *bool    `yaml:"enabled"`
	Tags        []string `yaml:"tags"`
	DependsOn   []string `yaml:"depends_on"`
	Content     string   `yaml:"-"`
}

func (s Skill) IsEnabled() bool {
	if s.Enabled == nil {
		return true
	}
	return *s.Enabled
}

type Loader struct {
	dir string
}

func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

// LoadAll reads all .md files from the skills directory, parses YAML frontmatter,
// and returns the parsed skills. Disabled skills are included but marked.
func (l *Loader) LoadAll() ([]Skill, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("skills: read dir: %w", err)
	}

	var result []Skill
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(l.dir, entry.Name()))
		if err != nil {
			continue
		}

		skill, err := parse(string(raw))
		if err != nil {
			continue
		}

		if skill.Name == "" {
			skill.Name = strings.TrimSuffix(entry.Name(), ".md")
		}

		result = append(result, skill)
	}

	return result, nil
}

// LoadEnabled returns only skills that are enabled.
func (l *Loader) LoadEnabled() ([]Skill, error) {
	all, err := l.LoadAll()
	if err != nil {
		return nil, err
	}

	var enabled []Skill
	for _, s := range all {
		if s.IsEnabled() {
			enabled = append(enabled, s)
		}
	}

	return enabled, nil
}

// FilterByTags returns skills that have at least one of the given tags.
func FilterByTags(skills []Skill, tags ...string) []Skill {
	if len(tags) == 0 {
		return skills
	}

	tagSet := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		tagSet[t] = struct{}{}
	}

	var result []Skill
	for _, s := range skills {
		for _, st := range s.Tags {
			if _, ok := tagSet[st]; ok {
				result = append(result, s)
				break
			}
		}
	}

	return result
}

// FormatForPrompt returns the skill contents joined for injection into a system prompt.
func FormatForPrompt(skills []Skill) string {
	var sb strings.Builder
	for _, s := range skills {
		sb.WriteString(s.Content)
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// Save writes a new skill to disk as a markdown file with YAML frontmatter.
// The filename is derived from the skill name (slugified).
func (l *Loader) Save(skill Skill) error {
	if skill.Name == "" {
		return fmt.Errorf("skills: name is required")
	}

	filename := slugify(skill.Name) + ".md"
	path := filepath.Join(l.dir, filename)

	enabled := skill.IsEnabled()
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", skill.Name))
	if skill.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %s\n", skill.Description))
	}
	sb.WriteString(fmt.Sprintf("enabled: %t\n", enabled))
	if len(skill.Tags) > 0 {
		sb.WriteString("tags:\n")
		for _, t := range skill.Tags {
			sb.WriteString(fmt.Sprintf("  - %s\n", t))
		}
	}
	sb.WriteString("---\n\n")
	sb.WriteString(skill.Content)
	sb.WriteString("\n")

	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

// slugify converts a skill name to a safe filename.
func slugify(name string) string {
	lower := strings.ToLower(name)
	lower = strings.ReplaceAll(lower, " ", "-")
	var safe strings.Builder
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			safe.WriteRune(r)
		}
	}
	return safe.String()
}

// parse splits a markdown file into YAML frontmatter and body content.
func parse(raw string) (Skill, error) {
	var skill Skill

	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "---") {
		skill.Content = trimmed
		return skill, nil
	}

	// Find closing ---
	rest := trimmed[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		skill.Content = trimmed
		return skill, nil
	}

	frontmatter := strings.TrimSpace(rest[:idx])
	body := strings.TrimSpace(rest[idx+4:])

	if err := yaml.Unmarshal([]byte(frontmatter), &skill); err != nil {
		return skill, fmt.Errorf("skills: parse frontmatter: %w", err)
	}

	skill.Content = body
	return skill, nil
}
