# Release Alpha Checklist

Target release: `v0.2.1-alpha`.

## Scope

- Public campaign list from `GET /api/v1/campaigns`.
- Donor checkout form on `/campaigns`.
- Midtrans Snap redirect from `POST /api/v1/donations`.
- Per-transaction Midtrans notification routing and idempotent callback updates.
- Direct Midtrans status reconciliation and donor-facing success, pending, failed, and expired pages.
- Password-authenticated admin dashboard with campaign management, encrypted payment settings, uploads, pagination, filters, and donation CSV export.
- Production migrations, Docker Compose, backup/restore scripts, and CI gates.

## Frontend Environment

Use an empty value when the frontend and backend share the same origin:

```env
PUBLIC_API_BASE_URL=
```

Use an absolute URL when the backend is hosted elsewhere:

```env
PUBLIC_API_BASE_URL=https://api.kebaikanku.id
```

This works on Cloudflare Pages, static hosting, VPS, Render, Railway, Fly.io, or any host that can serve the SvelteKit static build.

## Backend Environment

Required for sandbox E2E:

```env
APP_ENV=development
DB_DRIVER=sqlite
DB_DSN=kebaikanku-alpha.db
ADMIN_PASSWORD=change-this-admin-password
ADMIN_SESSION_SECRET=change-this-session-secret-at-least-32-characters
MIDTRANS_ENV=sandbox
MIDTRANS_SERVER_KEY=SB-Mid-server-...
MIDTRANS_CLIENT_KEY=SB-Mid-client-...
MIDTRANS_NOTIFICATION_URL=https://replace-with-your-public-host/api/v1/payments/midtrans/notification
MIDTRANS_NOTIFICATION_TOKEN=
```

`MIDTRANS_NOTIFICATION_URL` must be replaced with an HTTPS endpoint reachable by Midtrans. For a local E2E run, expose the backend first with `cloudflared tunnel --url http://localhost:8080`, then use the generated hostname plus `/api/v1/payments/midtrans/notification`. A deployed forwarding Worker may be used instead.

## Sandbox E2E

1. Start backend:

```bash
cd backend
go run cmd/api/main.go
```

2. Login and create a campaign with the session cookie:

```bash
curl -c /tmp/kebaikanku-admin-cookie.txt -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"password":"change-this-admin-password"}'

curl -b /tmp/kebaikanku-admin-cookie.txt -X POST http://localhost:8080/api/v1/campaigns \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "pilot-org",
    "title": "Bantu Renovasi Madrasah",
    "slug": "bantu-renovasi-madrasah",
    "description": "Campaign pilot untuk validasi pembayaran sandbox.",
    "category": "infak",
    "target_amount": 10000000,
    "end_date": "2026-12-31T23:59:59+07:00"
  }'
```

3. Start frontend:

```bash
cd frontend/landing
PUBLIC_API_BASE_URL=http://localhost:8080 npm run dev
```

4. Validate the production frontend build:

```bash
cd frontend/landing
npm run build
```

5. Open `/campaigns`, submit the donation form, and confirm it redirects to Midtrans sandbox.

6. After payment completion, verify donation export:

```bash
curl -b /tmp/kebaikanku-admin-cookie.txt http://localhost:8080/api/v1/donations/export
```

Confirm the exported row has `status` set to `success` and `provider_status` set to `settlement` or `capture`.

## Release Gate

- `go test ./...` passes in `backend`.
- `go vet ./...` passes in `backend`.
- `npm run build` passes in `frontend/landing`.
- `npm run build` passes in `frontend/dashboard`.
- Production migrations and container build pass in CI.
- Sandbox payment redirects to Midtrans.
- Sandbox notification marks the donation as `success`; verify it from `GET /api/v1/donations/export`.
