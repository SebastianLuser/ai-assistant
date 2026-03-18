package controller

import (
	"fmt"
	"log"
	"net/http"

	"asistente/pkg/domain"
	"asistente/pkg/usecase"
	"asistente/web"
)

// TelegramController handles Telegram Bot webhook endpoints.
type TelegramController struct {
	router      *usecase.MessageRouter
	channel     domain.Channel
	secretToken string
}

// NewTelegramController creates a new TelegramController.
func NewTelegramController(router *usecase.MessageRouter, channel domain.Channel, secretToken string) *TelegramController {
	return &TelegramController{
		router:      router,
		channel:     channel,
		secretToken: secretToken,
	}
}

// HandleWebhook receives incoming Telegram updates.
// Telegram sends X-Telegram-Bot-Api-Secret-Token header for verification.
func (c *TelegramController) HandleWebhook(req web.Request) web.Response {
	if c.secretToken != "" {
		token := req.Raw().Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if token != c.secretToken {
			return web.NewJSONResponse(http.StatusForbidden, map[string]string{"error": "invalid secret token"})
		}
	}

	var update domain.TelegramUpdate
	if err := web.DecodeJSON(req.Body(), &update); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	if update.Message != nil {
		msg := update.Message
		from := fmt.Sprintf("%d", msg.Chat.ID)
		messageID := fmt.Sprintf("%d", msg.MessageID)

		switch {
		case msg.Text != "":
			log.Printf("telegram: text from %s", from)
			go c.router.ProcessMessage(c.channel, from, messageID, msg.Text)
		case msg.Voice != nil:
			log.Printf("telegram: voice from %s", from)
			go c.router.ProcessAudioMessage(c.channel, from, messageID, msg.Voice.FileID)
		case msg.Audio != nil:
			log.Printf("telegram: audio from %s", from)
			go c.router.ProcessAudioMessage(c.channel, from, messageID, msg.Audio.FileID)
		}
	}

	return web.NewJSONResponse(http.StatusOK, map[string]string{"status": "ok"})
}
