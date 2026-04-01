package controller

import (
	"net/http"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

type ObsidianController struct {
	vault *clients.ObsidianVault
}

func NewObsidianController(vault *clients.ObsidianVault) *ObsidianController {
	return &ObsidianController{vault: vault}
}

func (c *ObsidianController) ReadNote(req web.Request) web.Response {
	path, ok := req.Query(domain.QueryParamPath)
	if !ok {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ObsidianNoteResponse{Error: "path is required"})
	}

	if err := domain.ValidatePath(path); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ObsidianNoteResponse{Error: err.Error()})
	}

	content, err := c.vault.ReadNote(path)
	if err != nil {
		return web.NewJSONResponse(http.StatusNotFound, domain.ObsidianNoteResponse{Error: "note not found"})
	}

	return web.NewJSONResponse(http.StatusOK, domain.ObsidianNoteResponse{
		Success: true, Path: path, Content: content,
	})
}

func (c *ObsidianController) WriteNote(req web.Request) web.Response {
	var payload domain.ObsidianNoteRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ObsidianNoteResponse{Error: "invalid body"})
	}

	if err := payload.Validate(); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ObsidianNoteResponse{Error: err.Error()})
	}

	if err := c.vault.WriteNote(payload.Path, payload.Content); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ObsidianNoteResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.ObsidianNoteResponse{Success: true, Path: payload.Path})
}

func (c *ObsidianController) ListNotes(req web.Request) web.Response {
	dir, _ := req.Query(domain.QueryParamDir)

	notes, err := c.vault.ListNotes(dir)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ObsidianListResponse{Error: err.Error()})
	}

	result := make([]domain.ObsidianNote, len(notes))
	for i, n := range notes {
		result[i] = domain.ObsidianNote{Path: n.Path, Title: n.Title, ModifiedAt: n.ModifiedAt}
	}

	return web.NewJSONResponse(http.StatusOK, domain.ObsidianListResponse{Success: true, Notes: result})
}

func (c *ObsidianController) SearchNotes(req web.Request) web.Response {
	q, ok := req.Query(domain.QueryParamQ)
	if !ok || q == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ObsidianListResponse{Error: "query parameter 'q' is required"})
	}

	notes, err := c.vault.SearchNotes(q)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ObsidianListResponse{Error: err.Error()})
	}

	result := make([]domain.ObsidianNote, len(notes))
	for i, n := range notes {
		result[i] = domain.ObsidianNote{Path: n.Path, Title: n.Title, ModifiedAt: n.ModifiedAt}
	}

	return web.NewJSONResponse(http.StatusOK, domain.ObsidianListResponse{Success: true, Notes: result})
}
