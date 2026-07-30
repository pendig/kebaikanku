# Dashboard

Single-admin dashboard for campaign setup and donation reconciliation. Authentication uses the backend's secure session cookie; the password is configured only in the backend environment.

## Local Development

```bash
npm install
PUBLIC_API_BASE_URL=http://localhost:8080 PUBLIC_LANDING_BASE_URL=http://localhost:8383 npm run dev
```

Open `http://localhost:8484`.

Configure the backend with `ADMIN_PASSWORD`. When frontend and API use different origins, both must use HTTPS in production and the API CORS allowlist must include the dashboard origin.

## Build

```bash
npm run build
```

Output is written to `build/` and can be served by any static host.
