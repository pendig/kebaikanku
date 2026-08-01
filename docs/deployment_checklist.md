# Landing Page Deployment & Static Hosting Checklist

This document is the operational checklist for deploying and validating the **kebaikanku.id** landing page as a static site (Phase 1, Issue #29).

---

## ☁️ Target Hosting Platform: Cloudflare Workers Static Assets

The landing repository includes `wrangler.jsonc` for Workers Static Assets. The generated `build` directory can also be hosted on Pages, but one project should use only one deployment model.

### Project Configurations
- **Root Directory:** `frontend/landing`
- **Build Command:** `npm ci && npm run build`
- **Deploy Command:** `npx wrangler deploy`
- **Static Assets Directory:** `build`

---

## 🛠️ Build and Pre-Deployment Validation

Before pushing changes to the release branch, execute the following commands locally:

### 1. Install Dependencies
```bash
cd frontend/landing
npm ci
```

### 2. Run Local Build Check
Ensure that the SvelteKit static adapter completes the prerendering without errors:
```bash
npm run build
```
*Expected Output:* SvelteKit generates a `build/` directory containing all pre-rendered HTML pages (`index.html`, `coming-soon/index.html`, `campaigns/index.html`, etc.) and CSS/JS assets.

### 3. Verify Prerendered Pages Locally
Run the Vite preview server to check routing and visual states locally before pushing:
```bash
npm run preview
```
Visit `http://localhost:4173` and click through all routes (`/`, `/campaigns`, `/coming-soon`, `/privacy`, `/terms`) to verify that the **Liquid Glass** animations, fonts, and dark mode display correctly.

---

## 🔄 Rollback Procedures

If the live landing page experiences styling anomalies or broken routing post-deployment:

1. **Cloudflare Dashboard:** Navigate to **Workers & Pages > kebaikanku-landing > Deployments**.
2. **Select Previous Stable Build:** Find the last successful deployment from the history.
3. **Rollback:** Select the last known-good deployment and use the available rollback action.

---

## 🌐 Future API Integration Strategy (Same Domain Strategy)

To integrate the Go backend REST API (`/api/v1/*`) behind the same domain (`kebaikanku.id`) without triggering CORS issues:

1. **Cloudflare Workers / Rules:** Set up a Cloudflare routing rule or worker on the same domain:
   - Paths matching `/api/v1/*` -> Forward traffic to the Go backend API server (e.g. deployed on a VPS or cloud instance).
   - All other paths (`/*`) -> Serve from Cloudflare Pages static hosting.
2. **Benefits:**
   - 0% CORS overhead (since cookies and tokens are served on the same top-level domain).
   - Faster response times due to Cloudflare edge routing caching.
