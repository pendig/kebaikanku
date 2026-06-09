# Contributing to kebaikanku.id

Thank you for your interest in contributing to **kebaikanku.id**! We welcome contributions from developers, designers, writers, and translators to make this open-source crowdfunding platform better for everyone.

---

## 🛡️ License Agreement & CLA
By contributing to this repository, you agree that:
1.  Your contributions will be licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.
2.  You grant the project maintainers a non-exclusive, perpetual, worldwide, royalty-free license to use, modify, and distribute your contributions under the **Commercial Enterprise License** (Dual-Licensing Model) to support the sustainability of the platform.

---

## 🚀 How to Contribute

### 1. Reporting Bugs & Suggesting Features
*   Check the issue tracker to see if your bug or feature request has already been reported.
*   Open a new issue with a clear description, steps to reproduce (for bugs), and expected behavior.

### 2. Developing & Pull Requests (PRs)
*   Fork the repository and create a new branch from `main`: `git checkout -b feature/your-feature-name`.
*   Ensure your code matches the project's formatting and styling guidelines.
*   Keep PRs focused and single-purpose. Large PRs are harder to review and merge.
*   Link the PR to the relevant issue.

---

## 🧭 Developer Workflow

### Backend
```bash
cd backend
cp .env.example .env
go run cmd/api/main.go
```

Before opening a backend PR, run:

```bash
cd backend
go fmt ./...
go vet ./...
go test ./...
```

### Landing Page
```bash
cd frontend/landing
npm install
npm run dev
```

Before opening a landing page PR, run:

```bash
cd frontend/landing
npm run build
```

### Local PostgreSQL
```bash
docker compose up -d postgres
```

Use PostgreSQL locally when working on production-like database behavior, payment reconciliation, or migration-sensitive changes.

---

## 💻 Code Style Guidelines

### Go Backend (Clean Architecture)
*   **Hexagonal Architecture**: Keep business logic (domain) isolated from infrastructure (handlers, database repositories, external integrations).
*   **ORM**: Use **GORM** for relational database access. Make sure repository queries are database-agnostic so they work on both SQLite (development) and PostgreSQL (production).
*   **Formatting**: Always run `go fmt ./...` and `go vet ./...` before committing.
*   **Naming Conventions**: Follow standard Go naming conventions (camelCase for variables, PascalCase for exported identifiers).

### SvelteKit Frontend
*   **CSS Style**: Use **TailwindCSS** for styles. Follow a mobile-first responsive approach.
*   **State Management**: Use Svelte stores or Svelte 5 runes (depending on Svelte version) for state sharing.
*   **Accessibility (a11y)**: Write semantic HTML. Ensure all buttons have descriptive labels and screen reader support.
*   **Formatting**: Run `npm run format` (using Prettier) before submitting.

### Localization & Multi-language (i18n)
*   We support multi-language translation. The priority is **Indonesian (id)**, but all UI strings must be localized using JSON files so they can be easily translated to **English (en)** and other languages.
*   Do not hardcode user-facing strings directly in Svelte files. Use the i18n translation keys instead.

---

## 💳 Payment Gateway Contributions

The first supported payment gateway is **Midtrans**. Payment-related PRs should preserve these rules:

*   Verify Midtrans notification signatures before updating donation state.
*   Treat callbacks as idempotent.
*   Keep campaign collected amount updates inside the same database transaction as donation status updates.
*   Do not log Midtrans server keys, JWTs, or full donor PII.

See `docs/payment-gateway.md` and `docs/security.md`.

---

## 💬 Community & Discussion
If you have any questions or want to discuss features, feel free to open a Discussion on GitHub or reach out to the project maintainers.
