package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
)

type ctxKey int

const (
	keyTraceID ctxKey = iota
	keyChannel
	keyProfile
	keyClassifiedTags
)

func NewTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(b)
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyTraceID, id)
}

func TraceID(ctx context.Context) string {
	if v, ok := ctx.Value(keyTraceID).(string); ok {
		return v
	}
	return ""
}

func WithChannel(ctx context.Context, ch string) context.Context {
	return context.WithValue(ctx, keyChannel, ch)
}

func Channel(ctx context.Context) string {
	if v, ok := ctx.Value(keyChannel).(string); ok {
		return v
	}
	return ""
}

func WithProfile(ctx context.Context, p string) context.Context {
	return context.WithValue(ctx, keyProfile, p)
}

func Profile(ctx context.Context) string {
	if v, ok := ctx.Value(keyProfile).(string); ok {
		return v
	}
	return ""
}

func WithClassifiedTags(ctx context.Context, tags []string) context.Context {
	return context.WithValue(ctx, keyClassifiedTags, tags)
}

func ClassifiedTags(ctx context.Context) []string {
	if v, ok := ctx.Value(keyClassifiedTags).([]string); ok {
		return v
	}
	return nil
}

func Logger(ctx context.Context) *slog.Logger {
	l := slog.Default()
	if tid := TraceID(ctx); tid != "" {
		l = l.With("trace_id", tid)
	}
	if ch := Channel(ctx); ch != "" {
		l = l.With("channel", ch)
	}
	return l
}
