# Backend

Go REST API service for kebaikanku.id.

## Status

Implemented:
- Config loading from environment and optional `.env`.
- Database connection with GORM.
- SQLite for local development and PostgreSQL for production-style deployments.
- GORM auto-migration for `Organization`, `Campaign`, `Donor`, and `Donation`.
- HTTP router and middleware stack using `go-chi/chi`.
- `GET /health`.

Not implemented yet:
- Auth endpoints.
- Campaign endpoints.
- Donation endpoints.
- Midtrans payment creation.
- Midtrans notification callback.
- JWT middleware and organization authorization.

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
| `MIDTRANS_NOTIFICATION_TOKEN` | empty | Optional internal token if the callback route is additionally guarded. |
| `JWT_SECRET` | empty | Required once auth is implemented. |

## Database Notes

The current backend runs `AutoMigrate` on startup. This is useful while the schema is young, but production should eventually move to explicit migrations with a tool such as `golang-migrate`.

Local SQLite example:

```env
DB_DRIVER=sqlite
DB_DSN=kebaikanku.db
```

Local PostgreSQL example:

```env
DB_DRIVER=postgres
DB_DSN=host=localhost user=kebaikanku password=kebaikanku dbname=kebaikanku port=5432 sslmode=disable TimeZone=Asia/Jakarta
```

## API Surface

Current:
- `GET /health`

Planned MVP:
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/campaigns`
- `POST /api/v1/campaigns`
- `POST /api/v1/donations`
- `POST /api/v1/payments/midtrans/notification`

See [../docs/api.md](../docs/api.md) and [../docs/payment-gateway.md](../docs/payment-gateway.md).
