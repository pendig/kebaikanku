# Deployment Guide

This document describes the intended deployment shape for kebaikanku.id.

## Components

| Component | Path | Suggested target |
| --- | --- | --- |
| Landing page | `frontend/landing` | Cloudflare Pages |
| Dashboard app | `frontend/dashboard` | Cloudflare Pages SPA, planned |
| Backend API | `backend` | Container host, VM, Fly.io, Railway, Render, or Kubernetes |
| Database | SQLite/PostgreSQL | SQLite for small self-hosted trials, PostgreSQL for production |
| Payment gateway | Midtrans | Sandbox for dev, production for live donations |

## Domains

Suggested domains:

- `kebaikanku.id`: static landing page.
- `app.kebaikanku.id`: dashboard SPA.
- `api.kebaikanku.id`: Go backend API.

The backend CORS configuration should allow only the active landing and dashboard origins.

## Landing Page on Cloudflare Pages

Settings:

| Setting | Value |
| --- | --- |
| Root directory | `frontend/landing` |
| Build command | `npm run build` |
| Build output directory | `build` |
| Node version | `18` or newer |

## Backend Deployment

Build:

```bash
cd backend
go build -o bin/api ./cmd/api
```

Run:

```bash
PORT=8080 ./bin/api
```

Required production environment variables:

```env
APP_ENV=production
PORT=8080
DB_DRIVER=postgres
DB_DSN=host=... user=... password=... dbname=... port=5432 sslmode=require TimeZone=Asia/Jakarta
JWT_SECRET=...
MIDTRANS_ENV=production
MIDTRANS_SERVER_KEY=...
MIDTRANS_CLIENT_KEY=...
```

## Local PostgreSQL

The root `docker-compose.yml` starts a local PostgreSQL database for development:

```bash
docker compose up -d postgres
```

Use:

```env
DB_DRIVER=postgres
DB_DSN=host=localhost user=kebaikanku password=kebaikanku dbname=kebaikanku port=5432 sslmode=disable TimeZone=Asia/Jakarta
```

## Database Strategy

SQLite is acceptable for:
- Local development.
- Demo deployments.
- Very small self-hosted pilots.

PostgreSQL is recommended for:
- Managed cloud.
- Production donation processing.
- Multi-admin dashboard usage.
- Payment reconciliation and reporting.

## Backups

Production deployments should define:

- Daily database backups.
- Backup retention period.
- Restore test cadence.
- Separate storage for uploaded receipts and report assets once uploads exist.

## CI/CD Target

Recommended GitHub Actions checks:

- `go test ./...` in `backend`.
- `go vet ./...` in `backend`.
- `npm ci` and `npm run build` in `frontend/landing`.
- Later: dashboard build and API integration tests.

## Deployment Checklist

- Production database is PostgreSQL.
- `APP_ENV=production`.
- Midtrans production keys are configured.
- Midtrans notification URL points to `/api/v1/payments/midtrans/notification`.
- TLS is enabled.
- CORS origins match deployed frontend domains.
- Secrets are stored in the platform secret manager, not committed files.
- Health check URL is monitored.
