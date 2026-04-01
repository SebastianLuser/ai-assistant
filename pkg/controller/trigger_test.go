package controller

import (
	"net/http"
	"testing"

	"jarvis/internal/hooks"
	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
)

func TestTriggerController_TriggerJob_MissingJobID(t *testing.T) {
	ctrl := NewTriggerController(nil)
	req := test.NewMockRequest()

	resp := ctrl.TriggerJob(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestTriggerController_TriggerJob_EmptyJobID(t *testing.T) {
	ctrl := NewTriggerController(nil)
	req := test.NewMockRequest().WithParam("job_id", "")

	resp := ctrl.TriggerJob(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestTriggerController_TriggerJob_Success(t *testing.T) {
	hooksReg := hooks.NewRegistry()
	jobs := []domain.Job{
		{
			ID: "test-job",
			RunFn: func() (string, error) {
				return "done", nil
			},
		},
	}
	scheduler := usecase.NewScheduler(jobs, hooksReg)
	ctrl := NewTriggerController(scheduler)
	req := test.NewMockRequest().WithParam("job_id", "test-job")

	resp := ctrl.TriggerJob(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestTriggerController_TriggerJob_NotFound(t *testing.T) {
	hooksReg := hooks.NewRegistry()
	scheduler := usecase.NewScheduler(nil, hooksReg)
	ctrl := NewTriggerController(scheduler)
	req := test.NewMockRequest().WithParam("job_id", "nonexistent")

	resp := ctrl.TriggerJob(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestTriggerController_ListJobs_Success(t *testing.T) {
	hooksReg := hooks.NewRegistry()
	jobs := []domain.Job{
		{ID: "job1"},
		{ID: "job2"},
	}
	scheduler := usecase.NewScheduler(jobs, hooksReg)
	ctrl := NewTriggerController(scheduler)
	req := test.NewMockRequest()

	resp := ctrl.ListJobs(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestTriggerController_ListJobs_Empty(t *testing.T) {
	hooksReg := hooks.NewRegistry()
	scheduler := usecase.NewScheduler(nil, hooksReg)
	ctrl := NewTriggerController(scheduler)
	req := test.NewMockRequest()

	resp := ctrl.ListJobs(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
