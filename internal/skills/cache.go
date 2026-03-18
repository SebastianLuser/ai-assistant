package skills

import (
	"os"
	"sync"
	"time"
)

type CachedLoader struct {
	loader    *Loader
	mu        sync.RWMutex
	skills    []Skill
	loadedAt  time.Time
	checkEvery time.Duration
}

const defaultCacheCheckInterval = 5 * time.Minute

func NewCachedLoader(loader *Loader) *CachedLoader {
	return &CachedLoader{
		loader:     loader,
		checkEvery: defaultCacheCheckInterval,
	}
}

// LoadEnabled returns cached enabled skills, reloading only if files changed.
func (c *CachedLoader) LoadEnabled() ([]Skill, error) {
	c.mu.RLock()
	if c.skills != nil && time.Since(c.loadedAt) < c.checkEvery {
		defer c.mu.RUnlock()
		return c.skills, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.skills != nil && time.Since(c.loadedAt) < c.checkEvery {
		return c.skills, nil
	}

	if c.skills != nil && !c.filesChanged() {
		c.loadedAt = time.Now()
		return c.skills, nil
	}

	loaded, err := c.loader.LoadEnabled()
	if err != nil {
		return nil, err
	}

	c.skills = loaded
	c.loadedAt = time.Now()
	return c.skills, nil
}

func (c *CachedLoader) filesChanged() bool {
	entries, err := os.ReadDir(c.loader.dir)
	if err != nil {
		return true
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return true
		}

		if info.ModTime().After(c.loadedAt) {
			return true
		}
	}

	return false
}

// Invalidate forces next LoadEnabled to reload from disk.
func (c *CachedLoader) Invalidate() {
	c.mu.Lock()
	c.skills = nil
	c.mu.Unlock()
}

// Dir returns the underlying loader's directory (for FilterByTags compatibility).
func (c *CachedLoader) Dir() string {
	return c.loader.dir
}

// LoadAll delegates to the underlying loader (bypasses cache).
func (c *CachedLoader) LoadAll() ([]Skill, error) {
	return c.loader.LoadAll()
}

// Save writes a skill to disk and invalidates the cache.
func (c *CachedLoader) Save(skill Skill) error {
	if err := c.loader.Save(skill); err != nil {
		return err
	}
	c.Invalidate()
	return nil
}

// SkillProvider is the interface used by controllers to load skills.
type SkillProvider interface {
	LoadEnabled() ([]Skill, error)
}

// SkillWriter extends SkillProvider with the ability to save new skills.
type SkillWriter interface {
	SkillProvider
	Save(skill Skill) error
}

var (
	_ SkillProvider = (*Loader)(nil)
	_ SkillProvider = (*CachedLoader)(nil)
)

