package controller

import (
	"net/http"

	"jarvis/pkg/usecase"
	"jarvis/web"
)

type HealthController struct {
	checker *usecase.HealthChecker
}

func NewHealthController(checker *usecase.HealthChecker) *HealthController {
	return &HealthController{checker: checker}
}

func (c *HealthController) DetailedHealth(req web.Request) web.Response {
	report := c.checker.Report()
	return web.NewJSONResponse(http.StatusOK, map[string]any{
		"success": true,
		"health":  report,
	})
}
