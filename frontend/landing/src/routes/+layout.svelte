<script>
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { t, locale } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { env } from '$env/dynamic/public';

	const {
		PUBLIC_SEO_SITE_NAME,
		PUBLIC_SEO_TITLE,
		PUBLIC_SEO_DESCRIPTION,
		PUBLIC_SEO_SITE_URL,
		PUBLIC_SEO_IMAGE,
		PUBLIC_SEO_KEYWORDS,
		PUBLIC_GA_TRACKING_ID,
		PUBLIC_GTM_CONTAINER_ID
	} = env;

	const siteName = PUBLIC_SEO_SITE_NAME || 'kebaikanku.id';
	const siteTitle = PUBLIC_SEO_TITLE || 'kebaikanku.id - Transparent Donation Platform';
	const siteDescription = PUBLIC_SEO_DESCRIPTION || 'Open-source crowdfunding and donation platform for zakat and kemanusiaan institutions. 0% platform fee for self-hosted.';
	const siteUrl = PUBLIC_SEO_SITE_URL || 'https://kebaikanku.id';
	const siteImage = PUBLIC_SEO_IMAGE || '/images/og-image.png';
	const siteKeywords = PUBLIC_SEO_KEYWORDS || 'donasi, zakat, open-source, crowdfunding, infak, sedekah, kemanusiaan, yayasan, filantropi';
	const gaTrackingId = PUBLIC_GA_TRACKING_ID || '';
	const gtmContainerId = PUBLIC_GTM_CONTAINER_ID || '';


	let { children } = $props();
	let isScrolled = $state(false);
	let isMobileMenuOpen = $state(false);
	let theme = $state('light');

	onMount(() => {
		const handleScroll = () => {
			isScrolled = window.scrollY > 20;
		};
		window.addEventListener('scroll', handleScroll);

		// Initialize theme
		const savedTheme = localStorage.getItem('theme');
		const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
		
		const updateTheme = () => {
			const isDark = savedTheme === 'dark' || (!savedTheme && prefersDark.matches);
			theme = isDark ? 'dark' : 'light';
			if (isDark) {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
		};
		updateTheme();

		prefersDark.addEventListener('change', updateTheme);

		return () => {
			window.removeEventListener('scroll', handleScroll);
			prefersDark.removeEventListener('change', updateTheme);
		};
	});

	function toggleLanguage() {
		locale.update((current) => (current === 'id' ? 'en' : 'id'));
	}

	function toggleTheme() {
		const nextTheme = theme === 'dark' ? 'light' : 'dark';
		localStorage.setItem('theme', nextTheme);
		theme = nextTheme;
		if (nextTheme === 'dark') {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	}
</script>

<svelte:head>
	<title>{siteTitle}</title>
	<link rel="icon" href={favicon} />
	<meta name="description" content={siteDescription} />
	<meta name="keywords" content={siteKeywords} />
	<link rel="canonical" href={$page ? $page.url.href : siteUrl} />

	<!-- Open Graph / Facebook -->
	<meta property="og:type" content="website" />
	<meta property="og:url" content={$page ? $page.url.href : siteUrl} />
	<meta property="og:title" content={siteTitle} />
	<meta property="og:description" content={siteDescription} />
	<meta property="og:image" content={siteImage.startsWith('http') ? siteImage : siteUrl + siteImage} />
	<meta property="og:site_name" content={siteName} />

	<!-- Twitter -->
	<meta property="twitter:card" content="summary_large_image" />
	<meta property="twitter:url" content={$page ? $page.url.href : siteUrl} />
	<meta property="twitter:title" content={siteTitle} />
	<meta property="twitter:description" content={siteDescription} />
	<meta property="twitter:image" content={siteImage.startsWith('http') ? siteImage : siteUrl + siteImage} />

	<!-- Google Fonts -->
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800&family=Plus+Jakarta+Sans:wght@300;400;500;600;700&display=swap" rel="stylesheet">

	{#if gtmContainerId}
		<!-- Google Tag Manager -->
		<script>
			(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
			new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
			j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
			'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
			})(window,document,'script','dataLayer','{gtmContainerId}');
		</script>
		<!-- End Google Tag Manager -->
	{/if}

	{#if gaTrackingId && !gtmContainerId}
		<!-- Google Analytics (gtag.js) -->
		<script async src="https://www.googletagmanager.com/gtag/js?id={gaTrackingId}"></script>
		<script>
			window.dataLayer = window.dataLayer || [];
			function gtag(){dataLayer.push(arguments);}
			gtag('js', new Date());
			gtag('config', '{gaTrackingId}');
		</script>
	{/if}
</svelte:head>


<div class="min-h-screen flex flex-col selection:bg-teal-500/30 selection:text-teal-300">
	<!-- Navbar -->
	<header
		class="fixed top-0 left-0 w-full z-50 transition-all duration-300 {isScrolled
			? 'bg-white/80 dark:bg-[#0b0f19]/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-800/50 py-3 shadow-lg shadow-black/5 dark:shadow-black/10'
			: 'bg-transparent py-5'}"
	>
		<div class="max-w-7xl mx-auto px-6 flex items-center justify-between">
			<!-- Logo -->
			<a href="/" class="flex items-center space-x-2.5 group">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-9 w-9 group-hover:scale-105 transition-transform duration-300 shrink-0" viewBox="0 0 512 512">
					<!-- Left Leaf -->
					<path d="M 256,210 C 256,180 210,130 175,100 C 140,120 180,210 256,210 Z" fill="#4EB73D" />
					<!-- Right Leaf -->
					<path d="M 256,210 C 256,180 302,130 337,100 C 372,120 332,210 256,210 Z" fill="#4EB73D" />
					<!-- Stem -->
					<path d="M 256,210 L 256,275" stroke="#4EB73D" stroke-width="24" stroke-linecap="round" />
					<!-- Bowl (Smile) -->
					<path d="M 150,230 C 150,350 362,350 362,230 L 332,230 C 332,315 180,315 180,230 Z" fill="#FAB100" stroke="#FAB100" stroke-width="8" stroke-linejoin="round" />
				</svg>
				<span class="text-xl font-bold tracking-tight text-gray-900 dark:text-white group-hover:text-[#4EB73D] dark:group-hover:text-[#59C648] transition-colors duration-300">
					kebaikanku<span class="text-[#FAB100] group-hover:text-[#FFA700] transition-colors">.id</span>
				</span>
			</a>

			<!-- Desktop Nav Links -->
			<nav class="hidden md:flex items-center space-x-8">
				<a href="/campaigns" class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.campaigns')}
				</a>
				<a href="/#features" class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.features')}
				</a>
				<a href="/#pricing" class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.pricing')}
				</a>
				<a href="/#faq" class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.faq')}
				</a>
				<a href="https://github.com/pendig/kebaikanku" target="_blank" rel="noopener" class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.docs')}
				</a>
			</nav>

			<!-- Actions -->
			<div class="hidden md:flex items-center space-x-4">
				<!-- Language Switcher -->
				<button
					onclick={toggleLanguage}
					class="px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-gray-900/50 text-xs font-semibold text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:border-gray-300 dark:hover:border-gray-700 transition-all flex items-center space-x-1.5 uppercase cursor-pointer"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-emerald-500 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
					</svg>
					<span>{$locale}</span>
				</button>

				<!-- Theme Switcher -->
				<button
					onclick={toggleTheme}
					class="p-2 rounded-lg border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-gray-900/50 text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:border-gray-300 dark:hover:border-gray-700 transition-all flex items-center justify-center cursor-pointer"
					aria-label="Toggle Theme"
				>
					{#if theme === 'dark'}
						<svg xmlns="http://www.w3.org/2000/svg" class="h-4.5 w-4.5 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707m12.728 0l-.707-.707M6.343 6.343l-.707-.707m2.828 9.9a5 5 0 117.072 0l-7.072 0z" />
						</svg>
					{:else}
						<svg xmlns="http://www.w3.org/2000/svg" class="h-4.5 w-4.5 text-slate-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
						</svg>
					{/if}
				</button>

				<!-- Log In (Points to coming-soon waitlist) -->
				<a
					href="/coming-soon"
					class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium text-sm px-4 py-2 rounded-lg transition-colors cursor-pointer"
				>
					{$t('nav.login')}
				</a>

				<!-- Register CTA (Points to coming-soon waitlist) -->
				<a
					href="/coming-soon"
					class="bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-white font-semibold text-sm px-4 py-2.5 rounded-xl shadow-md shadow-emerald-500/5 dark:shadow-emerald-500/10 hover:shadow-emerald-500/20 transition-all duration-300 hover:-translate-y-0.5 cursor-pointer"
				>
					{$t('nav.startFree')}
				</a>
			</div>

			<!-- Mobile Menu Toggle -->
			<button
				onclick={() => (isMobileMenuOpen = !isMobileMenuOpen)}
				class="md:hidden p-2 text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white focus:outline-none cursor-pointer"
				aria-label="Toggle Menu"
			>
				{#if isMobileMenuOpen}
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
				{:else}
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
					</svg>
				{/if}
			</button>
		</div>

		<!-- Mobile Menu Dropdown -->
		{#if isMobileMenuOpen}
			<div class="md:hidden bg-white dark:bg-[#0b0f19] border-b border-gray-200 dark:border-gray-800 px-6 py-4 space-y-4">
				<div class="flex flex-col space-y-3">
					<a
						href="/campaigns"
						onclick={() => (isMobileMenuOpen = false)}
						class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium py-1 cursor-pointer"
					>
						{$t('nav.campaigns')}
					</a>
					<a
						href="/#features"
						onclick={() => (isMobileMenuOpen = false)}
						class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium py-1 cursor-pointer"
					>
						{$t('nav.features')}
					</a>
					<a
						href="/#pricing"
						onclick={() => (isMobileMenuOpen = false)}
						class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium py-1 cursor-pointer"
					>
						{$t('nav.pricing')}
					</a>
					<a
						href="/#faq"
						onclick={() => (isMobileMenuOpen = false)}
						class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium py-1 cursor-pointer"
					>
						{$t('nav.faq')}
					</a>
					<a
						href="https://github.com/pendig/kebaikanku"
						target="_blank"
						rel="noopener"
						class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium py-1 cursor-pointer"
					>
						{$t('nav.docs')}
					</a>
				</div>
				<hr class="border-gray-200 dark:border-gray-800" />
				<div class="flex items-center justify-between">
					<div class="flex space-x-2">
						<button
							onclick={toggleLanguage}
							class="px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-gray-900/50 text-xs font-semibold text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white flex items-center space-x-1.5 uppercase cursor-pointer"
						>
							<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-emerald-500 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
							</svg>
							<span>{$locale}</span>
						</button>

						<button
							onclick={toggleTheme}
							class="px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-gray-900/50 text-xs font-semibold text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white flex items-center space-x-1.5 cursor-pointer"
						>
							{#if theme === 'dark'}
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707m12.728 0l-.707-.707M6.343 6.343l-.707-.707m2.828 9.9a5 5 0 117.072 0l-7.072 0z" />
								</svg>
								<span>Light</span>
							{:else}
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
								</svg>
								<span>Dark</span>
							{/if}
						</button>
					</div>
					<a href="/coming-soon" onclick={() => (isMobileMenuOpen = false)} class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white font-medium text-sm cursor-pointer">
						{$t('nav.login')}
					</a>
				</div>
				<a
					href="/coming-soon"
					onclick={() => (isMobileMenuOpen = false)}
					class="block text-center w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white font-semibold py-2.5 rounded-xl shadow-md shadow-emerald-500/10 cursor-pointer"
				>
					{$t('nav.startFree')}
				</a>
			</div>
		{/if}
	</header>

	<!-- Main Page Content -->
	<main class="flex-grow pt-24">
		{@render children()}
	</main>

	<!-- Footer -->
	<footer class="bg-slate-50 dark:bg-[#06080f] border-t border-gray-200 dark:border-gray-900 py-16">
		<div class="max-w-7xl mx-auto px-6 grid grid-cols-1 md:grid-cols-4 gap-12">
			<!-- Column 1: Info -->
			<div class="space-y-4 md:col-span-1.5">
				<div class="flex items-center space-x-2.5">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 shrink-0" viewBox="0 0 512 512">
						<!-- Left Leaf -->
						<path d="M 256,210 C 256,180 210,130 175,100 C 140,120 180,210 256,210 Z" fill="#4EB73D" />
						<!-- Right Leaf -->
						<path d="M 256,210 C 256,180 302,130 337,100 C 372,120 332,210 256,210 Z" fill="#4EB73D" />
						<!-- Stem -->
						<path d="M 256,210 L 256,275" stroke="#4EB73D" stroke-width="24" stroke-linecap="round" />
						<!-- Bowl (Smile) -->
						<path d="M 150,230 C 150,350 362,350 362,230 L 332,230 C 332,315 180,315 180,230 Z" fill="#FAB100" stroke="#FAB100" stroke-width="8" stroke-linejoin="round" />
					</svg>
					<span class="text-lg font-bold tracking-tight text-gray-900 dark:text-white">
						kebaikanku<span class="text-[#FAB100]">.id</span>
					</span>
				</div>
				<p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed max-w-sm">
					{$t('footer.desc')}
				</p>
				<div class="flex items-center space-x-3 pt-2">
					<a href="https://github.com/pendig/kebaikanku" target="_blank" rel="noopener" aria-label="GitHub Repository" class="w-8 h-8 rounded-lg bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 flex items-center justify-center text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:border-gray-300 dark:hover:border-gray-700 transition-colors">
						<svg class="w-4 h-4 fill-current" viewBox="0 0 24 24">
							<path d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.9-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.9 1.52 2.34 1.07 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.92 0-1.11.38-2 1.03-2.71-.1-.25-.45-1.29.1-2.64 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33.85 0 1.71.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.55 1.35.2 2.39.1 2.64.65.71 1.03 1.6 1.03 2.71 0 3.82-2.34 4.66-4.57 4.91.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/>
						</svg>
					</a>
				</div>
			</div>

			<!-- Column 2: Quick Links -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-gray-800 dark:text-gray-200 tracking-wider uppercase">
					{$t('footer.links_title')}
				</h3>
				<ul class="space-y-2.5 text-sm">
					<li>
						<a href="/campaigns" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors cursor-pointer">
							{$t('nav.campaigns')}
						</a>
					</li>
					<li>
						<a href="/coming-soon" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors cursor-pointer">
							{$t('nav.dashboard')} (Waitlist)
						</a>
					</li>
					<li>
						<a href="/#features" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors cursor-pointer">
							{$t('nav.features')}
						</a>
					</li>
					<li>
						<a href="/#pricing" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors cursor-pointer">
							{$t('nav.pricing')}
						</a>
					</li>
					<li>
						<a href="/#faq" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors cursor-pointer">
							{$t('nav.faq')}
						</a>
					</li>
				</ul>
			</div>

			<!-- Column 3: Legal -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-gray-800 dark:text-gray-200 tracking-wider uppercase">
					{$t('footer.legal_title')}
				</h3>
				<ul class="space-y-2.5 text-sm">
					<li>
						<a href="/terms" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors">
							{$t('footer.terms')}
						</a>
					</li>
					<li>
						<a href="/privacy" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors">
							{$t('footer.privacy')}
						</a>
					</li>
				</ul>
			</div>

			<!-- Column 4: License -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-gray-800 dark:text-gray-200 tracking-wider uppercase">
					License
				</h3>
				<p class="text-gray-600 dark:text-gray-400 text-xs leading-relaxed">
					This software is licensed under the <a href="https://www.gnu.org/licenses/agpl-3.0" target="_blank" rel="noopener" class="text-emerald-600 dark:text-emerald-400 hover:underline">AGPL-3.0</a> copyleft license. 
					<br class="mb-2" />
					Commercial dual-licensing is available for white-labeling, closed-source setups, and dedicated support.
				</p>
			</div>
		</div>

		<!-- Copyright -->
		<div class="max-w-7xl mx-auto px-6 mt-12 pt-8 border-t border-gray-200 dark:border-gray-900 flex flex-col md:flex-row items-center justify-between text-xs text-gray-600 dark:text-gray-500 gap-4">
			<span>{$t('footer.copyright')}</span>
			<div class="flex items-center space-x-4">
				<span class="px-2 py-0.5 rounded border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-gray-900/50 text-gray-600 dark:text-gray-400">AGPLv3 Open Source</span>
				<span class="px-2 py-0.5 rounded border border-emerald-200 dark:border-emerald-900/50 bg-emerald-50 dark:bg-emerald-950/20 text-emerald-600 dark:text-emerald-400">Enterprise Ready</span>
			</div>
		</div>
	</footer>
</div>
