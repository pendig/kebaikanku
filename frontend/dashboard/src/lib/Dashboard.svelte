<script>
	import { onMount } from 'svelte';
	import { page } from '$app/state';

	const apiBase = (import.meta.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');
	const landingBase = (import.meta.env.PUBLIC_LANDING_BASE_URL || 'http://127.0.0.1:18383').replace(/\/$/, '');

	function toDatetimeLocal(value) {
		const date = value ? new Date(value) : new Date();
		if (Number.isNaN(date.getTime())) return defaultEndDate();
		const pad = (number) => String(number).padStart(2, '0');
		return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
	}

	function defaultEndDate() {
		const date = new Date();
		date.setDate(date.getDate() + 30);
		date.setHours(23, 59, 0, 0);
		return toDatetimeLocal(date);
	}

	const blankCampaign = () => ({
		id: '',
		organization_id: import.meta.env.PUBLIC_ORGANIZATION_ID || 'pilot-org',
		title: '',
		slug: '',
		description: '',
		category: 'infak',
		subcategory: '',
		campaign_type: 'target_deadline',
		banner_url: '',
		location: '',
		beneficiary_note: '',
		target_amount: 10000000,
		end_date: defaultEndDate()
	});

	let { initialTab = 'overview' } = $props();
	let authState = $state('loading');
	let password = $state('');
	let authError = $state('');
	let mustChangePassword = $state(false);
	let tab = $state('overview');
	let campaigns = $state([]);
	let donations = $state([]);
	let donationFilter = $state('all');
	let donationCampaign = $state('all');
	let donationSort = $state('latest');
	let donationView = $state('table');
	let donationPage = $state(1);
	let donationLimit = 20;
	let donationTotal = $state(0);
	let campaignPage = $state(1);
	let campaignView = $state('table');
	let uploading = $state(false);
	let uploadProgress = $state(0);
	let uploadError = $state('');
	let paymentSettings = $state({ mode: 'sandbox', client_key: '', server_key: '', client_key_masked: '', server_key_masked: '', client_key_configured: false, server_key_configured: false, source: 'environment' });
	let editingPaymentSettings = $state(false);
	let settingsLoading = $state(false);
	let message = $state('');
	let error = $state('');
	let loading = $state(false);
	let downloadLoading = $state(false);
	let sidebarCollapsed = $state(false);
	let campaign = $state(blankCampaign());

	let activeCampaigns = $derived(campaigns.filter((item) => item.status === 'active'));
	let totalRaised = $derived(campaigns.reduce((sum, item) => sum + Number(item.collected_amount || 0), 0));
	let paidDonations = $derived(donations.filter((item) => ['success', 'settlement', 'paid'].includes(String(item.status).toLowerCase())));
	let filteredDonations = $derived(donations);
	let donationPages = $derived(Math.max(1, Math.ceil(donationTotal / donationLimit)));
	let campaignPages = $derived(Math.max(1, Math.ceil(campaigns.length / 10)));
	let visibleCampaigns = $derived(campaigns.slice((campaignPage - 1) * 10, campaignPage * 10));

	onMount(async () => {
		tab = initialTab;
		const mobile = matchMedia('(max-width: 840px)').matches;
		campaignView = localStorage.getItem(`campaign-view-${mobile ? 'mobile' : 'desktop'}`) || (mobile ? 'card' : 'table');
		donationView = localStorage.getItem(`donation-view-${mobile ? 'mobile' : 'desktop'}`) || (mobile ? 'card' : 'table');
		await checkSession();
	});

	async function api(path, options = {}) {
		const response = await fetch(`${apiBase}${path}`, {
			credentials: 'include',
			...options,
			headers: { ...(options.headers || {}) }
		});
		if (response.status === 401) authState = 'unauthenticated';
		return response;
	}

	async function checkSession() {
		authState = 'loading';
		authError = '';
		try {
			const response = await api('/api/v1/admin/session');
			const payload = await response.json().catch(() => null);
			if (!response.ok || payload?.success === false) {
				authState = 'unauthenticated';
				return;
			}
			mustChangePassword = Boolean(payload?.data?.session?.must_change_password ?? payload?.data?.must_change_password ?? payload?.data?.session?.using_default_password);
			authState = 'authenticated';
			await Promise.all([loadCampaigns(), loadDonations(), loadPaymentSettings()]);
		} catch {
			authState = 'unauthenticated';
			authError = 'Tidak dapat menghubungi layanan admin. Coba lagi.';
		}
	}

	async function login() {
		authError = '';
		loading = true;
		try {
			const response = await api('/api/v1/admin/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ password })
			});
			const payload = await response.json().catch(() => null);
			if (!response.ok || payload?.success === false) throw new Error(payload?.error?.message || 'Password tidak valid.');
			password = '';
			mustChangePassword = Boolean(payload?.data?.session?.must_change_password ?? payload?.data?.must_change_password ?? payload?.data?.session?.using_default_password);
			authState = 'authenticated';
			await Promise.all([loadCampaigns(), loadDonations(), loadPaymentSettings()]);
		} catch (err) {
			authState = 'unauthenticated';
			authError = err.message || 'Tidak dapat masuk.';
		} finally {
			loading = false;
		}
	}

	async function logout() {
		try {
			await api('/api/v1/admin/logout', { method: 'POST' });
		} finally {
			authState = 'unauthenticated';
			campaigns = [];
			donations = [];
			message = '';
			error = '';
			sidebarCollapsed = false;
		}
	}

	async function loadCampaigns() {
		loading = true;
		error = '';
		try {
			const response = await api('/api/v1/campaigns?limit=100&include_inactive=true');
			const payload = await response.json();
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Campaign tidak dapat dimuat.');
			campaigns = payload?.data?.campaigns || [];
			const editId = page.url.searchParams.get('id');
			if (initialTab === 'campaign_form' && editId) {
				const existing = campaigns.find((item) => item.id === editId);
				if (existing) fillCampaign(existing);
			}
		} catch (err) {
			error = err.message || 'Campaign tidak dapat dimuat.';
		} finally {
			loading = false;
		}
	}

	async function saveCampaign() {
		loading = true;
		error = '';
		message = '';
		try {
			const endDate = new Date(campaign.end_date);
			if (Number.isNaN(endDate.getTime())) throw new Error('Tanggal berakhir wajib diisi.');
			const body = JSON.stringify({
				...campaign,
				campaign_type: 'target_deadline',
				target_amount: Number(campaign.target_amount),
				end_date: endDate.toISOString()
			});
			const response = await api(campaign.id ? `/api/v1/campaigns/${campaign.id}` : '/api/v1/campaigns', {
				method: campaign.id ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body
			});
			const payload = await response.json();
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Campaign tidak dapat disimpan.');
			const saved = payload?.data?.campaign;
			window.location.href = `${landingBase}/campaigns/${saved?.slug || campaign.slug}`;
		} catch (err) {
			error = err.message || 'Campaign tidak dapat disimpan.';
		} finally {
			loading = false;
		}
	}

	async function updateCampaignStatus(item, status) {
		loading = true;
		error = '';
		try {
			const response = await api(`/api/v1/campaigns/${item.id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status })
			});
			const payload = await response.json().catch(() => null);
			if (!response.ok || payload?.success === false) throw new Error(payload?.error?.message || 'Status campaign tidak dapat diperbarui.');
			message = status === 'active' ? 'Campaign diaktifkan kembali.' : status === 'paused' ? 'Campaign dijeda.' : 'Campaign diselesaikan.';
			await loadCampaigns();
		} catch (err) {
			error = err.message || 'Status campaign tidak dapat diperbarui.';
		} finally {
			loading = false;
		}
	}

	async function loadDonations() {
		try {
			const params = new URLSearchParams({ page: donationPage, limit: donationLimit, sort: donationSort });
			if (donationFilter !== 'all') params.set('status', donationFilter);
			if (donationCampaign !== 'all') params.set('campaign_id', donationCampaign);
			const response = await api(`/api/v1/admin/donations?${params}`);
			const payload = await response.json().catch(() => null);
			if (!response.ok || payload?.success === false) throw new Error(payload?.error?.message || 'Donasi tidak dapat dimuat.');
			donations = payload?.data?.donations || [];
			donationTotal = Number(payload?.data?.pagination?.total ?? payload?.data?.total ?? donations.length);
		} catch (err) {
			error = err.message || 'Donasi tidak dapat dimuat.';
		}
	}

	function changeDonationFilter() {
		donationPage = 1;
		loadDonations();
	}

	function setCampaignView(view) {
		campaignView = view;
		localStorage.setItem(`campaign-view-${matchMedia('(max-width: 840px)').matches ? 'mobile' : 'desktop'}`, view);
	}

	function setDonationView(view) {
		donationView = view;
		localStorage.setItem(`donation-view-${matchMedia('(max-width: 840px)').matches ? 'mobile' : 'desktop'}`, view);
	}

	function uploadBanner(file) {
		if (!file) return;
		uploadError = '';
		if (!file.type.startsWith('image/')) { uploadError = 'Pilih file gambar.'; return; }
		if (file.size > 5 * 1024 * 1024) { uploadError = 'Ukuran gambar maksimal 5 MB.'; return; }
		uploading = true;
		uploadProgress = 0;
		const request = new XMLHttpRequest();
		request.open('POST', `${apiBase}/api/v1/admin/uploads`);
		request.withCredentials = true;
		request.upload.onprogress = (event) => { if (event.lengthComputable) uploadProgress = Math.round((event.loaded / event.total) * 100); };
		request.onload = () => {
			uploading = false;
			const payload = (() => { try { return JSON.parse(request.responseText); } catch { return null; } })();
			if (request.status < 200 || request.status >= 300 || !payload?.data?.url) { uploadError = payload?.error?.message || 'Gambar tidak dapat diunggah.'; return; }
			campaign.banner_url = payload.data.url;
			uploadProgress = 100;
		};
		request.onerror = () => { uploading = false; uploadError = 'Koneksi terputus saat mengunggah gambar.'; };
		const body = new FormData();
		body.append('file', file);
		request.send(body);
	}

	async function loadPaymentSettings() {
		try {
			const response = await api('/api/v1/admin/settings/payment');
			const payload = await response.json().catch(() => null);
			if (!response.ok || payload?.success === false) return;
			paymentSettings = { ...paymentSettings, ...(payload.data?.payment || {}), client_key: '', server_key: '' };
		} catch { /* Pengaturan tidak menghalangi halaman lain. */ }
	}

	async function savePaymentSettings() {
		settingsLoading = true;
		error = '';
		message = '';
		try {
			const response = await api('/api/v1/admin/settings/payment', {
				method: 'PUT', headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ mode: paymentSettings.mode, client_key: paymentSettings.client_key, server_key: paymentSettings.server_key })
			});
			const payload = await response.json().catch(() => null);
			if (!response.ok || payload?.success === false) throw new Error(payload?.error?.message || 'Pengaturan pembayaran tidak dapat disimpan.');
			message = 'Pengaturan Midtrans disimpan.';
			await loadPaymentSettings();
			editingPaymentSettings = false;
		} catch (err) { error = err.message === 'Midtrans keys do not match the selected mode.' ? 'Jenis key tidak cocok dengan mode. Gunakan key SB-Mid-* untuk Sandbox atau Mid-* untuk Production.' : err.message || 'Pengaturan pembayaran tidak dapat disimpan.'; }
		finally { settingsLoading = false; }
	}

	async function downloadDonations() {
		downloadLoading = true;
		error = '';
		try {
			const response = await api('/api/v1/donations/export');
			if (!response.ok) throw new Error((await response.text()) || 'CSV tidak dapat diunduh.');
			const blob = await response.blob();
			const url = URL.createObjectURL(blob);
			const link = document.createElement('a');
			link.href = url;
			link.download = 'donations.csv';
			document.body.appendChild(link);
			link.click();
			link.remove();
			URL.revokeObjectURL(url);
		} catch (err) {
			error = err.message || 'CSV tidak dapat diunduh.';
		} finally {
			downloadLoading = false;
		}
	}

	function fillCampaign(item) {
		campaign = { ...blankCampaign(), ...item, campaign_type: 'target_deadline', end_date: toDatetimeLocal(item.end_date) };
		message = 'Mengedit campaign.';
	}

	function slugify() {
		campaign.slug = campaign.title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
	}

	function parseCSV(csv) {
		const rows = [];
		let row = [], cell = '', quoted = false;
		for (let index = 0; index < csv.length; index++) {
			const char = csv[index];
			if (char === '"' && quoted && csv[index + 1] === '"') { cell += char; index++; }
			else if (char === '"') quoted = !quoted;
			else if (char === ',' && !quoted) { row.push(cell); cell = ''; }
			else if ((char === '\n' || char === '\r') && !quoted) {
				if (char === '\r' && csv[index + 1] === '\n') index++;
				row.push(cell); if (row.some(Boolean)) rows.push(row); row = []; cell = '';
			} else cell += char;
		}
		row.push(cell); if (row.some(Boolean)) rows.push(row);
		const [headers = [], ...records] = rows;
		return records.map((values) => Object.fromEntries(headers.map((key, index) => [key, values[index] || ''])));
	}

	function formatRupiah(amount) {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(amount || 0));
	}

	function formatAmountInput(amount) {
		return new Intl.NumberFormat('id-ID').format(Number(amount || 0));
	}

	function updateTargetAmount(event) {
		campaign.target_amount = Number(event.currentTarget.value.replace(/\D/g, '')) || 0;
		event.currentTarget.value = formatAmountInput(campaign.target_amount);
	}

	function formatDate(value) {
		const date = new Date(value);
		return Number.isNaN(date.getTime()) ? '—' : new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(date);
	}

	function progress(item) {
		if (!item?.target_amount) return 0;
		return Math.min(Math.round(((item.collected_amount || 0) / item.target_amount) * 100), 100);
	}

	function statusLabel(status) {
		return { active: 'Aktif', paused: 'Dijeda', completed: 'Selesai', pending: 'Menunggu', success: 'Berhasil', settlement: 'Berhasil', paid: 'Berhasil', failed: 'Gagal', expired: 'Kedaluwarsa' }[String(status).toLowerCase()] || status || '—';
	}

</script>

<svelte:head>
	<title>Admin | kebaikanku.id</title>
	<link rel="icon" href="/favicon.svg" />
</svelte:head>

{#if authState === 'loading'}
	<main class="auth-stage" aria-live="polite">
		<div class="loading-mark"><span></span><strong>kebaikanku.id</strong><p>Memeriksa sesi admin…</p></div>
	</main>
{:else if authState === 'unauthenticated'}
	<main class="auth-stage">
		<section class="login-shell" aria-labelledby="login-title">
			<div class="login-intro">
				<div class="brand"><span></span> kebaikanku.id</div>
				<p class="eyebrow">RUANG KERJA OPERATOR</p>
				<h1>Operasional campaign yang rapi dan terukur.</h1>
				<p>Kelola campaign, pantau donasi, dan unduh rekonsiliasi dari satu ruang kerja aman.</p>
			</div>
			<form class="login-form" onsubmit={(event) => { event.preventDefault(); login(); }}>
				<p class="eyebrow">MASUK ADMIN</p>
				<h2 id="login-title">Selamat datang kembali</h2>
				<p>Gunakan password admin yang diberikan pengelola sistem.</p>
				<label for="admin-password">Password admin</label>
				<input id="admin-password" bind:value={password} type="password" autocomplete="current-password" required aria-describedby="password-help" />
				<p id="password-help" class="field-help">Sesi disimpan sebagai cookie aman pada browser ini.</p>
				{#if authError}<p class="form-error" role="alert">{authError}</p>{/if}
				<button class="primary-action" disabled={loading}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M10 17l5-5-5-5m5 5H3m8-8h7v16h-7" /></svg>{loading ? 'Memproses…' : 'Masuk ke dashboard'}</button>
			</form>
		</section>
	</main>
{:else}
	<div class:sidebar-collapsed={sidebarCollapsed} class="app-shell">
		<aside class:collapsed={sidebarCollapsed} class="sidebar" aria-label="Navigasi admin">
			<div class="sidebar-brand-row"><div class="brand"><span class="brand-avatar"><img src="/favicon.svg" alt="" /></span><b>kebaikanku.id</b></div><button class="sidebar-toggle" onclick={() => (sidebarCollapsed = !sidebarCollapsed)} aria-label={sidebarCollapsed ? 'Buka sidebar' : 'Ciutkan sidebar'} aria-pressed={sidebarCollapsed}><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m15 18-6-6 6-6" /></svg></button></div>
			<p class="workspace-name">Workspace lembaga</p>
			<nav>
				<a class:active={tab === 'overview'} href="/" aria-label="Ringkasan"><svg class="nav-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 12 12 5l8 7v8h-5v-6H9v6H4v-8Z" /></svg><span>Ringkasan</span></a>
				<a class:active={tab === 'campaigns' || tab === 'campaign_form'} href="/campaigns" aria-label="Campaign"><svg class="nav-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 19.5V5.8A1.8 1.8 0 0 1 5.8 4h12.4A1.8 1.8 0 0 1 20 5.8v13.7M8 4v16m4-12h4m-4 4h4m-4 4h3" /></svg><span>Campaign</span></a>
				<a class:active={tab === 'donations'} href="/donations" aria-label="Donasi"><svg class="nav-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M3 6h18M6 3v6m12-6v6M5 10h14v10H5V10Zm4 4h6m-6 3h3" /></svg><span>Donasi</span></a>
				<a class:active={tab === 'settings'} href="/settings" aria-label="Pengaturan"><svg class="nav-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 15.25A3.25 3.25 0 1 0 12 8.75a3.25 3.25 0 0 0 0 6.5Zm0-11.25v2m0 12v2m8-8h-2M6 12H4m13.66-5.66-1.41 1.41M7.75 16.25l-1.41 1.41m11.32 0-1.41-1.41M7.75 7.75 6.34 6.34" /></svg><span>Pengaturan</span></a>
			</nav>
			<div class="sidebar-bottom">
				<p>Admin session</p>
				<button class="text-action" onclick={logout}><svg class="nav-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M10 17l5-5-5-5m5 5H3m8-8h7v16h-7" /></svg><span>Keluar</span></button>
			</div>
		</aside>

		<div class="main-frame">
			<header class="topbar">
				<div>
					<p class="eyebrow">ADMIN OPERASIONAL</p>
					<strong>{tab === 'overview' ? 'Ringkasan' : tab === 'campaign_form' ? (campaign.id ? 'Edit campaign' : 'Campaign baru') : tab === 'campaigns' ? 'Campaign' : tab === 'donations' ? 'Donasi' : 'Pengaturan'}</strong>
				</div>
				<div class="admin-indicator"><span class="admin-avatar"><img src="/favicon.svg" alt="" /></span><span class="online-dot"></span> Admin</div>
			</header>

			<main class="content">
				{#if mustChangePassword}<p class="notice warning" role="alert">Password admin dibuat otomatis saat instalasi. Atur <code>ADMIN_PASSWORD</code> lalu restart layanan sebelum digunakan di produksi.</p>{/if}
				{#if message}<p class="notice success" role="status">{message}</p>{/if}
				{#if error}<p class="notice danger" role="alert">{error}</p>{/if}

				{#if tab === 'overview'}
					<section class="page-heading"><div><p class="eyebrow">GAMBARAN HARI INI</p><h1>Operasional dalam kendali.</h1><p>Pantau campaign dan donasi yang masuk tanpa berpindah alat.</p></div><a class="primary-action" href="/campaign/form"><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14" /></svg>Buat campaign</a></section>
					<section class="metrics" aria-label="Metrik campaign">
						<article><p>Campaign aktif</p><strong>{activeCampaigns.length}</strong><span>dari {campaigns.length} campaign</span></article>
						<article><p>Dana terkumpul</p><strong>{formatRupiah(totalRaised)}</strong><span>akumulasi campaign</span></article>
						<article><p>Donasi berhasil</p><strong>{paidDonations.length}</strong><span>dari {donations.length} transaksi</span></article>
					</section>
					<section class="data-panel"><div class="section-header"><div><h2>Campaign terbaru</h2><p>Campaign aktif dan progres pengumpulan dana.</p></div><a class="text-link" href="/campaigns">Lihat semua</a></div>
						<div class="campaign-list">
							{#if loading}<p class="empty-state">Memuat campaign…</p>{:else if campaigns.length === 0}<p class="empty-state">Belum ada campaign. Buat campaign pertama untuk memulai.</p>{:else}
								{#each campaigns.slice(0, 5) as item (item.id)}
									<div class="campaign-row"><div class="campaign-name"><strong>{item.title}</strong><span>{item.category}{item.location ? ` · ${item.location}` : ''}</span></div><div class="progress-cell"><span>{formatRupiah(item.collected_amount)} / {formatRupiah(item.target_amount)}</span><div class="progress"><i style="width: {progress(item)}%"></i></div></div><span class="status {item.status}">{statusLabel(item.status)}</span></div>
								{/each}
							{/if}
						</div>
					</section>
				{:else if tab === 'campaigns'}
					<section class="page-heading"><div><p class="eyebrow">PORTOFOLIO CAMPAIGN</p><h1>Campaign</h1><p>Kelola status, informasi publik, dan target pengumpulan dana.</p></div><a class="primary-action" href="/campaign/form"><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14" /></svg>Buat campaign</a></section>
					<section class="data-panel table-panel"><div class="section-header"><div><h2>Semua campaign</h2><p>{campaigns.length} campaign tercatat.</p></div><div class="view-toolbar" aria-label="Tampilan campaign"><button class:active={campaignView === 'table'} onclick={() => setCampaignView('table')} aria-pressed={campaignView === 'table'}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 5h16v14H4V5Zm0 5h16M9 5v14" /></svg>Tabel</button><button class:active={campaignView === 'card'} onclick={() => setCampaignView('card')} aria-pressed={campaignView === 'card'}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 4h7v7H4V4Zm9 0h7v7h-7V4ZM4 13h7v7H4v-7Zm9 0h7v7h-7v-7Z" /></svg>Kartu</button></div></div>
						{#if campaignView === 'table'}<div class="table-scroll"><table><thead><tr><th>Campaign</th><th>Target</th><th>Terkumpul</th><th>Berakhir</th><th>Status</th><th>Aksi</th></tr></thead><tbody>
							{#each visibleCampaigns as item (item.id)}
								<tr><td><strong>{item.title}</strong><span>{item.category}{item.location ? ` · ${item.location}` : ''}</span></td><td>{formatRupiah(item.target_amount)}</td><td>{formatRupiah(item.collected_amount)}</td><td>{formatDate(item.end_date)}</td><td><span class="status {item.status}">{statusLabel(item.status)}</span></td><td><div class="row-actions"><a class="text-link action-link" href="/campaign/form?id={item.id}"><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m4 16.5-.5 4 4-.5L18 9.5 14.5 6 4 16.5ZM13.5 7l3.5 3.5" /></svg>Edit</a><a class="text-link action-link" href="{landingBase}/campaigns/{item.slug}" target="_blank" rel="noopener noreferrer"><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M3 12s3.25-5 9-5 9 5 9 5-3.25 5-9 5-9-5-9-5Zm9 2a2 2 0 1 0 0-4 2 2 0 0 0 0 4Z" /></svg>Lihat</a>{#if item.status === 'active'}<button class="text-link action-link" onclick={() => updateCampaignStatus(item, 'paused')} disabled={loading}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M8 6v12m8-12v12" /></svg>Jeda</button><button class="text-link action-link" onclick={() => updateCampaignStatus(item, 'completed')} disabled={loading}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m5 12 4 4L19 6" /></svg>Selesaikan</button>{:else if item.status === 'paused'}<button class="text-link action-link" onclick={() => updateCampaignStatus(item, 'active')} disabled={loading}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m9 6 7 6-7 6V6Z" /></svg>Aktifkan</button><button class="text-link action-link" onclick={() => updateCampaignStatus(item, 'completed')} disabled={loading}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m5 12 4 4L19 6" /></svg>Selesaikan</button>{/if}</div></td></tr>
							{:else}<tr><td colspan="6" class="empty-state">Belum ada campaign.</td></tr>{/each}
						</tbody></table></div>{:else}<div class="campaign-cards">{#each visibleCampaigns as item (item.id)}<article><div><span class="status {item.status}">{statusLabel(item.status)}</span><h3>{item.title}</h3><p>{item.category}{item.location ? ` · ${item.location}` : ''}</p></div><div class="card-numbers"><span>Terkumpul <strong>{formatRupiah(item.collected_amount)}</strong></span><span>Target <strong>{formatRupiah(item.target_amount)}</strong></span></div><div class="progress"><i style="width: {progress(item)}%"></i></div><p>Berakhir {formatDate(item.end_date)}</p><div class="row-actions"><a class="text-link action-link" href="/campaign/form?id={item.id}">Edit</a><a class="text-link action-link" href="{landingBase}/campaigns/{item.slug}" target="_blank" rel="noopener noreferrer">Lihat</a>{#if item.status === 'active'}<button class="text-link action-link" onclick={() => updateCampaignStatus(item, 'paused')}>Jeda</button><button class="text-link action-link" onclick={() => updateCampaignStatus(item, 'completed')}>Selesaikan</button>{:else if item.status === 'paused'}<button class="text-link action-link" onclick={() => updateCampaignStatus(item, 'active')}>Aktifkan</button>{/if}</div></article>{:else}<p class="empty-state">Belum ada campaign.</p>{/each}</div>{/if}
						{#if campaignPages > 1}<nav class="pagination" aria-label="Halaman campaign"><button disabled={campaignPage === 1} onclick={() => campaignPage--}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m15 18-6-6 6-6" /></svg>Sebelumnya</button><span>Halaman {campaignPage} dari {campaignPages}</span><button disabled={campaignPage === campaignPages} onclick={() => campaignPage++}>Berikutnya<svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m9 6 6 6-6 6" /></svg></button></nav>{/if}
					</section>
				{:else if tab === 'campaign_form'}
					<section class="page-heading"><div><p class="eyebrow">{campaign.id ? 'PERBARUI CAMPAIGN' : 'CAMPAIGN BARU'}</p><h1>{campaign.id ? campaign.title || 'Edit campaign' : 'Buat campaign'}</h1><p>Campaign memakai target dan batas waktu agar publik mendapat konteks yang jelas.</p></div></section>
					<section class="form-layout"><form class="data-panel campaign-form" onsubmit={(event) => { event.preventDefault(); saveCampaign(); }}><div class="section-header"><div><h2>Informasi campaign</h2><p>Bidang bertanda wajib diisi sebelum diterbitkan.</p></div></div>
						<label for="campaign-title">Judul<input id="campaign-title" bind:value={campaign.title} oninput={slugify} required /></label>
						<label for="campaign-description">Deskripsi<textarea id="campaign-description" bind:value={campaign.description} required></textarea></label>
						<div class="field-grid"><label for="campaign-target">Target donasi<span class="money-input"><b>Rp</b><input id="campaign-target" value={formatAmountInput(campaign.target_amount)} oninput={updateTargetAmount} inputmode="numeric" autocomplete="off" required /></span></label><label for="campaign-end-date">Batas waktu<input id="campaign-end-date" bind:value={campaign.end_date} type="datetime-local" required /></label></div>
						<details class="secondary-fields"><summary>Informasi tambahan <span>Slug, kategori, banner, dan detail penerima</span></summary><div class="details-content">
							<label for="campaign-slug">Slug<input id="campaign-slug" bind:value={campaign.slug} required pattern="[a-z0-9-]+" /></label>
							<div class="field-grid"><label for="campaign-category">Kategori<select id="campaign-category" bind:value={campaign.category}><option value="zakat_maal">Zakat Maal</option><option value="zakat_fitrah">Zakat Fitrah</option><option value="infak">Infak</option><option value="kemanusiaan">Kemanusiaan</option></select></label><label for="campaign-subcategory">Subkategori <span>(opsional)</span><input id="campaign-subcategory" bind:value={campaign.subcategory} placeholder="Pendidikan, Palestina…" /></label></div>
							<fieldset class="upload-field"><legend>Banner <span>(opsional)</span></legend><label for="campaign-upload">Pilih gambar<input id="campaign-upload" type="file" accept="image/png,image/jpeg,image/webp" onchange={(event) => uploadBanner(event.currentTarget.files?.[0])} /></label><span class="field-help">JPG, PNG, atau WebP maksimal 5 MB.</span>{#if uploading}<progress max="100" value={uploadProgress}>{uploadProgress}%</progress>{/if}{#if uploadError}<p class="form-error" role="alert">{uploadError}</p>{/if}{#if campaign.banner_url}<img class="banner-preview" src={campaign.banner_url} alt="Pratinjau banner campaign" />{/if}<label for="campaign-banner">Atau URL gambar<input id="campaign-banner" bind:value={campaign.banner_url} type="url" placeholder="https://…" /></label></fieldset>
							<label for="campaign-location">Lokasi <span>(opsional)</span><input id="campaign-location" bind:value={campaign.location} placeholder="Riau, Indonesia" /></label>
							<label for="campaign-beneficiary">Catatan penerima <span>(opsional)</span><textarea id="campaign-beneficiary" bind:value={campaign.beneficiary_note} placeholder="Contoh: 120 santri dan warga sekitar"></textarea></label>
						</div></details>
						<div class="form-actions"><button class="primary-action" disabled={loading}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 4h12l2 2v14H5V4Zm3 0v6h8V4M8 20v-6h8v6" /></svg>{loading ? 'Menyimpan…' : campaign.id ? 'Simpan perubahan' : 'Terbitkan campaign'}</button>{#if campaign.id}<a class="secondary-action" href="/campaign/form"><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m6 6 12 12M18 6 6 18" /></svg>Batal edit</a>{/if}</div>
					</form>
						<aside class="preview-panel">{#if campaign.banner_url}<img class="preview-banner" src={campaign.banner_url} alt="" role="presentation" />{/if}<p class="eyebrow">PRATINJAU PUBLIK</p><span class="status active">{campaign.category}</span><h2>{campaign.title || 'Judul campaign'}</h2><p>{campaign.location || 'Lokasi campaign'}</p><div class="preview-amount"><strong>{formatRupiah(0)}</strong><span>dari {formatRupiah(campaign.target_amount)}</span><div class="progress"><i style="width: {progress(campaign)}%"></i></div></div><p>{campaign.description || 'Deskripsi campaign akan tampil di sini.'}</p>{#if campaign.beneficiary_note}<p class="beneficiary">{campaign.beneficiary_note}</p>{/if}</aside>
					</section>
				{:else if tab === 'donations'}
					<section class="page-heading"><div><p class="eyebrow">REKONSILIASI DONASI</p><h1>Donasi</h1><p>Periksa transaksi yang masuk dan unduh CSV untuk pencatatan lembaga.</p></div></section>
					<section class="data-panel table-panel"><div class="section-header"><div><h2>Transaksi</h2><p>{donationTotal} transaksi ditemukan.</p></div><div class="view-toolbar icon-toolbar" aria-label="Tampilan donasi"><button class:active={donationView === 'table'} onclick={() => setDonationView('table')} aria-label="Tampilan tabel" title="Tampilan tabel" aria-pressed={donationView === 'table'}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 5h16v14H4V5Zm0 5h16M9 5v14" /></svg></button><button class:active={donationView === 'card'} onclick={() => setDonationView('card')} aria-label="Tampilan kartu" title="Tampilan kartu" aria-pressed={donationView === 'card'}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 4h7v7H4V4Zm9 0h7v7h-7V4ZM4 13h7v7H4v-7Zm9 0h7v7h-7v-7Z" /></svg></button></div></div>
						<div class="donation-filters"><label for="donation-status">Status<select id="donation-status" bind:value={donationFilter} onchange={changeDonationFilter}><option value="all">Semua status</option><option value="pending">Menunggu</option><option value="success">Berhasil</option><option value="failed">Gagal</option></select></label><label for="donation-campaign">Campaign<select id="donation-campaign" bind:value={donationCampaign} onchange={changeDonationFilter}><option value="all">Semua campaign</option>{#each campaigns as item (item.id)}<option value={item.id}>{item.title}</option>{/each}</select></label><label for="donation-sort">Urutkan<select id="donation-sort" bind:value={donationSort} onchange={changeDonationFilter}><option value="latest">Terbaru</option><option value="oldest">Terlama</option><option value="amount_desc">Nominal terbesar</option><option value="amount_asc">Nominal terkecil</option></select></label><div class="filter-actions"><button class="secondary-action icon-button" onclick={loadDonations} aria-label="Muat ulang donasi" title="Muat ulang"><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M20 6v5h-5M4 18v-5h5m9.5-3A7 7 0 0 0 6.4 6.4L4 9m16 6-2.4 2.6A7 7 0 0 1 5.5 14" /></svg></button><button class="primary-action" onclick={downloadDonations} disabled={downloadLoading}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3v12m-4-4 4 4 4-4M5 19h14" /></svg>{downloadLoading ? 'Menyiapkan…' : 'Unduh CSV'}</button></div></div>
						{#if donationView === 'table'}<div class="table-scroll"><table><thead><tr><th>ID / Order</th><th>Campaign</th><th>Donatur</th><th>Nominal</th><th>Status</th><th>Dibuat</th><th>Dibayar</th></tr></thead><tbody>
							{#each filteredDonations as donation (donation.id)}
								<tr><td><strong>{donation.id || '—'}</strong><span>{donation.provider_order_id || donation.provider_status || '—'}</span></td><td>{donation.campaign?.title || '—'}</td><td>{donation.donor?.name || 'Hamba Allah'}<span>{donation.donor?.phone_number?.startsWith('guest-') ? '' : donation.donor?.phone_number || ''}</span></td><td>{formatRupiah(donation.amount)}</td><td><span class="status {donation.status}">{statusLabel(donation.status)}</span></td><td>{formatDate(donation.created_at)}</td><td>{formatDate(donation.paid_at)}</td></tr>
							{:else}<tr><td colspan="7" class="empty-state">Belum ada transaksi untuk filter ini.</td></tr>{/each}
						</tbody></table></div>{:else}<div class="donation-cards">{#each filteredDonations as donation (donation.id)}<article><div class="card-heading"><div><p>{donation.campaign?.title || 'Tanpa campaign'}</p><h3>{donation.donor?.name || 'Hamba Allah'}</h3></div><span class="status {donation.status}">{statusLabel(donation.status)}</span></div><strong class="donation-amount">{formatRupiah(donation.amount)}</strong><dl><div><dt>Order</dt><dd>{donation.provider_order_id || donation.id}</dd></div><div><dt>Dibuat</dt><dd>{formatDate(donation.created_at)}</dd></div>{#if donation.paid_at}<div><dt>Dibayar</dt><dd>{formatDate(donation.paid_at)}</dd></div>{/if}</dl></article>{:else}<p class="empty-state">Belum ada transaksi untuk filter ini.</p>{/each}</div>{/if}
						<nav class="pagination" aria-label="Halaman donasi"><button disabled={donationPage === 1} onclick={() => { donationPage--; loadDonations(); }}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m15 18-6-6 6-6" /></svg>Sebelumnya</button><span>Halaman {donationPage} dari {donationPages}</span><button disabled={donationPage >= donationPages} onclick={() => { donationPage++; loadDonations(); }}>Berikutnya<svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m9 6 6 6-6 6" /></svg></button></nav>
					</section>
				{:else}
					<section class="page-heading"><div><p class="eyebrow">AKUN & PEMBAYARAN</p><h1>Pengaturan</h1><p>Kelola koneksi pembayaran dan sesi administrator.</p></div></section>
					<section class="settings-grid">
						<form class="data-panel settings-form" onsubmit={(event) => { event.preventDefault(); savePaymentSettings(); }}><div class="settings-title"><div><h2>Midtrans</h2><p>Nilai tersimpan mengoverride environment. Secret dienkripsi oleh server.</p></div>{#if !editingPaymentSettings}<button type="button" class="secondary-action" onclick={() => (editingPaymentSettings = true)}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m4 16.5-.5 4 4-.5L18 9.5 14.5 6 4 16.5ZM13.5 7l3.5 3.5" /></svg>Edit</button>{/if}</div><div class="masked-keys"><p><span>Client key</span><code>{paymentSettings.client_key_masked || 'Belum dikonfigurasi'}</code></p><p><span>Server key</span><code>{paymentSettings.server_key_masked || 'Belum dikonfigurasi'}</code></p></div>{#if editingPaymentSettings}<label for="payment-mode">Mode<select id="payment-mode" bind:value={paymentSettings.mode}><option value="sandbox">Sandbox</option><option value="production">Production</option></select></label><label for="midtrans-client-key">Client key baru <span>(isi bersama server key)</span><input id="midtrans-client-key" type="password" bind:value={paymentSettings.client_key} autocomplete="off" /></label><label for="midtrans-server-key">Server key baru <span>(isi bersama client key)</span><input id="midtrans-server-key" type="password" bind:value={paymentSettings.server_key} autocomplete="off" /></label><div class="form-actions"><button class="primary-action" disabled={settingsLoading}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 4h12l2 2v14H5V4Zm3 0v6h8V4M8 20v-6h8v6" /></svg>{settingsLoading ? 'Menyimpan…' : 'Simpan pengaturan'}</button><button type="button" class="secondary-action" onclick={() => { editingPaymentSettings = false; paymentSettings.client_key = ''; paymentSettings.server_key = ''; }}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m6 6 12 12M18 6 6 18" /></svg>Batal</button></div>{/if}<p class="field-help">Sumber aktif: {paymentSettings.source === 'override' ? 'pengaturan terenkripsi' : 'environment server'}.</p><details class="setup-guide"><summary>Panduan koneksi Midtrans</summary><ol><li>Buka <a href="https://dashboard.midtrans.com/settings/access-keys" target="_blank" rel="noopener noreferrer">Settings → Access Keys</a> di dashboard Midtrans.</li><li>Salin Client Key dan Server Key dari mode yang sama, lalu simpan di sini.</li><li>Isi Payment Notification URL dengan:</li></ol><code>{apiBase || 'https://api.domain-anda.id'}/api/v1/payments/midtrans/notification</code><p>URL harus HTTPS dan dapat diakses publik; alamat localhost tidak dapat dipanggil Midtrans.</p></details></form>
						<section class="data-panel settings-panel"><div><h2>Sesi admin</h2><p>Dashboard memakai cookie sesi aman; tidak ada token admin yang disimpan di browser.</p></div><button class="secondary-action" onclick={logout}><svg class="action-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M10 17l5-5-5-5m5 5H3m8-8h7v16h-7" /></svg>Keluar dari sesi</button></section>
					</section>
				{/if}
			</main>
			<nav class="bottom-tabs" aria-label="Navigasi utama">
				<a class:active={tab === 'overview'} href="/" aria-label="Ringkasan"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 12 12 5l8 7v8h-5v-6H9v6H4v-8Z" /></svg><span>Ringkasan</span></a>
				<a class:active={tab === 'campaigns' || tab === 'campaign_form'} href="/campaigns" aria-label="Campaign"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 19.5V5.8A1.8 1.8 0 0 1 5.8 4h12.4A1.8 1.8 0 0 1 20 5.8v13.7M8 4v16m4-12h4m-4 4h4m-4 4h3" /></svg><span>Campaign</span></a>
				<a class:active={tab === 'donations'} href="/donations" aria-label="Donasi"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 6h18M6 3v6m12-6v6M5 10h14v10H5V10Zm4 4h6m-6 3h3" /></svg><span>Donasi</span></a>
				<a class:active={tab === 'settings'} href="/settings" aria-label="Pengaturan"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 15.25A3.25 3.25 0 1 0 12 8.75a3.25 3.25 0 0 0 0 6.5Zm0-11.25v2m0 12v2m8-8h-2M6 12H4m13.66-5.66-1.41 1.41M7.75 16.25l-1.41 1.41m11.32 0-1.41-1.41M7.75 7.75 6.34 6.34" /></svg><span>Pengaturan</span></a>
			</nav>
		</div>
	</div>
{/if}

<style>
	:global(*) { box-sizing: border-box; }
	:global(body) { margin: 0; background: #eef2ec; color: #17261e; font-family: "Plus Jakarta Sans", system-ui, sans-serif; }
	button, input, textarea, select { font: inherit; }
	button, a { -webkit-tap-highlight-color: transparent; }
	.auth-stage { min-height: 100vh; display: grid; place-items: center; padding: 1.5rem; background: radial-gradient(circle at top left, #dbeac5, transparent 38%), #102b20; }
	.loading-mark { display: grid; gap: .65rem; justify-items: center; color: #f8f3e4; font-family: Outfit, "Plus Jakarta Sans", sans-serif; }
	.loading-mark span, .brand span { display: inline-block; width: .8rem; height: .8rem; border-radius: 99px; background: #eabf55; box-shadow: 0 0 0 6px rgba(234,191,85,.16); }
	.loading-mark p { margin: 0; color: #b9c8ba; font-family: "Plus Jakarta Sans", system-ui, sans-serif; font-size: .88rem; }
	.login-shell { width: min(100%, 58rem); display: grid; grid-template-columns: 1.15fr .85fr; overflow: hidden; border-radius: 1.75rem; background: #f9faf6; box-shadow: 0 2rem 6rem rgba(0,0,0,.32); }
	.login-intro { padding: clamp(2rem, 6vw, 4.5rem); color: #f8f3e4; background: linear-gradient(145deg, #123a2a, #0b2119); }
	.brand { display: flex; align-items: center; gap: .7rem; color: inherit; font: 800 1.15rem Outfit, "Plus Jakarta Sans", sans-serif; letter-spacing: -.03em; }
	.brand-avatar, .admin-avatar { display: grid !important; place-items: center; border-radius: 50%; background: linear-gradient(145deg, #65bd46, #eabf55) !important; color: #102b20; font: 800 .85rem Outfit, sans-serif; box-shadow: none !important; }
	.brand-avatar, .admin-avatar { width: 2.75rem !important; height: 2.75rem !important; }
	.brand-avatar img, .admin-avatar img { width: 1.8rem; height: 1.8rem; }
	.workspace-name { margin: .35rem 0 2rem 1.5rem; color: #8ea899; font-size: .75rem; }
	.eyebrow { margin: 0; color: #739f76; font-size: .68rem; font-weight: 800; letter-spacing: .14em; }
	.login-intro .eyebrow { margin-top: 5rem; color: #eabf55; }
	.login-intro h1 { max-width: 30rem; margin: 1rem 0; font: 700 clamp(2.25rem, 5vw, 4rem)/1.05 Outfit, "Plus Jakarta Sans", sans-serif; letter-spacing: -.055em; }
	.login-intro > p:last-child { max-width: 27rem; color: #c9d8ca; line-height: 1.7; }
	.login-form { display: grid; align-content: center; gap: .75rem; padding: clamp(2rem, 5vw, 3.5rem); }
	.login-form h2 { margin: 0; font: 700 2rem/1.1 Outfit, "Plus Jakarta Sans", sans-serif; letter-spacing: -.045em; }
	.login-form > p:not(.eyebrow):not(.field-help) { margin: 0 0 1rem; color: #617167; font-size: .9rem; line-height: 1.6; }
	label { display: grid; gap: .4rem; color: #34473a; font-size: .8rem; font-weight: 800; }
	label span { color: #829087; font-weight: 500; }
	input, textarea, select { width: 100%; border: 1px solid #d6ddd3; border-radius: .75rem; background: #fff; color: #17261e; padding: .78rem .9rem; transition: border-color .18s ease, box-shadow .18s ease; }
	input:focus, textarea:focus, select:focus { border-color: #3d7c4d; box-shadow: 0 0 0 3px rgba(61,124,77,.16); }
	textarea { min-height: 8rem; resize: vertical; }
	.field-help { margin: 0; color: #718076; font-size: .75rem; line-height: 1.5; }
	.form-error { margin: .25rem 0; border-radius: .75rem; background: #fff0ed; color: #a43b2a; padding: .75rem; font-size: .82rem; font-weight: 700; }
	.primary-action, .secondary-action, .text-action { border: 0; border-radius: .75rem; cursor: pointer; font-weight: 800; text-decoration: none; transition: transform .18s ease, background .18s ease; }
	.primary-action { display: inline-flex; align-items: center; justify-content: center; min-height: 2.75rem; padding: .72rem 1rem; background: #17643c; color: #fffdf6; box-shadow: 0 .65rem 1.1rem rgba(23,100,60,.18); }
	.primary-action, .secondary-action, .pagination button, .view-toolbar button { gap: .4rem; }
	.primary-action:hover { background: #104f2f; transform: translateY(-1px); }
	.secondary-action { display: inline-flex; align-items: center; justify-content: center; min-height: 2.75rem; padding: .7rem .95rem; border: 1px solid #cbd6ca; background: #fff; color: #214b31; }
	.text-action { padding: .45rem 0; background: transparent; color: #d3b15c; text-align: left; }
	button:disabled { cursor: not-allowed; opacity: .6; transform: none; }
	.app-shell { min-height: 100vh; display: grid; grid-template-columns: 16.5rem minmax(0, 1fr); background: #eef2ec; }
	.app-shell.sidebar-collapsed { grid-template-columns: 5rem minmax(0, 1fr); }
	.sidebar { display: flex; position: sticky; top: 0; height: 100vh; flex-direction: column; padding: 1.5rem 1rem; background: #102b20; color: #f8f3e4; }
	.sidebar-brand-row { display: flex; align-items: center; justify-content: space-between; gap: .5rem; }
	.sidebar-toggle { display: grid; flex: 0 0 2.75rem; width: 2.75rem; height: 2.75rem; place-items: center; border: 1px solid rgba(255,255,255,.15); border-radius: .55rem; background: transparent; color: #bdcfc0; cursor: pointer; }
	.sidebar-toggle:hover, .sidebar-toggle:focus-visible { background: rgba(255,255,255,.09); color: #fff; }
	.sidebar-toggle svg { fill: none; stroke: currentColor; stroke-width: 2; }
	.sidebar nav { display: grid; gap: .35rem; }
	.sidebar nav a { display: flex; align-items: center; gap: .75rem; min-height: 2.75rem; border-radius: .7rem; color: #bdcfc0; padding: .8rem 1rem; font-size: .88rem; font-weight: 700; text-decoration: none; }
	.sidebar nav a:hover, .sidebar nav a.active { background: rgba(234,191,85,.13); color: #f8f3e4; }
	.sidebar nav a.active { box-shadow: inset 3px 0 #eabf55; }
	.nav-icon, .action-icon, .bottom-tabs svg { flex: 0 0 auto; width: 1.15rem; height: 1.15rem; fill: none; stroke: currentColor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 1.8; }
	.sidebar.collapsed { align-items: center; padding-inline: .75rem; }
	.sidebar.collapsed .sidebar-brand-row { width: 100%; justify-content: center; }
	.sidebar.collapsed .brand, .sidebar.collapsed .workspace-name, .sidebar.collapsed nav a span, .sidebar.collapsed .sidebar-bottom p, .sidebar.collapsed .sidebar-bottom button span { display: none; }
	.sidebar.collapsed .sidebar-toggle svg { transform: rotate(180deg); }
	.sidebar.collapsed nav { width: 100%; margin-top: 2rem; }
	.sidebar.collapsed nav a { justify-content: center; padding-inline: .7rem; }
	.sidebar.collapsed .sidebar-bottom { width: 100%; padding-inline: 0; }
	.sidebar.collapsed .sidebar-bottom button { display: grid; width: 100%; place-items: center; }
	.sidebar-bottom { margin-top: auto; padding: 1rem; border-top: 1px solid rgba(255,255,255,.1); }
	.sidebar-bottom p { margin: 0 0 .25rem; color: #8ea899; font-size: .72rem; }
	.main-frame { min-width: 0; }
	.topbar { display: flex; min-height: 5.5rem; align-items: center; justify-content: space-between; gap: 1rem; padding: 1rem clamp(1.25rem, 4vw, 3rem); border-bottom: 1px solid #dce4da; background: rgba(250,252,247,.82); backdrop-filter: blur(12px); }
	.topbar strong { display: block; margin-top: .2rem; font: 700 1.25rem Outfit, "Plus Jakarta Sans", sans-serif; letter-spacing: -.035em; }
	.admin-indicator { display: flex; align-items: center; gap: .45rem; color: #48614e; font-size: .8rem; font-weight: 800; }
	.admin-indicator .online-dot { width: .5rem; height: .5rem; border-radius: 99px; background: #4f9c60; }
	.content { display: grid; gap: 1.5rem; max-width: 96rem; margin: 0 auto; padding: clamp(1.25rem, 4vw, 3rem); }
	.page-heading { display: flex; align-items: end; justify-content: space-between; gap: 1rem; }
	.page-heading h1 { margin: .35rem 0; font: 700 clamp(2rem, 4vw, 3rem)/1 Outfit, "Plus Jakarta Sans", sans-serif; letter-spacing: -.06em; }
	.page-heading p:not(.eyebrow) { max-width: 38rem; margin: 0; color: #607063; line-height: 1.6; }
	.notice { margin: 0; border-radius: .8rem; padding: .8rem 1rem; font-size: .85rem; font-weight: 700; }
	.notice.success { background: #e4f3e7; color: #17643c; }
	.notice.danger { background: #fff0ed; color: #a43b2a; }
	.notice.warning { background: #fff4d6; color: #755814; }
	.notice code { font-size: .78rem; }
	.metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; }
	.metrics article { padding: 1.25rem 0; border-bottom: 2px solid #ccd9cc; }
	.metrics article:nth-child(2) { border-color: #eabf55; }
	.metrics p, .metrics span { margin: 0; color: #718076; font-size: .77rem; }
	.metrics strong { display: block; margin: .35rem 0; font: 700 clamp(1.5rem, 3vw, 2rem)/1 Outfit, "Plus Jakarta Sans", sans-serif; letter-spacing: -.05em; }
	.data-panel { border: 1px solid #dce4da; border-radius: 1rem; background: #fbfcf8; padding: clamp(1rem, 3vw, 1.5rem); }
	.section-header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; margin-bottom: 1.25rem; }
	.section-header h2 { margin: 0; font: 700 1.15rem Outfit, "Plus Jakarta Sans", sans-serif; letter-spacing: -.03em; }
	.section-header p { margin: .25rem 0 0; color: #718076; font-size: .8rem; }
	.text-link { border: 0; background: transparent; color: #17643c; font-size: .8rem; font-weight: 800; text-decoration: none; cursor: pointer; }
	.campaign-list { display: grid; }
	.campaign-row { display: grid; grid-template-columns: minmax(10rem, 1fr) minmax(14rem, .9fr) auto; align-items: center; gap: 1rem; padding: 1rem 0; border-top: 1px solid #e3e9e2; }
	.campaign-name, td strong { display: grid; gap: .25rem; }
	.campaign-name span, td span { color: #718076; font-size: .74rem; }
	.progress-cell { display: grid; gap: .45rem; color: #637267; font-size: .72rem; }
	.progress { height: .42rem; overflow: hidden; border-radius: 99px; background: #e1e8df; }
	.progress i { display: block; height: 100%; border-radius: inherit; background: linear-gradient(90deg, #2e8050, #eabf55); }
	.status { display: inline-flex; width: max-content; border-radius: 99px; padding: .3rem .55rem; background: #edf1ec; color: #57675b; font-size: .68rem; font-weight: 800; text-transform: capitalize; }
	.status.active, .status.success, .status.settlement, .status.paid { background: #e2f2e5; color: #187140; }
	.status.paused, .status.pending { background: #fff4d6; color: #80611b; }
	.status.completed { background: #e7ebee; color: #52606a; }
	.status.failed, .status.expired { background: #fff0ed; color: #a43b2a; }
	.table-scroll { overflow-x: auto; }
	table { width: 100%; border-collapse: collapse; text-align: left; }
	th { padding: .7rem .8rem; border-bottom: 1px solid #dce4da; color: #738075; font-size: .68rem; letter-spacing: .08em; text-transform: uppercase; }
	td { padding: .85rem .8rem; border-bottom: 1px solid #e5ebe4; color: #35463a; font-size: .79rem; vertical-align: middle; }
	td strong { color: #1d3023; font-size: .78rem; }
	.row-actions, .toolbar, .form-actions { display: flex; flex-wrap: wrap; align-items: center; gap: .65rem; }
	.view-toolbar { display: flex; flex-wrap: wrap; align-items: center; gap: .4rem; }
	.view-toolbar > button:not(.text-action), .pagination button { min-height: 2.5rem; border: 1px solid #d2ddd1; border-radius: .65rem; background: #fff; color: #526357; padding: .5rem .75rem; font-weight: 750; cursor: pointer; }
	.view-toolbar > button.active { border-color: #17643c; background: #e5f1e6; color: #17643c; }
	.icon-toolbar button, .icon-button { width: 2.75rem; padding: .65rem; }
	.money-input { display: flex; align-items: center; overflow: hidden; border: 1px solid #d6ddd3; border-radius: .75rem; background: #fff; }
	.money-input:focus-within { border-color: #3d7c4d; box-shadow: 0 0 0 3px rgba(61,124,77,.16); }
	.money-input b { padding-left: .9rem; color: #607063; }
	.money-input input { border: 0; border-radius: 0; box-shadow: none; }
	.action-link { display: inline-flex; align-items: center; gap: .3rem; min-height: 2.75rem; padding: .35rem; }
	.empty-state { margin: 0; color: #718076; font-size: .85rem; padding: 1rem 0; }
	.form-layout { display: grid; grid-template-columns: minmax(0, 1.1fr) minmax(16rem, .7fr); align-items: start; gap: 1.5rem; }
	.campaign-form { display: grid; gap: 1rem; }
	.field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
	.secondary-fields { border: 1px solid #dce4da; border-radius: .85rem; background: #f4f7f2; }
	.secondary-fields summary { display: flex; cursor: pointer; justify-content: space-between; gap: 1rem; padding: 1rem; color: #294333; font-weight: 800; }
	.secondary-fields summary span { color: #718076; font-size: .72rem; font-weight: 500; }
	.details-content { display: grid; gap: 1rem; padding: 0 1rem 1rem; }
	.upload-field { display: grid; gap: .7rem; margin: 0; border: 1px dashed #b9c9b7; border-radius: .8rem; padding: 1rem; }
	.upload-field legend { color: #34473a; font-size: .8rem; font-weight: 800; }
	.upload-field legend span { color: #829087; font-weight: 500; }
	.upload-field input[type="file"] { padding: .5rem; }
	.upload-field progress { width: 100%; accent-color: #17643c; }
	.banner-preview, .preview-banner { display: block; width: 100%; object-fit: cover; border-radius: .7rem; }
	.banner-preview { max-height: 14rem; }
	.preview-banner { max-height: 12rem; margin-bottom: 1rem; }
	.preview-panel { position: sticky; top: 1.5rem; padding: 1.5rem; border-radius: 1rem; background: #173b2a; color: #f7f5eb; }
	.preview-panel h2 { margin: .85rem 0 .35rem; font: 700 1.7rem/1.1 Outfit, "Plus Jakarta Sans", sans-serif; letter-spacing: -.045em; }
	.preview-panel > p:not(.eyebrow) { color: #c4d2c5; font-size: .85rem; line-height: 1.65; }
	.preview-amount { display: grid; gap: .35rem; margin: 1.25rem 0; padding: 1rem 0; border-top: 1px solid rgba(255,255,255,.14); border-bottom: 1px solid rgba(255,255,255,.14); }
	.preview-amount strong { font: 700 1.25rem Outfit, "Plus Jakarta Sans", sans-serif; }
	.preview-amount span { color: #b9cbb9; font-size: .78rem; }
	.beneficiary { margin-top: 1rem; border-left: 2px solid #eabf55; padding-left: .8rem; }
	.settings-panel { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
	.settings-panel h2 { margin: 0; font: 700 1.1rem Outfit, "Plus Jakarta Sans", sans-serif; }
	.settings-panel p { margin: .35rem 0 0; color: #657468; font-size: .85rem; }
	.settings-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(16rem, .55fr); align-items: start; gap: 1rem; }
	.settings-form { display: grid; gap: 1rem; }
	.settings-form h2 { margin: 0; }
	.settings-form > div:first-child p { margin: .35rem 0 0; color: #657468; line-height: 1.55; }
	.settings-title { display: flex; align-items: start; justify-content: space-between; gap: 1rem; }
	.masked-keys { display: grid; gap: .6rem; }
	.masked-keys p { display: flex; flex-wrap: wrap; justify-content: space-between; gap: .5rem; margin: 0; border-radius: .7rem; background: #f1f5ef; padding: .75rem; }
	.masked-keys span { color: #718076; font-size: .75rem; font-weight: 750; }
	.masked-keys code { overflow-wrap: anywhere; color: #214b31; font-size: .78rem; }
	.setup-guide { border-top: 1px solid #dce4da; padding-top: .9rem; color: #526357; font-size: .78rem; line-height: 1.6; }
	.setup-guide summary { cursor: pointer; color: #214b31; font-weight: 800; }
	.setup-guide ol { margin: .75rem 0; padding-left: 1.25rem; }
	.setup-guide > code { display: block; overflow-wrap: anywhere; border-radius: .6rem; background: #102b20; color: #f8f3e4; padding: .7rem; }
	.setup-guide a { color: #17643c; font-weight: 700; }
	.campaign-cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 18rem), 1fr)); gap: 1rem; }
	.campaign-cards article { display: grid; gap: .85rem; border: 1px solid #dce4da; border-radius: .9rem; background: #fff; padding: 1rem; }
	.donation-filters { display: grid; grid-template-columns: repeat(3, minmax(10rem, 1fr)) auto; align-items: end; gap: .75rem; margin-bottom: 1rem; }
	.donation-filters label { gap: .35rem; }
	.filter-actions { display: flex; gap: .5rem; }
	.donation-cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 18rem), 1fr)); gap: .8rem; }
	.donation-cards article { display: grid; gap: .8rem; border: 1px solid #dce4da; border-radius: .9rem; background: #fff; padding: 1rem; }
	.card-heading { display: flex; align-items: start; justify-content: space-between; gap: .75rem; }
	.card-heading p, .card-heading h3 { margin: 0; }
	.card-heading p { color: #718076; font-size: .72rem; }
	.card-heading h3 { margin-top: .2rem; font: 700 1rem Outfit, sans-serif; }
	.donation-amount { font: 700 1.35rem Outfit, sans-serif; color: #17643c; }
	.donation-cards dl { display: grid; gap: .55rem; margin: 0; }
	.donation-cards dl div { display: grid; grid-template-columns: 4rem minmax(0, 1fr); gap: .5rem; font-size: .72rem; }
	.donation-cards dt { color: #718076; }
	.donation-cards dd { margin: 0; overflow-wrap: anywhere; color: #35463a; }
	.campaign-cards h3 { margin: .7rem 0 .2rem; font: 700 1.1rem Outfit, sans-serif; }
	.campaign-cards p { margin: 0; color: #718076; font-size: .76rem; }
	.card-numbers { display: grid; grid-template-columns: 1fr 1fr; gap: .75rem; }
	.card-numbers span { color: #718076; font-size: .7rem; }
	.card-numbers strong { display: block; margin-top: .2rem; color: #263a2d; font-size: .8rem; }
	.pagination { display: flex; align-items: center; justify-content: flex-end; gap: .75rem; margin-top: 1rem; color: #68776c; font-size: .75rem; }
	.pagination button:disabled { cursor: not-allowed; opacity: .45; }
	.bottom-tabs { display: none; }
	@media (max-width: 840px) {
		.app-shell, .app-shell.sidebar-collapsed { display: block; }
		.sidebar { display: none; }
		.main-frame { padding-bottom: calc(5rem + env(safe-area-inset-bottom)); }
		.topbar { min-height: 4.75rem; }
		.bottom-tabs { position: fixed; z-index: 20; right: .75rem; bottom: calc(.75rem + env(safe-area-inset-bottom)); left: .75rem; display: grid; grid-template-columns: repeat(4, 1fr); padding: .4rem; border: 1px solid rgba(255,255,255,.68); border-radius: 1.2rem; background: rgba(250,252,247,.94); box-shadow: 0 .75rem 2rem rgba(16,43,32,.18); backdrop-filter: blur(16px); }
		.bottom-tabs a { display: grid; min-width: 0; min-height: 3.6rem; place-items: center; align-content: center; gap: .2rem; border-radius: .85rem; color: #66776b; font-size: .62rem; font-weight: 800; text-decoration: none; }
		.bottom-tabs a.active { background: #e6efe3; color: #17643c; }
		.bottom-tabs svg { width: 1.25rem; height: 1.25rem; }
		.metrics, .form-layout, .settings-grid { grid-template-columns: 1fr; }
		.donation-filters { grid-template-columns: 1fr 1fr; }
		.preview-panel { position: static; }
	}
	@media (max-width: 640px) { .login-shell { grid-template-columns: 1fr; } .login-intro { padding: 2rem; } .login-intro .eyebrow { margin-top: 3rem; } .page-heading, .section-header, .settings-panel { align-items: flex-start; flex-direction: column; } .primary-action { width: 100%; } .metrics { gap: .45rem; } .metrics article { padding: .9rem 0; } .campaign-row { grid-template-columns: 1fr; gap: .6rem; } .field-grid, .donation-filters { grid-template-columns: 1fr; } .filter-actions { display: grid; grid-template-columns: 1fr 1fr; } .filter-actions .primary-action, .filter-actions .secondary-action { width: 100%; } .table-panel { max-width: 100%; overflow: hidden; } .pagination { justify-content: space-between; } .secondary-fields summary { display: grid; } .settings-title { align-items: flex-start; } }
	@media (prefers-reduced-motion: reduce) { *, *::before, *::after { scroll-behavior: auto !important; animation-duration: .01ms !important; animation-iteration-count: 1 !important; transition-duration: .01ms !important; } }
</style>
