import "../../chunks/index-server.js";
import { _ as attr, n as ensure_array_like, r as head, v as escape_html } from "../../chunks/server.js";
//#region src/routes/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		"".replace(/\/$/, "");
		let token = "";
		let campaigns = [];
		let donations = [];
		let campaign = {
			organization_id: "pilot-org",
			title: "",
			slug: "",
			description: "",
			category: "infak",
			target_amount: 1e7,
			end_date: "2026-12-31T23:59:59+07:00"
		};
		function formatRupiah(amount) {
			return new Intl.NumberFormat("id-ID", {
				style: "currency",
				currency: "IDR",
				maximumFractionDigits: 0
			}).format(Number(amount || 0));
		}
		head("1uha8ag", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Dashboard | kebaikanku.id</title>`);
			});
		});
		$$renderer.push(`<main class="min-h-screen bg-slate-100 p-5 text-slate-950"><div class="mx-auto max-w-7xl space-y-5"><header class="flex flex-col gap-3 border-b border-slate-300 pb-5 md:flex-row md:items-end md:justify-between"><div><p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Alpha dashboard</p> <h1 class="text-3xl font-black tracking-tight">kebaikanku.id</h1> <p class="text-sm text-slate-600">Campaign operations with admin token. Full signup/login later.</p></div> <div class="flex gap-2"><input${attr("value", token)} class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm md:w-80 svelte-1uha8ag" placeholder="CAMPAIGN_ADMIN_TOKEN"/> <button class="rounded-lg bg-slate-950 px-4 py-2 text-sm font-bold text-white">Save</button></div></header> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <section class="grid gap-5 lg:grid-cols-[420px_1fr]"><form class="rounded-xl border border-slate-300 bg-white p-5 shadow-sm"><h2 class="mb-4 text-lg font-black">Create campaign</h2> <label class="svelte-1uha8ag">Title<input${attr("value", campaign.title)} required="" class="svelte-1uha8ag"/></label> <label class="svelte-1uha8ag">Slug<input${attr("value", campaign.slug)} required="" pattern="[a-z0-9-]+" class="svelte-1uha8ag"/></label> <label class="svelte-1uha8ag">Description<textarea required="" class="svelte-1uha8ag">`);
		const $$body = escape_html(campaign.description);
		if ($$body) $$renderer.push(`${$$body}`);
		$$renderer.push(`</textarea></label> <label class="svelte-1uha8ag">Category `);
		$$renderer.select({
			value: campaign.category,
			class: ""
		}, ($$renderer) => {
			$$renderer.option({ value: "infak" }, ($$renderer) => {
				$$renderer.push(`Infak`);
			});
			$$renderer.option({ value: "kemanusiaan" }, ($$renderer) => {
				$$renderer.push(`Kemanusiaan`);
			});
			$$renderer.option({ value: "zakat_maal" }, ($$renderer) => {
				$$renderer.push(`Zakat Maal`);
			});
			$$renderer.option({ value: "zakat_fitrah" }, ($$renderer) => {
				$$renderer.push(`Zakat Fitrah`);
			});
		}, "svelte-1uha8ag");
		$$renderer.push(`</label> <div class="grid grid-cols-2 gap-3"><label class="svelte-1uha8ag">Target<input${attr("value", campaign.target_amount)} min="1" step="1" type="number" required="" class="svelte-1uha8ag"/></label> <label class="svelte-1uha8ag">End date<input${attr("value", campaign.end_date)} required="" class="svelte-1uha8ag"/></label></div> <button${attr("disabled", true, true)} class="mt-4 w-full rounded-lg bg-emerald-600 px-4 py-3 text-sm font-black text-white disabled:opacity-50">Create</button></form> <div class="space-y-5"><section class="rounded-xl border border-slate-300 bg-white p-5 shadow-sm"><div class="mb-4 flex items-center justify-between"><h2 class="text-lg font-black">Campaigns</h2> <button class="rounded-lg border border-slate-300 px-3 py-2 text-xs font-bold">Refresh</button></div> <div class="grid gap-3 md:grid-cols-2"><!--[-->`);
		const each_array = ensure_array_like(campaigns);
		for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
			let item = each_array[$$index];
			$$renderer.push(`<article class="rounded-lg border border-slate-200 p-4"><p class="text-xs font-bold uppercase text-emerald-700">${escape_html(item.category)}</p> <h3 class="font-black">${escape_html(item.title)}</h3> <p class="mt-1 line-clamp-2 text-sm text-slate-600">${escape_html(item.description)}</p> <p class="mt-3 text-sm font-bold">${escape_html(formatRupiah(item.collected_amount))} / ${escape_html(formatRupiah(item.target_amount))}</p></article>`);
		}
		$$renderer.push(`<!--]--></div></section> <section class="rounded-xl border border-slate-300 bg-white p-5 shadow-sm"><div class="mb-4 flex items-center justify-between"><h2 class="text-lg font-black">Donation status</h2> <button${attr("disabled", true, true)} class="rounded-lg border border-slate-300 px-3 py-2 text-xs font-bold disabled:opacity-50">Load CSV</button></div> <div class="overflow-x-auto"><table class="w-full text-left text-sm"><thead><tr class="border-b border-slate-200 text-xs uppercase text-slate-500"><th class="svelte-1uha8ag">Campaign</th><th class="svelte-1uha8ag">Donor</th><th class="svelte-1uha8ag">Amount</th><th class="svelte-1uha8ag">Status</th><th class="svelte-1uha8ag">Paid</th></tr></thead><tbody><!--[-->`);
		const each_array_1 = ensure_array_like(donations);
		for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
			let donation = each_array_1[$$index_1];
			$$renderer.push(`<tr class="border-b border-slate-100"><td class="svelte-1uha8ag">${escape_html(donation.campaign)}</td><td class="svelte-1uha8ag">${escape_html(donation.donor_name)}</td><td class="svelte-1uha8ag">${escape_html(formatRupiah(donation.amount))}</td><td class="svelte-1uha8ag">${escape_html(donation.status)}</td><td class="svelte-1uha8ag">${escape_html(donation.paid_at)}</td></tr>`);
		}
		$$renderer.push(`<!--]--></tbody></table></div></section></div></section></div></main>`);
	});
}
//#endregion
export { _page as default };
