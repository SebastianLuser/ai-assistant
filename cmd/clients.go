package main

import (
	"log"

	"jarvis/clients"
	"jarvis/config"
	"jarvis/pkg/domain"
)

type Clients struct {
	AI       domain.AIProvider
	AILight  domain.AIProvider
	Sheets   *clients.SheetsClient
	WhatsApp *clients.WhatsAppClient
	Calendar *clients.CalendarClient
	Notion   *clients.NotionClient
	Obsidian *clients.ObsidianVault
	GitHub   *clients.GitHubClient
	Jira     *clients.JiraClient
	Spotify  *clients.SpotifyClient
	Todoist  *clients.TodoistClient
	Gmail    *clients.GmailClient
	ClickUp  *clients.ClickUpClient
	Figma       *clients.FigmaClient
	Telegram    *clients.TelegramClient
	Transcriber domain.Transcriber
}

func NewClients(cfg config.Config) Clients {
	ai := newAIProviderWithFailover(cfg)
	return Clients{
		AI:       ai,
		AILight:  newAILightProvider(cfg, ai),
		Sheets:   newSheetsClient(cfg),
		WhatsApp: newWhatsAppClient(cfg),
		Calendar: newCalendarClient(cfg),
		Notion:   newNotionClient(cfg),
		Obsidian: newObsidianVault(cfg),
		GitHub:   newGitHubClient(cfg),
		Jira:     newJiraClient(cfg),
		Spotify:  newSpotifyClient(cfg),
		Todoist:  newTodoistClient(cfg),
		Gmail:    newGmailClient(cfg),
		ClickUp:  newClickUpClient(cfg),
		Figma:       newFigmaClient(cfg),
		Telegram:    newTelegramClient(cfg),
		Transcriber: newTranscriber(cfg),
	}
}

func newAIProviderWithFailover(cfg config.Config) domain.AIProvider {
	primary := newAIProvider(cfg)
	fallback := newAIFallback(cfg)
	return clients.NewFailoverProvider(primary, fallback)
}

func newAIProvider(cfg config.Config) domain.AIProvider {
	switch cfg.AIProvider {
	case "openai":
		log.Printf("AI provider: OpenAI (model: %s)", cfg.OpenAIModel)
		return clients.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	default:
		log.Printf("AI provider: Claude (model: %s)", cfg.ClaudeModel)
		return clients.NewClaudeClient(cfg.ClaudeAPIKey, cfg.ClaudeModel)
	}
}

func newAIFallback(cfg config.Config) domain.AIProvider {
	switch cfg.AIProvider {
	case "openai":
		if cfg.ClaudeAPIKey == "" {
			return nil
		}
		log.Printf("AI fallback: Claude (model: %s)", cfg.ClaudeModel)
		return clients.NewClaudeClient(cfg.ClaudeAPIKey, cfg.ClaudeModel)
	default:
		if cfg.OpenAIAPIKey == "" {
			return nil
		}
		log.Printf("AI fallback: OpenAI (model: %s)", cfg.OpenAIModel)
		return clients.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	}
}

func newAILightProvider(cfg config.Config, primary domain.AIProvider) domain.AIProvider {
	switch cfg.AIProvider {
	case "openai":
		if cfg.OpenAIModelLight == "" || cfg.OpenAIModelLight == cfg.OpenAIModel {
			return primary
		}
		log.Printf("AI light provider: OpenAI (model: %s)", cfg.OpenAIModelLight)
		return clients.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModelLight)
	default:
		if cfg.ClaudeModelLight == "" || cfg.ClaudeModelLight == cfg.ClaudeModel {
			return primary
		}
		log.Printf("AI light provider: Claude (model: %s)", cfg.ClaudeModelLight)
		return clients.NewClaudeClient(cfg.ClaudeAPIKey, cfg.ClaudeModelLight)
	}
}

func newSheetsClient(cfg config.Config) *clients.SheetsClient {
	if cfg.SheetsID == "" || cfg.SheetsCredFile == "" {
		return nil
	}
	client, err := clients.NewSheetsClient(cfg.SheetsCredFile, cfg.SheetsID)
	if err != nil {
		log.Printf("WARNING: sheets client not available: %v", err)
		return nil
	}
	return client
}

func newWhatsAppClient(cfg config.Config) *clients.WhatsAppClient {
	if cfg.WhatsAppPhoneID == "" || cfg.WhatsAppToken == "" {
		return nil
	}
	log.Println("WhatsApp client configured")
	return clients.NewWhatsAppClient(cfg.WhatsAppPhoneID, cfg.WhatsAppToken)
}

func newCalendarClient(cfg config.Config) *clients.CalendarClient {
	if cfg.GoogleCalendarID == "" || cfg.SheetsCredFile == "" {
		return nil
	}
	client, err := clients.NewCalendarClient(cfg.SheetsCredFile, cfg.GoogleCalendarID)
	if err != nil {
		log.Printf("WARNING: calendar client not available: %v", err)
		return nil
	}
	return client
}

func newNotionClient(cfg config.Config) *clients.NotionClient {
	if cfg.NotionAPIKey == "" {
		return nil
	}
	log.Println("Notion client configured")
	return clients.NewNotionClient(cfg.NotionAPIKey)
}

func newObsidianVault(cfg config.Config) *clients.ObsidianVault {
	if cfg.ObsidianVaultPath == "" {
		return nil
	}
	log.Printf("Obsidian vault: %s", cfg.ObsidianVaultPath)
	return clients.NewObsidianVault(cfg.ObsidianVaultPath)
}

func newGitHubClient(cfg config.Config) *clients.GitHubClient {
	if cfg.GitHubToken == "" {
		return nil
	}
	log.Println("GitHub client configured")
	return clients.NewGitHubClient(cfg.GitHubToken)
}

func newJiraClient(cfg config.Config) *clients.JiraClient {
	if cfg.JiraBaseURL == "" || cfg.JiraEmail == "" || cfg.JiraAPIToken == "" {
		return nil
	}
	log.Println("Jira client configured")
	return clients.NewJiraClient(cfg.JiraBaseURL, cfg.JiraEmail, cfg.JiraAPIToken)
}

func newSpotifyClient(cfg config.Config) *clients.SpotifyClient {
	if cfg.SpotifyAccessToken == "" {
		return nil
	}
	log.Println("Spotify client configured")
	return clients.NewSpotifyClient(cfg.SpotifyAccessToken)
}

func newTodoistClient(cfg config.Config) *clients.TodoistClient {
	if cfg.TodoistAPIToken == "" {
		return nil
	}
	log.Println("Todoist client configured")
	return clients.NewTodoistClient(cfg.TodoistAPIToken)
}

func newGmailClient(cfg config.Config) *clients.GmailClient {
	if cfg.GmailUserEmail == "" || cfg.SheetsCredFile == "" {
		return nil
	}
	client, err := clients.NewGmailClient(cfg.SheetsCredFile, cfg.GmailUserEmail)
	if err != nil {
		log.Printf("WARNING: gmail client not available: %v", err)
		return nil
	}
	log.Println("Gmail client configured")
	return client
}

func newClickUpClient(cfg config.Config) *clients.ClickUpClient {
	if cfg.ClickUpAPIToken == "" {
		return nil
	}
	log.Println("ClickUp client configured")
	return clients.NewClickUpClient(cfg.ClickUpAPIToken, cfg.ClickUpTeamID)
}

func newTranscriber(cfg config.Config) domain.Transcriber {
	if cfg.OpenAIAPIKey == "" {
		return nil
	}
	log.Println("Transcriber configured (Whisper)")
	return clients.NewOpenAIClient(cfg.OpenAIAPIKey, "")
}

func newTelegramClient(cfg config.Config) *clients.TelegramClient {
	if cfg.TelegramBotToken == "" {
		return nil
	}
	log.Println("Telegram client configured")
	return clients.NewTelegramClient(cfg.TelegramBotToken)
}

func newFigmaClient(cfg config.Config) *clients.FigmaClient {
	if cfg.FigmaAccessToken == "" {
		return nil
	}
	log.Println("Figma client configured")
	return clients.NewFigmaClient(cfg.FigmaAccessToken)
}
