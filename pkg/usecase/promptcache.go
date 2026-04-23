package usecase

import (
	"sync"
	"sync/atomic"
	"time"
)

type promptEntry struct {
	prompt    string
	createdAt time.Time
}

type PromptCache struct {
	mu      sync.RWMutex
	entries map[string]promptEntry
	ttl     time.Duration
	hits    int64
	misses  int64
}

func NewPromptCache(ttl time.Duration) *PromptCache {
	return &PromptCache{
		entries: make(map[string]promptEntry),
		ttl:     ttl,
	}
}

func (pc *PromptCache) Get(key string) (string, bool) {
	pc.mu.RLock()
	entry, ok := pc.entries[key]
	pc.mu.RUnlock()

	if !ok || time.Since(entry.createdAt) > pc.ttl {
		atomic.AddInt64(&pc.misses, 1)
		return "", false
	}
	atomic.AddInt64(&pc.hits, 1)
	return entry.prompt, true
}

func (pc *PromptCache) Set(key, prompt string) {
	pc.mu.Lock()
	pc.entries[key] = promptEntry{prompt: prompt, createdAt: time.Now()}
	pc.mu.Unlock()
}

func (pc *PromptCache) Clear() {
	pc.mu.Lock()
	pc.entries = make(map[string]promptEntry)
	pc.mu.Unlock()
}

func (pc *PromptCache) Stats() (hits, misses int64) {
	return atomic.LoadInt64(&pc.hits), atomic.LoadInt64(&pc.misses)
}
