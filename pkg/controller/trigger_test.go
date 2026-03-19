package controller

import (
	"net/http"
	"testing"

	"asistente/test"

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
