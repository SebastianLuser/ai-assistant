# Jarvis

Microservicio Go que funciona como backend de un asistente personal accesible por WhatsApp. Parsea gastos, guarda notas con busqueda semantica, mantiene memoria de conversaciones, y ejecuta acciones proactivas via cron jobs.

## Features

- **Chat con memoria**: Conversaciones persistentes con compactacion automatica
- **Finanzas**: Parseo de gastos en lenguaje natural, guardado en Google Sheets
- **Notas**: Guardado con embeddings, busqueda hibrida (vector + FTS)
- **Cron jobs**: Briefing diario, alertas de presupuesto, journal via WhatsApp
- **Habitos**: Tracking con streaks
- **Links**: Guardado y busqueda semantica
- **Proyectos**: Status summary con AI
- **Integraciones**: Google Calendar, Notion, Obsidian, GitHub, Jira, Spotify, Todoist, Gmail, ClickUp

## Quick Start

```bash
cp .env.example .env
# Editar .env con las API keys necesarias

make run
# o
go run ./cmd
```

## Stack

| Componente | Tecnologia |
|---|---|
| Language | Go 1.25+ |
| Framework | Gin (via abstraccion en `web/`) |
| Database | PostgreSQL |
| AI | Claude API (Anthropic) o OpenAI |
| Deploy | Docker multi-stage |

## Architecture

```
cmd/         <- Entry point (main 3 lineas) + wiring (App, Clients, Controllers structs)
clients/     <- External API clients (13 integraciones)
pkg/domain/  <- Models, types, errors, validations
pkg/controller/ <- HTTP handlers
pkg/usecase/ <- Business logic
pkg/service/ <- Data access (Postgres, Sheets, embeddings)
internal/    <- Hooks, skills, middleware
```

Detalle completo en [CLAUDE.md](CLAUDE.md).

## Development

```bash
make build        # CGO_ENABLED=0 go build -o jarvis ./cmd
make test         # go test -race ./...
make test-cover   # test + coverage report
make vet          # go vet ./...
make docker       # docker compose up -d --build jarvis
```

## Testing

190+ tests unitarios. Ver [TESTING.md](TESTING.md) para convenciones y guia completa.

```bash
go test ./...
```

## API

Health: `GET /health`

Todos los endpoints bajo `/api/`. Ver tabla completa en [CLAUDE.md](CLAUDE.md#api-endpoints).

## Environment Variables

Ver `.env.example` para la lista completa con defaults. Las integraciones son opcionales — si no se configuran, se desactivan automaticamente.

## Deploy

```bash
make docker-all   # Levanta jarvis + postgres + dependencias
```

## License

Private.
