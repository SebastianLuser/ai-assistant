package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jarvis/pkg/domain"

	"gopkg.in/yaml.v3"
)

type Loader struct {
	dir string
}

func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

func (l *Loader) LoadAll() ([]domain.AgentDefinition, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("agents: read dir: %w", err)
	}

	var result []domain.AgentDefinition
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(l.dir, entry.Name()))
		if err != nil {
			continue
		}

		agent, err := parse(string(raw))
		if err != nil {
			continue
		}

		if agent.ID == "" {
			agent.ID = strings.TrimSuffix(entry.Name(), ".md")
		}

		result = append(result, agent)
	}

	return result, nil
}

func parse(raw string) (domain.AgentDefinition, error) {
	var agent domain.AgentDefinition

	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "---") {
		agent.SystemPrompt = trimmed
		return agent, nil
	}

	rest := trimmed[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		agent.SystemPrompt = trimmed
		return agent, nil
	}

	frontmatter := strings.TrimSpace(rest[:idx])
	body := strings.TrimSpace(rest[idx+4:])

	if err := yaml.Unmarshal([]byte(frontmatter), &agent); err != nil {
		return agent, fmt.Errorf("agents: parse frontmatter: %w", err)
	}

	agent.SystemPrompt = body
	return agent, nil
}
