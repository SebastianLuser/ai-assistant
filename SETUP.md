# Setup — Cosas que tenes que hacer vos antes del deploy

Esta guia lista todo lo que no se puede automatizar: cuentas a crear, API keys a generar, webhooks a configurar. Segui el orden sugerido.

Una vez que tengas todo, pasa a [DEPLOY.md](DEPLOY.md) para el deploy en Contabo + Cloudflare + Coolify.

---

## Checklist rapido

### Obligatorio (sin esto no arranca)
- [ ] Anthropic Claude API key
- [ ] WhatsApp Business Cloud API (5 valores)
- [ ] Google Service Account + `credentials.json` + Sheets + Calendar
- [ ] Dominio + cuenta Cloudflare
- [ ] VPS Contabo
- [ ] Secretos random (`WEBHOOK_SECRET`, `POSTGRES_PASSWORD`, `WHATSAPP_VERIFY_TOKEN`)

### Recomendado
- [ ] OpenAI API key (fallback del AI)
- [ ] Gmail habilitado en el mismo service account

### Opcional (por integracion)
- [ ] Notion, GitHub, Jira, Spotify, Todoist, ClickUp, Figma, Telegram, Obsidian

---

## 1. Claude API (OBLIGATORIA)

1. Entra a https://console.anthropic.com
2. Sign up con tu mail
3. **Settings** -> **Billing** -> carga credito (minimo USD 5)
4. **Settings** -> **API Keys** -> **Create Key** -> nombre: `jarvis-prod`
5. Copia la key que empieza con `sk-ant-...` y guardala en tu password manager

**Vars que te da:**
```
CLAUDE_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-sonnet-4-6
```

---

## 2. OpenAI API (recomendada como fallback)

1. https://platform.openai.com -> Sign up
2. **Billing** -> agregar metodo de pago y cargar saldo
3. **API keys** -> **Create new secret key** -> copia `sk-...`

**Vars:**
```
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4o
```

---

## 3. WhatsApp Business Cloud API (OBLIGATORIA)

Meta te da un numero de prueba gratis y 1000 conversaciones/mes.

### 3.1 Crear la app
1. Entra a https://developers.facebook.com
2. **My Apps** -> **Create App**
3. Tipo: **Business**
4. Nombre: `jarvis-personal`
5. Agrega producto **WhatsApp** -> **Set up**

### 3.2 API Setup (valores temporales de prueba)
En **WhatsApp** -> **API Setup**:
- **Phone number ID** (abajo del numero de prueba) -> copiar
- **Temporary access token** -> copiar (ATENCION: dura 24h, ver 3.4 para el permanente)
- Agrega tu numero personal en **To** para recibir mensajes de prueba

**Vars:**
```
WHATSAPP_PHONE_NUMBER_ID=...
WHATSAPP_ACCESS_TOKEN=...   # temporal; reemplazar en 3.4
WHATSAPP_TO_NUMBER=549...   # tu numero con codigo de pais, sin + ni espacios
```

### 3.3 App Secret
1. **App Settings** -> **Basic**
2. **App Secret** -> click **Show** -> copia

**Var:**
```
WHATSAPP_APP_SECRET=...
```

### 3.4 Access token permanente (IMPORTANTE para prod)
El temporal dura 24h. Para prod necesitas uno permanente via System User:

1. https://business.facebook.com -> **Business Settings**
2. **Users** -> **System Users** -> **Add** -> nombre `jarvis`, role **Admin**
3. Click el system user -> **Add Assets** -> **Apps** -> selecciona tu app -> permisos **Full control**
4. **Generate New Token**:
   - App: tu app
   - Token expiration: **Never**
   - Permisos: `whatsapp_business_messaging` + `whatsapp_business_management`
5. Copia el token generado (este es el que va a prod)

Reemplaza `WHATSAPP_ACCESS_TOKEN` con este.

### 3.5 Verify token (lo inventas vos)
```bash
openssl rand -hex 16
```

**Var:**
```
WHATSAPP_VERIFY_TOKEN=...
```

### 3.6 Configurar webhook
Esto lo haces DESPUES de deployar (ver [DEPLOY.md](DEPLOY.md) seccion 7).

---

## 4. Google (Sheets + Calendar + Gmail con un solo JSON)

### 4.1 Crear proyecto
1. https://console.cloud.google.com
2. Arriba a la izquierda -> **Select project** -> **New Project** -> nombre: `jarvis`
3. Esperar creacion y seleccionarlo

### 4.2 Habilitar APIs
**APIs & Services** -> **Library** -> buscar y habilitar una por una:
- Google Sheets API
- Google Calendar API
- Gmail API (solo si vas a usar lectura de mails)

### 4.3 Service Account
1. **APIs & Services** -> **Credentials**
2. **Create Credentials** -> **Service Account**
3. Nombre: `jarvis-service`
4. Skip los pasos opcionales -> **Done**
5. Click sobre el service account recien creado
6. **Keys** -> **Add Key** -> **Create new key** -> **JSON** -> se descarga el archivo
7. Renombralo a `credentials.json` y guardalo fuera del repo (el `.gitignore` ya lo cubre)
8. Copia el email del service account (`jarvis-service@jarvis-xxxxx.iam.gserviceaccount.com`)

### 4.4 Google Sheets
1. https://sheets.google.com -> crea un sheet nuevo llamado `Jarvis Finanzas`
2. Crea una tab llamada `Gastos` (con columnas: Fecha, Descripcion, Monto, Categoria, Moneda)
3. Click en **Share** -> pega el email del service account -> permiso **Editor** -> Send
4. De la URL copia el ID: `https://docs.google.com/spreadsheets/d/ESTE_ES_EL_ID/edit`

**Vars:**
```
GOOGLE_SHEETS_ID=...
GOOGLE_SHEETS_NAME=Gastos
GOOGLE_CREDENTIALS_FILE=credentials.json
```

### 4.5 Google Calendar
1. https://calendar.google.com -> en la barra lateral **My calendars** crea uno nuevo: `Jarvis`
2. Settings del calendario -> **Share with specific people** -> pega el email del service account -> permiso **Make changes to events**
3. Baja hasta **Integrate calendar** -> copia **Calendar ID** (termina en `@group.calendar.google.com`)

**Var:**
```
GOOGLE_CALENDAR_ID=...@group.calendar.google.com
```

### 4.6 Gmail (opcional, complicado)
Gmail via service account requiere Google Workspace con Domain-Wide Delegation. Si solo tenes Gmail personal, saltea esto. Si tenes Workspace:

1. Admin console -> **Security** -> **API controls** -> **Domain-wide delegation**
2. **Add new** -> Client ID = el del service account -> scopes:
   - `https://www.googleapis.com/auth/gmail.readonly`
   - `https://www.googleapis.com/auth/gmail.modify`

**Var:**
```
GMAIL_USER_EMAIL=tu@tudominio.com
```

---

## 5. Dominio + Cloudflare (OBLIGATORIO)

### 5.1 Dominio
Si no tenes uno:
- Namecheap, Porkbun, o Cloudflare Registrar (~USD 10/ano)

### 5.2 Cloudflare
1. https://dash.cloudflare.com -> **Add site** -> tu dominio
2. Plan gratuito
3. Cambia los nameservers en tu registrar por los que te da Cloudflare
4. Espera propagacion (5-30 min)

El DNS record y el resto de la config lo haces en [DEPLOY.md](DEPLOY.md) seccion 4.

---

## 6. VPS Contabo (OBLIGATORIO)

1. https://contabo.com -> **VPS**
2. Elegir **VPS S** (4 vCPU, 8 GB RAM, 50 GB NVMe, ~EUR 5/mes)
3. Region: EU
4. Image: Ubuntu 24.04
5. SSH key: pegar tu clave publica (`cat ~/.ssh/id_ed25519.pub`). Si no tenes una, generarla con `ssh-keygen -t ed25519`
6. Desactivar paneles opcionales (Webuzo/Plesk)
7. Checkout. Espera el mail con la IP (puede tardar 15 min a algunas horas)

---

## 7. Secretos random (generar localmente)

Corre estos comandos y guarda cada output en tu password manager:

```bash
# Password de Postgres
openssl rand -base64 32

# Webhook secret (protege /api/*)
openssl rand -hex 32

# WhatsApp verify token (ya generado en 3.5)
openssl rand -hex 16

# Telegram secret token (si vas a usar Telegram)
openssl rand -hex 16
```

**Vars:**
```
POSTGRES_PASSWORD=...
WEBHOOK_SECRET=...
WHATSAPP_VERIFY_TOKEN=...
TELEGRAM_SECRET_TOKEN=...
```

---

## 8. Integraciones opcionales

Saltealas si no las vas a usar. Cada una se activa sola si la var esta presente.

### 8.1 Notion
1. https://www.notion.so/my-integrations -> **New integration**
2. Type: Internal -> crear -> copiar **Internal Integration Token** (`secret_...`)
3. Abri una pagina de Notion -> `...` arriba a la derecha -> **Connections** -> agregar tu integracion
4. Copia el ID de la pagina (los 32 caracteres en la URL, sin guiones)

```
NOTION_API_KEY=secret_...
NOTION_DEFAULT_PAGE_ID=...
```

### 8.2 GitHub
1. github.com -> avatar -> **Settings** -> **Developer settings** -> **Personal access tokens** -> **Fine-grained tokens**
2. **Generate new token**:
   - Nombre: `jarvis`
   - Expiration: 1 year (o no expira)
   - Repositories: All o seleccionar
   - Permissions: Contents (read), Issues (read/write), Pull requests (read), Metadata (read)
3. Copia `github_pat_...`

```
GITHUB_TOKEN=github_pat_...
```

### 8.3 Jira
1. https://id.atlassian.com/manage-profile/security/api-tokens
2. **Create API token** -> nombre `jarvis` -> copia

```
JIRA_BASE_URL=https://tuempresa.atlassian.net
JIRA_EMAIL=tu@email.com
JIRA_API_TOKEN=...
```

### 8.4 Spotify
1. https://developer.spotify.com/dashboard -> **Create app**
2. Redirect URI: `http://localhost:8080/callback`
3. Necesitas un access token via Authorization Code flow con scopes `user-read-playback-state` + `user-modify-playback-state`
4. El token se vence, vas a necesitar refresh (o solo para demo)

```
SPOTIFY_ACCESS_TOKEN=...
```

### 8.5 Todoist
1. todoist.com -> **Settings** -> **Integrations** -> **Developer** -> copiar API token

```
TODOIST_API_TOKEN=...
```

### 8.6 ClickUp
1. clickup.com -> avatar -> **Settings** -> **Apps**
2. **Generate API Token** -> copiar (`pk_...`)
3. `CLICKUP_TEAM_ID` = el numero que aparece en la URL de tu workspace

```
CLICKUP_API_TOKEN=pk_...
CLICKUP_TEAM_ID=...
```

### 8.7 Figma
1. figma.com -> avatar -> **Settings** -> **Security** -> **Personal access tokens**
2. **Generate new token** -> nombre `jarvis` -> copiar

```
FIGMA_ACCESS_TOKEN=figd_...
```

### 8.8 Telegram
1. Abri Telegram -> busca [@BotFather](https://t.me/BotFather)
2. `/newbot` -> nombre y username (debe terminar en `bot`)
3. BotFather te da un token tipo `123456:ABC-DEF...`

```
TELEGRAM_BOT_TOKEN=123456:ABC...
TELEGRAM_BOT_USERNAME=tu_bot
TELEGRAM_SECRET_TOKEN=...   # el que generaste en paso 7
```

El webhook lo configuras post-deploy (ver DEPLOY.md seccion 8).

### 8.9 Obsidian
Solo tiene sentido si el vault esta en el mismo server (raro). Dejar vacio.

---

## 9. Armar tu .env final

Copia `.env.example` a `.env` y completa con todos los valores que juntaste.

```bash
cp .env.example .env
```

Abrilo con tu editor y pega cada valor en el lugar correspondiente.

**IMPORTANTE**: nunca commitees `.env` ni `credentials.json`. Ambos estan en `.gitignore`.

---

## 10. Listo para deploy

Cuando tengas:
- [x] `.env` completo con las vars obligatorias
- [x] `credentials.json` descargado
- [x] VPS Contabo con IP
- [x] Dominio apuntando a Cloudflare

Pasa a [DEPLOY.md](DEPLOY.md).

---

## Post-deploy: cosas manuales que quedan

Estas tres las hacen solo despues de que el servicio este corriendo con SSL:

1. **WhatsApp webhook**: Meta -> tu app -> WhatsApp -> Configuration -> Callback URL + Verify Token + suscribir a `messages`
2. **Telegram webhook** (si aplica): `curl` a `api.telegram.org/bot.../setWebhook`
3. **Cloudflare proxy naranja**: cambiar DNS record de gris a naranja cuando Let's Encrypt ya haya emitido el cert

Detalle completo en [DEPLOY.md](DEPLOY.md).
