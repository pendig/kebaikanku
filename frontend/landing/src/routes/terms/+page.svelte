<script>
	import { t } from '$lib/i18n';
</script>

<svelte:head>
	<title>{$t('footer.terms')} | kebaikanku.id</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-6 py-12 md:py-20">
	<div class="space-y-6 border-b border-gray-200 dark:border-gray-900 pb-10">
		<h1 class="text-3xl md:text-5xl font-extrabold text-gray-900 dark:text-white">
			{$t('footer.terms')}
		</h1>
		<p class="text-sm text-gray-500 dark:text-gray-400">Last updated: June 9, 2026</p>
	</div>

	<div class="mt-10 prose dark:prose-invert prose-emerald max-w-none text-gray-700 dark:text-gray-300 space-y-8 text-sm md:text-base leading-relaxed">
		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">1. Acceptance of Terms</h2>
			<p>
				By accessing or using the kebaikanku.id platform or hosting your own instance using our open-source codebase, you agree to be bound by these Terms of Service. If you are using the platform on behalf of a registered zakat institution or non-profit foundation, you represent that you have the authority to bind that entity to these terms.
			</p>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">2. Open Source Licensing & Proprietary Use</h2>
			<p>
				The core kebaikanku.id codebase is open-source and licensed under the <strong>GNU Affero General Public License v3 (AGPL-3.0)</strong>. You are free to self-host, copy, modify, and distribute the software, provided that any network deployment (SaaS) of modified versions is also made publicly available under the AGPL-3.0 license.
			</p>
			<p>
				For organizations requiring custom white-label setups (removal of kebaikanku.id branding), closed-source extensions, or dedicated on-premise installation support, a separate <strong>Commercial Enterprise License</strong> must be acquired from the maintainers.
			</p>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">3. Platform Fees & Financial Transactions</h2>
			<ul class="list-disc pl-6 space-y-2">
				<li>
					<strong>Self-Hosted Instances:</strong> kebaikanku.id charges a <strong>0% platform fee</strong>. All donation payouts are routed directly through the Midtrans account configured by your organization. You are solely responsible for paying direct Midtrans transaction fees according to their terms.
				</li>
			</ul>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">4. Zakat & Donation Compliance</h2>
			<p>
				kebaikanku.id provides technical tools to calculate zakat and process donations. The registered organization is solely responsible for:
			</p>
			<ul class="list-disc pl-6 space-y-2">
				<li>Ensuring all campaigns conform to local regulations and Islamic jurisprudential (sharia) guidelines for zakat management.</li>
				<li>Verifying the identity of beneficiaries (Mustahik) and ensuring proper distribution of funds.</li>
				<li>Issuing valid tax deduction receipts (Bukti Setor Zakat) and maintaining legal operational permits.</li>
			</ul>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">5. AI Features & Accuracy (Roadmap)</h2>
			<p>
				AI-assisted customer service, zakat calculations, and report generation are roadmap items and are not provided by the current release. If released, their terms and safeguards will be documented before use.
			</p>
			<ul class="list-disc pl-6 space-y-2">
				<li>Any future AI-generated text or calculations must be verified by a human administrator before finalization.</li>
			</ul>
		</section>

		<section class="space-y-3">
			<h2 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">6. Limitation of Liability</h2>
			<p>
				IN NO EVENT SHALL KEBAIKANKU.ID, ITS MAINTAINERS, OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
			</p>
		</section>
	</div>
</div>
