package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WhatsAppClient struct {
	phoneNumberID string
	accessToken   string
	httpClient    *http.Client
}

func NewWhatsAppClient(phoneNumberID, accessToken string) *WhatsAppClient {
	return &WhatsAppClient{
		phoneNumberID: phoneNumberID,
		accessToken:   accessToken,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
}

const whatsAppMaxMessageLen = 4096

func (c *WhatsAppClient) SendTextMessage(to, text string) error {
	chunks := splitMessage(text, whatsAppMaxMessageLen)
	for _, chunk := range chunks {
		if err := c.sendMessage(to, map[string]any{
			"messaging_product": "whatsapp",
			"to":                to,
			"type":              "text",
			"text":              map[string]string{"body": chunk},
		}); err != nil {
			return err
		}
	}
	return nil
}

// MarkAsRead sends a read receipt so the user sees blue checkmarks.
func (c *WhatsAppClient) MarkAsRead(messageID string) error {
	return c.sendMessage("", map[string]any{
		"messaging_product": "whatsapp",
		"status":            "read",
		"message_id":        messageID,
	})
}

func (c *WhatsAppClient) sendMessage(to string, body map[string]any) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", c.phoneNumberID)

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("whatsapp: marshal body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("whatsapp: create request: %w", err)
	}

	req.Header.Set(headerContentType, contentTypeJSON)
	req.Header.Set(headerAuthorization, "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// splitMessage breaks text into chunks that fit within WhatsApp's character limit.
// It tries to split at newlines, then spaces, falling back to hard cuts.
func splitMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for len(text) > 0 {
		if len(text) <= maxLen {
			chunks = append(chunks, text)
			break
		}

		chunk := text[:maxLen]
		cutAt := maxLen

		if idx := lastIndexByte(chunk, '\n'); idx > maxLen/2 {
			cutAt = idx + 1
		} else if idx := lastIndexByte(chunk, ' '); idx > maxLen/2 {
			cutAt = idx + 1
		}

		chunks = append(chunks, text[:cutAt])
		text = text[cutAt:]
	}
	return chunks
}

// DownloadMedia downloads a media file from WhatsApp Cloud API.
// First gets the download URL, then downloads the binary data.
func (c *WhatsAppClient) DownloadMedia(mediaID string) ([]byte, string, error) {
	// Step 1: Get download URL
	urlEndpoint := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", mediaID)
	req, err := http.NewRequest(http.MethodGet, urlEndpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("whatsapp: create media request: %w", err)
	}
	req.Header.Set(headerAuthorization, "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("whatsapp: get media url: %w", err)
	}
	defer resp.Body.Close()

	var mediaInfo struct {
		URL      string `json:"url"`
		MimeType string `json:"mime_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mediaInfo); err != nil {
		return nil, "", fmt.Errorf("whatsapp: parse media info: %w", err)
	}

	// Step 2: Download binary data
	dlReq, err := http.NewRequest(http.MethodGet, mediaInfo.URL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("whatsapp: create download request: %w", err)
	}
	dlReq.Header.Set(headerAuthorization, "Bearer "+c.accessToken)

	dlResp, err := c.httpClient.Do(dlReq)
	if err != nil {
		return nil, "", fmt.Errorf("whatsapp: download media: %w", err)
	}
	defer dlResp.Body.Close()

	data, err := io.ReadAll(dlResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("whatsapp: read media: %w", err)
	}

	return data, mediaInfo.MimeType, nil
}

// Name implements domain.Channel.
func (c *WhatsAppClient) Name() string { return "whatsapp" }

// SendMessage implements domain.Channel.
func (c *WhatsAppClient) SendMessage(to, text string) error {
	return c.SendTextMessage(to, text)
}

// AckMessage implements domain.Channel by sending a read receipt.
func (c *WhatsAppClient) AckMessage(messageID string) error {
	return c.MarkAsRead(messageID)
}

func lastIndexByte(s string, b byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == b {
			return i
		}
	}
	return -1
}
