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

var (
	errLinkEmbed = errors.New("embed failed")
	errLinkSave  = errors.New("save failed")
	errLinkFTS   = errors.New("fts failed")
	linkVector   = []float64{0.5, 0.5}
)

func TestLinkUseCase_Save_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", "Example https://example.com").Return(linkVector, nil)
	repo.On("Save", "[Example](https://example.com)", []string{"link"}, linkVector).Return(int64(42), nil)
	uc := usecase.NewLinkUseCase(repo, embedder)

	id, err := uc.Save("https://example.com", "Example", nil)

	require.NoError(t, err)
	assert.Equal(t, int64(42), id)
	repo.AssertExpectations(t)
	embedder.AssertExpectations(t)
}

func TestLinkUseCase_Save_WithTags(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", "Docs https://docs.go.dev").Return(linkVector, nil)
	repo.On("Save", "[Docs](https://docs.go.dev)", []string{"link", "golang"}, linkVector).Return(int64(1), nil)
	uc := usecase.NewLinkUseCase(repo, embedder)

	id, err := uc.Save("https://docs.go.dev", "Docs", []string{"golang"})

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	repo.AssertExpectations(t)
}

func TestLinkUseCase_Save_EmbedError(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", "x https://x.com").Return([]float64(nil), errLinkEmbed)
	uc := usecase.NewLinkUseCase(repo, embedder)

	_, err := uc.Save("https://x.com", "x", nil)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrLinkSave))
}

func TestLinkUseCase_Save_RepoError(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", "x https://x.com").Return(linkVector, nil)
	repo.On("Save", "[x](https://x.com)", []string{"link"}, linkVector).Return(int64(0), errLinkSave)
	uc := usecase.NewLinkUseCase(repo, embedder)

	_, err := uc.Save("https://x.com", "x", nil)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrLinkSave))
}

func TestLinkUseCase_Search_FiltersLinkTag(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	results := []domain.Memory{
		{ID: 1, Content: "[Go](https://go.dev)", Tags: []string{"link", "golang"}},
		{ID: 2, Content: "nota random", Tags: []string{"note"}},
		{ID: 3, Content: "[Rust](https://rust-lang.org)", Tags: []string{"link"}},
	}
	repo.On("SearchFTS", "dev", 15).Return(results, nil)
	uc := usecase.NewLinkUseCase(repo, embedder)

	filtered, err := uc.Search("dev", 5)

	require.NoError(t, err)
	assert.Len(t, filtered, 2)
	assert.Equal(t, int64(1), filtered[0].ID)
	assert.Equal(t, int64(3), filtered[1].ID)
	repo.AssertExpectations(t)
}

func TestLinkUseCase_Search_RespectsLimit(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	results := []domain.Memory{
		{ID: 1, Tags: []string{"link"}},
		{ID: 2, Tags: []string{"link"}},
		{ID: 3, Tags: []string{"link"}},
	}
	repo.On("SearchFTS", "test", 3).Return(results, nil)
	uc := usecase.NewLinkUseCase(repo, embedder)

	filtered, err := uc.Search("test", 1)

	require.NoError(t, err)
	assert.Len(t, filtered, 1)
}

func TestLinkUseCase_Search_NoResults_ReturnsEmptySlice(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	repo.On("SearchFTS", "nada", 15).Return([]domain.Memory{}, nil)
	uc := usecase.NewLinkUseCase(repo, embedder)

	filtered, err := uc.Search("nada", 5)

	require.NoError(t, err)
	assert.Equal(t, []domain.Memory{}, filtered)
}

func TestLinkUseCase_Search_Error(t *testing.T) {
	repo := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	repo.On("SearchFTS", "q", 15).Return([]domain.Memory(nil), errLinkFTS)
	uc := usecase.NewLinkUseCase(repo, embedder)

	_, err := uc.Search("q", 5)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrLinkSearch))
}
