<script>
	import { selectedArticle } from '../stores/ui';
	import { markAsRead, starArticle } from '../api/client';
	import { updateArticle } from '../stores/articles';

	function close() {
		$selectedArticle = null;
	}

	function handleBackdropClick(e) {
		if (e.target === e.currentTarget) {
			close();
		}
	}

	async function toggleRead() {
		if (!$selectedArticle) return;
		const newReadState = !$selectedArticle.is_read;
		const articleId = $selectedArticle.id;

		$selectedArticle = { ...$selectedArticle, is_read: newReadState };
		updateArticle(articleId, { is_read: newReadState });

		try {
			await markAsRead(articleId, newReadState);
		} catch (error) {
			$selectedArticle = { ...$selectedArticle, is_read: !newReadState };
			updateArticle(articleId, { is_read: !newReadState });
		}
	}

	async function toggleStar() {
		if (!$selectedArticle) return;
		const newStarState = !$selectedArticle.is_starred;
		const articleId = $selectedArticle.id;

		$selectedArticle = { ...$selectedArticle, is_starred: newStarState };
		updateArticle(articleId, { is_starred: newStarState });

		try {
			await starArticle(articleId, newStarState);
		} catch (error) {
			$selectedArticle = { ...$selectedArticle, is_starred: !newStarState };
			updateArticle(articleId, { is_starred: !newStarState });
		}
	}

	function formatDate(dateString) {
		const date = new Date(dateString);
		return date.toLocaleString();
	}
</script>

{#if $selectedArticle}
	<div
		class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
		on:click={handleBackdropClick}
		on:keypress
		role="button"
		tabindex="-1"
	>
		<div class="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
			<!-- Header -->
			<div class="p-6 border-b flex items-center justify-between">
				<h2 class="text-2xl font-bold flex-1">{$selectedArticle.title}</h2>
				<button
					on:click={close}
					class="text-gray-500 hover:text-gray-700 text-2xl leading-none ml-4"
				>
					×
				</button>
			</div>

			<!-- Content -->
			<div class="flex-1 overflow-y-auto p-6">
				<div class="mb-4 text-sm text-gray-600">
					<div class="font-semibold">{$selectedArticle.feed_title}</div>
					{#if $selectedArticle.author}
						<div>By {$selectedArticle.author}</div>
					{/if}
					<div>{formatDate($selectedArticle.published_at)}</div>
				</div>

				{#if $selectedArticle.image_url}
					<img
						src={$selectedArticle.image_url}
						alt={$selectedArticle.title}
						class="w-full h-auto rounded-lg mb-4"
					/>
				{/if}

				{#if $selectedArticle.content}
					<div class="prose max-w-none">
						{@html $selectedArticle.content}
					</div>
				{:else if $selectedArticle.description}
					<p class="text-gray-700">{$selectedArticle.description}</p>
				{/if}

				<div class="mt-6">
					<a
						href={$selectedArticle.url}
						target="_blank"
						rel="noopener noreferrer"
						class="text-blue-500 hover:text-blue-700"
					>
						Read original article →
					</a>
				</div>
			</div>

			<!-- Footer -->
			<div class="p-6 border-t flex justify-between items-center">
				<div class="flex gap-2">
					<button
						on:click={toggleRead}
						class="px-4 py-2 rounded-lg border hover:bg-gray-100 transition-colors"
					>
						{$selectedArticle.is_read ? '✓ Read' : 'Mark as Read'}
					</button>
					<button
						on:click={toggleStar}
						class="px-4 py-2 rounded-lg border hover:bg-gray-100 transition-colors"
					>
						{$selectedArticle.is_starred ? '★ Starred' : '☆ Star'}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
