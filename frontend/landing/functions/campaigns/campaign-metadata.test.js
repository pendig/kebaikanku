import assert from 'node:assert/strict';
import test from 'node:test';
import { injectCampaignMetadata } from './[slug].js';

test('injects escaped campaign metadata without leaking default tags', () => {
	const html = '<html><head><title>Default</title><meta name="description" content="default"><meta property="og:title" content="default"><link rel="canonical" href="https://old.test"></head><body></body></html>';
	const result = injectCampaignMetadata(html, {
		slug: 'bantu-anak',
		title: 'Bantu & Anak',
		description: 'Aman "dan" terpercaya<script>alert(1)</script>',
		banner_url: '/uploads/banner.jpg'
	}, 'https://kebaikanku.id', 'https://api.kebaikanku.id');

	assert.match(result, /<title>Bantu &amp; Anak \| kebaikanku\.id<\/title>/);
	assert.match(result, /https:\/\/kebaikanku\.id\/campaigns\/bantu-anak/);
	assert.match(result, /https:\/\/api\.kebaikanku\.id\/uploads\/banner\.jpg/);
	assert.doesNotMatch(result, /<script>alert\(1\)<\/script>/);
	assert.equal((result.match(/name="description"/g) || []).length, 1);
	assert.equal((result.match(/property="og:title"/g) || []).length, 1);
});
