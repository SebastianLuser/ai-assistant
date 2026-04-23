package controller

import (
	"net/http"

	"jarvis/internal/skills"
	"jarvis/web"
)

type SkillsQAController struct {
	loader skills.SkillProvider
	rubric skills.QualityRubric
}

func NewSkillsQAController(loader skills.SkillProvider, rubric skills.QualityRubric) *SkillsQAController {
	return &SkillsQAController{loader: loader, rubric: rubric}
}

func (c *SkillsQAController) Report(req web.Request) web.Response {
	loaded, err := c.loader.LoadEnabled()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, map[string]any{
			"success": false, "error": err.Error(),
		})
	}

	results := skills.ValidateAll(loaded, c.rubric)

	valid := 0
	for _, r := range results {
		if r.Valid {
			valid++
		}
	}

	return web.NewJSONResponse(http.StatusOK, map[string]any{
		"success": true,
		"total":   len(results),
		"valid":   valid,
		"invalid": len(results) - valid,
		"results": results,
	})
}

func (c *SkillsQAController) Validate(req web.Request) web.Response {
	var payload struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Content     string   `json:"content"`
	}
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, map[string]any{
			"success": false, "error": "invalid body",
		})
	}

	skill := skills.Skill{
		Name:        payload.Name,
		Description: payload.Description,
		Tags:        payload.Tags,
		Content:     payload.Content,
	}

	result := skills.ValidateSkill(skill, c.rubric)

	return web.NewJSONResponse(http.StatusOK, map[string]any{
		"success": true,
		"result":  result,
	})
}
