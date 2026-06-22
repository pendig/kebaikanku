<script>
	import { onMount } from 'svelte';
	import { t } from '$lib/i18n';

	const apiBase = (import.meta.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');

	let campaigns = $state([]);
	let selectedCategory = $state('all');
	let searchQuery = $state('');
	let activeCampaign = $state(null);
	let loading = $state(true);
	let error = $state('');
	let submitting = $state(false);
	let donationError = $state('');
	let form = $state({
		name: '',
		phone_number: '',
		email: '',
		amount: 100000,
		platform_tip: 0
	});

	let filteredCampaigns = $derived(
		campaigns.filter((campaign) => {
			if (!campaign) return false;
			const category = campaign.category || '';
			const matchesCategory = selectedCategory === 'all' || category === selectedCategory;
			const q = searchQuery.toLowerCase();
			const matchesSearch =
				(campaign.title || '').toLowerCase().includes(q) ||
				(campaign.description || '').toLowerCase().includes(q);
			return matchesCategory && matchesSearch;
		})
	);

	$effect(() => {
		const currentId = activeCampaign?.id;
		const stillVisible = currentId
			? filteredCampaigns.some((campaign) => campaign?.id === currentId)
			: false;
		if (!stillVisible) {
			activeCampaign = filteredCampaigns[0] || null;
		}
	});

	onMount(loadCampaigns);

	async function loadCampaigns() {
		loading = true;
		error = '';
		try {
			const response = await fetch(`${apiBase}/api/v1/campaigns?limit=100`);
			const payload = await response.json();
			if (!response.ok || !payload?.success) {
				throw new Error(payload?.error?.message || 'Could not load campaigns.');
			}
			campaigns = payload?.data?.campaigns || [];
			activeCampaign = campaigns[0] || null;
		} catch (err) {
			error = err.message || 'Could not load campaigns.';
		} finally {
			loading = false;
		}
	}

	async function submitDonation() {
		if (!activeCampaign || submitting) return;

		submitting = true;
		donationError = '';
		try {
			const response = await fetch(`${apiBase}/api/v1/donations`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					campaign_id: activeCampaign.id,
					donor: {
						name: form.name,
						phone_number: form.phone_number,
						email: form.email
					},
					amount: Number(form.amount),
					platform_tip: Number(form.platform_tip || 0),
					payment_method: 'midtrans_snap'
				})
			});
			const payload = await response.json();
			if (!response.ok || !payload?.success) {
				throw new Error(payload?.error?.message || 'Could not start payment.');
			}
			const redirectUrl = payload?.data?.payment?.redirect_url;
			if (!redirectUrl) {
				throw new Error('Payment redirect URL is missing.');
			}
			window.location.href = redirectUrl;
		} catch (err) {
			donationError = err.message || 'Could not start payment.';
			submitting = false;
		}
	}

	function selectCampaign(campaign) {
		activeCampaign = campaign;
		donationError = '';
	}

	function formatRupiah(amount) {
		return new Intl.NumberFormat('id-ID', {
			style: 'currency',
			currency: 'IDR',
			maximumFractionDigits: 0
		}).format(amount || 0);
	}

	function progress(campaign) {
		if (!campaign?.target_amount) return 0;
		return Math.min(Math.round(((campaign.collected_amount || 0) / campaign.target_amount) * 100), 100);
	}
</script>

<svelte:head>
	<title>{$t('campaigns.title')} | kebaikanku.id</title>
	<meta name="description" content={$t('campaigns.subtitle')} />
</svelte:head>

<section class="min-h-screen bg-slate-50 px-6 pb-16 pt-28 text-gray-950 dark:bg-[#0b0f19] dark:text-white">
	<div class="mx-auto grid max-w-7xl gap-8 lg:grid-cols-[1fr_420px]">
		<div class="space-y-6">
			<div class="space-y-3">
				<p class="text-xs font-bold uppercase tracking-[0.2em] text-emerald-600 dark:text-emerald-400">
					{$t('campaigns.badge')}
				</p>
				<h1 class="text-3xl font-extrabold tracking-tight md:text-5xl">{$t('campaigns.title')}</h1>
				<p class="max-w-2xl text-sm leading-6 text-gray-600 dark:text-gray-400">{$t('campaigns.subtitle')}</p>
			</div>

			<div class="flex flex-col gap-3 rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-900 dark:bg-gray-950/40 md:flex-row md:items-center md:justify-between">
				<div class="flex flex-wrap gap-2">
					{#each ['all', 'zakat_maal', 'zakat_fitrah', 'infak', 'kemanusiaan'] as category}
						<button
							type="button"
							onclick={() => (selectedCategory = category)}
							class="rounded-xl px-3 py-2 text-xs font-bold transition {selectedCategory === category ? 'bg-emerald-500 text-white' : 'border border-gray-200 bg-white text-gray-600 hover:border-emerald-300 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300'}"
						>
							{$t(`campaigns.category_${category}`)}
						</button>
					{/each}
				</div>
				<input
					bind:value={searchQuery}
					class="w-full rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900 md:max-w-xs"
					placeholder={$t('campaigns.search_placeholder')}
				/>
			</div>

			{#if loading}
				<div class="rounded-2xl border border-gray-200 bg-white p-8 text-sm text-gray-500 dark:border-gray-900 dark:bg-gray-950/40">{$t('campaigns.loading')}</div>
			{:else if error}
				<div class="rounded-2xl border border-red-200 bg-red-50 p-8 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-950/20 dark:text-red-300">{error}</div>
			{:else if filteredCampaigns.length === 0}
				<div class="rounded-2xl border border-gray-200 bg-white p-8 text-sm text-gray-500 dark:border-gray-900 dark:bg-gray-950/40">{$t('campaigns.empty')}</div>
			{:else}
				<div class="grid gap-4 md:grid-cols-2">
					{#each filteredCampaigns as campaign (campaign.id)}
						<button
							type="button"
							onclick={() => selectCampaign(campaign)}
							class="group rounded-2xl border p-5 text-left transition {activeCampaign?.id === campaign.id ? 'border-emerald-400 bg-emerald-50 shadow-lg shadow-emerald-500/10 dark:bg-emerald-950/20' : 'border-gray-200 bg-white hover:border-emerald-300 dark:border-gray-900 dark:bg-gray-950/40'}"
						>
							<div class="mb-4 flex items-start justify-between gap-3">
								<h2 class="text-lg font-extrabold leading-tight group-hover:text-emerald-600 dark:group-hover:text-emerald-400">{campaign.title}</h2>
								<span class="shrink-0 rounded-full bg-teal-50 px-2.5 py-1 text-[10px] font-bold uppercase text-teal-700 dark:bg-teal-950/40 dark:text-teal-300">{campaign.category}</span>
							</div>
							<p class="line-clamp-3 text-sm leading-6 text-gray-600 dark:text-gray-400">{campaign.description}</p>
							<div class="mt-5 space-y-2">
								<div class="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-900">
									<div class="h-full rounded-full bg-gradient-to-r from-emerald-500 to-teal-400" style="width: {progress(campaign)}%"></div>
								</div>
								<div class="flex justify-between text-xs text-gray-500">
									<span>{formatRupiah(campaign.collected_amount)} {$t('campaigns.collected')}</span>
									<span>{formatRupiah(campaign.target_amount)}</span>
								</div>
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>

		<aside class="h-fit rounded-2xl border border-gray-200 bg-white p-6 shadow-xl shadow-gray-200/60 dark:border-gray-900 dark:bg-gray-950/70 dark:shadow-black/20 lg:sticky lg:top-28">
			{#if activeCampaign}
				<div class="mb-6 space-y-2">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-600 dark:text-emerald-400">{$t('campaigns.donate_now')}</p>
					<h2 class="text-2xl font-extrabold leading-tight">{activeCampaign.title}</h2>
					<p class="text-sm leading-6 text-gray-600 dark:text-gray-400">{activeCampaign.description}</p>
				</div>

				<form class="space-y-4" onsubmit={(event) => { event.preventDefault(); submitDonation(); }}>
					<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">
						{$t('campaigns.form_name')}
						<input bind:value={form.name} required class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900" />
					</label>
					<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">
						{$t('campaigns.form_phone')}
						<input bind:value={form.phone_number} required class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900" />
					</label>
					<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">
						{$t('campaigns.form_email')}
						<input bind:value={form.email} type="email" class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900" />
					</label>
					<div class="grid grid-cols-2 gap-3">
						<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">
							{$t('campaigns.form_amount')}
							<input bind:value={form.amount} required min="1" step="1" type="number" class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900" />
						</label>
						<label class="block text-xs font-bold text-gray-600 dark:text-gray-300">
							{$t('campaigns.form_tip')}
							<input bind:value={form.platform_tip} min="0" step="1" type="number" class="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-gray-800 dark:bg-gray-900" />
						</label>
					</div>
					{#if donationError}
						<p class="rounded-xl bg-red-50 p-3 text-xs font-semibold text-red-700 dark:bg-red-950/20 dark:text-red-300">{donationError}</p>
					{/if}
					<button
						type="submit"
						disabled={submitting}
						class="w-full rounded-xl bg-gradient-to-r from-emerald-500 to-teal-500 px-5 py-3.5 text-sm font-extrabold text-white shadow-lg shadow-emerald-500/20 transition hover:from-emerald-600 hover:to-teal-600 disabled:cursor-not-allowed disabled:opacity-60"
					>
						{submitting ? $t('campaigns.redirecting') : $t('campaigns.pay_with_midtrans')}
					</button>
				</form>
			{:else}
				<p class="text-sm text-gray-500">{$t('campaigns.select_first')}</p>
			{/if}
		</aside>
	</div>
</section>
