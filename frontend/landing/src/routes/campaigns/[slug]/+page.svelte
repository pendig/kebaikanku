<script>
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	const apiBase = (import.meta.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');
	const presetAmounts = [10000, 100000, 1000000];

	let campaign = $state(null);
	let loading = $state(true);
	let error = $state('');
	let submitting = $state(false);
	let donationError = $state('');
	let recoveryRequired = $state(false);
	let checkoutKey = $state('');
	let checkoutURL = $state('');
	let step = $state(0);
	let form = $state({
		name: '',
		phone_number: '',
		email: '',
		amount: 10000,
		anonymous: false
	});

	onMount(() => {
		const saved = sessionStorage.getItem(`donation:${page.params.slug}`);
		if (saved) {
			try {
				const state = JSON.parse(saved);
				form = { ...form, ...state.form };
				checkoutURL = state.checkoutURL || '';
			} catch {
				sessionStorage.removeItem(`donation:${page.params.slug}`);
			}
		}
		step = stepFromURL();
		loadCampaign();
		const resetNavigation = () => (submitting = false);
		window.addEventListener('pageshow', resetNavigation);
		return () => window.removeEventListener('pageshow', resetNavigation);
	});


	function stepFromURL() {
		const value = page.url.searchParams.get('step');
		if (value === 'nominal') return 1;
		if (value === 'contact') return 2;
		if (value === 'confirm') return 3;
		return 0;
	}

	function setStep(next) {
		step = next;
		persistCheckout();
		const names = ['', 'nominal', 'contact', 'confirm'];
		const suffix = next ? `?step=${names[next]}` : '';
		goto(`/campaigns/${page.params.slug}${suffix}`, { keepFocus: true, noScroll: true });
	}

	async function loadCampaign() {
		loading = true;
		error = '';
		try {
			const response = await fetch(`${apiBase}/api/v1/campaigns/${page.params.slug}`);
			const payload = await response.json();
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Could not load campaign.');
			campaign = payload?.data?.campaign || null;
		} catch (err) {
			error = err.message || 'Could not load campaign.';
		} finally {
			loading = false;
		}
	}

	async function submitDonation() {
		if (!campaign || submitting) return;
		if (checkoutURL) {
			window.location.assign(checkoutURL);
			return;
		}
		submitting = true;
		donationError = '';
		recoveryRequired = false;
		checkoutKey ||= crypto.randomUUID();
		try {
			const response = await fetch(`${apiBase}/api/v1/donations`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', 'Idempotency-Key': checkoutKey },
				body: JSON.stringify({
					campaign_id: campaign.id,
					donor: {
						name: form.anonymous ? '' : form.name,
						phone_number: form.phone_number,
						email: form.email
					},
					anonymous: form.anonymous,
					amount: Number(form.amount),
					platform_tip: 0,
					payment_method: 'midtrans_snap'
				})
			});
			const payload = await response.json().catch(() => null);
			if (!response.ok || !payload?.success) {
				if (payload?.error?.code === 'IDEMPOTENCY_PREVIOUSLY_FAILED') {
					donationError = 'Percobaan pembayaran sebelumnya tidak dapat dilanjutkan. Ubah nominal atau data donatur untuk memulai percobaan baru.';
					recoveryRequired = true;
					submitting = false;
					return;
				}
				if (payload?.error?.code === 'VALIDATION_FAILED') {
					donationError = payload.error.message || 'Data donasi belum valid. Periksa kembali nominal dan kontak Anda.';
					recoveryRequired = true;
					submitting = false;
					return;
				}
				throw new Error(payload?.error?.message || '');
			}
			const redirectUrl = payload?.data?.payment?.redirect_url;
			if (!redirectUrl) throw new Error();
			checkoutKey = '';
			checkoutURL = redirectUrl;
			persistCheckout();
			window.location.assign(redirectUrl);
		} catch (error) {
			donationError = error?.message || 'Pembayaran belum dapat dimulai. Coba lagi; nominal dan data Anda tetap dipakai.';
			submitting = false;
		}
	}

	function clearCheckoutState() {
		donationError = '';
		recoveryRequired = false;
		checkoutKey = '';
		checkoutURL = '';
		persistCheckout();
	}

	function persistCheckout() {
		if (typeof sessionStorage === 'undefined') return;
		sessionStorage.setItem(`donation:${page.params.slug}`, JSON.stringify({ form, checkoutURL }));
	}

	function recoverCheckout() {
		clearCheckoutState();
		setStep(1);
	}

	function setAmount(amount) {
		form.amount = amount;
		clearCheckoutState();
	}

	function nextStep() {
		donationError = '';
		if (step === 1 && (form.amount < 2000 || form.amount > 100000000)) {
			donationError = 'Nominal donasi harus antara Rp 2.000 dan Rp 100.000.000.';
			return;
		}
		if (step === 2 && !form.anonymous && !form.name.trim()) {
			donationError = 'Nama wajib diisi, atau pilih anonim.';
			return;
		}
		setStep(step + 1);
	}

	function formatRupiah(amount) {
		return new Intl.NumberFormat('id-ID', {
			style: 'currency',
			currency: 'IDR',
			maximumFractionDigits: 0
		}).format(amount || 0);
	}

	function formatPlain(amount) {
		return new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 }).format(amount || 0);
	}

	function parseRupiah(value) {
		form.amount = Number(String(value).replace(/\D/g, '')) || 0;
		clearCheckoutState();
	}

	function progress(value) {
		if (!value?.target_amount) return 0;
		return Math.min(Math.round(((value.collected_amount || 0) / value.target_amount) * 100), 100);
	}

	function formatDate(value) {
		const date = new Date(value);
		if (!value || Number.isNaN(date.getTime()) || date.getFullYear() <= 1970) return 'Tanpa batas waktu';
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' }).format(date);
	}

	function absoluteURL(value) {
		if (!value) return `${page.url.origin}/images/og-image.png`;
		try { return new URL(value, page.url.origin).href; } catch { return value; }
	}

	let canonicalURL = $derived(`${page.url.origin}${page.url.pathname}`);
	let metaTitle = $derived(campaign?.title ? `${campaign.title} | kebaikanku.id` : 'Kampanye | kebaikanku.id');
	let metaDescription = $derived((campaign?.description || 'Bantu kampanye terpercaya melalui kebaikanku.id').replace(/\s+/g, ' ').trim().slice(0, 160));
	let metaImage = $derived(absoluteURL(campaign?.banner_url));
</script>

<svelte:head>
	<title>{metaTitle}</title>
	<meta name="description" content={metaDescription} />
	<meta name="robots" content="index, follow, max-image-preview:large" />
	<link rel="canonical" href={canonicalURL} />
	<meta property="og:type" content="article" />
	<meta property="og:site_name" content="kebaikanku.id" />
	<meta property="og:url" content={canonicalURL} />
	<meta property="og:title" content={metaTitle} />
	<meta property="og:description" content={metaDescription} />
	<meta property="og:image" content={metaImage} />
	<meta property="og:image:alt" content={campaign?.title ? `Banner ${campaign.title}` : 'kebaikanku.id'} />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={metaTitle} />
	<meta name="twitter:description" content={metaDescription} />
	<meta name="twitter:image" content={metaImage} />
</svelte:head>

<section class="min-h-screen bg-slate-50 px-6 pb-16 pt-20 text-gray-950 dark:bg-[#0b0f19] dark:text-white md:pt-24">
	<div class="mx-auto max-w-6xl">
		{#if loading}
			<div class="rounded-3xl border border-gray-200 bg-white p-8 text-sm text-gray-500 dark:border-gray-900 dark:bg-gray-950/40">Memuat kampanye...</div>
		{:else if error || !campaign}
			<div class="rounded-3xl border border-red-200 bg-red-50 p-8 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-950/20 dark:text-red-300">{error || 'Kampanye tidak ditemukan.'}</div>
		{:else}
			<div class="grid gap-8 lg:grid-cols-[1fr_380px]">
				<article class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-xl shadow-gray-200/50 dark:border-gray-900 dark:bg-gray-950/60 dark:shadow-black/20">
					{#if campaign.banner_url}
						<img src={campaign.banner_url} alt="Banner {campaign.title}" class="h-72 w-full object-cover" />
					{/if}
					<div class="space-y-6 p-6 md:p-8">
						<div class="space-y-3">
							<p class="text-xs font-extrabold uppercase tracking-[0.18em] text-emerald-600 dark:text-emerald-400">{campaign.category}{campaign.subcategory ? ` / ${campaign.subcategory}` : ''}</p>
							<h1 class="text-3xl font-black tracking-tight md:text-5xl">{campaign.title}</h1>
							{#if campaign.location}<p class="text-sm font-bold text-gray-500 dark:text-gray-400">{campaign.location}</p>{/if}
						</div>

						<div class="space-y-3 rounded-2xl bg-slate-50 p-5 dark:bg-gray-900/70">
							<div class="flex flex-wrap justify-between gap-3 text-sm">
								<strong>{formatRupiah(campaign.collected_amount)} terkumpul</strong>
								<span class="text-gray-500">{campaign.campaign_type === 'open_amount' ? 'Tanpa batas' : `dari ${formatRupiah(campaign.target_amount)}`}</span>
							</div>
							<div class="h-3 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-800"><div class="h-full rounded-full bg-gradient-to-r from-emerald-500 to-teal-400" style="width: {campaign.campaign_type === 'open_amount' ? 100 : progress(campaign)}%"></div></div>
							<p class="text-xs font-semibold text-gray-500">Berakhir: {formatDate(campaign.end_date)}</p>
						</div>

						{#if campaign.beneficiary_note}
							<div class="rounded-2xl border border-amber-200 bg-amber-50 p-5 text-sm font-semibold text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/20 dark:text-amber-100">{campaign.beneficiary_note}</div>
						{/if}

						<div class="prose prose-slate max-w-none whitespace-pre-line text-gray-700 dark:prose-invert dark:text-gray-300">{campaign.description}</div>
					</div>
				</article>

				<aside class="h-fit rounded-3xl border border-gray-200 bg-white p-6 shadow-xl shadow-gray-200/60 dark:border-gray-900 dark:bg-gray-950/70 dark:shadow-black/20 lg:sticky lg:top-28">
					{#if step === 0}
						<div class="space-y-4">
							<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-600 dark:text-emerald-400">Bantu sekarang</p>
							<h2 class="text-2xl font-extrabold leading-tight">Siap ikut bantu?</h2>
							<p class="text-sm leading-6 text-gray-600 dark:text-gray-400">Pilih nominal donasi di langkah berikutnya, lalu isi kontak singkat dan konfirmasi.</p>
							<button type="button" onclick={() => setStep(1)} class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-emerald-500 to-yellow-400 px-5 py-3.5 text-sm font-extrabold text-slate-950 shadow-lg shadow-emerald-500/20">Bantu Sekarang <span aria-hidden="true">→</span></button>
							<div class="border-t border-gray-200 pt-4 text-xs leading-5 text-gray-500 dark:border-gray-800 dark:text-gray-400">
								Pembayaran dilanjutkan ke halaman aman Midtrans. Metode pembayaran dan status transaksi dikonfirmasi oleh Midtrans.
							</div>
						</div>
					{:else}
						<div class="mb-6 space-y-2">
							<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-600 dark:text-emerald-400">Langkah {step} dari 3</p>
							<h2 class="text-2xl font-extrabold leading-tight">{step === 1 ? 'Pilih nominal donasi' : step === 2 ? 'Kontak dan metode bayar' : 'Konfirmasi donasi'}</h2>
						</div>

						<form class="space-y-4" onsubmit={(event) => { event.preventDefault(); step < 3 ? nextStep() : submitDonation(); }}>
							{#if step === 1}
								<div class="grid gap-3 sm:grid-cols-2">
									{#each presetAmounts as amount}
										<button type="button" onclick={() => setAmount(amount)} disabled={submitting} class="flex items-center justify-between rounded-xl border px-4 py-4 text-left text-lg font-black {form.amount === amount ? 'border-teal-300 bg-teal-400/10 text-teal-300' : 'border-gray-200 bg-white text-teal-600 dark:border-gray-800 dark:bg-gray-900'} disabled:cursor-not-allowed disabled:opacity-60">
											<span>{formatRupiah(amount)}</span><span>›</span>
										</button>
									{/each}
								</div>
								<label for="donation-amount" class="block text-xs font-bold text-gray-600 dark:text-gray-300">Nominal lainnya
									<div class="mt-1 flex rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm dark:border-gray-800 dark:bg-gray-900"><span class="mr-3 font-black text-gray-400">Rp</span><input id="donation-amount" value={formatPlain(form.amount)} oninput={(event) => parseRupiah(event.currentTarget.value)} inputmode="numeric" disabled={submitting} aria-invalid={Boolean(donationError)} class="w-full bg-transparent text-right text-lg font-black text-teal-500 outline-none disabled:cursor-not-allowed" /></div>
								</label>
								<p class="text-xs font-semibold text-gray-500">Minimal Rp 2.000 dan maksimal Rp 100.000.000</p>
							{:else if step === 2}
								<label class="flex items-center gap-3 rounded-xl border border-gray-200 p-3 text-sm font-bold dark:border-gray-800"><input bind:checked={form.anonymous} onchange={clearCheckoutState} type="checkbox" disabled={submitting} class="h-4 w-4" /> Tampilkan sebagai Hamba Allah</label>
								{#if !form.anonymous}<label for="donor-name" class="block text-xs font-bold text-gray-600 dark:text-gray-300">Nama<input id="donor-name" bind:value={form.name} oninput={clearCheckoutState} required disabled={submitting} aria-invalid={Boolean(donationError)} class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 disabled:cursor-not-allowed dark:border-gray-800 dark:bg-gray-900" /></label>{/if}
								<label for="donor-phone" class="block text-xs font-bold text-gray-600 dark:text-gray-300">No. HP (opsional)<input id="donor-phone" bind:value={form.phone_number} oninput={clearCheckoutState} disabled={submitting} aria-invalid={Boolean(donationError)} class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 disabled:cursor-not-allowed dark:border-gray-800 dark:bg-gray-900" /></label>
								<label for="donor-email" class="block text-xs font-bold text-gray-600 dark:text-gray-300">Email (opsional)<input id="donor-email" bind:value={form.email} oninput={clearCheckoutState} type="email" disabled={submitting} aria-invalid={Boolean(donationError)} class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 disabled:cursor-not-allowed dark:border-gray-800 dark:bg-gray-900" /></label>
								<p class="text-xs font-semibold text-gray-500">Metode pembayaran dipilih di halaman aman Midtrans.</p>
							{:else}
								<div class="space-y-3 rounded-2xl bg-slate-50 p-4 text-sm dark:bg-gray-900/70">
									<div class="flex justify-between gap-4"><span>Nama Donatur</span><strong>{form.anonymous ? 'Hamba Allah' : form.name}</strong></div>
									<div class="flex justify-between gap-4"><span>No. HP</span><strong>{form.phone_number || '-'}</strong></div>
									<div class="flex justify-between gap-4"><span>Email</span><strong>{form.email || '-'}</strong></div>
									<div class="border-t border-gray-200 pt-3 dark:border-gray-800"></div>
									<div class="flex justify-between gap-4"><span>Nilai Donasi</span><strong>{formatRupiah(form.amount)}</strong></div>
									<div class="flex justify-between gap-4 text-lg"><span>Total Bayar</span><strong>{formatRupiah(form.amount)}</strong></div>
								</div>
							{/if}

							{#if donationError}
								<div role="alert" class="space-y-3 rounded-xl bg-red-50 p-3 text-xs font-semibold text-red-700 dark:bg-red-950/20 dark:text-red-300">
									<p>{donationError}</p>
									{#if recoveryRequired}<button type="button" onclick={recoverCheckout} class="inline-flex min-h-11 items-center gap-2 rounded-lg bg-red-700 px-4 py-2 font-extrabold text-white hover:bg-red-800 dark:bg-red-300 dark:text-red-950"><span aria-hidden="true">✎</span> Ubah data donasi</button>{/if}
								</div>
							{/if}
							<div class="grid grid-cols-2 gap-3">
								{#if step > 1}<button type="button" onclick={() => setStep(step - 1)} disabled={submitting} class="inline-flex items-center justify-center gap-2 rounded-xl border border-gray-200 px-5 py-3.5 text-sm font-extrabold disabled:cursor-not-allowed disabled:opacity-60 dark:border-gray-800"><span aria-hidden="true">←</span> Kembali</button>{/if}
								<button type="submit" disabled={submitting || recoveryRequired} class="{step === 1 ? 'col-span-2' : ''} inline-flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-emerald-500 to-yellow-400 px-5 py-3.5 text-sm font-extrabold text-slate-950 shadow-lg shadow-emerald-500/20 disabled:cursor-not-allowed disabled:opacity-60">{submitting ? 'Mengalihkan...' : step < 3 ? 'Selanjutnya' : 'Lanjutkan Pembayaran'} <span aria-hidden="true">→</span></button>
							</div>
						</form>
					{/if}
				</aside>
			</div>
		{/if}
	</div>
</section>
