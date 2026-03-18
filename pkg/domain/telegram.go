package domain

// Telegram webhook payload types.
// See: https://core.telegram.org/bots/api#update

// TelegramUpdate represents an incoming update from Telegram.
type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

// TelegramMessage represents a Telegram message.
type TelegramMessage struct {
	MessageID int              `json:"message_id"`
	From      *TelegramUser    `json:"from,omitempty"`
	Chat      TelegramChat     `json:"chat"`
	Text      string           `json:"text"`
	Voice     *TelegramVoice   `json:"voice,omitempty"`
	Audio     *TelegramAudio   `json:"audio,omitempty"`
	Photo     []TelegramPhoto  `json:"photo,omitempty"`
	Document  *TelegramDocument `json:"document,omitempty"`
	Caption   string           `json:"caption,omitempty"`
}

// TelegramPhoto represents a photo size variant.
type TelegramPhoto struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size"`
}

// TelegramDocument represents a file sent as document.
type TelegramDocument struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int    `json:"file_size"`
}

// TelegramVoice represents a Telegram voice message (ogg/opus).
type TelegramVoice struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type"`
	FileSize int    `json:"file_size"`
}

// TelegramAudio represents a Telegram audio file.
type TelegramAudio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type"`
	FileSize int    `json:"file_size"`
	Title    string `json:"title"`
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

// Sentinel errors for Telegram.
var (
	ErrTelegramRequest = New("telegram api request failed")
	ErrTelegramParse   = New("failed to parse telegram response")
)
