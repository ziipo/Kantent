<script>
	import { onMount } from 'svelte';
	import ArticleCard from './ArticleCard.svelte';
	import { articles, isLoading, hasMore, loadArticles } from '../stores/articles';

	export let onArticleClick;

	let sentinel;
	let observer;
	let columnCount = 4;
	let containerWidth = 0;

	// Distribute articles into columns
	$: columns = distributeIntoColumns($articles, columnCount);

	function distributeIntoColumns(items, numColumns) {
		const cols = Array.from({ length: numColumns }, () => []);
		items.forEach((item, index) => {
			cols[index % numColumns].push(item);
		});
		return cols;
	}

	function updateColumnCount() {
		if (typeof window !== 'undefined') {
			const width = window.innerWidth;
			if (width >= 1280) columnCount = 4;
			else if (width >= 1024) columnCount = 3;
			else if (width >= 640) columnCount = 2;
			else columnCount = 1;
		}
	}

	onMount(() => {
		updateColumnCount();

		const resizeObserver = new ResizeObserver(() => {
			updateColumnCount();
		});

		if (typeof window !== 'undefined') {
			resizeObserver.observe(document.body);
		}

		// Set up intersection observer for infinite scroll
		observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting && $hasMore && !$isLoading) {
					loadArticles(false);
				}
			},
			{ threshold: 0.1 }
		);

		if (sentinel) {
			observer.observe(sentinel);
		}

		return () => {
			if (observer) observer.disconnect();
			resizeObserver.disconnect();
		};
	});
</script>

<div class="container mx-auto px-4 py-6">
	{#if $articles.length === 0 && !$isLoading}
		<div class="text-center py-12 text-gray-500">
			No articles yet. Add some feeds to get started!
		</div>
	{:else}
		<div class="masonry-grid" style="--column-count: {columnCount}">
			{#each columns as column, i (i)}
				<div class="masonry-column">
					{#each column as article (article.id)}
						<div class="masonry-item">
							<ArticleCard {article} on:click={() => onArticleClick(article)} />
						</div>
					{/each}
				</div>
			{/each}
		</div>
	{/if}

	{#if $isLoading}
		<div class="text-center py-8">
			<div class="inline-block animate-pulse text-gray-500">Loading...</div>
		</div>
	{/if}

	<!-- Sentinel for infinite scroll -->
	<div bind:this={sentinel} class="h-4"></div>
</div>

<style>
	.masonry-grid {
		display: grid;
		grid-template-columns: repeat(var(--column-count), 1fr);
		gap: 1rem;
	}

	.masonry-column {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.masonry-item {
		break-inside: avoid;
	}
</style>
