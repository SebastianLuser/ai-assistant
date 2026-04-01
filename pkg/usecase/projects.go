package usecase

import (
	"fmt"
	"strings"

	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

const projectStatusSystemPrompt = `Sos un asistente de gestión de proyectos. Hablas en español rioplatense.
Dado un conjunto de notas de un proyecto, generá un resumen de estado breve y útil.
Incluí: progreso general, últimas actividades, y posibles próximos pasos.
Respondé en texto plano, sin markdown.`

const projectSearchLimit = 20

type ProjectUseCase struct {
	repo     service.MemoryService
	embedder service.Embedder
	ai       domain.AIProvider
}

func NewProjectUseCase(repo service.MemoryService, embedder service.Embedder, ai domain.AIProvider) *ProjectUseCase {
	return &ProjectUseCase{repo: repo, embedder: embedder, ai: ai}
}

func (uc *ProjectUseCase) GetStatus(name string) (*domain.ProjectStatusResponse, error) {
	results, err := uc.repo.SearchFTS(name, projectSearchLimit)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrProjectStatus, err)
	}

	var projectNotes []domain.Memory
	for _, m := range results {
		if hasTag(m.Tags, name) || hasTag(m.Tags, "project") {
			projectNotes = append(projectNotes, m)
		}
	}

	if len(projectNotes) == 0 {
		projectNotes = results
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Proyecto: %s\nNotas encontradas: %d\n\n", name, len(projectNotes)))
	for _, n := range projectNotes {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", n.CreatedAt.Format(domain.DateFormatISO), n.Content))
	}
	sb.WriteString("\nGenerá un resumen de estado del proyecto.")

	summary, err := uc.ai.Complete(projectStatusSystemPrompt, sb.String())
	if err != nil {
		return nil, domain.Wrapf(domain.ErrProjectStatus, err)
	}

	return &domain.ProjectStatusResponse{
		Success:   true,
		Name:      name,
		Summary:   summary,
		NoteCount: len(projectNotes),
	}, nil
}
