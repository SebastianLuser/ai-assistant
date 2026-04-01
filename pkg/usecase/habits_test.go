package usecase_test

import (
	"errors"
	"testing"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errHabitFailed = errors.New("habit store error")

func TestHabitUseCase_LogHabit_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("LogHabit", "exercise").Return(nil)
	uc := usecase.NewHabitUseCase(repo)

	err := uc.LogHabit("exercise")

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHabitUseCase_LogHabit_Error(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("LogHabit", "exercise").Return(errHabitFailed)
	uc := usecase.NewHabitUseCase(repo)

	err := uc.LogHabit("exercise")

	require.Error(t, err)
	assert.True(t, errors.Is(err, errHabitFailed))
}

func TestHabitUseCase_GetStreak_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("GetHabitStreak", "exercise").Return(5, 30, nil)
	uc := usecase.NewHabitUseCase(repo)

	streak, total, err := uc.GetStreak("exercise")

	require.NoError(t, err)
	assert.Equal(t, 5, streak)
	assert.Equal(t, 30, total)
	repo.AssertExpectations(t)
}

func TestHabitUseCase_GetStreak_Error(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("GetHabitStreak", "exercise").Return(0, 0, errHabitFailed)
	uc := usecase.NewHabitUseCase(repo)

	_, _, err := uc.GetStreak("exercise")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrHabitQuery))
}

func TestHabitUseCase_ListToday_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("ListHabitsToday").Return([]string{"exercise", "read"}, nil)
	uc := usecase.NewHabitUseCase(repo)

	habits, err := uc.ListToday()

	require.NoError(t, err)
	assert.Equal(t, []string{"exercise", "read"}, habits)
	repo.AssertExpectations(t)
}

func TestHabitUseCase_ListToday_NilReturnsEmptySlice(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("ListHabitsToday").Return([]string(nil), nil)
	uc := usecase.NewHabitUseCase(repo)

	habits, err := uc.ListToday()

	require.NoError(t, err)
	assert.Equal(t, []string{}, habits)
}

func TestHabitUseCase_ListToday_Error(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("ListHabitsToday").Return([]string(nil), errHabitFailed)
	uc := usecase.NewHabitUseCase(repo)

	_, err := uc.ListToday()

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrHabitQuery))
}
