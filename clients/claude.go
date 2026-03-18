package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"asistente/pkg/domain"
)

const (
	claudeDefaultBaseURL   = "https://api.anthropic.com/v1/messages"
	claudeDefaultMaxTokens = 2048
	claudeDefaultTimeout   = 30 * time.Second
	anthropicVersion       = "2023-06-01"
	claudeHeaderAPIKey     = "x-api-key"
	claudeHeaderVersion    = "anthropic-version"
)

// Compile-time check: *ClaudeClient implements domain.AIProvider.
var _ domain.AIProvider = (*ClaudeClient)(nil)

type ClaudeClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []claudeContentBlock
}

type claudeContentBlock struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
	ToolUseID string         `json:"tool_use_id,omitempty"`
	Content   string         `json:"content,omitempty"`
	IsError   bool           `json:"is_error,omitempty"`
}

type claudeRequest struct {
	Model     string                `json:"model"`
	MaxTokens int                   `json:"max_tokens"`
	System    string                `json:"system,omitempty"`
	Messages  []claudeMessage       `json:"messages"`
	Tools     []domain.ToolDefinition `json:"tools,omitempty"`
}

type claudeResponse struct {
	Content    []claudeContentBlock `json:"content"`
	StopReason string               `json:"stop_reason"`
	Usage      domain.Usage         `json:"usage"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewClaudeClient(apiKey, model string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    claudeDefaultBaseURL,
		httpClient: &http.Client{Timeout: claudeDefaultTimeout},
	}
}

func NewClaudeClientWithBaseURL(apiKey, model, baseURL string) *ClaudeClient {
	c := NewClaudeClient(apiKey, model)
	c.baseURL = baseURL
	return c
}

func (c *ClaudeClient) Complete(system, userMessage string, opts ...domain.CompletionOption) (string, error) {
	return c.CompleteMessages(system, []domain.Message{
		{Role: domain.RoleUser, Content: userMessage},
	}, opts...)
}

// CompleteWithUsage is like Complete but also returns token usage.
func (c *ClaudeClient) CompleteWithUsage(system, userMessage string, opts ...domain.CompletionOption) (string, domain.Usage, error) {
	return c.completeMessagesWithUsage(system, []domain.Message{
		{Role: domain.RoleUser, Content: userMessage},
	}, opts...)
}

func (c *ClaudeClient) CompleteMessages(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, error) {
	text, usage, err := c.completeMessagesWithUsage(system, messages, opts...)
	if err == nil && (usage.InputTokens > 0 || usage.OutputTokens > 0) {
		log.Printf("claude: model=%s in=%d out=%d", c.model, usage.InputTokens, usage.OutputTokens)
	}
	return text, err
}

func (c *ClaudeClient) CompleteJSON(system, userMessage string, target any, opts ...domain.CompletionOption) error {
	text, err := c.Complete(system, userMessage, opts...)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(text), target); err != nil {
		return domain.Wrapf(domain.ErrClaudeJSON, err)
	}

	return nil
}

// CompleteWithTools sends a request with tool definitions and returns content blocks + stop reason.
func (c *ClaudeClient) CompleteWithTools(system string, messages []domain.Message, tools []domain.ToolDefinition, opts ...domain.CompletionOption) ([]domain.ContentBlock, string, error) {
	cfg := domain.ApplyOptions(4096, opts...)

	reqBody := claudeRequest{
		Model:     c.model,
		MaxTokens: cfg.MaxTokens,
		System:    system,
		Messages:  domainToClaudeMessages(messages),
		Tools:     tools,
	}

	result, err := c.doRequest(reqBody)
	if err != nil {
		return nil, "", err
	}

	blocks := make([]domain.ContentBlock, len(result.Content))
	for i, b := range result.Content {
		blocks[i] = domain.ContentBlock{
			Type:  b.Type,
			Text:  b.Text,
			ID:    b.ID,
			Name:  b.Name,
			Input: b.Input,
		}
	}

	log.Printf("claude: model=%s in=%d out=%d stop=%s", c.model, result.Usage.InputTokens, result.Usage.OutputTokens, result.StopReason)
	return blocks, result.StopReason, nil
}

var _ domain.ToolUseProvider = (*ClaudeClient)(nil)

func (c *ClaudeClient) completeMessagesWithUsage(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, domain.Usage, error) {
	cfg := domain.ApplyOptions(claudeDefaultMaxTokens, opts...)

	reqBody := claudeRequest{
		Model:     c.model,
		MaxTokens: cfg.MaxTokens,
		System:    system,
		Messages:  domainToClaudeMessages(messages),
	}

	result, err := c.doRequest(reqBody)
	if err != nil {
		return "", domain.Usage{}, err
	}

	// Extract first text block.
	for _, b := range result.Content {
		if b.Type == "text" {
			return b.Text, result.Usage, nil
		}
	}

	return "", result.Usage, domain.ErrClaudeEmpty
}

func (c *ClaudeClient) doRequest(reqBody claudeRequest) (claudeResponse, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return claudeResponse{}, domain.Wrapf(domain.ErrClaudeMarshal, err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return claudeResponse{}, domain.Wrapf(domain.ErrClaudeRequest, err)
	}

	req.Header.Set(headerContentType, contentTypeJSON)
	req.Header.Set(claudeHeaderAPIKey, c.apiKey)
	req.Header.Set(claudeHeaderVersion, anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return claudeResponse{}, domain.Wrapf(domain.ErrClaudeSend, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return claudeResponse{}, domain.Wrapf(domain.ErrClaudeRead, err)
	}

	var result claudeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return claudeResponse{}, domain.Wrapf(domain.ErrClaudeUnmarshal, err)
	}

	if result.Error != nil {
		return claudeResponse{}, domain.Wrap(domain.ErrClaudeAPI, result.Error.Type+": "+result.Error.Message)
	}

	if len(result.Content) == 0 {
		return claudeResponse{}, domain.ErrClaudeEmpty
	}

	return result, nil
}

// domainToClaudeMessages converts domain messages to Claude API format.
// Handles both simple text and structured content blocks (for tool use).
func domainToClaudeMessages(messages []domain.Message) []claudeMessage {
	apiMsgs := make([]claudeMessage, len(messages))
	for i, m := range messages {
		if len(m.ContentBlocks) > 0 {
			blocks := make([]claudeContentBlock, len(m.ContentBlocks))
			for j, b := range m.ContentBlocks {
				blocks[j] = claudeContentBlock{
					Type:      b.Type,
					Text:      b.Text,
					ID:        b.ID,
					Name:      b.Name,
					Input:     b.Input,
					ToolUseID: b.ID,
					Content:   b.Text,
				}
				// Fix: tool_result blocks use content+tool_use_id, not text
				if b.Type == "tool_result" {
					blocks[j].Text = ""
				}
			}
			apiMsgs[i] = claudeMessage{Role: m.Role, Content: blocks}
		} else {
			apiMsgs[i] = claudeMessage{Role: m.Role, Content: m.Content}
		}
	}
	return apiMsgs
}
