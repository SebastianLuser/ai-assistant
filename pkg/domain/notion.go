package domain

type NotionCreateRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Validate checks that a Notion page creation request is valid.
func (r NotionCreateRequest) Validate() error {
	if r.Title == "" {
		return Wrap(ErrValidation, "title is required")
	}
	return nil
}

type NotionResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

type NotionSearchRequest struct {
	Query string `json:"query"`
}

type NotionSearchResponse struct {
	Success bool          `json:"success"`
	Results []NotionPage  `json:"results"`
	Error   string        `json:"error,omitempty"`
}

type NotionPage struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
