package hooks

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ExternalConfig struct {
	Hooks []ExternalHookDef `yaml:"hooks"`
}

func LoadExternalConfig(path string) ([]ExternalHookDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read hooks config: %w", err)
	}

	var cfg ExternalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse hooks config: %w", err)
	}

	for i, h := range cfg.Hooks {
		if h.Name == "" {
			return nil, fmt.Errorf("hook at index %d missing name", i)
		}
		if h.Event == "" {
			return nil, fmt.Errorf("hook %q missing event", h.Name)
		}
		if h.Type != "command" && h.Type != "webhook" {
			return nil, fmt.Errorf("hook %q has invalid type %q (must be command or webhook)", h.Name, h.Type)
		}
		if h.Type == "command" && h.Command == "" {
			return nil, fmt.Errorf("hook %q of type command missing command", h.Name)
		}
		if h.Type == "webhook" && h.URL == "" {
			return nil, fmt.Errorf("hook %q of type webhook missing url", h.Name)
		}
	}

	return cfg.Hooks, nil
}
