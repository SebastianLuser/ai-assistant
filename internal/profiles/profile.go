package profiles

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	AllowedSkills []string `yaml:"allowed_skills"`
	AllowedTools  []string `yaml:"allowed_tools"`
	AllowedAgents []string `yaml:"allowed_agents"`
	AllowedRules  []string `yaml:"allowed_rules"`
	ExtraPrompt   string   `yaml:"extra_prompt"`
}

type Loader struct {
	dir string
}

func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

func (l *Loader) Load(name string) (*Profile, error) {
	path := filepath.Join(l.dir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profiles: load %s: %w", name, err)
	}

	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("profiles: parse %s: %w", name, err)
	}

	if p.Name == "" {
		p.Name = name
	}
	return &p, nil
}

func (l *Loader) LoadAll() ([]Profile, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("profiles: read dir: %w", err)
	}

	var result []Profile
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		name := entry.Name()[:len(entry.Name())-5]
		p, err := l.Load(name)
		if err != nil {
			continue
		}
		result = append(result, *p)
	}
	return result, nil
}

func (p *Profile) AllowsSkillTag(tag string) bool {
	if len(p.AllowedSkills) == 0 {
		return true
	}
	for _, s := range p.AllowedSkills {
		if s == tag {
			return true
		}
	}
	return false
}

func (p *Profile) AllowsTool(name string) bool {
	if len(p.AllowedTools) == 0 {
		return true
	}
	for _, t := range p.AllowedTools {
		if t == name {
			return true
		}
	}
	return false
}

func (p *Profile) AllowsAgent(id string) bool {
	if len(p.AllowedAgents) == 0 {
		return true
	}
	for _, a := range p.AllowedAgents {
		if a == id {
			return true
		}
	}
	return false
}

func (p *Profile) AllowsRule(name string) bool {
	if len(p.AllowedRules) == 0 {
		return true
	}
	for _, r := range p.AllowedRules {
		if r == name {
			return true
		}
	}
	return false
}
