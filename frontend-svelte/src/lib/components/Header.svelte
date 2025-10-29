<script>
	import { showFeedManager, stats } from '../stores/ui';
	import { setFilters, filters } from '../stores/articles';

	export let currentFilters = { unread: false, feedId: null };

	$: isUnreadActive = currentFilters.unread;

	function handleShowAll() {
		setFilters({ unread: false, feedId: null });
	}

	function handleShowUnread() {
		setFilters({ unread: true, feedId: null });
	}
</script>

<header class="sticky top-0 z-40 bg-white border-b">
	<div class="container mx-auto px-4 py-4">
		<div class="flex items-center justify-between mb-4">
			<h1 class="text-2xl font-bold">Kantent</h1>
			<button
				on:click={() => ($showFeedManager = true)}
				class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
			>
				Manage Feeds
			</button>
		</div>

		<div class="flex items-center gap-4">
			<button
				on:click={handleShowAll}
				class="px-4 py-2 rounded-lg transition-colors {!isUnreadActive
					? 'bg-blue-500 text-white'
					: 'bg-gray-200 hover:bg-gray-300'}"
			>
				All ({$stats.totalArticles})
			</button>
			<button
				on:click={handleShowUnread}
				class="px-4 py-2 rounded-lg transition-colors {isUnreadActive
					? 'bg-blue-500 text-white'
					: 'bg-gray-200 hover:bg-gray-300'}"
			>
				Unread ({$stats.unreadCount})
			</button>
		</div>
	</div>
</header>
