# Landing Page

Static SvelteKit landing page for kebaikanku.id.

## Status

Implemented:
- Homepage.
- Terms page at `/terms`.
- Privacy page at `/privacy`.
- Indonesian and English localization.
- Static prerendering through `@sveltejs/adapter-static`.
- TailwindCSS v4 through the Vite plugin.

## Requirements

- Node.js 18 or newer.
- npm.

## Local Development

1. Install dependencies:
   ```bash
   npm install
   ```

2. Copy the environment variables example:
   ```bash
   cp .env.example .env
   ```

3. Start the local Vite server:
   ```bash
   npm run dev
   ```

Vite starts on `http://localhost:8383` (strictPort) by default.


## Static Build

```bash
npm run build
```

The static output is generated in `build/`.

Preview the production build:

```bash
npm run preview
```

## Localization

Translations live in:

- `src/lib/locales/id.json`
- `src/lib/locales/en.json`

The translation helper lives in `src/lib/i18n.js`. Indonesian (`id`) is the default locale. The selected locale is stored in browser `localStorage` as `lang`.

When adding visible user-facing text, add keys to both locale files and use the `t` store from `$lib/i18n`.

## Deployment

Recommended target: Cloudflare Pages.

Cloudflare Pages settings:

| Setting | Value |
| --- | --- |
| Root directory | `frontend/landing` |
| Build command | `npm run build` |
| Build output directory | `build` |
| Node version | `18` or newer |

The current routes are prerendered static pages, so the output can be served by any static host.

Campaign detail metadata is rendered at the Cloudflare edge by `functions/campaigns/[slug].js`,
so social crawlers receive campaign-specific canonical, Open Graph, Twitter, and JSON-LD tags
without running JavaScript. Set `API_BASE_URL` to the public backend origin and
`PUBLIC_SEO_SITE_URL` to the public landing origin in Cloudflare Pages. If the API is unavailable,
the function safely serves the normal static fallback page.

## Related Docs

- [../../docs/deployment.md](../../docs/deployment.md)
- [../../docs/localization.md](../../docs/localization.md)
