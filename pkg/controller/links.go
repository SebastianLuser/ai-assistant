package controller

import (
	"net/http"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/web"
)

type LinkController struct {
	usecase *usecase.LinkUseCase
}

func NewLinkController(uc *usecase.LinkUseCase) *LinkController {
	return &LinkController{usecase: uc}
}

func (c *LinkController) PostLink(req web.Request) web.Response {
	var payload domain.LinkSaveRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.LinkResponse{Error: "invalid body"})
	}

	if err := payload.Validate(); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.LinkResponse{Error: err.Error()})
	}

	if payload.Title == "" {
		payload.Title = payload.URL
	}

	id, err := c.usecase.Save(payload.URL, payload.Title, payload.Tags)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.LinkResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.LinkResponse{Success: true, ID: id, Message: "Link guardado"})
}

func (c *LinkController) GetSearch(req web.Request) web.Response {
	q, ok := req.Query(domain.QueryParamQ)
	if !ok || q == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.LinkListResponse{Error: "query parameter 'q' is required"})
	}

	results, err := c.usecase.Search(q, domain.DefaultSearchLimit)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.LinkListResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.LinkListResponse{Success: true, Results: results})
}
