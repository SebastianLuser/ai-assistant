package usecase

import (
	"fmt"
	"strings"
	"time"

	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

const expenseSystemPrompt = `Sos un parser de gastos en español argentino.
Extraé la información del gasto del mensaje del usuario.

Reglas:
- "lucas" o "luquitas" = multiplicar por 1000 (ej: "5 lucas" = 5000)
- Si no se menciona quién pagó, usá el sender que te paso
- Si menciona dólares/USD/usd, poné el monto en amount_usd y amount en 0
- La fecha es hoy salvo que diga otra cosa
- Categorías válidas: Supermercado, Restaurante, Transporte, Servicios, Salud, Ropa, Entretenimiento, Educacion, Hogar, Otro

Respondé SOLO con JSON válido, sin markdown, sin explicación:
{"amount": number, "amount_usd": number, "category": "string", "description": "string", "paid_by": "string", "date": "YYYY-MM-DD"}`

const summarySystemPrompt = `Sos un asistente financiero personal. Hablas en español rioplatense.
Genera un resumen claro y conciso de los gastos del periodo indicado.
Incluí: total en ARS, total en USD (si hay), desglose por categoría, y un consejo breve.
Respondé en texto plano, sin markdown.`

type FinanceUseCase struct {
	ai        domain.AIProvider
	repo      service.FinanceService
	memorySvc service.MemoryService
}

func NewFinanceUseCase(ai domain.AIProvider, repo service.FinanceService) *FinanceUseCase {
	return &FinanceUseCase{
		ai:   ai,
		repo: repo,
	}
}

func (uc *FinanceUseCase) SetMemoryService(svc service.MemoryService) {
	uc.memorySvc = svc
}

func (uc *FinanceUseCase) GetSummary(period string) (*domain.FinanceSummaryResponse, error) {
	from, to := periodRange(period)

	var expenses []domain.Expense
	if uc.memorySvc != nil {
		var err error
		expenses, err = uc.memorySvc.ListExpenses(from, to)
		if err != nil {
			return nil, domain.Wrapf(domain.ErrFinanceSummary, err)
		}
	}

	totalARS := 0.0
	totalUSD := 0.0
	byCategory := make(map[string]float64)
	for _, e := range expenses {
		totalARS += e.Amount
		totalUSD += e.AmountUSD
		byCategory[e.Category] += e.Amount + e.AmountUSD
	}

	prompt := fmt.Sprintf("Periodo: %s (del %s al %s)\nTotal ARS: $%.0f\nTotal USD: US$%.0f\nGastos por categoría:\n", period, from, to, totalARS, totalUSD)
	for cat, amt := range byCategory {
		prompt += fmt.Sprintf("- %s: $%.0f\n", cat, amt)
	}
	if len(expenses) == 0 {
		prompt += "No hay gastos registrados en este periodo.\n"
	}
	prompt += "\nGenerá un resumen financiero breve."

	summary, err := uc.ai.Complete(summarySystemPrompt, prompt)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrFinanceSummary, err)
	}

	return &domain.FinanceSummaryResponse{
		Success:    true,
		Summary:    summary,
		Period:     period,
		TotalARS:   totalARS,
		TotalUSD:   totalUSD,
		ByCategory: byCategory,
	}, nil
}

func periodRange(period string) (string, string) {
	now := time.Now()
	to := now.Format(domain.DateFormatISO)

	switch period {
	case domain.PeriodWeek:
		from := now.AddDate(0, 0, -7).Format(domain.DateFormatISO)
		return from, to
	case domain.PeriodYear:
		from := now.AddDate(-1, 0, 0).Format(domain.DateFormatISO)
		return from, to
	default:
		from := now.AddDate(0, -1, 0).Format(domain.DateFormatISO)
		return from, to
	}
}

func (uc *FinanceUseCase) ProcessExpense(message, sender string) (string, error) {
	userMsg := "Sender: " + sender + "\nFecha de hoy: " + time.Now().Format(domain.DateFormatISO) + "\nMensaje: " + message

	var expense domain.ParsedExpense
	if err := uc.ai.CompleteJSON(expenseSystemPrompt, userMsg, &expense); err != nil {
		return "", domain.Wrapf(domain.ErrFinanceParseExpense, err)
	}

	if expense.Date == "" {
		expense.Date = time.Now().Format(domain.DateFormatISO)
	}

	if expense.PaidBy == "" {
		expense.PaidBy = sender
	}

	if uc.repo != nil {
		if err := uc.repo.SaveExpense(expense); err != nil {
			return "", domain.Wrapf(domain.ErrFinanceWriteSheets, err)
		}
	}

	return FormatExpenseResponse(expense), nil
}

// FormatExpenseResponse builds a user-friendly confirmation message from a parsed expense.
func FormatExpenseResponse(e domain.ParsedExpense) string {
	emoji := domain.CategoryEmoji[e.Category]
	if emoji == "" {
		emoji = domain.DefaultEmoji
	}

	var amountStr string
	if e.AmountUSD > 0 {
		amountStr = fmt.Sprintf("US$%.0f", e.AmountUSD)
	} else {
		amountStr = "$" + FormatNumber(e.Amount)
	}

	date, _ := time.Parse(domain.DateFormatISO, e.Date)
	dateStr := date.Format(domain.DateFormatDisplay)

	return fmt.Sprintf("%s Anotado!\n%s — %s\n%s pago %s el %s",
		emoji, e.Category, e.Description, e.PaidBy, amountStr, dateStr)
}

// FormatNumber formats a number with comma-separated thousands (e.g. 5000 → "5,000").
func FormatNumber(n float64) string {
	s := fmt.Sprintf("%.0f", n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
		if len(s) > remainder {
			result.WriteString(",")
		}
	}
	for i := remainder; i < len(s); i += 3 {
		if i > remainder {
			result.WriteString(",")
		}
		result.WriteString(s[i : i+3])
	}
	return result.String()
}
