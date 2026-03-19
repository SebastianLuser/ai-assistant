package domain

const (
	DateFormatISO     = "2006-01-02"
	DateFormatDisplay = "02/01/2006"

	CategorySupermercado    = "Supermercado"
	CategoryRestaurante     = "Restaurante"
	CategoryTransporte      = "Transporte"
	CategoryServicios       = "Servicios"
	CategorySalud           = "Salud"
	CategoryRopa            = "Ropa"
	CategoryEntretenimiento = "Entretenimiento"
	CategoryEducacion       = "Educacion"
	CategoryHogar           = "Hogar"
	CategoryOtro            = "Otro"

	DefaultSender = "Sebas"
	DefaultEmoji  = "\U0001F4E6"
)

type ParsedExpense struct {
	Amount      float64 `json:"amount"`
	AmountUSD   float64 `json:"amount_usd"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	PaidBy      string  `json:"paid_by"`
	Date        string  `json:"date"`
}

const (
	maxExpenseMessageLen = 5000
)

type ExpenseRequest struct {
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

// Validate checks that an expense request is valid.
func (r ExpenseRequest) Validate() error {
	if r.Message == "" {
		return Wrap(ErrValidation, "message is required")
	}
	if len(r.Message) > maxExpenseMessageLen {
		return Wrap(ErrValidation, "message exceeds maximum length")
	}
	return nil
}

type ExpenseResponse struct {
	Success  bool   `json:"success"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
}

const (
	QueryParamPeriod = "period"

	PeriodWeek  = "week"
	PeriodMonth = "month"
	PeriodYear  = "year"
)

type FinanceSummaryResponse struct {
	Success    bool               `json:"success"`
	Summary    string             `json:"summary,omitempty"`
	Period     string             `json:"period,omitempty"`
	TotalARS   float64            `json:"total_ars,omitempty"`
	TotalUSD   float64            `json:"total_usd,omitempty"`
	ByCategory map[string]float64 `json:"by_category,omitempty"`
	Error      string             `json:"error,omitempty"`
}

type Expense struct {
	ID          int64   `json:"id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	AmountUSD   float64 `json:"amount_usd"`
	PaidBy      string  `json:"paid_by"`
}

var CategoryEmoji = map[string]string{
	CategorySupermercado:    "\U0001F6D2",
	CategoryRestaurante:     "\U0001F354",
	CategoryTransporte:      "\U0001F697",
	CategoryServicios:       "\U0001F4F1",
	CategorySalud:           "\U0001F48A",
	CategoryRopa:            "\U0001F455",
	CategoryEntretenimiento: "\U0001F3AE",
	CategoryEducacion:       "\U0001F4DA",
	CategoryHogar:           "\U0001F3E0",
	CategoryOtro:            DefaultEmoji,
}
