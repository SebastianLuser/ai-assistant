package controller

import (
	"net/http"

	"jarvis/pkg/domain"
	"jarvis/pkg/usecase"
	"jarvis/web"
)

type FinanceController struct {
	usecase *usecase.FinanceUseCase
}

func NewFinanceController(uc *usecase.FinanceUseCase) *FinanceController {
	return &FinanceController{usecase: uc}
}

func (c *FinanceController) GetSummary(req web.Request) web.Response {
	period, _ := req.Query(domain.QueryParamPeriod)
	if period == "" {
		period = domain.PeriodMonth
	}

	result, err := c.usecase.GetSummary(period)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.FinanceSummaryResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, result)
}

func (c *FinanceController) PostExpense(req web.Request) web.Response {
	var payload domain.ExpenseRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ExpenseResponse{Error: "invalid request body"})
	}

	if payload.Message == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.ExpenseResponse{Error: "message is required"})
	}

	if payload.Sender == "" {
		payload.Sender = domain.DefaultSender
	}

	response, err := c.usecase.ProcessExpense(payload.Message, payload.Sender)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.ExpenseResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.ExpenseResponse{Success: true, Response: response})
}
