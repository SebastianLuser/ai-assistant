package controller

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"asistente/clients"
	"asistente/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validNoteWriteBody    = `{"path":"notes/test.md","content":"hello world"}`
	pathTraversalBody     = `{"path":"../../../etc/passwd","content":"x"}`
	absolutePathBody      = `{"path":"/etc/passwd","content":"x"}`
	emptyPathWriteBody    = `{"path":"","content":"hello"}`
	emptyContentWriteBody = `{"path":"test.md","content":""}`
	invalidObsidianJSON   = `{nope`
)

func TestObsidianController_WriteNote_InvalidJSON(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(invalidObsidianJSON)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestObsidianController_WriteNote_PathTraversal(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(pathTraversalBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: path must not contain '..'", errorFromBody(t, resp.Body))
}

func TestObsidianController_WriteNote_AbsolutePath(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(absolutePathBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: path must be relative", errorFromBody(t, resp.Body))
}

func TestObsidianController_WriteNote_EmptyPath(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(emptyPathWriteBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: path is required", errorFromBody(t, resp.Body))
}

func TestObsidianController_WriteNote_EmptyContent(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(emptyContentWriteBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: content is required", errorFromBody(t, resp.Body))
}

func TestObsidianController_ReadNote_MissingPath(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest()

	resp := ctrl.ReadNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestObsidianController_ReadNote_PathTraversal(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithQuery("path", "../secret.md")

	resp := ctrl.ReadNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: path must not contain '..'", errorFromBody(t, resp.Body))
}

func TestObsidianController_WriteNote_Success(t *testing.T) {
	tmpDir := t.TempDir()
	vault := clients.NewObsidianVault(tmpDir)
	ctrl := NewObsidianController(vault)
	req := test.NewMockRequest().WithBody(`{"path":"test.md","content":"hello world"}`)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
	content, err := os.ReadFile(filepath.Join(tmpDir, "test.md"))
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(content))
}

func TestObsidianController_ReadNote_Success(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte("hello"), 0644))
	vault := clients.NewObsidianVault(tmpDir)
	ctrl := NewObsidianController(vault)
	req := test.NewMockRequest().WithQuery("path", "test.md")

	resp := ctrl.ReadNote(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestObsidianController_ReadNote_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	vault := clients.NewObsidianVault(tmpDir)
	ctrl := NewObsidianController(vault)
	req := test.NewMockRequest().WithQuery("path", "nonexistent.md")

	resp := ctrl.ReadNote(req)

	assert.Equal(t, http.StatusNotFound, resp.Status)
}

func TestObsidianController_ListNotes_Success(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("b"), 0644))
	vault := clients.NewObsidianVault(tmpDir)
	ctrl := NewObsidianController(vault)
	req := test.NewMockRequest()

	resp := ctrl.ListNotes(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestObsidianController_ListNotes_WithDir(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "note.md"), []byte("a"), 0644))
	vault := clients.NewObsidianVault(tmpDir)
	ctrl := NewObsidianController(vault)
	req := test.NewMockRequest().WithQuery("dir", "sub")

	resp := ctrl.ListNotes(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestObsidianController_SearchNotes_MissingQuery(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest()

	resp := ctrl.SearchNotes(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestObsidianController_SearchNotes_EmptyQuery(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithQuery("q", "")

	resp := ctrl.SearchNotes(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestObsidianController_SearchNotes_Success(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte("hello world"), 0644))
	vault := clients.NewObsidianVault(tmpDir)
	ctrl := NewObsidianController(vault)
	req := test.NewMockRequest().WithQuery("q", "hello")

	resp := ctrl.SearchNotes(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestObsidianController_WriteNote_SubdirCreated(t *testing.T) {
	tmpDir := t.TempDir()
	vault := clients.NewObsidianVault(tmpDir)
	ctrl := NewObsidianController(vault)
	req := test.NewMockRequest().WithBody(validNoteWriteBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
	_, err := os.Stat(filepath.Join(tmpDir, "notes", "test.md"))
	assert.NoError(t, err)
}
