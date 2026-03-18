package domain

import (
	"log"
	"time"
)

const (
	DeliveryModeWhatsApp = "whatsapp"
	DeliveryModeLog      = "log"
	DeliveryModeWebhook  = "webhook"

	JobDailyBriefing  = "daily-briefing"
	JobWeeklyFinance  = "weekly-finance"
	JobBudgetAlert    = "budget-alert"
	JobDailyJournal   = "daily-journal"
	JobSessionPruning = "session-pruning"

	DailyBriefingHour  = 8
	WeeklyFinanceHour  = 20
	BudgetAlertHour    = 21
	DailyJournalHour   = 22
	SessionPruningHour = 3

	SessionTTLDays = 30
	MinRunInterval    = 2 * time.Minute

	CronSystemPrompt     = "Sos un asistente personal. Genera contenido conciso y util."
	BriefingSystemPrompt = "Sos un asistente personal. Hablas en español rioplatense."

	DateFormatCronDisplay = "Monday 02/01/2006"
	TimeFormatCronDisplay = "15:04"
)

// WhatsAppSender is an interface for sending WhatsApp messages.
type WhatsAppSender interface {
	SendTextMessage(to, text string) error
}

// Channel represents a messaging channel (WhatsApp, Telegram, CLI, etc).
type Channel interface {
	// Name returns the channel identifier (e.g. "whatsapp", "telegram", "cli").
	Name() string
	// SendMessage sends a text reply to the given recipient.
	SendMessage(to, text string) error
	// AckMessage acknowledges receipt (e.g. read receipts). Optional — noop if unsupported.
	AckMessage(messageID string) error
}

// MessageMeta carries metadata about an incoming message.
type MessageMeta struct {
	IsGroup  bool   // true if the message comes from a group chat
	BotName  string // the bot's username, for mention detection
}

type DeliveryConfig struct {
	Mode     string
	To       string
	WAClient WhatsAppSender
}

type Job struct {
	ID       string
	Hour     int
	Minute   int
	Weekday  *time.Weekday
	Prompt   string
	Delivery DeliveryConfig
	RunFn    func() (string, error)
	LastRun  time.Time
	AI       AIProvider
}

func (j *Job) ShouldRun(now time.Time) bool {
	if now.Hour() != j.Hour || now.Minute() != j.Minute {
		return false
	}

	if j.Weekday != nil && now.Weekday() != *j.Weekday {
		return false
	}

	if !j.LastRun.IsZero() && now.Sub(j.LastRun) < MinRunInterval {
		return false
	}

	return true
}

func (j *Job) Execute() (string, error) {
	if j.RunFn != nil {
		return j.RunFn()
	}

	if j.AI == nil {
		return "", Wrap(ErrCronNoClaude, j.ID)
	}

	return j.AI.Complete(CronSystemPrompt, j.Prompt)
}

func (j *Job) Deliver(result string) error {
	switch j.Delivery.Mode {
	case DeliveryModeWhatsApp:
		if j.Delivery.WAClient != nil && j.Delivery.To != "" {
			return j.Delivery.WAClient.SendTextMessage(j.Delivery.To, result)
		}
		log.Printf("[cron:%s -> WA not configured] %s", j.ID, result)
		return nil
	case DeliveryModeLog:
		log.Printf("[cron:%s] %s", j.ID, result)
		return nil
	default:
		log.Printf("[cron:%s] %s", j.ID, result)
		return nil
	}
}
