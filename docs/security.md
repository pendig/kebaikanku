# Security Notes

kebaikanku.id handles donation transactions, donor contact data, institution accounts, and future AI-assisted conversations. Security should be treated as product functionality, not an afterthought.

## Data Classes

Sensitive data:
- Donor name, phone number, email, and transaction history.
- Institution profile data.
- Payment references and reconciliation records.
- Admin session cookies; account password hashes once multi-user self-service auth ships.
- Midtrans server keys and notification payloads.
- Future AI chat logs, receipt photos, and distribution reports.

## Authentication

The controlled pilot uses one `ADMIN_PASSWORD` from the backend secret store. A successful login creates a signed, time-limited, `HttpOnly` session cookie; the dashboard never stores the password or a bearer token. Keep `ADMIN_SESSION_SECRET` separate from the password in production and rotate both after suspected exposure.

Multi-user institution accounts remain a later phase. That phase must hash stored passwords with a modern password hashing algorithm and add role checks.

## Payment Security

Midtrans implementation must:
- Verify notification signatures.
- Keep sandbox and production keys separate.
- Treat callbacks as untrusted input until verified.
- Process payment status changes idempotently.
- Avoid logging raw secrets or full donor PII.

See [payment-gateway.md](payment-gateway.md).

## CORS

Allowed origins should be specific in production:

- `https://kebaikanku.id`
- `https://app.kebaikanku.id`
- Any managed Cloudflare Pages preview domains intentionally used by the team.

Avoid broad wildcard origins for production deployments.

## Rate Limiting

The API currently rate-limits public campaign, donation, and waitlist routes. Add separate limits for:

- Login attempts.
- Payment callback endpoint.

Payment callbacks should be protected primarily through provider signature verification, not only IP allowlists.

## AI Data Handling

Future AI features should follow these rules:

- Do not send passwords, bank credentials, or full payment secrets to LLM providers.
- Minimize donor PII in prompts.
- Record whether generated reports were reviewed by a human before publication.
- Provide a way to disable AI processing for self-hosted deployments that require stricter data control.

## Logging

Logs should include:
- Request ID.
- Route.
- Status code.
- Latency.
- Internal object IDs when needed.

Logs should not include:
- Passwords.
- Session cookies.
- Midtrans server keys.
- Full Authorization headers.
- Full donor contact details unless explicitly needed in secure audit logs.

## Pre-Launch Security Checklist

- `ADMIN_PASSWORD` and `ADMIN_SESSION_SECRET` configured outside the repo.
- Admin cookies are `HttpOnly`, `Secure` in production, and time-limited.
- Midtrans signature verification implemented and tested.
- CORS locked to production domains.
- Rate limiting added for auth and donation routes.
- Error responses do not leak stack traces or raw database errors.
- Production uses HTTPS.
- Database backups are encrypted or access-controlled.
- Privacy policy matches actual AI/payment/data behavior.
