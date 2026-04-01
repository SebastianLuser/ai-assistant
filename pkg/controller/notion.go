package controller

import (
	"net/http"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

type NotionController struct {
	client   *clients.NotionClient
	parentID string
}

func NewNotionController(client *clients.NotionClient, defaultParentID string) *NotionController {
	return &NotionController{client: client, parentID: defaultParentID}
}

func (c *NotionController) CreatePage(req web.Request) web.Response {
	var payload domain.NotionCreateRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.NotionResponse{Error: "invalid body"})
	}

	if payload.Title == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.NotionResponse{Error: "title is required"})
	}

	id, err := c.client.CreatePage(c.parentID, payload.Title, payload.Content)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.NotionResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.NotionResponse{Success: true, ID: id})
}

func (c *NotionController) GetPage(req web.Request) web.Response {
	pageID, ok := req.Param(domain.PathParamID)
	if !ok || pageID == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.NotionResponse{Error: "id is required"})
	}

	page, err := c.client.GetPage(pageID)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.NotionResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.NotionResponse{
		Success: true, ID: page.ID, Title: page.Title,
	})
}
