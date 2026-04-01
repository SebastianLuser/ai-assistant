package clients

import (
	"log"

	"jarvis/pkg/domain"
)

// FailoverProvider wraps a primary and fallback AIProvider.
// If the primary fails, it transparently retries with the fallback.
type FailoverProvider struct {
	primary  domain.AIProvider
	fallback domain.AIProvider
}

// NewFailoverProvider creates a provider that falls back to a secondary on error.
// If fallback is nil, it behaves identically to the primary.
func NewFailoverProvider(primary, fallback domain.AIProvider) domain.AIProvider {
	if fallback == nil {
		return primary
	}
	return &FailoverProvider{primary: primary, fallback: fallback}
}

func (f *FailoverProvider) Complete(system, userMessage string, opts ...domain.CompletionOption) (string, error) {
	result, err := f.primary.Complete(system, userMessage, opts...)
	if err == nil {
		return result, nil
	}

	log.Printf("ai-failover: primary failed (%v), trying fallback", err)
	return f.fallback.Complete(system, userMessage, opts...)
}

func (f *FailoverProvider) CompleteMessages(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, error) {
	result, err := f.primary.CompleteMessages(system, messages, opts...)
	if err == nil {
		return result, nil
	}

	log.Printf("ai-failover: primary failed (%v), trying fallback", err)
	return f.fallback.CompleteMessages(system, messages, opts...)
}

func (f *FailoverProvider) CompleteJSON(system, userMessage string, target any, opts ...domain.CompletionOption) error {
	err := f.primary.CompleteJSON(system, userMessage, target, opts...)
	if err == nil {
		return nil
	}

	log.Printf("ai-failover: primary failed (%v), trying fallback", err)
	return f.fallback.CompleteJSON(system, userMessage, target, opts...)
}

// CompleteWithTools implements ToolUseProvider with failover.
func (f *FailoverProvider) CompleteWithTools(system string, messages []domain.Message, tools []domain.ToolDefinition, opts ...domain.CompletionOption) ([]domain.ContentBlock, string, error) {
	if tp, ok := f.primary.(domain.ToolUseProvider); ok {
		blocks, stop, err := tp.CompleteWithTools(system, messages, tools, opts...)
		if err == nil {
			return blocks, stop, nil
		}
		log.Printf("ai-failover: primary tool use failed (%v), trying fallback", err)
	}

	if tp, ok := f.fallback.(domain.ToolUseProvider); ok {
		return tp.CompleteWithTools(system, messages, tools, opts...)
	}

	return nil, "", domain.Wrap(domain.ErrClaudeAPI, "no tool use provider available")
}
