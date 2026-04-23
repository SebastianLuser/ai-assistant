package usecase

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

type IntegrationStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LastCheck string `json:"last_check,omitempty"`
	Error     string `json:"error,omitempty"`
}

type TopTool struct {
	Name       string  `json:"name"`
	UsageCount int64   `json:"usage_count"`
	ErrorRate  float64 `json:"error_rate"`
}

type CatalogSummary struct {
	TotalTools       int       `json:"total_tools"`
	TotalUsage       int64     `json:"total_usage"`
	TopTools         []TopTool `json:"top_tools"`
	OverallErrorRate float64   `json:"overall_error_rate"`
}

type HealthReport struct {
	Status       string              `json:"status"`
	Uptime       string              `json:"uptime"`
	Integrations []IntegrationStatus `json:"integrations"`
	Catalog      *CatalogSummary     `json:"catalog,omitempty"`
}

type IntegrationCheck struct {
	Name  string
	Check func() error
}

type HealthChecker struct {
	mu        sync.RWMutex
	checks    []IntegrationCheck
	catalog   service.CatalogService
	report    *HealthReport
	startedAt time.Time
	stopCh    chan struct{}
}

func NewHealthChecker(checks []IntegrationCheck, catalog service.CatalogService) *HealthChecker {
	return &HealthChecker{
		checks:    checks,
		catalog:   catalog,
		startedAt: time.Now(),
		stopCh:    make(chan struct{}),
	}
}

func (hc *HealthChecker) Start(interval time.Duration) {
	hc.runChecks()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				hc.runChecks()
			case <-hc.stopCh:
				return
			}
		}
	}()
}

func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

func (hc *HealthChecker) Report() HealthReport {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	if hc.report == nil {
		return HealthReport{Status: "unknown", Uptime: hc.uptime()}
	}
	report := *hc.report
	report.Uptime = hc.uptime()
	return report
}

func (hc *HealthChecker) runChecks() {
	now := time.Now().Format(time.RFC3339)
	var integrations []IntegrationStatus
	overall := "healthy"

	for _, c := range hc.checks {
		status := IntegrationStatus{Name: c.Name, LastCheck: now}
		if err := c.Check(); err != nil {
			status.Status = "down"
			status.Error = err.Error()
			overall = "degraded"
		} else {
			status.Status = "ok"
		}
		integrations = append(integrations, status)
	}

	var catalogSummary *CatalogSummary
	if hc.catalog != nil {
		catalogSummary = hc.buildCatalogSummary()
	}

	hc.mu.Lock()
	hc.report = &HealthReport{
		Status:       overall,
		Integrations: integrations,
		Catalog:      catalogSummary,
	}
	hc.mu.Unlock()
}

func (hc *HealthChecker) buildCatalogSummary() *CatalogSummary {
	entries, err := hc.catalog.GetAll()
	if err != nil {
		return nil
	}

	summary := &CatalogSummary{TotalTools: len(entries)}
	var totalErrors int64
	for _, e := range entries {
		summary.TotalUsage += e.UsageCount
		totalErrors += e.ErrorCount
	}

	if summary.TotalUsage > 0 {
		summary.OverallErrorRate = float64(totalErrors) / float64(summary.TotalUsage)
	}

	summary.TopTools = topToolsFromEntries(entries, 5)
	return summary
}

func topToolsFromEntries(entries []domain.CatalogEntry, limit int) []TopTool {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].UsageCount > entries[j].UsageCount
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}
	var top []TopTool
	for _, e := range entries {
		var errRate float64
		if e.UsageCount > 0 {
			errRate = float64(e.ErrorCount) / float64(e.UsageCount)
		}
		top = append(top, TopTool{
			Name:       e.Name,
			UsageCount: e.UsageCount,
			ErrorRate:  errRate,
		})
	}
	return top
}

func (hc *HealthChecker) uptime() string {
	d := time.Since(hc.startedAt)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
