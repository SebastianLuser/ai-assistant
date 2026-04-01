# Deploy Guide — Hetzner + Coolify

## Prerequisitos

- Cuenta en [Hetzner Cloud](https://console.hetzner.cloud)
- Repo en GitHub
- Dominio o subdomain (o DuckDNS gratis)
- API keys: Claude/OpenAI, WhatsApp Business, Google credentials, etc.

---

## 1. Crear VPS en Hetzner (~2 min)

1. Ir a [console.hetzner.cloud](https://console.hetzner.cloud)
2. **Create Server**:
   - **Location**: Nuremberg (eu-central) o el más cercano
   - **Image**: Ubuntu 24.04
   - **Type**: CX22 (2 vCPU, 4GB RAM, 40GB SSD) — **~€4.5/mes**
   - **SSH Key**: agregar tu clave pública
   - **Name**: `asistente`
3. Click **Create & Buy**
4. Anotar la IP pública

> Tu stack (1 binario Go + SQLite) usa ~200MB RAM. CX22 sobra.

---

## 2. Instalar Coolify (~1 min)

```bash
ssh root@TU_IP
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
```

Esperá 2-3 minutos. Accedé a `http://TU_IP:8000` y creá tu cuenta admin.

---

## 3. Conectar dominio

### Opción A: Dominio propio
```
asistente.tudominio.com  →  A  →  TU_IP
```


## 4. Configurar proyecto en Coolify

1. **Projects** → **New** → **New Resource** → **Public Repository**
2. URL: `https://github.com/TU_USUARIO/jarvis`
3. Branch: `main`
4. **Build Pack**: Docker Compose

### Variables de entorno

En Coolify → **Environment Variables**:

```env
# Requeridas
CLAUDE_API_KEY=sk-ant-...
WHATSAPP_PHONE_NUMBER_ID=tu-phone-id
WHATSAPP_ACCESS_TOKEN=tu-token
WHATSAPP_TO_NUMBER=5491112345678
WHATSAPP_VERIFY_TOKEN=un-token-random
WHATSAPP_APP_SECRET=tu-app-secret-de-meta

# Opcionales
OPENAI_API_KEY=sk-...
GOOGLE_SHEETS_ID=tu-spreadsheet-id
GOOGLE_CALENDAR_ID=tu-calendar@group.calendar.google.com
GMAIL_USER_EMAIL=tu@gmail.com
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
TELEGRAM_SECRET_TOKEN=un-secret-random
FIGMA_ACCESS_TOKEN=tu-figma-token
NOTION_API_KEY=secret_...
GITHUB_TOKEN=ghp_...
TODOIST_API_TOKEN=...
```

### Subir credentials.json (Google)

```bash
scp credentials.json root@TU_IP:/opt/coolify/credentials.json
```

---

## 5. SSL (automático)

1. En Coolify → configuración del recurso → **Domains**
2. Agregar: `asistente.tudominio.com`
3. Coolify genera certificado Let's Encrypt automáticamente

---

## 6. Configurar WhatsApp webhook

1. [developers.facebook.com](https://developers.facebook.com) → tu app → WhatsApp → Configuration
2. **Callback URL**: `https://TU_DOMINIO/webhook/whatsapp`
3. **Verify Token**: mismo valor que `WHATSAPP_VERIFY_TOKEN`
4. Subscribir al campo `messages`

---

## 7. Configurar Telegram webhook (opcional)

```bash
curl "https://api.telegram.org/botTU_BOT_TOKEN/setWebhook\
?url=https://TU_DOMINIO/webhook/telegram\
&secret_token=TU_TELEGRAM_SECRET_TOKEN"
```

---

## 8. Verificar

```bash
# Health check
curl https://TU_DOMINIO/health

# Jobs cron disponibles
curl https://TU_DOMINIO/api/triggers/jobs

# Disparar briefing manualmente
curl -X POST https://TU_DOMINIO/api/triggers/job/daily-briefing

# Ver pairing code (para autorizar nuevos senders)
curl https://TU_DOMINIO/api/pairing-code

# Probar chat
curl -X POST https://TU_DOMINIO/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"hola","sender":"test"}'

# Listar skills
curl https://TU_DOMINIO/api/skills

# Crear un skill nuevo
curl -X POST https://TU_DOMINIO/api/skills \
  -H "Content-Type: application/json" \
  -d '{"name":"test","content":"Sos un asistente de prueba.","tags":["test"]}'
```

---

## 9. Deploy automático

1. Coolify → configuración → **Webhooks** → copiar URL
2. GitHub → **Settings** → **Webhooks** → **Add webhook**
3. Pegar URL, content type `application/json`, trigger en `push`

Cada `git push` a `main` → rebuild + redeploy automático.

---

## 10. Testear localmente

```bash
make test          # Correr 280+ tests
make build         # Build local
make docker        # Docker local en http://localhost:8080
curl localhost:8080/health
```

---

## Arquitectura en producción

```
Internet
  │
  ├── WhatsApp (Meta) ──→ POST /webhook/whatsapp
  ├── Telegram ──────────→ POST /webhook/telegram
  └── Browser/CLI ───────→ GET/POST /api/*
  │
  ▼
Coolify (reverse proxy + SSL)
  │
  ▼
Docker: asistente (Go + SQLite)
  ├── Port 8080
  ├── /app/data/asistente.db  (volumen persistente)
  ├── /app/skills/             (skills hot-reload)
  └── /app/credentials.json    (Google service account)
```

**Cron jobs** (automáticos):
| Job | Hora | Descripción |
|-----|------|-------------|
| `daily-briefing` | 08:00 | Calendario + Gmail + gastos → resumen por WhatsApp |
| `weekly-finance` | 20:00 dom | Recordatorio de gastos semanales |
| `budget-alert` | 21:00 | Alerta de presupuesto |
| `daily-journal` | 22:00 | Prompt de journaling |
| `session-pruning` | 03:00 | Limpia sesiones > 30 días |

Todos los jobs se pueden disparar manualmente: `POST /api/triggers/job/:id`

---

## Troubleshooting

```bash
# Ver logs
ssh root@TU_IP
docker logs -f $(docker ps -q --filter name=asistente)

# Backup SQLite
docker cp $(docker ps -q --filter name=asistente):/app/data/asistente.db ./backup.db

# Reiniciar
# → Desde la UI de Coolify: botón Restart

# WhatsApp webhook no funciona
# 1. curl https://TU_DOMINIO/health → debe responder
# 2. Verify token debe coincidir
# 3. Meta requiere HTTPS (Coolify lo maneja)

# Telegram webhook no funciona
# curl "https://api.telegram.org/botTU_TOKEN/getWebhookInfo"
```

---

*Última actualización: 18 de marzo de 2026*
