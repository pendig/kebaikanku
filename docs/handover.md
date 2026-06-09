# Project Handover & Developer Specification

This document is the practical handover guide for developers and AI agents continuing **kebaikanku.id**.

## Project Status

As of June 9, 2026, the repository is in the MVP bootstrap phase.

Completed:

1. Repository setup and licensing.
2. Root `README.md` and `CONTRIBUTING.md`.
3. Go backend skeleton in `backend`.
4. Config loading through environment variables and optional `.env`.
5. GORM database initialization with SQLite/PostgreSQL selection.
6. Core domain models: `Organization`, `Campaign`, `Donor`, `Donation`.
7. `GET /health` route.
8. SvelteKit static landing page in `frontend/landing`.
9. Landing routes: `/`, `/terms`, `/privacy`.
10. Lightweight Indonesian/English localization.
11. Developer docs for API, deployment, security, payment gateway, roadmap, database, architecture, and localization.
12. Local PostgreSQL `docker-compose.yml`.
13. Environment examples.

Not completed yet:

1. `frontend/dashboard`.
2. Auth endpoints.
3. Campaign endpoints.
4. Donation endpoint.
5. Midtrans integration implementation.
6. Midtrans payment notification callback.
7. JWT middleware.
8. CI/CD workflows.
9. Explicit database migrations.

## Current Codebase

Important files:

- `backend/cmd/api/main.go`: backend entrypoint and current route setup.
- `backend/internal/config/config.go`: environment config loader.
- `backend/internal/database/database.go`: GORM connection and auto-migration.
- `backend/internal/domain/*.go`: current domain models.
- `frontend/landing/src/routes/+page.svelte`: landing homepage.
- `frontend/landing/src/routes/+layout.svelte`: shared layout, navigation, and footer.
- `frontend/landing/src/lib/i18n.js`: localization store.
- `frontend/landing/src/lib/locales/id.json`: Indonesian translations.
- `frontend/landing/src/lib/locales/en.json`: English translations.

## Running Locally

Backend:

```bash
cd backend
cp .env.example .env
go run cmd/api/main.go
```

Health check:

```bash
curl http://localhost:8080/health
```

Landing page:

```bash
cd frontend/landing
npm install
npm run dev
```

Local PostgreSQL:

```bash
docker compose up -d postgres
```

## Docs Map

- `docs/architecture.md`: system architecture and boundaries.
- `docs/database.md`: GORM and schema model plan.
- `docs/api.md`: planned REST API contract.
- `docs/payment-gateway.md`: Midtrans-first payment design.
- `docs/deployment.md`: Cloudflare Pages, backend, database, and production checklist.
- `docs/security.md`: auth, payment, CORS, AI, logging, and pre-launch notes.
- `docs/localization.md`: i18n strategy.
- `docs/roadmap.md`: dependency-ordered milestone plan.

## Payment Gateway Decision

The initial payment gateway is **Midtrans**.

Implement Midtrans before adding Xendit or another provider. The first production-safe payment milestone should include:

- Pending local donation creation.
- Midtrans Snap/Core API transaction creation.
- Midtrans notification signature verification.
- Idempotent callback processing.
- Transactional update of donation status and campaign collected amount.

See `docs/payment-gateway.md`.

## Recommended Next Implementation Order

1. Add backend repository and service layers for current domain models.
2. Implement auth register/login with password hashing and JWT.
3. Implement public campaign list and authorized campaign creation.
4. Add Midtrans config and client.
5. Implement donation creation through Midtrans.
6. Implement Midtrans notification callback.
7. Scaffold `frontend/dashboard`.
8. Add CI checks for backend and landing builds.

## Guardrails

- Keep core domain logic separate from HTTP handlers and external provider code.
- Keep Midtrans integration behind a small adapter boundary so future providers can be added later.
- Do not log secrets, JWTs, Midtrans server keys, or full donor PII.
- Treat payment callbacks as untrusted until verified.
- Avoid broad CORS wildcards in production.
- Keep all user-facing frontend strings in locale files.
