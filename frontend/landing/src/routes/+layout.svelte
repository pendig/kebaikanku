<script>
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { t, locale } from '$lib/i18n';
	import { onMount } from 'svelte';

	let { children } = $props();
	let isScrolled = $state(false);
	let isMobileMenuOpen = $state(false);

	onMount(() => {
		const handleScroll = () => {
			isScrolled = window.scrollY > 20;
		};
		window.addEventListener('scroll', handleScroll);
		return () => {
			window.removeEventListener('scroll', handleScroll);
		};
	});

	function toggleLanguage() {
		locale.update((current) => (current === 'id' ? 'en' : 'id'));
	}
</script>

<svelte:head>
	<title>kebaikanku.id - Transparent Donation Platform</title>
	<link rel="icon" href={favicon} />
	<meta name="description" content="Open-source crowdfunding and donation platform for zakat and kemanusiaan institutions. 0% platform fee for self-hosted." />
	<!-- Google Fonts -->
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800&family=Plus+Jakarta+Sans:wght@300;400;500;600;700&display=swap" rel="stylesheet">
	<style>
		:global(body) {
			font-family: 'Plus Jakarta Sans', sans-serif;
			background-color: #0b0f19;
			color: #f3f4f6;
			overflow-x: hidden;
		}
		:global(h1, h2, h3, h4, h5, h6) {
			font-family: 'Outfit', sans-serif;
		}
	</style>
</svelte:head>

<div class="min-h-screen flex flex-col selection:bg-teal-500/30 selection:text-teal-300">
	<!-- Navbar -->
	<header
		class="fixed top-0 left-0 w-full z-50 transition-all duration-300 {isScrolled
			? 'bg-[#0b0f19]/80 backdrop-blur-md border-b border-gray-800/50 py-3 shadow-lg shadow-black/10'
			: 'bg-transparent py-5'}"
	>
		<div class="max-w-7xl mx-auto px-6 flex items-center justify-between">
			<!-- Logo -->
			<a href="/" class="flex items-center space-x-2.5 group">
				<div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-emerald-500 to-teal-400 flex items-center justify-center shadow-md shadow-emerald-500/20 group-hover:scale-105 transition-transform duration-300">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6.5 w-6.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
					</svg>
				</div>
				<span class="text-xl font-bold tracking-tight text-white group-hover:text-emerald-400 transition-colors duration-300">
					kebaikanku<span class="text-emerald-400">.id</span>
				</span>
			</a>

			<!-- Desktop Nav Links -->
			<nav class="hidden md:flex items-center space-x-8">
				<a href="/#features" class="text-gray-300 hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.features')}
				</a>
				<a href="/#pricing" class="text-gray-300 hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.pricing')}
				</a>
				<a href="/#faq" class="text-gray-300 hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.faq')}
				</a>
				<a href="https://github.com/kebaikankuid/kebaikanku" target="_blank" rel="noopener" class="text-gray-300 hover:text-white font-medium text-sm transition-colors duration-200">
					{$t('nav.docs')}
				</a>
			</nav>

			<!-- Actions -->
			<div class="hidden md:flex items-center space-x-4">
				<!-- Language Switcher -->
				<button
					onclick={toggleLanguage}
					class="px-3 py-1.5 rounded-lg border border-gray-800 bg-gray-900/50 text-xs font-semibold text-gray-300 hover:text-white hover:border-gray-700 transition-all flex items-center space-x-1.5 uppercase cursor-pointer"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5c-.006 2.16-.454 4.223-1.251 6.1A18.007 18.007 0 0118 9m-6-4H3" />
					</svg>
					<span>{$locale}</span>
				</button>

				<!-- Log In -->
				<a
					href="https://app.kebaikanku.id/login"
					class="text-gray-300 hover:text-white font-medium text-sm px-4 py-2 rounded-lg transition-colors"
				>
					{$t('nav.login')}
				</a>

				<!-- Register CTA -->
				<a
					href="https://app.kebaikanku.id/register"
					class="bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-white font-semibold text-sm px-4 py-2.5 rounded-xl shadow-md shadow-emerald-500/10 hover:shadow-emerald-500/25 transition-all duration-300 hover:-translate-y-0.5"
				>
					{$t('nav.startFree')}
				</a>
			</div>

			<!-- Mobile Menu Toggle -->
			<button
				onclick={() => (isMobileMenuOpen = !isMobileMenuOpen)}
				class="md:hidden p-2 text-gray-400 hover:text-white focus:outline-none cursor-pointer"
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
			<div class="md:hidden bg-[#0b0f19] border-b border-gray-800 px-6 py-4 space-y-4">
				<div class="flex flex-col space-y-3">
					<a
						href="/#features"
						onclick={() => (isMobileMenuOpen = false)}
						class="text-gray-300 hover:text-white font-medium py-1"
					>
						{$t('nav.features')}
					</a>
					<a
						href="/#pricing"
						onclick={() => (isMobileMenuOpen = false)}
						class="text-gray-300 hover:text-white font-medium py-1"
					>
						{$t('nav.pricing')}
					</a>
					<a
						href="/#faq"
						onclick={() => (isMobileMenuOpen = false)}
						class="text-gray-300 hover:text-white font-medium py-1"
					>
						{$t('nav.faq')}
					</a>
					<a
						href="https://github.com/kebaikankuid/kebaikanku"
						target="_blank"
						rel="noopener"
						class="text-gray-300 hover:text-white font-medium py-1"
					>
						{$t('nav.docs')}
					</a>
				</div>
				<hr class="border-gray-800" />
				<div class="flex items-center justify-between">
					<button
						onclick={toggleLanguage}
						class="px-3 py-1.5 rounded-lg border border-gray-800 bg-gray-900/50 text-xs font-semibold text-gray-300 hover:text-white flex items-center space-x-1.5 uppercase cursor-pointer"
					>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5c-.006 2.16-.454 4.223-1.251 6.1A18.007 18.007 0 0118 9m-6-4H3" />
						</svg>
						<span>{$locale}</span>
					</button>
					<a href="https://app.kebaikanku.id/login" class="text-gray-300 hover:text-white font-medium text-sm">
						{$t('nav.login')}
					</a>
				</div>
				<a
					href="https://app.kebaikanku.id/register"
					class="block text-center w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white font-semibold py-2.5 rounded-xl shadow-md shadow-emerald-500/10"
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
	<footer class="bg-[#06080f] border-t border-gray-900 py-16">
		<div class="max-w-7xl mx-auto px-6 grid grid-cols-1 md:grid-cols-4 gap-12">
			<!-- Column 1: Info -->
			<div class="space-y-4 md:col-span-1.5">
				<div class="flex items-center space-x-2.5">
					<div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-emerald-500 to-teal-400 flex items-center justify-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
						</svg>
					</div>
					<span class="text-lg font-bold tracking-tight text-white">
						kebaikanku<span class="text-emerald-400">.id</span>
					</span>
				</div>
				<p class="text-gray-400 text-sm leading-relaxed max-w-sm">
					{$t('footer.desc')}
				</p>
				<div class="flex items-center space-x-3 pt-2">
					<a href="https://github.com/kebaikankuid/kebaikanku" target="_blank" rel="noopener" aria-label="GitHub Repository" class="w-8 h-8 rounded-lg bg-gray-900 border border-gray-800 flex items-center justify-center text-gray-400 hover:text-white hover:border-gray-700 transition-colors">
						<svg class="w-4 h-4 fill-current" viewBox="0 0 24 24">
							<path d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.9-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.9 1.52 2.34 1.07 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.92 0-1.11.38-2 1.03-2.71-.1-.25-.45-1.29.1-2.64 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33.85 0 1.71.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.55 1.35.2 2.39.1 2.64.65.71 1.03 1.6 1.03 2.71 0 3.82-2.34 4.66-4.57 4.91.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/>
						</svg>
					</a>
				</div>
			</div>

			<!-- Column 2: Quick Links -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-gray-200 tracking-wider uppercase">
					{$t('footer.links_title')}
				</h3>
				<ul class="space-y-2.5 text-sm">
					<li>
						<a href="/#features" class="text-gray-400 hover:text-white transition-colors">
							{$t('nav.features')}
						</a>
					</li>
					<li>
						<a href="/#pricing" class="text-gray-400 hover:text-white transition-colors">
							{$t('nav.pricing')}
						</a>
					</li>
					<li>
						<a href="/#faq" class="text-gray-400 hover:text-white transition-colors">
							{$t('nav.faq')}
						</a>
					</li>
					<li>
						<a href="https://github.com/kebaikankuid/kebaikanku" target="_blank" rel="noopener" class="text-gray-400 hover:text-white transition-colors">
							{$t('nav.docs')}
						</a>
					</li>
				</ul>
			</div>

			<!-- Column 3: Legal -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-gray-200 tracking-wider uppercase">
					{$t('footer.legal_title')}
				</h3>
				<ul class="space-y-2.5 text-sm">
					<li>
						<a href="/terms" class="text-gray-400 hover:text-white transition-colors">
							{$t('footer.terms')}
						</a>
					</li>
					<li>
						<a href="/privacy" class="text-gray-400 hover:text-white transition-colors">
							{$t('footer.privacy')}
						</a>
					</li>
				</ul>
			</div>

			<!-- Column 4: License -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-gray-200 tracking-wider uppercase">
					License
				</h3>
				<p class="text-gray-400 text-xs leading-relaxed">
					This software is licensed under the <a href="https://www.gnu.org/licenses/agpl-3.0" target="_blank" rel="noopener" class="text-emerald-400 hover:underline">AGPL-3.0</a> copyleft license. 
					<br class="mb-2" />
					Commercial dual-licensing is available for white-labeling, closed-source setups, and dedicated support.
				</p>
			</div>
		</div>

		<!-- Copyright -->
		<div class="max-w-7xl mx-auto px-6 mt-12 pt-8 border-t border-gray-900 flex flex-col md:flex-row items-center justify-between text-xs text-gray-500 gap-4">
			<span>{$t('footer.copyright')}</span>
			<div class="flex items-center space-x-4">
				<span class="px-2 py-0.5 rounded border border-gray-800 bg-gray-900/50">AGPLv3 Open Source</span>
				<span class="px-2 py-0.5 rounded border border-emerald-900/50 bg-emerald-950/20 text-emerald-400">Enterprise Ready</span>
			</div>
		</div>
	</footer>
</div>
