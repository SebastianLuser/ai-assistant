package controller

import (
	"net/http"
	"strconv"
	"strings"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

// FigmaController handles Figma API endpoints.
type FigmaController struct {
	client *clients.FigmaClient
}

// NewFigmaController creates a new FigmaController.
func NewFigmaController(client *clients.FigmaClient) *FigmaController {
	return &FigmaController{client: client}
}

// GetFile returns metadata for a Figma file.
func (c *FigmaController) GetFile(req web.Request) web.Response {
	fileKey, ok := req.Param(domain.PathParamFileKey)
	if !ok || fileKey == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.FigmaFileResponse{Error: "file_key is required"})
	}

	file, err := c.client.GetFile(fileKey)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.FigmaFileResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.FigmaFileResponse{
		Success: true,
		File: &domain.FigmaFile{
			Name:         file.Name,
			LastModified: file.LastModified,
			ThumbnailURL: file.ThumbnailURL,
			Version:      file.Version,
		},
	})
}

// GetNodes returns specific nodes from a Figma file.
func (c *FigmaController) GetNodes(req web.Request) web.Response {
	fileKey, ok := req.Param(domain.PathParamFileKey)
	if !ok || fileKey == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.FigmaNodesResponse{Error: "file_key is required"})
	}

	idsStr, _ := req.Query(domain.QueryParamNodeIDs)
	if idsStr == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.FigmaNodesResponse{Error: "ids query param is required"})
	}

	nodeIDs := strings.Split(idsStr, ",")
	nodes, err := c.client.GetNodes(fileKey, nodeIDs)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.FigmaNodesResponse{Error: err.Error()})
	}

	domainNodes := make(map[string]domain.FigmaNodeDetail, len(nodes))
	for id, n := range nodes {
		components := make(map[string]domain.FigmaComponent, len(n.Components))
		for cid, comp := range n.Components {
			components[cid] = domain.FigmaComponent{Key: comp.Key, Name: comp.Name, Description: comp.Description}
		}
		domainNodes[id] = domain.FigmaNodeDetail{Document: n.Document, Components: components}
	}

	return web.NewJSONResponse(http.StatusOK, domain.FigmaNodesResponse{Success: true, Nodes: domainNodes})
}

// GetImages renders nodes as images and returns download URLs.
func (c *FigmaController) GetImages(req web.Request) web.Response {
	fileKey, ok := req.Param(domain.PathParamFileKey)
	if !ok || fileKey == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.FigmaImagesResponse{Error: "file_key is required"})
	}

	idsStr, _ := req.Query(domain.QueryParamNodeIDs)
	if idsStr == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.FigmaImagesResponse{Error: "ids query param is required"})
	}

	format, _ := req.Query(domain.QueryParamFormat)
	scaleStr, _ := req.Query(domain.QueryParamScale)
	scale := 2.0
	if scaleStr != "" {
		if s, err := strconv.ParseFloat(scaleStr, 64); err == nil {
			scale = s
		}
	}

	nodeIDs := strings.Split(idsStr, ",")
	images, err := c.client.GetImages(fileKey, nodeIDs, format, scale)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.FigmaImagesResponse{Error: err.Error()})
	}

	domainImages := make([]domain.FigmaImage, len(images))
	for i, img := range images {
		domainImages[i] = domain.FigmaImage{NodeID: img.NodeID, ImageURL: img.ImageURL}
	}

	return web.NewJSONResponse(http.StatusOK, domain.FigmaImagesResponse{Success: true, Images: domainImages})
}

// GetComments returns comments on a Figma file.
func (c *FigmaController) GetComments(req web.Request) web.Response {
	fileKey, ok := req.Param(domain.PathParamFileKey)
	if !ok || fileKey == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.FigmaCommentsResponse{Error: "file_key is required"})
	}

	comments, err := c.client.GetComments(fileKey)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.FigmaCommentsResponse{Error: err.Error()})
	}

	domainComments := make([]domain.FigmaComment, len(comments))
	for i, cm := range comments {
		domainComments[i] = domain.FigmaComment{
			ID: cm.ID, Message: cm.Message, CreatedAt: cm.CreatedAt, User: cm.User.Handle,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.FigmaCommentsResponse{Success: true, Comments: domainComments})
}

// GetProjectFiles lists files in a Figma project.
func (c *FigmaController) GetProjectFiles(req web.Request) web.Response {
	projectID, ok := req.Param(domain.PathParamProjectID)
	if !ok || projectID == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.FigmaProjectFilesResponse{Error: "project_id is required"})
	}

	files, err := c.client.GetProjectFiles(projectID)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.FigmaProjectFilesResponse{Error: err.Error()})
	}

	domainFiles := make([]domain.FigmaProjectFile, len(files))
	for i, f := range files {
		domainFiles[i] = domain.FigmaProjectFile{
			Key: f.Key, Name: f.Name, ThumbnailURL: f.ThumbnailURL, LastModified: f.LastModified,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.FigmaProjectFilesResponse{Success: true, Files: domainFiles})
}

// GetComponents returns the components in a Figma file.
func (c *FigmaController) GetComponents(req web.Request) web.Response {
	fileKey, ok := req.Param(domain.PathParamFileKey)
	if !ok || fileKey == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.FigmaComponentsResponse{Error: "file_key is required"})
	}

	components, err := c.client.GetComponents(fileKey)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.FigmaComponentsResponse{Error: err.Error()})
	}

	domainComponents := make([]domain.FigmaComponent, len(components))
	for i, comp := range components {
		domainComponents[i] = domain.FigmaComponent{Key: comp.Key, Name: comp.Name, Description: comp.Description}
	}

	return web.NewJSONResponse(http.StatusOK, domain.FigmaComponentsResponse{Success: true, Components: domainComponents})
}
