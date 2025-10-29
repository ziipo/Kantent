import { writable } from 'svelte/store';
import {
	fetchFeeds as apiFetchFeeds,
	createFeed as apiCreateFeed,
	deleteFeed as apiDeleteFeed,
	refreshFeed as apiRefreshFeed
} from '../api/client';
import { loadArticles } from './articles';

export const feeds = writable([]);
export const isLoadingFeeds = writable(false);
export const isCreatingFeed = writable(false);

// Load all feeds
export async function loadFeeds() {
	isLoadingFeeds.set(true);
	try {
		const data = await apiFetchFeeds();
		feeds.set(data);
	} catch (error) {
		console.error('Failed to load feeds:', error);
	} finally {
		isLoadingFeeds.set(false);
	}
}

// Create a new feed
export async function createFeed(feedData) {
	isCreatingFeed.set(true);
	try {
		await apiCreateFeed(feedData);
		await loadFeeds();
		await loadArticles(true);
	} catch (error) {
		console.error('Failed to create feed:', error);
		throw error;
	} finally {
		isCreatingFeed.set(false);
	}
}

// Delete a feed
export async function deleteFeed(feedId) {
	try {
		await apiDeleteFeed(feedId);
		await loadFeeds();
		await loadArticles(true);
	} catch (error) {
		console.error('Failed to delete feed:', error);
	}
}

// Refresh a feed
export async function refreshFeed(feedId) {
	try {
		await apiRefreshFeed(feedId);
		await loadArticles(true);
	} catch (error) {
		console.error('Failed to refresh feed:', error);
	}
}
