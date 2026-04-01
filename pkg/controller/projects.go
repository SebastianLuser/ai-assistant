package controller

import (
	"net/http"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/web"
)

type ProjectController struct {
	usecase *usecase.ProjectUseCase
}

func NewProjectController(uc *usecase.ProjectUseCase) *ProjectController {
	return &ProjectController{usecase: uc}
}

func (c *ProjectController) GetStatus(req web.Request) web.Response {
	name, ok := req.Param(domain.PathParamName)
	if !ok || name == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ProjectStatusResponse{Error: "project name is required"})
	}

	result, err := c.usecase.GetStatus(name)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ProjectStatusResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, result)
}
