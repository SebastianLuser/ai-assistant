package domain

// Message represents a chat message for any AI provider.
// For simple text, use Content. For tool use, use ContentBlocks.
type Message struct {
	Role          string         `json:"role"`
	Content       string         `json:"content,omitempty"`
	ContentBlocks []ContentBlock `json:"content_blocks,omitempty"`
}

// Usage tracks token consumption from an AI API call.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// CompletionConfig holds options for a single AI completion call.
type CompletionConfig struct {
	MaxTokens int
}

// CompletionOption modifies a CompletionConfig.
type CompletionOption func(*CompletionConfig)

// WithMaxTokens sets the max output tokens for a completion call.
func WithMaxTokens(n int) CompletionOption {
	return func(c *CompletionConfig) {
		c.MaxTokens = n
	}
}

// ApplyOptions returns a CompletionConfig with all options applied.
func ApplyOptions(defaultMaxTokens int, opts ...CompletionOption) CompletionConfig {
	cfg := CompletionConfig{MaxTokens: defaultMaxTokens}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// AIProvider is the interface for any LLM provider (Claude, OpenAI, Gemini, Ollama, etc).
type AIProvider interface {
	Complete(system, userMessage string, opts ...CompletionOption) (string, error)
	CompleteMessages(system string, messages []Message, opts ...CompletionOption) (string, error)
	CompleteJSON(system, userMessage string, target any, opts ...CompletionOption) error
}

// ToolUseProvider extends AIProvider with tool/function calling support.
type ToolUseProvider interface {
	CompleteWithTools(system string, messages []Message, tools []ToolDefinition, opts ...CompletionOption) ([]ContentBlock, string, error)
}

// Transcriber converts audio to text (e.g. Whisper).
type Transcriber interface {
	Transcribe(audioData []byte, mimeType string) (string, error)
}

// MediaDownloader downloads media files from a messaging platform.
type MediaDownloader interface {
	DownloadMedia(mediaID string) ([]byte, string, error) // returns data, mimeType, error
}
