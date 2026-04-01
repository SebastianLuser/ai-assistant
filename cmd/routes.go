package main

import (
	"context"

	"jarvis/boot"
	"jarvis/internal/middleware"
	"jarvis/web"
	webgin "jarvis/web/gin"
)

func setupRoutes(c Controllers) boot.RoutesMapper[boot.GinRouter] {
	return func(_ context.Context, _ boot.Config, router boot.GinRouter) {
		registerRoutes(router, c)
	}
}

func middlewareMapper(webhookSecret string) boot.MiddlewareMapper[boot.GinMiddlewareRouter] {
	return func(_ context.Context, _ boot.Config, router boot.GinMiddlewareRouter) {
		if webhookSecret != "" {
			router.Use(webgin.NewInterceptor(middleware.WebhookAuth(webhookSecret)))
		}
	}
}

func registerRoutes(router boot.GinRouter, c Controllers) {
	router.GET("/health", webgin.NewHandlerJSON(func(req web.Request) web.Response {
		return web.NewJSONResponse(200, map[string]string{"status": "healthy"})
	}))

	webhook := router.Group("/webhook")
	if c.WhatsApp != nil {
		webhook.GET("/whatsapp", webgin.NewHandlerRaw(c.WhatsApp.VerifyWebhook))
		webhook.POST("/whatsapp", webgin.NewHandlerJSON(c.WhatsApp.HandleWebhook))
	}
	if c.Telegram != nil {
		webhook.POST("/telegram", webgin.NewHandlerJSON(c.Telegram.HandleWebhook))
	}

	api := router.Group("/api")

	api.POST("/finance/expense", webgin.NewHandlerJSON(c.Finance.PostExpense))
	api.GET("/finance/summary", webgin.NewHandlerJSON(c.Finance.GetSummary))

	api.POST("/memory/note", webgin.NewHandlerJSON(c.Memory.PostNote))
	api.GET("/memory/search", webgin.NewHandlerJSON(c.Memory.GetSearch))
	api.DELETE("/memory/note/:id", webgin.NewHandlerJSON(c.Memory.DeleteNote))

	api.POST("/chat", webgin.NewHandlerJSON(c.Chat.PostChat))

	if c.Notion != nil {
		api.POST("/notion/page", webgin.NewHandlerJSON(c.Notion.CreatePage))
		api.GET("/notion/page/:id", webgin.NewHandlerJSON(c.Notion.GetPage))
	}

	if c.Obsidian != nil {
		api.GET("/obsidian/note", webgin.NewHandlerJSON(c.Obsidian.ReadNote))
		api.POST("/obsidian/note", webgin.NewHandlerJSON(c.Obsidian.WriteNote))
		api.GET("/obsidian/notes", webgin.NewHandlerJSON(c.Obsidian.ListNotes))
		api.GET("/obsidian/search", webgin.NewHandlerJSON(c.Obsidian.SearchNotes))
	}

	if c.Calendar != nil {
		api.GET("/calendar/today", webgin.NewHandlerJSON(c.Calendar.GetTodayEvents))
		api.POST("/calendar/event", webgin.NewHandlerJSON(c.Calendar.CreateEvent))
	}

	if c.GitHub != nil {
		api.GET("/github/repos", webgin.NewHandlerJSON(c.GitHub.ListRepos))
		api.GET("/github/:owner/:repo/issues", webgin.NewHandlerJSON(c.GitHub.ListIssues))
		api.POST("/github/:owner/:repo/issues", webgin.NewHandlerJSON(c.GitHub.CreateIssue))
		api.GET("/github/:owner/:repo/pulls", webgin.NewHandlerJSON(c.GitHub.ListPRs))
	}

	if c.Jira != nil {
		api.GET("/jira/my-issues", webgin.NewHandlerJSON(c.Jira.GetMyIssues))
		api.GET("/jira/issue/:key", webgin.NewHandlerJSON(c.Jira.GetIssue))
		api.POST("/jira/issue", webgin.NewHandlerJSON(c.Jira.CreateIssue))
	}

	if c.Spotify != nil {
		api.GET("/spotify/playing", webgin.NewHandlerJSON(c.Spotify.GetCurrentlyPlaying))
		api.POST("/spotify/play", webgin.NewHandlerJSON(c.Spotify.Play))
		api.POST("/spotify/pause", webgin.NewHandlerJSON(c.Spotify.Pause))
		api.POST("/spotify/next", webgin.NewHandlerJSON(c.Spotify.Next))
	}

	if c.Todoist != nil {
		api.GET("/todoist/tasks", webgin.NewHandlerJSON(c.Todoist.GetTasks))
		api.POST("/todoist/task", webgin.NewHandlerJSON(c.Todoist.CreateTask))
		api.POST("/todoist/task/:id/complete", webgin.NewHandlerJSON(c.Todoist.CompleteTask))
	}

	if c.Gmail != nil {
		api.GET("/gmail/unread", webgin.NewHandlerJSON(c.Gmail.ListUnread))
		api.GET("/gmail/message/:id", webgin.NewHandlerJSON(c.Gmail.GetMessage))
	}

	if c.ClickUp != nil {
		api.GET("/clickup/tasks", webgin.NewHandlerJSON(c.ClickUp.GetMyTasks))
		api.GET("/clickup/task/:id", webgin.NewHandlerJSON(c.ClickUp.GetTask))
		api.POST("/clickup/task", webgin.NewHandlerJSON(c.ClickUp.CreateTask))
		api.PUT("/clickup/task/:id/status", webgin.NewHandlerJSON(c.ClickUp.UpdateTaskStatus))
	}

	if c.Habit != nil {
		api.POST("/habits/log", webgin.NewHandlerJSON(c.Habit.PostLog))
		api.GET("/habits/streak", webgin.NewHandlerJSON(c.Habit.GetStreak))
		api.GET("/habits/today", webgin.NewHandlerJSON(c.Habit.GetToday))
	}

	if c.Link != nil {
		api.POST("/links", webgin.NewHandlerJSON(c.Link.PostLink))
		api.GET("/links/search", webgin.NewHandlerJSON(c.Link.GetSearch))
	}

	if c.Project != nil {
		api.GET("/projects/:name/status", webgin.NewHandlerJSON(c.Project.GetStatus))
	}

	if c.Skill != nil {
		api.GET("/skills", webgin.NewHandlerJSON(c.Skill.ListSkills))
		api.POST("/skills", webgin.NewHandlerJSON(c.Skill.CreateSkill))
	}

	if c.Trigger != nil {
		api.GET("/triggers/jobs", webgin.NewHandlerJSON(c.Trigger.ListJobs))
		api.POST("/triggers/job/:job_id", webgin.NewHandlerJSON(c.Trigger.TriggerJob))
	}

	if c.Pairing != nil {
		api.GET("/pairing-code", webgin.NewHandlerJSON(c.Pairing))
	}

	if c.Usage != nil {
		api.GET("/usage", webgin.NewHandlerJSON(c.Usage.GetUsage))
	}

	if c.Figma != nil {
		api.GET("/figma/file/:file_key", webgin.NewHandlerJSON(c.Figma.GetFile))
		api.GET("/figma/file/:file_key/nodes", webgin.NewHandlerJSON(c.Figma.GetNodes))
		api.GET("/figma/file/:file_key/images", webgin.NewHandlerJSON(c.Figma.GetImages))
		api.GET("/figma/file/:file_key/comments", webgin.NewHandlerJSON(c.Figma.GetComments))
		api.GET("/figma/file/:file_key/components", webgin.NewHandlerJSON(c.Figma.GetComponents))
		api.GET("/figma/project/:project_id/files", webgin.NewHandlerJSON(c.Figma.GetProjectFiles))
	}
}
