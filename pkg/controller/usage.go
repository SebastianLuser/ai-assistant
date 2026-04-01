package controller

import (
	"net/http"

	"jarvis/pkg/usecase"
	"jarvis/web"
)

// UsageController exposes token usage and cost tracking.
type UsageController struct {
	tracker *usecase.UsageTracker
}

// NewUsageController creates a new UsageController.
func NewUsageController(tracker *usecase.UsageTracker) *UsageController {
	return &UsageController{tracker: tracker}
}

// GetUsage returns usage stats for all sessions.
func (c *UsageController) GetUsage(req web.Request) web.Response {
	sessions := c.tracker.GetAll()
	totalCost := c.tracker.TotalCostUSD()

	return web.NewJSONResponse(http.StatusOK, map[string]any{
		"success":        true,
		"total_cost_usd": totalCost,
		"sessions":       sessions,
	})
}
