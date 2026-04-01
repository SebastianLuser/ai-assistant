package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"

	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
)

func newTestWAController(verifyToken, appSecret string) *WhatsAppController {
	return NewWhatsAppController(nil, nil, verifyToken, appSecret)
}

func TestWhatsAppController_VerifyWebhook_ValidToken(t *testing.T) {
	ctrl := newTestWAController("my-token", "")
	req := test.NewMockRequest().
		WithQuery("hub.mode", "subscribe").
		WithQuery("hub.verify_token", "my-token").
		WithQuery("hub.challenge", "challenge-123")

	resp := ctrl.VerifyWebhook(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	assert.Equal(t, []byte("challenge-123"), resp.Body)
}

func TestWhatsAppController_VerifyWebhook_InvalidToken(t *testing.T) {
	ctrl := newTestWAController("my-token", "")
	req := test.NewMockRequest().
		WithQuery("hub.mode", "subscribe").
		WithQuery("hub.verify_token", "wrong-token").
		WithQuery("hub.challenge", "challenge-123")

	resp := ctrl.VerifyWebhook(req)

	assert.Equal(t, http.StatusForbidden, resp.Status)
}

func TestWhatsAppController_VerifyWebhook_MissingParams(t *testing.T) {
	ctrl := newTestWAController("my-token", "")
	req := test.NewMockRequest().
		WithQuery("hub.mode", "subscribe")

	resp := ctrl.VerifyWebhook(req)

	assert.Equal(t, http.StatusForbidden, resp.Status)
}

func TestWhatsAppController_HandleWebhook_InvalidSignature(t *testing.T) {
	ctrl := newTestWAController("", "app-secret")
	payload := domain.WhatsAppWebhookPayload{Object: "whatsapp_business_account"}
	body, _ := json.Marshal(payload)

	req := test.NewMockRequest().
		WithBody(string(body)).
		WithHeader("X-Hub-Signature-256", "sha256=invalid")

	resp := ctrl.HandleWebhook(req)

	assert.Equal(t, http.StatusForbidden, resp.Status)
}

func TestWhatsAppController_HandleWebhook_ValidSignature(t *testing.T) {
	secret := "test-secret"
	payload := domain.WhatsAppWebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []domain.WhatsAppEntry{{
			Changes: []domain.WhatsAppChange{{
				Value: domain.WhatsAppValue{
					Statuses: []domain.WhatsAppStatus{{ID: "1", Status: "delivered"}},
				},
			}},
		}},
	}
	body, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctrl := newTestWAController("", secret)
	req := test.NewMockRequest().
		WithBody(string(body)).
		WithHeader("X-Hub-Signature-256", sig)

	resp := ctrl.HandleWebhook(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestWhatsAppController_HandleWebhook_NoSecretConfigured(t *testing.T) {
	ctrl := newTestWAController("", "")
	payload := domain.WhatsAppWebhookPayload{Object: "whatsapp_business_account"}
	body, _ := json.Marshal(payload)

	req := test.NewMockRequest().WithBody(string(body))

	resp := ctrl.HandleWebhook(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
