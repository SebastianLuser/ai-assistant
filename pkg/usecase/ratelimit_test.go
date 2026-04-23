package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToolLimit_Allow_WithinLimit(t *testing.T) {
	lim := newToolLimit(3, time.Second)

	assert.True(t, lim.Allow())
	assert.True(t, lim.Allow())
	assert.True(t, lim.Allow())
}

func TestToolLimit_Allow_Exceeded(t *testing.T) {
	lim := newToolLimit(2, time.Second)

	assert.True(t, lim.Allow())
	assert.True(t, lim.Allow())
	assert.False(t, lim.Allow())
}

func TestToolLimit_Allow_WindowExpires(t *testing.T) {
	lim := newToolLimit(1, 50*time.Millisecond)

	assert.True(t, lim.Allow())
	assert.False(t, lim.Allow())

	time.Sleep(60 * time.Millisecond)

	assert.True(t, lim.Allow())
}

func TestToolLimit_Reset(t *testing.T) {
	lim := newToolLimit(1, time.Second)

	assert.True(t, lim.Allow())
	assert.False(t, lim.Allow())

	lim.Reset()

	assert.True(t, lim.Allow())
}
