package controller

import (
	"net/http"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

type CalendarController struct {
	client *clients.CalendarClient
}

func NewCalendarController(client *clients.CalendarClient) *CalendarController {
	return &CalendarController{client: client}
}

func (c *CalendarController) GetTodayEvents(req web.Request) web.Response {
	events, err := c.client.GetTodayEvents()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.CalendarListResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.CalendarListResponse{
		Success: true, Events: toCalendarEvents(events),
	})
}

func (c *CalendarController) CreateEvent(req web.Request) web.Response {
	var payload domain.CalendarEventRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.CalendarEventResponse{Error: "invalid body"})
	}

	start, end, err := payload.Validate()
	if err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.CalendarEventResponse{Error: err.Error()})
	}

	id, err := c.client.CreateEvent(payload.Summary, start, end)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.CalendarEventResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.CalendarEventResponse{Success: true, ID: id})
}

func toCalendarEvents(events []clients.CalendarEvent) []domain.CalendarEvent {
	result := make([]domain.CalendarEvent, len(events))
	for i, e := range events {
		result[i] = domain.CalendarEvent{
			ID: e.ID, Summary: e.Summary, Start: e.Start, End: e.End, Location: e.Location,
		}
	}
	return result
}
