package controller

import (
	"encoding/json"
	"net/http"
	"testing"

	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validEventBody     = `{"summary":"Meeting","start":"2026-03-11T10:00:00Z","end":"2026-03-11T11:00:00Z"}`
	noSummaryEventBody = `{"summary":"","start":"2026-03-11T10:00:00Z","end":"2026-03-11T11:00:00Z"}`
	badStartEventBody  = `{"summary":"x","start":"not-a-date","end":"2026-03-11T11:00:00Z"}`
	endBeforeStartBody = `{"summary":"x","start":"2026-03-11T12:00:00Z","end":"2026-03-11T10:00:00Z"}`
	invalidEventJSON   = `{broken`
)

func TestCalendarController_CreateEvent_InvalidJSON(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(invalidEventJSON)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestCalendarController_CreateEvent_EmptySummary(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(noSummaryEventBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: summary is required", errorFromBody(t, resp.Body))
}

func TestCalendarController_CreateEvent_InvalidStartFormat(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(badStartEventBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: invalid start time format, use RFC3339", errorFromBody(t, resp.Body))
}

func TestCalendarController_CreateEvent_EndBeforeStart(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(endBeforeStartBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: end must be after start", errorFromBody(t, resp.Body))
}

func errorFromBody(t *testing.T, body []byte) string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	v, _ := m["error"].(string)
	return v
}
