package usecase

import (
	"encoding/json"
	"fmt"

	"jarvis/pkg/domain"
)

// ToolHandler executes a tool and returns a text result.
type ToolHandler func(input map[string]any) (string, error)

// ToolRegistry holds tool definitions and their handler functions.
type ToolRegistry struct {
	definitions []domain.ToolDefinition
	handlers    map[string]ToolHandler
}

// NewToolRegistry creates an empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		handlers: make(map[string]ToolHandler),
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(def domain.ToolDefinition, handler ToolHandler) {
	r.definitions = append(r.definitions, def)
	r.handlers[def.Name] = handler
}

// Definitions returns all registered tool definitions.
func (r *ToolRegistry) Definitions() []domain.ToolDefinition {
	return r.definitions
}

// Execute runs a tool by name with the given input.
func (r *ToolRegistry) Execute(name string, input map[string]any) (string, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return "", domain.Wrap(domain.ErrToolNotFound, name)
	}
	return handler(input)
}

// Helper to extract a string param from tool input.
func inputString(input map[string]any, key string) string {
	v, ok := input[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

// Helper to marshal a result to JSON string.
func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
