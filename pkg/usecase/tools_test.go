package usecase

import (
	"testing"
	"time"

	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Suppress unused import
var _ = time.Now

func TestToolRegistry_Register_And_Execute(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "greet"}, func(input map[string]any) (string, error) {
		return "hello " + inputString(input, "name"), nil
	})

	result, err := reg.Execute("greet", map[string]any{"name": "world"})

	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestToolRegistry_Execute_NotFound(t *testing.T) {
	reg := NewToolRegistry()

	_, err := reg.Execute("nonexistent", nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrToolNotFound)
}

func TestToolRegistry_Definitions_ReturnsAll(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "tool_a"}, func(input map[string]any) (string, error) { return "", nil })
	reg.Register(domain.ToolDefinition{Name: "tool_b"}, func(input map[string]any) (string, error) { return "", nil })

	defs := reg.Definitions()

	require.Len(t, defs, 2)
	assert.Equal(t, "tool_a", defs[0].Name)
	assert.Equal(t, "tool_b", defs[1].Name)
}

func TestToolRegistry_Definitions_Empty(t *testing.T) {
	reg := NewToolRegistry()

	defs := reg.Definitions()

	assert.Empty(t, defs)
}

func TestToolRegistry_Execute_HandlerError(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "fail"}, func(input map[string]any) (string, error) {
		return "", assert.AnError
	})

	_, err := reg.Execute("fail", map[string]any{})

	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestInputString_ExistingKey(t *testing.T) {
	input := map[string]any{"name": "test"}

	result := inputString(input, "name")

	assert.Equal(t, "test", result)
}

func TestInputString_MissingKey(t *testing.T) {
	input := map[string]any{"other": "value"}

	result := inputString(input, "name")

	assert.Equal(t, "", result)
}

func TestInputString_NonStringValue(t *testing.T) {
	input := map[string]any{"count": 42}

	result := inputString(input, "count")

	assert.Equal(t, "42", result)
}

func TestInputString_NilMap(t *testing.T) {
	result := inputString(nil, "key")

	assert.Equal(t, "", result)
}

func TestInputString_BoolValue(t *testing.T) {
	input := map[string]any{"flag": true}

	result := inputString(input, "flag")

	assert.Equal(t, "true", result)
}

func TestToJSON_Struct(t *testing.T) {
	data := struct {
		Name string `json:"name"`
	}{Name: "test"}

	result := toJSON(data)

	assert.Equal(t, `{"name":"test"}`, result)
}

func TestToJSON_Map(t *testing.T) {
	data := map[string]string{"key": "value"}

	result := toJSON(data)

	assert.Equal(t, `{"key":"value"}`, result)
}

func TestToJSON_Slice(t *testing.T) {
	data := []int{1, 2, 3}

	result := toJSON(data)

	assert.Equal(t, `[1,2,3]`, result)
}

func TestToJSON_String(t *testing.T) {
	result := toJSON("hello")

	assert.Equal(t, `"hello"`, result)
}

func TestToJSON_UnmarshalableValue(t *testing.T) {
	// Channels can't be marshaled to JSON
	ch := make(chan int)

	result := toJSON(ch)

	assert.NotEmpty(t, result)
}

// --- ReminderManager tests ---

func TestReminderManager_Schedule(t *testing.T) {
	called := make(chan string, 1)
	sender := func(text string) {
		called <- text
	}
	rm := NewReminderManager(sender)

	rm.Schedule(1*time.Millisecond, "test reminder")

	select {
	case msg := <-called:
		assert.Equal(t, "test reminder", msg)
	case <-time.After(1 * time.Second):
		t.Fatal("reminder did not fire within timeout")
	}
}

func TestReminderManager_Schedule_NilSender(t *testing.T) {
	rm := NewReminderManager(nil)

	// Should not panic
	rm.Schedule(1*time.Millisecond, "test")
	time.Sleep(10 * time.Millisecond)
}

// --- BuildToolRegistry tests ---

func TestBuildToolRegistry_NilClients_OnlyUsageTool(t *testing.T) {
	reg := BuildToolRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	defs := reg.Definitions()

	// Should have at least the get_usage tool
	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}
	assert.True(t, names["get_usage"])
}

func TestBuildToolRegistry_WithMemoryService_RegistersNoteTools(t *testing.T) {
	repo := new(test.MockMemoryService)

	reg := BuildToolRegistry(nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	defs := reg.Definitions()
	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}

	assert.True(t, names["save_note"])
	assert.True(t, names["search_notes"])
}

func TestBuildToolRegistry_WithReminderMgr_RegistersReminder(t *testing.T) {
	rm := NewReminderManager(nil)
	reg := BuildToolRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, rm)

	defs := reg.Definitions()
	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}

	assert.True(t, names["set_reminder"])
}

func TestBuildToolRegistry_SaveNoteHandler_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("Save", "my note", []string(nil), []float64(nil)).Return(int64(42), nil)

	reg := BuildToolRegistry(nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	result, err := reg.Execute("save_note", map[string]any{"content": "my note"})

	assert.NoError(t, err)
	assert.Contains(t, result, "42")
}

func TestBuildToolRegistry_SaveNoteHandler_WithEmbedder(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", "my note").Return([]float64{0.1, 0.2}, nil)
	repo.On("Save", "my note", []string(nil), []float64{0.1, 0.2}).Return(int64(1), nil)

	reg := BuildToolRegistry(nil, repo, embedder, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	result, err := reg.Execute("save_note", map[string]any{"content": "my note"})

	assert.NoError(t, err)
	assert.Contains(t, result, "1")
}

func TestBuildToolRegistry_SearchNotesHandler_FTS(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SearchFTS", "hello", 5).Return([]domain.Memory{{ID: 1, Content: "hello world"}}, nil)

	reg := BuildToolRegistry(nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	result, err := reg.Execute("search_notes", map[string]any{"query": "hello"})

	assert.NoError(t, err)
	assert.Contains(t, result, "hello world")
}

func TestBuildToolRegistry_SearchNotesHandler_WithEmbedder_Hybrid(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", "hello").Return([]float64{0.1}, nil)
	repo.On("SearchHybrid", "hello", []float64{0.1}, 5, 0.6, 0.4).Return([]domain.Memory{{ID: 1, Content: "hybrid"}}, nil)

	reg := BuildToolRegistry(nil, repo, embedder, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	result, err := reg.Execute("search_notes", map[string]any{"query": "hello"})

	assert.NoError(t, err)
	assert.Contains(t, result, "hybrid")
}

func TestBuildToolRegistry_GetUsageHandler(t *testing.T) {
	reg := BuildToolRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	result, err := reg.Execute("get_usage", map[string]any{})

	assert.NoError(t, err)
	assert.Contains(t, result, "Usage tracking")
}

func TestBuildToolRegistry_SetReminderHandler_Success(t *testing.T) {
	rm := NewReminderManager(nil)
	reg := BuildToolRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, rm)

	result, err := reg.Execute("set_reminder", map[string]any{"message": "call mom", "minutes": 30.0})

	assert.NoError(t, err)
	assert.Contains(t, result, "30 minutos")
}

func TestBuildToolRegistry_SetReminderHandler_InvalidMinutes(t *testing.T) {
	rm := NewReminderManager(nil)
	reg := BuildToolRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, rm)

	_, err := reg.Execute("set_reminder", map[string]any{"message": "test", "minutes": 0.0})

	assert.Error(t, err)
}

func TestBuildToolRegistry_SaveNoteHandler_WithTags(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("Save", "tagged note", []string{"important"}, []float64(nil)).Return(int64(1), nil)

	reg := BuildToolRegistry(nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	result, err := reg.Execute("save_note", map[string]any{
		"content": "tagged note",
		"tags":    []any{"important"},
	})

	assert.NoError(t, err)
	assert.Contains(t, result, "1")
}

func TestBuildToolRegistry_SearchNotesHandler_CustomLimit(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SearchFTS", "test", 10).Return([]domain.Memory{}, nil)

	reg := BuildToolRegistry(nil, repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	_, err := reg.Execute("search_notes", map[string]any{"query": "test", "limit": 10.0})

	assert.NoError(t, err)
}

func TestToolRegistry_MultipleRegistrations_LastWins(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(domain.ToolDefinition{Name: "tool"}, func(input map[string]any) (string, error) {
		return "first", nil
	})
	reg.Register(domain.ToolDefinition{Name: "tool"}, func(input map[string]any) (string, error) {
		return "second", nil
	})

	result, err := reg.Execute("tool", nil)

	require.NoError(t, err)
	assert.Equal(t, "second", result)
}
