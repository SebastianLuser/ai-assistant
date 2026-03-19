package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	githubBaseURL        = "https://api.github.com"
	githubDefaultTimeout = 15 * time.Second
	githubAcceptJSON     = "application/vnd.github+json"
)

// GitHubRepo represents a GitHub repository.
type GitHubRepo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string `json:"html_url"`
}

// GitHubIssue represents a GitHub issue.
type GitHubIssue struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"html_url"`
}

// GitHubPullRequest represents a GitHub pull request.
type GitHubPullRequest struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"html_url"`
	Draft  bool   `json:"draft"`
}

// GitHubClient is the GitHub API client.
type GitHubClient struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		baseURL:    githubBaseURL,
		httpClient: &http.Client{Timeout: githubDefaultTimeout},
	}
}

// NewGitHubClientWithBaseURL creates a GitHub client pointing at a custom base URL (for testing).
func NewGitHubClientWithBaseURL(token, baseURL string) *GitHubClient {
	c := NewGitHubClient(token)
	c.baseURL = baseURL
	return c
}

// ListRepos returns the authenticated user's repositories.
func (c *GitHubClient) ListRepos() ([]GitHubRepo, error) {
	resp, err := c.doRequest(http.MethodGet, "/user/repos", nil)
	if err != nil {
		return nil, err
	}

	var repos []GitHubRepo
	if err := json.Unmarshal(resp, &repos); err != nil {
		return nil, fmt.Errorf("github: parse repos: %w", err)
	}

	return repos, nil
}

// ListIssues returns issues for the given repository.
func (c *GitHubClient) ListIssues(owner, repo string) ([]GitHubIssue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var issues []GitHubIssue
	if err := json.Unmarshal(resp, &issues); err != nil {
		return nil, fmt.Errorf("github: parse issues: %w", err)
	}

	return issues, nil
}

// CreateIssue creates a new issue in the given repository.
func (c *GitHubClient) CreateIssue(owner, repo, title, body string) (GitHubIssue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	payload := map[string]string{
		"title": title,
		"body":  body,
	}

	resp, err := c.doRequest(http.MethodPost, path, payload)
	if err != nil {
		return GitHubIssue{}, err
	}

	var issue GitHubIssue
	if err := json.Unmarshal(resp, &issue); err != nil {
		return GitHubIssue{}, fmt.Errorf("github: parse create issue: %w", err)
	}

	return issue, nil
}

// ListPRs returns pull requests for the given repository.
func (c *GitHubClient) ListPRs(owner, repo string) ([]GitHubPullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var prs []GitHubPullRequest
	if err := json.Unmarshal(resp, &prs); err != nil {
		return nil, fmt.Errorf("github: parse pull requests: %w", err)
	}

	return prs, nil
}

func (c *GitHubClient) doRequest(method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("github: marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("github: create request: %w", err)
	}

	req.Header.Set(headerAuthorization, "Bearer "+c.token)
	req.Header.Set("Accept", githubAcceptJSON)
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("github: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
