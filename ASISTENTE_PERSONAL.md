# 🦞 Asistente Personal — Diseño Completo del Stack

> Documento de arquitectura, decisiones técnicas y roadmap de implementación.
> Basado en análisis de OpenClaw, investigación de herramientas 2025-2026, y diseño propio.

---

## 📋 Tabla de Contenidos

1. [Visión General](#1-visión-general)
2. [El Problema que Resuelve](#2-el-problema-que-resuelve)
3. [Stack Tecnológico](#3-stack-tecnológico)
4. [Infraestructura — Oracle Cloud](#4-infraestructura--oracle-cloud)
5. [Arquitectura Completa](#5-arquitectura-completa)
6. [Cómo Funcionan las APIs](#6-cómo-funcionan-las-apis)
7. [n8n — El Orquestador](#7-n8n--el-orquestador)
8. [El Microservicio Go](#8-el-microservicio-go)
9. [Lecciones del Análisis de OpenClaw](#9-lecciones-del-análisis-de-openclaw)
10. [Ideas Robadas de OpenClaw para el Microservicio](#10-ideas-robadas-de-openclaw-para-el-microservicio)
11. [Módulos del Microservicio](#11-módulos-del-microservicio)
12. [Flujo Completo de un Mensaje](#12-flujo-completo-de-un-mensaje)
13. [Costo Mensual](#13-costo-mensual)
14. [Roadmap de Setup](#14-roadmap-de-setup)
15. [Decisiones de Diseño y Por Qué](#15-decisiones-de-diseño-y-por-qué)
16. [Comparación con Alternativas](#16-comparación-con-alternativas)

---

## 1. Visión General

El objetivo es construir un **asistente personal propio** que:

- Corre en infraestructura propia (no depende de ningún SaaS)
- Se controla por **WhatsApp**, la app que ya usás todo el día
- Tiene **memoria real** entre conversaciones (no olvida todo al cerrar)
- Ejecuta acciones reales: guarda gastos, crea eventos, toma notas
- Es **proactivo**: manda briefings matutinos sin que lo pidas
- Conecta todos los contextos de tu vida: finanzas, proyectos, facultad, trabajo
- Cuesta **~$5-10/mes** en total (solo Claude API + dominio)

La diferencia con usar Claude.ai directamente es que **este asistente actúa**. No solo responde preguntas, sino que guarda en Sheets, crea eventos en Calendar, escribe en Notion, y te manda resúmenes proactivos. Además tiene memoria persistente de semanas o meses de conversaciones.

---

## 2. El Problema que Resuelve

### El problema con Claude.ai / ChatGPT sueltos
- No tienen acceso a tus datos reales (Sheets, Calendar, etc.)
- Olvidan todo cuando cerrás el chat
- No pueden actuar, solo responden
- No son proactivos, solo reactivos

### El problema con OpenClaw y similares
- Código de terceros corriendo con acceso a tus cuentas y archivos
- Problemas serios de seguridad (21,000 instancias expuestas públicamente)
- No controlás qué datos salen de tu máquina
- Dependés del roadmap de otra persona

### La solución
Construir tu propio stack donde:
- **Sabés exactamente qué código corre**
- **Sabés exactamente qué datos van a dónde**
- **Podés extenderlo como quieras** sin limitaciones de plugins
- **Es tuyo para siempre**, no depende de que otra empresa mantenga el proyecto

---

## 3. Stack Tecnológico

### Visión general

```
┌─────────────────────────────────────────────────────────────────┐
│                    ORACLE CLOUD VM (gratis)                      │
│                                                                   │
│  ┌──────────┐    ┌─────────────┐    ┌──────────────────────┐    │
│  │  Nginx   │    │  PostgreSQL │    │       Redis          │    │
│  │  :80/443 │    │  (datos n8n)│    │  (colas de jobs)     │    │
│  └────┬─────┘    └─────────────┘    └──────────────────────┘    │
│       │                                                           │
│  ┌────▼─────┐    ┌─────────────────────────────────────────┐    │
│  │   n8n    │    │        Microservicio Go                  │    │
│  │  :5678   │    │             :8080                        │    │
│  │(orquest.)│    │  /finance  /projects  /memory  /cron     │    │
│  └──────────┘    └─────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
          │                        │
          ▼                        ▼
    SERVICIOS EXTERNOS       SERVICIOS EXTERNOS
    ├── WhatsApp API         ├── Claude API
    ├── Google Calendar      ├── Google Sheets API
    └── Gmail                └── SQLite (local, memoria vectorial)
```

### Componente por componente

| Componente | Tecnología | Propósito | Costo |
|---|---|---|---|
| Servidor | Oracle Cloud Free Tier | VM con 4 vCPU / 24GB RAM | $0 |
| Reverse proxy | Nginx | HTTPS, routing de tráfico | $0 |
| Certificado SSL | Let's Encrypt | HTTPS válido | $0 |
| Dominio | DuckDNS | Subdominio gratis | $0 |
| Base de datos | PostgreSQL 15 | Datos de n8n | $0 |
| Cola de jobs | Redis 7 | Múltiples workflows simultáneos | $0 |
| Orquestador | n8n (self-hosted) | Conecta WA con el microservicio | $0 |
| Microservicio | Go 1.22 | Lógica de negocio propia | $0 |
| Memoria semántica | SQLite + sqlite-vec | Búsqueda vectorial local | $0 |
| IA | Claude API (Anthropic) | El cerebro del asistente | ~$5-10/mes |
| Mensajería | WhatsApp Business Cloud API | Canal de comunicación | $0 (vol. bajo) |
| Notas | Google Sheets / Obsidian / Notion | Almacenamiento de datos | $0 |

### Por qué cada tecnología

**Oracle Cloud Free Tier** — 4 vCPU ARM + 24GB RAM gratis para siempre (no es trial). Es el doble de RAM de lo que necesitaría todo el stack junto. El único risk es que Oracle podría cambiar su política de free tier, pero al estar todo dockerizado, migrar a otro VPS sería copiar dos archivos.

**n8n** — Open source, self-hosted, con 1000+ integraciones. La diferencia con Zapier o Make es que no pagás por ejecución, corrés todo en tu propio servidor. Para flujos como WA → Claude → Sheets es ideal: tres nodos, sin código. No tiene sentido escribir Go para algo tan lineal.

**Microservicio Go** — Para lógica que n8n no puede manejar bien: memoria persistente entre sesiones, análisis de datos complejos, integración con herramientas que no tienen nodo en n8n, y cualquier lógica de negocio específica. Go es lo que ya sabés del trabajo y aplica perfectamente: interfaces limpias, bajo consumo de memoria, deploy trivial como binario Docker.

**SQLite + sqlite-vec** — En lugar de montar Postgres + pgvector o un servicio como Pinecone para memoria vectorial, `sqlite-vec` corre todo en un solo archivo. Para uso personal es más que suficiente y no agrega dependencias externas.

**Claude API** — El modelo que mejor sigue instrucciones complejas en español rioplatense. Precio similar a OpenAI para el mismo nivel de calidad. Y el punto clave: el plan Pro de $20/mes de Claude.ai **no cubre** la API, son sistemas de facturación separados.

---

## 4. Infraestructura — Oracle Cloud

### Qué es una VM en la nube

Cuando creás una VM en Oracle Cloud, estás alquilando esto:

```
Oracle Cloud Infrastructure (OCI)
│
└── Tu cuenta (Tenancy)
    └── Compartment (carpeta lógica de recursos)
        └── VCN — Virtual Cloud Network (tu red privada)
            ├── Subnet (segmento de red)
            └── VM Instance
                ├── CPU: 4 vCPU ARM Ampere
                ├── RAM: 24 GB
                ├── Storage: 200 GB
                └── IP Pública: xxx.xxx.xxx.xxx
```

La VM corre Ubuntu 22.04. Vos te conectás por SSH desde tu PC y tenés acceso root completo.

### SSH — cómo te conectás

SSH es el protocolo para controlar servidores Linux desde tu terminal. Funciona con un par de claves criptográficas:

```
Tu PC                              Oracle VM
┌─────────────────┐               ┌─────────────────┐
│  Clave Privada  │ ──── SSH ───► │  Clave Pública  │
│  (solo tuya,    │  encriptado   │  (Oracle la      │
│   nunca sale    │               │   guarda al      │
│   de tu PC)     │               │   crear la VM)   │
└─────────────────┘               └─────────────────┘
```

Oracle te genera el par de claves. Vos bajás la privada y la usás así:

```bash
ssh -i ~/.ssh/oracle_key ubuntu@TU_IP_PUBLICA
# Y ya estás dentro de la VM
ubuntu@vm:~$
```

### Docker — cómo vive el stack en la VM

Sin Docker, instalarías cada app manualmente con sus dependencias: un caos de versiones, conflictos, configuraciones. Con Docker, cada app vive en un contenedor aislado con todo lo que necesita:

```
VM Ubuntu 22.04
│
├── Contenedor: nginx          → el portero, maneja HTTPS
├── Contenedor: n8n            → el orquestador
├── Contenedor: PostgreSQL     → datos de n8n
├── Contenedor: Redis          → colas de ejecución
└── Contenedor: asistente-go   → tu microservicio
```

Docker Compose es el archivo YAML que describe todos estos contenedores y cómo se conectan. Un solo comando los levanta todos:

```bash
docker compose up -d   # -d = background, no bloquea la terminal
```

### Nginx — el portero del tráfico

Todo el tráfico que llega a la VM entra por Nginx primero. Él decide a quién mandárselo basado en la URL:

```
INTERNET :443 (HTTPS)
        │
      Nginx
        │
        ├── /webhook/* ──────────────► n8n :5678
        ├── /api/finance/* ──────────► microservicio :8080
        └── /n8n/* ──────────────────► n8n UI :5678
```

Nginx también termina SSL: el certificado HTTPS vive en Nginx, los servicios de adentro se hablan en HTTP simple porque están en la misma red interna de Docker.

### El docker-compose.yml completo

```yaml
services:
  postgres:
    image: postgres:15
    restart: always
    environment:
      POSTGRES_USER: n8n
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: n8n
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    restart: always

  n8n:
    image: n8nio/n8n
    restart: always
    ports:
      - "5678:5678"
    environment:
      - DB_TYPE=postgresdb
      - DB_POSTGRESDB_HOST=postgres
      - QUEUE_BULL_REDIS_HOST=redis
      - N8N_BASIC_AUTH_ACTIVE=true
      - N8N_BASIC_AUTH_USER=${N8N_USER}
      - N8N_BASIC_AUTH_PASSWORD=${N8N_PASSWORD}
      - WEBHOOK_URL=https://${DOMAIN}
    volumes:
      - n8n_data:/home/node/.n8n
    depends_on:
      - postgres
      - redis

  asistente:
    build: ./microservicio
    restart: always
    ports:
      - "8080:8080"
    env_file:
      - .env
    volumes:
      - ./credentials.json:/app/credentials.json:ro
      - sqlite_data:/app/data
    depends_on:
      - postgres

  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/letsencrypt:ro

volumes:
  postgres_data:
  n8n_data:
  sqlite_data:
```

---

## 5. Arquitectura Completa

### Diagrama de comunicación

```
VOS (WhatsApp)
      │
      │ mensaje de texto / audio / imagen
      ▼
  META SERVERS
      │
      │ POST webhook (mensaje entrante)
      ▼
    NGINX :443
      │
      │ forward a n8n
      ▼
     N8N
      │
      ├── ¿Es un gasto?
      │       │ POST /finance/expense
      │       ▼
      │   MICROSERVICIO GO
      │       │
      │       ├── Claude API (parsea el gasto)
      │       └── Google Sheets API (guarda la fila)
      │
      ├── ¿Es una nota de voz?
      │       │
      │       ├── Whisper API (transcribe audio)
      │       └── POST /memory/note → microservicio
      │
      ├── ¿Es pregunta libre?
      │       │
      │       └── Claude directamente desde n8n
      │
      └── respuesta → META API → VOS (WhatsApp)
```

### Las dos capas de inteligencia

**n8n** maneja el "qué hacer": recibe el mensaje, detecta la intención (es un gasto / es una nota / es una pregunta), y rutea al sistema correcto.

**Microservicio Go** maneja el "cómo hacerlo": tiene la lógica de negocio, la memoria persistente, y las integraciones complejas con APIs externas.

### Cuándo usa n8n vs el microservicio

| Tipo de tarea | Sistema | Por qué |
|---|---|---|
| Rutear mensaje de WA | n8n | Es un workflow lineal |
| Guardar gasto en Sheets | Microservicio Go | Necesita parseo con Claude + lógica de categorías |
| Crear evento en Calendar | n8n | Integración directa disponible en n8n |
| Daily briefing (cron) | n8n | Simple, trigger temporal |
| Respuesta con memoria | Microservicio Go | Necesita buscar en historial vectorial |
| Transcripción de audio | n8n (Whisper node) | Integración directa |
| Análisis de finanzas del mes | Microservicio Go | Lógica compleja sobre datos históricos |
| Notas de proyectos con búsqueda semántica | Microservicio Go | Requiere embeddings y búsqueda vectorial |

---

## 6. Cómo Funcionan las APIs

### Claude API — API Key simple

```
Tu request:
POST https://api.anthropic.com/v1/messages
Headers:
  x-api-key: sk-ant-...
  anthropic-version: 2023-06-01
Body:
  {
    "model": "${CLAUDE_MODEL}",
    "max_tokens": 1024,
    "system": "Sos un asistente que parsea gastos...",
    "messages": [{"role": "user", "content": "gasté 5000 en el super"}]
  }

Respuesta:
  {
    "content": [{"type": "text", "text": "{\"amount\": 5000, \"category\": \"Supermercado\"}"}]
  }
```

Facturación: por tokens consumidos. ~$3 por millón de tokens de input con Sonnet. Para uso personal moderado (bot de WA + daily briefing) son $2-5/mes.

**Importante:** El plan Pro de Claude.ai ($20/mes) no cubre la API. Son sistemas separados. La API se paga aparte en platform.anthropic.com.

### Google APIs — OAuth 2.0 y Service Account

Google tiene dos mecanismos de autenticación dependiendo del caso:

**Service Account** (para el microservicio Go → Sheets):
- Creás una cuenta de servicio en Google Cloud Console
- Descargás un archivo JSON con credenciales
- Compartís tu Spreadsheet con el email de la service account
- Tu código usa el JSON para autenticarse automáticamente, sin intervención humana

```go
// En Go, autenticación con service account
svc, err := sheets.NewService(ctx, 
    option.WithCredentialsFile("credentials.json"))
```

**OAuth 2.0** (para n8n → Gmail y Google Calendar):
- n8n te redirige a Google para que autorices
- Google le da un token a n8n
- n8n renueva el token automáticamente

La diferencia: Service Account es para servidores que actúan en nombre de sí mismos. OAuth es para actuar en nombre de un usuario.

### WhatsApp Business Cloud API — Meta Developer Portal

```
Meta Developer Portal
        │
        └── App de tipo Business
            └── Producto: WhatsApp
                ├── Phone Number ID   → identifica tu número
                ├── Access Token      → para enviar mensajes
                └── Webhook URL       → donde Meta manda mensajes entrantes
                    │
                    └── https://tudominio.duckdns.org/webhook/whatsapp
                                              ▲
                                    n8n escucha acá

Mensaje entrante:
  Meta hace POST a tu webhook con:
  {
    "entry": [{
      "changes": [{
        "value": {
          "messages": [{
            "from": "5491112345678",
            "text": {"body": "gasté 5000 en el super"}
          }]
        }
      }]
    }]
  }

Enviar respuesta:
  POST https://graph.facebook.com/v17.0/{PHONE_NUMBER_ID}/messages
  {
    "to": "5491112345678",
    "type": "text",
    "text": {"body": "🛒 Anotado! $5000 - Supermercado"}
  }
```

**Costo:** Gratis para volumen bajo (conversaciones iniciadas por el usuario). Empezás a pagar cuando superás cierto volumen mensual, lo cual para uso personal personal nunca pasa.

**El número:** No puede ser tu número personal. Necesitás un número dedicado al bot. Podés usar un número de prueba gratuito que Meta te da para testing, y después un número real (lo más barato es usar Twilio o un SIM físico).

---

## 7. n8n — El Orquestador

### Conceptos core

**Trigger** — el inicio de cada workflow:
- `Webhook` → algún sistema externo llama a una URL (WhatsApp mensajes entrantes)
- `Schedule` → cron expression ("todos los días a las 8am")
- `Manual` → vos lo iniciás a mano para testear

**Nodo** — cada paso del workflow. Algunos importantes:
- `HTTP Request` → llama a cualquier API (incluyendo tu microservicio)
- `IF` → lógica condicional
- `Code` → JavaScript custom para lógica que no tiene nodo
- `WhatsApp Business Cloud` → recibir y enviar mensajes
- `Google Sheets` → leer y escribir celdas
- `Google Calendar` → crear y leer eventos
- `Gmail` → leer y enviar mails

**Credenciales** — las API keys y tokens OAuth. Se guardan una vez, encriptadas en Postgres, y las reutilizás en todos los workflows.

### El workflow principal — WhatsApp Router

```
[WhatsApp Trigger]
        │
        ▼
[Code Node — detectar intención]
  if (msg.includes("gasté") || msg.includes("pagué") || msg.includes("lucas"))
    → "gasto"
  else if (msg.includes("recordame") || msg.includes("agendame"))
    → "recordatorio"
  else
    → "pregunta"
        │
        ├── "gasto" ──────────────► [HTTP Request: POST /finance/expense]
        │                                     │
        │                            [WhatsApp Send: response.body]
        │
        ├── "recordatorio" ────────► [Google Calendar: Create Event]
        │                                     │
        │                            [WhatsApp Send: confirmación]
        │
        └── "pregunta" ────────────► [Claude: chat completion]
                                              │
                                     [WhatsApp Send: respuesta]
```

### El workflow de Daily Briefing

```
[Schedule Trigger: 0 8 * * *]  ← todos los días 8am
        │
        ▼
[Google Calendar: listar eventos de hoy]
        │
        ▼
[Gmail: listar mails sin leer]
        │
        ▼
[HTTP GET /finance/summary?period=week]  ← microservicio
        │
        ▼
[Claude: armar resumen con todo el contexto]
  System: "Sos un asistente personal. Armá un briefing matutino..."
  User: "{eventos} {mails} {gastos_semana}"
        │
        ▼
[WhatsApp Send: briefing a tu número]
```

### Por qué n8n y no Make o Zapier

- **Make (Integromat):** Cobra por crédito/operación. 100k tareas pueden ser $500+/mes. n8n cobra por ejecución completa, no por operación individual.
- **Zapier:** Caro, cerrado, limitado. 20 workflows en el plan básico de $20/mes.
- **n8n cloud:** $20/mes. No tiene sentido cuando podés self-hostearlo en Oracle gratis.
- **n8n self-hosted:** El software es gratis y open source. Solo pagás el servidor, que en este caso es $0.

---

## 8. El Microservicio Go

### Por qué Go y no otro lenguaje

- Es lo que ya sabés del trabajo anterior en MeLi
- Mismo paradigma: microservicios, interfaces, DI, hexagonal architecture
- Deploy trivial: un binario estático en Docker, ~10MB de imagen
- Bajo consumo de memoria: el servicio completo usa ~50MB de RAM
- Tipado estático: los errores aparecen en compile time, no en producción

### Estructura del proyecto

```
asistente/
├── cmd/
│   └── server/
│       └── main.go              → entry point, configura router
├── config/
│   └── config.go                → carga variables de entorno
├── internal/
│   ├── finance/
│   │   ├── service.go           → lógica de gastos
│   │   └── handler.go           → HTTP handler
│   ├── memory/
│   │   ├── store.go             → SQLite + sqlite-vec
│   │   └── embeddings.go        → genera embeddings con Claude
│   ├── context/
│   │   ├── engine.go            → ingest / assemble / compact
│   │   └── session.go           → manejo de sesiones por usuario
│   ├── cron/
│   │   ├── scheduler.go         → cron jobs propios
│   │   └── jobs.go              → definición de jobs (briefing, etc.)
│   ├── projects/                → (próximo módulo)
│   ├── university/              → (próximo módulo)
│   └── work/                    → (próximo módulo)
├── pkg/
│   ├── claude/
│   │   └── client.go            → wrapper de Anthropic API
│   └── sheets/
│       └── client.go            → wrapper de Google Sheets API
├── docker-compose.yml
├── Dockerfile
└── .env.example
```

### Interfaz del Channel (patrón de OpenClaw)

En lugar de hardcodear WhatsApp en toda la lógica, se abstrae el canal:

```go
// pkg/channel/channel.go
type Channel interface {
    Send(to string, message string) error
    CanDo(action Action) bool
}

type Action string
const (
    ActionSend      Action = "send"
    ActionReactions Action = "reactions"
    ActionMedia     Action = "media"
)

// Implementación concreta de WhatsApp
type WhatsAppChannel struct {
    phoneNumberID string
    accessToken   string
    httpClient    *http.Client
}

func (w *WhatsAppChannel) Send(to, message string) error {
    // llama a graph.facebook.com/messages
}

func (w *WhatsAppChannel) CanDo(action Action) bool {
    switch action {
    case ActionSend, ActionMedia:
        return true
    case ActionReactions:
        return false // WA Business API no soporta reactions
    }
    return false
}
```

Cuando en el futuro se quiera agregar Telegram, es implementar la misma interfaz. La lógica de negocio no cambia.

### El Context Engine (idea de OpenClaw)

La parte más valiosa del análisis: cómo manejar la memoria de conversación sin gastar todos los tokens.

```go
// internal/context/engine.go
type ContextEngine interface {
    Ingest(sessionID string, msg Message) error
    Assemble(sessionID string) ([]Message, error)
    Compact(sessionID string) error
}

// Ingest: guarda el mensaje en memoria
func (e *engine) Ingest(sessionID string, msg Message) error {
    // 1. Guarda en historial de sesión (SQLite)
    // 2. Genera embedding del mensaje
    // 3. Indexa en sqlite-vec para búsqueda semántica futura
}

// Assemble: arma el contexto para mandar a Claude
func (e *engine) Assemble(sessionID string) ([]Message, error) {
    // 1. Trae los últimos N mensajes del historial
    // 2. Verifica que no supere el budget de tokens
    // 3. Si supera → llama a Compact primero
    // 4. Devuelve los mensajes listos para enviar a Claude
}

// Compact: el truco más valioso del análisis de OpenClaw
// Cuando la conversación es muy larga, Claude la resume
// y se descartan los mensajes viejos pero se guarda el resumen
func (e *engine) Compact(sessionID string) error {
    messages := e.loadSession(sessionID)
    
    summary, err := e.claude.Complete(
        "Resumí esta conversación en un párrafo conciso. Preservá hechos importantes.",
        formatMessages(messages),
    )
    
    // Reemplaza el historial largo por el resumen
    e.saveSession(sessionID, []Message{
        {Role: "system", Content: "Resumen de conversación anterior: " + summary},
    })
}
```

### La Memoria Vectorial (sqlite-vec)

En lugar de un servicio externo de vectores (Pinecone, Weaviate), se usa SQLite con la extensión `sqlite-vec`. Todo queda en un archivo local.

```go
// internal/memory/store.go
type MemoryStore struct {
    db *sql.DB
}

// SaveMemory guarda una nota con su embedding
func (m *MemoryStore) SaveMemory(content string, embedding []float32) error {
    // INSERT en tabla memories con el vector
}

// Search busca memorias semánticamente similares
func (m *MemoryStore) Search(query string, limit int) ([]Memory, error) {
    // 1. Genera embedding de la query
    // 2. Busca en sqlite-vec por similitud coseno
    // 3. Devuelve las memorias más relevantes con score
}
```

Esto permite preguntas como:
- "¿Qué decidí sobre el sistema de cartas del juego?" → busca semánticamente en todas tus notas
- "¿Cuánto gasté el mes pasado en transporte?" → busca en historial de finanzas
- "¿Qué me dijo el profe sobre el parcial?" → busca en notas de UADE

### El Módulo Finance — implementado

```go
// internal/finance/service.go
func (s *Service) ProcessMessage(msg, sender string) (*ParseResult, error) {
    // 1. Manda el texto a Claude con un prompt que entiende lunfardo:
    //    "5 lucas de nafta" → {amount: 5000, category: "Transporte"}
    //    "Maca pagó la farmacia, 3000" → {amount: 3000, paid_by: "Maca", category: "Salud"}
    //    "$20 USD de Netflix" → {amount_usd: 20, category: "Entretenimiento"}

    // 2. Escribe la fila en Google Sheets:
    //    | Fecha | Descripción | Categoría | Monto ARS | Monto USD | Pagó |

    // 3. Devuelve respuesta amigable con emoji por categoría
    return &ParseResult{
        Response: "🛒 Anotado! $5,000 - Supermercado\nSebas pagó el 09/03/2026",
    }, nil
}
```

### El Cron Scheduler (patrón de OpenClaw)

OpenClaw tiene un sistema de cron jobs que son la parte "proactiva" del asistente. La idea adaptada a Go:

```go
// internal/cron/jobs.go
type Job struct {
    ID       string
    Schedule string   // cron expression
    Prompt   string   // qué pedirle a Claude
    Delivery Delivery // cómo entregar el resultado
}

type Delivery struct {
    Mode    string // "whatsapp" | "webhook" | "none"
    To      string // número de WA
}

// Jobs definidos
var DefaultJobs = []Job{
    {
        ID:       "daily-briefing",
        Schedule: "0 8 * * *", // todos los días a las 8am
        Prompt:   "Armá un briefing matutino con eventos del día, mails importantes, y resumen de gastos de la semana",
        Delivery: Delivery{Mode: "whatsapp", To: "+5491112345678"},
    },
    {
        ID:       "weekly-finance",
        Schedule: "0 20 * * 0", // domingos a las 8pm
        Prompt:   "Hacé un resumen de gastos de la semana y compará con el presupuesto mensual",
        Delivery: Delivery{Mode: "whatsapp", To: "+5491112345678"},
    },
}
```

---

## 9. Lecciones del Análisis de OpenClaw

OpenClaw (antes Clawdbot, Moltbot) es el agente personal de IA open source más popular de 2026. Pasó de 1,000 a 196,000 stars en GitHub en semanas. Fue creado por Peter Steinberger (fundador de PSPDFKit).

### Qué hace que sea poderoso

El código (TypeScript, Node.js) revela una arquitectura muy pensada:

**1. Skills como Markdown, no como código**
Las Skills no son plugins programados. Son archivos `.md` con frontmatter YAML que describen las capacidades al modelo. Claude los lee como contexto y "sabe" qué herramientas tiene disponibles. Esto hace que sea trivial agregar nuevas capacidades sin recompilar.

```markdown
---
name: notion
metadata:
  openclaw:
    emoji: 📝
    requires:
      env: [NOTION_API_KEY]
---
# notion
Use the Notion API to create/read/update pages...
```

**2. Memoria vectorial híbrida**
Usan búsqueda semántica (embeddings) + full-text search (FTS) + temporal decay (recuerdos viejos pesan menos). Y lo más importante: usan SQLite con sqlite-vec, sin necesidad de bases de datos separadas.

**3. Heartbeat / CronService**
La diferencia entre un chatbot y un asistente real es que el asistente actúa sin que vos lo pidas. El CronService de OpenClaw corre jobs en background que preguntan al modelo "¿qué debería hacer ahora basado en mi contexto actual?" y entrega los resultados al canal configurado.

**4. Context Engine con Compact**
Cuando la conversación supera el contexto máximo del modelo, en lugar de truncar (que pierde información) o fallar, el engine comprime todo el historial en un resumen y continúa. Transparente para el usuario.

**5. Channel abstraction**
Todos los canales de mensajería (WhatsApp, Telegram, Discord, iMessage, etc.) implementan la misma interfaz. La lógica de negocio no sabe en qué canal está hablando.

### Los problemas reales de OpenClaw

1. **Seguridad:** El Gateway corre en puerto 18789 y muchos usuarios lo expusieron públicamente. Pillar Security registró actividad de explotación activa en menos de una hora de exponer una instancia de prueba. Hubo casos de harvesting de API keys.

2. **Dependencia:** Si OpenClaw deja de mantenerse (el creador ya se fue a OpenAI), el proyecto podría estancarse.

3. **Complejidad:** El repo tiene 10,729 commits, 3,100 issues abiertos, y es TypeScript. Para entender qué pasa en un bug, hay que meterse en un codebase enorme.

4. **Pérdida de control:** No sabés exactamente qué datos van a dónde, qué skills de la comunidad hacen realmente, o si hay prompt injection en alguna skill instalada.

### Por qué construir el propio en Go en lugar de usar OpenClaw

| Criterio | OpenClaw | Stack propio en Go |
|---|---|---|
| Setup | 30 min | 4-8 horas |
| Control total | No | Sí |
| Entender el código | Difícil (10k commits) | Sí (es tuyo) |
| Seguridad | Riesgosa si se expone | Controlada |
| Flexibilidad | Skills del marketplace | Código propio |
| Dependencia | Del proyecto externo | Independiente |
| Curva de aprendizaje | Baja | Alta (pero aprendés infra) |

Para alguien no técnico: OpenClaw. Para un dev que quiere control, aprender el proceso, y construir algo que escala como él quiera: stack propio.

---

## 10. Ideas Robadas de OpenClaw para el Microservicio

Estas son las ideas concretas extraídas del análisis del repo que se van a implementar en Go:

### Idea 1: Skill como configuración Markdown

En lugar de hardcodear las capacidades del asistente en el código, definirlas en archivos `.md`:

```
asistente/
└── skills/
    ├── finance.md      → "Podés registrar gastos. Formatos: '5000 en nafta', '5 lucas de super'..."
    ├── projects.md     → "Podés consultar el estado de Mythological Oath, Vaultbreakers..."
    ├── university.md   → "Podés agregar fechas de entrega de UADE, consultar el estado de materias..."
    └── work.md         → "Podés guardar notas de EDUCABOT, registrar tareas..."
```

Claude recibe estos archivos como parte del system prompt. Si se quiere agregar una nueva capacidad, se agrega un `.md`, sin tocar Go.

### Idea 2: Compact del Context Engine

El patrón más valioso. Cuando el historial de conversación supera ~50,000 tokens, en lugar de truncar:

```
Historial largo (100 mensajes, 80k tokens)
        │
        ▼ Claude compact
"Resumen: el usuario registró gastos de $45,000 ARS en febrero.
 Tiene una entrega de UADE el 15 de marzo sobre Unity optimization.
 Decidió usar hexagonal architecture para el microservicio..."
        │
        ▼
Historial nuevo (1 mensaje de resumen, ~500 tokens)
+ continúa la conversación con contexto preservado
```

### Idea 3: Temporal Decay en memoria

No todas las memorias valen igual. Los recuerdos recientes pesan más que los viejos al hacer búsqueda semántica:

```go
// Score final = similarity_score * time_decay_factor
func timeDecay(createdAt time.Time) float64 {
    daysSince := time.Since(createdAt).Hours() / 24
    return math.Exp(-0.05 * daysSince) // decae exponencialmente
}
```

### Idea 4: Action Gate por canal

```go
// Cada canal declara qué puede hacer
type WhatsAppChannel struct{}
func (w WhatsAppChannel) CanDo(action Action) bool {
    return action == ActionSend || action == ActionMedia
    // no puede hacer reactions, no puede editar mensajes enviados
}
```

### Idea 5: Isolated Agent para cron jobs

En OpenClaw, los cron jobs corren en un "isolated agent" separado del contexto de conversación normal. Esto es importante: el job del daily briefing no debe contaminar el contexto de las conversaciones activas del día.

```go
// El briefing corre con su propio contexto
func (s *Scheduler) runJob(job Job) {
    ctx := s.buildJobContext(job) // contexto aislado, sin historial de conversaciones
    response, _ := s.claude.Complete(ctx, job.Prompt)
    s.deliver(job.Delivery, response)
}
```

---

## 11. Módulos del Microservicio

### Módulo 1: Finance (implementado)

**Endpoints:**
- `POST /finance/expense` → registra un gasto desde texto libre
- `GET /finance/summary` → resumen por periodo (implementado)

**Flujo:**
```
"gasté 5000 en el super"
        │
        ▼ Claude parsea
{amount: 5000, category: "Supermercado", paid_by: "Sebas", date: "2026-03-09"}
        │
        ▼ Google Sheets API
Fila nueva en la hoja "Gastos"
| 2026-03-09 | Supermercado | Supermercado | 5000 | 0 | Sebas |
        │
        ▼ respuesta
"🛒 Anotado! $5,000 - Supermercado\nSebas pagó el 09/03/2026"
```

**Categorías soportadas:** Supermercado, Restaurante, Transporte, Servicios, Salud, Ropa, Entretenimiento, Educación, Hogar, Otro.

**Soporte ARS/USD:** Si se menciona dólares, se guarda en la columna USD. Las formulas de tu Sheets ya convierten automáticamente.

### Módulo 2: Memory (implementado)

**Endpoints:**
- `POST /memory/note` → guarda una nota con embedding
- `GET /memory/search?q=...` → búsqueda semántica en todas las notas
- `DELETE /memory/note/:id` → elimina una nota

**Casos de uso:**
```
"guardá que el sistema de cartas de Mythological Oath usa un pool de 40 cartas"
        │ POST /memory/note
        ▼
{
  content: "El sistema de cartas de Mythological Oath usa un pool de 40 cartas",
  embedding: [0.234, -0.891, ...],  // vector generado por Claude
  timestamp: "2026-03-09",
  tags: ["mythological-oath", "game-design"]
}

// Más tarde:
"¿cuántas cartas tiene el pool del juego?"
        │ GET /memory/search?q=cartas+pool+juego
        ▼
"El sistema de cartas de Mythological Oath usa un pool de 40 cartas (nota del 09/03)"
```

### Módulo 3: Projects (implementado)

**Endpoints:**
- `POST /projects/note` → agrega nota a un proyecto
- `GET /projects/:name/status` → estado actual del proyecto
- `GET /projects/:name/context` → contexto completo para Claude

**Proyectos iniciales:**
- `mythological-oath` — el roguelike card game
- `vaultbreakers` — proyecto de UADE

**Flujo:**
```
"¿cuánto me falta para el MVP de Mythological Oath?"
        │
        ▼ GET /projects/mythological-oath/context
Contexto: GDD, features implementadas, features pendientes, última nota
        │
        ▼ Claude con el contexto
"Te falta: el sistema de boss, las cartas pasivas de tier 3, 
 y la UI del menú principal. Basado en tus notas, llevás el 60% del MVP."
```

### Módulo 4: University / Work (cubiertos por Memory + Projects)

Los modulos University y Work del diseno original se resuelven con el sistema generico de Memory (notas con tags y busqueda semantica) y Projects (status con AI). No se necesitan endpoints especificos — guardar una nota con tag "uade" o "educabot" logra lo mismo con mas flexibilidad.

### Modulos adicionales implementados (no en diseno original)

- **Habits** — `POST /api/habits/log`, `GET /api/habits/streak`, `GET /api/habits/today`
- **Links** — `POST /api/links`, `GET /api/links/search` (guardado con embeddings)
- **GitHub** — repos, issues, PRs
- **Jira** — my issues, crear issue, transiciones
- **Spotify** — currently playing, play/pause/next
- **Todoist** — tasks CRUD + complete
- **Gmail** — unread list + read message
- **ClickUp** — tasks CRUD por equipo
- **Notion** — crear y leer paginas
- **Obsidian** — CRUD de notas en vault local

---

## 12. Flujo Completo de un Mensaje

### Ejemplo real: "gasté 5 lucas de nafta"

```
1. Vos escribís "gasté 5 lucas de nafta" en WhatsApp al número del bot

2. Meta (servidor de WhatsApp) recibe el mensaje y hace:
   POST https://tudominio.duckdns.org/webhook/whatsapp
   {
     "entry": [{"changes": [{"value": {
       "messages": [{"from": "5491112345678", "text": {"body": "gasté 5 lucas de nafta"}}]
     }}]}]
   }

3. Nginx recibe en :443, termina SSL, y pasa a n8n en :5678

4. n8n procesa el workflow "WhatsApp Router":
   - Extrae: from="5491112345678", text="gasté 5 lucas de nafta"
   - Code node detecta intención → "gasto"
   - Hace POST a microservicio:
     POST http://asistente:8080/finance/expense
     Headers: X-Webhook-Secret: tu_secret_random
     Body: {"message": "gasté 5 lucas de nafta", "sender": "Sebas"}

5. El microservicio Go recibe el request:
   - Verifica el webhook secret
   - Llama a Claude API:
     System: "Sos un asistente que extrae gastos en español argentino..."
     User: "gasté 5 lucas de nafta"
   
6. Claude API responde:
   {"amount": 5000, "category": "Transporte", "description": "Nafta", 
    "paid_by": "Sebas", "date": "2026-03-09", "amount_usd": 0}

7. El microservicio llama a Google Sheets API:
   - Agrega fila: ["2026-03-09", "Nafta", "Transporte", 5000, 0, "Sebas"]
   - Sheets responde OK

8. El microservicio responde a n8n:
   {"success": true, "response": "🚗 Anotado!\nTransporte — Nafta\nSebas pagó $5,000 el 09/03/2026"}

9. n8n toma el response y llama a Meta API:
   POST https://graph.facebook.com/v17.0/{PHONE_NUMBER_ID}/messages
   {"to": "5491112345678", "type": "text", 
    "text": {"body": "🚗 Anotado!\nTransporte — Nafta\nSebas pagó $5,000 el 09/03/2026"}}

10. Meta entrega el mensaje a tu WhatsApp

⏱️ Tiempo total: ~2-3 segundos
🪙 Costo en tokens: ~300 tokens = $0.0009 (menos de un centavo)
```

---

## 13. Costo Mensual

### Breakdown completo

| Servicio | Costo | Notas |
|---|---|---|
| Oracle Cloud VM | $0 | Free tier para siempre. 4vCPU + 24GB RAM |
| DuckDNS (dominio) | $0 | Subdominio gratis |
| Let's Encrypt (SSL) | $0 | Certificado HTTPS gratis |
| n8n (software) | $0 | Open source, self-hosted |
| PostgreSQL | $0 | Corre en Docker en la VM |
| Redis | $0 | Corre en Docker en la VM |
| WhatsApp Business API | $0* | Probablemente sin costo en uso personal, sujeto al pricing vigente de Meta |
| Google APIs | $0 | Dentro del free tier para uso personal |
| Claude API (Anthropic) | ~$5-10/mes | El único costo real |
| **TOTAL** | **~$5-10/mes** | |

### Estimación de tokens con Claude API

Asumiendo uso moderado:
- 20 gastos por mes → ~6,000 tokens → $0.02
- Daily briefing (30 días) → ~150,000 tokens → $0.45
- Preguntas libres (50/mes) → ~50,000 tokens → $0.15
- Notas y memoria → ~20,000 tokens → $0.06

**Total estimado: ~$0.70/mes en tokens**

En la práctica, con uso real y más activo, es razonable presupuestar $3-8/mes en Claude API para tener margen.

---

## 14. Roadmap de Setup

### Día 1 — Infraestructura base

**Oracle Cloud:**
1. Crear cuenta en cloud.oracle.com/free
2. Crear VM: Ubuntu 22.04, shape VM.Standard.A1.Flex (4 vCPU / 24GB RAM)
3. Descargar la clave SSH privada
4. Abrir puertos en Security List: 22 (SSH), 80 (HTTP), 443 (HTTPS)
5. Conectarse: `ssh -i oracle_key ubuntu@TU_IP`

**Docker:**
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install docker.io docker-compose-plugin -y
sudo systemctl enable docker
sudo usermod -aG docker ubuntu
```

**Dominio:**
1. Crear cuenta en duckdns.org
2. Registrar `sebas-asistente.duckdns.org`
3. Apuntarlo a la IP de Oracle

**SSL:**
```bash
sudo apt install certbot -y
sudo certbot certonly --standalone -d sebas-asistente.duckdns.org
```

### Día 2 — n8n + Primera integración

1. Subir el `docker-compose.yml` a la VM
2. `docker compose up -d`
3. Acceder a `https://sebas-asistente.duckdns.org:5678`
4. Configurar credenciales de Google (OAuth)
5. Primer workflow: Schedule → Gmail fetch → Log (solo para verificar)

### Día 3 — WhatsApp

1. Crear app en developers.facebook.com
2. Agregar producto WhatsApp
3. Configurar webhook: `https://sebas-asistente.duckdns.org/webhook/whatsapp`
4. En n8n: agregar WhatsApp Trigger
5. Primer workflow funcional: WA recibe → WA responde "hola!"

### Día 4 — Microservicio Go

1. Compilar el microservicio: `go build -o asistente ./cmd/server`
2. Build de la imagen Docker: `docker build -t asistente .`
3. Agregar el servicio al docker-compose.yml
4. `docker compose up -d asistente`
5. Test: `curl -X POST localhost:8080/finance/expense -d '{"message":"gasté 5000"}'`

### Día 5 — Google Sheets

1. Google Cloud Console → crear proyecto
2. Habilitar Google Sheets API
3. Crear Service Account → descargar `credentials.json`
4. Compartir el Spreadsheet con el email de la service account
5. Test end-to-end: mensaje WA → microservicio → fila nueva en Sheets

### Semana 2 en adelante — Modulos adicionales (completado)

Todo implementado: Memory con busqueda hibrida, Context Engine con Compact, Projects generico, Habits, Links, 13 integraciones externas, 4 cron jobs, 191 tests. Ver seccion 17 para detalle completo.

---

## 15. Decisiones de Diseño y Por Qué

### ¿Por qué no usar OpenClaw directamente?

OpenClaw es el stack más maduro para esto, con 196,000 stars y una comunidad activa. Sin embargo:

1. Es TypeScript / Node.js, no Go. Agregar lógica custom requiere aprender el framework.
2. El modelo de seguridad es preocupante: instancias expuestas, prompt injection en skills, harvesting de API keys documentado por investigadores de seguridad.
3. Dependés del roadmap externo. El creador ya se fue a OpenAI.
4. Para casos de uso personales específicos (finanzas en ARS/USD, proyectos del juego, UADE), el overhead de adaptar OpenClaw supera el de construir los módulos necesarios.

La decisión es estudiar la arquitectura de OpenClaw (que es excelente) y aplicar los patrones en Go con control total.

### ¿Por qué n8n + microservicio y no todo en Go?

Para flujos lineales simples (WA → Claude → Google Calendar), n8n es 10x más rápido de implementar y mantener que código Go. No tiene sentido escribir un cliente de Google Calendar en Go cuando n8n ya lo tiene resuelto con autenticación OAuth y todo.

Para lógica con estado, memoria, análisis de datos, o integración con herramientas sin nodo en n8n: Go.

La división es: **n8n para el routing y las integraciones simples, Go para la lógica de negocio compleja**.

### ¿Por qué Oracle Cloud y no Railway o DigitalOcean?

Railway y Render son más fáciles de setup pero cuestan $5-10/mes. Oracle es gratis para siempre y tiene más RAM (24GB vs 1GB de Railway).

El trade-off: Oracle requiere más configuración inicial (Linux, Docker, Nginx, SSL). Pero una vez configurado es fire-and-forget. Y con 24GB de RAM, hay espacio para crecer: agregar Ollama para modelos locales, montar más servicios, etc.

### ¿Por qué sqlite-vec y no Postgres con pgvector?

Para uso personal con volúmenes bajos (miles de notas, no millones), sqlite-vec es más que suficiente. No necesita un servicio separado, no necesita configuración de índices, y el backup es copiar un archivo. pgvector agregaría complejidad sin beneficio real a esta escala.

### ¿Por qué WhatsApp y no Telegram?

WhatsApp es donde ya estás hablando todos los días. La fricción de aprender a hablar con un bot en una app nueva es alta. Con WhatsApp, es el mismo lugar donde hablás con Maca, con familia, con el trabajo — el bot vive ahí naturalmente.

Técnicamente Telegram tiene una API más amigable para bots. Pero la usabilidad gana a la conveniencia técnica.

---

## 16. Comparación con Alternativas

### OpenClaw vs Stack propio

| Criterio | OpenClaw | Stack propio |
|---|---|---|
| Tiempo de setup | 30 min | 4-8 horas |
| Conocimiento requerido | Bajo | Dev mid-senior |
| Control sobre el código | Ninguno | Total |
| Seguridad | Riesgosa si mal configurada | Controlada |
| Extensibilidad | Skills del marketplace | Código propio en Go |
| Dependencia externa | Alta | Solo Claude API |
| Costo | $0 (+ API de IA) | $0 (+ API de IA) |
| Escalabilidad | Limitada al framework | Ilimitada |

### n8n vs Make vs Zapier (para el orquestador)

| Criterio | n8n self-hosted | Make | Zapier |
|---|---|---|---|
| Costo | $0 | $9-16/mes | $20+/mes |
| Control de datos | Total | En cloud de Make | En cloud de Zapier |
| Integraciones | 1000+ | 2800+ | 8000+ |
| Complejidad setup | Media | Baja | Muy baja |
| Código custom | JavaScript en nodos | Limitado | Muy limitado |
| Self-hosted | Sí | No | No |

### Railway vs Oracle vs VPS propio (para el servidor)

| Criterio | Oracle Cloud | Railway | Hetzner VPS |
|---|---|---|---|
| Costo | $0 para siempre | $5/mes | €4.50/mes |
| RAM | 24 GB | 1 GB | 4 GB |Bien 
| Setup | 60 min | 5 min | 30 min |
| Control | Total | Limitado | Total |
| Riesgo | Política de free tier | Estabilidad de empresa | Bajo |
| Ideal para | Este proyecto | Prototipos rápidos | Proyectos productivos |

---

## Resumen Ejecutivo

**Asistente personal propio** accesible por WhatsApp, corriendo en una VM de Oracle Cloud gratis, con n8n como orquestador de workflows y un microservicio Go como motor de logica de negocio.

**Estado: microservicio completo y testeado.** 38 endpoints, 13 integraciones externas, 191 tests, validaciones de datos en domain, SQL externalizado, arquitectura limpia con capas separadas.

El asistente puede:
- Registrar gastos con lenguaje natural y guardarlos en Google Sheets
- Chat con memoria persistente y compactacion automatica
- Busqueda semantica hibrida (vector + FTS) sobre notas
- Briefings diarios automaticos, alertas de presupuesto, journal via WhatsApp
- Tracking de habitos con streaks
- Guardado y busqueda de links
- Status de proyectos con AI
- Integraciones con Calendar, Notion, Obsidian, GitHub, Jira, Spotify, Todoist, Gmail, ClickUp

Agregar un nuevo modulo es: domain struct + controller + usecase + ruta en `routes.go`. Agregar un nuevo client es un archivo en `clients/`.

**Costo total: ~$5-10/mes** (solo Claude/OpenAI API). Todo lo demas corre en infraestructura gratuita.

**Pendiente para produccion:** deploy a Oracle Cloud, configurar n8n + WhatsApp webhook de entrada, rate limiting.

---

## 17. Estado Actual de Implementacion (Marzo 2026)

### Implementado y funcionando

| Modulo | Status | Notas |
|---|---|---|
| **Finance** | Completo | Parse de gastos + Sheets + summary por periodo |
| **Memory** | Completo | Notas con embeddings, busqueda hibrida (vector + FTS5 + fallback) |
| **Chat** | Completo | Conversaciones persistentes con compact automatico |
| **Cron Jobs** | Completo | Daily briefing, weekly finance, budget alert, daily journal |
| **Habits** | Completo | Log + streak tracking + today |
| **Links** | Completo | Guardado con embeddings + busqueda semantica |
| **Projects** | Completo | Status summary con AI sobre notas del proyecto |
| **Calendar** | Completo | Eventos de hoy + crear evento |
| **Notion** | Completo | Crear y leer paginas |
| **Obsidian** | Completo | CRUD de notas en vault local |
| **WhatsApp** | Completo | Envio de mensajes via Business Cloud API |
| **GitHub** | Completo | Repos, issues, PRs |
| **Jira** | Completo | My issues, crear issue, transiciones |
| **Spotify** | Completo | Playing, play/pause/next |
| **Todoist** | Completo | Tasks CRUD + complete |
| **Gmail** | Completo | Listar unread + leer mensaje |
| **ClickUp** | Completo | Tasks CRUD por equipo |

### Decisiones que cambiaron vs el diseno original

| Diseno original | Implementacion real | Por que |
|---|---|---|
| `internal/` para toda la logica | `pkg/` con capas (domain/controller/usecase/service) | Mejor separacion y testabilidad |
| sqlite-vec para vectores | Embeddings propios via Claude/OpenAI + cosine similarity | Mas simple, sin extension nativa |
| Modulos University y Work separados | Cubiertos por Memory + Projects generico | Un sistema generico es mas flexible que modulos hardcodeados |
| n8n como unico orquestador de WhatsApp | Microservicio con WhatsApp client propio + cron interno | Menos dependencia de n8n, cron jobs nativos en Go |
| Solo Claude API | Claude + OpenAI (configurable via AI_PROVIDER) | Flexibilidad de proveedor |
| Clients en `pkg/` (un paquete por client) | `clients/` top-level (un archivo por client) | Menos paquetes, tipos prefijados, mas limpio |
| SQL inline en codigo Go | `.sql` files en `pkg/service/sqldata/` via go:embed | Queries legibles, separadas del codigo |

### Arquitectura actual

```
cmd/main.go         → 3 lineas (NewApp → Run → Close)
cmd/server.go       → App struct con NewApp(), Run(), Close()
cmd/clients.go      → Clients struct (13 clientes)
cmd/controller.go   → Controllers struct (15 controllers)
cmd/scheduler.go    → 4 cron jobs
cmd/routes.go       → 38 endpoints
clients/            → 13 clients + common.go
pkg/domain/         → Models + errors + Validate() methods
pkg/controller/     → 15 HTTP handlers
pkg/usecase/        → 7 usecases + scheduler
pkg/service/        → SQLite + Postgres + Sheets + embeddings + cache
pkg/service/sqldata/ → 24 archivos .sql (12 sqlite + 12 postgres)
```

### Testing

- **191 tests** across 22 test files
- Domain validation tests (path traversal, dates, URLs, max length)
- Controller tests con `errorFromBody()` helper
- UseCase tests con mocks (testify/mock)
- Service tests (SQLite integration + embeddings unit)
- Convenciones: AAA sin comentarios, `errors.Is()`, expected errors como constantes, sin `Contains`

### Pendiente / Ideas futuras

- [ ] Deploy a Oracle Cloud (infra lista en docker-compose, falta el push)
- [ ] Configurar n8n workflows de WhatsApp en produccion
- [ ] OAuth refresh para Spotify (token expira)
- [ ] Webhook de entrada para WhatsApp (actualmente solo envio)
- [ ] Rate limiting en endpoints publicos
- [ ] Metrics/observability (Prometheus o similar)
- [ ] Backup automatico de SQLite
- [ ] Multi-user (actualmente single-user con DefaultSender)

---

*Documento original generado el 9 de marzo de 2026.*
*Stack diseñado en base a analisis de OpenClaw repo (MIT License), investigacion de herramientas 2025-2026, y sesiones de diseno.*
*Seccion 17 actualizada el 11 de marzo de 2026 con estado real de implementacion.*
