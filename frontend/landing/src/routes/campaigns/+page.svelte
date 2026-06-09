<script>
	import { t } from '$lib/i18n';

	// Category filter state
	let selectedCategory = $state('all');
	// Search keyword state
	let searchQuery = $state('');

	// Donation modal state
	let isModalOpen = $state(false);
	let activeCampaign = $state(null);

	const mockCampaigns = [
		{
			id: '1',
			title: 'Bantu Renovasi Madrasah Al-Ikhlas',
			description: 'Renovasi bangunan sekolah madrasah di pelosok Jawa Barat yang sudah lapuk dimakan usia demi kenyamanan belajar anak-anak.',
			category: 'zakat',
			targetAmount: 120000000,
			collectedAmount: 75000000,
			daysLeft: 24,
			institution: 'Yayasan Bina Umat'
		},
		{
			id: '2',
			title: 'Solidaritas Pangan & Medis untuk Palestina',
			description: 'Penyaluran paket makanan pokok, obat-obatan darurat, dan akses air bersih untuk warga terdampak di posko pengungsian.',
			category: 'kemanusiaan',
			targetAmount: 500000000,
			collectedAmount: 320500000,
			daysLeft: 12,
			institution: 'Aksi Tanggap Mulia'
		},
		{
			id: '3',
			title: 'Program Ambulans Gratis Yayasan Peduli',
			description: 'Pengadaan armada mobil ambulans layanan operasional darurat gratis bagi kaum dhuafa, lansia, dan warga miskin kota.',
			category: 'infak',
			targetAmount: 250000000,
			collectedAmount: 98200000,
			daysLeft: 45,
			institution: 'Yayasan Peduli Nusantara'
		},
		{
			id: '4',
			title: 'Zakat Maal Pemberdayaan UMKM Syariah',
			description: 'Penyaluran modal usaha bergulir tanpa bunga serta pembinaan kewirausahaan syariah untuk para asnaf mustahik produktif.',
			category: 'zakat',
			targetAmount: 80000000,
			collectedAmount: 64000000,
			daysLeft: 18,
			institution: 'Baitul Maal Amanah'
		},
		{
			id: '5',
			title: 'Bantuan Logistik Bencana Banjir Bandang',
			description: 'Penyediaan posko darurat hangat, selimut, pakaian bersih layak pakai, dan makanan siap saji untuk pengungsi terdampak banjir.',
			category: 'kemanusiaan',
			targetAmount: 150000000,
			collectedAmount: 112900000,
			daysLeft: 8,
			institution: 'Relawan Hijau Indonesia'
		}
	];

	// Filter and search logic
	let filteredCampaigns = $derived(
		mockCampaigns.filter((camp) => {
			const matchesCategory = selectedCategory === 'all' || camp.category === selectedCategory;
			const matchesSearch =
				camp.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
				camp.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
				camp.institution.toLowerCase().includes(searchQuery.toLowerCase());
			return matchesCategory && matchesSearch;
		})
	);

	function formatRupiah(amount) {
		return new Intl.NumberFormat('id-ID', {
			style: 'currency',
			currency: 'IDR',
			maximumFractionDigits: 0
		}).format(amount);
	}

	function calculatePercentage(collected, target) {
		return Math.min(Math.round((collected / target) * 100), 100);
	}

	function openDonation(campaign) {
		activeCampaign = campaign;
		isModalOpen = true;
	}

	function closeModal() {
		isModalOpen = false;
		activeCampaign = null;
	}
</script>

<svelte:head>
	<title>{$t('campaigns.title')} | kebaikanku.id</title>
	<meta name="description" content={$t('campaigns.subtitle')} />
</svelte:head>

<section class="relative py-12 md:py-20 px-6 overflow-hidden">
	<!-- Ambient lights -->
	<div class="absolute inset-0 bg-slate-50 dark:bg-[#0b0f19] z-0 transition-colors duration-200">
		<div class="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[300px] bg-emerald-500/5 blur-[120px] rounded-full"></div>
	</div>

	<div class="max-w-7xl mx-auto relative z-10 space-y-12">
		<!-- Header -->
		<div class="text-center space-y-4 max-w-3xl mx-auto">
			<h1 class="text-3xl md:text-5xl font-extrabold tracking-tight text-gray-900 dark:text-white leading-tight">
				{$t('campaigns.title')}
			</h1>
			<p class="text-gray-600 dark:text-gray-400 text-sm md:text-base leading-relaxed">
				{$t('campaigns.subtitle')}
			</p>
		</div>

		<!-- Filters & Search Controls -->
		<div class="flex flex-col md:flex-row items-center justify-between gap-6 bg-white/40 dark:bg-gray-950/40 p-4 rounded-2xl border border-gray-200 dark:border-gray-900 shadow-lg dark:shadow-xl backdrop-blur-md">
			<!-- Categories tabs -->
			<div class="flex flex-wrap items-center gap-2 w-full md:w-auto">
				<button
					onclick={() => (selectedCategory = 'all')}
					class="px-4 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer {selectedCategory === 'all'
						? 'bg-emerald-500 text-white'
						: 'bg-white/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:text-gray-950 dark:hover:text-white hover:border-gray-300 dark:hover:border-gray-700'}"
				>
					{$t('campaigns.filter_all')}
				</button>
				<button
					onclick={() => (selectedCategory = 'zakat')}
					class="px-4 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer {selectedCategory === 'zakat'
						? 'bg-emerald-500 text-white'
						: 'bg-white/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:text-gray-950 dark:hover:text-white hover:border-gray-300 dark:hover:border-gray-700'}"
				>
					{$t('campaigns.filter_zakat')}
				</button>
				<button
					onclick={() => (selectedCategory = 'kemanusiaan')}
					class="px-4 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer {selectedCategory === 'kemanusiaan'
						? 'bg-emerald-500 text-white'
						: 'bg-white/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:text-gray-950 dark:hover:text-white hover:border-gray-300 dark:hover:border-gray-700'}"
				>
					{$t('campaigns.filter_kemanusiaan')}
				</button>
				<button
					onclick={() => (selectedCategory = 'infak')}
					class="px-4 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer {selectedCategory === 'infak'
						? 'bg-emerald-500 text-white'
						: 'bg-white/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:text-gray-950 dark:hover:text-white hover:border-gray-300 dark:hover:border-gray-700'}"
				>
					{$t('campaigns.filter_infak')}
				</button>
			</div>

			<!-- Search -->
			<div class="relative w-full md:max-w-xs">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder={$t('campaigns.search_placeholder')}
					class="w-full pl-10 pr-4 py-2.5 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900/40 text-gray-950 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500/20 text-xs shadow-sm dark:shadow-none"
				/>
				<div class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
				</div>
			</div>
		</div>

		<!-- Campaigns Grid -->
		{#if filteredCampaigns.length > 0}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
				{#each filteredCampaigns as campaign (campaign.id)}
					<div class="bg-white dark:bg-gray-950/20 border border-gray-200 dark:border-gray-900 rounded-3xl p-6 flex flex-col justify-between space-y-6 hover:border-emerald-500/30 hover:bg-slate-50 dark:hover:bg-gray-950/40 transition-all duration-300 relative group overflow-hidden shadow-sm dark:shadow-none">
						<!-- Top border shimmer -->
						<div class="absolute inset-x-0 top-0 h-[1px] bg-gradient-to-r from-transparent via-emerald-500/20 dark:via-emerald-500/10 to-transparent"></div>
						
						<div class="space-y-4">
							<!-- Institution & Badge -->
							<div class="flex items-center justify-between text-[11px] text-gray-500 dark:text-gray-400">
								<span>{campaign.institution}</span>
								<span class="px-2.5 py-0.5 rounded-full border border-teal-200 dark:border-teal-500/20 bg-teal-50 dark:bg-teal-500/5 text-teal-650 dark:text-teal-400 font-semibold uppercase">
									{campaign.category}
								</span>
							</div>

							<!-- Title -->
							<h3 class="text-lg font-bold text-gray-900 dark:text-white leading-tight group-hover:text-emerald-500 dark:group-hover:text-emerald-400 transition-colors duration-300">
								{campaign.title}
							</h3>
							
							<!-- Description -->
							<p class="text-gray-650 dark:text-gray-400 text-xs leading-relaxed line-clamp-3">
								{campaign.description}
							</p>
						</div>

						<div class="space-y-4 pt-2">
							<!-- Progress bar -->
							<div class="space-y-1.5">
								<div class="w-full bg-gray-100 dark:bg-gray-900 h-2 rounded-full overflow-hidden">
									<div
										class="bg-gradient-to-r from-emerald-500 to-teal-400 h-2 rounded-full transition-all duration-500"
										style="width: {calculatePercentage(campaign.collectedAmount, campaign.targetAmount)}%"
									></div>
								</div>
								<div class="flex justify-between items-center text-[10px] text-gray-500 dark:text-gray-450">
									<span>{calculatePercentage(campaign.collectedAmount, campaign.targetAmount)}% {$t('campaigns.collected')}</span>
									<span>{campaign.daysLeft} {$t('campaigns.days_left')}</span>
								</div>
							</div>

							<!-- Amounts -->
							<div class="flex items-baseline justify-between pt-1">
								<div class="flex flex-col">
									<span class="text-[9px] uppercase tracking-wider text-gray-500">{$t('campaigns.collected')}</span>
									<span class="text-sm font-extrabold text-gray-900 dark:text-white">{formatRupiah(campaign.collectedAmount)}</span>
								</div>
								<div class="flex flex-col text-right">
									<span class="text-[9px] uppercase tracking-wider text-gray-500">{$t('campaigns.target')}</span>
									<span class="text-xs font-semibold text-gray-600 dark:text-gray-400">{formatRupiah(campaign.targetAmount)}</span>
								</div>
							</div>

							<!-- CTA Button -->
							<button
								onclick={() => openDonation(campaign)}
								class="w-full bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-white font-bold py-3 rounded-xl transition-all duration-300 text-xs shadow-md shadow-emerald-500/5 hover:shadow-emerald-500/15 cursor-pointer"
							>
								{$t('campaigns.donate_now')}
							</button>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<!-- Empty State -->
			<div class="text-center py-20 bg-white dark:bg-gray-950/20 rounded-3xl border border-gray-200 dark:border-gray-900 shadow-sm dark:shadow-none">
				<div class="w-12 h-12 rounded-full bg-gray-50 dark:bg-gray-900 border border-gray-100 dark:border-gray-800 flex items-center justify-center mx-auto text-gray-500 mb-4">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<h3 class="text-gray-900 dark:text-white font-bold mb-1">Kampanye tidak ditemukan / Campaign not found</h3>
				<p class="text-xs text-gray-600 dark:text-gray-500">Coba ubah filter atau kata kunci pencarian Anda.</p>
			</div>
		{/if}
	</div>
</section>

<!-- Simulated Payment / Integration Coming Soon Modal -->
{#if isModalOpen && activeCampaign}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-6 bg-black/50 dark:bg-black/75 backdrop-blur-md transition-opacity">
		<div class="relative w-full max-w-md bg-white dark:bg-gray-950 border border-gray-200 dark:border-gray-800 rounded-3xl p-6 shadow-2xl overflow-hidden animate-scale-up">
			<!-- Glowing decorative edge -->
			<div class="absolute inset-x-0 top-0 h-[1px] bg-gradient-to-r from-transparent via-emerald-500/30 to-transparent"></div>
			
			<div class="space-y-6 text-center">
				<!-- Icon -->
				<div class="w-14 h-14 bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-200 dark:border-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded-full flex items-center justify-center mx-auto shadow-md dark:shadow-lg dark:shadow-emerald-500/5">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
					</svg>
				</div>

				<div class="space-y-2">
					<h3 class="text-xl font-bold text-gray-900 dark:text-white">Integrasi Pembayaran Sedang Diproses</h3>
					<span class="text-xs font-semibold text-emerald-650 dark:text-emerald-400 px-2 py-0.5 rounded bg-emerald-50 dark:bg-emerald-500/5 border border-emerald-200 dark:border-emerald-500/10 inline-block uppercase">
						Midtrans Snap API
					</span>
					<p class="text-xs text-gray-650 dark:text-gray-400 leading-relaxed pt-2">
						Sistem donasi langsung dan payment gateway Midtrans untuk kampanye <strong class="text-gray-900 dark:text-white">"{activeCampaign.title}"</strong> saat ini sedang dibangun dalam roadmap MVP.
					</p>
				</div>

				<!-- Actions -->
				<div class="flex flex-col gap-3 pt-2">
					<a
						href="/coming-soon"
						onclick={closeModal}
						class="w-full bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-white font-bold py-3 rounded-xl transition-all duration-300 text-xs shadow-md shadow-emerald-500/5 hover:-translate-y-0.5 text-center"
					>
						Ikut Antrean Rilis (Waitlist)
					</a>
					<button
						onclick={closeModal}
						class="w-full bg-gray-100 hover:bg-gray-200 dark:bg-gray-900 dark:hover:bg-gray-800 border border-gray-200 dark:border-gray-800 text-gray-800 dark:text-white font-semibold py-3 rounded-xl transition-all duration-300 text-xs cursor-pointer"
					>
						Kembali ke Daftar Kampanye
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
