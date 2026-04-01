package controller

import (
	"net/http"
	"strconv"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/web"
)

const defaultMaxResults = 10

// GmailController handles Gmail API endpoints.
type GmailController struct {
	client *clients.GmailClient
}

// NewGmailController creates a new GmailController.
func NewGmailController(client *clients.GmailClient) *GmailController {
	return &GmailController{client: client}
}

// ListUnread returns unread emails.
func (c *GmailController) ListUnread(req web.Request) web.Response {
	maxResults := defaultMaxResults
	if v, ok := req.Query(domain.QueryParamMaxResults); ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxResults = n
		}
	}

	emails, err := c.client.ListUnread(maxResults)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.GmailListResponse{Error: err.Error()})
	}

	domainEmails := make([]domain.GmailEmail, len(emails))
	for i, e := range emails {
		domainEmails[i] = domain.GmailEmail{
			ID: e.ID, From: e.From, Subject: e.Subject, Snippet: e.Snippet, Date: e.Date,
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.GmailListResponse{
		Success: true, Emails: domainEmails,
	})
}

// GetMessage returns a single email by ID.
func (c *GmailController) GetMessage(req web.Request) web.Response {
	messageID, ok := req.Param(domain.PathParamID)
	if !ok || messageID == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GmailMessageResponse{Error: "id is required"})
	}

	email, err := c.client.GetMessage(messageID)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.GmailMessageResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.GmailMessageResponse{
		Success: true,
		Email: &domain.GmailEmail{
			ID: email.ID, From: email.From, Subject: email.Subject,
			Snippet: email.Snippet, Date: email.Date,
		},
	})
}
