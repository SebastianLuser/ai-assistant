package hooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchConditions_Empty(t *testing.T) {
	assert.True(t, matchConditions(nil, "anything"))
	assert.True(t, matchConditions(map[string]string{}, "anything"))
}

func TestMatchConditions_MapPayload(t *testing.T) {
	payload := map[string]any{"intent": "expense", "channel": "whatsapp"}

	assert.True(t, matchConditions(map[string]string{"intent": "expense"}, payload))
	assert.False(t, matchConditions(map[string]string{"intent": "note"}, payload))
	assert.False(t, matchConditions(map[string]string{"missing": "key"}, payload))
}

func TestMatchConditions_StructPayload(t *testing.T) {
	type Msg struct {
		Type string `json:"type"`
	}
	payload := Msg{Type: "text"}

	assert.True(t, matchConditions(map[string]string{"type": "text"}, payload))
	assert.False(t, matchConditions(map[string]string{"type": "audio"}, payload))
}

func TestExternalHookDef_Timeout(t *testing.T) {
	def := ExternalHookDef{Timeout: "5s"}
	assert.Equal(t, 5*time.Second, def.timeout())

	def2 := ExternalHookDef{}
	assert.Equal(t, 10*time.Second, def2.timeout())

	def3 := ExternalHookDef{Timeout: "invalid"}
	assert.Equal(t, 10*time.Second, def3.timeout())
}

func TestExternalHookDef_Method(t *testing.T) {
	def := ExternalHookDef{}
	assert.Equal(t, http.MethodPost, def.method())

	def2 := ExternalHookDef{Method: "PUT"}
	assert.Equal(t, "PUT", def2.method())
}

func TestExecuteWebhook_Success(t *testing.T) {
	var received []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = json.Marshal(map[string]string{"method": r.Method, "content_type": r.Header.Get("Content-Type")})
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	def := ExternalHookDef{
		Name:    "test-hook",
		Type:    "webhook",
		URL:     server.URL,
		Timeout: "5s",
	}

	event := Event{Type: "test", Payload: map[string]string{"key": "value"}, Timestamp: time.Now()}
	err := executeWebhook(context.Background(), def, event)

	require.NoError(t, err)
	require.NotNil(t, received)

	var info map[string]string
	json.Unmarshal(received, &info)
	assert.Equal(t, "POST", info["method"])
	assert.Equal(t, "application/json", info["content_type"])
}

func TestExecuteWebhook_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	def := ExternalHookDef{
		Name:    "test-hook",
		Type:    "webhook",
		URL:     server.URL,
		Timeout: "5s",
	}

	event := Event{Type: "test", Timestamp: time.Now()}
	err := executeWebhook(context.Background(), def, event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned 500")
}

func TestExecuteCommand_Success(t *testing.T) {
	def := ExternalHookDef{
		Name:    "echo-hook",
		Type:    "command",
		Command: "cat > /dev/null",
		Timeout: "5s",
	}

	event := Event{Type: "test", Payload: "data", Timestamp: time.Now()}
	err := executeCommand(context.Background(), def, event)

	assert.NoError(t, err)
}

func TestExecuteExternal_UnknownType(t *testing.T) {
	def := ExternalHookDef{Type: "ftp"}
	event := Event{Type: "test", Timestamp: time.Now()}

	err := executeExternal(context.Background(), def, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown hook type")
}

func TestRegistry_RegisterExternal_EmitTriggersWebhook(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	registry := NewRegistry()
	registry.RegisterExternal([]ExternalHookDef{{
		Name:    "test",
		Event:   MessageProcessed,
		Type:    "webhook",
		URL:     server.URL,
		Timeout: "5s",
	}})

	registry.Emit(context.Background(), MessageProcessed, nil)

	assert.True(t, called)
}

func TestRegistry_RegisterExternal_ConditionsMismatch_SkipsHook(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	registry := NewRegistry()
	registry.RegisterExternal([]ExternalHookDef{{
		Name:       "conditional",
		Event:      MessageProcessed,
		Type:       "webhook",
		URL:        server.URL,
		Timeout:    "5s",
		Conditions: map[string]string{"intent": "expense"},
	}})

	registry.Emit(context.Background(), MessageProcessed, map[string]any{"intent": "note"})

	assert.False(t, called)
}

func TestRegistry_RegisterExternal_ConditionsMatch_RunsHook(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	registry := NewRegistry()
	registry.RegisterExternal([]ExternalHookDef{{
		Name:       "conditional",
		Event:      MessageProcessed,
		Type:       "webhook",
		URL:        server.URL,
		Timeout:    "5s",
		Conditions: map[string]string{"intent": "expense"},
	}})

	registry.Emit(context.Background(), MessageProcessed, map[string]any{"intent": "expense"})

	assert.True(t, called)
}

func TestRegistry_External_DifferentEvent_NotTriggered(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	registry := NewRegistry()
	registry.RegisterExternal([]ExternalHookDef{{
		Name:  "test",
		Event: CronJobCompleted,
		Type:  "webhook",
		URL:   server.URL,
	}})

	registry.Emit(context.Background(), MessageProcessed, nil)

	assert.False(t, called)
}
