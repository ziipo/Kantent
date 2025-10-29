<script>
	import { onMount } from 'svelte';
	import Header from '../lib/components/Header.svelte';
	import MasonryGrid from '../lib/components/MasonryGrid.svelte';
	import ArticleModal from '../lib/components/ArticleModal.svelte';
	import FeedManager from '../lib/components/FeedManager.svelte';
	import { selectedArticle } from '../lib/stores/ui';
	import { loadArticles, filters } from '../lib/stores/articles';
	import { loadFeeds } from '../lib/stores/feeds';
	import { fetchStats } from '../lib/api/client';
	import { stats } from '../lib/stores/ui';
	import { markAsRead } from '../lib/api/client';
	import { updateArticle } from '../lib/stores/articles';

	onMount(async () => {
		// Load initial data
		await Promise.all([loadArticles(true), loadFeeds(), loadStats()]);
	});

	async function loadStats() {
		try {
			const data = await fetchStats();
			$stats = data;
		} catch (error) {
			console.error('Failed to load stats:', error);
		}
	}

	async function handleArticleClick(article) {
		$selectedArticle = article;

		// Mark as read when opening
		if (!article.is_read) {
			updateArticle(article.id, { is_read: true });
			try {
				await markAsRead(article.id, true);
				// Reload stats after marking as read
				await loadStats();
			} catch (error) {
				updateArticle(article.id, { is_read: false });
			}
		}
	}
</script>

<svelte:head>
	<title>Kantent - RSS Reader</title>
</svelte:head>

<div class="min-h-screen bg-gray-100">
	<Header currentFilters={$filters} />
	<MasonryGrid onArticleClick={handleArticleClick} />
	<ArticleModal />
	<FeedManager />
</div>
