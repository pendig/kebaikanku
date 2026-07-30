const FETCH_TIMEOUT_MS = 3000;

export async function onRequestGet(context) {
	const assetURL = new URL('/404.html', context.request.url);
	const asset = context.env.ASSETS
		? await context.env.ASSETS.fetch(assetURL)
		: await context.next();
	const html = await asset.text();
	const apiBase = String(context.env.API_BASE_URL || context.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');

	if (!apiBase) return htmlResponse(html, asset);

	try {
		const response = await fetch(`${apiBase}/api/v1/campaigns/${encodeURIComponent(context.params.slug)}`, {
			headers: { Accept: 'application/json' },
			signal: AbortSignal.timeout(FETCH_TIMEOUT_MS)
		});
		const payload = await response.json();
		if (!response.ok || !payload?.success || !payload?.data?.campaign) return htmlResponse(html, asset);

		const siteBase = String(context.env.PUBLIC_SEO_SITE_URL || new URL(context.request.url).origin).replace(/\/$/, '');
		return htmlResponse(injectCampaignMetadata(html, payload.data.campaign, siteBase, apiBase), asset);
	} catch {
		return htmlResponse(html, asset);
	}
}

export function injectCampaignMetadata(html, campaign, siteBase, apiBase) {
	const canonical = `${siteBase}/campaigns/${encodeURIComponent(campaign.slug)}`;
	const title = `${plainText(campaign.title, 70)} | kebaikanku.id`;
	const description = plainText(campaign.description || `Bantu ${campaign.title} melalui kebaikanku.id`, 160);
	const image = absoluteURL(campaign.banner_url || '/images/og-image.png', campaign.banner_url ? apiBase : siteBase);
	const structuredData = JSON.stringify({
		'@context': 'https://schema.org',
		'@type': 'DonateAction',
		name: campaign.title,
		description,
		image,
		target: { '@type': 'EntryPoint', urlTemplate: canonical }
	}).replace(/</g, '\\u003c');
	const tags = [
		`<title>${escapeHTML(title)}</title>`,
		`<meta name="description" content="${escapeHTML(description)}">`,
		'<meta name="robots" content="index, follow, max-image-preview:large">',
		`<link rel="canonical" href="${escapeHTML(canonical)}">`,
		'<meta property="og:type" content="article">',
		'<meta property="og:site_name" content="kebaikanku.id">',
		`<meta property="og:url" content="${escapeHTML(canonical)}">`,
		`<meta property="og:title" content="${escapeHTML(title)}">`,
		`<meta property="og:description" content="${escapeHTML(description)}">`,
		`<meta property="og:image" content="${escapeHTML(image)}">`,
		`<meta property="og:image:alt" content="${escapeHTML(`Banner ${campaign.title}`)}">`,
		'<meta name="twitter:card" content="summary_large_image">',
		`<meta name="twitter:title" content="${escapeHTML(title)}">`,
		`<meta name="twitter:description" content="${escapeHTML(description)}">`,
		`<meta name="twitter:image" content="${escapeHTML(image)}">`,
		`<script type="application/ld+json">${structuredData}</script>`
	].join('\n');

	const withoutDefaults = html
		.replace(/<title>[\s\S]*?<\/title>/gi, '')
		.replace(/<meta\s+(?:name|property)=["'](?:description|robots|og:[^"']+|twitter:[^"']+)["'][^>]*>/gi, '')
		.replace(/<link\s+rel=["']canonical["'][^>]*>/gi, '');
	return withoutDefaults.replace('</head>', `${tags}\n</head>`);
}

function plainText(value, limit) {
	return String(value || '').replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim().slice(0, limit);
}

function escapeHTML(value) {
	return String(value).replace(/[&<>"']/g, (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[character]);
}

function absoluteURL(value, base) {
	try { return new URL(value, `${base}/`).href; } catch { return value; }
}

function htmlResponse(html, asset) {
	const headers = new Headers(asset.headers);
	headers.set('Content-Type', 'text/html; charset=utf-8');
	headers.set('Cache-Control', 'public, max-age=60, s-maxage=300');
	return new Response(html, { status: asset.ok ? 200 : asset.status, headers });
}
