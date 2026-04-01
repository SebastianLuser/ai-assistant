package usecase_test

import (
	"errors"
	"testing"
	"time"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errProjectFTS = errors.New("fts failed")
	projectNotes  = []domain.Memory{
		{ID: 1, Content: "setup repo", Tags: []string{"jarvis", "project"}, CreatedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Content: "add tests", Tags: []string{"jarvis"}, CreatedAt: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)},
	}
)

func TestProjectUseCase_GetStatus_Success(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SearchFTS", "jarvis", 20).Return(projectNotes, nil)
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "El proyecto avanza bien"})
	defer srv.Close()
	uc := usecase.NewProjectUseCase(repo, nil, ai)

	result, err := uc.GetStatus("jarvis")

	require.NoError(t, err)
	assert.Equal(t, true, result.Success)
	assert.Equal(t, "jarvis", result.Name)
	assert.Equal(t, "El proyecto avanza bien", result.Summary)
	assert.Equal(t, 2, result.NoteCount)
	repo.AssertExpectations(t)
}

func TestProjectUseCase_GetStatus_FTSError(t *testing.T) {
	repo := new(test.MockMemoryService)
	repo.On("SearchFTS", "myproj", 20).Return([]domain.Memory(nil), errProjectFTS)
	uc := usecase.NewProjectUseCase(repo, nil, nil)

	_, err := uc.GetStatus("myproj")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrProjectStatus))
}

func TestProjectUseCase_GetStatus_NoTaggedNotes_UsesAllResults(t *testing.T) {
	repo := new(test.MockMemoryService)
	untagged := []domain.Memory{
		{ID: 1, Content: "random note", Tags: []string{"other"}, CreatedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
	}
	repo.On("SearchFTS", "myproj", 20).Return(untagged, nil)
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: "status"})
	defer srv.Close()
	uc := usecase.NewProjectUseCase(repo, nil, ai)

	result, err := uc.GetStatus("myproj")

	require.NoError(t, err)
	assert.Equal(t, 1, result.NoteCount)
	repo.AssertExpectations(t)
}
