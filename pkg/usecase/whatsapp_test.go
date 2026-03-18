package usecase

import (
	"testing"

	"asistente/pkg/domain"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hel...", truncate("hello world", 3))
}

func TestGeneratePairingCode(t *testing.T) {
	code := generatePairingCode()

	assert.Len(t, code, 8)
	assert.NotEqual(t, "00000000", code)
}

func TestToolRegistry_Execute(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(
		domain.ToolDefinition{Name: "test_tool", Description: "test"},
		func(input map[string]any) (string, error) {
			return "result: " + inputString(input, "key"), nil
		},
	)

	result, err := reg.Execute("test_tool", map[string]any{"key": "value"})

	assert.NoError(t, err)
	assert.Equal(t, "result: value", result)
}

func TestToolRegistry_Execute_NotFound(t *testing.T) {
	reg := NewToolRegistry()

	_, err := reg.Execute("nonexistent", nil)

	assert.Error(t, err)
}

func TestInputString(t *testing.T) {
	assert.Equal(t, "hello", inputString(map[string]any{"k": "hello"}, "k"))
	assert.Equal(t, "", inputString(map[string]any{}, "k"))
	assert.Equal(t, "42", inputString(map[string]any{"k": 42}, "k"))
}
