package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadExternalConfig_Valid(t *testing.T) {
	content := `hooks:
  - name: test-webhook
    event: message_processed
    type: webhook
    url: https://example.com/hook
    method: POST
    timeout: 5s
    async: true
  - name: test-command
    event: cron_job_completed
    type: command
    command: "/bin/echo hello"
    timeout: 10s
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.yaml")
	os.WriteFile(path, []byte(content), 0644)

	defs, err := LoadExternalConfig(path)

	require.NoError(t, err)
	assert.Len(t, defs, 2)
	assert.Equal(t, "test-webhook", defs[0].Name)
	assert.Equal(t, "webhook", defs[0].Type)
	assert.Equal(t, "https://example.com/hook", defs[0].URL)
	assert.True(t, defs[0].Async)
	assert.Equal(t, "test-command", defs[1].Name)
	assert.Equal(t, "command", defs[1].Type)
}

func TestLoadExternalConfig_MissingFile(t *testing.T) {
	_, err := LoadExternalConfig("/nonexistent/hooks.yaml")
	assert.Error(t, err)
}

func TestLoadExternalConfig_MissingName(t *testing.T) {
	content := `hooks:
  - event: message_processed
    type: webhook
    url: https://example.com
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.yaml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := LoadExternalConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing name")
}

func TestLoadExternalConfig_MissingEvent(t *testing.T) {
	content := `hooks:
  - name: test
    type: webhook
    url: https://example.com
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.yaml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := LoadExternalConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing event")
}

func TestLoadExternalConfig_InvalidType(t *testing.T) {
	content := `hooks:
  - name: test
    event: message_processed
    type: ftp
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.yaml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := LoadExternalConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestLoadExternalConfig_WebhookMissingURL(t *testing.T) {
	content := `hooks:
  - name: test
    event: message_processed
    type: webhook
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.yaml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := LoadExternalConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing url")
}

func TestLoadExternalConfig_CommandMissingCommand(t *testing.T) {
	content := `hooks:
  - name: test
    event: message_processed
    type: command
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.yaml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := LoadExternalConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing command")
}

func TestLoadExternalConfig_EmptyHooks(t *testing.T) {
	content := `hooks: []
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.yaml")
	os.WriteFile(path, []byte(content), 0644)

	defs, err := LoadExternalConfig(path)
	require.NoError(t, err)
	assert.Empty(t, defs)
}

func TestLoadExternalConfig_WithConditions(t *testing.T) {
	content := `hooks:
  - name: expense-only
    event: message_processed
    type: webhook
    url: https://example.com
    conditions:
      intent: expense
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.yaml")
	os.WriteFile(path, []byte(content), 0644)

	defs, err := LoadExternalConfig(path)
	require.NoError(t, err)
	assert.Len(t, defs, 1)
	assert.Equal(t, "expense", defs[0].Conditions["intent"])
}
