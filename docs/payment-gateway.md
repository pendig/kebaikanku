# Payment Gateway

The initial payment gateway for kebaikanku.id is **Midtrans**.

Xendit is intentionally left as a future provider option. The first MVP should make the donation lifecycle reliable with Midtrans before introducing a provider abstraction with multiple production backends.

## Goals

- Create donation payments through Midtrans.
- Receive and verify Midtrans notifications.
- Keep donation status transitions idempotent.
- Update campaign collected amounts exactly once per successful donation.
- Preserve enough payment metadata for reconciliation and support.

## Recommended Midtrans Flow

1. Donor submits a donation request to `POST /api/v1/donations`.
2. Backend creates a local `Donation` row with status `pending`.
3. Backend creates a Midtrans transaction using Snap or Core API.
4. Backend returns a Snap token or redirect URL to the frontend.
5. Donor completes payment on Midtrans.
6. Midtrans sends a notification to `POST /api/v1/payments/midtrans/notification`.
7. Backend verifies the notification signature.
8. Backend maps the Midtrans status to an internal donation status.
9. Backend updates the donation and campaign inside one database transaction.

## Environment Variables

```env
MIDTRANS_ENV=sandbox
MIDTRANS_SERVER_KEY=SB-Mid-server-your-server-key
MIDTRANS_CLIENT_KEY=SB-Mid-client-your-client-key
MIDTRANS_NOTIFICATION_TOKEN=
```

`MIDTRANS_NOTIFICATION_TOKEN` is optional and should not replace Midtrans signature verification. It can be used only as an additional internal guard if the deployed route sits behind a gateway rule.

## Internal Status Mapping

| Midtrans status | Fraud status | Internal status | Notes |
| --- | --- | --- | --- |
| `capture` | `accept` | `success` | Card payment accepted. |
| `settlement` | any | `success` | Funds settled. |
| `pending` | any | `pending` | Waiting for payment. |
| `deny` | any | `failed` | Denied by Midtrans or issuing bank. |
| `cancel` | any | `failed` | Canceled payment. |
| `expire` | any | `failed` | Payment expired. |
| `failure` | any | `failed` | Payment failed. |

## Data Model Additions

The current `Donation` model is enough for early scaffolding, but Midtrans implementation should add provider fields before production use:

- `Provider`: `midtrans`.
- `ProviderOrderID`: local order ID sent to Midtrans.
- `ProviderTransactionID`: Midtrans transaction ID.
- `ProviderStatus`: raw latest Midtrans transaction status.
- `ProviderPayload`: raw notification JSON or a normalized audit record.
- `PaidAt`: successful payment timestamp.

These fields help with reconciliation and make duplicate notifications safe to process.

## Idempotency Rules

- Use the local donation ID or a deterministic prefixed order ID as `order_id`.
- Reject unknown `order_id` values.
- If a donation is already `success`, repeated success notifications must not increment `Campaign.CollectedAmount` again.
- If a donation is `failed`, later success should be handled deliberately based on Midtrans transaction history, not blindly accepted.
- Store enough raw provider data to investigate mismatches.

## Security Rules

- Verify Midtrans `signature_key`.
- Do not log server keys, client keys, JWTs, or full donor PII.
- Use HTTPS in every deployed environment.
- Keep sandbox and production keys separate.
- Restrict CORS to known frontend domains.

## Future Provider Abstraction

After Midtrans is stable, introduce a payment provider interface around these operations:

- `CreateDonationPayment`
- `VerifyNotification`
- `MapProviderStatus`
- `ExtractProviderReferences`

Xendit can then be implemented as another adapter without changing campaign or donation business logic.
