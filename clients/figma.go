package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	figmaBaseURL        = "https://api.figma.com/v1"
	figmaDefaultTimeout = 30 * time.Second
)

// FigmaFile represents a Figma file's metadata.
type FigmaFile struct {
	Name         string `json:"name"`
	LastModified string `json:"lastModified"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Version      string `json:"version"`
}

// FigmaComponent represents a component in a Figma file.
type FigmaComponent struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FigmaNode represents a node in a Figma file.
type FigmaNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// FigmaNodeDetail includes the full node document from the API.
type FigmaNodeDetail struct {
	Document json.RawMessage `json:"document"`
	// Components maps component ID -> FigmaComponent
	Components map[string]FigmaComponent `json:"components"`
}

// FigmaImage maps node IDs to image URLs.
type FigmaImage struct {
	NodeID   string `json:"node_id"`
	ImageURL string `json:"image_url"`
}

// FigmaComment represents a comment on a Figma file.
type FigmaComment struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Handle string `json:"handle"`
	} `json:"user"`
}

// FigmaProjectFile represents a file inside a Figma project.
type FigmaProjectFile struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	ThumbnailURL string `json:"thumbnail_url"`
	LastModified string `json:"last_modified"`
}

// FigmaClient is the Figma REST API client.
type FigmaClient struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewFigmaClient creates a new Figma API client.
func NewFigmaClient(token string) *FigmaClient {
	return &FigmaClient{
		token:      token,
		baseURL:    figmaBaseURL,
		httpClient: &http.Client{Timeout: figmaDefaultTimeout},
	}
}

// NewFigmaClientWithBaseURL creates a Figma client pointing at a custom base URL (for testing).
func NewFigmaClientWithBaseURL(token, baseURL string) *FigmaClient {
	c := NewFigmaClient(token)
	c.baseURL = baseURL
	return c
}

// GetFile returns metadata for a Figma file.
func (c *FigmaClient) GetFile(fileKey string) (FigmaFile, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("/files/%s?depth=1", fileKey))
	if err != nil {
		return FigmaFile{}, err
	}

	var result FigmaFile
	if err := json.Unmarshal(resp, &result); err != nil {
		return FigmaFile{}, fmt.Errorf("figma: parse file: %w", err)
	}

	return result, nil
}

// GetNodes returns specific nodes from a Figma file.
func (c *FigmaClient) GetNodes(fileKey string, nodeIDs []string) (map[string]FigmaNodeDetail, error) {
	ids := strings.Join(nodeIDs, ",")
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("/files/%s/nodes?ids=%s", fileKey, ids))
	if err != nil {
		return nil, err
	}

	var result struct {
		Nodes map[string]FigmaNodeDetail `json:"nodes"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("figma: parse nodes: %w", err)
	}

	return result.Nodes, nil
}

// GetImages renders nodes as images and returns download URLs.
// Format can be "png", "svg", "jpg", or "pdf".
func (c *FigmaClient) GetImages(fileKey string, nodeIDs []string, format string, scale float64) ([]FigmaImage, error) {
	if format == "" {
		format = "png"
	}
	if scale <= 0 {
		scale = 2
	}

	ids := strings.Join(nodeIDs, ",")
	path := fmt.Sprintf("/images/%s?ids=%s&format=%s&scale=%.0f", fileKey, ids, format, scale)

	resp, err := c.doRequest(http.MethodGet, path)
	if err != nil {
		return nil, err
	}

	var result struct {
		Images map[string]string `json:"images"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("figma: parse images: %w", err)
	}

	var images []FigmaImage
	for id, url := range result.Images {
		images = append(images, FigmaImage{NodeID: id, ImageURL: url})
	}

	return images, nil
}

// GetComments returns comments on a Figma file.
func (c *FigmaClient) GetComments(fileKey string) ([]FigmaComment, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("/files/%s/comments", fileKey))
	if err != nil {
		return nil, err
	}

	var result struct {
		Comments []FigmaComment `json:"comments"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("figma: parse comments: %w", err)
	}

	return result.Comments, nil
}

// GetProjectFiles lists files in a Figma project.
func (c *FigmaClient) GetProjectFiles(projectID string) ([]FigmaProjectFile, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("/projects/%s/files", projectID))
	if err != nil {
		return nil, err
	}

	var result struct {
		Files []FigmaProjectFile `json:"files"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("figma: parse project files: %w", err)
	}

	return result.Files, nil
}

// GetComponents returns the components in a Figma file.
func (c *FigmaClient) GetComponents(fileKey string) ([]FigmaComponent, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("/files/%s/components", fileKey))
	if err != nil {
		return nil, err
	}

	var result struct {
		Meta struct {
			Components []FigmaComponent `json:"components"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("figma: parse components: %w", err)
	}

	return result.Meta.Components, nil
}

func (c *FigmaClient) doRequest(method, path string) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("figma: create request: %w", err)
	}

	req.Header.Set("X-Figma-Token", c.token)
	req.Header.Set("Accept", contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("figma: send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("figma: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("figma: api error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
