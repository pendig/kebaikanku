<script>
	import { onMount } from 'svelte';
	import { page } from '$app/state';

	const apiBase = (import.meta.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');
	const landingBase = (import.meta.env.PUBLIC_LANDING_BASE_URL || 'http://127.0.0.1:18383').replace(/\/$/, '');
	const blankCampaign = () => ({
		id: '', organization_id: 'pilot-org', title: '', slug: '', description: '', category: 'infaq', subcategory: '', campaign_type: 'target_deadline', banner_url: '', location: '', beneficiary_note: '', target_amount: 10000000, end_date: '2026-12-31T23:59'
	});

	let { initialTab = 'campaigns' } = $props();
	let token = $state('');
	let tab = $state('campaigns');
	let campaigns = $state([]);
	let donations = $state([]);
	let message = $state('');
	let error = $state('');
	let loading = $state(false);
	let campaign = $state(blankCampaign());

	onMount(() => {
		tab = initialTab;
		token = sessionStorage.getItem('campaign_admin_token') || '';
		loadCampaigns();
		if (token) loadDonations();
	});


	function saveToken() {
		token = token.trim();
		if (token) sessionStorage.setItem('campaign_admin_token', token);
		else sessionStorage.removeItem('campaign_admin_token');
		message = 'Token saved.';
		if (token) loadDonations();
	}

	async function api(path, options = {}) {
		const headers = { ...(options.headers || {}) };
		if (token) headers.Authorization = `Bearer ${token}`;
		return fetch(`${apiBase}${path}`, { ...options, headers });
	}

	async function loadCampaigns() {
		loading = true; error = '';
		try {
			const response = await api('/api/v1/campaigns?limit=100');
			const payload = await response.json();
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Could not load campaigns.');
			campaigns = payload?.data?.campaigns || [];
			const editId = page.url.searchParams.get('id');
			if (initialTab === 'campaign_form' && editId) {
				const existing = campaigns.find((item) => item.id === editId);
				if (existing) fillCampaign(existing);
			}
		} catch (err) { error = err.message || 'Could not load campaigns.'; }
		finally { loading = false; }
	}

	async function saveCampaign() {
		loading = true; error = ''; message = '';
		try {
			const body = JSON.stringify({ ...campaign, target_amount: Number(campaign.target_amount), end_date: new Date(campaign.end_date).toISOString() });
			const response = await api(campaign.id ? `/api/v1/campaigns/${campaign.id}` : '/api/v1/campaigns', { method: campaign.id ? 'PUT' : 'POST', headers: { 'Content-Type': 'application/json' }, body });
			const payload = await response.json();
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Could not save campaign.');
			message = campaign.id ? 'Campaign updated.' : 'Campaign created.';
			campaign = blankCampaign();
			await loadCampaigns();
		} catch (err) { error = err.message || 'Could not save campaign.'; }
		finally { loading = false; }
	}

	async function loadDonations() {
		error = '';
		try {
			const response = await api('/api/v1/donations/export');
			if (!response.ok) throw new Error((await response.text()) || 'Could not load donations.');
			donations = parseCSV(await response.text());
		} catch (err) { error = err.message || 'Could not load donations.'; }
	}

	function fillCampaign(item) {
		campaign = { ...blankCampaign(), ...item, end_date: item.end_date ? new Date(item.end_date).toISOString().slice(0, 16) : '2026-12-31T23:59' };
		message = 'Editing campaign.';
	}

	function slugify() {
		campaign.slug = campaign.title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
	}

	function parseCSV(csv) {
		const [head, ...rows] = csv.trim().split(/\r?\n/);
		if (!head) return [];
		const headers = head.split(',');
		return rows.map((row) => Object.fromEntries(headers.map((key, i) => [key, row.split(',')[i] || ''])));
	}

	function formatRupiah(amount) {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(amount || 0));
	}

	function progress(item) {
		if (!item?.target_amount || item.campaign_type === 'open_amount') return 0;
		return Math.min(Math.round(((item.collected_amount || 0) / item.target_amount) * 100), 100);
	}
</script>

<svelte:head><title>Dashboard | kebaikanku.id</title></svelte:head>

<main>
	<div>
		<header>
			<div>
				<p>Alpha dashboard</p>
				<h1>kebaikanku.id</h1>
				<p>Campaign builder dengan preview publik. Signup/login nanti.</p>
			</div>
			<div><input bind:value={token} type="password" autocomplete="off" placeholder="CAMPAIGN_ADMIN_TOKEN" /><button onclick={saveToken}>Save</button></div>
		</header>

		<nav class="topnav">
			<a class:active={tab === 'campaigns'} href="/campaigns">Campaigns</a>
			<a class:active={tab === 'donations'} href="/donations">Donations</a>
			<a class:active={tab === 'settings'} href="/settings">Settings</a>
		</nav>

		{#if message}<p class="notice ok">{message}</p>{/if}
		{#if error}<p class="notice bad">{error}</p>{/if}

		{#if tab === 'campaigns'}
			<section>
				<div class="section-head"><h2>Campaigns</h2><a class="button" href="/campaign/form">Create campaign</a></div>
				<div class="campaign-cards">
					{#each campaigns as item (item.id)}
						<article>
							{#if item.banner_url}<img src={item.banner_url} alt="Banner {item.title}" />{/if}
							<div class="card-body">
								<div class="card-title"><h3>{item.title}</h3><span>{item.category}{item.subcategory ? ` / ${item.subcategory}` : ''}</span></div>
								{#if item.location}<p class="loc">{item.location}</p>{/if}
								<p class="desc">{item.description}</p>
								{#if item.beneficiary_note}<p class="benefit">{item.beneficiary_note}</p>{/if}
								<div class="bar"><i style="width: {item.campaign_type === 'open_amount' ? 100 : progress(item)}%"></i></div>
								<div class="amounts"><span>{formatRupiah(item.collected_amount)} terkumpul</span><span>{item.campaign_type === 'open_amount' ? 'Unlimited' : formatRupiah(item.target_amount)}</span></div>
								<div class="actions"><a class="button ghost" href="/campaign/form?id={item.id}">Edit</a><a class="button ghost" href="{landingBase}/campaigns/{item.slug}" target="_blank">Preview</a></div>
							</div>
						</article>
					{/each}
				</div>
			</section>
		{:else if tab === 'campaign_form'}
			<section class="builder">
				<form onsubmit={(event) => { event.preventDefault(); saveCampaign(); }}>
					<h2>{campaign.id ? 'Edit campaign' : 'Create campaign'}</h2>
					<label>Title<input bind:value={campaign.title} oninput={slugify} required /></label>
					<label>Slug<input bind:value={campaign.slug} required pattern="[a-z0-9-]+" /></label>
					<label>Description<textarea bind:value={campaign.description} required></textarea></label>
					<label>Banner URL<input bind:value={campaign.banner_url} placeholder="https://..." /></label>
					<div class="two"><label>Category<select bind:value={campaign.category}><option value="zakat">Zakat</option><option value="infaq">Infaq</option><option value="qurban">Qurban</option><option value="donasi_lainnya">Donasi Lainnya</option></select></label><label>Sub kategori<input bind:value={campaign.subcategory} placeholder="Pendidikan, Palestina, Masjid..." /></label></div>
					<label>Campaign type<select bind:value={campaign.campaign_type}><option value="target_deadline">Target + batas waktu</option><option value="target_open">Target tanpa batas waktu</option><option value="open_amount">Unlimited amount campaign</option></select></label>
					<div class="two"><label>Target<input bind:value={campaign.target_amount} min="1" step="1" type="number" required /></label><label>End date<input bind:value={campaign.end_date} type="datetime-local" required /></label></div>
					<label>Location<input bind:value={campaign.location} placeholder="Riau, Indonesia" /></label>
					<label>Beneficiary note<textarea bind:value={campaign.beneficiary_note} placeholder="Contoh: 120 santri dan warga sekitar"></textarea></label>
					<div class="actions"><button disabled={loading || !token}>{campaign.id ? 'Update' : 'Create'}</button>{#if campaign.id}<a class="button ghost" href="/campaign/form">Cancel edit</a>{/if}</div>
				</form>

				<aside class="preview">
					<h2>Preview public page</h2>
					<article>
						{#if campaign.banner_url}<img src={campaign.banner_url} alt="Banner preview" />{/if}
						<p class="badge">{campaign.category}{campaign.subcategory ? ` / ${campaign.subcategory}` : ''}</p>
						<h3>{campaign.title || 'Judul campaign'}</h3>
						<p>{campaign.location || 'Lokasi campaign'}</p>
						<div class="meter"><strong>{formatRupiah(0)} terkumpul</strong><span>{campaign.campaign_type === 'open_amount' ? 'unlimited' : `dari ${formatRupiah(campaign.target_amount)}`}</span><i style="width: {progress(campaign)}%"></i></div>
						{#if campaign.beneficiary_note}<p class="benefit">{campaign.beneficiary_note}</p>{/if}
						<p>{campaign.description || 'Deskripsi campaign akan tampil di sini.'}</p>
					</article>
				</aside>
			</section>
		{:else if tab === 'donations'}
			<section><div class="section-head"><h2>Donation status</h2><button onclick={loadDonations} disabled={!token}>Load CSV</button></div><table><thead><tr><th>Campaign</th><th>Donor</th><th>Amount</th><th>Status</th><th>Paid</th></tr></thead><tbody>{#each donations as donation}<tr><td>{donation.campaign}</td><td>{donation.donor_name}</td><td>{formatRupiah(donation.amount)}</td><td>{donation.status}</td><td>{donation.paid_at}</td></tr>{/each}</tbody></table></section>
		{:else}
			<section><h2>Settings</h2><p>Token admin sementara untuk alpha. Auth/register nanti.</p></section>
		{/if}
	</div>
</main>

<style>
	:global(body){margin:0;background:linear-gradient(135deg,#04101c,#0b3154 58%,#04101c);color:#e5f3ff;font-family:"Plus Jakarta Sans",Inter,system-ui,sans-serif}main{min-height:100vh;padding:2rem}main>div{max-width:80rem;margin:auto;display:grid;gap:1.25rem}header,section,form,.preview article{border:1px solid rgba(148,163,184,.18);border-radius:1.5rem;background:rgba(6,17,31,.76);padding:1.5rem;box-shadow:0 24px 80px rgba(0,0,0,.25)}header{display:flex;justify-content:space-between;gap:1rem;flex-wrap:wrap}h1{font-size:clamp(2.2rem,5vw,3.5rem);margin:.35rem 0;font-weight:950;letter-spacing:-.06em}h2,h3,p{margin:0}header p:first-child,.badge{color:#6ee7b7;text-transform:uppercase;letter-spacing:.18em;font-size:.75rem;font-weight:900}.topnav{display:flex;gap:.75rem;flex-wrap:wrap}.topnav a,.button{border-radius:.85rem;background:rgba(8,24,44,.72);color:#e5f3ff;padding:.75rem 1rem;font-weight:950;text-decoration:none;display:inline-flex;align-items:center;justify-content:center}.topnav a.active{outline:2px solid #22c55e}.builder{display:grid;gap:1.25rem}.two{display:grid;grid-template-columns:1fr 1fr;gap:.8rem}label{display:block;margin-top:.75rem;color:#b8c8dd;font-size:.76rem;font-weight:850}input,textarea,select{box-sizing:border-box;margin-top:.25rem;width:100%;border:1px solid rgba(148,163,184,.26);border-radius:.85rem;background:rgba(8,24,44,.72);color:#f8fafc;padding:.78rem .85rem;font:inherit}textarea{min-height:7rem}button,.button{border:0;border-radius:.85rem;background:#22c55e;color:#04101c;padding:.75rem 1rem;font-weight:950;cursor:pointer}.button.ghost{background:transparent;color:#86efac;border:1px solid rgba(134,239,172,.35)}button:disabled{opacity:.55;cursor:not-allowed}.actions,.section-head{display:flex;gap:.75rem;justify-content:space-between;align-items:center}.notice{border-radius:1rem;padding:1rem;font-weight:800}.ok{background:rgba(16,185,129,.16)}.bad{background:rgba(239,68,68,.18)}.preview img{width:100%;height:220px;object-fit:cover;border-radius:1rem;margin-bottom:1rem}.preview h3{font-size:2rem;line-height:1;margin:.6rem 0}.meter{margin:1rem 0;padding:1rem;border-radius:1rem;background:rgba(15,23,42,.75);display:grid;gap:.5rem}.meter span{color:#94a3b8}.meter i{display:block;height:.55rem;border-radius:999px;background:linear-gradient(90deg,#22c55e,#facc15)}.benefit{border:1px solid rgba(245,158,11,.28);background:rgba(245,158,11,.09);border-radius:1rem;padding:1rem;color:#fde68a!important}.campaign-cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:1rem}.campaign-cards article{overflow:hidden;border:1px solid rgba(148,163,184,.18);border-radius:1.2rem;background:rgba(8,24,44,.65)}.campaign-cards img{width:100%;height:180px;object-fit:cover}.card-body{display:grid;gap:.9rem;padding:1.1rem}.card-title{display:flex;gap:1rem;align-items:flex-start;justify-content:space-between}.card-title span{border-radius:999px;background:rgba(250,204,21,.13);color:#fde68a;padding:.35rem .65rem;font-size:.7rem;font-weight:900;text-transform:uppercase}.loc{color:#86efac;text-transform:uppercase;letter-spacing:.14em;font-size:.76rem;font-weight:900}.desc{color:#b8c8dd;line-height:1.55}.bar{height:.5rem;border-radius:999px;background:rgba(15,23,42,.85);overflow:hidden}.bar i{display:block;height:100%;border-radius:999px;background:#22c55e}.amounts{display:flex;justify-content:space-between;gap:1rem;color:#94a3b8;font-size:.85rem}table{width:100%;border-collapse:collapse}th,td{padding:.75rem;text-align:left}th{color:#6ee7b7}@media(min-width:980px){.builder{grid-template-columns:520px 1fr}}@media(max-width:760px){main{padding:1rem}.two{grid-template-columns:1fr}}
</style>
