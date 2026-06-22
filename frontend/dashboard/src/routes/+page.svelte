<script>
	import { onMount } from 'svelte';

	const apiBase = (import.meta.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');

	let token = $state('');
	let campaigns = $state([]);
	let donations = $state([]);
	let message = $state('');
	let error = $state('');
	let loading = $state(false);
	let campaign = $state({
		organization_id: 'pilot-org',
		title: '',
		slug: '',
		description: '',
		category: 'infak',
		target_amount: 10000000,
		end_date: '2026-12-31T23:59'
	});

	onMount(() => {
		token = sessionStorage.getItem('campaign_admin_token') || '';
		loadCampaigns();
		if (token) loadDonations();
	});

	function saveToken() {
		token = token.trim();
		if (token) {
			sessionStorage.setItem('campaign_admin_token', token);
			loadDonations();
		} else {
			sessionStorage.removeItem('campaign_admin_token');
			donations = [];
		}
		message = 'Token saved.';
		error = '';
	}

	async function api(path, options = {}) {
		const headers = {
			...(options.headers || {})
		};
		if (token) headers.Authorization = `Bearer ${token}`;
		const response = await fetch(`${apiBase}${path}`, { ...options, headers });
		return response;
	}

	async function loadCampaigns() {
		loading = true;
		error = '';
		try {
			const response = await api('/api/v1/campaigns?limit=100');
			const payload = await response.json();
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Could not load campaigns.');
			campaigns = payload?.data?.campaigns || [];
		} catch (err) {
			error = err.message || 'Could not load campaigns.';
		} finally {
			loading = false;
		}
	}

	async function createCampaign() {
		loading = true;
		error = '';
		message = '';
		try {
			const response = await api('/api/v1/campaigns', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					...campaign,
					target_amount: Number(campaign.target_amount),
					end_date: new Date(campaign.end_date).toISOString()
				})
			});
			const payload = await response.json();
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Could not create campaign.');
			message = 'Campaign created.';
			campaign.title = '';
			campaign.slug = '';
			campaign.description = '';
			await loadCampaigns();
		} catch (err) {
			error = err.message || 'Could not create campaign.';
		} finally {
			loading = false;
		}
	}

	async function loadDonations() {
		error = '';
		try {
			const response = await api('/api/v1/donations/export');
			if (!response.ok) {
				let errMsg = 'Could not load donations.';
				const body = await response.text();
				try {
					const payload = JSON.parse(body);
					errMsg = payload?.error?.message || errMsg;
				} catch {
					errMsg = body || errMsg;
				}
				throw new Error(errMsg);
			}
			const csv = await response.text();
			donations = parseCSV(csv);
		} catch (err) {
			error = err.message || 'Could not load donations.';
		}
	}

	function parseCSV(csv) {
		const rows = [];
		let row = [];
		let cell = '';
		let quoted = false;
		for (let i = 0; i < csv.length; i += 1) {
			const char = csv[i];
			const next = csv[i + 1];
			if (char === '"' && quoted && next === '"') {
				cell += '"';
				i += 1;
			} else if (char === '"') {
				quoted = !quoted;
			} else if (char === ',' && !quoted) {
				row.push(cell);
				cell = '';
			} else if ((char === '\n' || char === '\r') && !quoted) {
				if (char === '\r' && next === '\n') i += 1;
				row.push(cell);
				if (row.some(Boolean)) rows.push(row);
				row = [];
				cell = '';
			} else {
				cell += char;
			}
		}
		row.push(cell);
		if (row.some(Boolean)) rows.push(row);
		const [headers = [], ...records] = rows;
		return records.map((values) => Object.fromEntries(headers.map((key, index) => [key, values[index] || ''])));
	}

	function slugify() {
		campaign.slug = campaign.title
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-|-$/g, '');
	}

	function formatRupiah(amount) {
		return new Intl.NumberFormat('id-ID', {
			style: 'currency',
			currency: 'IDR',
			maximumFractionDigits: 0
		}).format(Number(amount || 0));
	}
</script>

<svelte:head>
	<title>Dashboard | kebaikanku.id</title>
</svelte:head>

<main class="min-h-screen bg-slate-100 p-5 text-slate-950">
	<div class="mx-auto max-w-7xl space-y-5">
		<header class="flex flex-col gap-3 border-b border-slate-300 pb-5 md:flex-row md:items-end md:justify-between">
			<div>
				<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Alpha dashboard</p>
				<h1 class="text-3xl font-black tracking-tight">kebaikanku.id</h1>
				<p class="text-sm text-slate-600">Campaign operations with admin token. Full signup/login later.</p>
			</div>
			<div class="flex gap-2">
				<input bind:value={token} type="password" autocomplete="off" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm md:w-80" placeholder="CAMPAIGN_ADMIN_TOKEN" />
				<button onclick={saveToken} class="rounded-lg bg-slate-950 px-4 py-2 text-sm font-bold text-white">Save</button>
			</div>
		</header>

		{#if message}
			<p class="rounded-lg bg-emerald-100 p-3 text-sm font-semibold text-emerald-800">{message}</p>
		{/if}
		{#if error}
			<p class="rounded-lg bg-red-100 p-3 text-sm font-semibold text-red-800">{error}</p>
		{/if}

		<section class="grid gap-5 lg:grid-cols-[420px_1fr]">
			<form class="rounded-xl border border-slate-300 bg-white p-5 shadow-sm" onsubmit={(event) => { event.preventDefault(); createCampaign(); }}>
				<h2 class="mb-4 text-lg font-black">Create campaign</h2>
				<label>Title<input bind:value={campaign.title} oninput={slugify} required /></label>
				<label>Slug<input bind:value={campaign.slug} required pattern="[a-z0-9-]+" /></label>
				<label>Description<textarea bind:value={campaign.description} required></textarea></label>
				<label>Category
					<select bind:value={campaign.category}>
						<option value="infak">Infak</option>
						<option value="kemanusiaan">Kemanusiaan</option>
						<option value="zakat_maal">Zakat Maal</option>
						<option value="zakat_fitrah">Zakat Fitrah</option>
					</select>
				</label>
				<div class="grid grid-cols-2 gap-3">
					<label>Target<input bind:value={campaign.target_amount} min="1" step="1" type="number" required /></label>
					<label>End date<input bind:value={campaign.end_date} type="datetime-local" required /></label>
				</div>
				<button disabled={loading || !token} class="mt-4 w-full rounded-lg bg-emerald-600 px-4 py-3 text-sm font-black text-white disabled:opacity-50">Create</button>
			</form>

			<div class="space-y-5">
				<section class="rounded-xl border border-slate-300 bg-white p-5 shadow-sm">
					<div class="mb-4 flex items-center justify-between">
						<h2 class="text-lg font-black">Campaigns</h2>
						<button onclick={loadCampaigns} class="rounded-lg border border-slate-300 px-3 py-2 text-xs font-bold">Refresh</button>
					</div>
					<div class="grid gap-3 md:grid-cols-2">
						{#each campaigns as item (item.id)}
							<article class="rounded-lg border border-slate-200 p-4">
								<p class="text-xs font-bold uppercase text-emerald-700">{item.category}</p>
								<h3 class="font-black">{item.title}</h3>
								<p class="mt-1 line-clamp-2 text-sm text-slate-600">{item.description}</p>
								<p class="mt-3 text-sm font-bold">{formatRupiah(item.collected_amount)} / {formatRupiah(item.target_amount)}</p>
							</article>
						{/each}
					</div>
				</section>

				<section class="rounded-xl border border-slate-300 bg-white p-5 shadow-sm">
					<div class="mb-4 flex items-center justify-between">
						<h2 class="text-lg font-black">Donation status</h2>
						<button onclick={loadDonations} disabled={!token} class="rounded-lg border border-slate-300 px-3 py-2 text-xs font-bold disabled:opacity-50">Load CSV</button>
					</div>
					<div class="overflow-x-auto">
						<table class="w-full text-left text-sm">
							<thead>
								<tr class="border-b border-slate-200 text-xs uppercase text-slate-500">
									<th>Campaign</th><th>Donor</th><th>Amount</th><th>Status</th><th>Paid</th>
								</tr>
							</thead>
							<tbody>
								{#each donations as donation}
									<tr class="border-b border-slate-100">
										<td>{donation.campaign}</td>
										<td>{donation.donor_name}</td>
										<td>{formatRupiah(donation.amount)}</td>
										<td>{donation.status}</td>
										<td>{donation.paid_at}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</section>
			</div>
		</section>
	</div>
</main>

<style>
	label {
		display: block;
		margin-top: 0.75rem;
		font-size: 0.75rem;
		font-weight: 800;
		color: #475569;
	}
	input, textarea, select {
		margin-top: 0.25rem;
		width: 100%;
		border: 1px solid #cbd5e1;
		border-radius: 0.5rem;
		padding: 0.7rem 0.75rem;
		font-size: 0.875rem;
		outline: none;
	}
	textarea {
		min-height: 6rem;
	}
	main {
		min-height: 100vh;
		background: #f1f5f9;
		padding: 1.25rem;
		color: #020617;
	}
	main > div {
		max-width: 80rem;
		margin: 0 auto;
	}
	header {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		gap: 1rem;
		border-bottom: 1px solid #cbd5e1;
		padding-bottom: 1.25rem;
	}
	header p {
		color: #475569;
	}
	form,
	section section,
	article {
		border: 1px solid #cbd5e1;
		border-radius: 0.85rem;
		background: white;
		padding: 1.25rem;
	}
	button {
		border: 0;
		border-radius: 0.6rem;
		background: #0f172a;
		color: white;
		padding: 0.7rem 1rem;
		font-weight: 800;
		cursor: pointer;
	}
	button:disabled {
		cursor: not-allowed;
		opacity: 0.55;
	}
	th, td {
		padding: 0.7rem 0.5rem;
		white-space: nowrap;
	}
	@media (min-width: 900px) {
		main > div > section {
			display: grid;
			grid-template-columns: 420px 1fr;
			gap: 1.25rem;
		}
	}
</style>
