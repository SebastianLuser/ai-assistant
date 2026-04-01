package controller

import (
	"net/http"
	"strconv"

	"jarvis/pkg/domain"
	"jarvis/pkg/service"
	"jarvis/pkg/usecase"
	"jarvis/web"
)

type MemoryController struct {
	store    service.MemoryService
	embedder service.Embedder
	search   *usecase.MemoryUseCase
}

func NewMemoryController(store service.MemoryService, embedder service.Embedder) *MemoryController {
	return &MemoryController{
		store:    store,
		embedder: embedder,
		search:   usecase.NewMemoryUseCase(store, embedder),
	}
}

func (c *MemoryController) PostNote(req web.Request) web.Response {
	var payload domain.NoteRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.NoteResponse{Error: "invalid body"})
	}

	if err := payload.Validate(); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.NoteResponse{Error: err.Error()})
	}

	embedding, err := c.embedder.Embed(payload.Content)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.NoteResponse{Error: "failed to generate embedding"})
	}

	id, err := c.store.Save(payload.Content, payload.Tags, embedding)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.NoteResponse{Error: "failed to save note"})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.NoteResponse{Success: true, ID: id, Message: "Nota guardada"})
}

func (c *MemoryController) GetSearch(req web.Request) web.Response {
	q, ok := req.Query(domain.QueryParamQ)
	if !ok || q == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.NoteResponse{Error: "query parameter 'q' is required"})
	}

	limit := parseLimit(req)
	mode, _ := req.Query(domain.QueryParamMode)

	results, err := c.executeSearch(q, mode, limit)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.NoteResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.SearchResponse{Results: results})
}

func parseLimit(req web.Request) int {
	limitStr, exists := req.Query(domain.QueryParamLimit)
	if !exists || limitStr == "" {
		return domain.DefaultSearchLimit
	}

	n, err := strconv.Atoi(limitStr)
	if err != nil || n <= 0 {
		return domain.DefaultSearchLimit
	}

	return n
}

func (c *MemoryController) executeSearch(query, mode string, limit int) ([]domain.Memory, error) {
	switch mode {
	case domain.SearchModeFTS:
		return c.store.SearchFTS(query, limit)
	case domain.SearchModeVector:
		return c.searchWithEmbedding(query, limit)
	case domain.SearchModeHybrid:
		return c.searchHybrid(query, limit)
	default:
		return c.search.FallbackSearch(query, limit)
	}
}

func (c *MemoryController) searchWithEmbedding(query string, limit int) ([]domain.Memory, error) {
	embedding, err := c.embedder.Embed(query)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrEmbedGenerate, err)
	}
	return c.store.Search(embedding, limit)
}

func (c *MemoryController) searchHybrid(query string, limit int) ([]domain.Memory, error) {
	embedding, err := c.embedder.Embed(query)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrEmbedGenerate, err)
	}
	return c.store.SearchHybrid(query, embedding, limit, domain.DefaultVecWeight, domain.DefaultFTSWeight)
}

func (c *MemoryController) DeleteNote(req web.Request) web.Response {
	idStr, ok := req.Param(domain.PathParamID)
	if !ok {
		return web.NewJSONResponse(http.StatusBadRequest, domain.NoteResponse{Error: "id is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.NoteResponse{Error: "invalid id"})
	}

	if err := c.store.Delete(id); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.NoteResponse{Error: "failed to delete"})
	}

	return web.NewJSONResponse(http.StatusOK, domain.NoteResponse{Success: true})
}
