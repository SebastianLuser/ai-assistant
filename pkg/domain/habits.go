package domain

import "time"

const (
	QueryParamName = "name"
)

type Habit struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	LoggedAt time.Time `json:"logged_at"`
}

const (
	maxHabitNameLen = 200
)

type HabitLogRequest struct {
	Name string `json:"name"`
}

// Validate checks that a habit log request is valid.
func (r HabitLogRequest) Validate() error {
	if r.Name == "" {
		return Wrap(ErrValidation, "name is required")
	}
	if len(r.Name) > maxHabitNameLen {
		return Wrap(ErrValidation, "name exceeds maximum length")
	}
	return nil
}

type HabitLogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type HabitStreakResponse struct {
	Success       bool   `json:"success"`
	Name          string `json:"name"`
	CurrentStreak int    `json:"current_streak"`
	Total         int    `json:"total"`
	Error         string `json:"error,omitempty"`
}

type HabitsTodayResponse struct {
	Success bool     `json:"success"`
	Habits  []string `json:"habits"`
	Error   string   `json:"error,omitempty"`
}
