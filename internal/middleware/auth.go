package middleware

import (
	"net/http"

	"jarvis/web"
)

// WebhookAuth validates the X-Webhook-Secret header against the expected secret.
func WebhookAuth(secret string) web.Interceptor {
	return func(req web.InterceptedRequest) web.Response {
		if secret == "" {
			return req.Next()
		}

		headers, ok := req.Header("X-Webhook-Secret")
		if !ok || len(headers) == 0 || headers[0] != secret {
			return web.NewJSONResponse(http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
		}

		return req.Next()
	}
}
