package controller

import (
	"net/http"

	"asistente/pkg/usecase"
	"asistente/web"
)

// TriggerController handles webhook trigger endpoints.
type TriggerController struct {
	scheduler *usecase.Scheduler
}

// NewTriggerController creates a new TriggerController.
func NewTriggerController(scheduler *usecase.Scheduler) *TriggerController {
	return &TriggerController{scheduler: scheduler}
}

// TriggerJob manually executes a cron job by ID.
func (c *TriggerController) TriggerJob(req web.Request) web.Response {
	jobID, ok := req.Param("job_id")
	if !ok || jobID == "" {
		return web.NewJSONResponse(http.StatusBadRequest, map[string]any{
			"success": false, "error": "job_id is required",
		})
	}

	result, err := c.scheduler.TriggerJob(jobID)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, map[string]any{
			"success": false, "error": err.Error(),
		})
	}

	return web.NewJSONResponse(http.StatusOK, map[string]any{
		"success": true, "job_id": jobID, "result": result,
	})
}

// ListJobs returns all registered job IDs.
func (c *TriggerController) ListJobs(req web.Request) web.Response {
	jobs := c.scheduler.ListJobs()

	return web.NewJSONResponse(http.StatusOK, map[string]any{
		"success": true, "jobs": jobs,
	})
}
