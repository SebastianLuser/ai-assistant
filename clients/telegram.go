package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	telegramBaseURL        = "https://api.telegram.org/bot"
	telegramDefaultTimeout = 15 * time.Second
	telegramMaxMessageLen  = 4096
)

// TelegramUpdate represents an incoming Telegram update (message).
type TelegramUpdate struct {
	UpdateID int             `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

// TelegramMessage represents a Telegram message.
type TelegramMessage struct {
	MessageID int           `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      TelegramChat  `json:"chat"`
	Text      string        `json:"text"`
}

// TelegramUser represents a Telegram user.
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// TelegramChat represents a Telegram chat.
type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// TelegramClient is the Telegram Bot API client.
type TelegramClient struct {
	token      string
	httpClient *http.Client
}

// NewTelegramClient creates a new Telegram Bot API client.
func NewTelegramClient(token string) *TelegramClient {
	return &TelegramClient{
		token:      token,
		httpClient: &http.Client{Timeout: telegramDefaultTimeout},
	}
}

// Name implements domain.Channel.
func (c *TelegramClient) Name() string { return "telegram" }

// SendMessage implements domain.Channel.
func (c *TelegramClient) SendMessage(to, text string) error {
	return c.SendTextMessage(to, text)
}

// AckMessage implements domain.Channel. Telegram has no read receipts,
// so we send a "typing" action as acknowledgment.
func (c *TelegramClient) AckMessage(messageID string) error {
	// messageID is not used — we don't have chatID here.
	// AckMessage is best-effort, so we just return nil.
	return nil
}

// SendTextMessage sends a text message to a Telegram chat.
// Automatically splits messages longer than 4096 characters.
func (c *TelegramClient) SendTextMessage(chatID, text string) error {
	chunks := splitMessage(text, telegramMaxMessageLen)
	for _, chunk := range chunks {
		body := map[string]any{
			"chat_id":    chatID,
			"text":       chunk,
			"parse_mode": "Markdown",
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("telegram: marshal body: %w", err)
		}

		url := telegramBaseURL + c.token + "/sendMessage"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("telegram: create request: %w", err)
		}

		req.Header.Set(headerContentType, contentTypeJSON)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("telegram: send request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("telegram: api error %d: %s", resp.StatusCode, string(respBody))
		}
	}
	return nil
}

// SendChatAction sends a "typing" indicator to a chat.
func (c *TelegramClient) SendChatAction(chatID, action string) error {
	body := map[string]string{
		"chat_id": chatID,
		"action":  action,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("telegram: marshal action: %w", err)
	}

	url := telegramBaseURL + c.token + "/sendChatAction"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("telegram: create request: %w", err)
	}
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send action: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// DownloadMedia downloads a file by file_id from Telegram.
// Implements domain.MediaDownloader.
func (c *TelegramClient) DownloadMedia(fileID string) ([]byte, string, error) {
	// Step 1: get file path
	url := telegramBaseURL + c.token + "/getFile?file_id=" + fileID
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("telegram: get file: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("telegram: parse file info: %w", err)
	}

	// Step 2: download
	dlURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.token, result.Result.FilePath)
	dlResp, err := c.httpClient.Get(dlURL)
	if err != nil {
		return nil, "", fmt.Errorf("telegram: download file: %w", err)
	}
	defer dlResp.Body.Close()

	data, err := io.ReadAll(dlResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("telegram: read file: %w", err)
	}

	// Infer mime type from file extension
	mime := "audio/ogg"
	if strings.HasSuffix(result.Result.FilePath, ".mp3") {
		mime = "audio/mpeg"
	} else if strings.HasSuffix(result.Result.FilePath, ".m4a") {
		mime = "audio/mp4"
	}

	return data, mime, nil
}

// SetWebhook registers a webhook URL with Telegram.
func (c *TelegramClient) SetWebhook(webhookURL, secretToken string) error {
	body := map[string]string{
		"url":          webhookURL,
		"secret_token": secretToken,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("telegram: marshal webhook: %w", err)
	}

	url := telegramBaseURL + c.token + "/setWebhook"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("telegram: create request: %w", err)
	}
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram: webhook error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
