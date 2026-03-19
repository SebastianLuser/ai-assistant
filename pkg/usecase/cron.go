package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"asistente/clients"
	"asistente/internal/hooks"
	"asistente/pkg/domain"
	"asistente/pkg/service"
)

const schedulerTickInterval = 1 * time.Minute

type Scheduler struct {
	jobs     []domain.Job
	hooks    *hooks.Registry
	mu       sync.Mutex
	stopCh   chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

func NewScheduler(jobs []domain.Job, hooksRegistry *hooks.Registry) *Scheduler {
	return &Scheduler{
		jobs:   jobs,
		hooks:  hooksRegistry,
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go func() {
		defer close(s.done)
		ticker := time.NewTicker(schedulerTickInterval)
		defer ticker.Stop()

		log.Printf("cron: scheduler started with %d jobs", len(s.jobs))

		for {
			select {
			case now := <-ticker.C:
				s.mu.Lock()
				for i := range s.jobs {
					if s.jobs[i].ShouldRun(now) {
						go s.runJob(&s.jobs[i], now)
					}
				}
				s.mu.Unlock()
			case <-s.stopCh:
				log.Println("cron: scheduler stopped")
				return
			}
		}
	}()
}

// TriggerJob manually executes a job by ID, regardless of schedule.
func (s *Scheduler) TriggerJob(jobID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.jobs {
		if s.jobs[i].ID == jobID {
			log.Printf("cron: manual trigger for job %s", jobID)
			result, err := s.jobs[i].Execute()
			if err != nil {
				return "", err
			}
			if err := s.jobs[i].Deliver(result); err != nil {
				return result, err
			}
			s.jobs[i].LastRun = time.Now()

			s.hooks.Emit(context.Background(), hooks.CronJobCompleted, map[string]string{
				"job_id": jobID, "result": result, "trigger": "manual",
			})

			return result, nil
		}
	}

	return "", fmt.Errorf("job not found: %s", jobID)
}

// ListJobs returns the IDs of all registered jobs.
func (s *Scheduler) ListJobs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]string, len(s.jobs))
	for i, j := range s.jobs {
		ids[i] = j.ID
	}
	return ids
}

func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	<-s.done
}

func (s *Scheduler) runJob(job *domain.Job, now time.Time) {
	log.Printf("cron: running job %s", job.ID)

	result, err := job.Execute()
	if err != nil {
		log.Printf("cron: job %s failed: %v", job.ID, err)
		return
	}

	if err := job.Deliver(result); err != nil {
		log.Printf("cron: job %s delivery failed: %v", job.ID, err)
		return
	}

	s.mu.Lock()
	job.LastRun = now
	s.mu.Unlock()

	log.Printf("cron: job %s completed", job.ID)

	s.hooks.Emit(context.Background(), hooks.CronJobCompleted, map[string]string{
		"job_id": job.ID, "result": result,
	})
}

func NewDailyBriefingJob(ai domain.AIProvider, waNumber string, waClient *clients.WhatsAppClient, calendarClient *clients.CalendarClient, gmailClient *clients.GmailClient, memorySvc service.MemoryService) domain.Job {
	return domain.Job{
		ID:     domain.JobDailyBriefing,
		Hour:   domain.DailyBriefingHour,
		Minute: 0,
		Delivery: domain.DeliveryConfig{
			Mode:     domain.DeliveryModeWhatsApp,
			To:       waNumber,
			WAClient: waClient,
		},
		AI: ai,
		RunFn: func() (string, error) {
			var parts []string
			parts = append(parts, "Fecha de hoy: "+time.Now().Format(domain.DateFormatCronDisplay))

			if calendarClient != nil {
				events, err := calendarClient.GetTodayEvents()
				if err == nil && len(events) > 0 {
					var evStr []string
					for _, e := range events {
						evStr = append(evStr, fmt.Sprintf("- %s (%s)", e.Summary, e.Start.Format(domain.TimeFormatCronDisplay)))
					}
					parts = append(parts, "Eventos de hoy:\n"+strings.Join(evStr, "\n"))
				}
			}

			if gmailClient != nil {
				emails, err := gmailClient.ListUnread(5)
				if err == nil && len(emails) > 0 {
					var emStr []string
					for _, e := range emails {
						emStr = append(emStr, fmt.Sprintf("- %s (de %s)", e.Subject, e.From))
					}
					parts = append(parts, "Emails sin leer:\n"+strings.Join(emStr, "\n"))
				}
			}

			if memorySvc != nil {
				today := time.Now().Format(domain.DateFormatISO)
				expenses, err := memorySvc.ListExpenses(today, today)
				if err == nil && len(expenses) > 0 {
					total := 0.0
					for _, e := range expenses {
						total += e.Amount
					}
					parts = append(parts, fmt.Sprintf("Gastos de hoy: $%.0f (%d gastos)", total, len(expenses)))
				}
			}

			prompt := strings.Join(parts, "\n\n") + "\n\nArma un briefing matutino breve con esta info. Inclui un saludo motivador."
			return ai.Complete(domain.BriefingSystemPrompt, prompt)
		},
	}
}

func NewWeeklyFinanceJob(ai domain.AIProvider, waNumber string, waClient *clients.WhatsAppClient) domain.Job {
	sunday := time.Sunday
	return domain.Job{
		ID:      domain.JobWeeklyFinance,
		Hour:    domain.WeeklyFinanceHour,
		Minute:  0,
		Weekday: &sunday,
		Prompt:  "Genera un recordatorio para revisar los gastos de la semana.",
		Delivery: domain.DeliveryConfig{
			Mode:     domain.DeliveryModeWhatsApp,
			To:       waNumber,
			WAClient: waClient,
		},
		AI: ai,
	}
}

func NewBudgetAlertJob(ai domain.AIProvider, waNumber string, waClient *clients.WhatsAppClient) domain.Job {
	return domain.Job{
		ID:     domain.JobBudgetAlert,
		Hour:   domain.BudgetAlertHour,
		Minute: 0,
		Delivery: domain.DeliveryConfig{
			Mode:     domain.DeliveryModeWhatsApp,
			To:       waNumber,
			WAClient: waClient,
		},
		AI: ai,
		RunFn: func() (string, error) {
			prompt := fmt.Sprintf(
				"Fecha: %s\nGenerá un mensaje breve de alerta de presupuesto diario. "+
					"Recordale al usuario revisar si sus gastos del día están dentro de lo razonable. "+
					"Sé conciso y motivador.",
				time.Now().Format(domain.DateFormatCronDisplay),
			)
			return ai.Complete(domain.CronSystemPrompt, prompt)
		},
	}
}

func NewDailyJournalJob(ai domain.AIProvider, waNumber string, waClient *clients.WhatsAppClient) domain.Job {
	return domain.Job{
		ID:     domain.JobDailyJournal,
		Hour:   domain.DailyJournalHour,
		Minute: 0,
		Delivery: domain.DeliveryConfig{
			Mode:     domain.DeliveryModeWhatsApp,
			To:       waNumber,
			WAClient: waClient,
		},
		RunFn: func() (string, error) {
			return "¿Cómo fue tu día? Contame qué pasó, qué lograste, o qué te gustaría recordar de hoy.", nil
		},
	}
}

// NewSessionPruningJob deletes conversation sessions older than SessionTTLDays.
// Runs daily at 3am. Delivery is log-only (no notification).
func NewSessionPruningJob(memorySvc service.MemoryService) domain.Job {
	return domain.Job{
		ID:       domain.JobSessionPruning,
		Hour:     domain.SessionPruningHour,
		Minute:   0,
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeLog},
		RunFn: func() (string, error) {
			pruned, err := memorySvc.PruneSessions(domain.SessionTTLDays)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("pruned %d stale session(s) older than %d days", pruned, domain.SessionTTLDays), nil
		},
	}
}
