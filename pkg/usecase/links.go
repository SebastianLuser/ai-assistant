package usecase

import (
	"fmt"

	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

type LinkUseCase struct {
	repo     service.MemoryService
	embedder service.Embedder
}

func NewLinkUseCase(repo service.MemoryService, embedder service.Embedder) *LinkUseCase {
	return &LinkUseCase{repo: repo, embedder: embedder}
}

func (uc *LinkUseCase) Save(url, title string, tags []string) (int64, error) {
	content := fmt.Sprintf("[%s](%s)", title, url)

	allTags := append([]string{domain.LinkTag}, tags...)

	embedding, err := uc.embedder.Embed(title + " " + url)
	if err != nil {
		return 0, domain.Wrapf(domain.ErrLinkSave, err)
	}

	id, err := uc.repo.Save(content, allTags, embedding)
	if err != nil {
		return 0, domain.Wrapf(domain.ErrLinkSave, err)
	}

	return id, nil
}

func (uc *LinkUseCase) Search(query string, limit int) ([]domain.Memory, error) {
	results, err := uc.repo.SearchFTS(query, limit*3)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrLinkSearch, err)
	}

	var filtered []domain.Memory
	for _, m := range results {
		if hasTag(m.Tags, domain.LinkTag) {
			filtered = append(filtered, m)
		}
		if len(filtered) >= limit {
			break
		}
	}

	if filtered == nil {
		filtered = []domain.Memory{}
	}
	return filtered, nil
}

func hasTag(tags []string, target string) bool {
	for _, t := range tags {
		if t == target {
			return true
		}
	}
	return false
}
