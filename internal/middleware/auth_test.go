package middleware

import (
	"context"
	"net/http"
	"testing"

	"jarvis/test"
	"jarvis/web"

	"github.com/stretchr/testify/assert"
)

const testSecret = "my-webhook-secret"

type mockInterceptedRequest struct {
	*test.MockRequest
	nextCalled bool
}

func (m *mockInterceptedRequest) Next() web.Response {
	m.nextCalled = true
	return web.NewJSONResponse(http.StatusOK, nil)
}

func (m *mockInterceptedRequest) Writer() http.ResponseWriter { return nil }
func (m *mockInterceptedRequest) Apply(_ context.Context)     {}

func runAuth(secret string, req *test.MockRequest) (web.Response, bool) {
	fn := WebhookAuth(secret)
	ir := &mockInterceptedRequest{MockRequest: req}
	resp := fn(ir)
	return resp, ir.nextCalled
}

func TestWebhookAuth_EmptySecret_PassesThrough(t *testing.T) {
	req := test.NewMockRequest()

	_, nextCalled := runAuth("", req)

	assert.True(t, nextCalled)
}

func TestWebhookAuth_ValidSecret_PassesThrough(t *testing.T) {
	req := test.NewMockRequest()
	req.HeadersMap = map[string][]string{"X-Webhook-Secret": {testSecret}}

	_, nextCalled := runAuth(testSecret, req)

	assert.True(t, nextCalled)
}

func TestWebhookAuth_InvalidSecret_Returns401(t *testing.T) {
	req := test.NewMockRequest()
	req.HeadersMap = map[string][]string{"X-Webhook-Secret": {"wrong"}}

	resp, nextCalled := runAuth(testSecret, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, resp.Status)
}

func TestWebhookAuth_MissingHeader_Returns401(t *testing.T) {
	req := test.NewMockRequest()

	resp, nextCalled := runAuth(testSecret, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, resp.Status)
}
