<script>
	import { t } from '$lib/i18n';
</script>

<div class="max-w-4xl mx-auto px-6 py-12 md:py-20">
	<div class="space-y-6 border-b border-gray-900 pb-10">
		<h1 class="text-3xl md:text-5xl font-extrabold text-white">
			{$t('footer.privacy')}
		</h1>
		<p class="text-sm text-gray-500">Last updated: June 9, 2026</p>
	</div>

	<div class="mt-10 prose prose-invert prose-emerald max-w-none text-gray-300 space-y-8 text-sm md:text-base leading-relaxed">
		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-white">1. Information We Collect</h2>
			<p>
				We collect information necessary to process donations, contact donors, and optimize platform usage.
			</p>
			<ul class="list-disc pl-6 space-y-2">
				<li><strong>Donor Information:</strong> Name, email address, phone number, and transaction history.</li>
				<li><strong>Institution Profiles:</strong> Organization names, addresses, tax registry documents, bank account details, and staff credentials.</li>
				<li><strong>Usage & Chat logs:</strong> Technical telemetry data, and conversation history in case you interact with our AI-powered Customer Service.</li>
			</ul>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-white">2. Data Sovereignty & Ownership</h2>
			<p>
				Data ownership is key to trust. Unlike commercial donation sites:
			</p>
			<ul class="list-disc pl-6 space-y-2">
				<li><strong>For Self-Hosted Deployments:</strong> Your organization has absolute sovereignty over all donor, transaction, and beneficiary data. No data is sent to kebaikanku.id servers. Your data is stored directly in your SQLite/Postgres database.</li>
				<li><strong>For Managed Cloud Deployments:</strong> Your database is strictly isolated. We do not monetize, sell, or analyze your donor profiles for cross-marketing purposes.</li>
			</ul>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-white">3. How We Use Your Data</h2>
			<p>
				We use collected data strictly for:
			</p>
			<ul class="list-disc pl-6 space-y-2">
				<li>Initiating payment requests, QRIS generation, and Virtual Account validation via payment gateways.</li>
				<li>Generating official tax-deductible Zakat certificates (BSZ).</li>
				<li>Enabling Customer Service AI (via LLM APIs) to respond accurately to donor queries.</li>
				<li>Sending donation receipt updates and campaign progress reports via email or WhatsApp.</li>
			</ul>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-white">4. Payment Gateways & Third-Party Processors</h2>
			<p>
				Financial transactions are processed directly by Midtrans for the initial MVP. We do not store full credit card details or bank passwords on our servers. Your financial information is handled securely by Midtrans under its privacy and security rules.
			</p>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-white">5. AI Processing (LLM Integrations)</h2>
			<p>
				When using AI features (such as WhatsApp Donation chats or AI CS), conversation texts may be securely processed through LLM provider APIs (e.g., OpenAI, Anthropic, or Google Gemini). We ensure that no highly sensitive personal identifiers (like passwords or full bank credentials) are sent to these AI processors.
			</p>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-white">6. Updates to This Policy</h2>
			<p>
				We may update this Privacy Policy from time to time. If we make material modifications, we will post the revised version on this page and update the "Last updated" date. We recommend checking this page regularly for updates.
			</p>
		</section>
	</div>
</div>
