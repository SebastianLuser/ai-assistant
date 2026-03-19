package usecase

import (
	"testing"

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
