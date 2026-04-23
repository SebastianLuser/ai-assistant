package usecase

import (
	"context"
	"testing"

	"jarvis/pkg/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolRegistry_DryRun_Preview(t *testing.T) {
	reg := NewToolRegistry()

	executed := false
	previewed := false

	def := domain.ToolDefinition{Name: "save_expense", Description: "Save expense"}
	reg.RegisterWithPreview(def,
		func(_ context.Context, input map[string]any) (string, error) {
			executed = true
			return "saved", nil
		},
		func(_ context.Context, input map[string]any) (string, error) {
			previewed = true
			return `{"tool":"save_expense","would_do":"Registrar $5000 en Comida","confirm_hint":"Decí 'dale' para confirmar"}`, nil
		},
	)
	reg.SetDryRunTools([]string{"save_expense"})

	assert.True(t, reg.IsDryRun("save_expense"))
	assert.True(t, reg.HasPreview("save_expense"))

	result, err := reg.Preview(context.Background(),"save_expense", map[string]any{"amount": 5000})

	require.NoError(t, err)
	assert.True(t, previewed)
	assert.False(t, executed)
	assert.Contains(t, result, "would_do")
}

func TestToolRegistry_DryRun_ExecuteWithConfirm(t *testing.T) {
	reg := NewToolRegistry()

	def := domain.ToolDefinition{Name: "save_expense", Description: "Save expense"}
	reg.RegisterWithPreview(def,
		func(_ context.Context, input map[string]any) (string, error) { return "saved", nil },
		func(_ context.Context, input map[string]any) (string, error) { return "preview", nil },
	)
	reg.SetDryRunTools([]string{"save_expense"})

	result, err := reg.Execute(context.Background(),"save_expense", map[string]any{"confirmed": true})

	require.NoError(t, err)
	assert.Equal(t, "saved", result)
}

func TestToolRegistry_NoDryRun_ExecutesDirectly(t *testing.T) {
	reg := NewToolRegistry()

	def := domain.ToolDefinition{Name: "get_usage", Description: "Get usage"}
	reg.Register(def, func(_ context.Context, input map[string]any) (string, error) {
		return "usage data", nil
	})

	assert.False(t, reg.IsDryRun("get_usage"))
	assert.False(t, reg.HasPreview("get_usage"))

	result, err := reg.Execute(context.Background(),"get_usage", nil)

	require.NoError(t, err)
	assert.Equal(t, "usage data", result)
}

func TestToolRegistry_Preview_NotFound(t *testing.T) {
	reg := NewToolRegistry()

	_, err := reg.Preview(context.Background(),"nonexistent", nil)

	assert.Error(t, err)
}

func TestInputBool(t *testing.T) {
	assert.True(t, inputBool(map[string]any{"confirmed": true}, "confirmed"))
	assert.False(t, inputBool(map[string]any{"confirmed": false}, "confirmed"))
	assert.False(t, inputBool(map[string]any{}, "confirmed"))
	assert.False(t, inputBool(map[string]any{"confirmed": "yes"}, "confirmed"))
}
