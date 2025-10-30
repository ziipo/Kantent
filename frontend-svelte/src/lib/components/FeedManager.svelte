<script>
	import { showFeedManager } from '../stores/ui';
	import { feeds, isLoadingFeeds, isCreatingFeed, loadFeeds, createFeed, deleteFeed, refreshFeed } from '../stores/feeds';
	import { setFilters } from '../stores/articles';
	import { discoverFeeds, resolveYouTubeChannel } from '../api/client';

	let activeTab = 'discover';
	let newFeedUrl = '';
	let subreddit = '';
	let sortBy = 'hot';
	let youtubeInput = '';
	let discoverUrl = '';
	let discoveredFeeds = [];
	let isDiscovering = false;

	function close() {
		$showFeedManager = false;
	}

	function handleBackdropClick(e) {
		if (e.target === e.currentTarget) {
			close();
		}
	}

	async function handleAddRssFeed(e) {
		e.preventDefault();
		if (!newFeedUrl.trim()) return;

		try {
			await createFeed({ url: newFeedUrl, title: 'Loading...' });
			newFeedUrl = '';
		} catch (error) {
			alert('Failed to add feed: ' + error.message);
		}
	}

	async function handleAddRedditFeed(e) {
		e.preventDefault();
		if (!subreddit.trim()) return;

		const cleanSubreddit = subreddit.replace(/^r\//, '').trim();
		const redditUrl = `https://www.reddit.com/r/${cleanSubreddit}/${sortBy}.rss`;
		const feedTitle = `r/${cleanSubreddit} (${sortBy})`;

		try {
			await createFeed({ url: redditUrl, title: feedTitle });
			subreddit = '';
		} catch (error) {
			alert('Failed to add Reddit feed: ' + error.message);
		}
	}

	async function handleAddYouTubeFeed(e) {
		e.preventDefault();
		if (!youtubeInput.trim()) return;

		try {
			const result = await resolveYouTubeChannel(youtubeInput);
			const feedTitle = `YouTube: ${youtubeInput}`;

			await createFeed({ url: result.rss_url, title: feedTitle });
			youtubeInput = '';
		} catch (error) {
			alert('Failed to resolve YouTube channel: ' + error.message);
		}
	}

	async function handleDiscoverFeeds(e) {
		e.preventDefault();
		if (!discoverUrl.trim()) return;

		isDiscovering = true;
		try {
			const feeds = await discoverFeeds(discoverUrl);
			discoveredFeeds = feeds;
		} catch (error) {
			alert('Failed to discover feeds: ' + error.message);
		} finally {
			isDiscovering = false;
		}
	}

	async function handleAddDiscoveredFeed(feedUrl) {
		try {
			await createFeed({ url: feedUrl, title: 'Loading...' });
		} catch (error) {
			alert('Failed to add feed: ' + error.message);
		}
	}

	async function handleRefresh(feedId) {
		await refreshFeed(feedId);
	}

	async function handleDelete(feed) {
		if (confirm(`Delete feed "${feed.title}"?`)) {
			await deleteFeed(feed.id);
		}
	}

	function handleSelectFeed(feedId) {
		setFilters({ unread: false, feedId });
		close();
	}
</script>

{#if $showFeedManager}
	<div
		class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
		on:click={handleBackdropClick}
		on:keypress
		role="button"
		tabindex="-1"
	>
		<div class="bg-white rounded-lg max-w-2xl w-full max-h-[80vh] overflow-hidden flex flex-col">
			<!-- Header -->
			<div class="p-6 border-b flex items-center justify-between">
				<h2 class="text-2xl font-bold">Manage Feeds</h2>
				<button
					on:click={close}
					class="text-gray-500 hover:text-gray-700 text-2xl leading-none"
				>
					×
				</button>
			</div>

			<!-- Tabs -->
			<div class="border-b overflow-x-auto">
				<div class="flex min-w-max">
					<button
						on:click={() => (activeTab = 'discover')}
						class="flex-1 px-3 py-3 font-medium transition-colors whitespace-nowrap {activeTab ===
						'discover'
							? 'text-blue-600 border-b-2 border-blue-600'
							: 'text-gray-500 hover:text-gray-700'}"
					>
						Discover
					</button>
					<button
						on:click={() => (activeTab = 'rss')}
						class="flex-1 px-3 py-3 font-medium transition-colors whitespace-nowrap {activeTab ===
						'rss'
							? 'text-blue-600 border-b-2 border-blue-600'
							: 'text-gray-500 hover:text-gray-700'}"
					>
						RSS/Atom
					</button>
					<button
						on:click={() => (activeTab = 'reddit')}
						class="flex-1 px-3 py-3 font-medium transition-colors whitespace-nowrap {activeTab ===
						'reddit'
							? 'text-blue-600 border-b-2 border-blue-600'
							: 'text-gray-500 hover:text-gray-700'}"
					>
						Reddit
					</button>
					<button
						on:click={() => (activeTab = 'youtube')}
						class="flex-1 px-3 py-3 font-medium transition-colors whitespace-nowrap {activeTab ===
						'youtube'
							? 'text-blue-600 border-b-2 border-blue-600'
							: 'text-gray-500 hover:text-gray-700'}"
					>
						YouTube
					</button>
				</div>
			</div>

			<!-- Tab Content -->
			<div class="p-6 border-b">
				{#if activeTab === 'discover'}
					<div class="space-y-3">
						<form on:submit={handleDiscoverFeeds} class="space-y-3">
							<input
								type="url"
								placeholder="Enter website URL (e.g., https://example.com)"
								bind:value={discoverUrl}
								class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
								required
							/>
							<button
								type="submit"
								disabled={isDiscovering}
								class="w-full px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
							>
								{isDiscovering ? 'Discovering...' : 'Discover Feeds'}
							</button>
						</form>

						{#if discoveredFeeds.length > 0}
							<div class="mt-4 space-y-2">
								<h3 class="font-semibold text-sm text-gray-700">
									Found {discoveredFeeds.length} feed(s):
								</h3>
								{#each discoveredFeeds as feed}
									<div
										class="flex items-center justify-between p-3 border rounded-lg bg-gray-50"
									>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium truncate">{feed.title || 'Untitled Feed'}</p>
											<p class="text-xs text-gray-500 truncate">{feed.url}</p>
											<span class="text-xs text-gray-400">{feed.type}</span>
										</div>
										<button
											on:click={() => handleAddDiscoveredFeed(feed.url)}
											disabled={$isCreatingFeed}
											class="ml-3 px-3 py-1 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
										>
											Add
										</button>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{:else if activeTab === 'rss'}
					<form on:submit={handleAddRssFeed} class="space-y-3">
						<input
							type="url"
							placeholder="Enter RSS or Atom feed URL (e.g., https://example.com/feed.xml)"
							bind:value={newFeedUrl}
							class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
							required
						/>
						<p class="text-xs text-gray-500">
							Supports RSS 1.0/2.0, Atom, and JSON Feed formats
						</p>
						<button
							type="submit"
							disabled={$isCreatingFeed}
							class="w-full px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{$isCreatingFeed ? 'Adding...' : 'Add Feed'}
						</button>
					</form>
				{:else if activeTab === 'reddit'}
					<form on:submit={handleAddRedditFeed} class="space-y-3">
						<div class="flex gap-2">
							<input
								type="text"
								placeholder="Subreddit name (e.g., technology or r/technology)"
								bind:value={subreddit}
								class="flex-1 px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
								required
							/>
							<select
								bind:value={sortBy}
								class="px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
							>
								<option value="hot">Hot</option>
								<option value="new">New</option>
								<option value="top">Top</option>
								<option value="rising">Rising</option>
							</select>
						</div>
						<button
							type="submit"
							disabled={$isCreatingFeed}
							class="w-full px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{$isCreatingFeed ? 'Adding...' : 'Add Reddit Feed'}
						</button>
					</form>
				{:else}
					<form on:submit={handleAddYouTubeFeed} class="space-y-3">
						<input
							type="text"
							placeholder="Channel URL or ID (e.g., https://youtube.com/@channelname or UCxxxxx)"
							bind:value={youtubeInput}
							class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
							required
						/>
						<p class="text-xs text-gray-500">
							Paste any YouTube channel URL, or just the channel ID (starts with UC)
						</p>
						<button
							type="submit"
							disabled={$isCreatingFeed}
							class="w-full px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{$isCreatingFeed ? 'Adding...' : 'Add YouTube Feed'}
						</button>
					</form>
				{/if}
			</div>

			<!-- Feed List -->
			<div class="flex-1 overflow-y-auto p-6">
				{#if $isLoadingFeeds}
					<div class="text-center py-8 text-gray-500">Loading feeds...</div>
				{:else if $feeds.length === 0}
					<div class="text-center py-8 text-gray-500">
						No feeds yet. Add one above to get started!
					</div>
				{:else}
					<div class="space-y-2">
						{#each $feeds as feed (feed.id)}
							<div
								class="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50"
							>
								<div
									class="flex-1 min-w-0 cursor-pointer"
									on:click={() => handleSelectFeed(feed.id)}
									on:keypress
									role="button"
									tabindex="0"
								>
									<h3 class="font-semibold truncate">{feed.title}</h3>
									<p class="text-sm text-gray-500 truncate">{feed.url}</p>
									{#if feed.last_error}
										<p class="text-xs text-red-500 mt-1 truncate">Error: {feed.last_error}</p>
									{/if}
								</div>
								<div class="flex gap-2 ml-4 flex-shrink-0">
									<button
										on:click={() => handleRefresh(feed.id)}
										class="px-3 py-1 text-blue-600 hover:bg-blue-50 rounded"
										title="Refresh feed"
									>
										↻
									</button>
									<button
										on:click={() => handleDelete(feed)}
										class="px-3 py-1 text-red-600 hover:bg-red-50 rounded"
										title="Delete feed"
									>
										Delete
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
