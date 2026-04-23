package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceID_Roundtrip(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", TraceID(ctx))

	ctx = WithTraceID(ctx, "abc123")
	assert.Equal(t, "abc123", TraceID(ctx))
}

func TestChannel_Roundtrip(t *testing.T) {
	ctx := WithChannel(context.Background(), "whatsapp")
	assert.Equal(t, "whatsapp", Channel(ctx))
}

func TestProfile_Roundtrip(t *testing.T) {
	ctx := WithProfile(context.Background(), "work")
	assert.Equal(t, "work", Profile(ctx))
}

func TestClassifiedTags_Roundtrip(t *testing.T) {
	ctx := WithClassifiedTags(context.Background(), []string{"finance", "sheets"})
	assert.Equal(t, []string{"finance", "sheets"}, ClassifiedTags(ctx))
}

func TestClassifiedTags_Empty(t *testing.T) {
	assert.Nil(t, ClassifiedTags(context.Background()))
}

func TestNewTraceID_Length(t *testing.T) {
	id := NewTraceID()
	assert.Len(t, id, 32)
}

func TestLogger_WithTraceID(t *testing.T) {
	ctx := WithTraceID(context.Background(), "test-trace")
	ctx = WithChannel(ctx, "telegram")
	l := Logger(ctx)
	assert.NotNil(t, l)
}
