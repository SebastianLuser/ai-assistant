package domain

// AgentDefinition defines a specialized sub-agent with its own persona, tools, and routing rules.
// Inspired by OpenClaw's agent structure (SOUL.md + tools + bindings).
type AgentDefinition struct {
	// ID is the unique identifier for this agent (e.g. "finance", "dev", "study").
	ID string `json:"id"`
	// Name is the display name.
	Name string `json:"name"`
	// Description explains what this agent does (used by the orchestrator to decide routing).
	Description string `json:"description"`
	// SystemPrompt is the SOUL — persona, tone, boundaries, expertise.
	SystemPrompt string `json:"system_prompt"`
	// AllowedTools limits which tools this agent can use. Empty = all tools.
	AllowedTools []string `json:"allowed_tools,omitempty"`
	// DeniedTools explicitly blocks tools. Takes precedence over AllowedTools.
	DeniedTools []string `json:"denied_tools,omitempty"`
}

// SubAgentTask represents a task delegated to a sub-agent.
type SubAgentTask struct {
	AgentID string `json:"agent_id"`
	Task    string `json:"task"`
	RunID   string `json:"run_id"`
}

// SubAgentResult is the announcement from a completed sub-agent.
type SubAgentResult struct {
	AgentID string `json:"agent_id"`
	RunID   string `json:"run_id"`
	Task    string `json:"task"`
	Result  string `json:"result"`
	Status  string `json:"status"` // "success", "error", "timeout"
}
