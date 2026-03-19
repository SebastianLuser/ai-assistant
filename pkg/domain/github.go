package domain

// GitHub path parameter constants.
const (
	PathParamOwner = "owner"
	PathParamRepo  = "repo"
)

// GitHubRepo represents a GitHub repository.
type GitHubRepo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string `json:"url"`
}

// GitHubIssue represents a GitHub issue.
type GitHubIssue struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

// GitHubPullRequest represents a GitHub pull request.
type GitHubPullRequest struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
	Draft  bool   `json:"draft"`
}

const (
	maxGitHubIssueTitleLen = 500
)

// GitHubCreateIssueRequest is the payload for creating a GitHub issue.
type GitHubCreateIssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Validate checks that a GitHub issue creation request is valid.
func (r GitHubCreateIssueRequest) Validate() error {
	if r.Title == "" {
		return Wrap(ErrValidation, "title is required")
	}
	if len(r.Title) > maxGitHubIssueTitleLen {
		return Wrap(ErrValidation, "title exceeds maximum length")
	}
	return nil
}

// GitHubRepoListResponse is the response for listing repos.
type GitHubRepoListResponse struct {
	Success bool         `json:"success"`
	Repos   []GitHubRepo `json:"repos,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// GitHubIssueListResponse is the response for listing issues.
type GitHubIssueListResponse struct {
	Success bool          `json:"success"`
	Issues  []GitHubIssue `json:"issues,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// GitHubIssueResponse is the response for creating an issue.
type GitHubIssueResponse struct {
	Success bool   `json:"success"`
	Issue   *GitHubIssue `json:"issue,omitempty"`
	Error   string `json:"error,omitempty"`
}

// GitHubPRListResponse is the response for listing pull requests.
type GitHubPRListResponse struct {
	Success      bool                `json:"success"`
	PullRequests []GitHubPullRequest `json:"pull_requests,omitempty"`
	Error        string              `json:"error,omitempty"`
}
