# Roadmap

This roadmap is ordered by implementation dependency, not by marketing priority.

## Phase 0: Foundation

Status: complete for the controlled-pilot baseline.

- Repository licensing and contribution docs.
- Go backend skeleton.
- GORM database setup.
- Core domain models.
- Static landing page.
- Architecture, database, localization, and handover docs.

## Phase 1: Developer Readiness

Status: complete.

- Backend README.
- Landing README.
- API contract.
- Deployment guide.
- Security notes.
- Payment gateway design.
- Environment examples.
- Local PostgreSQL compose file.

Exit state:
- A new developer can run the backend and landing page locally.
- A new developer understands the planned API and Midtrans integration before coding.

## Phase 2: Backend API Core

Status: complete for the single-organization pilot; multi-institution self-service remains later work.

- Add repository layer for organizations, campaigns, donors, and donations.
- Add password login and signed session-cookie middleware.
- Implement public campaign reads and authenticated campaign management.

Exit state:
- The pilot administrator can log in and create campaigns for the bootstrapped organization.
- Public users can list active campaigns.

## Phase 3: Donation and Midtrans MVP

Status: complete.

- Add Midtrans config and client.
- Implement `POST /api/v1/donations`.
- Create local pending donations before calling Midtrans.
- Return Midtrans Snap token or redirect URL.
- Implement `POST /api/v1/payments/midtrans/notification`.
- Verify Midtrans notification signatures.
- Update donation and campaign totals idempotently.

Exit state:
- A donor can start a Midtrans payment.
- Midtrans callback can mark a donation as successful or failed.
- Campaign collected amount updates exactly once per successful donation.

## Phase 4: Dashboard SPA

Status: complete for the single-admin pilot.

- Scaffold `frontend/dashboard`.
- Add auth screens.
- Add campaign list and campaign creation.
- Add donation/payment status views.
- Add API client and signed session-cookie handling.
- Deploy as Cloudflare Pages SPA.

Exit state:
- The pilot administrator can operate the MVP without direct API calls.

## Phase 5: Reporting and Operations

Status: partially complete. Production migrations, CI, backup/restore scripts, donation CSV export, and deployment documentation are available; reports and audit logs remain.

- Add campaign report model.
- Add donor receipt and zakat certificate generation.
- Add basic admin audit logs.

Exit state:
- The platform is ready for a controlled single-organization pilot.

## Phase 6: AI Roadmap

- Add AI gateway service boundary.
- Add WhatsApp CS and zakat calculator workflow.
- Add AI report draft generation from receipt photos and notes.
- Add human review before donor broadcast.

Exit state:
- AI features support staff efficiency without bypassing human accountability.
