package usecase

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testDate = "2026-03-10"
)

func TestFormatNumber_Small(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "0"},
		{"hundreds", 500, "500"},
		{"thousands", 5000, "5,000"},
		{"tens_of_thousands", 15000, "15,000"},
		{"hundreds_of_thousands", 150000, "150,000"},
		{"millions", 1500000, "1,500,000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNumber(tt.input)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatExpenseResponse_ARS(t *testing.T) {
	expense := domain.ParsedExpense{
		Amount:      5000,
		Category:    "Supermercado",
		Description: "Compras",
		PaidBy:      "Sebas",
		Date:        testDate,
	}

	result := FormatExpenseResponse(expense)

	expected := "\U0001F6D2 Anotado!\nSupermercado — Compras\nSebas pago $5,000 el 10/03/2026"
	assert.Equal(t, expected, result)
}

func TestFormatExpenseResponse_USD(t *testing.T) {
	expense := domain.ParsedExpense{
		Amount:      0,
		AmountUSD:   20,
		Category:    "Entretenimiento",
		Description: "Netflix",
		PaidBy:      "Sebas",
		Date:        testDate,
	}

	result := FormatExpenseResponse(expense)

	expected := "\U0001F3AE Anotado!\nEntretenimiento — Netflix\nSebas pago US$20 el 10/03/2026"
	assert.Equal(t, expected, result)
}

func TestFormatExpenseResponse_CategoryEmoji(t *testing.T) {
	tests := []struct {
		category string
		emoji    string
	}{
		{"Supermercado", "\U0001F6D2"},
		{"Restaurante", "\U0001F354"},
		{"Transporte", "\U0001F697"},
		{"Servicios", "\U0001F4F1"},
		{"Salud", "\U0001F48A"},
		{"Ropa", "\U0001F455"},
		{"Entretenimiento", "\U0001F3AE"},
		{"Educacion", "\U0001F4DA"},
		{"Hogar", "\U0001F3E0"},
		{"Otro", "\U0001F4E6"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			expense := domain.ParsedExpense{
				Amount:   1000,
				Category: tt.category,
				PaidBy:   "Test",
				Date:     testDate,
			}

			result := FormatExpenseResponse(expense)

			expected := tt.emoji + " Anotado!\n" + tt.category + " — \nTest pago $1,000 el 10/03/2026"
			assert.Equal(t, expected, result)
		})
	}
}

func TestFormatExpenseResponse_UnknownCategory_DefaultEmoji(t *testing.T) {
	expense := domain.ParsedExpense{
		Amount:   1000,
		Category: "CategoriaInventada",
		PaidBy:   "Test",
		Date:     testDate,
	}

	result := FormatExpenseResponse(expense)

	expected := "\U0001F4E6 Anotado!\nCategoriaInventada — \nTest pago $1,000 el 10/03/2026"
	assert.Equal(t, expected, result)
}

func TestFormatExpenseResponse_DateFormat(t *testing.T) {
	expense := domain.ParsedExpense{
		Amount:   1000,
		Category: "Otro",
		PaidBy:   "Sebas",
		Date:     "2026-03-10",
	}

	result := FormatExpenseResponse(expense)

	expected := "\U0001F4E6 Anotado!\nOtro — \nSebas pago $1,000 el 10/03/2026"
	assert.Equal(t, expected, result)
}

// --- ProcessExpense tests ---

func TestFinanceUseCase_ProcessExpense_Success(t *testing.T) {
	ai := new(test.MockAIProvider)
	repo := new(test.MockFinanceService)

	expense := domain.ParsedExpense{
		Amount:      5000,
		Category:    "Supermercado",
		Description: "Compras",
		PaidBy:      "Sebas",
		Date:        "2026-03-19",
	}

	ai.On("CompleteJSON", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		target := args.Get(2).(*domain.ParsedExpense)
		*target = expense
	}).Return(nil)
	repo.On("SaveExpense", expense).Return(nil)

	uc := NewFinanceUseCase(ai, repo)

	result, err := uc.ProcessExpense("gaste 5000 en el super", "Sebas")

	require.NoError(t, err)
	assert.Contains(t, result, "Anotado!")
	assert.Contains(t, result, "Supermercado")
	ai.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestFinanceUseCase_ProcessExpense_AIError(t *testing.T) {
	ai := new(test.MockAIProvider)
	ai.On("CompleteJSON", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("ai error"))

	uc := NewFinanceUseCase(ai, nil)

	_, err := uc.ProcessExpense("gaste 5000", "Sebas")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrFinanceParseExpense))
}

func TestFinanceUseCase_ProcessExpense_RepoError(t *testing.T) {
	ai := new(test.MockAIProvider)
	repo := new(test.MockFinanceService)

	ai.On("CompleteJSON", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		target := args.Get(2).(*domain.ParsedExpense)
		*target = domain.ParsedExpense{Amount: 1000, Category: "Otro", PaidBy: "Sebas", Date: "2026-03-19"}
	}).Return(nil)
	repo.On("SaveExpense", mock.Anything).Return(errors.New("sheets error"))

	uc := NewFinanceUseCase(ai, repo)

	_, err := uc.ProcessExpense("gaste 1000", "Sebas")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrFinanceWriteSheets))
}

func TestFinanceUseCase_ProcessExpense_NoRepo_SkipsSave(t *testing.T) {
	ai := new(test.MockAIProvider)

	ai.On("CompleteJSON", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		target := args.Get(2).(*domain.ParsedExpense)
		*target = domain.ParsedExpense{Amount: 1000, Category: "Otro", PaidBy: "Sebas", Date: "2026-03-19"}
	}).Return(nil)

	uc := NewFinanceUseCase(ai, nil)

	result, err := uc.ProcessExpense("gaste 1000", "Sebas")

	require.NoError(t, err)
	assert.Contains(t, result, "Anotado!")
}

func TestFinanceUseCase_ProcessExpense_EmptyDate_SetsToday(t *testing.T) {
	ai := new(test.MockAIProvider)

	ai.On("CompleteJSON", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		target := args.Get(2).(*domain.ParsedExpense)
		*target = domain.ParsedExpense{Amount: 1000, Category: "Otro", PaidBy: "Sebas", Date: ""}
	}).Return(nil)

	uc := NewFinanceUseCase(ai, nil)

	result, err := uc.ProcessExpense("gaste 1000", "Sebas")

	require.NoError(t, err)
	today := time.Now().Format(domain.DateFormatISO)
	todayDisplay, _ := time.Parse(domain.DateFormatISO, today)
	assert.Contains(t, result, todayDisplay.Format(domain.DateFormatDisplay))
}

func TestFinanceUseCase_ProcessExpense_EmptyPaidBy_UsesSender(t *testing.T) {
	ai := new(test.MockAIProvider)

	ai.On("CompleteJSON", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		target := args.Get(2).(*domain.ParsedExpense)
		*target = domain.ParsedExpense{Amount: 1000, Category: "Otro", PaidBy: "", Date: "2026-03-19"}
	}).Return(nil)

	uc := NewFinanceUseCase(ai, nil)

	result, err := uc.ProcessExpense("gaste 1000", "Laura")

	require.NoError(t, err)
	assert.Contains(t, result, "Laura")
}

// --- GetSummary tests ---

func TestFinanceUseCase_GetSummary_Success(t *testing.T) {
	ai := new(test.MockAIProvider)
	repo := new(test.MockMemoryService)

	expenses := []domain.Expense{
		{Amount: 5000, Category: "Supermercado", PaidBy: "Sebas"},
		{Amount: 3000, AmountUSD: 10, Category: "Entretenimiento", PaidBy: "Sebas"},
	}
	repo.On("ListExpenses", mock.Anything, mock.Anything).Return(expenses, nil)
	ai.On("Complete", mock.Anything, mock.Anything).Return("Resumen financiero", nil)

	uc := NewFinanceUseCase(ai, nil)
	uc.SetMemoryService(repo)

	result, err := uc.GetSummary("month")

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "Resumen financiero", result.Summary)
	assert.Equal(t, "month", result.Period)
	assert.Equal(t, 8000.0, result.TotalARS)
	assert.Equal(t, 10.0, result.TotalUSD)
	assert.Len(t, result.ByCategory, 2)
}

func TestFinanceUseCase_GetSummary_NoMemoryService(t *testing.T) {
	ai := new(test.MockAIProvider)
	ai.On("Complete", mock.Anything, mock.Anything).Return("No hay gastos", nil)

	uc := NewFinanceUseCase(ai, nil)

	result, err := uc.GetSummary("week")

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 0.0, result.TotalARS)
}

func TestFinanceUseCase_GetSummary_ListExpensesError(t *testing.T) {
	ai := new(test.MockAIProvider)
	repo := new(test.MockMemoryService)
	repo.On("ListExpenses", mock.Anything, mock.Anything).Return([]domain.Expense(nil), errors.New("db error"))

	uc := NewFinanceUseCase(ai, nil)
	uc.SetMemoryService(repo)

	_, err := uc.GetSummary("month")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrFinanceSummary))
}

func TestFinanceUseCase_GetSummary_AIError(t *testing.T) {
	ai := new(test.MockAIProvider)
	repo := new(test.MockMemoryService)
	repo.On("ListExpenses", mock.Anything, mock.Anything).Return([]domain.Expense{}, nil)
	ai.On("Complete", mock.Anything, mock.Anything).Return("", errors.New("ai error"))

	uc := NewFinanceUseCase(ai, nil)
	uc.SetMemoryService(repo)

	_, err := uc.GetSummary("month")

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrFinanceSummary))
}

// --- periodRange tests ---

func TestPeriodRange_Week(t *testing.T) {
	from, to := periodRange("week")

	fromTime, _ := time.Parse(domain.DateFormatISO, from)
	toTime, _ := time.Parse(domain.DateFormatISO, to)

	diff := toTime.Sub(fromTime)
	assert.InDelta(t, 7, diff.Hours()/24, 1)
}

func TestPeriodRange_Year(t *testing.T) {
	from, to := periodRange("year")

	fromTime, _ := time.Parse(domain.DateFormatISO, from)
	toTime, _ := time.Parse(domain.DateFormatISO, to)

	diff := toTime.Sub(fromTime)
	assert.InDelta(t, 365, diff.Hours()/24, 2)
}

func TestPeriodRange_Month_Default(t *testing.T) {
	from, to := periodRange("month")

	fromTime, _ := time.Parse(domain.DateFormatISO, from)
	toTime, _ := time.Parse(domain.DateFormatISO, to)

	diff := toTime.Sub(fromTime)
	assert.InDelta(t, 30, diff.Hours()/24, 2)
}

func TestPeriodRange_UnknownPeriod_DefaultsToMonth(t *testing.T) {
	fromMonth, _ := periodRange("month")
	fromUnknown, _ := periodRange("unknown")

	assert.Equal(t, fromMonth, fromUnknown)
}

// Suppress unused import warnings
var _ = json.Marshal
