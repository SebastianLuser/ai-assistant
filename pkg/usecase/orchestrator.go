package usecase

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"asistente/pkg/domain"
)

// AgentOrchestrator manages multiple specialized sub-agents.
// The main agent can delegate tasks to sub-agents, which run with their own
// system prompt and tool subset. Inspired by OpenClaw's sessions_spawn pattern.
type AgentOrchestrator struct {
	ai      domain.ToolUseProvider
	tools   *ToolRegistry
	agents  map[string]domain.AgentDefinition
	ordered []domain.AgentDefinition // for the orchestrator prompt

	mu      sync.Mutex
	results map[string]domain.SubAgentResult // runID -> result
}

// NewAgentOrchestrator creates an orchestrator with the given agent definitions.
func NewAgentOrchestrator(ai domain.ToolUseProvider, tools *ToolRegistry, agents []domain.AgentDefinition) *AgentOrchestrator {
	agentMap := make(map[string]domain.AgentDefinition, len(agents))
	for _, a := range agents {
		agentMap[a.ID] = a
	}
	return &AgentOrchestrator{
		ai:      ai,
		tools:   tools,
		agents:  agentMap,
		ordered: agents,
		results: make(map[string]domain.SubAgentResult),
	}
}

// Run executes the orchestrator agent. It has access to all tools plus a special
// "delegate_to_agent" tool that spawns sub-agent runs.
// The orchestrator decides whether to handle the task itself or delegate.
func (o *AgentOrchestrator) Run(baseSystem string, messages []domain.Message) (string, error) {
	// Build orchestrator system prompt with agent catalog
	system := o.buildOrchestratorPrompt(baseSystem)

	// Create a tool registry that includes delegate_to_agent
	orchTools := o.buildOrchestratorTools()

	agent := &AgentUseCase{
		ai:       o.ai,
		tools:    orchTools,
		maxTurns: agentMaxTurns,
	}

	return agent.Run(system, messages)
}

func (o *AgentOrchestrator) buildOrchestratorPrompt(base string) string {
	var sb strings.Builder
	sb.WriteString(base)
	sb.WriteString("\n## Agentes especializados disponibles\n\n")
	sb.WriteString("Podés delegar tareas a estos agentes especializados usando la herramienta `delegate_to_agent`. ")
	sb.WriteString("Cada agente tiene su propia expertise y herramientas. Usá delegación cuando la tarea encaje mejor con un agente específico. ")
	sb.WriteString("Podés delegar a varios agentes en paralelo si la tarea tiene partes independientes.\n\n")

	for _, a := range o.ordered {
		sb.WriteString(fmt.Sprintf("- **%s** (`%s`): %s\n", a.Name, a.ID, a.Description))
	}
	sb.WriteString("\nSi la tarea es simple o no encaja con ningún agente, resolvela vos directamente con tus herramientas.\n")

	return sb.String()
}

func (o *AgentOrchestrator) buildOrchestratorTools() *ToolRegistry {
	orchTools := NewToolRegistry()

	// Copy all existing tools
	for _, def := range o.tools.Definitions() {
		name := def.Name
		orchTools.Register(def, func(input map[string]any) (string, error) {
			return o.tools.Execute(name, input)
		})
	}

	// Add the delegation tool
	orchTools.Register(domain.ToolDefinition{
		Name:        "delegate_to_agent",
		Description: "Delega una tarea a un agente especializado. El agente ejecuta la tarea con su propia expertise y herramientas, y devuelve el resultado.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent_id": map[string]any{
					"type":        "string",
					"description": "ID del agente al que delegar",
					"enum":        o.agentIDs(),
				},
				"task": map[string]any{
					"type":        "string",
					"description": "Descripción de la tarea a delegar. Sé específico.",
				},
			},
			"required": []string{"agent_id", "task"},
		},
	}, func(input map[string]any) (string, error) {
		agentID := inputString(input, "agent_id")
		task := inputString(input, "task")
		return o.runSubAgent(agentID, task)
	})

	return orchTools
}

func (o *AgentOrchestrator) agentIDs() []string {
	ids := make([]string, len(o.ordered))
	for i, a := range o.ordered {
		ids[i] = a.ID
	}
	return ids
}

// runSubAgent executes a task with a specialized sub-agent.
func (o *AgentOrchestrator) runSubAgent(agentID, task string) (string, error) {
	agentDef, ok := o.agents[agentID]
	if !ok {
		return "", fmt.Errorf("agent not found: %s", agentID)
	}

	log.Printf("orchestrator: delegating to agent '%s': %s", agentID, truncate(task, 80))

	// Build sub-agent with filtered tools
	subTools := o.filterTools(agentDef)

	subAgent := &AgentUseCase{
		ai:       o.ai,
		tools:    subTools,
		maxTurns: agentMaxTurns,
	}

	// Run with the sub-agent's system prompt
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: task},
	}

	result, err := subAgent.Run(agentDef.SystemPrompt, messages)
	status := "success"
	if err != nil {
		status = "error"
		log.Printf("orchestrator: agent '%s' failed: %v", agentID, err)
		result = fmt.Sprintf("Error del agente %s: %v", agentDef.Name, err)
	}

	log.Printf("orchestrator: agent '%s' completed (%s): %s", agentID, status, truncate(result, 100))

	return result, nil
}

// filterTools returns a ToolRegistry with only the tools allowed for this agent.
// Sub-agents never get the delegate_to_agent tool (prevents recursion).
func (o *AgentOrchestrator) filterTools(agent domain.AgentDefinition) *ToolRegistry {
	denied := make(map[string]bool)
	denied["delegate_to_agent"] = true // always deny delegation for sub-agents
	for _, t := range agent.DeniedTools {
		denied[t] = true
	}

	allowed := make(map[string]bool)
	for _, t := range agent.AllowedTools {
		allowed[t] = true
	}

	filtered := NewToolRegistry()
	for _, def := range o.tools.Definitions() {
		if denied[def.Name] {
			continue
		}
		if len(allowed) > 0 && !allowed[def.Name] {
			continue
		}
		name := def.Name
		filtered.Register(def, func(input map[string]any) (string, error) {
			return o.tools.Execute(name, input)
		})
	}
	return filtered
}
