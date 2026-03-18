package usecase

import (
	"fmt"
	"time"

	"asistente/clients"
	"asistente/pkg/domain"
	"asistente/pkg/service"
)

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
	notionClient *clients.NotionClient,
	obsidianClient *clients.ObsidianVault,
) *ToolRegistry {
	r := NewToolRegistry()

	// --- Finance ---
	if financeUC != nil {
		r.Register(domain.ToolDefinition{
			Name:        "save_expense",
			Description: "Registra un gasto. El usuario dice algo como 'gasté 500 en el super' y se parsea automáticamente.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message": map[string]any{"type": "string", "description": "El mensaje del usuario describiendo el gasto"},
				},
				"required": []string{"message"},
			},
		}, func(input map[string]any) (string, error) {
			return financeUC.ProcessExpense(inputString(input, "message"), "Sebas")
		})
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
		}, func(input map[string]any) (string, error) {
			content := inputString(input, "content")
			var tags []string
			if t, ok := input["tags"].([]any); ok {
				for _, v := range t {
					if s, ok := v.(string); ok {
						tags = append(tags, s)
					}
				}
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
			Description: "Busca en las notas guardadas del usuario.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string", "description": "Texto a buscar"},
				},
				"required": []string{"query"},
			},
		}, func(input map[string]any) (string, error) {
			results, err := memorySvc.SearchFTS(inputString(input, "query"), 5)
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
		}, func(input map[string]any) (string, error) {
			events, err := calendarClient.GetTodayEvents()
			if err != nil {
				return "", err
			}
			return toJSON(events), nil
		})

		r.Register(domain.ToolDefinition{
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
		}, func(input map[string]any) (string, error) {
			start, _ := time.Parse(time.RFC3339, inputString(input, "start"))
			end, _ := time.Parse(time.RFC3339, inputString(input, "end"))
			event, err := calendarClient.CreateEvent(inputString(input, "summary"), start, end)
			if err != nil {
				return "", err
			}
			return toJSON(event), nil
		})
	}

	// --- Gmail ---
	if gmailClient != nil {
		r.Register(domain.ToolDefinition{
			Name:        "get_unread_emails",
			Description: "Lista los emails no leídos (últimos 5).",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		}, func(input map[string]any) (string, error) {
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
		}, func(input map[string]any) (string, error) {
			tasks, err := todoistClient.GetTasks()
			if err != nil {
				return "", err
			}
			return toJSON(tasks), nil
		})

		r.Register(domain.ToolDefinition{
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
		}, func(input map[string]any) (string, error) {
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
		})
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
		}, func(input map[string]any) (string, error) {
			issues, err := githubClient.ListIssues(inputString(input, "owner"), inputString(input, "repo"))
			if err != nil {
				return "", err
			}
			return toJSON(issues), nil
		})

		r.Register(domain.ToolDefinition{
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
		}, func(input map[string]any) (string, error) {
			issue, err := githubClient.CreateIssue(inputString(input, "owner"), inputString(input, "repo"), inputString(input, "title"), inputString(input, "body"))
			if err != nil {
				return "", err
			}
			return toJSON(issue), nil
		})
	}

	// --- Jira ---
	if jiraClient != nil {
		r.Register(domain.ToolDefinition{
			Name:        "get_my_jira_issues",
			Description: "Lista mis issues asignados en Jira.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		}, func(input map[string]any) (string, error) {
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
		}, func(input map[string]any) (string, error) {
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
		}, func(input map[string]any) (string, error) {
			return "Canción siguiente", spotifyClient.Next()
		})
	}

	return r
}
