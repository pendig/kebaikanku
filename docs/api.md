# API Contract

This document defines the planned MVP REST API contract for kebaikanku.id.

Current implementation status:
- `GET /health` is implemented.
- `GET /api/v1/campaigns` and `GET /api/v1/campaigns/{slug}` are implemented.
- `POST /api/v1/campaigns` is implemented with a temporary `CAMPAIGN_ADMIN_TOKEN` bearer token for pilot data.
- `POST /api/v1/donations` creates local pending donations and returns Midtrans Snap checkout data.
- `POST /api/v1/payments/midtrans/notification` verifies Midtrans signatures and applies idempotent status updates.
- `GET /api/v1/donations/export` returns a CSV export for operator reconciliation.
- Auth endpoints remain planned.

## Conventions

Base path:

```text
/api/v1
```

Payload format:
- Request body: JSON.
- Response body: JSON.
- Authenticated endpoints use `Authorization: Bearer <jwt>`.

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

### `POST /api/v1/auth/register`

Registers an institution account.

Request:

```json
{
  "name": "Yayasan Contoh",
  "email": "admin@example.org",
  "password": "strong-password",
  "address": "Jakarta"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "organization": {
      "id": "uuid",
      "name": "Yayasan Contoh",
      "email": "admin@example.org",
      "status": "pending"
    }
  }
}
```

### `POST /api/v1/auth/login`

Authenticates an institution account.

Request:

```json
{
  "email": "admin@example.org",
  "password": "strong-password"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "token": "jwt",
    "organization": {
      "id": "uuid",
      "name": "Yayasan Contoh",
      "email": "admin@example.org",
      "status": "active"
    }
  }
}
```

## Campaigns

### `GET /api/v1/campaigns`

Public list of active campaigns.

Query parameters:
- `category`
- `page`
- `limit`

### `POST /api/v1/campaigns`

Creates a campaign. For the MVP pilot, this requires `Authorization: Bearer <CAMPAIGN_ADMIN_TOKEN>`. Replace this with organization JWT once dashboard auth exists.

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

Returns donation reconciliation data as CSV. Requires `Authorization: Bearer <CAMPAIGN_ADMIN_TOKEN>`.

Response:

```http
Content-Type: text/csv
```

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
