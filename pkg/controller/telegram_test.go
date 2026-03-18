package controller

import (
	"encoding/json"
	"net/http"
	"testing"

	"asistente/pkg/domain"
	"asistente/test"

	"github.com/stretchr/testify/assert"
)

func TestTelegramController_HandleWebhook_InvalidSecretToken(t *testing.T) {
	ctrl := NewTelegramController(nil, nil, "my-secret", "")
	req := test.NewMockRequest().
		WithBody(`{}`).
		WithHeader("X-Telegram-Bot-Api-Secret-Token", "wrong")

	resp := ctrl.HandleWebhook(req)

	assert.Equal(t, http.StatusForbidden, resp.Status)
}

func TestTelegramController_HandleWebhook_ValidSecretToken(t *testing.T) {
	ctrl := NewTelegramController(nil, nil, "my-secret", "")
	update := domain.TelegramUpdate{UpdateID: 1}
	body, _ := json.Marshal(update)

	req := test.NewMockRequest().
		WithBody(string(body)).
		WithHeader("X-Telegram-Bot-Api-Secret-Token", "my-secret")

	resp := ctrl.HandleWebhook(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestTelegramController_HandleWebhook_NoSecretConfigured(t *testing.T) {
	ctrl := NewTelegramController(nil, nil, "", "")
	update := domain.TelegramUpdate{UpdateID: 1}
	body, _ := json.Marshal(update)

	req := test.NewMockRequest().WithBody(string(body))

	resp := ctrl.HandleWebhook(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestTelegramController_HandleWebhook_InvalidPayload(t *testing.T) {
	ctrl := NewTelegramController(nil, nil, "", "")
	req := test.NewMockRequest().WithBody(`{invalid`)

	resp := ctrl.HandleWebhook(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}
