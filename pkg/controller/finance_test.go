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
	validExpenseBody = `{"message":"gaste 5000 en el super","sender":"Sebas"}`
	noMessageBody    = `{"message":"","sender":"Sebas"}`
	noSenderBody     = `{"message":"gaste 5000 en el super"}`
	invalidJSONBody  = `{invalid`

	mockExpenseJSON = `{"amount":5000,"category":"Supermercado","description":"super","paid_by":"","date":"2026-03-10"}`
)

func TestFinanceController_PostExpense_InvalidJSON(t *testing.T) {
	ctrl := NewFinanceController(nil)
	req := test.NewMockRequest().WithBody(invalidJSONBody)

	resp := ctrl.PostExpense(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFinanceController_PostExpense_EmptyMessage(t *testing.T) {
	ctrl := NewFinanceController(nil)
	req := test.NewMockRequest().WithBody(noMessageBody)

	resp := ctrl.PostExpense(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFinanceController_PostExpense_DefaultSender(t *testing.T) {
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: mockExpenseJSON})
	defer srv.Close()
	uc := usecase.NewFinanceUseCase(ai, nil)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest().WithBody(noSenderBody)

	resp := ctrl.PostExpense(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}

func TestFinanceController_PostExpense_WithSender(t *testing.T) {
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: mockExpenseJSON})
	defer srv.Close()
	uc := usecase.NewFinanceUseCase(ai, nil)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest().WithBody(validExpenseBody)

	resp := ctrl.PostExpense(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}

func TestFinanceController_PostExpense_AIError(t *testing.T) {
	srv, ai := test.NewMockClaudeServerError(500, "api_error", "server error")
	defer srv.Close()
	uc := usecase.NewFinanceUseCase(ai, nil)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest().WithBody(validExpenseBody)

	resp := ctrl.PostExpense(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestFinanceController_GetSummary_DefaultPeriod(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("ListExpenses", mock.Anything, mock.Anything).Return([]domain.Expense{}, nil)
	ai := new(test.MockAIProvider)
	ai.On("Complete", mock.Anything, mock.Anything).Return("resumen del mes", nil)
	uc := usecase.NewFinanceUseCase(ai, nil)
	uc.SetMemoryService(store)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest()

	resp := ctrl.GetSummary(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFinanceController_GetSummary_WithPeriod(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("ListExpenses", mock.Anything, mock.Anything).Return([]domain.Expense{}, nil)
	ai := new(test.MockAIProvider)
	ai.On("Complete", mock.Anything, mock.Anything).Return("resumen semanal", nil)
	uc := usecase.NewFinanceUseCase(ai, nil)
	uc.SetMemoryService(store)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest().WithQuery("period", "week")

	resp := ctrl.GetSummary(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFinanceController_GetSummary_NilMemoryService(t *testing.T) {
	ai := new(test.MockAIProvider)
	ai.On("Complete", mock.Anything, mock.Anything).Return("sin datos", nil)
	uc := usecase.NewFinanceUseCase(ai, nil)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest()

	resp := ctrl.GetSummary(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestFinanceController_GetSummary_StoreError(t *testing.T) {
	store := new(test.MockMemoryService)
	store.On("ListExpenses", mock.Anything, mock.Anything).Return([]domain.Expense(nil), domain.ErrStoreOpen)
	uc := usecase.NewFinanceUseCase(nil, nil)
	uc.SetMemoryService(store)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest()

	resp := ctrl.GetSummary(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestFinanceController_GetSummary_AIError(t *testing.T) {
	ai := new(test.MockAIProvider)
	ai.On("Complete", mock.Anything, mock.Anything).Return("", domain.ErrClaudeAPI)
	uc := usecase.NewFinanceUseCase(ai, nil)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest()

	resp := ctrl.GetSummary(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}
