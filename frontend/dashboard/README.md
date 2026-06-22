# Dashboard

Alpha operator dashboard for campaign setup and donation reconciliation.

This dashboard uses `CAMPAIGN_ADMIN_TOKEN` for now. Full register/login is intentionally postponed.

## Local Development

```bash
npm install
PUBLIC_API_BASE_URL=http://localhost:8080 npm run dev
```

Open `http://localhost:8484`.

## Build

```bash
npm run build
```

Output is written to `build/` and can be served by any static host.
