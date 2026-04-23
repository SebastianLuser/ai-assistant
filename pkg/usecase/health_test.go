package usecase

import (
	"errors"
	"testing"
	"time"

	"jarvis/pkg/domain"

	"github.com/stretchr/testify/assert"
)

type mockCatalogForHealth struct {
	entries []domain.CatalogEntry
}

func (m *mockCatalogForHealth) RecordUsage(string, string, bool) error { return nil }
func (m *mockCatalogForHealth) GetAll() ([]domain.CatalogEntry, error) {
	return m.entries, nil
}
func (m *mockCatalogForHealth) GetByName(string, string) (*domain.CatalogEntry, error) {
	return nil, nil
}

func TestHealthChecker_AllOK(t *testing.T) {
	checks := []IntegrationCheck{
		{Name: "calendar", Check: func() error { return nil }},
		{Name: "github", Check: func() error { return nil }},
	}

	hc := NewHealthChecker(checks, nil)
	hc.runChecks()

	report := hc.Report()

	assert.Equal(t, "healthy", report.Status)
	assert.Len(t, report.Integrations, 2)
	assert.Equal(t, "ok", report.Integrations[0].Status)
	assert.Equal(t, "ok", report.Integrations[1].Status)
}

func TestHealthChecker_Degraded(t *testing.T) {
	checks := []IntegrationCheck{
		{Name: "calendar", Check: func() error { return nil }},
		{Name: "github", Check: func() error { return errors.New("token expired") }},
	}

	hc := NewHealthChecker(checks, nil)
	hc.runChecks()

	report := hc.Report()

	assert.Equal(t, "degraded", report.Status)
	assert.Equal(t, "ok", report.Integrations[0].Status)
	assert.Equal(t, "down", report.Integrations[1].Status)
	assert.Equal(t, "token expired", report.Integrations[1].Error)
}

func TestHealthChecker_CatalogSummary(t *testing.T) {
	catalog := &mockCatalogForHealth{
		entries: []domain.CatalogEntry{
			{Name: "save_expense", UsageCount: 100, SuccessCount: 90, ErrorCount: 10},
			{Name: "create_event", UsageCount: 50, SuccessCount: 48, ErrorCount: 2},
			{Name: "save_note", UsageCount: 30, SuccessCount: 30, ErrorCount: 0},
		},
	}

	hc := NewHealthChecker(nil, catalog)
	hc.runChecks()

	report := hc.Report()

	assert.NotNil(t, report.Catalog)
	assert.Equal(t, 3, report.Catalog.TotalTools)
	assert.Equal(t, int64(180), report.Catalog.TotalUsage)
	assert.InDelta(t, 12.0/180.0, report.Catalog.OverallErrorRate, 0.001)
	assert.Len(t, report.Catalog.TopTools, 3)
	assert.Equal(t, "save_expense", report.Catalog.TopTools[0].Name)
}

func TestHealthChecker_Uptime(t *testing.T) {
	hc := NewHealthChecker(nil, nil)
	hc.startedAt = time.Now().Add(-2 * time.Hour)

	report := hc.Report()

	assert.Contains(t, report.Uptime, "h")
}

func TestTopToolsFromEntries_LimitTo5(t *testing.T) {
	var entries []domain.CatalogEntry
	for i := 0; i < 10; i++ {
		entries = append(entries, domain.CatalogEntry{
			Name:       "tool_" + string(rune('a'+i)),
			UsageCount: int64(10 - i),
		})
	}

	top := topToolsFromEntries(entries, 5)

	assert.Len(t, top, 5)
	assert.Equal(t, int64(10), top[0].UsageCount)
}
