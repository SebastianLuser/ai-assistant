package controller

import (
	"net/http"
	"testing"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProjectController_GetStatus_MissingName(t *testing.T) {
	ctrl := NewProjectController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "project name is required", errorFromBody(t, resp.Body))
}

func TestProjectController_GetStatus_EmptyName(t *testing.T) {
	ctrl := NewProjectController(nil)
	req := test.NewMockRequest().WithParam("name", "")

	resp := ctrl.GetStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestProjectController_GetStatus_Success(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchFTS", "myproject", mock.Anything).Return([]domain.Memory{
		{ID: 1, Content: "myproject update: all good"},
	}, nil)

	ai := new(test.MockAIProvider)
	ai.On("Complete", mock.Anything, mock.Anything).Return("Project is going well", nil)

	embedder := new(test.MockEmbedder)
	uc := usecase.NewProjectUseCase(store, embedder, ai)
	ctrl := NewProjectController(uc)
	req := test.NewMockRequest().WithParam("name", "myproject")

	resp := ctrl.GetStatus(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestProjectController_GetStatus_StoreError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchFTS", "myproject", mock.Anything).Return([]domain.Memory(nil), domain.ErrStoreOpen)

	embedder := new(test.MockEmbedder)
	uc := usecase.NewProjectUseCase(store, embedder, nil)
	ctrl := NewProjectController(uc)
	req := test.NewMockRequest().WithParam("name", "myproject")

	resp := ctrl.GetStatus(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}
