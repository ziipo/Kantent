import { writable } from 'svelte/store';

export const selectedArticle = writable(null);
export const showFeedManager = writable(false);
export const stats = writable({ totalFeeds: 0, totalArticles: 0, unreadCount: 0 });
