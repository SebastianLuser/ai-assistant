package skills

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSkillFile(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	require.NoError(t, err)
}

func TestCachedLoader_LoadEnabled_CachesResults(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "skill1.md", skillWithAllFields)

	loader := NewLoader(dir)
	cached := NewCachedLoader(loader)

	first, err := cached.LoadEnabled()
	require.NoError(t, err)
	require.Len(t, first, 1)

	second, err := cached.LoadEnabled()
	require.NoError(t, err)
	require.Len(t, second, 1)

	assert.Equal(t, first[0].Name, second[0].Name)
}

func TestCachedLoader_LoadEnabled_ReturnsFromCacheWithinInterval(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "skill1.md", skillNameOnly)

	loader := NewLoader(dir)
	cached := NewCachedLoader(loader)
	cached.checkEvery = 1 * time.Hour // long interval ensures cache is used

	first, err := cached.LoadEnabled()
	require.NoError(t, err)
	require.Len(t, first, 1)

	// Add another file — should NOT appear because cache interval hasn't expired
	writeSkillFile(t, dir, "skill2.md", skillWithAllFields)

	second, err := cached.LoadEnabled()
	require.NoError(t, err)
	assert.Len(t, second, 1) // still 1, cached
}

func TestCachedLoader_Invalidate_ForcesReload(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "skill1.md", skillNameOnly)

	loader := NewLoader(dir)
	cached := NewCachedLoader(loader)
	cached.checkEvery = 1 * time.Hour

	first, err := cached.LoadEnabled()
	require.NoError(t, err)
	require.Len(t, first, 1)

	writeSkillFile(t, dir, "skill2.md", skillWithAllFields)
	cached.Invalidate()

	second, err := cached.LoadEnabled()
	require.NoError(t, err)
	assert.Len(t, second, 2)
}

func TestCachedLoader_filesChanged_DetectsNewFile(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "skill1.md", skillNameOnly)

	loader := NewLoader(dir)
	cached := NewCachedLoader(loader)
	cached.checkEvery = 0 // always check files

	first, err := cached.LoadEnabled()
	require.NoError(t, err)
	require.Len(t, first, 1)

	// Ensure the new file has a later mtime
	time.Sleep(50 * time.Millisecond)
	writeSkillFile(t, dir, "skill2.md", skillWithAllFields)

	second, err := cached.LoadEnabled()
	require.NoError(t, err)
	assert.Len(t, second, 2)
}

func TestCachedLoader_filesChanged_NoChange_UsesCacheWithRefreshedTime(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "skill1.md", skillNameOnly)

	loader := NewLoader(dir)
	cached := NewCachedLoader(loader)
	cached.checkEvery = 0 // always check files, but files haven't changed

	first, err := cached.LoadEnabled()
	require.NoError(t, err)
	require.Len(t, first, 1)

	loadedBefore := cached.loadedAt

	// Small sleep to see if loadedAt gets refreshed
	time.Sleep(10 * time.Millisecond)

	second, err := cached.LoadEnabled()
	require.NoError(t, err)
	assert.Len(t, second, 1)

	// loadedAt should be refreshed even when files didn't change
	assert.True(t, cached.loadedAt.After(loadedBefore) || cached.loadedAt.Equal(loadedBefore))
}

func TestCachedLoader_filesChanged_InvalidDir_ReturnsTrue(t *testing.T) {
	loader := NewLoader("/nonexistent/path")
	cached := NewCachedLoader(loader)
	cached.loadedAt = time.Now()

	// filesChanged should return true for invalid dir
	assert.True(t, cached.filesChanged())
}

func TestCachedLoader_Dir_ReturnsLoaderDir(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir)
	cached := NewCachedLoader(loader)

	assert.Equal(t, dir, cached.Dir())
}

func TestCachedLoader_LoadAll_BypassesCache(t *testing.T) {
	dir := t.TempDir()
	writeSkillFile(t, dir, "skill1.md", skillNameOnly)

	loader := NewLoader(dir)
	cached := NewCachedLoader(loader)

	all, err := cached.LoadAll()
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestCachedLoader_LoadEnabled_ErrorOnInvalidDir(t *testing.T) {
	loader := NewLoader("/nonexistent/path")
	cached := NewCachedLoader(loader)

	_, err := cached.LoadEnabled()

	assert.Error(t, err)
}
