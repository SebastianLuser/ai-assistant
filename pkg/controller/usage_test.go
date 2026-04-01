package controller

import (
	"net/http"
	"testing"

	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
)

func TestUsageController_New(t *testing.T) {
	ctrl := NewUsageController(nil)

	assert.NotNil(t, ctrl)
}

func TestUsageController_GetUsage_Success(t *testing.T) {
	tracker := usecase.NewUsageTracker()
	ctrl := NewUsageController(tracker)
	req := test.NewMockRequest()

	resp := ctrl.GetUsage(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
