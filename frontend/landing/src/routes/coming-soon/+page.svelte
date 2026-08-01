<svelte:head>
	<title>Kampanye | kebaikanku.id</title>
	<meta name="description" content="Lihat kampanye aktif dan berdonasi dengan aman." />
</svelte:head>

<section class="relative flex min-h-[70vh] items-center justify-center overflow-hidden px-6 py-20">
	<div class="absolute inset-0 z-0 bg-slate-50 dark:bg-[#0b0f19]">
		<div class="absolute inset-0 bg-[linear-gradient(to_right,#1f293708_1px,transparent_1px),linear-gradient(to_bottom,#1f293708_1px,transparent_1px)] bg-[size:4rem_4rem]"></div>
		<div class="absolute left-1/2 top-1/2 h-[500px] w-[500px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-emerald-500/5 blur-[120px]"></div>
	</div>

	<div class="relative z-10 w-full max-w-2xl text-center">
		<div class="rounded-3xl border border-gray-200 bg-white/40 p-8 shadow-xl shadow-black/5 backdrop-blur-xl dark:border-gray-800/80 dark:bg-gray-950/40 dark:shadow-black/80 md:p-12">
			<div class="space-y-6">
				<div class="inline-flex items-center space-x-2 rounded-full border border-teal-200 bg-teal-50 px-3 py-1 text-xs font-semibold text-teal-600 dark:border-teal-500/20 dark:bg-teal-500/5 dark:text-teal-400">
					<span class="inline-flex h-2 w-2 rounded-full bg-teal-500"></span>
					<span>Donasi publik tersedia</span>
				</div>
				<h1 class="text-3xl font-extrabold leading-tight tracking-tight text-gray-900 dark:text-white md:text-4xl">Temukan kampanye yang ingin Anda bantu</h1>
				<p class="mx-auto max-w-lg text-sm leading-relaxed text-gray-600 dark:text-gray-400 md:text-base">Dashboard operator masih akses terbatas. Untuk berdonasi, pilih kampanye aktif dan lanjutkan ke halaman pembayaran aman.</p>
				<a href="/campaigns" class="inline-flex items-center justify-center rounded-xl bg-gradient-to-r from-emerald-500 to-teal-500 px-6 py-3.5 text-sm font-bold text-white shadow-lg shadow-emerald-500/10 transition-colors hover:from-emerald-600 hover:to-teal-600">Lihat kampanye</a>
			</div>
		</div>
	</div>
</section>
