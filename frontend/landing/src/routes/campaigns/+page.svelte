<script>
	import { onMount } from 'svelte';
	import { t } from '$lib/i18n';

	const apiBase = (import.meta.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');

	let campaigns = $state([]);
	let selectedCategory = $state('all');
	let searchQuery = $state('');
	let loading = $state(true);
	let error = $state('');

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
		} catch (err) {
			error = err.message || 'Could not load campaigns.';
		} finally {
			loading = false;
		}
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
	<div class="mx-auto max-w-7xl">
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
					{#each [{id:'all', label:'Semua'}, {id:'zakat', label:'Zakat'}, {id:'infaq', label:'Infaq'}, {id:'qurban', label:'Qurban'}, {id:'donasi_lainnya', label:'Donasi Lainnya'}] as category}
						<button
							type="button"
							onclick={() => (selectedCategory = category.id)}
							class="rounded-xl px-3 py-2 text-xs font-bold transition {selectedCategory === category.id ? 'bg-emerald-500 text-white' : 'border border-gray-200 bg-white text-gray-600 hover:border-emerald-300 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300'}"
						>
							{category.label}
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
						<article class="group overflow-hidden rounded-2xl border transition border-gray-200 bg-white hover:border-emerald-300 dark:border-gray-900 dark:bg-gray-950/40">
							<div class="w-full p-5 text-left">
								{#if campaign.banner_url}
									<img src={campaign.banner_url} alt="Banner {campaign.title}" class="-mx-5 -mt-5 mb-5 h-40 w-[calc(100%+2.5rem)] object-cover" loading="lazy" />
								{/if}
								<div class="mb-4 flex items-start justify-between gap-3">
									<h2 class="text-lg font-extrabold leading-tight group-hover:text-emerald-600 dark:group-hover:text-emerald-400">{campaign.title}</h2>
									<span class="shrink-0 rounded-full bg-teal-50 px-2.5 py-1 text-[10px] font-bold uppercase text-teal-700 dark:bg-teal-950/40 dark:text-teal-300">{campaign.category}{campaign.subcategory ? ` / ${campaign.subcategory}` : ''}</span>
								</div>
								{#if campaign.location}
									<p class="mb-2 text-xs font-bold uppercase tracking-[0.16em] text-emerald-700 dark:text-emerald-300">{campaign.location}</p>
								{/if}
								<p class="line-clamp-3 text-sm leading-6 text-gray-600 dark:text-gray-400">{campaign.description}</p>
								{#if campaign.beneficiary_note}
									<p class="mt-3 rounded-xl bg-amber-50 px-3 py-2 text-xs font-semibold text-amber-800 dark:bg-amber-950/30 dark:text-amber-200">{campaign.beneficiary_note}</p>
								{/if}
								<div class="mt-5 space-y-2">
									<div class="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-900">
										<div class="h-full rounded-full bg-gradient-to-r from-emerald-500 to-teal-400" style="width: {progress(campaign)}%"></div>
									</div>
									<div class="flex justify-between text-xs text-gray-500">
										<span>{formatRupiah(campaign.collected_amount)} {$t('campaigns.collected')}</span>
										<span>{campaign.campaign_type === 'open_amount' ? 'Unlimited' : formatRupiah(campaign.target_amount)}</span>
									</div>
								</div>
							</div>
							<a href="/campaigns/{campaign.slug}" class="mx-5 mb-5 inline-flex rounded-xl border border-emerald-200 px-4 py-2 text-xs font-extrabold text-emerald-700 hover:bg-emerald-50 dark:border-emerald-900 dark:text-emerald-300 dark:hover:bg-emerald-950/30">Lihat detail</a>
						</article>
					{/each}
				</div>
			{/if}
		</div>

	</div>
</section>
