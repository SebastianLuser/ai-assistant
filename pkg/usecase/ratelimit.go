package usecase

import (
	"sync"
	"time"
)

type toolLimit struct {
	maxCalls int
	window   time.Duration
	mu       sync.Mutex
	calls    []time.Time
}

func newToolLimit(maxCalls int, window time.Duration) *toolLimit {
	return &toolLimit{
		maxCalls: maxCalls,
		window:   window,
	}
}

func (tl *toolLimit) Allow() bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-tl.window)

	valid := tl.calls[:0]
	for _, t := range tl.calls {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	tl.calls = valid

	if len(tl.calls) >= tl.maxCalls {
		return false
	}

	tl.calls = append(tl.calls, now)
	return true
}

func (tl *toolLimit) Reset() {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.calls = tl.calls[:0]
}
