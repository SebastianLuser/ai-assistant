package usecase

import (
	"context"
	"testing"
	"time"

	"jarvis/internal/hooks"
	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testJobID = "test-job"
)

func TestJob_ShouldRun_CorrectTime(t *testing.T) {
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)

	assert.True(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongHour(t *testing.T) {
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongMinute(t *testing.T) {
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 10, 8, 1, 0, 0, time.UTC)

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongWeekday(t *testing.T) {
	sunday := time.Sunday
	job := domain.Job{ID: testJobID, Hour: 20, Minute: 0, Weekday: &sunday}
	monday := time.Date(2026, 3, 9, 20, 0, 0, 0, time.UTC) // Monday

	assert.False(t, job.ShouldRun(monday))
}

func TestJob_ShouldRun_CorrectWeekday(t *testing.T) {
	sunday := time.Sunday
	job := domain.Job{ID: testJobID, Hour: 20, Minute: 0, Weekday: &sunday}
	sun := time.Date(2026, 3, 15, 20, 0, 0, 0, time.UTC) // Sunday

	assert.True(t, job.ShouldRun(sun))
}

func TestJob_ShouldRun_TooSoonAfterLastRun(t *testing.T) {
	now := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0, LastRun: now.Add(-30 * time.Second)}

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_EnoughTimeAfterLastRun(t *testing.T) {
	now := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0, LastRun: now.Add(-3 * time.Minute)}

	assert.True(t, job.ShouldRun(now))
}

func TestJob_Execute_WithRunFn(t *testing.T) {
	job := domain.Job{
		ID:    testJobID,
		RunFn: func() (string, error) { return "custom result", nil },
	}

	result, err := job.Execute()

	assert.NoError(t, err)
	assert.Equal(t, "custom result", result)
}

func TestJob_Execute_NoClaude_NoRunFn_ReturnsError(t *testing.T) {
	job := domain.Job{ID: testJobID}

	_, err := job.Execute()

	assert.Error(t, err)
}

func TestJob_Deliver_LogMode(t *testing.T) {
	job := domain.Job{
		ID:       testJobID,
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeLog},
	}

	err := job.Deliver("test result")

	assert.NoError(t, err)
}

func TestJob_Deliver_WhatsApp_NotConfigured_NoError(t *testing.T) {
	job := domain.Job{
		ID:       testJobID,
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeWhatsApp},
	}

	err := job.Deliver("test result")

	assert.NoError(t, err)
}

func TestJob_Deliver_DefaultMode_NoError(t *testing.T) {
	job := domain.Job{
		ID:       testJobID,
		Delivery: domain.DeliveryConfig{Mode: "unknown"},
	}

	err := job.Deliver("test result")

	assert.NoError(t, err)
}

// --- Scheduler tests ---

func TestScheduler_TriggerJob_Success(t *testing.T) {
	job := domain.Job{
		ID:       "trigger-test",
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeLog},
		RunFn:    func() (string, error) { return "triggered", nil },
	}
	hooksRegistry := hooks.NewRegistry()
	s := NewScheduler([]domain.Job{job}, hooksRegistry)

	result, err := s.TriggerJob("trigger-test")

	require.NoError(t, err)
	assert.Equal(t, "triggered", result)
}

func TestScheduler_TriggerJob_NotFound(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	s := NewScheduler([]domain.Job{}, hooksRegistry)

	_, err := s.TriggerJob("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "job not found")
}

func TestScheduler_TriggerJob_ExecuteError(t *testing.T) {
	job := domain.Job{
		ID:       "fail-job",
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeLog},
		RunFn:    func() (string, error) { return "", assert.AnError },
	}
	hooksRegistry := hooks.NewRegistry()
	s := NewScheduler([]domain.Job{job}, hooksRegistry)

	_, err := s.TriggerJob("fail-job")

	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestScheduler_TriggerJob_UpdatesLastRun(t *testing.T) {
	job := domain.Job{
		ID:       "time-job",
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeLog},
		RunFn:    func() (string, error) { return "ok", nil },
	}
	hooksRegistry := hooks.NewRegistry()
	s := NewScheduler([]domain.Job{job}, hooksRegistry)

	assert.True(t, s.jobs[0].LastRun.IsZero())

	_, err := s.TriggerJob("time-job")
	require.NoError(t, err)

	assert.False(t, s.jobs[0].LastRun.IsZero())
}

func TestScheduler_TriggerJob_EmitsHook(t *testing.T) {
	job := domain.Job{
		ID:       "hook-job",
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeLog},
		RunFn:    func() (string, error) { return "result", nil },
	}
	hooksRegistry := hooks.NewRegistry()
	hookCalled := false
	hooksRegistry.Register(hooks.CronJobCompleted, func(ctx context.Context, event hooks.Event) error {
		hookCalled = true
		return nil
	})
	s := NewScheduler([]domain.Job{job}, hooksRegistry)

	_, _ = s.TriggerJob("hook-job")

	assert.True(t, hookCalled)
}

func TestScheduler_ListJobs_ReturnsAllIDs(t *testing.T) {
	jobs := []domain.Job{
		{ID: "job-a"},
		{ID: "job-b"},
		{ID: "job-c"},
	}
	hooksRegistry := hooks.NewRegistry()
	s := NewScheduler(jobs, hooksRegistry)

	ids := s.ListJobs()

	assert.Equal(t, []string{"job-a", "job-b", "job-c"}, ids)
}

func TestScheduler_ListJobs_Empty(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	s := NewScheduler([]domain.Job{}, hooksRegistry)

	ids := s.ListJobs()

	assert.Empty(t, ids)
}

// --- NewSessionPruningJob tests ---

func TestNewSessionPruningJob_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("PruneSessions", domain.SessionTTLDays).Return(int64(5), nil)

	job := NewSessionPruningJob(repo)

	assert.Equal(t, domain.JobSessionPruning, job.ID)
	assert.Equal(t, domain.SessionPruningHour, job.Hour)
	assert.Equal(t, domain.DeliveryModeLog, job.Delivery.Mode)

	result, err := job.RunFn()

	require.NoError(t, err)
	assert.Contains(t, result, "pruned 5 stale session(s)")
	repo.AssertExpectations(t)
}

func TestNewSessionPruningJob_Error(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("PruneSessions", domain.SessionTTLDays).Return(int64(0), assert.AnError)

	job := NewSessionPruningJob(repo)

	_, err := job.RunFn()

	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestNewSessionPruningJob_ZeroPruned(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("PruneSessions", domain.SessionTTLDays).Return(int64(0), nil)

	job := NewSessionPruningJob(repo)

	result, err := job.RunFn()

	require.NoError(t, err)
	assert.Contains(t, result, "pruned 0 stale session(s)")
}

// --- NewDailyBriefingJob tests ---

func TestNewDailyBriefingJob_Structure(t *testing.T) {
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "briefing"})
	defer srv.Close()

	job := NewDailyBriefingJob(ai, "123", nil, nil, nil, nil)

	assert.Equal(t, domain.JobDailyBriefing, job.ID)
	assert.Equal(t, domain.DailyBriefingHour, job.Hour)
	assert.Equal(t, domain.DeliveryModeWhatsApp, job.Delivery.Mode)
	assert.Equal(t, "123", job.Delivery.To)
	assert.NotNil(t, job.RunFn)
}

func TestNewDailyBriefingJob_RunFn_NoClients(t *testing.T) {
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "good morning"})
	defer srv.Close()

	job := NewDailyBriefingJob(ai, "123", nil, nil, nil, nil)

	result, err := job.RunFn()

	require.NoError(t, err)
	assert.Equal(t, "good morning", result)
}

func TestNewDailyBriefingJob_RunFn_WithMemoryService(t *testing.T) {
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "briefing with expenses"})
	defer srv.Close()

	repo := new(test.MockMemoryService)
	expenses := []domain.Expense{
		{Amount: 1000, Category: "Supermercado"},
	}
	repo.On("ListExpenses", mock.Anything, mock.Anything).Return(expenses, nil)

	job := NewDailyBriefingJob(ai, "123", nil, nil, nil, repo)

	result, err := job.RunFn()

	require.NoError(t, err)
	assert.Equal(t, "briefing with expenses", result)
	repo.AssertExpectations(t)
}

// --- NewWeeklyFinanceJob tests ---

func TestNewWeeklyFinanceJob_Structure(t *testing.T) {
	job := NewWeeklyFinanceJob(nil, "123", nil)

	assert.Equal(t, domain.JobWeeklyFinance, job.ID)
	assert.Equal(t, domain.WeeklyFinanceHour, job.Hour)
	assert.NotNil(t, job.Weekday)
	assert.Equal(t, time.Sunday, *job.Weekday)
	assert.NotEmpty(t, job.Prompt)
	assert.Equal(t, domain.DeliveryModeWhatsApp, job.Delivery.Mode)
}

// --- NewBudgetAlertJob tests ---

func TestNewBudgetAlertJob_Structure(t *testing.T) {
	job := NewBudgetAlertJob(nil, "123", nil)

	assert.Equal(t, domain.JobBudgetAlert, job.ID)
	assert.Equal(t, domain.BudgetAlertHour, job.Hour)
	assert.NotNil(t, job.RunFn)
	assert.Equal(t, domain.DeliveryModeWhatsApp, job.Delivery.Mode)
}

func TestNewBudgetAlertJob_RunFn_Success(t *testing.T) {
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "budget alert"})
	defer srv.Close()

	job := NewBudgetAlertJob(ai, "123", nil)

	result, err := job.RunFn()

	require.NoError(t, err)
	assert.Equal(t, "budget alert", result)
}

// --- Scheduler Start/Stop tests ---

func TestScheduler_StartStop(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	s := NewScheduler([]domain.Job{}, hooksRegistry)

	s.Start()
	s.Stop()
	// If we get here without hanging, the test passes
}

func TestScheduler_StopIdempotent(t *testing.T) {
	hooksRegistry := hooks.NewRegistry()
	s := NewScheduler([]domain.Job{}, hooksRegistry)

	s.Start()
	s.Stop()
	// Calling Stop again should not panic due to stopOnce
	s.Stop()
}

func TestNewDailyJournalJob_ReturnsStaticMessage(t *testing.T) {
	job := NewDailyJournalJob(nil, "123", nil)

	assert.Equal(t, domain.JobDailyJournal, job.ID)
	assert.Equal(t, domain.DailyJournalHour, job.Hour)

	result, err := job.RunFn()

	require.NoError(t, err)
	assert.Contains(t, result, "Contame")
}
