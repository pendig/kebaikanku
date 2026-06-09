<script>
	import { t } from '$lib/i18n';
	import { onMount } from 'svelte';

	let email = $state('');
	let emailError = $state('');
	let isSubmitted = $state(false);
	let isLoading = $state(false);
	let totalSignups = $state(0);

	onMount(() => {
		// Calculate current signups count (simulated)
		const saved = localStorage.getItem('kbk_waitlist_emails');
		const emailList = saved ? JSON.parse(saved) : [];
		totalSignups = 142 + emailList.length; // Base mock number + actual local entries
	});

	function validateEmail(val) {
		const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		return re.test(String(val).toLowerCase());
	}

	function handleSubmit(e) {
		e.preventDefault();
		emailError = '';

		if (!email) {
			emailError = 'Email wajib diisi / Email is required';
			return;
		}

		if (!validateEmail(email)) {
			emailError = 'Format email tidak valid / Invalid email format';
			return;
		}

		isLoading = true;

		// Simulated premium transition loading (1.2s delay)
		setTimeout(() => {
			isLoading = false;
			isSubmitted = true;
			
			// Save locally
			const saved = localStorage.getItem('kbk_waitlist_emails');
			const emailList = saved ? JSON.parse(saved) : [];
			if (!emailList.includes(email)) {
				emailList.push(email);
				localStorage.setItem('kbk_waitlist_emails', JSON.stringify(emailList));
				totalSignups += 1;
			}
			
			// Mock Analytics tracking
			console.log('Waitlist Signup Event Triggered:', {
				email: email,
				timestamp: new Date().toISOString()
			});
		}, 1200);
	}
</script>

<svelte:head>
	<title>{$t('coming_soon.title')} | kebaikanku.id</title>
	<meta name="description" content={$t('coming_soon.waitlist_desc')} />
</svelte:head>

<section class="relative min-h-[70vh] flex items-center justify-center py-20 px-6 overflow-hidden">
	<!-- Glass background accents -->
	<div class="absolute inset-0 bg-slate-50 dark:bg-[#0b0f19] z-0 transition-colors duration-200">
		<div class="absolute inset-0 bg-[linear-gradient(to_right,#1f293708_1px,transparent_1px),linear-gradient(to_bottom,#1f293708_1px,transparent_1px)] bg-[size:4rem_4rem]"></div>
		<div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] bg-emerald-500/5 blur-[120px] rounded-full"></div>
		<div class="absolute top-1/3 left-1/4 w-[300px] h-[300px] bg-teal-500/5 blur-[100px] rounded-full"></div>
	</div>

	<div class="max-w-2xl w-full relative z-10 text-center">
		<!-- Liquid Glass Card Container -->
		<div class="relative rounded-3xl border border-gray-200 dark:border-gray-800/80 bg-white/40 dark:bg-gray-950/40 p-8 md:p-12 shadow-xl dark:shadow-2xl shadow-black/5 dark:shadow-black/80 backdrop-blur-xl overflow-hidden group">
			<!-- Glowing decorative edge -->
			<div class="absolute inset-x-0 top-0 h-[1px] bg-gradient-to-r from-transparent via-emerald-500/30 to-transparent"></div>
			
			{#if !isSubmitted}
				<div class="space-y-6">
					<!-- Badge -->
					<div class="inline-flex items-center space-x-2 px-3 py-1 rounded-full border border-teal-200 dark:border-teal-500/20 bg-teal-50 dark:bg-teal-500/5 text-xs font-semibold text-teal-600 dark:text-teal-400">
						<span class="flex h-2 w-2 relative">
							<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-teal-400 opacity-75"></span>
							<span class="relative inline-flex rounded-full h-2 w-2 bg-teal-500"></span>
						</span>
						<span>Dashboard coming soon</span>
					</div>

					<!-- Headings -->
					<h1 class="text-3xl md:text-4xl font-extrabold tracking-tight text-gray-900 dark:text-white leading-tight">
						{$t('coming_soon.title')}
					</h1>
					<p class="text-gray-600 dark:text-gray-400 text-sm md:text-base leading-relaxed max-w-lg mx-auto">
						{$t('coming_soon.subtitle')}
					</p>

					<div class="py-4 px-6 rounded-2xl bg-gray-50 dark:bg-gray-900/30 border border-gray-200 dark:border-gray-800/50 max-w-md mx-auto text-left space-y-2">
						<div class="flex items-center space-x-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-emerald-500 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
							</svg>
							<span>Dashboard Roadmap</span>
						</div>
						<ul class="text-xs text-gray-600 dark:text-gray-500 space-y-1.5 list-disc pl-4 leading-normal">
							<li><strong class="text-gray-900 dark:text-gray-300">Self-Hosted:</strong> Open source, 0% platform fee, connect your Midtrans account directly.</li>
							<li><strong class="text-gray-900 dark:text-gray-300">Managed Cloud SaaS:</strong> Setup in 5 mins, low cost, automatic updates & backups.</li>
						</ul>
					</div>

					<!-- Waitlist Box -->
					<div class="pt-4 max-w-md mx-auto">
						<h2 class="text-lg font-bold text-gray-900 dark:text-white mb-2">{$t('coming_soon.waitlist_title')}</h2>
						<p class="text-xs text-gray-600 dark:text-gray-400 mb-6">{$t('coming_soon.waitlist_desc')}</p>

						<form onsubmit={handleSubmit} class="space-y-4 text-left">
							<div class="relative">
								<input
									type="email"
									bind:value={email}
									placeholder={$t('coming_soon.placeholder_email')}
									disabled={isLoading}
									class="w-full px-4 py-3.5 rounded-xl border {emailError ? 'border-red-500/50 focus:border-red-500' : 'border-gray-200 dark:border-gray-800 focus:border-emerald-500'} bg-white dark:bg-gray-900/60 text-gray-950 dark:text-white placeholder-gray-450 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/20 transition-all text-sm shadow-sm dark:shadow-none"
								/>
								{#if emailError}
									<p class="text-red-550 dark:text-red-400 text-xs mt-1.5 pl-1">{emailError}</p>
								{/if}
							</div>
							
							<button
								type="submit"
								disabled={isLoading}
								class="w-full bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 disabled:opacity-50 text-white font-bold py-3.5 rounded-xl shadow-lg shadow-emerald-500/10 transition-all duration-300 flex items-center justify-center space-x-2 text-sm cursor-pointer"
							>
								{#if isLoading}
									<!-- Spinner -->
									<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
									<span>Processing...</span>
								{:else}
									<span>{$t('coming_soon.btn_submit')}</span>
								{/if}
							</button>
						</form>

						<!-- Social Proof/Stats -->
						<div class="mt-6 text-center text-xs text-gray-600 dark:text-gray-500">
							<span class="text-emerald-600 dark:text-emerald-400 font-bold">{totalSignups}</span> yayasan & organisasi telah bergabung di antrean.
						</div>
					</div>
				</div>
			{:else}
				<!-- Success Screen -->
				<div class="space-y-6 py-8 animate-fade-in">
					<div class="w-16 h-16 bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-200 dark:border-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded-full flex items-center justify-center mx-auto shadow-md dark:shadow-lg dark:shadow-emerald-500/5">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
						</svg>
					</div>

					<h2 class="text-2xl md:text-3xl font-extrabold text-gray-900 dark:text-white">
						{$t('coming_soon.success_title')}
					</h2>
					<p class="text-gray-600 dark:text-gray-400 text-sm md:text-base leading-relaxed max-w-md mx-auto">
						{$t('coming_soon.success_desc')}
					</p>

					<div class="pt-6">
						<a
							href="/"
							class="inline-block bg-gray-100 hover:bg-gray-200 dark:bg-gray-900 dark:hover:bg-gray-800 border border-gray-200 dark:border-gray-800 text-gray-800 dark:text-white font-semibold px-6 py-3 rounded-xl transition-all duration-300 text-sm"
						>
							{$t('coming_soon.back_home')}
						</a>
					</div>
				</div>
			{/if}
		</div>
	</div>
</section>
