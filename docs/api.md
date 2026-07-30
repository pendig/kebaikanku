# API Contract

This document defines the planned MVP REST API contract for kebaikanku.id.

Current implementation status:
- `GET /health` is implemented.
- `GET /api/v1/campaigns` and `GET /api/v1/campaigns/{slug}` are implemented.
- Single-admin login, logout, and session endpoints are implemented with a signed HTTP-only cookie.
- Campaign mutations and donation export require an authenticated admin session.
- `POST /api/v1/donations` creates local pending donations and returns Midtrans Snap checkout data.
- `POST /api/v1/payments/midtrans/notification` verifies Midtrans signatures and applies idempotent status updates.
- `GET /api/v1/donations/export` returns a CSV export for operator reconciliation.

## Conventions

Base path:

```text
/api/v1
```

Payload format:
- Request body: JSON.
- Response body: JSON.
- Authenticated admin endpoints use the session cookie issued by `/api/v1/admin/login`.

Success response shape:

```json
{
  "success": true,
  "data": {}
}
```

Error response shape:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Field 'amount' must be greater than zero"
  }
}
```

## Health

### `GET /health`

Returns server health and current server time.

## Waitlist

### `POST /api/v1/waitlist`

Registers an email for the landing-page waitlist.
When SMTP is configured, the API queues a confirmation email after the address is stored.
When `WAITLIST_ADMIN_EMAIL` is configured, the API also queues an admin notification.

Request:

```json
{
  "email": "admin@example.org",
  "website": "",
  "source": "landing-coming-soon"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "8b2f9e9e-0e5a-4db6-b6a3-f6f8b4f6dd1e",
    "email_queued": true,
    "admin_email_queued": true
  }
}
```

Error examples:

```json
{
  "success": false,
  "error": {
    "code": "DUPLICATE_EMAIL",
    "message": "Email is already registered in the waitlist."
  }
}
```

## Auth

### `POST /api/v1/admin/login`

Authenticates the single pilot administrator using `ADMIN_PASSWORD` and sets the session cookie.

Request:

```json
{
  "password": "strong-password"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "authenticated": true
  }
}
```

`GET /api/v1/admin/session` reports whether the request has a valid session. `POST /api/v1/admin/logout` expires it.

## Campaigns

### `GET /api/v1/campaigns`

Public list of active campaigns.

Query parameters:
- `category`
- `page`
- `limit`

### `POST /api/v1/campaigns`

Creates a campaign. Requires an authenticated admin session.

Request:

```json
{
  "title": "Bantu Renovasi Madrasah",
  "slug": "bantu-renovasi-madrasah",
  "description": "Campaign description",
  "category": "zakat_maal",
  "target_amount": 100000000,
  "end_date": "2026-12-31T23:59:59+07:00"
}
```

## Donations

### `POST /api/v1/donations`

Creates a donation and starts a Midtrans payment flow.

Request:

```json
{
  "campaign_id": "uuid",
  "donor": {
    "name": "Budi",
    "phone_number": "+6281234567890",
    "email": "budi@example.com"
  },
  "amount": 250000,
  "platform_tip": 5000,
  "payment_method": "midtrans_snap"
}
```

### `GET /api/v1/donations/export`

Returns donation reconciliation data as CSV. Requires an authenticated admin session.

Response:

```http
Content-Type: text/csv
```

### `GET /api/v1/admin/donations`

Returns authenticated, paginated donation JSON. Supports `page`, `limit` (maximum 100), and `status=pending|success|failed`.

### `POST /api/v1/admin/uploads`

Accepts an authenticated multipart field named `file`. JPEG, PNG, and WebP images up to 5 MB are stored under a random filename and return `{ "data": { "url": "..." } }`.

### `GET|PUT /api/v1/admin/settings/payment`

Reads or changes the effective Midtrans mode (`sandbox` or `production`). A PUT may include `server_key` and `client_key` together; overrides are encrypted with `ADMIN_SETTINGS_ENCRYPTION_KEY`, take precedence over environment values, and are never returned by the API. Sending neither key resets the override to environment keys for the selected mode.

```csv
id,campaign,donor_name,donor_phone,amount,platform_tip,status,provider_status,created_at,paid_at
8f89c67a-12bc-401d-9e12-3a8bc02d41ab,Bantu Renovasi Madrasah,Budi,+6281234567890,250000,5000,success,settlement,2026-06-22T19:10:00+07:00,2026-06-22T19:12:00+07:00
```

## Payments

### `POST /api/v1/payments/midtrans/notification`

Receives Midtrans payment notifications.

Rules:
- Verify Midtrans signature before mutating data.
- Treat notifications as idempotent.
- Store the provider transaction ID and raw status transition for auditability.
- Only increment `Campaign.CollectedAmount` once for each successful donation.

See [payment-gateway.md](payment-gateway.md) for the Midtrans lifecycle design.
