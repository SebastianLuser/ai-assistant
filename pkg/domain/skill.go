package domain

// SkillCreateRequest is the payload for creating a new skill.
type SkillCreateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
}

const (
	maxSkillNameLen    = 200
	maxSkillContentLen = 50_000
)

// Validate checks that a skill creation request is valid.
func (r SkillCreateRequest) Validate() error {
	if r.Name == "" {
		return Wrap(ErrValidation, "name is required")
	}
	if len(r.Name) > maxSkillNameLen {
		return Wrap(ErrValidation, "name exceeds maximum length")
	}
	if r.Content == "" {
		return Wrap(ErrValidation, "content is required")
	}
	if len(r.Content) > maxSkillContentLen {
		return Wrap(ErrValidation, "content exceeds maximum length")
	}
	return nil
}

// SkillResponse is the response for skill operations.
type SkillResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SkillListResponse is the response for listing skills.
type SkillListResponse struct {
	Success bool          `json:"success"`
	Skills  []SkillInfo   `json:"skills,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// SkillInfo is a summary of a skill (no content).
type SkillInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Enabled     bool     `json:"enabled"`
}
