package controller

import (
	"net/http"

	"jarvis/pkg/service"
	"jarvis/web"
)

type CatalogController struct {
	catalog service.CatalogService
}

func NewCatalogController(catalog service.CatalogService) *CatalogController {
	return &CatalogController{catalog: catalog}
}

func (c *CatalogController) List(req web.Request) web.Response {
	entries, err := c.catalog.GetAll()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, map[string]any{
			"success": false, "error": err.Error(),
		})
	}

	type entryResponse struct {
		Name        string  `json:"name"`
		Type        string  `json:"type"`
		UsageCount  int64   `json:"usage_count"`
		LastUsed    string  `json:"last_used"`
		SuccessRate float64 `json:"success_rate"`
		ErrorCount  int64   `json:"error_count"`
	}

	var result []entryResponse
	for _, e := range entries {
		result = append(result, entryResponse{
			Name:        e.Name,
			Type:        e.Type,
			UsageCount:  e.UsageCount,
			LastUsed:    service.LastUsedFormatted(e.LastUsed),
			SuccessRate: e.SuccessRate(),
			ErrorCount:  e.ErrorCount,
		})
	}

	return web.NewJSONResponse(http.StatusOK, map[string]any{
		"success": true,
		"entries": result,
	})
}

func (c *CatalogController) Get(req web.Request) web.Response {
	name, _ := req.Param("name")
	entryType, _ := req.Query("type")
	if entryType == "" {
		entryType = "tool"
	}

	entry, err := c.catalog.GetByName(name, entryType)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, map[string]any{
			"success": false, "error": err.Error(),
		})
	}
	if entry == nil {
		return web.NewJSONResponse(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "entry not found",
		})
	}

	return web.NewJSONResponse(http.StatusOK, map[string]any{
		"success": true,
		"entry": map[string]any{
			"name":         entry.Name,
			"type":         entry.Type,
			"usage_count":  entry.UsageCount,
			"last_used":    service.LastUsedFormatted(entry.LastUsed),
			"success_rate": entry.SuccessRate(),
			"error_count":  entry.ErrorCount,
			"tags":         entry.Tags,
			"created_at":   entry.CreatedAt,
		},
	})
}
