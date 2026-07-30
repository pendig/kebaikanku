<script>
	import { t } from '$lib/i18n';
	
	// FAQ accordion state
	let openFaq = $state(-1);
	
	function toggleFaq(index) {
		openFaq = openFaq === index ? -1 : index;
	}

	function handleMouseMove(e) {
		const card = e.currentTarget;
		const rect = card.getBoundingClientRect();
		const x = e.clientX - rect.left;
		const y = e.clientY - rect.top;
		card.style.setProperty('--x', `${x}px`);
		card.style.setProperty('--y', `${y}px`);
	}
</script>


<!-- Hero Section -->
<section class="relative pt-12 pb-24 md:py-32 overflow-hidden">
	<!-- Background grid and gradients -->
	<div class="absolute inset-0 bg-slate-50 dark:bg-[#0b0f19] z-0 transition-colors duration-200">
		<div class="absolute inset-0 bg-[linear-gradient(to_right,#1f29370a_1px,transparent_1px),linear-gradient(to_bottom,#1f29370a_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)]"></div>
		<div class="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[300px] bg-emerald-500/10 blur-[120px] rounded-full"></div>
		<div class="absolute top-40 left-1/3 w-[300px] h-[300px] bg-teal-500/5 blur-[100px] rounded-full"></div>
	</div>

	<div class="max-w-7xl mx-auto px-6 relative z-10 text-center space-y-8">
		<!-- Title -->
		<h1 class="text-4xl md:text-6xl lg:text-7xl font-extrabold tracking-tight max-w-5xl mx-auto leading-[1.15] text-gray-900 dark:text-white">
			{$t('hero.title').split('&')[0]} <span class="bg-gradient-to-r from-emerald-500 via-teal-400 to-emerald-500 dark:from-emerald-400 dark:via-teal-300 dark:to-emerald-400 bg-clip-text text-transparent">{$t('hero.title').split('&')[1] || ''}</span>
		</h1>

		<!-- Subtitle -->
		<p class="text-gray-600 dark:text-gray-400 text-lg md:text-xl max-w-3xl mx-auto leading-relaxed">
			{$t('hero.subtitle')}
		</p>

		<!-- CTAs -->
		<div class="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
			<a
				href="/campaigns"
				class="w-full sm:w-auto bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-white font-bold px-8 py-4 rounded-xl shadow-lg shadow-emerald-500/20 hover:shadow-emerald-500/30 transition-all duration-300 hover:-translate-y-0.5 text-center cursor-pointer"
			>
				{$t('hero.cta_primary')}
			</a>
			<a
				href="https://github.com/pendig/kebaikanku"
				target="_blank"
				rel="noopener"
				class="w-full sm:w-auto border border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-gray-900/50 hover:bg-gray-50 dark:hover:bg-gray-800/80 hover:border-gray-300 dark:hover:border-gray-700 text-gray-800 dark:text-white font-semibold px-8 py-4 rounded-xl transition-all duration-300 text-center cursor-pointer"
			>
				{$t('hero.cta_secondary')}
			</a>
		</div>

		<!-- Mockup/Analytics Showcase -->
		<div class="pt-16 max-w-5xl mx-auto">
			<div class="relative rounded-2xl border border-gray-200 dark:border-gray-800 bg-white/40 dark:bg-gray-950/40 p-4 shadow-xl dark:shadow-2xl shadow-black/5 dark:shadow-black/50 backdrop-blur-sm">
				<div class="absolute inset-0 bg-gradient-to-t from-slate-50 dark:from-[#0b0f19] via-transparent to-transparent z-10 rounded-2xl"></div>
				<!-- Simulated Dashboard Window -->
				<div class="rounded-xl overflow-hidden border border-gray-200 dark:border-gray-800/80 bg-slate-100 dark:bg-[#0d1321] text-left">
					<!-- Window bar -->
					<div class="bg-slate-200 dark:bg-[#090e1a] px-4 py-3 border-b border-gray-200 dark:border-gray-900 flex items-center justify-between">
						<div class="flex items-center space-x-2">
							<div class="w-3 h-3 rounded-full bg-red-500/70"></div>
							<div class="w-3 h-3 rounded-full bg-yellow-500/70"></div>
							<div class="w-3 h-3 rounded-full bg-green-500/70"></div>
						</div>
						<div class="text-xs text-gray-600 dark:text-gray-500 font-mono">app.kebaikanku.id/campaigns</div>
						<div class="w-12"></div>
					</div>
					<!-- Window body placeholder/design -->
					<div class="p-6 md:p-8 grid grid-cols-1 md:grid-cols-3 gap-6">
						<div class="bg-white dark:bg-gray-900/40 rounded-xl p-5 border border-gray-200 dark:border-gray-800/50 shadow-sm dark:shadow-none">
							<span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase">Zakat Maal Campaign</span>
							<h3 class="text-lg font-bold text-gray-900 dark:text-white mt-1">Bantu Renovasi Madrasah</h3>
							<div class="w-full bg-gray-200 dark:bg-gray-800 h-2 rounded-full mt-4 overflow-hidden">
								<div class="bg-emerald-500 h-2 rounded-full w-3/4"></div>
							</div>
							<div class="flex justify-between items-center mt-3 text-xs">
								<span class="text-gray-600 dark:text-gray-400">75% Collected</span>
								<span class="text-emerald-600 dark:text-emerald-400 font-bold">Rp 75,000,000</span>
							</div>
						</div>
						<div class="bg-white dark:bg-gray-900/40 rounded-xl p-5 border border-gray-200 dark:border-gray-800/50 shadow-sm dark:shadow-none">
							<span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase">Biaya platform</span>
							<h3 class="text-lg font-bold text-emerald-600 dark:text-emerald-400 mt-1">0% untuk self-hosting</h3>
							<p class="text-xs text-gray-600 dark:text-gray-500 mt-2">Deployment mandiri tidak mengenakan biaya platform wajib. Biaya transaksi payment gateway dapat berlaku.</p>
							<div class="mt-4 inline-flex items-center text-xs text-emerald-600 dark:text-emerald-400 font-semibold bg-emerald-50 dark:bg-emerald-500/5 border border-emerald-200 dark:border-emerald-500/10 px-2.5 py-1 rounded-full">
								Deployment mandiri
							</div>
						</div>
						<div class="bg-white dark:bg-gray-900/40 rounded-xl p-5 border border-gray-200 dark:border-gray-800/50 shadow-sm dark:shadow-none">
							<span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase">Fitur tersedia</span>
							<h3 class="text-lg font-bold text-gray-900 dark:text-white mt-1">Kampanye dan checkout</h3>
							<p class="text-xs text-gray-600 dark:text-gray-500 mt-2">Laporan AI dan dashboard multi-pengguna masih berada di roadmap.</p>
							<div class="mt-4 flex -space-x-2 overflow-hidden">
								<div class="inline-block h-6 w-6 rounded-full ring-2 ring-white dark:ring-gray-900 bg-emerald-500 flex items-center justify-center text-[10px] font-bold text-white">WA</div>
								<div class="inline-block h-6 w-6 rounded-full ring-2 ring-white dark:ring-gray-900 bg-teal-600 flex items-center justify-center text-[10px] font-bold text-white">AI</div>
								<div class="inline-block h-6 w-6 rounded-full ring-2 ring-white dark:ring-gray-900 bg-gray-700 flex items-center justify-center text-[10px] font-bold text-white">QR</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</section>

<!-- Stats / Comparison Section -->
<section id="comparison" class="py-20 bg-white dark:bg-[#090c15] border-y border-gray-200 dark:border-gray-950 transition-colors duration-200">
	<div class="max-w-7xl mx-auto px-6">
		<div class="text-center space-y-4 mb-16">
			<h2 class="text-3xl md:text-5xl font-bold text-gray-900 dark:text-white">{$t('stats.header')}</h2>
			<p class="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">{$t('stats.subheader')}</p>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
			<!-- Public donation flow -->
			<div class="bg-gray-100/30 dark:bg-gray-950/30 border border-gray-200 dark:border-gray-900 rounded-2xl p-8 space-y-6 relative overflow-hidden group glow-card" onmousemove={handleMouseMove} role="presentation">
				<div class="absolute top-0 right-0 w-24 h-24 bg-red-500/5 blur-xl rounded-full"></div>
				<div class="w-12 h-12 rounded-xl bg-red-50 dark:bg-red-500/5 border border-red-200 dark:border-red-500/10 flex items-center justify-center text-red-600 dark:text-red-400 group-hover:scale-110 transition-transform duration-300">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 animate-pulse-slow" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<h3 class="text-xl font-bold text-gray-900 dark:text-white">{$t('stats.traditional_title')}</h3>
				<div class="text-3xl font-extrabold text-red-600 dark:text-red-400">{$t('stats.traditional_fee')}</div>
				<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed">{$t('stats.traditional_desc')}</p>
			</div>

			<!-- Our Self-Hosted Model -->
			<div class="bg-gradient-to-b from-emerald-500/5 to-transparent border border-emerald-200 dark:border-emerald-500/20 rounded-2xl p-8 space-y-6 relative overflow-hidden shadow-sm dark:shadow-lg dark:shadow-emerald-500/5 glow-card" onmousemove={handleMouseMove} role="presentation">
				<div class="absolute top-0 right-0 w-24 h-24 bg-emerald-500/10 blur-xl rounded-full"></div>
				<div class="w-12 h-12 rounded-xl bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-200 dark:border-emerald-500/20 flex items-center justify-center text-emerald-600 dark:text-emerald-400 group-hover:scale-110 transition-transform duration-300">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 animate-float" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
					</svg>
				</div>
				<h3 class="text-xl font-bold text-gray-900 dark:text-white">{$t('stats.our_title')}</h3>
				<div class="text-3xl font-extrabold text-emerald-600 dark:text-emerald-400">{$t('stats.our_fee')}</div>
				<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed">{$t('stats.our_desc')}</p>
			</div>

		</div>
	</div>
</section>

<!-- Features Grid -->
<section id="features" class="py-24 max-w-7xl mx-auto px-6">
	<div class="text-center space-y-4 mb-20">
		<span class="text-xs font-bold tracking-wider text-emerald-600 dark:text-emerald-400 uppercase bg-emerald-50 dark:bg-emerald-500/5 border border-emerald-200 dark:border-emerald-500/10 px-3 py-1 rounded-full">
			{$t('features.badge')}
		</span>
		<h2 class="text-3xl md:text-5xl font-bold text-gray-900 dark:text-white max-w-3xl mx-auto leading-tight">
			{$t('features.title')}
		</h2>
		<p class="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
			{$t('features.subtitle')}
		</p>
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
		<!-- Feature 1 -->
		<div class="bg-white dark:bg-gray-950/20 border border-gray-200 dark:border-gray-900 rounded-2xl p-7 hover:border-emerald-500/30 hover:bg-slate-50 dark:hover:bg-gray-950/40 transition-all duration-300 group shadow-sm dark:shadow-none glow-card" onmousemove={handleMouseMove} role="presentation">
			<div class="w-10 h-10 rounded-lg bg-emerald-50 dark:bg-emerald-500/5 border border-emerald-200 dark:border-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center group-hover:scale-105 transition-transform duration-300">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5.5 w-5.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
				</svg>
			</div>
			<h3 class="text-lg font-bold text-gray-900 dark:text-white mt-5">{$t('features.f1_title')}</h3>
			<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed mt-2.5">{$t('features.f1_desc')}</p>
		</div>

		<!-- Feature 2 -->
		<div class="bg-white dark:bg-gray-950/20 border border-gray-200 dark:border-gray-900 rounded-2xl p-7 hover:border-emerald-500/30 hover:bg-slate-50 dark:hover:bg-gray-950/40 transition-all duration-300 group shadow-sm dark:shadow-none glow-card" onmousemove={handleMouseMove} role="presentation">
			<div class="w-10 h-10 rounded-lg bg-emerald-50 dark:bg-emerald-500/5 border border-emerald-200 dark:border-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center group-hover:scale-105 transition-transform duration-300">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5.5 w-5.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path class="animate-draw" stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
				</svg>
			</div>
			<h3 class="text-lg font-bold text-gray-900 dark:text-white mt-5">{$t('features.f2_title')}</h3>
			<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed mt-2.5">{$t('features.f2_desc')}</p>
		</div>

		<!-- Feature 3 -->
		<div class="bg-white dark:bg-gray-950/20 border border-gray-200 dark:border-gray-900 rounded-2xl p-7 hover:border-emerald-500/30 hover:bg-slate-50 dark:hover:bg-gray-950/40 transition-all duration-300 group shadow-sm dark:shadow-none glow-card" onmousemove={handleMouseMove} role="presentation">
			<div class="w-10 h-10 rounded-lg bg-teal-50 dark:bg-teal-500/5 border border-teal-200 dark:border-teal-500/10 text-teal-600 dark:text-teal-400 flex items-center justify-center group-hover:scale-105 transition-transform duration-300">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5.5 w-5.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
					<circle cx="9" cy="12" r="1.2" fill="currentColor" class="animate-typing-dot-1" />
					<circle cx="12" cy="12" r="1.2" fill="currentColor" class="animate-typing-dot-2" />
					<circle cx="15" cy="12" r="1.2" fill="currentColor" class="animate-typing-dot-3" />
				</svg>
			</div>
			<h3 class="text-lg font-bold text-gray-900 dark:text-white mt-5">{$t('features.f3_title')}</h3>
			<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed mt-2.5">{$t('features.f3_desc')}</p>
		</div>

		<!-- Feature 4 -->
		<div class="bg-white dark:bg-gray-950/20 border border-gray-200 dark:border-gray-900 rounded-2xl p-7 hover:border-emerald-500/30 hover:bg-slate-50 dark:hover:bg-gray-950/40 transition-all duration-300 group shadow-sm dark:shadow-none glow-card" onmousemove={handleMouseMove} role="presentation">
			<div class="w-10 h-10 rounded-lg bg-teal-50 dark:bg-teal-500/5 border border-teal-200 dark:border-teal-500/10 text-teal-600 dark:text-teal-400 flex items-center justify-center group-hover:scale-105 transition-transform duration-300">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5.5 w-5.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path class="animate-float" stroke-linecap="round" stroke-linejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
				</svg>
			</div>
			<h3 class="text-lg font-bold text-gray-900 dark:text-white mt-5">{$t('features.f4_title')}</h3>
			<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed mt-2.5">{$t('features.f4_desc')}</p>
		</div>

		<!-- Feature 5 -->
		<div class="bg-white dark:bg-gray-950/20 border border-gray-200 dark:border-gray-900 rounded-2xl p-7 hover:border-emerald-500/30 hover:bg-slate-50 dark:hover:bg-gray-950/40 transition-all duration-300 group shadow-sm dark:shadow-none glow-card" onmousemove={handleMouseMove} role="presentation">
			<div class="w-10 h-10 rounded-lg bg-emerald-50 dark:bg-emerald-500/5 border border-emerald-200 dark:border-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center group-hover:scale-105 transition-transform duration-300">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5.5 w-5.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path class="animate-float" stroke-linecap="round" stroke-linejoin="round" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
				</svg>
			</div>
			<h3 class="text-lg font-bold text-gray-900 dark:text-white mt-5">{$t('features.f5_title')}</h3>
			<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed mt-2.5">{$t('features.f5_desc')}</p>
		</div>

		<!-- Feature 6 -->
		<div class="bg-white dark:bg-gray-950/20 border border-gray-200 dark:border-gray-900 rounded-2xl p-7 hover:border-emerald-500/30 hover:bg-slate-50 dark:hover:bg-gray-950/40 transition-all duration-300 group shadow-sm dark:shadow-none glow-card" onmousemove={handleMouseMove} role="presentation">
			<div class="w-10 h-10 rounded-lg bg-emerald-50 dark:bg-emerald-500/5 border border-emerald-200 dark:border-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center group-hover:scale-105 transition-transform duration-300">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5.5 w-5.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<!-- Animated lock shackle -->
					<path class="transition-transform duration-300 group-hover:-translate-y-0.75" stroke-linecap="round" stroke-linejoin="round" d="M8 11V7a4 4 0 118 0v4" />
					<path stroke-linecap="round" stroke-linejoin="round" d="M4 11h16v8a2 2 0 01-2 2H6a2 2 0 01-2-2v-8z" />
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 14v3" />
				</svg>
			</div>
			<h3 class="text-lg font-bold text-gray-900 dark:text-white mt-5">{$t('features.f6_title')}</h3>
			<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed mt-2.5">{$t('features.f6_desc')}</p>
		</div>
	</div>
</section>

<!-- Pricing Section -->
<section id="pricing" class="py-24 bg-white dark:bg-[#090c15] border-y border-gray-200 dark:border-gray-950 transition-colors duration-200">
	<div class="max-w-7xl mx-auto px-6">
		<div class="text-center space-y-4 mb-20">
			<h2 class="text-3xl md:text-5xl font-bold text-gray-900 dark:text-white">{$t('pricing.title')}</h2>
			<p class="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">{$t('pricing.subtitle')}</p>
		</div>

		<div class="mx-auto grid max-w-xl grid-cols-1 gap-8">
			<!-- Self-Hosted Open-Source -->
			<div class="bg-slate-50 dark:bg-gray-950/50 border border-gray-200 dark:border-gray-900 rounded-2xl p-8 flex flex-col justify-between space-y-8 relative group shadow-sm dark:shadow-none glow-card" onmousemove={handleMouseMove} role="presentation">
				<div class="space-y-6">
					<span class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 px-2.5 py-1 rounded-full bg-emerald-50 dark:bg-emerald-500/5 border border-emerald-200 dark:border-emerald-500/10 inline-block">
						{$t('pricing.self_title')}
					</span>
					<div class="flex items-baseline space-x-1">
						<span class="text-4xl md:text-5xl font-extrabold text-gray-900 dark:text-white">{$t('pricing.self_price')}</span>
						<span class="text-gray-500 text-sm">/ {$t('pricing.self_period')}</span>
					</div>
					<p class="text-gray-600 dark:text-gray-400 text-sm">{$t('pricing.self_desc')}</p>
					<hr class="border-gray-200 dark:border-gray-900" />
					<ul class="space-y-3.5 text-sm">
						<li class="flex items-start space-x-3 text-gray-700 dark:text-gray-300">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
								<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
							</svg>
							<span>{$t('pricing.self_f1')}</span>
						</li>
						<li class="flex items-start space-x-3 text-gray-700 dark:text-gray-300">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
								<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
							</svg>
							<span>{$t('pricing.self_f2')}</span>
						</li>
						<li class="flex items-start space-x-3 text-gray-700 dark:text-gray-300">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
								<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
							</svg>
							<span>{$t('pricing.self_f3')}</span>
						</li>
						<li class="flex items-start space-x-3 text-gray-700 dark:text-gray-300">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
								<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
							</svg>
							<span>{$t('pricing.self_f4')}</span>
						</li>
					</ul>
				</div>
				<a
					href="https://github.com/pendig/kebaikanku"
					target="_blank"
					rel="noopener"
					class="block text-center w-full bg-gray-900 dark:bg-white dark:hover:bg-gray-100 hover:bg-gray-800 dark:text-gray-950 text-white font-bold py-3.5 rounded-xl border border-gray-800 dark:border-transparent transition-colors cursor-pointer"
				>
					{$t('pricing.self_cta')}
				</a>
			</div>

		</div>
	</div>
</section>

<!-- FAQ Accordion Section -->
<section id="faq" class="py-24 max-w-4xl mx-auto px-6">
	<div class="text-center space-y-4 mb-16">
		<h2 class="text-3xl md:text-5xl font-bold text-gray-900 dark:text-white">{$t('faq.title')}</h2>
		<p class="text-gray-600 dark:text-gray-400">{$t('faq.subtitle')}</p>
	</div>

	<div class="space-y-4">
		<!-- Question 1 -->
		<div class="border border-gray-200 dark:border-gray-900 rounded-xl bg-white dark:bg-gray-950/20 overflow-hidden shadow-sm dark:shadow-none">
			<button
				onclick={() => toggleFaq(1)}
				class="w-full text-left px-6 py-5 flex items-center justify-between text-gray-900 dark:text-white font-bold hover:bg-gray-100/50 dark:hover:bg-gray-900/20 transition-colors cursor-pointer"
			>
				<span>{$t('faq.q1')}</span>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5 text-emerald-600 dark:text-emerald-400 transition-transform duration-300 {openFaq === 1 ? 'rotate-180' : ''}"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="2"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
				</svg>
			</button>
			{#if openFaq === 1}
				<div class="px-6 pb-5 text-sm text-gray-600 dark:text-gray-400 leading-relaxed border-t border-gray-100 dark:border-gray-900/50 pt-3">
					{$t('faq.a1')}
				</div>
			{/if}
		</div>

		<!-- Question 2 -->
		<div class="border border-gray-200 dark:border-gray-900 rounded-xl bg-white dark:bg-gray-950/20 overflow-hidden shadow-sm dark:shadow-none">
			<button
				onclick={() => toggleFaq(2)}
				class="w-full text-left px-6 py-5 flex items-center justify-between text-gray-900 dark:text-white font-bold hover:bg-gray-100/50 dark:hover:bg-gray-900/20 transition-colors cursor-pointer"
			>
				<span>{$t('faq.q3')}</span>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5 text-emerald-600 dark:text-emerald-400 transition-transform duration-300 {openFaq === 2 ? 'rotate-180' : ''}"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="2"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
				</svg>
			</button>
			{#if openFaq === 2}
				<div class="px-6 pb-5 text-sm text-gray-600 dark:text-gray-400 leading-relaxed border-t border-gray-100 dark:border-gray-900/50 pt-3">
					{$t('faq.a3')}
				</div>
			{/if}
		</div>

	</div>
</section>

<style>
	@keyframes float {
		0%, 100% { transform: translateY(0); }
		50% { transform: translateY(-6px); }
	}
	.animate-float {
		animation: float 5s ease-in-out infinite;
	}

	@keyframes pulse-slow {
		0%, 100% { opacity: 1; transform: scale(1); }
		50% { opacity: 0.8; transform: scale(0.96); }
	}
	.animate-pulse-slow {
		animation: pulse-slow 3s ease-in-out infinite;
	}

	@keyframes draw {
		to { stroke-dashoffset: 0; }
	}
	.animate-draw {
		stroke-dasharray: 60;
		stroke-dashoffset: 60;
		animation: draw 2.5s cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	@keyframes typing-dot {
		0%, 100% { opacity: 0.2; transform: translateY(0); }
		50% { opacity: 1; transform: translateY(-2.5px); }
	}
	.animate-typing-dot-1 {
		animation: typing-dot 1.2s infinite ease-in-out;
	}
	.animate-typing-dot-2 {
		animation: typing-dot 1.2s infinite ease-in-out 0.2s;
	}
	.animate-typing-dot-3 {
		animation: typing-dot 1.2s infinite ease-in-out 0.4s;
	}

	@media (prefers-reduced-motion: reduce) {
		*, *::before, *::after {
			animation-duration: 0.01ms !important;
			animation-iteration-count: 1 !important;
			transition-duration: 0.01ms !important;
			scroll-behavior: auto !important;
		}
	}
</style>
