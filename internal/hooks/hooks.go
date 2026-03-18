package hooks

import (
	"context"
	"log"
	"sync"
	"time"
)

const (
	MessageReceived          = "message_received"
	BeforeResponse           = "before_response"
	AfterCompaction          = "after_compaction"
	CronJobCompleted         = "cron_job_completed"
	WhatsAppMessageProcessed = "whatsapp_message_processed" // deprecated: use MessageProcessed
	MessageProcessed         = "message_processed"
)

type Event struct {
	Type      string
	Payload   any
	Timestamp time.Time
}

type HookFunc func(ctx context.Context, event Event) error

type Registry struct {
	mu    sync.RWMutex
	hooks map[string][]HookFunc
}

func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string][]HookFunc),
	}
}

// Register adds a hook function for the given event type.
func (r *Registry) Register(eventType string, fn HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[eventType] = append(r.hooks[eventType], fn)
}

// Emit fires all hooks for the given event type sequentially.
// Errors are logged but do not stop execution.
func (r *Registry) Emit(ctx context.Context, eventType string, payload any) {
	r.mu.RLock()
	fns := r.hooks[eventType]
	r.mu.RUnlock()

	if len(fns) == 0 {
		return
	}

	event := Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	for _, fn := range fns {
		if err := fn(ctx, event); err != nil {
			log.Printf("hooks: %s handler error: %v", eventType, err)
		}
	}
}
