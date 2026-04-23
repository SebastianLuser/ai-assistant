# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

Jarvis — microservicio Go que funciona como backend de un asistente personal accesible por WhatsApp. Parsea gastos, guarda notas con busqueda semantica, mantiene memoria de conversaciones, y ejecuta acciones proactivas via cron jobs.

## Tech Stack

- **Language**: Go 1.24+
- **Framework**: Gin (via abstraccion framework-agnostic en `web/`)
- **Database**: PostgreSQL (tsvector con diccionario spanish para FTS)
- **AI**: Claude API (Anthropic) o OpenAI para parseo de lenguaje natural y embeddings
- **Integrations**: Google Sheets, Google Calendar, Notion, Obsidian, WhatsApp Business (direct webhook), GitHub, Jira, Spotify, Todoist, Gmail, ClickUp
- **Deploy**: Docker multi-stage, Contabo VPS + Cloudflare + Coolify

## Development Commands

```bash
make run          # go run ./cmd
make build        # CGO_ENABLED=0 go build -o jarvis ./cmd
make test         # go test ./...
make test-race    # go test -race (requires CGO_ENABLED=1)
make test-cover   # test + coverage report
make vet          # go vet ./...
make docker       # docker compose up -d --build jarvis
make docker-all   # docker compose up -d --build (full stack)
make docker-down  # docker compose down
make clean        # remove binaries and coverage
```

## Architecture

Layered architecture siguiendo el patron del api-template:

```
cmd/                    ← Entry point + wiring
├── main.go             ← main() → NewApp → Run (3 lineas)
├── server.go           ← App struct, NewApp(), Run(), Close(), signal handling
├── clients.go          ← Clients struct + NewClients() + factory functions
├── controller.go       ← Controllers struct + NewControllers() con nil-checks
├── service.go          ← NewMemoryService, NewFinanceService, NewEmbedder
├── scheduler.go        ← NewScheduler() con los 4 cron jobs
└── routes.go           ← registerRoutes(), setupRoutes(), middlewareMapper()

clients/                ← External API clients (single package, prefixed types)
├── claude.go           ← ClaudeClient (Anthropic API)
├── openai.go           ← OpenAIClient (OpenAI API)
├── sheets.go           ← SheetsClient (Google Sheets)
├── calendar.go         ← CalendarClient (Google Calendar)
├── notion.go           ← NotionClient (Notion API)
├── obsidian.go         ← ObsidianVault (local filesystem)
├── whatsapp.go         ← WhatsAppClient (WhatsApp Business Cloud API)
├── github.go           ← GitHubClient (GitHub API)
├── jira.go             ← JiraClient (Jira API)
├── spotify.go          ← SpotifyClient (Spotify API)
├── todoist.go          ← TodoistClient (Todoist API)
├── gmail.go            ← GmailClient (Gmail API)
├── clickup.go          ← ClickUpClient (ClickUp API)
└── common.go           ← Shared HTTP constants

pkg/
├── domain/             ← Models, types, sentinel errors, constants, Validate() methods
│   ├── catalog.go      ← CatalogEntry (usage/success/error tracking per tool/skill/agent)
│   └── dryrun.go       ← PreviewResult for dry-run mode
├── controller/         ← HTTP handlers (decode → validate → usecase → response)
│   ├── catalog.go      ← Catalog API controller
│   ├── health.go       ← Health API controller (detailed integration status)
│   └── skills_qa.go    ← Skills QA report + validate controller
├── usecase/            ← Business logic (ProcessExpense, FallbackSearch, Compact, Scheduler)
│   ├── ratelimit.go    ← Sliding window rate limiter per tool
│   ├── promptcache.go  ← In-memory prompt cache with TTL
│   └── health.go       ← Health checker with periodic integration checks + catalog summary
└── service/            ← Data access interfaces + implementations (Postgres, Sheets)
    ├── catalog.go      ← CatalogService interface + PGCatalogService + NullCatalogService
    └── sqldata/        ← SQL queries as embedded .sql files
        ├── postgres/   ← insert/, select/, delete/ con archivos .sql individuales
        └── queries.go  ← Constantes Postgres via go:embed

internal/
├── agents/             ← Agent definition loader (YAML frontmatter .md)
├── hooks/              ← Event hook system (Register/Emit)
│   ├── config.go       ← Hook YAML config loader
│   └── external.go     ← External webhook/command hooks
├── middleware/          ← Trace ID + webhook auth interceptors
├── profiles/           ← YAML profile loader (skill/tool/agent/rule permissions)
├── rules/              ← Rule engine (triggers: tags, time_range, channel, day_of_week)
├── skills/             ← Skill loader (YAML frontmatter + markdown)
│   ├── deps.go         ← Skill dependency validation
│   ├── triggers.go     ← Post-tool skill triggers
│   └── qa.go           ← Skill QA rubric validation
└── tracing/            ← Context-based trace propagation + slog integration

test/
└── mocks.go            ← MockMemoryService, MockEmbedder, MockAIProvider, MockWhatsAppSender, MockRequest, MockClaudeServer

web/                    ← Framework-agnostic HTTP abstractions (from template)
boot/                   ← Server bootstrap (from template)
config/
├── profiles/           ← Runtime profile YAML files (full, work, personal, study)
└── hooks.yaml          ← External hook definitions (command/webhook)
agents/                 ← Agent definition markdown files (assistant, dev, finance, planner, study)
rules/                  ← Runtime rule markdown files (dev, finance, habits, morning)
db/                     ← Migrator + SQL migrations (postgres)
skills/                 ← Skill markdown files with YAML frontmatter
```

### Data Flow

```
HTTP Request → Controller → UseCase → Service (DB/API) → UseCase → Controller → HTTP Response
```

### Key Patterns

- **App struct**: `cmd/server.go` define `App` con `NewApp()`, `Run()`, `Close()`. `main()` son 3 lineas
- **Clients struct**: Agrupa todos los clientes externos en `cmd/clients.go`
- **Controllers struct**: Agrupa todos los controllers con nil-checks en `cmd/controller.go`
- **Framework-agnostic handlers**: Controllers usan `web.Request` / `web.Response`, no `*gin.Context`
- **Optional integrations**: Nil clients skip registration (Notion, Obsidian, Calendar, etc.)
- **Domain validation**: Request structs tienen metodos `Validate()` para datos (path traversal, fechas, URLs, max length)
- **SQL as files**: Queries en archivos `.sql` individuales organizados por operacion, embebidos via `go:embed`
- **Skills as config**: Markdown files en `skills/` inyectados como Claude system prompt
- **Context Engine**: Ingest → Assemble → Compact (multi-stage with fallback)
- **Hybrid search**: Vector + FTS5 with fallback chain
- **Graceful shutdown**: SIGINT/SIGTERM handling con deferred cleanup
- **Channel-agnostic MessageRouter**: Intent-based routing (expense/note/chat) via `domain.Channel` interface — same router for WhatsApp, Telegram, CLI, etc.
- **AI failover**: `FailoverProvider` wraps primary + fallback AIProvider — if Claude fails, falls back to OpenAI automatically
- **Skills auto-generated**: Skills can be created at runtime via `POST /api/skills` — the AI or user can create new skills from chat
- **Webhook triggers**: Cron jobs can be triggered manually via `POST /api/triggers/job/:job_id`
- **Agent orchestration**: `AgentOrchestrator` delegates to sub-agents (assistant, dev, finance, planner, study) with own system prompts loaded from `agents/`
- **Runtime profiles**: YAML profiles in `config/profiles/` control which skills/tools/agents/rules are active per context (full, work, personal, study)
- **Rules engine**: Markdown rules in `rules/` with YAML frontmatter triggers (tags, time_range, channel, day_of_week) — matched and injected into prompts at runtime
- **Tool rate limiting**: Sliding window per-tool rate limits prevent abuse (`toolLimit` in `usecase/ratelimit.go`)
- **Prompt caching**: In-memory TTL cache keyed by profile+channel+tags avoids recomputing system prompts (`PromptCache`)
- **Catalog tracking**: Automatic usage/success/error counting per tool/skill/agent via `CatalogService` + `catalog` table in Postgres
- **Tracing**: Context-propagated TraceID across request lifecycle via `internal/tracing` — injected by `middleware.TraceID()`, logged with `tracing.Logger(ctx)`
- **External hooks**: YAML-configured hooks (`config/hooks.yaml`) fire shell commands or HTTP webhooks on internal events
- **Skill QA**: Rubric-based validation scores skill quality (required fields, content length, WIP markers) via `skills.ValidateSkill()`
- **Skill dependencies**: `DependencyChecker` validates that required integrations are available before enabling a skill
- **Dry-run mode**: Tools configured via `DRY_RUN_TOOLS` env return `PreviewResult` instead of executing side-effects

## API Endpoints

| Method | Route | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/ping` | Ping/pong (built-in) |
| GET | `/webhook/whatsapp` | Meta webhook verification (challenge/response) |
| POST | `/webhook/whatsapp` | Incoming WhatsApp messages (intent detection + routing) |
| POST | `/api/finance/expense` | Parse expense and save to Sheets |
| GET | `/api/finance/summary` | Financial summary by period |
| POST | `/api/memory/note` | Save note with embedding |
| GET | `/api/memory/search?q=&mode=&limit=` | Search notes (fts/vector/hybrid/fallback) |
| DELETE | `/api/memory/note/:id` | Delete note |
| POST | `/api/chat` | Chat with persistent memory |
| POST | `/api/notion/page` | Create Notion page |
| GET | `/api/notion/page/:id` | Get Notion page |
| GET | `/api/obsidian/note?path=` | Read Obsidian note |
| POST | `/api/obsidian/note` | Write Obsidian note |
| GET | `/api/obsidian/notes?dir=` | List Obsidian notes |
| GET | `/api/obsidian/search?q=` | Search Obsidian vault |
| GET | `/api/calendar/today` | Today's events |
| POST | `/api/calendar/event` | Create calendar event |
| GET | `/api/github/repos` | List GitHub repos |
| GET | `/api/github/:owner/:repo/issues` | List issues |
| POST | `/api/github/:owner/:repo/issues` | Create issue |
| GET | `/api/github/:owner/:repo/pulls` | List pull requests |
| GET | `/api/jira/my-issues` | My Jira issues |
| GET | `/api/jira/issue/:key` | Get Jira issue |
| POST | `/api/jira/issue` | Create Jira issue |
| GET | `/api/spotify/playing` | Currently playing track |
| POST | `/api/spotify/play` | Resume playback |
| POST | `/api/spotify/pause` | Pause playback |
| POST | `/api/spotify/next` | Next track |
| GET | `/api/todoist/tasks` | List tasks |
| POST | `/api/todoist/task` | Create task |
| POST | `/api/todoist/task/:id/complete` | Complete task |
| GET | `/api/gmail/unread` | List unread emails |
| GET | `/api/gmail/message/:id` | Get email |
| GET | `/api/clickup/tasks` | List ClickUp tasks |
| GET | `/api/clickup/task/:id` | Get ClickUp task |
| POST | `/api/clickup/task` | Create ClickUp task |
| POST | `/api/habits/log` | Log habit |
| GET | `/api/habits/streak` | Get habit streak |
| GET | `/api/habits/today` | Today's habits |
| POST | `/api/links` | Save link |
| GET | `/api/links/search` | Search links |
| GET | `/api/projects/:name/status` | Project status summary |
| GET | `/api/figma/file/:file_key` | Get Figma file metadata |
| GET | `/api/figma/file/:file_key/nodes?ids=` | Get specific nodes |
| GET | `/api/figma/file/:file_key/images?ids=&format=&scale=` | Render nodes as images |
| GET | `/api/figma/file/:file_key/comments` | List file comments |
| GET | `/api/figma/file/:file_key/components` | List file components |
| GET | `/api/figma/project/:project_id/files` | List project files |
| GET | `/api/skills` | List enabled skills |
| POST | `/api/skills` | Create a new skill |
| GET | `/api/skills/report` | QA report for all loaded skills |
| POST | `/api/skills/validate` | Validate a single skill against QA rubric |
| GET | `/api/catalog` | List all catalog entries (usage stats) |
| GET | `/api/catalog/:name?type=` | Get catalog entry by name and type |
| GET | `/api/health` | Detailed health with integration status + catalog summary |
| GET | `/api/triggers/jobs` | List registered cron jobs |
| POST | `/api/triggers/job/:job_id` | Manually trigger a cron job |

## Environment Variables

See `.env.example` for the full list with defaults. New config vars:

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENTS_DIR` | `agents` | Directory for agent definition markdown files |
| `RULES_DIR` | `rules` | Directory for runtime rule markdown files |
| `PROFILES_DIR` | `config/profiles` | Directory for YAML profile files |
| `DEFAULT_PROFILE` | `full` | Active profile name at startup |
| `HOOKS_CONFIG_FILE` | `config/hooks.yaml` | Path to external hooks YAML config |
| `DRY_RUN_TOOLS` | _(empty)_ | Comma-separated tool names that run in preview mode |

## Naming Conventions

- **Packages**: lowercase, single word (`domain`, `controller`, `usecase`, `service`)
- **Types/Structs**: PascalCase (`FinanceController`, `MemoryUseCase`, `ParsedExpense`)
- **Functions/Methods**: PascalCase for exported, camelCase for unexported
- **File names**: snake_case (`memory_postgres.go`, `finance_sheets.go`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **JSON fields**: snake_case (`session_id`, `paid_by`, `amount_usd`)
- **Sentinel errors**: `Err` prefix PascalCase (`ErrStoreOpen`, `ErrClaudeAPI`, `ErrValidation`)
- **Client types**: Prefijo del servicio (`ClaudeClient`, `SheetsClient`, `GitHubIssue`, `CalendarEvent`)
- **SQL constants**: PascalCase (`SaveMemory`, `LoadConversation`, etc.)

## Error Handling

- All sentinel errors in `pkg/domain/errors.go`
- Use `domain.Wrap(sentinel, detail)` or `domain.Wrapf(sentinel, cause)` to wrap errors
- Compare errors with `errors.Is()`, never string comparison
- Controllers return typed response structs, never `map[string]any`
- Validation errors use `domain.ErrValidation` sentinel

## Validation

Request structs en `pkg/domain/` tienen metodos `Validate() error` para validaciones de datos:

- **Path traversal**: `ObsidianNoteRequest`, `ValidatePath()` — rechazan `..` y paths absolutos
- **Date format**: `CalendarEventRequest` (RFC3339), `TodoistCreateTaskRequest` (YYYY-MM-DD)
- **Cross-field**: `CalendarEventRequest` — end debe ser posterior a start
- **URL format**: `LinkSaveRequest` — solo http/https con host valido
- **Max length**: `ChatRequest`, `NoteRequest` — previenen payloads excesivos

Controllers llaman `payload.Validate()` y devuelven 400 con el error message.

## Testing

- 190+ tests across 20+ test files
- `testify` for assertions (`assert`, `require`)
- AAA pattern without comments, separated by blank lines
- Table-driven tests for multiple scenarios
- Mocks in `test/mocks.go` (MockMemoryService, MockEmbedder, MockAIProvider, MockRequest, MockClaudeServer)
- See `TESTING.md` for full guide

## Important Notes

- `web/` and `boot/` are the template base — do not modify unless necessary
- Controllers never use `*gin.Context` directly
- All integrations are optional — nil if not configured
- PostgreSQL uses `tsvector` with spanish dictionary for FTS
- SQL queries live in `.sql` files under `pkg/service/sqldata/`, never inline strings
- Skills are hot-reloaded from disk on every chat request
- `credentials.json` (Google service account) never committed to repo
- Cron jobs run in isolated goroutines
