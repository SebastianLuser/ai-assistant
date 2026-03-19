package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const notionBaseURL = "https://api.notion.com/v1"

type NotionClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type NotionPage struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content,omitempty"`
	LastEdited time.Time `json:"last_edited_time"`
}

func NewNotionClient(apiKey string) *NotionClient {
	return &NotionClient{
		apiKey:     apiKey,
		baseURL:    notionBaseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// NewNotionClientWithBaseURL creates a Notion client pointing at a custom base URL (for testing).
func NewNotionClientWithBaseURL(apiKey, baseURL string) *NotionClient {
	c := NewNotionClient(apiKey)
	c.baseURL = baseURL
	return c
}

func (c *NotionClient) CreatePage(parentPageID, title, content string) (string, error) {
	body := map[string]any{
		"parent": map[string]string{"page_id": parentPageID},
		"properties": map[string]any{
			"title": map[string]any{
				"title": []map[string]any{
					{"text": map[string]string{"content": title}},
				},
			},
		},
		"children": []map[string]any{
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]any{
					"rich_text": []map[string]any{
						{"type": "text", "text": map[string]string{"content": content}},
					},
				},
			},
		},
	}

	resp, err := c.doRequest(http.MethodPost, "/pages", body)
	if err != nil {
		return "", err
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("notion: parse create response: %w", err)
	}

	return result.ID, nil
}

func (c *NotionClient) GetPage(pageID string) (NotionPage, error) {
	resp, err := c.doRequest(http.MethodGet, "/pages/"+pageID, nil)
	if err != nil {
		return NotionPage{}, err
	}

	var raw struct {
		ID             string `json:"id"`
		LastEditedTime string `json:"last_edited_time"`
		Properties     struct {
			Title struct {
				Title []struct {
					PlainText string `json:"plain_text"`
				} `json:"title"`
			} `json:"title"`
		} `json:"properties"`
	}

	if err := json.Unmarshal(resp, &raw); err != nil {
		return NotionPage{}, fmt.Errorf("notion: parse page: %w", err)
	}

	page := NotionPage{ID: raw.ID}
	if len(raw.Properties.Title.Title) > 0 {
		page.Title = raw.Properties.Title.Title[0].PlainText
	}
	page.LastEdited, _ = time.Parse(time.RFC3339, raw.LastEditedTime)

	return page, nil
}

func (c *NotionClient) QueryDatabase(databaseID string, filter map[string]any) ([]NotionPage, error) {
	body := map[string]any{}
	if filter != nil {
		body["filter"] = filter
	}

	resp, err := c.doRequest(http.MethodPost, "/databases/"+databaseID+"/query", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []struct {
			ID             string `json:"id"`
			LastEditedTime string `json:"last_edited_time"`
			Properties     struct {
				Name struct {
					Title []struct {
						PlainText string `json:"plain_text"`
					} `json:"title"`
				} `json:"Name"`
			} `json:"properties"`
		} `json:"results"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("notion: parse query: %w", err)
	}

	pages := make([]NotionPage, len(result.Results))
	for i, r := range result.Results {
		pages[i] = NotionPage{ID: r.ID}
		if len(r.Properties.Name.Title) > 0 {
			pages[i].Title = r.Properties.Name.Title[0].PlainText
		}
		pages[i].LastEdited, _ = time.Parse(time.RFC3339, r.LastEditedTime)
	}

	return pages, nil
}

func (c *NotionClient) doRequest(method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("notion: marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("notion: create request: %w", err)
	}

	req.Header.Set(headerAuthorization, "Bearer "+c.apiKey)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("notion: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("notion: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("notion: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
