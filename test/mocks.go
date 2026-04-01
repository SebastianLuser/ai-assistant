package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"

	"jarvis/clients"
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
	"jarvis/web"

	"github.com/stretchr/testify/mock"
)

// MockMemoryService mocks service.MemoryService.
type MockMemoryService struct {
	mock.Mock
}

func (m *MockMemoryService) Save(content string, tags []string, embedding []float64) (int64, error) {
	args := m.Called(content, tags, embedding)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMemoryService) Search(queryEmbedding []float64, limit int) ([]domain.Memory, error) {
	args := m.Called(queryEmbedding, limit)
	return args.Get(0).([]domain.Memory), args.Error(1)
}

func (m *MockMemoryService) SearchFTS(query string, limit int) ([]domain.Memory, error) {
	args := m.Called(query, limit)
	return args.Get(0).([]domain.Memory), args.Error(1)
}

func (m *MockMemoryService) SearchHybrid(query string, queryEmbedding []float64, limit int, vecWeight, ftsWeight float64) ([]domain.Memory, error) {
	args := m.Called(query, queryEmbedding, limit, vecWeight, ftsWeight)
	return args.Get(0).([]domain.Memory), args.Error(1)
}

func (m *MockMemoryService) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMemoryService) SaveConversation(sessionID, role, content string) error {
	args := m.Called(sessionID, role, content)
	return args.Error(0)
}

func (m *MockMemoryService) LoadConversation(sessionID string, limit int) ([]domain.ConversationMessage, error) {
	args := m.Called(sessionID, limit)
	return args.Get(0).([]domain.ConversationMessage), args.Error(1)
}

func (m *MockMemoryService) ClearConversation(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockMemoryService) ReplaceConversation(sessionID string, msgs []domain.ConversationMessage) error {
	args := m.Called(sessionID, msgs)
	return args.Error(0)
}

func (m *MockMemoryService) LogHabit(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockMemoryService) GetHabitStreak(name string) (int, int, error) {
	args := m.Called(name)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockMemoryService) ListHabitsToday() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockMemoryService) ListExpenses(from, to string) ([]domain.Expense, error) {
	args := m.Called(from, to)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockMemoryService) PruneSessions(olderThanDays int) (int64, error) {
	args := m.Called(olderThanDays)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMemoryService) Close() error {
	args := m.Called()
	return args.Error(0)
}

var _ service.MemoryService = (*MockMemoryService)(nil)

// MockEmbedder mocks service.Embedder.
type MockEmbedder struct {
	mock.Mock
}

func (m *MockEmbedder) Embed(text string) ([]float64, error) {
	args := m.Called(text)
	return args.Get(0).([]float64), args.Error(1)
}

var _ service.Embedder = (*MockEmbedder)(nil)

// MockAIProvider mocks domain.AIProvider using testify/mock.
type MockAIProvider struct {
	mock.Mock
}

func (m *MockAIProvider) Complete(system, userMessage string, opts ...domain.CompletionOption) (string, error) {
	args := m.Called(system, userMessage)
	return args.String(0), args.Error(1)
}

func (m *MockAIProvider) CompleteMessages(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, error) {
	args := m.Called(system, messages)
	return args.String(0), args.Error(1)
}

func (m *MockAIProvider) CompleteJSON(system, userMessage string, target any, opts ...domain.CompletionOption) error {
	args := m.Called(system, userMessage, target)
	return args.Error(0)
}

var _ domain.AIProvider = (*MockAIProvider)(nil)

// ClaudeResponse is a helper for building mock Claude API responses.
type ClaudeResponse struct {
	Text string
}

// NewMockClaudeServer creates a mock HTTP server that returns Claude-format responses.
// It returns domain.AIProvider for use in tests.
func NewMockClaudeServer(responses ...ClaudeResponse) (*httptest.Server, domain.AIProvider) {
	idx := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp ClaudeResponse
		if idx < len(responses) {
			resp = responses[idx]
			idx++
		} else if len(responses) > 0 {
			resp = responses[len(responses)-1]
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": resp.Text},
			},
			"usage": map[string]int{"input_tokens": 10, "output_tokens": 10},
		})
	}))

	client := clients.NewClaudeClientWithBaseURL("test-key", "test-model", srv.URL)
	return srv, client
}

// MockRequest implements web.Request for controller tests.
type MockRequest struct {
	Ctx        context.Context
	ParamsMap  map[string]string
	QueriesMap map[string]string
	HeadersMap map[string][]string
	BodyStr    string
}

func NewMockRequest() *MockRequest {
	return &MockRequest{
		Ctx:        context.Background(),
		ParamsMap:  make(map[string]string),
		QueriesMap: make(map[string]string),
	}
}

func (m *MockRequest) WithParam(key, value string) *MockRequest {
	m.ParamsMap[key] = value
	return m
}

func (m *MockRequest) WithQuery(key, value string) *MockRequest {
	m.QueriesMap[key] = value
	return m
}

func (m *MockRequest) WithBody(body string) *MockRequest {
	m.BodyStr = body
	return m
}

func (m *MockRequest) WithHeader(key, value string) *MockRequest {
	if m.HeadersMap == nil {
		m.HeadersMap = make(map[string][]string)
	}
	m.HeadersMap[key] = append(m.HeadersMap[key], value)
	return m
}

func (m *MockRequest) Context() context.Context { return m.Ctx }
func (m *MockRequest) Raw() *http.Request {
	r := &http.Request{Header: make(http.Header)}
	if m.HeadersMap != nil {
		for k, v := range m.HeadersMap {
			r.Header[k] = v
		}
	}
	return r
}
func (m *MockRequest) DeclaredPath() string { return "" }
func (m *MockRequest) Params() []web.Param                                { return nil }
func (m *MockRequest) Queries() url.Values                                { return nil }
func (m *MockRequest) Headers() http.Header                               { return nil }
func (m *MockRequest) Body() io.ReadCloser                                { return io.NopCloser(bytes.NewBufferString(m.BodyStr)) }
func (m *MockRequest) Header(key string) ([]string, bool) {
	if m.HeadersMap != nil {
		v, ok := m.HeadersMap[key]
		return v, ok
	}
	return nil, false
}
func (m *MockRequest) FormFile(_ string) (*multipart.FileHeader, error)   { return nil, nil }
func (m *MockRequest) FormValue(_ string) (string, bool)                  { return "", false }
func (m *MockRequest) MultipartForm() (*multipart.Form, error)            { return nil, nil }

func (m *MockRequest) Param(key string) (string, bool) {
	v, ok := m.ParamsMap[key]
	return v, ok
}

func (m *MockRequest) Query(key string) (string, bool) {
	v, ok := m.QueriesMap[key]
	return v, ok
}

var _ web.Request = (*MockRequest)(nil)

// MockWhatsAppSender mocks domain.WhatsAppSender.
type MockWhatsAppSender struct {
	mock.Mock
}

func (m *MockWhatsAppSender) SendTextMessage(to, text string) error {
	args := m.Called(to, text)
	return args.Error(0)
}

func (m *MockWhatsAppSender) MarkAsRead(messageID string) error {
	args := m.Called(messageID)
	return args.Error(0)
}

var _ domain.WhatsAppSender = (*MockWhatsAppSender)(nil)

// MockFinanceService mocks service.FinanceService.
type MockFinanceService struct {
	mock.Mock
}

func (m *MockFinanceService) SaveExpense(expense domain.ParsedExpense) error {
	args := m.Called(expense)
	return args.Error(0)
}

var _ service.FinanceService = (*MockFinanceService)(nil)

// MockTranscriber mocks domain.Transcriber.
type MockTranscriber struct {
	mock.Mock
}

func (m *MockTranscriber) Transcribe(audioData []byte, mimeType string) (string, error) {
	args := m.Called(audioData, mimeType)
	return args.String(0), args.Error(1)
}

var _ domain.Transcriber = (*MockTranscriber)(nil)

// MockMediaChannel implements domain.Channel and domain.MediaDownloader.
type MockMediaChannel struct {
	MockChannel
}

func (m *MockMediaChannel) DownloadMedia(mediaID string) ([]byte, string, error) {
	args := m.Called(mediaID)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

var _ domain.MediaDownloader = (*MockMediaChannel)(nil)

// MockChannel mocks domain.Channel for testing the MessageRouter.
type MockChannel struct {
	mock.Mock
	ChannelName string
}

func (m *MockChannel) Name() string { return m.ChannelName }

func (m *MockChannel) SendMessage(to, text string) error {
	args := m.Called(to, text)
	return args.Error(0)
}

func (m *MockChannel) AckMessage(messageID string) error {
	args := m.Called(messageID)
	return args.Error(0)
}

var _ domain.Channel = (*MockChannel)(nil)

func NewMockClaudeServerError(statusCode int, errType, errMsg string) (*httptest.Server, domain.AIProvider) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprintf(w, `{"error":{"type":"%s","message":"%s"}}`, errType, errMsg)
	}))

	client := clients.NewClaudeClientWithBaseURL("test-key", "test-model", srv.URL)
	return srv, client
}
