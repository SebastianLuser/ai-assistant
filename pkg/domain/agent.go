package domain

// AgentDefinition defines a specialized sub-agent with its own persona, tools, and routing rules.
// Inspired by OpenClaw's agent structure (SOUL.md + tools + bindings).
type AgentDefinition struct {
	ID             string   `json:"id" yaml:"id"`
	Name           string   `json:"name" yaml:"name"`
	Description    string   `json:"description" yaml:"description"`
	SystemPrompt   string   `json:"system_prompt" yaml:"-"`
	AllowedTools   []string `json:"allowed_tools,omitempty" yaml:"allowed_tools"`
	DeniedTools    []string `json:"denied_tools,omitempty" yaml:"denied_tools"`
	DelegatesTo    []string `json:"delegates_to,omitempty" yaml:"delegates_to"`
	VetoRules      []string `json:"veto_rules,omitempty" yaml:"veto_rules"`
	Priority       int      `json:"priority,omitempty" yaml:"priority"`
	MaxTokenBudget int      `json:"max_token_budget,omitempty" yaml:"max_token_budget"`
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
