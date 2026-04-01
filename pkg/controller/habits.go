package controller

import (
	"net/http"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/web"
)

type HabitController struct {
	usecase *usecase.HabitUseCase
}

func NewHabitController(uc *usecase.HabitUseCase) *HabitController {
	return &HabitController{usecase: uc}
}

func (c *HabitController) PostLog(req web.Request) web.Response {
	var payload domain.HabitLogRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.HabitLogResponse{Error: "invalid body"})
	}

	if payload.Name == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.HabitLogResponse{Error: "name is required"})
	}

	if err := c.usecase.LogHabit(payload.Name); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.HabitLogResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.HabitLogResponse{Success: true, Message: "Hábito registrado"})
}

func (c *HabitController) GetStreak(req web.Request) web.Response {
	name, ok := req.Query(domain.QueryParamName)
	if !ok || name == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.HabitStreakResponse{Error: "query parameter 'name' is required"})
	}

	streak, total, err := c.usecase.GetStreak(name)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.HabitStreakResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.HabitStreakResponse{
		Success:       true,
		Name:          name,
		CurrentStreak: streak,
		Total:         total,
	})
}

func (c *HabitController) GetToday(req web.Request) web.Response {
	habits, err := c.usecase.ListToday()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.HabitsTodayResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.HabitsTodayResponse{Success: true, Habits: habits})
}
