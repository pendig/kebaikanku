# Release Alpha Checklist

Target release: `v0.2.0-alpha`.

## Scope

- Public campaign list from `GET /api/v1/campaigns`.
- Donor checkout form on `/campaigns`.
- Midtrans Snap redirect from `POST /api/v1/donations`.
- Midtrans notification callback updates donations idempotently.
- Donation CSV export for operator reconciliation.

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
CAMPAIGN_ADMIN_TOKEN=change-me
MIDTRANS_ENV=sandbox
MIDTRANS_SERVER_KEY=SB-Mid-server-...
MIDTRANS_CLIENT_KEY=SB-Mid-client-...
```

## Sandbox E2E

1. Start backend:

```bash
cd backend
go run cmd/api/main.go
```

2. Create a campaign:

```bash
curl -X POST http://localhost:8080/api/v1/campaigns \
  -H "Authorization: Bearer $CAMPAIGN_ADMIN_TOKEN" \
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

4. Open `/campaigns`, submit the donation form, and confirm it redirects to Midtrans sandbox.

5. After payment completion, verify donation export:

```bash
curl http://localhost:8080/api/v1/donations/export \
  -H "Authorization: Bearer $CAMPAIGN_ADMIN_TOKEN"
```

## Release Gate

- `go test ./...` passes in `backend`.
- `npm run build` passes in `frontend/landing`.
- Sandbox payment redirects to Midtrans.
- Sandbox notification marks the donation as `success`.
