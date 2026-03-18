package domain

// Meta WhatsApp Cloud API webhook payload types.
// See: https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/components

// WhatsAppWebhookPayload is the top-level envelope Meta sends to the webhook.
type WhatsAppWebhookPayload struct {
	Object string         `json:"object"`
	Entry  []WhatsAppEntry `json:"entry"`
}

// WhatsAppEntry represents a single entry in the webhook payload.
type WhatsAppEntry struct {
	ID      string           `json:"id"`
	Changes []WhatsAppChange `json:"changes"`
}

// WhatsAppChange represents a change notification.
type WhatsAppChange struct {
	Value WhatsAppValue `json:"value"`
	Field string        `json:"field"`
}

// WhatsAppValue contains the message data.
type WhatsAppValue struct {
	MessagingProduct string                    `json:"messaging_product"`
	Metadata         WhatsAppMetadata          `json:"metadata"`
	Contacts         []WhatsAppContact         `json:"contacts,omitempty"`
	Messages         []WhatsAppIncomingMessage  `json:"messages,omitempty"`
	Statuses         []WhatsAppStatus           `json:"statuses,omitempty"`
}

// WhatsAppMetadata contains phone number metadata.
type WhatsAppMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// WhatsAppContact represents a sender contact.
type WhatsAppContact struct {
	Profile WhatsAppProfile `json:"profile"`
	WaID    string          `json:"wa_id"`
}

// WhatsAppProfile contains the sender's profile info.
type WhatsAppProfile struct {
	Name string `json:"name"`
}

// WhatsAppIncomingMessage represents a single incoming message.
type WhatsAppIncomingMessage struct {
	From      string           `json:"from"`
	ID        string           `json:"id"`
	Timestamp string           `json:"timestamp"`
	Type      string           `json:"type"`
	Text      WhatsAppTextBody  `json:"text,omitempty"`
	Audio     *WhatsAppAudioBody `json:"audio,omitempty"`
}

// WhatsAppTextBody contains the text content of a message.
type WhatsAppTextBody struct {
	Body string `json:"body"`
}

// WhatsAppAudioBody contains audio message metadata.
type WhatsAppAudioBody struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
}

// WhatsAppStatus represents a message status update (sent, delivered, read).
type WhatsAppStatus struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// WhatsAppVerifyRequest holds Meta's webhook verification parameters.
type WhatsAppVerifyRequest struct {
	Mode      string
	Token     string
	Challenge string
}

// Validate checks that all verification fields are present.
func (r WhatsAppVerifyRequest) Validate() error {
	if r.Mode == "" || r.Token == "" || r.Challenge == "" {
		return Wrap(ErrValidation, "missing verification parameters")
	}
	if r.Mode != "subscribe" {
		return Wrap(ErrValidation, "invalid hub.mode: "+r.Mode)
	}
	return nil
}

// ExtractTextMessages returns all text messages from a webhook payload,
// filtering out status updates and non-text messages.
func ExtractTextMessages(payload WhatsAppWebhookPayload) []WhatsAppIncomingMessage {
	var msgs []WhatsAppIncomingMessage
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				if msg.Type == "text" && msg.Text.Body != "" {
					msgs = append(msgs, msg)
				}
			}
		}
	}
	return msgs
}

// ExtractAudioMessages returns all audio/voice messages from a webhook payload.
func ExtractAudioMessages(payload WhatsAppWebhookPayload) []WhatsAppIncomingMessage {
	var msgs []WhatsAppIncomingMessage
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				if (msg.Type == "audio" || msg.Type == "voice") && msg.Audio != nil {
					msgs = append(msgs, msg)
				}
			}
		}
	}
	return msgs
}
