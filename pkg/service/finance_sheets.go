package service

import (
	"jarvis/clients"
	"jarvis/pkg/domain"
)

type SheetsFinanceService struct {
	client    *clients.SheetsClient
	sheetName string
}

func NewSheetsFinanceService(client *clients.SheetsClient, sheetName string) *SheetsFinanceService {
	return &SheetsFinanceService{client: client, sheetName: sheetName}
}

func (r *SheetsFinanceService) SaveExpense(expense domain.ParsedExpense) error {
	row := []any{
		expense.Date,
		expense.Description,
		expense.Category,
		expense.Amount,
		expense.AmountUSD,
		expense.PaidBy,
	}
	return r.client.AppendRow(r.sheetName, row)
}

var _ FinanceService = (*SheetsFinanceService)(nil)
