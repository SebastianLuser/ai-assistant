package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"jarvis/internal/hooks"
	"jarvis/internal/tracing"
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

// ToolHandler executes a tool and returns a text result.
type ToolHandler func(ctx context.Context, input map[string]any) (string, error)

// ToolRegistry holds tool definitions and their handler functions.
type ToolRegistry struct {
	definitions []domain.ToolDefinition
	handlers    map[string]ToolHandler
	previewers  map[string]ToolHandler
	dryRunTools map[string]bool
	limits      map[string]*toolLimit
	catalog     service.CatalogService
	hooks       *hooks.Registry
}

// NewToolRegistry creates an empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		handlers:    make(map[string]ToolHandler),
		previewers:  make(map[string]ToolHandler),
		dryRunTools: make(map[string]bool),
		limits:      make(map[string]*toolLimit),
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

// FilteredDefinitions returns tool definitions filtered to only the given names.
// If names is empty, returns all definitions.
func (r *ToolRegistry) FilteredDefinitions(names []string) []domain.ToolDefinition {
	if len(names) == 0 {
		return r.definitions
	}
	allowed := make(map[string]struct{}, len(names))
	for _, n := range names {
		allowed[n] = struct{}{}
	}
	var filtered []domain.ToolDefinition
	for _, d := range r.definitions {
		if _, ok := allowed[d.Name]; ok {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// RegisterWithLimit adds a tool with a rate limit (max calls per time window).
func (r *ToolRegistry) RegisterWithLimit(def domain.ToolDefinition, handler ToolHandler, maxCalls int, window time.Duration) {
	r.Register(def, handler)
	r.limits[def.Name] = newToolLimit(maxCalls, window)
}

// RegisterWithPreview adds a tool that supports dry-run preview.
func (r *ToolRegistry) RegisterWithPreview(def domain.ToolDefinition, handler, previewer ToolHandler) {
	r.Register(def, handler)
	r.previewers[def.Name] = previewer
}

// SetDryRunTools marks specific tools as dry-run by default.
func (r *ToolRegistry) SetDryRunTools(names []string) {
	for _, n := range names {
		r.dryRunTools[n] = true
	}
}

// IsDryRun checks if a tool should run in preview mode.
func (r *ToolRegistry) IsDryRun(name string) bool {
	return r.dryRunTools[name]
}

// Preview executes the previewer for a tool, returning a preview result.
func (r *ToolRegistry) Preview(ctx context.Context, name string, input map[string]any) (string, error) {
	previewer, ok := r.previewers[name]
	if !ok {
		return "", domain.Wrap(domain.ErrToolNotFound, name+" (no previewer)")
	}
	return previewer(ctx, input)
}

// HasPreview returns true if the tool has a registered previewer.
func (r *ToolRegistry) HasPreview(name string) bool {
	_, ok := r.previewers[name]
	return ok
}

// SetCatalog attaches a CatalogService for usage tracking.
func (r *ToolRegistry) SetCatalog(catalog service.CatalogService) {
	r.catalog = catalog
}

// SetHooks attaches a hooks registry for event emission on tool completion.
func (r *ToolRegistry) SetHooks(h *hooks.Registry) {
	r.hooks = h
}

// Execute runs a tool by name with the given input and records usage metrics.
func (r *ToolRegistry) Execute(ctx context.Context, name string, input map[string]any) (string, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return "", domain.Wrap(domain.ErrToolNotFound, name)
	}

	if lim, ok := r.limits[name]; ok && !lim.Allow() {
		return "", fmt.Errorf("rate limit: %s exceeded %d calls in %v, try a different approach", name, lim.maxCalls, lim.window)
	}

	result, err := handler(ctx, input)

	if r.catalog != nil {
		if recErr := r.catalog.RecordUsage(name, domain.CatalogTypeTool, err == nil); recErr != nil {
			tracing.Logger(ctx).Warn("catalog: record failed", "tool", name, "err", recErr)
		}
	}

	if err == nil && r.hooks != nil {
		r.hooks.Emit(ctx, hooks.ToolCompleted, map[string]string{"tool": name})
	}

	return result, err
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

// Helper to extract a boolean param from tool input.
func inputBool(input map[string]any, key string) bool {
	v, ok := input[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// Helper to marshal a result to JSON string.
func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
