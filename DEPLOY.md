# Deploy Guide — Contabo + Cloudflare + Coolify

Esta guia asume que ya tenes todas las API keys. Si todavia no las tenes, mira [SETUP.md](SETUP.md) primero.

---

## Arquitectura

```
Internet
  |
  v
Cloudflare (DNS + proxy + DDoS)
  |
  v
Contabo VPS (Ubuntu 24.04)
  |
  v
Coolify (reverse proxy + Let's Encrypt SSL)
  |
  v
Docker: jarvis (Go) + postgres (red interna, no exposed)
```

---

## 1. Crear VPS en Contabo (~15 min)

1. Ir a [contabo.com](https://contabo.com) -> **VPS**
2. Elegir **VPS S** (4 vCPU, 8 GB RAM, 50 GB NVMe, ~EUR 5/mes)
3. Configuracion:
   - **Region**: European Union (o la mas cercana a vos)
   - **Image**: Ubuntu 24.04
   - **Object Storage / Add-ons**: ninguno
   - **Login**: SSH key (pega tu clave publica)
   - Desactiva cualquier panel web opcional (Webuzo/Plesk)
4. Checkout. Contabo tarda entre 15 min y algunas horas en aprovisionar (mas lento que Hetzner).
5. Cuando recibas el mail con la IP, anotala.

---

## 2. Hardening basico

```bash
ssh root@TU_IP
apt update && apt upgrade -y
apt install -y ufw fail2ban
ufw allow 22,80,443,8000/tcp
ufw --force enable
systemctl enable --now fail2ban
```

Opcional pero recomendado: crear usuario no-root y deshabilitar login root.

---

## 3. Instalar Coolify (~3-5 min)

```bash
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
```

Cuando termine, abri `http://TU_IP:8000` en el navegador y crea la cuenta admin. Anotala en un password manager.

---

## 4. Configurar Cloudflare

### 4.1 Agregar dominio
1. [dash.cloudflare.com](https://dash.cloudflare.com) -> **Add site** -> tu dominio
2. Plan gratuito
3. Cambia los nameservers en tu registrar (Namecheap/GoDaddy/etc) por los que te da Cloudflare
4. Espera la propagacion (5-30 min; Cloudflare te avisa cuando termina)

### 4.2 DNS record
1. **DNS** -> **Records** -> **Add record**
2. Type: `A`, Name: `jarvis`, Content: `IP_DE_CONTABO`
3. **Proxy status**: **DNS only** (nube gris) — IMPORTANTE, al principio desactivado para que Let's Encrypt pueda validar via HTTP-01

### 4.3 SSL
1. **SSL/TLS** -> **Overview** -> modo **Full (strict)**
2. **SSL/TLS** -> **Edge Certificates** -> activa **Always Use HTTPS**

### 4.4 Proxy naranja (despues del primer deploy)
Cuando ya tengas el sitio funcionando con cert Let's Encrypt, volve al DNS y cambia la nube a **Proxied** (naranja). Esto activa DDoS + cache + ocultamiento de IP.

### 4.5 Bypass cache para webhooks (importante)
**Rules** -> **Configuration Rules** -> **Create rule**
- Nombre: `webhooks-no-cache`
- When: `URI Path` `starts with` `/webhook/`
- Then: **Cache Level** = `Bypass`

Sin esto, Cloudflare puede cachear respuestas de webhooks de WhatsApp/Telegram y romper las entregas.

---

## 5. Proyecto en Coolify

### 5.1 Crear recurso
1. Coolify UI -> **Projects** -> **New Project** -> "jarvis"
2. **New Resource** -> **Public Repository**
3. URL: `https://github.com/Pineapple-Pixels/Jarvis`
4. Branch: `main`
5. **Build Pack**: Docker Compose

### 5.2 Variables de entorno
En **Environment Variables** pega todas las del `.env` (ver [SETUP.md](SETUP.md) para como conseguir cada una).

Como minimo:
```env
# DB (generar con: openssl rand -base64 32)
POSTGRES_PASSWORD=un-password-muy-fuerte

# AI
AI_PROVIDER=claude
CLAUDE_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-sonnet-4-6

# WhatsApp
WHATSAPP_PHONE_NUMBER_ID=...
WHATSAPP_ACCESS_TOKEN=...
WHATSAPP_TO_NUMBER=549...
WHATSAPP_VERIFY_TOKEN=...
WHATSAPP_APP_SECRET=...

# Seguridad
WEBHOOK_SECRET=...

# Google
GOOGLE_SHEETS_ID=...
GOOGLE_CALENDAR_ID=...@group.calendar.google.com
GOOGLE_CREDENTIALS_FILE=/app/credentials.json
```

### 5.3 Subir credentials.json
El archivo de Google service account no va en el repo. Dos opciones:

**Opcion A — scp directo al host:**
```bash
scp credentials.json root@IP_CONTABO:/data/coolify/projects/jarvis/credentials.json
```
Y en Coolify agrega un bind mount al servicio `jarvis`: `./credentials.json:/app/credentials.json:ro`.

**Opcion B — Coolify File Mounts:**
En la configuracion del servicio -> **Storages** / **File Mounts** -> crear archivo en path `/app/credentials.json` y pegar el contenido.

### 5.4 Dominio
1. Configuracion del recurso -> **Domains**
2. Agregar: `https://jarvis.tudominio.com`
3. Guardar. Coolify emite el cert Let's Encrypt automaticamente (requiere que Cloudflare este en **DNS only** durante la emision)

### 5.5 Deploy
Click **Deploy**. Mira los logs. La primera build tarda ~2-5 min.

---

## 6. Verificacion

```bash
curl https://jarvis.tudominio.com/health
# -> {"status":"ok"}

curl https://jarvis.tudominio.com/api/triggers/jobs
# lista de cron jobs registrados
```

Si `/health` responde, podes activar la nube naranja de Cloudflare (paso 4.4).

---

## 7. Configurar webhook de WhatsApp

1. [developers.facebook.com](https://developers.facebook.com) -> tu app -> **WhatsApp** -> **Configuration**
2. **Webhook** -> **Edit**:
   - **Callback URL**: `https://jarvis.tudominio.com/webhook/whatsapp`
   - **Verify Token**: el mismo valor de `WHATSAPP_VERIFY_TOKEN`
3. Click **Verify and Save**
4. **Webhook fields** -> suscribite a `messages`
5. Manda un WhatsApp al numero de prueba y confirma en los logs de Coolify

---

## 8. Configurar webhook de Telegram (opcional)

```bash
curl "https://api.telegram.org/botTU_BOT_TOKEN/setWebhook?url=https://jarvis.tudominio.com/webhook/telegram&secret_token=TU_TELEGRAM_SECRET_TOKEN"
```

Verificar:
```bash
curl "https://api.telegram.org/botTU_BOT_TOKEN/getWebhookInfo"
```

---

## 9. Deploy automatico (CI/CD)

1. Coolify -> configuracion del recurso -> **Webhooks** -> copia la URL
2. GitHub -> repo -> **Settings** -> **Webhooks** -> **Add webhook**
3. Payload URL: la de Coolify
4. Content type: `application/json`
5. Events: `Just the push event`
6. Active: si

Cada `git push` a `main` dispara rebuild + redeploy automatico.

---

## 10. Backups

### Postgres
Coolify soporta backups automaticos a S3. Configuracion del recurso -> **Backups**:
- Schedule: `0 3 * * *` (3 AM diario)
- Retention: 7 dias
- Destination: S3/Backblaze/cualquier S3-compatible

### credentials.json + .env
Guarda copias en tu password manager (1Password/Bitwarden). Estos archivos no estan en git.

---

## Troubleshooting

```bash
# Ver logs en vivo
ssh root@IP_CONTABO
docker logs -f $(docker ps -q --filter name=jarvis)

# Entrar al container
docker exec -it $(docker ps -q --filter name=jarvis) sh

# Ver estado de Postgres
docker exec -it $(docker ps -q --filter name=postgres) psql -U jarvis -d jarvis -c "\dt"

# Ver migraciones aplicadas
docker exec -it $(docker ps -q --filter name=postgres) psql -U jarvis -d jarvis -c "SELECT * FROM schema_migrations;"

# Reiniciar servicio
# -> Desde la UI de Coolify: boton Restart
```

### Problemas comunes

- **Let's Encrypt no emite cert**: Cloudflare debe estar en **DNS only** (nube gris) durante la emision. Activa el proxy naranja recien despues.
- **WhatsApp no entrega mensajes**: verifica que la Rule de bypass cache para `/webhook/*` este activa.
- **`POSTGRES_PASSWORD is required`** al deployar: falta la var de entorno en Coolify.
- **502 Bad Gateway**: el container esta buildeando o crasheo. Mira logs.
- **Webhook verify falla**: el `WHATSAPP_VERIFY_TOKEN` en Meta y en Coolify deben ser exactamente iguales.

---

## Cron jobs

Corren automaticamente dentro del container:

| Job | Hora | Descripcion |
|-----|------|-------------|
| `daily-briefing` | 08:00 | Calendario + Gmail + gastos -> resumen por WhatsApp |
| `weekly-finance` | 20:00 dom | Recordatorio de gastos semanales |
| `budget-alert` | 21:00 | Alerta de presupuesto |
| `daily-journal` | 22:00 | Prompt de journaling |
| `session-pruning` | 03:00 | Limpia sesiones > 30 dias |

Disparar manualmente:
```bash
curl -X POST https://jarvis.tudominio.com/api/triggers/job/daily-briefing
```

---

## Costo mensual aproximado

| Item | Costo |
|------|-------|
| Contabo VPS S | EUR 5 |
| Cloudflare | Gratis |
| Dominio | ~USD 10/ano |
| Claude API (uso personal) | ~USD 5-15 |
| WhatsApp Business | Gratis (1000 conv/mes) |
| **Total** | **~USD 12-22/mes** |
