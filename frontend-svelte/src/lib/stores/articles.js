import { writable, derived } from 'svelte/store';
import { fetchArticles as apiFetchArticles } from '../api/client';

// Articles state
export const articles = writable([]);
export const isLoading = writable(false);
export const hasMore = writable(true);
export const filters = writable({ unread: false, feedId: null });

// Load articles
export async function loadArticles(reset = false) {
	const currentFilters = get(filters);
	const currentArticles = get(articles);

	if (reset) {
		articles.set([]);
		hasMore.set(true);
	}

	isLoading.set(true);

	try {
		const offset = reset ? 0 : currentArticles.length;
		const data = await apiFetchArticles({
			offset,
			limit: 12,
			...currentFilters
		});

		if (reset) {
			articles.set(data);
		} else {
			articles.update(a => [...a, ...data]);
		}

		if (data.length < 12) {
			hasMore.set(false);
		}
	} catch (error) {
		console.error('Failed to load articles:', error);
	} finally {
		isLoading.set(false);
	}
}

// Update filters and reload
export function setFilters(newFilters) {
	filters.set(newFilters);
	loadArticles(true);
}

// Update a single article
export function updateArticle(articleId, updates) {
	articles.update(arts =>
		arts.map(a => a.id === articleId ? { ...a, ...updates } : a)
	);
}

// Helper to get current value
function get(store) {
	let value;
	store.subscribe(v => value = v)();
	return value;
}
