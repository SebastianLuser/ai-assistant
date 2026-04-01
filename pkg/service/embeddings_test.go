package service

import (
	"errors"
	"testing"

	"jarvis/pkg/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errAIComplete = errors.New("ai failed")
	testVector64  = make([]float64, 64)
)

type stubAI struct {
	response string
	err      error
}

func (s *stubAI) Complete(system, userMessage string, opts ...domain.CompletionOption) (string, error) {
	return s.response, s.err
}

func (s *stubAI) CompleteMessages(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, error) {
	return s.response, s.err
}

func (s *stubAI) CompleteJSON(system, userMessage string, target any, opts ...domain.CompletionOption) error {
	return s.err
}

var _ domain.AIProvider = (*stubAI)(nil)

func TestAIEmbedder_Embed_Success(t *testing.T) {
	ai := &stubAI{response: "[0.1, 0.2, 0.3]"}
	embedder := NewAIEmbedder(ai)

	vec, err := embedder.Embed("hello")

	require.NoError(t, err)
	assert.Equal(t, []float64{0.1, 0.2, 0.3}, vec)
}

func TestAIEmbedder_Embed_AIError(t *testing.T) {
	ai := &stubAI{err: errAIComplete}
	embedder := NewAIEmbedder(ai)

	_, err := embedder.Embed("hello")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrEmbedGenerate))
}

func TestAIEmbedder_Embed_InvalidJSON(t *testing.T) {
	ai := &stubAI{response: "not json"}
	embedder := NewAIEmbedder(ai)

	_, err := embedder.Embed("hello")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrEmbedParse))
}

func TestNewClaudeEmbedder_IsAlias(t *testing.T) {
	ai := &stubAI{response: "[1.0]"}

	e1 := NewAIEmbedder(ai)
	e2 := NewClaudeEmbedder(ai)

	v1, _ := e1.Embed("x")
	v2, _ := e2.Embed("x")
	assert.Equal(t, v1, v2)
}
