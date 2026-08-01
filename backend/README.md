# Backend

Go REST API service for kebaikanku.id.

## Status

Implemented:
- Config loading from environment and optional `.env`.
- Database connection with GORM.
- SQLite for local development and PostgreSQL for production-style deployments.
- GORM AutoMigrate for local development and tracked SQL migrations for production.
- HTTP router and middleware stack using `go-chi/chi`.
- `GET /health` and `GET /readyz`.
- Public campaign list/detail endpoints.
- Temporary bearer-token campaign creation for pilot data.
- Donation creation with Midtrans Snap checkout.
- Midtrans notification callback with idempotent campaign total updates.
- CSV donation export for operator reconciliation.

Not implemented yet:
- Auth endpoints.
- JWT middleware and organization authorization.
- Full dashboard campaign management.

## Requirements

- Go `1.26.2` as declared in `go.mod`.
- SQLite for the default local database.
- PostgreSQL only when testing production-style database behavior.

## Local Setup

```bash
cp .env.example .env
go run cmd/api/main.go
```

The API listens on `http://localhost:8080` by default.

Health check:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/readyz
```

## Build

```bash
go build -o bin/api ./cmd/api
```

## Environment Variables

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP listen port. |
| `APP_ENV` | `development` | Runtime environment label. |
| `DB_DRIVER` | `sqlite` | Supported values: `sqlite`, `postgres`. |
| `DB_DSN` | `kebaikanku.db` | SQLite file path or PostgreSQL DSN. |
| `MIDTRANS_ENV` | `sandbox` | `sandbox` for development, `production` for live payments. |
| `MIDTRANS_SERVER_KEY` | empty | Midtrans server key. Keep secret. |
| `MIDTRANS_CLIENT_KEY` | empty | Midtrans client key for frontend Snap usage. |
| `MIDTRANS_NOTIFICATION_URL` | empty | Public webhook URL sent per Snap transaction through `X-Override-Notification`. |
| `MIDTRANS_NOTIFICATION_TOKEN` | empty | Optional internal token if the callback route is additionally guarded. |
| `ADMIN_PASSWORD` | empty | Single-admin password. If empty on first install, a random password is printed once and stored only as a bcrypt hash. Explicit production values require at least 12 characters. |
| `ADMIN_SESSION_SECRET` | empty | HMAC secret for admin session cookies. Production requires at least 32 characters. |
| `ADMIN_SETTINGS_ENCRYPTION_KEY` | empty | Base64-encoded 32-byte AES key used only to encrypt dashboard payment overrides. |
| `UPLOAD_DIR` | `uploads` | Persistent directory for uploaded campaign images. |
| `PUBLIC_UPLOAD_BASE_URL` | `/uploads` | Public URL prefix returned by the image upload endpoint. |
| `PUBLIC_LANDING_URL` | `http://127.0.0.1:18481` | Landing origin used for the Midtrans finish redirect (`/payments/{donation_id}`). |
| `SMTP_HOST` | empty | SMTP host for transactional email. If empty, waitlist email is skipped. |
| `SMTP_PORT` | `587` | SMTP port. Local dev uses `1025` with Mailpit-compatible SMTP. |
| `SMTP_USER` | empty | SMTP username. |
| `SMTP_PASS` | empty | SMTP password. |
| `SMTP_FROM` | empty | Sender email address. Required with `SMTP_HOST`. |
| `SMTP_FROM_NAME` | `kebaikanku.id` | Sender display name. |
| `SMTP_ENCRYPTION` | empty | Set to `null` for local SMTP servers that do not advertise AUTH/TLS. |
| `WAITLIST_ADMIN_EMAIL` | empty | Optional admin recipient for new waitlist signup notifications. |
| `WAITLIST_EMAIL_URL` | `https://kebaikanku.id/coming-soon` | Link included in waitlist confirmation emails. |

## Database Migrations

Development uses GORM AutoMigrate. Production (`APP_ENV=production`) deliberately skips it and requires the ordered SQL files in `migrations/` to be applied before the API starts. `docker-compose.production.yml` runs `migrate/migrate` against those files and tracks them in the separate `kebaikanku_migrations` table.

Never edit an applied migration. Add a new forward-only `NNNNNN_description.up.sql` file. Production rollback is a database restore plus the previously deployed API image; do not run destructive down migrations against donation data.

Local SQLite example:

```env
DB_DRIVER=sqlite
DB_DSN=kebaikanku.db
SMTP_HOST=100.65.30.81
SMTP_PORT=1025
SMTP_USER=any
SMTP_PASS=any
SMTP_FROM=testing@penadigital.id
SMTP_ENCRYPTION=null
SMTP_FROM_NAME=Penadigital Dev
WAITLIST_ADMIN_EMAIL=admin@example.com
WAITLIST_EMAIL_URL=http://127.0.0.1:4173/coming-soon
```

Local PostgreSQL example:

```env
DB_DRIVER=postgres
DB_DSN=host=localhost user=kebaikanku password=kebaikanku dbname=kebaikanku port=5432 sslmode=disable TimeZone=Asia/Jakarta
```

## API Surface

Current:
- `GET /health`
- `GET /readyz`
- `GET /api/v1/campaigns`
- `GET /api/v1/campaigns/{slug}`
- `POST /api/v1/admin/login`, `POST /api/v1/admin/logout`, and `GET /api/v1/admin/session`
- `POST /api/v1/campaigns` with an authenticated admin session
- `PUT /api/v1/campaigns/{id}` and `PATCH /api/v1/campaigns/{id}/status` with an authenticated admin session
- `POST /api/v1/donations`
- `GET /api/v1/donations/export` with an authenticated admin session
- `POST /api/v1/payments/midtrans/notification`

Multi-user institution accounts remain a later phase; the production pilot intentionally uses one environment-configured admin.

See [../docs/api.md](../docs/api.md) and [../docs/payment-gateway.md](../docs/payment-gateway.md).
