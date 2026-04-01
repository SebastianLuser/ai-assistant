# Testing Guide

## Requisitos

- **Go 1.24+**

### Correr tests

```bash
# Todos
go test ./...

# Con verbose
go test -v ./...

# Con coverage
go test -cover ./...

# Coverage report HTML
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Un paquete especifico
go test -v ./pkg/domain/...
go test -v ./pkg/controller/...
go test -v ./pkg/usecase/...
go test -v ./pkg/service/...
go test -v ./clients/...
go test -v ./internal/hooks/...
go test -v ./internal/skills/...
go test -v ./internal/middleware/...
```

En Docker:

```bash
docker run --rm -v $(pwd):/app -w /app golang:1.25 go test ./...
```

## Estructura de tests

```
clients/
└── claude_test.go            ← ClaudeClient (headers, body, errors, JSON)

pkg/domain/
├── errors_test.go            ← Wrap, Wrapf, Is, Unwrap, sentinels
└── validate_test.go          ← Validate() en todos los request structs

pkg/controller/
├── finance_test.go           ← PostExpense (JSON, empty, default sender)
├── memory_test.go            ← PostNote, GetSearch, DeleteNote
├── conversation_test.go      ← PostChat (JSON, empty message)
├── calendar_test.go          ← CreateEvent (validation, dates, end>start)
├── obsidian_test.go          ← WriteNote, ReadNote (path traversal, absolute)
├── links_test.go             ← PostLink (URL format, scheme, empty)
├── todoist_test.go           ← CreateTask (content, date format)
├── habits_test.go            ← PostLog, GetStreak, GetToday
├── notion_test.go            ← CreatePage, GetPage (validation)
└── projects_test.go          ← GetStatus (missing name)

pkg/usecase/
├── finance_test.go           ← FormatNumber, FormatExpenseResponse
├── memory_test.go            ← FallbackSearch (hybrid, FTS, vector, errors)
├── conversation_test.go      ← Ingest, Assemble, Compact, FormatMessages
├── cron_test.go              ← ShouldRun, Execute, Deliver
├── habits_test.go            ← LogHabit, GetStreak, ListToday
├── links_test.go             ← Save, Search (tag filter, limit, errors)
└── projects_test.go          ← GetStatus (success, FTS error, no tags)

pkg/service/
├── embeddings_test.go        ← AIEmbedder (success, AI error, JSON parse)
└── embeddings_cache_test.go  ← CachedEmbedder (hit, miss, eviction, error)

internal/
├── hooks/hooks_test.go       ← Register, Emit, payloads
├── skills/loader_test.go     ← LoadAll, LoadEnabled, frontmatter, cache
└── middleware/auth_test.go   ← WebhookAuth (valid, invalid, empty)
```

**190+ tests totales.**

## Convenciones

### Patron AAA (Arrange, Act, Assert)

Separar las tres secciones con lineas en blanco. No poner comentarios `// Arrange`, `// Act`, `// Assert`.

```go
func TestStore_Save_ReturnsID(t *testing.T) {
	store := newTestStore(t)

	id, err := store.Save("contenido", []string{"tag"}, []float64{0.1})

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}
```

### Assertions

Siempre usar `testify`. Nunca `if` + `t.Fatal`.

```go
require.NoError(t, err)
assert.Equal(t, expected, actual)
assert.True(t, condition)
assert.Len(t, slice, 3)
```

- `require.*` para precondiciones (falla inmediatamente si no se cumple)
- `assert.*` para verificaciones (continua ejecutando)

### Comparacion de errores

Siempre `errors.Is()` para verificar tipo. `assert.Equal()` con el error string exacto para verificar mensaje. Nunca `assert.Contains()` en errors.

```go
// Bien
assert.True(t, errors.Is(err, domain.ErrValidation))
assert.Equal(t, "validation error: path is required", err.Error())

// Mal
assert.Contains(t, err.Error(), "path")
assert.Equal(t, "store failure", err.Error()) // sin errors.Is
```

### Expected errors como constantes

Definir los mensajes esperados como constantes al inicio del archivo:

```go
const (
	errPathRequired  = "validation error: path is required"
	errPathTraversal = "validation error: path must not contain '..'"
)

func TestObsidian_Validate_EmptyPath(t *testing.T) {
	r := domain.ObsidianNoteRequest{Path: "", Content: "x"}

	err := r.Validate()

	assert.True(t, errors.Is(err, domain.ErrValidation))
	assert.Equal(t, errPathRequired, err.Error())
}
```

### Verificar error en response body de controllers

Usar `errorFromBody()` helper para deserializar JSON y comparar el campo error exacto:

```go
func errorFromBody(t *testing.T, body []byte) string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	v, _ := m["error"].(string)
	return v
}

func TestController_CreateEvent_EmptySummary(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(noSummaryBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: summary is required", errorFromBody(t, resp.Body))
}
```

### Naming

```go
// Formato: Test<Struct>_<Method>_<Scenario>
func TestHabitUseCase_LogHabit_Success(t *testing.T) {}
func TestObsidianController_WriteNote_PathTraversal(t *testing.T) {}
func TestCachedEmbedder_Eviction(t *testing.T) {}
```

### Constantes y variables de test

Siempre al inicio del archivo, antes de helpers y tests.

```go
const (
	testSessionID = "test-session"
	testContent   = "El pool de cartas tiene 40 cartas"
)

var (
	errStoreFailure = errors.New("store failure")
	testTags        = []string{"game", "cards"}
	testEmbedding   = []float64{0.1, 0.2, 0.3}
)
```

### Table-driven tests

Para escenarios repetitivos:

```go
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "0"},
		{"thousands", 5000, "5,000"},
		{"millions", 1500000, "1,500,000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNumber(tt.input)

			assert.Equal(t, tt.expected, result)
		})
	}
}
```

### Mocks

Los mocks estan en `test/mocks.go` y usan `testify/mock`.

**MockMemoryService** — mockea `service.MemoryService`:

```go
repo := new(test.MockMemoryService)
repo.On("Save", "content", []string{"tag"}, mock.Anything).Return(int64(1), nil)
repo.On("LogHabit", "exercise").Return(nil)
repo.On("GetHabitStreak", "exercise").Return(5, 30, nil)

// Verificar que se llamo
repo.AssertExpectations(t)

// Verificar que NO se llamo
repo.AssertNotCalled(t, "ReplaceConversation")
```

**MockEmbedder** — mockea `service.Embedder`:

```go
embedder := new(test.MockEmbedder)
embedder.On("Embed", "texto").Return([]float64{0.1, 0.2}, nil)
```

**MockAIProvider** — mockea `domain.AIProvider`:

```go
ai := new(test.MockAIProvider)
ai.On("Complete", "system", "user msg").Return("response", nil)
```

**MockClaudeServer** — levanta un HTTP server fake:

```go
srv, ai := test.NewMockClaudeServer(
    test.ClaudeResponse{Text: `{"amount": 5000}`},
)
defer srv.Close()
```

**MockRequest** — mockea `web.Request` para controller tests:

```go
req := test.NewMockRequest().
    WithBody(`{"name":"exercise"}`).
    WithQuery("name", "exercise").
    WithParam("id", "5")
```

### Helpers

Cada archivo de test puede tener helpers privados al inicio:

```go
func newTestStore(t *testing.T) *PGMemoryService {
	t.Helper()
	// Use a test PostgreSQL instance
	store, err := NewPGMemoryService(os.Getenv("TEST_POSTGRES_DSN"))
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}
```

## CI (GitHub Actions)

```yaml
name: tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go test -race -cover ./...
```
