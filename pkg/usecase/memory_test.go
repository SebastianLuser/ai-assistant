package usecase_test

import (
	"errors"
	"testing"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testQuery = "cartas del juego"
	testLimit = 5
)

var (
	errSearchFailed = errors.New("search failed")
	testVector      = []float64{0.1, 0.2, 0.3}
	hybridMemory    = []domain.Memory{{ID: 1, Content: "hybrid result", Score: 0.9}}
	ftsMemory       = []domain.Memory{{ID: 2, Content: "fts result", Score: 0.8}}
	vectorMemory    = []domain.Memory{{ID: 3, Content: "vector result", Score: 0.7}}
	emptyMemories   = []domain.Memory{}
)

func TestSearchUseCase_HybridSuccess_ReturnsHybridResults(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", testQuery).Return(testVector, nil)
	repo.On("SearchHybrid", testQuery, testVector, testLimit, 0.6, 0.4).Return(hybridMemory, nil)

	searcher := usecase.NewMemoryUseCase(repo, embedder)

	results, err := searcher.FallbackSearch(testQuery, testLimit)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "hybrid result", results[0].Content)
}

func TestSearchUseCase_HybridFails_FallsBackToFTS(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", testQuery).Return(testVector, nil)
	repo.On("SearchHybrid", testQuery, testVector, testLimit, 0.6, 0.4).Return(emptyMemories, errSearchFailed)
	repo.On("SearchFTS", testQuery, testLimit).Return(ftsMemory, nil)

	searcher := usecase.NewMemoryUseCase(repo, embedder)

	results, err := searcher.FallbackSearch(testQuery, testLimit)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "fts result", results[0].Content)
}

func TestSearchUseCase_HybridAndFTSFail_FallsBackToVector(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", testQuery).Return(testVector, nil)
	repo.On("SearchHybrid", testQuery, testVector, testLimit, 0.6, 0.4).Return(emptyMemories, errSearchFailed)
	repo.On("SearchFTS", testQuery, testLimit).Return(emptyMemories, errSearchFailed)
	repo.On("Search", testVector, testLimit).Return(vectorMemory, nil)

	searcher := usecase.NewMemoryUseCase(repo, embedder)

	results, err := searcher.FallbackSearch(testQuery, testLimit)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "vector result", results[0].Content)
}

func TestSearchUseCase_EmbeddingFails_FallsBackToFTS(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", testQuery).Return([]float64(nil), errSearchFailed)
	repo.On("SearchFTS", testQuery, testLimit).Return(ftsMemory, nil)

	searcher := usecase.NewMemoryUseCase(repo, embedder)

	results, err := searcher.FallbackSearch(testQuery, testLimit)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "fts result", results[0].Content)
}

func TestSearchUseCase_AllFail_ReturnsError(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", testQuery).Return([]float64(nil), errSearchFailed)
	repo.On("SearchFTS", testQuery, testLimit).Return(emptyMemories, errSearchFailed)

	searcher := usecase.NewMemoryUseCase(repo, embedder)

	_, err := searcher.FallbackSearch(testQuery, testLimit)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errSearchFailed))
}

func TestSearchUseCase_HybridEmpty_FallsBackToFTS(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", testQuery).Return(testVector, nil)
	repo.On("SearchHybrid", testQuery, testVector, testLimit, 0.6, 0.4).Return(emptyMemories, nil)
	repo.On("SearchFTS", testQuery, testLimit).Return(ftsMemory, nil)

	searcher := usecase.NewMemoryUseCase(repo, embedder)

	results, err := searcher.FallbackSearch(testQuery, testLimit)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "fts result", results[0].Content)
}
