package domain

import "time"

const (
	CatalogTypeTool  = "tool"
	CatalogTypeSkill = "skill"
	CatalogTypeAgent = "agent"
)

type CatalogEntry struct {
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	UsageCount   int64      `json:"usage_count"`
	LastUsed     *time.Time `json:"last_used,omitempty"`
	SuccessCount int64      `json:"success_count"`
	ErrorCount   int64      `json:"error_count"`
	Tags         []string   `json:"tags,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (e CatalogEntry) SuccessRate() float64 {
	if e.UsageCount == 0 {
		return 0
	}
	return float64(e.SuccessCount) / float64(e.UsageCount)
}
