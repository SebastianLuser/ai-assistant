package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/web"
)

type WhatsAppController struct {
	router      *usecase.MessageRouter
	channel     domain.Channel
	verifyToken string
	appSecret   string
}

func NewWhatsAppController(router *usecase.MessageRouter, channel domain.Channel, verifyToken, appSecret string) *WhatsAppController {
	return &WhatsAppController{
		router:      router,
		channel:     channel,
		verifyToken: verifyToken,
		appSecret:   appSecret,
	}
}

// VerifyWebhook handles Meta's GET verification challenge.
func (c *WhatsAppController) VerifyWebhook(req web.Request) web.Response {
	mode, _ := req.Query("hub.mode")
	token, _ := req.Query("hub.verify_token")
	challenge, _ := req.Query("hub.challenge")

	verify := domain.WhatsAppVerifyRequest{
		Mode:      mode,
		Token:     token,
		Challenge: challenge,
	}

	if err := verify.Validate(); err != nil {
		return web.NewResponse(http.StatusForbidden, []byte("invalid verification request"))
	}

	if verify.Token != c.verifyToken {
		return web.NewResponse(http.StatusForbidden, []byte("invalid verify token"))
	}

	return web.NewResponse(http.StatusOK, []byte(verify.Challenge))
}

// HandleWebhook receives incoming WhatsApp messages from Meta.
// It validates the signature, returns 200 immediately, and processes messages async.
func (c *WhatsAppController) HandleWebhook(req web.Request) web.Response {
	body, err := io.ReadAll(req.Body())
	if err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, map[string]string{"error": "failed to read body"})
	}

	if c.appSecret != "" {
		signature := req.Raw().Header.Get("X-Hub-Signature-256")
		if !c.verifySignature(body, signature) {
			return web.NewJSONResponse(http.StatusForbidden, map[string]string{"error": "invalid signature"})
		}
	}

	var payload domain.WhatsAppWebhookPayload
	if err := web.DecodeJSON(strings.NewReader(string(body)), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	textMsgs := domain.ExtractTextMessages(payload)
	for _, msg := range textMsgs {
		go c.router.ProcessMessage(c.channel, msg.From, msg.ID, msg.Text.Body)
	}

	audioMsgs := domain.ExtractAudioMessages(payload)
	for _, msg := range audioMsgs {
		go c.router.ProcessAudioMessage(c.channel, msg.From, msg.ID, msg.Audio.ID)
	}

	imgMsgs := domain.ExtractImageMessages(payload)
	for _, msg := range imgMsgs {
		mediaID := ""
		caption := ""
		if msg.Image != nil {
			mediaID = msg.Image.ID
			caption = msg.Image.Caption
		} else if msg.Document != nil {
			mediaID = msg.Document.ID
			caption = msg.Document.Caption
		}
		go c.router.ProcessImageMessage(c.channel, msg.From, msg.ID, mediaID, caption)
	}

	total := len(textMsgs) + len(audioMsgs) + len(imgMsgs)
	if total > 0 {
		log.Printf("whatsapp: webhook received %d text + %d audio + %d image message(s)", len(textMsgs), len(audioMsgs), len(imgMsgs))
	}

	return web.NewJSONResponse(http.StatusOK, map[string]string{"status": "ok"})
}

func (c *WhatsAppController) verifySignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	expectedPrefix := "sha256="
	if !strings.HasPrefix(signature, expectedPrefix) {
		return false
	}

	sigHex := signature[len(expectedPrefix):]
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(c.appSecret))
	mac.Write(body)
	expected := mac.Sum(nil)

	return hmac.Equal(sigBytes, expected)
}
