package controller

import (
	"net/http"
	"testing"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
)

const (
	validHabitBody   = `{"name":"exercise"}`
	emptyHabitBody   = `{"name":""}`
	invalidHabitJSON = `{nope`
)

func TestHabitController_PostLog_InvalidJSON(t *testing.T) {
	ctrl := NewHabitController(nil)
	req := test.NewMockRequest().WithBody(invalidHabitJSON)

	resp := ctrl.PostLog(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestHabitController_PostLog_EmptyName(t *testing.T) {
	ctrl := NewHabitController(nil)
	req := test.NewMockRequest().WithBody(emptyHabitBody)

	resp := ctrl.PostLog(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestHabitController_PostLog_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("LogHabit", "exercise").Return(nil)
	uc := usecase.NewHabitUseCase(repo)
	ctrl := NewHabitController(uc)
	req := test.NewMockRequest().WithBody(validHabitBody)

	resp := ctrl.PostLog(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
	repo.AssertExpectations(t)
}

func TestHabitController_GetStreak_MissingName(t *testing.T) {
	ctrl := NewHabitController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetStreak(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestHabitController_GetStreak_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("GetHabitStreak", "exercise").Return(5, 30, nil)
	uc := usecase.NewHabitUseCase(repo)
	ctrl := NewHabitController(uc)
	req := test.NewMockRequest().WithQuery("name", "exercise")

	resp := ctrl.GetStreak(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	repo.AssertExpectations(t)
}

func TestHabitController_GetToday_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("ListHabitsToday").Return([]string{"exercise"}, nil)
	uc := usecase.NewHabitUseCase(repo)
	ctrl := NewHabitController(uc)
	req := test.NewMockRequest()

	resp := ctrl.GetToday(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	repo.AssertExpectations(t)
}

func TestHabitController_PostLog_StoreError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("LogHabit", "exercise").Return(domain.ErrStoreOpen)
	uc := usecase.NewHabitUseCase(repo)
	ctrl := NewHabitController(uc)
	req := test.NewMockRequest().WithBody(validHabitBody)

	resp := ctrl.PostLog(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestHabitController_GetStreak_StoreError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("GetHabitStreak", "exercise").Return(0, 0, domain.ErrStoreOpen)
	uc := usecase.NewHabitUseCase(repo)
	ctrl := NewHabitController(uc)
	req := test.NewMockRequest().WithQuery("name", "exercise")

	resp := ctrl.GetStreak(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestHabitController_GetToday_StoreError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("ListHabitsToday").Return([]string(nil), domain.ErrStoreOpen)
	uc := usecase.NewHabitUseCase(repo)
	ctrl := NewHabitController(uc)
	req := test.NewMockRequest()

	resp := ctrl.GetToday(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}
