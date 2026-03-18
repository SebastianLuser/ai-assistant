package clients

import (
	"errors"
	"testing"

	"asistente/pkg/domain"

	"github.com/stretchr/testify/assert"
)

type stubProvider struct {
	response string
	err      error
	called   bool
}

func (s *stubProvider) Complete(system, userMessage string, opts ...domain.CompletionOption) (string, error) {
	s.called = true
	return s.response, s.err
}

func (s *stubProvider) CompleteMessages(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, error) {
	s.called = true
	return s.response, s.err
}

func (s *stubProvider) CompleteJSON(system, userMessage string, target any, opts ...domain.CompletionOption) error {
	s.called = true
	return s.err
}

func TestFailoverProvider_PrimarySucceeds(t *testing.T) {
	primary := &stubProvider{response: "primary ok"}
	fallback := &stubProvider{response: "fallback ok"}
	provider := NewFailoverProvider(primary, fallback)

	result, err := provider.Complete("sys", "msg")

	assert.NoError(t, err)
	assert.Equal(t, "primary ok", result)
	assert.True(t, primary.called)
	assert.False(t, fallback.called)
}

func TestFailoverProvider_PrimaryFails_FallbackSucceeds(t *testing.T) {
	primary := &stubProvider{err: errors.New("claude down")}
	fallback := &stubProvider{response: "fallback ok"}
	provider := NewFailoverProvider(primary, fallback)

	result, err := provider.Complete("sys", "msg")

	assert.NoError(t, err)
	assert.Equal(t, "fallback ok", result)
	assert.True(t, primary.called)
	assert.True(t, fallback.called)
}

func TestFailoverProvider_BothFail(t *testing.T) {
	primary := &stubProvider{err: errors.New("claude down")}
	fallback := &stubProvider{err: errors.New("openai down")}
	provider := NewFailoverProvider(primary, fallback)

	_, err := provider.Complete("sys", "msg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openai down")
}

func TestFailoverProvider_NilFallback_ReturnsPrimary(t *testing.T) {
	primary := &stubProvider{response: "only primary"}
	provider := NewFailoverProvider(primary, nil)

	result, err := provider.Complete("sys", "msg")

	assert.NoError(t, err)
	assert.Equal(t, "only primary", result)
}

func TestFailoverProvider_CompleteMessages(t *testing.T) {
	primary := &stubProvider{err: errors.New("fail")}
	fallback := &stubProvider{response: "fallback messages"}
	provider := NewFailoverProvider(primary, fallback)

	result, err := provider.CompleteMessages("sys", []domain.Message{{Role: "user", Content: "hi"}})

	assert.NoError(t, err)
	assert.Equal(t, "fallback messages", result)
}

func TestFailoverProvider_CompleteJSON(t *testing.T) {
	primary := &stubProvider{err: errors.New("fail")}
	fallback := &stubProvider{}
	provider := NewFailoverProvider(primary, fallback)

	err := provider.CompleteJSON("sys", "msg", nil)

	assert.NoError(t, err)
	assert.True(t, fallback.called)
}
