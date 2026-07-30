<script>
	import { onMount } from 'svelte';
	import { page } from '$app/state';

	const apiBase = (import.meta.env.PUBLIC_API_BASE_URL || '').replace(/\/$/, '');
	let payment = $state(null);
	let loading = $state(true);
	let error = $state('');
	let timer;

	let kind = $derived(payment?.status === 'success'
		? 'success'
		: ['expire', 'expired'].includes(payment?.provider_status)
			? 'expired'
			: payment?.status === 'failed' ? 'failed' : 'pending');
	let copy = $derived(kind === 'success'
		? { title: 'Donasi berhasil', text: 'Terima kasih. Pembayaran telah diterima dan donasi sudah tercatat.' }
		: kind === 'expired'
			? { title: 'Pembayaran kedaluwarsa', text: 'Batas waktu pembayaran telah lewat. Silakan kembali ke kampanye untuk membuat pembayaran baru.' }
			: kind === 'failed'
				? { title: 'Pembayaran tidak berhasil', text: 'Transaksi dibatalkan atau ditolak. Dana tidak tercatat sebagai donasi.' }
				: { title: 'Menunggu pembayaran', text: 'Selesaikan pembayaran di Midtrans. Status halaman ini diperiksa langsung ke Midtrans.' });

	onMount(() => {
		for (let index = sessionStorage.length - 1; index >= 0; index--) {
			const key = sessionStorage.key(index);
			if (key?.startsWith('donation:')) sessionStorage.removeItem(key);
		}
		refresh();
		return () => clearTimeout(timer);
	});

	async function refresh() {
		clearTimeout(timer);
		loading = !payment;
		error = '';
		try {
			const response = await fetch(`${apiBase}/api/v1/donations/${encodeURIComponent(page.params.id)}/status`);
			const payload = await response.json().catch(() => null);
			if (!response.ok || !payload?.success) throw new Error(payload?.error?.message || 'Status pembayaran tidak dapat diperiksa.');
			payment = payload.data;
			if (payment.status === 'pending') timer = setTimeout(refresh, 4000);
		} catch (cause) {
			error = cause.message || 'Status pembayaran tidak dapat diperiksa.';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>{copy.title} | kebaikanku.id</title>
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

<section class="min-h-screen bg-slate-50 px-6 pb-28 pt-24 text-gray-950 dark:bg-[#0b0f19] dark:text-white">
	<div class="mx-auto max-w-xl">
		<div class="rounded-3xl border border-gray-200 bg-white p-7 text-center shadow-xl shadow-gray-200/50 dark:border-gray-900 dark:bg-gray-950/70 dark:shadow-black/20 md:p-10">
			<div class="mx-auto mb-5 grid h-16 w-16 place-items-center rounded-full {kind === 'success' ? 'bg-emerald-100 text-emerald-700' : kind === 'pending' ? 'bg-amber-100 text-amber-700' : 'bg-red-100 text-red-700'}" aria-hidden="true">
				{#if kind === 'success'}
					<svg viewBox="0 0 24 24" class="h-8 w-8" fill="none" stroke="currentColor" stroke-width="2.5"><path d="m5 12 4 4L19 6" /></svg>
				{:else if kind === 'pending'}
					<svg viewBox="0 0 24 24" class="h-8 w-8" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>
				{:else}
					<svg viewBox="0 0 24 24" class="h-8 w-8" fill="none" stroke="currentColor" stroke-width="2.5"><path d="m7 7 10 10M17 7 7 17" /></svg>
				{/if}
			</div>
			<p class="text-xs font-extrabold uppercase tracking-[0.18em] text-emerald-600 dark:text-emerald-400">Status pembayaran</p>
			<h1 class="mt-2 text-3xl font-black tracking-tight">{copy.title}</h1>
			<p class="mx-auto mt-3 max-w-md text-sm leading-6 text-gray-600 dark:text-gray-400">{copy.text}</p>

			{#if loading}
				<p class="mt-6 text-sm font-bold text-gray-500" role="status">Memeriksa status...</p>
			{:else if error}
				<p class="mt-6 rounded-xl bg-red-50 p-3 text-sm font-semibold text-red-700 dark:bg-red-950/20 dark:text-red-300" role="alert">{error}</p>
			{:else if payment}
				<dl class="mt-7 space-y-3 rounded-2xl bg-slate-50 p-5 text-left text-sm dark:bg-gray-900/70">
					<div class="flex justify-between gap-4"><dt>ID donasi</dt><dd class="max-w-[65%] break-all text-right font-bold">{payment.donation_id}</dd></div>
					<div class="flex justify-between gap-4"><dt>Status Midtrans</dt><dd class="font-bold capitalize">{payment.provider_status || payment.status}</dd></div>
				</dl>
			{/if}

			<div class="mt-7 grid gap-3 sm:grid-cols-2">
				<a href="/campaigns" class="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-emerald-500 to-yellow-400 px-5 py-3 text-sm font-extrabold text-slate-950"><svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 12h16M10 6l-6 6 6 6" /></svg> Lihat kampanye</a>
				<button type="button" onclick={refresh} disabled={loading} aria-label="Periksa ulang status pembayaran" class="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl border border-gray-200 px-5 py-3 text-sm font-extrabold disabled:opacity-60 dark:border-gray-800"><svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 7v5h-5M4 17v-5h5" /><path d="M6 9a7 7 0 0 1 12-2l2 5M18 15a7 7 0 0 1-12 2l-2-5" /></svg> Periksa ulang</button>
			</div>
		</div>
	</div>
</section>
