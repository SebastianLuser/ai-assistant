package skills

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type QAIssue struct {
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
}

type ValidationResult struct {
	Skill    string    `json:"skill"`
	Valid    bool      `json:"valid"`
	Score    float64   `json:"score"`
	Issues   []QAIssue `json:"issues"`
}

type QualityRubric struct {
	MinDescriptionLen int      `yaml:"min_description_len"`
	RequiredFields    []string `yaml:"required_fields"`
	MaxContentLen     int      `yaml:"max_content_len"`
	MinTags           int      `yaml:"min_tags"`
}

func DefaultRubric() QualityRubric {
	return QualityRubric{
		MinDescriptionLen: 10,
		RequiredFields:    []string{"name", "description", "tags"},
		MaxContentLen:     50000,
		MinTags:           1,
	}
}

func LoadRubric(path string) (QualityRubric, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultRubric(), nil
	}
	var r QualityRubric
	if err := yaml.Unmarshal(data, &r); err != nil {
		return DefaultRubric(), fmt.Errorf("qa: parse rubric: %w", err)
	}
	return r, nil
}

func ValidateSkill(skill Skill, rubric QualityRubric) ValidationResult {
	result := ValidationResult{Skill: skill.Name, Valid: true}
	maxScore := 0
	earned := 0

	for _, field := range rubric.RequiredFields {
		maxScore++
		switch field {
		case "name":
			if skill.Name == "" {
				result.Issues = append(result.Issues, QAIssue{"error", "required_field", "name is required"})
			} else {
				earned++
			}
		case "description":
			if skill.Description == "" {
				result.Issues = append(result.Issues, QAIssue{"error", "required_field", "description is required"})
			} else if len(skill.Description) < rubric.MinDescriptionLen {
				result.Issues = append(result.Issues, QAIssue{"warning", "min_description", fmt.Sprintf("description too short (min %d chars)", rubric.MinDescriptionLen)})
			} else {
				earned++
			}
		case "tags":
			if len(skill.Tags) < rubric.MinTags {
				result.Issues = append(result.Issues, QAIssue{"error", "min_tags", fmt.Sprintf("at least %d tag(s) required", rubric.MinTags)})
			} else {
				earned++
			}
		}
	}

	maxScore++
	if skill.Content == "" {
		result.Issues = append(result.Issues, QAIssue{"error", "empty_content", "skill content is empty"})
	} else {
		earned++
	}

	maxScore++
	if rubric.MaxContentLen > 0 && len(skill.Content) > rubric.MaxContentLen {
		result.Issues = append(result.Issues, QAIssue{"warning", "max_content", fmt.Sprintf("content exceeds %d chars", rubric.MaxContentLen)})
	} else {
		earned++
	}

	maxScore++
	if strings.Contains(skill.Content, "TODO") || strings.Contains(skill.Content, "FIXME") {
		result.Issues = append(result.Issues, QAIssue{"warning", "wip_markers", "content contains TODO/FIXME markers"})
	} else {
		earned++
	}

	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			result.Valid = false
			break
		}
	}

	if maxScore > 0 {
		result.Score = float64(earned) / float64(maxScore)
	}

	return result
}

func ValidateAll(skills []Skill, rubric QualityRubric) []ValidationResult {
	seen := make(map[string]bool)
	var results []ValidationResult

	for _, s := range skills {
		r := ValidateSkill(s, rubric)

		if seen[s.Name] {
			r.Issues = append(r.Issues, QAIssue{"error", "duplicate_name", fmt.Sprintf("duplicate skill name: %s", s.Name)})
			r.Valid = false
		}
		seen[s.Name] = true

		results = append(results, r)
	}

	return results
}
