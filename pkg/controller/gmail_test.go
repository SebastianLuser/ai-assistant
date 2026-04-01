package controller

import (
	"net/http"
	"testing"

	"jarvis/test"

	"github.com/stretchr/testify/assert"
)

func TestGmailController_GetMessage_MissingID(t *testing.T) {
	ctrl := NewGmailController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetMessage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGmailController_GetMessage_EmptyID(t *testing.T) {
	ctrl := NewGmailController(nil)
	req := test.NewMockRequest().WithParam("id", "")

	resp := ctrl.GetMessage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}
