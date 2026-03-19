package usecase

import (
	"testing"

	"asistente/pkg/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
