package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPromptCache_SetAndGet(t *testing.T) {
	pc := NewPromptCache(5 * time.Minute)

	pc.Set("key1", "prompt content")

	val, ok := pc.Get("key1")

	assert.True(t, ok)
	assert.Equal(t, "prompt content", val)
}

func TestPromptCache_Miss(t *testing.T) {
	pc := NewPromptCache(5 * time.Minute)

	val, ok := pc.Get("nonexistent")

	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestPromptCache_Expired(t *testing.T) {
	pc := NewPromptCache(1 * time.Millisecond)

	pc.Set("key1", "prompt content")
	time.Sleep(5 * time.Millisecond)

	val, ok := pc.Get("key1")

	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestPromptCache_Stats(t *testing.T) {
	pc := NewPromptCache(5 * time.Minute)

	pc.Set("key1", "value")
	pc.Get("key1")
	pc.Get("key1")
	pc.Get("missing")

	hits, misses := pc.Stats()

	assert.Equal(t, int64(2), hits)
	assert.Equal(t, int64(1), misses)
}

func TestPromptCache_Clear(t *testing.T) {
	pc := NewPromptCache(5 * time.Minute)

	pc.Set("key1", "value")
	pc.Clear()

	_, ok := pc.Get("key1")

	assert.False(t, ok)
}
