<script>
	import { markAsRead, starArticle } from '../api/client';
	import { updateArticle } from '../stores/articles';

	export let article;

	async function toggleRead(e) {
		e.stopPropagation();
		const newReadState = !article.is_read;
		updateArticle(article.id, { is_read: newReadState });
		try {
			await markAsRead(article.id, newReadState);
		} catch (error) {
			updateArticle(article.id, { is_read: !newReadState });
		}
	}

	async function toggleStar(e) {
		e.stopPropagation();
		const newStarState = !article.is_starred;
		updateArticle(article.id, { is_starred: newStarState });
		try {
			await starArticle(article.id, newStarState);
		} catch (error) {
			updateArticle(article.id, { is_starred: !newStarState });
		}
	}

	function formatDate(dateString) {
		const date = new Date(dateString);
		return date.toLocaleDateString();
	}
</script>

<div
	class="bg-white rounded-lg shadow-sm hover:shadow-lg transition-shadow cursor-pointer group h-full"
	on:click
	on:keypress
	role="button"
	tabindex="0"
>
	{#if article.image_url}
		<div class="overflow-hidden rounded-t-lg">
			<img
				src={article.image_url}
				alt={article.title}
				class="w-full h-auto object-cover group-hover:scale-105 transition-transform duration-300"
				loading="lazy"
			/>
		</div>
	{/if}

	<div class="p-4">
		<h3 class="font-semibold text-lg mb-2 {article.is_read ? 'opacity-60' : ''}">{article.title}</h3>

		{#if article.description}
			<p class="text-sm text-gray-600 mb-3 line-clamp-3">{article.description}</p>
		{/if}

		<div class="flex items-center justify-between text-xs text-gray-500">
			<div class="flex flex-col gap-1">
				<span class="font-medium">{article.feed_title}</span>
				<span>{formatDate(article.published_at)}</span>
			</div>

			<div class="flex gap-2">
				<button
					on:click={toggleRead}
					class="px-3 py-1 rounded hover:bg-gray-200 transition-colors"
					title={article.is_read ? 'Mark as unread' : 'Mark as read'}
				>
					{article.is_read ? '✓' : '○'}
				</button>
				<button
					on:click={toggleStar}
					class="px-3 py-1 rounded hover:bg-gray-200 transition-colors"
					title={article.is_starred ? 'Unstar' : 'Star'}
				>
					{article.is_starred ? '★' : '☆'}
				</button>
			</div>
		</div>
	</div>
</div>
