package usecase

import (
	"log"

	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

type MemoryUseCase struct {
	repo     service.MemoryService
	embedder service.Embedder
}

func NewMemoryUseCase(repo service.MemoryService, embedder service.Embedder) *MemoryUseCase {
	return &MemoryUseCase{repo: repo, embedder: embedder}
}

func (uc *MemoryUseCase) FallbackSearch(query string, limit int) ([]domain.Memory, error) {
	embedding, embErr := uc.embedder.Embed(query)

	if embErr == nil {
		results, err := uc.repo.SearchHybrid(query, embedding, limit, domain.DefaultVecWeight, domain.DefaultFTSWeight)
		if err == nil && len(results) > 0 {
			return results, nil
		}
		if err != nil {
			log.Printf("fallback: hybrid search failed: %v", err)
		}
	}

	results, err := uc.repo.SearchFTS(query, limit)
	if err == nil && len(results) > 0 {
		log.Println("fallback: using FTS results")
		return results, nil
	}
	if err != nil {
		log.Printf("fallback: FTS search failed: %v", err)
	}

	if embErr == nil {
		results, err := uc.repo.Search(embedding, limit)
		if err == nil && len(results) > 0 {
			log.Println("fallback: using vector results")
			return results, nil
		}
		if err != nil {
			log.Printf("fallback: vector search failed: %v", err)
		}
	}

	if embErr != nil {
		return nil, embErr
	}
	return []domain.Memory{}, nil
}
