package domain

// ToolDefinition defines a tool that Claude can call.
// Follows the Anthropic tool use API format.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// ToolCall represents Claude's decision to invoke a tool.
type ToolCall struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// ToolResult is the response sent back to Claude after executing a tool.
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// ContentBlock represents a polymorphic content block in Claude's response.
// A response can contain both text and tool_use blocks.
type ContentBlock struct {
	Type  string         `json:"type"`
	Text  string         `json:"text,omitempty"`
	ID    string         `json:"id,omitempty"`
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`
}

// StopReason constants from Claude's API.
const (
	StopReasonEndTurn = "end_turn"
	StopReasonToolUse = "tool_use"
	StopReasonMaxToks = "max_tokens"
)

// Sentinel errors for tool use and agent.
var (
	ErrToolExecution = New("tool execution failed")
	ErrToolNotFound  = New("tool not found")
	ErrAgentMaxTurns = New("agent exceeded maximum turns")
	ErrTranscription = New("audio transcription failed")
)
