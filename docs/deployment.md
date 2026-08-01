# Deployment Guide

This document describes the intended deployment shape for kebaikanku.id.

## Components

| Component | Path | Suggested target |
| --- | --- | --- |
| Landing page | `frontend/landing` | Cloudflare Pages |
| Dashboard app | `frontend/dashboard` | Static SPA host such as Cloudflare Pages |
| Backend API | `backend` | Container host, VM, Fly.io, Railway, Render, or Kubernetes |
| Database | SQLite/PostgreSQL | SQLite for small self-hosted trials, PostgreSQL for production |
| Payment gateway | Midtrans | Sandbox for dev, production for live donations |

### OpenAI Sites compatibility

The current static SvelteKit frontends can be adapted for Sites, but the repository is not yet a Sites project: it has no `.openai/hosting.json` and does not emit the required Cloudflare Worker-compatible `dist/server/index.js`. Sites does not run a persistent Go or Bun server process. A Sites-native full stack must expose Worker-compatible JavaScript/TypeScript handlers and use supported bindings such as D1 for relational data and R2 for uploads. The existing Go API, PostgreSQL database, and Midtrans webhook can remain on an external HTTPS backend while Sites hosts the landing page or dashboard.

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

### Environment Variables (SEO Configuration)

The static landing page is configured at build-time using public environment variables. In production, these should be set in your Cloudflare Pages dashboard (or equivalent hosting platform) environment configuration.

| Variable | Description | Example / Default |
| --- | --- | --- |
| `PUBLIC_SEO_SITE_NAME` | The brand or site name used in metadata. | `kebaikanku.id` |
| `PUBLIC_SEO_TITLE` | The default browser title bar text for the home page. | `kebaikanku.id - Transparent Donation Platform` |
| `PUBLIC_SEO_DESCRIPTION` | The default site description meta tag for search engines. | `Open-source crowdfunding and donation platform for zakat and kemanusiaan institutions. 0% platform fee for self-hosted.` |
| `PUBLIC_SEO_SITE_URL` | The absolute canonical root URL of the website. | `https://kebaikanku.id` |
| `PUBLIC_SEO_IMAGE` | The path or absolute URL of the Open Graph sharing banner. | `/images/og-image.png` |
| `PUBLIC_SEO_KEYWORDS` | Comma-separated list of keywords for metadata tags. | `donasi, zakat, open-source, crowdfunding, infak, sedekah` |
| `PUBLIC_GA_TRACKING_ID` | Google Analytics (GA4) measurement ID (optional, e.g., `G-XXXXXXXXXX`). | `""` (disabled if blank) |
| `PUBLIC_GTM_CONTAINER_ID` | Google Tag Manager Container ID (optional, e.g., `GTM-XXXXXXX`). | `""` (disabled if blank) |


## Backend Deployment

The production package is the API image, PostgreSQL, and a one-shot migration service. It does not expose PostgreSQL outside Docker; put a TLS reverse proxy in front of the API.

1. Copy `.env.production.example` to `.env.production` and replace every placeholder. Keep this file only in the host secret store.
2. Use a long random PostgreSQL password and URL-encode it in `MIGRATION_DATABASE_URL` if it has special URL characters.
3. Start the stack:

```bash
docker compose --env-file .env.production -f docker-compose.production.yml up -d --build
```

4. Confirm migrations completed and the API is ready:

```bash
docker compose --env-file .env.production -f docker-compose.production.yml logs migrate
curl http://127.0.0.1:8080/readyz
```

The one-shot `bootstrap` service provisions the controlled-pilot organization before the API starts. Set `PILOT_ORGANIZATION_ID`, `PILOT_ORGANIZATION_NAME`, and `PILOT_ORGANIZATION_EMAIL`; configure the dashboard build with the same public `PUBLIC_ORGANIZATION_ID` (or retain the default `pilot-org`). Organization self-service belongs to the later multi-institution auth work.

`/health` is a liveness response; `/readyz` is the endpoint for load balancers and deployment verification. The backend runs with `APP_ENV=production`, so it will not change the schema itself.

Required production environment variables:

```env
APP_ENV=production
PORT=8080
DB_DRIVER=postgres
DB_DSN=host=... user=... password=... dbname=... port=5432 sslmode=require TimeZone=Asia/Jakarta
MIGRATION_DATABASE_URL=postgres://...?...&sslmode=require&x-migrations-table=kebaikanku_migrations
CORS_ALLOWED_ORIGINS=https://kebaikanku.id,https://app.kebaikanku.id
ADMIN_PASSWORD=...
ADMIN_SESSION_SECRET=...
ADMIN_SETTINGS_ENCRYPTION_KEY=... # openssl rand -base64 32
UPLOAD_DIR=/data/uploads
PUBLIC_UPLOAD_BASE_URL=https://api.kebaikanku.id/uploads
PUBLIC_LANDING_URL=https://kebaikanku.id
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

Create a compressed PostgreSQL backup:

```bash
./scripts/backup-postgres.sh
```

Schedule it from the host at least daily and copy the resulting `backups/*.dump` off-host. Keep a retention period appropriate to the institution's reconciliation policy, and test restore monthly.

To restore, first stop the API, then run the guarded script and start the migration/API services again:

```bash
docker compose --env-file .env.production -f docker-compose.production.yml stop api
RESTORE_CONFIRM=restore ./scripts/restore-postgres.sh backups/kebaikanku-YYYYMMDDTHHMMSSZ.dump
docker compose --env-file .env.production -f docker-compose.production.yml up -d migrate api
curl http://127.0.0.1:8080/readyz
```

For an application rollback, restore the last verified backup first, then deploy the previous API image. Migrations are forward-only because deleting payment records during a down migration is not safe.

Production deployments should also define:

- Daily database backups.
- Backup retention period.
- Restore test cadence.
- Keep the `/data/uploads` volume persistent and include it in the backup policy for campaign banners.

## CI

GitHub Actions runs:

- `go test ./...` in `backend`.
- `go vet ./...` in `backend`.
- `npm ci` and `npm run build` in both frontend applications.
- API container build.

## Deployment Checklist

- Production database is PostgreSQL.
- `APP_ENV=production`.
- Midtrans production keys are configured.
- Midtrans notification URL points to `/api/v1/payments/midtrans/notification`.
- TLS is enabled.
- CORS origins match deployed frontend domains.
- `CORS_ALLOWED_ORIGINS` lists only those HTTPS origins.
- Secrets are stored in the platform secret manager, not committed files.
- Health check URL is monitored.
- `/readyz` is configured as the deployment/load-balancer readiness probe.
- Migration job completed successfully before the API release.
- A backup was taken and its restore procedure has been tested.
