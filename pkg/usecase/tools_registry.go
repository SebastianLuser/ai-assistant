package usecase

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"jarvis/clients"
	"jarvis/internal/skills"
	"jarvis/internal/tracing"
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

// ReminderSender is called when a reminder fires.
type ReminderSender func(text string)

// ReminderManager tracks pending reminders so they can be created from tools.
type ReminderManager struct {
	mu       sync.Mutex
	sender   ReminderSender
}

// NewReminderManager creates a reminder manager.
func NewReminderManager(sender ReminderSender) *ReminderManager {
	return &ReminderManager{sender: sender}
}

// Schedule creates a one-shot reminder that fires after the given duration.
func (rm *ReminderManager) Schedule(after time.Duration, text string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	go func() {
		time.Sleep(after)
		log.Printf("reminder: firing: %s", text)
		if rm.sender != nil {
			rm.sender(text)
		}
	}()
}

// BuildToolRegistry creates a ToolRegistry with handlers for all available integrations.
// Only registers tools for non-nil clients.
func BuildToolRegistry(
	financeUC *FinanceUseCase,
	memorySvc service.MemoryService,
	embedder service.Embedder,
	calendarClient *clients.CalendarClient,
	gmailClient *clients.GmailClient,
	todoistClient *clients.TodoistClient,
	githubClient *clients.GitHubClient,
	jiraClient *clients.JiraClient,
	spotifyClient *clients.SpotifyClient,
	_ *clients.NotionClient,
	_ *clients.ObsidianVault,
	skillWriter skills.SkillWriter,
	reminderMgr *ReminderManager,
) *ToolRegistry {
	r := NewToolRegistry()

	// --- Finance ---
	if financeUC != nil {
		r.RegisterWithLimit(domain.ToolDefinition{
			Name:        "save_expense",
			Description: "Registra un gasto. El usuario dice algo como 'gasté 500 en el super' y se parsea automáticamente.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message": map[string]any{"type": "string", "description": "El mensaje del usuario describiendo el gasto"},
				},
				"required": []string{"message"},
			},
		}, func(_ context.Context, input map[string]any) (string, error) {
			return financeUC.ProcessExpense(inputString(input, "message"), "Sebas")
		}, 5, time.Minute)
	}

	// --- Memory / Notes ---
	if memorySvc != nil {
		r.Register(domain.ToolDefinition{
			Name:        "save_note",
			Description: "Guarda una nota o recordatorio para el usuario.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content": map[string]any{"type": "string", "description": "Contenido de la nota"},
					"tags":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tags opcionales"},
				},
				"required": []string{"content"},
			},
		}, func(ctx context.Context, input map[string]any) (string, error) {
			content := inputString(input, "content")
			var tags []string
			if t, ok := input["tags"].([]any); ok {
				for _, v := range t {
					if s, ok := v.(string); ok {
						tags = append(tags, s)
					}
				}
			}
			if p := tracing.Profile(ctx); p != "" {
				tags = append(tags, "profile:"+p)
			}
			if ch := tracing.Channel(ctx); ch != "" {
				tags = append(tags, "channel:"+ch)
			}
			for _, t := range tracing.ClassifiedTags(ctx) {
				tags = append(tags, "auto:"+t)
			}
			var emb []float64
			if embedder != nil {
				emb, _ = embedder.Embed(content)
			}
			id, err := memorySvc.Save(content, tags, emb)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Nota guardada (id: %d)", id), nil
		})

		r.Register(domain.ToolDefinition{
			Name:        "search_notes",
			Description: "Busca en las notas guardadas del usuario. Usa búsqueda semántica si hay embeddings, o texto completo si no.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string", "description": "Texto a buscar"},
					"limit": map[string]any{"type": "number", "description": "Cantidad máxima de resultados (default 5)"},
				},
				"required": []string{"query"},
			},
		}, func(_ context.Context, input map[string]any) (string, error) {
			query := inputString(input, "query")
			limit := 5
			if l, ok := input["limit"].(float64); ok && l > 0 {
				limit = int(l)
			}

			// Try hybrid search first (vector + FTS) if embedder available
			if embedder != nil {
				emb, err := embedder.Embed(query)
				if err == nil {
					results, err := memorySvc.SearchHybrid(query, emb, limit, 0.6, 0.4)
					if err == nil && len(results) > 0 {
						return toJSON(results), nil
					}
				}
			}

			// Fallback to FTS
			results, err := memorySvc.SearchFTS(query, limit)
			if err != nil {
				return "", err
			}
			return toJSON(results), nil
		})
	}

	// --- Calendar ---
	if calendarClient != nil {
		r.Register(domain.ToolDefinition{
			Name:        "get_today_events",
			Description: "Muestra los eventos del calendario de hoy.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		}, func(_ context.Context, input map[string]any) (string, error) {
			events, err := calendarClient.GetTodayEvents()
			if err != nil {
				return "", err
			}
			return toJSON(events), nil
		})

		r.RegisterWithLimit(domain.ToolDefinition{
			Name:        "create_event",
			Description: "Crea un evento en el calendario.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"summary": map[string]any{"type": "string", "description": "Título del evento"},
					"start":   map[string]any{"type": "string", "description": "Fecha/hora inicio RFC3339"},
					"end":     map[string]any{"type": "string", "description": "Fecha/hora fin RFC3339"},
				},
				"required": []string{"summary", "start", "end"},
			},
		}, func(_ context.Context, input map[string]any) (string, error) {
			start, err := time.Parse(time.RFC3339, inputString(input, "start"))
			if err != nil {
				return "", fmt.Errorf("invalid start date: %w", err)
			}
			end, err := time.Parse(time.RFC3339, inputString(input, "end"))
			if err != nil {
				return "", fmt.Errorf("invalid end date: %w", err)
			}
			event, err := calendarClient.CreateEvent(inputString(input, "summary"), start, end)
			if err != nil {
				return "", err
			}
			return toJSON(event), nil
		}, 5, time.Minute)
	}

	// --- Gmail ---
	if gmailClient != nil {
		r.Register(domain.ToolDefinition{
			Name:        "get_unread_emails",
			Description: "Lista los emails no leídos (últimos 5).",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		}, func(_ context.Context, input map[string]any) (string, error) {
			emails, err := gmailClient.ListUnread(5)
			if err != nil {
				return "", err
			}
			return toJSON(emails), nil
		})
	}

	// --- Todoist ---
	if todoistClient != nil {
		r.Register(domain.ToolDefinition{
			Name:        "list_tasks",
			Description: "Lista las tareas pendientes de Todoist.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		}, func(_ context.Context, input map[string]any) (string, error) {
			tasks, err := todoistClient.GetTasks()
			if err != nil {
				return "", err
			}
			return toJSON(tasks), nil
		})

		r.RegisterWithLimit(domain.ToolDefinition{
			Name:        "create_task",
			Description: "Crea una tarea nueva en Todoist.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content":  map[string]any{"type": "string", "description": "Texto de la tarea"},
					"due_date": map[string]any{"type": "string", "description": "Fecha límite YYYY-MM-DD (opcional)"},
				},
				"required": []string{"content"},
			},
		}, func(_ context.Context, input map[string]any) (string, error) {
			dueDate := inputString(input, "due_date")
			var dueDatePtr *string
			if dueDate != "" {
				dueDatePtr = &dueDate
			}
			task, err := todoistClient.CreateTask(inputString(input, "content"), dueDatePtr)
			if err != nil {
				return "", err
			}
			return toJSON(task), nil
		}, 5, time.Minute)
	}

	// --- GitHub ---
	if githubClient != nil {
		r.Register(domain.ToolDefinition{
			Name:        "list_github_issues",
			Description: "Lista issues de un repo de GitHub.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"owner": map[string]any{"type": "string"},
					"repo":  map[string]any{"type": "string"},
				},
				"required": []string{"owner", "repo"},
			},
		}, func(_ context.Context, input map[string]any) (string, error) {
			issues, err := githubClient.ListIssues(inputString(input, "owner"), inputString(input, "repo"))
			if err != nil {
				return "", err
			}
			return toJSON(issues), nil
		})

		r.RegisterWithLimit(domain.ToolDefinition{
			Name:        "create_github_issue",
			Description: "Crea un issue en un repo de GitHub.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"owner": map[string]any{"type": "string"},
					"repo":  map[string]any{"type": "string"},
					"title": map[string]any{"type": "string"},
					"body":  map[string]any{"type": "string"},
				},
				"required": []string{"owner", "repo", "title"},
			},
		}, func(_ context.Context, input map[string]any) (string, error) {
			issue, err := githubClient.CreateIssue(inputString(input, "owner"), inputString(input, "repo"), inputString(input, "title"), inputString(input, "body"))
			if err != nil {
				return "", err
			}
			return toJSON(issue), nil
		}, 5, time.Minute)
	}

	// --- Jira ---
	if jiraClient != nil {
		r.Register(domain.ToolDefinition{
			Name:        "get_my_jira_issues",
			Description: "Lista mis issues asignados en Jira.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		}, func(_ context.Context, input map[string]any) (string, error) {
			issues, err := jiraClient.GetMyIssues()
			if err != nil {
				return "", err
			}
			return toJSON(issues), nil
		})
	}

	// --- Spotify ---
	if spotifyClient != nil {
		r.Register(domain.ToolDefinition{
			Name:        "spotify_now_playing",
			Description: "Muestra qué canción se está reproduciendo en Spotify.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		}, func(_ context.Context, input map[string]any) (string, error) {
			track, err := spotifyClient.GetCurrentlyPlaying()
			if err != nil {
				return "", err
			}
			return toJSON(track), nil
		})

		r.Register(domain.ToolDefinition{
			Name:        "spotify_next",
			Description: "Salta a la siguiente canción en Spotify.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		}, func(_ context.Context, input map[string]any) (string, error) {
			return "Canción siguiente", spotifyClient.Next()
		})
	}

	// --- Self-modifying skills ---
	if skillWriter != nil {
		r.RegisterWithLimit(domain.ToolDefinition{
			Name:        "create_skill",
			Description: "Crea una nueva habilidad/skill para el asistente. Usá esto cuando el usuario te pida recordar un comportamiento, estilo, o conocimiento específico para futuras conversaciones.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":        map[string]any{"type": "string", "description": "Nombre corto del skill (slug)"},
					"description": map[string]any{"type": "string", "description": "Descripción breve de qué hace el skill"},
					"content":     map[string]any{"type": "string", "description": "Instrucciones del skill (system prompt)"},
					"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tags para clasificar el skill"},
				},
				"required": []string{"name", "content"},
			},
		}, func(_ context.Context, input map[string]any) (string, error) {
			enabled := true
			var tags []string
			if t, ok := input["tags"].([]any); ok {
				for _, v := range t {
					if s, ok := v.(string); ok {
						tags = append(tags, s)
					}
				}
			}
			skill := skills.Skill{
				Name:        inputString(input, "name"),
				Description: inputString(input, "description"),
				Content:     inputString(input, "content"),
				Tags:        tags,
				Enabled:     &enabled,
			}
			if err := skillWriter.Save(skill); err != nil {
				return "", err
			}
			return fmt.Sprintf("Skill '%s' creado. Se activará en las próximas conversaciones.", skill.Name), nil
		}, 3, time.Minute)
	}

	// --- Usage tracking ---
	r.Register(domain.ToolDefinition{
		Name:        "get_usage",
		Description: "Muestra el consumo de tokens y costo estimado de la sesión actual y el total.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	}, func(_ context.Context, input map[string]any) (string, error) {
		return "Usage tracking disponible via GET /api/usage", nil
	})

	// --- Scheduled reminders ---
	if reminderMgr != nil {
		r.RegisterWithLimit(domain.ToolDefinition{
			Name:        "set_reminder",
			Description: "Programa un recordatorio que se enviará después de un tiempo. Ej: 'recordame en 30 minutos que tengo que llamar'.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message": map[string]any{"type": "string", "description": "Texto del recordatorio"},
					"minutes": map[string]any{"type": "number", "description": "En cuántos minutos recordar"},
				},
				"required": []string{"message", "minutes"},
			},
		}, func(_ context.Context, input map[string]any) (string, error) {
			msg := inputString(input, "message")
			minutes := 0.0
			if m, ok := input["minutes"].(float64); ok {
				minutes = m
			}
			if minutes <= 0 {
				return "", fmt.Errorf("minutes must be positive")
			}
			duration := time.Duration(minutes) * time.Minute
			reminderMgr.Schedule(duration, "⏰ Recordatorio: "+msg)
			return fmt.Sprintf("Recordatorio programado para dentro de %.0f minutos.", minutes), nil
		}, 5, time.Minute)
	}

	return r
}
