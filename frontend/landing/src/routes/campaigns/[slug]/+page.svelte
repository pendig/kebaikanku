<script>
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	const apiBase = (import.meta.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');
	const presetAmounts = [50000, 100000, 250000, 500000, 1000000, 2500000, 5000000, 10000000];

	let campaign = $state(null);
	let loading = $state(true);
	let error = $state('');
	let submitting = $state(false);
	let donationError = $state('');
	let step = $state(0);
	let form = $state({
		name: '',
		phone_number: '',
		email: '',
		amount: 50000,
		anonymous: false,
		payment_method: 'qris'
	});

	onMount(() => {
		step = stepFromURL();
		loadCampaign();
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
		submitting = true;
		donationError = '';
		try {
			const response = await fetch(`${apiBase}/api/v1/donations`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
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
			const payload = await response.json();
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Could not start payment.');
			const redirectUrl = payload?.data?.payment?.redirect_url;
			if (!redirectUrl) throw new Error('Payment redirect URL is missing.');
			window.location.href = redirectUrl;
		} catch (err) {
			donationError = err.message || 'Could not start payment.';
			submitting = false;
		}
	}

	function nextStep() {
		donationError = '';
		if (step === 1 && form.amount < 2000) {
			donationError = 'Minimal donasi Rp 2.000.';
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
	}

	function progress(value) {
		if (!value?.target_amount) return 0;
		return Math.min(Math.round(((value.collected_amount || 0) / value.target_amount) * 100), 100);
	}

	function formatDate(value) {
		if (!value) return 'Tanpa batas waktu';
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' }).format(new Date(value));
	}
</script>

<svelte:head>
	<title>{campaign?.title || 'Campaign'} | kebaikanku.id</title>
	<meta name="description" content={campaign?.description || 'Detail campaign kebaikanku.id'} />
</svelte:head>

<section class="min-h-screen bg-slate-50 px-6 pb-16 pt-28 text-gray-950 dark:bg-[#0b0f19] dark:text-white">
	<div class="mx-auto max-w-6xl">
		<a href="/campaigns" class="mb-6 inline-flex text-sm font-bold text-emerald-700 dark:text-emerald-300">← Semua campaign</a>

		{#if loading}
			<div class="rounded-3xl border border-gray-200 bg-white p-8 text-sm text-gray-500 dark:border-gray-900 dark:bg-gray-950/40">Loading campaign...</div>
		{:else if error || !campaign}
			<div class="rounded-3xl border border-red-200 bg-red-50 p-8 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-950/20 dark:text-red-300">{error || 'Campaign was not found.'}</div>
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
								<span class="text-gray-500">{campaign.campaign_type === 'open_amount' ? 'unlimited' : `dari ${formatRupiah(campaign.target_amount)}`}</span>
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
							<button type="button" onclick={() => setStep(1)} class="w-full rounded-xl bg-gradient-to-r from-emerald-500 to-yellow-400 px-5 py-3.5 text-sm font-extrabold text-slate-950 shadow-lg shadow-emerald-500/20">Bantu Sekarang</button>
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
										<button type="button" onclick={() => (form.amount = amount)} class="flex items-center justify-between rounded-xl border px-4 py-4 text-left text-lg font-black {form.amount === amount ? 'border-teal-300 bg-teal-400/10 text-teal-300' : 'border-gray-200 bg-white text-teal-600 dark:border-gray-800 dark:bg-gray-900'}">
											<span>{formatRupiah(amount)}</span><span>›</span>
										</button>
									{/each}
								</div>
								<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">Nominal lainnya
									<div class="mt-1 flex rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm dark:border-gray-800 dark:bg-gray-900"><span class="mr-3 font-black text-gray-400">Rp</span><input value={formatPlain(form.amount)} oninput={(event) => parseRupiah(event.currentTarget.value)} inputmode="numeric" class="w-full bg-transparent text-right text-lg font-black text-teal-500 outline-none" /></div>
								</label>
								<p class="text-xs font-semibold text-gray-500">Minimal Rp 2.000 dan maksimal Rp 100.000.000</p>
							{:else if step === 2}
								<label class="flex items-center gap-3 rounded-xl border border-gray-200 p-3 text-sm font-bold dark:border-gray-800"><input bind:checked={form.anonymous} type="checkbox" class="h-4 w-4" /> Tampilkan sebagai Hamba Allah</label>
								{#if !form.anonymous}<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">Nama<input bind:value={form.name} required class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900" /></label>{/if}
								<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">No. HP (optional)<input bind:value={form.phone_number} class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900" /></label>
								<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">Email (optional)<input bind:value={form.email} type="email" class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900" /></label>
								<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">Metode pembayaran<select bind:value={form.payment_method} class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900"><option value="qris">QRIS</option><option value="virtual_account">Virtual Account</option><option value="ewallet">E-Wallet</option><option value="bank_transfer">Bank Transfer</option><option value="retail_outlet">Gerai Retail</option></select></label>
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

							{#if donationError}<p class="rounded-xl bg-red-50 p-3 text-xs font-semibold text-red-700 dark:bg-red-950/20 dark:text-red-300">{donationError}</p>{/if}
							<div class="grid grid-cols-2 gap-3">
								{#if step > 1}<button type="button" onclick={() => setStep(step - 1)} class="rounded-xl border border-gray-200 px-5 py-3.5 text-sm font-extrabold dark:border-gray-800">Kembali</button>{/if}
								<button type="submit" disabled={submitting} class="{step === 1 ? 'col-span-2' : ''} rounded-xl bg-gradient-to-r from-emerald-500 to-yellow-400 px-5 py-3.5 text-sm font-extrabold text-slate-950 shadow-lg shadow-emerald-500/20 disabled:cursor-not-allowed disabled:opacity-60">{submitting ? 'Mengalihkan...' : step < 3 ? 'Selanjutnya' : 'Lanjutkan Pembayaran'}</button>
							</div>
						</form>
					{/if}
				</aside>
			</div>
		{/if}
	</div>
</section>
