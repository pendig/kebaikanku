# Landing Page Deployment & Static Hosting Checklist

This document is the operational checklist for deploying and validating the **kebaikanku.id** landing page as a static site (Phase 1, Issue #29).

---

## ☁️ Target Hosting Platform: Cloudflare Pages

We recommend deploying the static landing page to **Cloudflare Pages** because of its native global CDN, zero-cost static tier, automatic TLS certificates, and seamless edge routing.

### Project Configurations
- **Framework Preset:** `SvelteKit` (or `None` with static outputs)
- **Root Directory:** `frontend/landing`
- **Build Command:** `npm run build`
- **Build Output Directory:** `build`
- **Node.js Version:** `18` or newer (Set `NODE_VERSION` environment variable to `20` in Cloudflare console)

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

1. **Cloudflare Pages Dashboard:** Navigate to **kebaikanku-landing > Deployments**.
2. **Select Previous Stable Build:** Find the last successful deployment from the history.
3. **Rollback:** Click the three dots icon next to the build and select **Rollback to this deployment**. This instantly updates the CDN routing at the edge in less than 5 seconds.

---

## 🌐 Future API Integration Strategy (Same Domain Strategy)

To integrate the Go backend REST API (`/api/v1/*`) behind the same domain (`kebaikanku.id`) without triggering CORS issues:

1. **Cloudflare Workers / Rules:** Set up a Cloudflare routing rule or worker on the same domain:
   - Paths matching `/api/v1/*` -> Forward traffic to the Go backend API server (e.g. deployed on a VPS or cloud instance).
   - All other paths (`/*`) -> Serve from Cloudflare Pages static hosting.
2. **Benefits:**
   - 0% CORS overhead (since cookies and tokens are served on the same top-level domain).
   - Faster response times due to Cloudflare edge routing caching.
