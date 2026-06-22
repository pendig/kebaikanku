# MVP Scope

This document keeps the first usable MVP small. It is not a full PRD.

## Goal

Let one institution publish donation campaigns and receive Midtrans payments with enough operational visibility to run a controlled pilot.

## Must Ship

- Public campaign listing and campaign detail.
- Manual campaign creation path through seed data, script, or direct admin operation.
- Donation creation endpoint with donor data, amount, and optional platform tip.
- Midtrans Snap payment creation returning a token or redirect URL.
- Midtrans notification callback with signature verification.
- Idempotent donation status updates.
- Campaign collected amount updated exactly once per successful donation.
- Basic donation list/export for operator reconciliation.
- Production configuration notes for PostgreSQL, Midtrans keys, HTTPS, CORS, and secrets.

## Not MVP

- Full institution dashboard.
- Multi-user role management.
- Xendit or any second payment gateway.
- Payment provider abstraction beyond what Midtrans needs.
- AI assistant, AI reports, or WhatsApp automation.
- Donor receipt and zakat certificate generation.
- Advanced audit log UI.
- Automated backup/restore scripts.
- White-label or enterprise packaging.
- Cloudflare Pages Function waitlist hardening.

## Build Order

1. Resolve the current branch divergence before feature work.
2. Implement public campaign read endpoints.
3. Add the smallest campaign creation path for pilot data.
4. Implement donation creation with local `pending` records.
5. Integrate Midtrans Snap payment creation.
6. Implement Midtrans notification callback.
7. Make successful callbacks idempotent.
8. Add a basic operator donation export.
9. Lock production config and deployment notes.

## Acceptance

- A donor can open a public campaign page.
- A donor can start a Midtrans payment.
- A paid Midtrans transaction marks the donation as `success`.
- Repeated Midtrans callbacks do not double-count collected amount.
- An operator can reconcile successful donations without database access.

## Deliberate Shortcut

Campaign management can be manual for the first pilot. Build the dashboard after the payment lifecycle is proven.
