package middleware

import (
	"jarvis/internal/tracing"
	"jarvis/web"
)

func TraceID() web.Interceptor {
	return func(req web.InterceptedRequest) web.Response {
		tid := tracing.NewTraceID()
		ctx := tracing.WithTraceID(req.Context(), tid)
		req.Apply(ctx)
		resp := req.Next()
		resp.Headers.Set("X-Trace-ID", tid)
		return resp
	}
}
