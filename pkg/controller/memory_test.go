package controller

import (
	"net/http"
	"testing"

	"jarvis/pkg/domain"
	"jarvis/pkg/service"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validNoteBody   = `{"content":"el pool de cartas tiene 40","tags":["game"]}`
	emptyNoteBody   = `{"content":""}`
	invalidNoteBody = `{broken`
)

var testEmbedding = []float64{0.1, 0.2, 0.3}

type stubEmbedder struct{}

func (s *stubEmbedder) Embed(_ string) ([]float64, error) { return testEmbedding, nil }

var _ service.Embedder = (*stubEmbedder)(nil)

type failEmbedder struct{}

func (f *failEmbedder) Embed(_ string) ([]float64, error) { return nil, domain.ErrEmbedGenerate }

var _ service.Embedder = (*failEmbedder)(nil)

func TestMemoryController_PostNote_InvalidJSON(t *testing.T) {
	store := new(test.MockMemoryService)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithBody(invalidNoteBody)

	resp := ctrl.PostNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestMemoryController_PostNote_EmptyContent(t *testing.T) {
	store := new(test.MockMemoryService)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithBody(emptyNoteBody)

	resp := ctrl.PostNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestMemoryController_PostNote_Success(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("Save", "el pool de cartas tiene 40", []string{"game"}, testEmbedding).Return(int64(1), nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithBody(validNoteBody)

	resp := ctrl.PostNote(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
	store.AssertExpectations(t)
}

func TestMemoryController_GetSearch_MissingQuery(t *testing.T) {
	store := new(test.MockMemoryService)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest()

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestMemoryController_GetSearch_FTSMode(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchFTS", "cartas", 5).Return([]domain.Memory{{ID: 1, Content: "cartas"}}, nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "cartas").WithQuery("mode", "fts")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
}

func TestMemoryController_GetSearch_CustomLimit(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchFTS", "test", 3).Return([]domain.Memory{}, nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "fts").WithQuery("limit", "3")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
}

func TestMemoryController_DeleteNote_InvalidID(t *testing.T) {
	store := new(test.MockMemoryService)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithParam("id", "abc")

	resp := ctrl.DeleteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestMemoryController_DeleteNote_MissingID(t *testing.T) {
	store := new(test.MockMemoryService)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest()

	resp := ctrl.DeleteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestMemoryController_DeleteNote_Success(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("Delete", int64(5)).Return(nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithParam("id", "5")

	resp := ctrl.DeleteNote(req)

	require.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
}

func TestMemoryController_DeleteNote_StoreError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("Delete", int64(5)).Return(domain.ErrStoreOpen)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithParam("id", "5")

	resp := ctrl.DeleteNote(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestMemoryController_PostNote_EmbedError(t *testing.T) {
	store := new(test.MockMemoryService)
	ctrl := NewMemoryController(store, &failEmbedder{})
	req := test.NewMockRequest().WithBody(validNoteBody)

	resp := ctrl.PostNote(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestMemoryController_PostNote_SaveError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("Save", "el pool de cartas tiene 40", []string{"game"}, testEmbedding).Return(int64(0), domain.ErrStoreOpen)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithBody(validNoteBody)

	resp := ctrl.PostNote(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestMemoryController_GetSearch_VectorMode(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("Search", testEmbedding, 5).Return([]domain.Memory{{ID: 1, Content: "test"}}, nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "vector")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
}

func TestMemoryController_GetSearch_HybridMode(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchHybrid", "test", testEmbedding, 5, domain.DefaultVecWeight, domain.DefaultFTSWeight).Return([]domain.Memory{{ID: 1}}, nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "hybrid")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
}

func TestMemoryController_GetSearch_FallbackMode(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchHybrid", "test", testEmbedding, 5, domain.DefaultVecWeight, domain.DefaultFTSWeight).Return([]domain.Memory{{ID: 1}}, nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "fallback")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestMemoryController_GetSearch_DefaultMode(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchHybrid", "test", testEmbedding, 5, domain.DefaultVecWeight, domain.DefaultFTSWeight).Return([]domain.Memory{{ID: 1}}, nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestMemoryController_GetSearch_InvalidLimit(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchFTS", "test", 5).Return([]domain.Memory{}, nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "fts").WithQuery("limit", "abc")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
}

func TestMemoryController_GetSearch_NegativeLimit(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchFTS", "test", 5).Return([]domain.Memory{}, nil)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "fts").WithQuery("limit", "-1")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	store.AssertExpectations(t)
}

func TestMemoryController_GetSearch_StoreError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("SearchFTS", "test", 5).Return([]domain.Memory(nil), domain.ErrStoreOpen)
	ctrl := NewMemoryController(store, &stubEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "fts")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestMemoryController_GetSearch_VectorEmbedError(t *testing.T) {
	store := new(test.MockMemoryService)
	ctrl := NewMemoryController(store, &failEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "vector")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestMemoryController_GetSearch_HybridEmbedError(t *testing.T) {
	store := new(test.MockMemoryService)
	ctrl := NewMemoryController(store, &failEmbedder{})
	req := test.NewMockRequest().WithQuery("q", "test").WithQuery("mode", "hybrid")

	resp := ctrl.GetSearch(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}
