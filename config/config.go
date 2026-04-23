package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port               string
	AIProvider         string
	ClaudeAPIKey       string
	ClaudeModel        string
	ClaudeModelLight   string
	OpenAIAPIKey       string
	OpenAIModel        string
	OpenAIModelLight   string
	CompactThreshold   int
	MaxHistoryMsgs     int
	SheetsID           string
	SheetsCredFile     string
	SheetsSheetName    string
	WebhookSecret      string
	WhatsAppPhoneID    string
	WhatsAppToken      string
	WhatsAppTo         string
	WhatsAppVerifyToken string
	WhatsAppAppSecret  string
	SkillsDir          string
	NotionAPIKey       string
	NotionPageID       string
	ObsidianVaultPath  string
	GoogleCalendarID   string
	PostgresDSN        string
	GitHubToken        string
	JiraBaseURL        string
	JiraEmail          string
	JiraAPIToken       string
	SpotifyAccessToken string
	TodoistAPIToken    string
	GmailUserEmail     string
	ClickUpAPIToken    string
	ClickUpTeamID      string
	FigmaAccessToken   string
	TelegramBotToken    string
	TelegramSecretToken string
	TelegramBotUsername string
	RulesDir            string
	DryRunTools         string
	ProfilesDir         string
	DefaultProfile      string
	AgentsDir           string
	HooksConfigFile     string
}

func Load() Config {
	return Config{
		Port:               envOr("PORT", "8080"),
		AIProvider:         envOr("AI_PROVIDER", "claude"),
		ClaudeAPIKey:       os.Getenv("CLAUDE_API_KEY"),
		ClaudeModel:        envOr("CLAUDE_MODEL", "claude-sonnet-4-6"),
		ClaudeModelLight:   envOr("CLAUDE_MODEL_LIGHT", ""),
		OpenAIAPIKey:       os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:        envOr("OPENAI_MODEL", "gpt-4o"),
		OpenAIModelLight:   envOr("OPENAI_MODEL_LIGHT", ""),
		CompactThreshold:   envOrInt("COMPACT_THRESHOLD", 20),
		MaxHistoryMsgs:     envOrInt("MAX_HISTORY_MESSAGES", 30),
		SheetsID:           os.Getenv("GOOGLE_SHEETS_ID"),
		SheetsCredFile:     envOr("GOOGLE_CREDENTIALS_FILE", "credentials.json"),
		SheetsSheetName:    envOr("GOOGLE_SHEETS_NAME", "Gastos"),
		WebhookSecret:      os.Getenv("WEBHOOK_SECRET"),
		WhatsAppPhoneID:    os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
		WhatsAppToken:      os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		WhatsAppTo:         os.Getenv("WHATSAPP_TO_NUMBER"),
		WhatsAppVerifyToken: os.Getenv("WHATSAPP_VERIFY_TOKEN"),
		WhatsAppAppSecret:  os.Getenv("WHATSAPP_APP_SECRET"),
		SkillsDir:          envOr("SKILLS_DIR", "skills"),
		NotionAPIKey:       os.Getenv("NOTION_API_KEY"),
		NotionPageID:       os.Getenv("NOTION_DEFAULT_PAGE_ID"),
		ObsidianVaultPath:  os.Getenv("OBSIDIAN_VAULT_PATH"),
		GoogleCalendarID:   os.Getenv("GOOGLE_CALENDAR_ID"),
		PostgresDSN:        os.Getenv("POSTGRES_DSN"),
		GitHubToken:        os.Getenv("GITHUB_TOKEN"),
		JiraBaseURL:        os.Getenv("JIRA_BASE_URL"),
		JiraEmail:          os.Getenv("JIRA_EMAIL"),
		JiraAPIToken:       os.Getenv("JIRA_API_TOKEN"),
		SpotifyAccessToken: os.Getenv("SPOTIFY_ACCESS_TOKEN"),
		TodoistAPIToken:    os.Getenv("TODOIST_API_TOKEN"),
		GmailUserEmail:     os.Getenv("GMAIL_USER_EMAIL"),
		ClickUpAPIToken:    os.Getenv("CLICKUP_API_TOKEN"),
		ClickUpTeamID:      os.Getenv("CLICKUP_TEAM_ID"),
		FigmaAccessToken:    os.Getenv("FIGMA_ACCESS_TOKEN"),
		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramSecretToken: os.Getenv("TELEGRAM_SECRET_TOKEN"),
		TelegramBotUsername: os.Getenv("TELEGRAM_BOT_USERNAME"),
		RulesDir:            envOr("RULES_DIR", "rules"),
		DryRunTools:         os.Getenv("DRY_RUN_TOOLS"),
		ProfilesDir:         envOr("PROFILES_DIR", "config/profiles"),
		DefaultProfile:      envOr("DEFAULT_PROFILE", "full"),
		AgentsDir:           envOr("AGENTS_DIR", "agents"),
		HooksConfigFile:     envOr("HOOKS_CONFIG_FILE", "config/hooks.yaml"),
	}
}

func envOrInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
