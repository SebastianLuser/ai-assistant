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

const (
	validLinkBody   = `{"url":"https://example.com","title":"Example"}`
	emptyURLBody    = `{"url":"","title":"Example"}`
	invalidURLBody  = `{"url":"not a url","title":"Example"}`
	ftpURLBody      = `{"url":"ftp://files.example.com","title":"Files"}`
	invalidLinkJSON = `{bad`
)

func TestLinkController_PostLink_InvalidJSON(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithBody(invalidLinkJSON)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestLinkController_PostLink_EmptyURL(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithBody(emptyURLBody)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: url is required", errorFromBody(t, resp.Body))
}

func TestLinkController_PostLink_InvalidURL(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithBody(invalidURLBody)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: url must be a valid http or https URL", errorFromBody(t, resp.Body))
}

func TestLinkController_PostLink_FTPScheme(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithBody(ftpURLBody)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: url must be a valid http or https URL", errorFromBody(t, resp.Body))
}

func TestLinkController_PostLink_Success(t *testing.T) {
	store := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", mock.Anything).Return([]float64{0.1}, nil)
	store.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(int64(42), nil)
	uc := usecase.NewLinkUseCase(store, embedder)
	ctrl := NewLinkController(uc)
	req := test.NewMockRequest().WithBody(validLinkBody)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
	store.AssertExpectations(t)
}

func TestLinkController_PostLink_DefaultTitle(t *testing.T) {
	store := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", mock.Anything).Return([]float64{0.1}, nil)
	store.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	uc := usecase.NewLinkUseCase(store, embedder)
	ctrl := NewLinkController(uc)
	req := test.NewMockRequest().WithBody(`{"url":"https://example.com"}`)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}

func TestLinkController_PostLink_SaveError(t *testing.T) {
	store := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	embedder.On("Embed", mock.Anything).Return([]float64{0.1}, nil)
	store.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), domain.ErrStoreOpen)
	uc := usecase.NewLinkUseCase(store, embedder)
	ctrl := NewLinkController(uc)
	req := test.NewMockRequest().WithBody(validLinkBody)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestLinkController_GetSearch_MissingQuery(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestLinkController_GetSearch_EmptyQuery(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithQuery("q", "")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestLinkController_GetSearch_Success(t *testing.T) {
	store := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	store.On("SearchFTS", "golang", 15).Return([]domain.Memory{
		{ID: 1, Content: "[Go](https://golang.org)", Tags: []string{"link"}},
	}, nil)
	uc := usecase.NewLinkUseCase(store, embedder)
	ctrl := NewLinkController(uc)
	req := test.NewMockRequest().WithQuery("q", "golang")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
}

func TestLinkController_GetSearch_StoreError(t *testing.T) {
	store := new(test.MockMemoryService)
	embedder := new(test.MockEmbedder)
	store.On("SearchFTS", "test", 15).Return([]domain.Memory(nil), domain.ErrStoreOpen)
	uc := usecase.NewLinkUseCase(store, embedder)
	ctrl := NewLinkController(uc)
	req := test.NewMockRequest().WithQuery("q", "test")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}
