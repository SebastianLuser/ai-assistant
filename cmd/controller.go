package main

import (
	"net/http"

	"asistente/config"
	"asistente/internal/hooks"
	"asistente/internal/skills"
	"asistente/pkg/controller"
	"asistente/pkg/domain"
	"asistente/pkg/service"
	"asistente/pkg/usecase"
	"asistente/web"
)

type Controllers struct {
	Finance  *controller.FinanceController
	Memory   *controller.MemoryController
	Chat     *controller.ConversationController
	WhatsApp *controller.WhatsAppController
	Notion   *controller.NotionController
	Obsidian *controller.ObsidianController
	Calendar *controller.CalendarController
	GitHub   *controller.GitHubController
	Jira     *controller.JiraController
	Spotify  *controller.SpotifyController
	Todoist  *controller.TodoistController
	Gmail    *controller.GmailController
	ClickUp  *controller.ClickUpController
	Habit    *controller.HabitController
	Link     *controller.LinkController
	Project  *controller.ProjectController
	Figma    *controller.FigmaController
	Telegram *controller.TelegramController
	Skill    *controller.SkillController
	Trigger  *controller.TriggerController
	Usage    *controller.UsageController
	Pairing  web.Handler
}

func NewControllers(
	cl Clients,
	cfg config.Config,
	memorySvc service.MemoryService,
	financeSvc service.FinanceService,
	embedder service.Embedder,
	skillsLoader skills.SkillProvider,
	hooksRegistry *hooks.Registry,
	scheduler *usecase.Scheduler,
) Controllers {
	financeUC := usecase.NewFinanceUseCase(cl.AI, financeSvc)
	financeUC.SetMemoryService(memorySvc)

	chatUC := usecase.NewConversationUseCase(memorySvc, cl.AI, hooksRegistry, cfg.MaxHistoryMsgs, cfg.CompactThreshold)

	c := Controllers{
		Finance: controller.NewFinanceController(financeUC),
		Memory:  controller.NewMemoryController(memorySvc, embedder),
		Chat:    controller.NewConversationController(chatUC, cl.AI, skillsLoader, hooksRegistry),
		Habit:   controller.NewHabitController(usecase.NewHabitUseCase(memorySvc)),
		Link:    controller.NewLinkController(usecase.NewLinkUseCase(memorySvc, embedder)),
		Project: controller.NewProjectController(usecase.NewProjectUseCase(memorySvc, embedder, cl.AI)),
	}

	if cl.Notion != nil {
		c.Notion = controller.NewNotionController(cl.Notion, cfg.NotionPageID)
	}
	if cl.Obsidian != nil {
		c.Obsidian = controller.NewObsidianController(cl.Obsidian)
	}
	if cl.Calendar != nil {
		c.Calendar = controller.NewCalendarController(cl.Calendar)
	}
	if cl.GitHub != nil {
		c.GitHub = controller.NewGitHubController(cl.GitHub)
	}
	if cl.Jira != nil {
		c.Jira = controller.NewJiraController(cl.Jira)
	}
	if cl.Spotify != nil {
		c.Spotify = controller.NewSpotifyController(cl.Spotify)
	}
	if cl.Todoist != nil {
		c.Todoist = controller.NewTodoistController(cl.Todoist)
	}
	if cl.Gmail != nil {
		c.Gmail = controller.NewGmailController(cl.Gmail)
	}
	if cl.ClickUp != nil {
		c.ClickUp = controller.NewClickUpController(cl.ClickUp)
	}

	if sw, ok := skillsLoader.(skills.SkillWriter); ok {
		c.Skill = controller.NewSkillController(sw)
	}

	c.Trigger = controller.NewTriggerController(scheduler)

	usageTracker := usecase.NewUsageTracker()
	c.Usage = controller.NewUsageController(usageTracker)

	if cl.Figma != nil {
		c.Figma = controller.NewFigmaController(cl.Figma)
	}

	// Shared MessageRouter for all messaging channels.
	var router *usecase.MessageRouter
	if cl.WhatsApp != nil || cl.Telegram != nil {
		// Reminder manager sends reminders via the first available channel.
		var reminderMgr *usecase.ReminderManager
		if cl.WhatsApp != nil && cfg.WhatsAppTo != "" {
			reminderMgr = usecase.NewReminderManager(func(text string) {
				_ = cl.WhatsApp.SendTextMessage(cfg.WhatsAppTo, text)
			})
		}

		var sw skills.SkillWriter
		if writer, ok := skillsLoader.(skills.SkillWriter); ok {
			sw = writer
		}

		toolReg := usecase.BuildToolRegistry(financeUC, memorySvc, embedder, cl.Calendar, cl.Gmail, cl.Todoist, cl.GitHub, cl.Jira, cl.Spotify, cl.Notion, cl.Obsidian, sw, reminderMgr)

		var agent *usecase.AgentUseCase
		var orchestrator *usecase.AgentOrchestrator
		if tp, ok := cl.AI.(domain.ToolUseProvider); ok {
			agent = usecase.NewAgentUseCase(tp, toolReg)
			orchestrator = usecase.NewAgentOrchestrator(tp, toolReg, usecase.DefaultAgents())
		}

		var transcriber domain.Transcriber
		if cl.Transcriber != nil {
			transcriber = cl.Transcriber
		}

		router = usecase.NewMessageRouter(chatUC, cl.AI, agent, orchestrator, transcriber, skillsLoader, hooksRegistry, usageTracker, cfg.WhatsAppTo)
	}

	if router != nil {
		c.Pairing = func(req web.Request) web.Response {
			return web.NewJSONResponse(http.StatusOK, map[string]any{
				"success": true, "pairing_code": router.GetPairingCode(),
			})
		}
	}

	if cl.WhatsApp != nil && cfg.WhatsAppVerifyToken != "" && router != nil {
		c.WhatsApp = controller.NewWhatsAppController(router, cl.WhatsApp, cfg.WhatsAppVerifyToken, cfg.WhatsAppAppSecret)
	}

	if cl.Telegram != nil && router != nil {
		c.Telegram = controller.NewTelegramController(router, cl.Telegram, cfg.TelegramSecretToken, cfg.TelegramBotUsername)
	}

	return c
}
