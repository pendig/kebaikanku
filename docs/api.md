# API Contract

This document defines the planned MVP REST API contract for kebaikanku.id.

Current implementation status:
- `GET /health` is implemented.
- All `/api/v1/*` endpoints below are planned and should be implemented before dashboard integration.

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
When SMTP is configured, the API sends a confirmation email after the address is stored.

Request:

```json
{
  "email": "admin@example.org",
  "website": "",
  "source": "landing-coming-soon",
  "submitted_at": 1718200000
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "8b2f9e9e-0e5a-4db6-b6a3-f6f8b4f6dd1e",
    "email_sent": true
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

Creates a campaign. Requires organization JWT.

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

Response:

```json
{
  "success": true,
  "data": {
    "donation_id": "uuid",
    "status": "pending",
    "payment": {
      "provider": "midtrans",
      "snap_token": "snap-token",
      "redirect_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-token"
    }
  }
}
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
