const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

async function apiRequest(endpoint, options = {}) {
	const response = await fetch(`${API_URL}${endpoint}`, {
		headers: {
			'Content-Type': 'application/json',
			...options.headers
		},
		...options
	});

	if (!response.ok) {
		throw new Error(`API error: ${response.statusText}`);
	}

	// Handle 204 No Content
	if (response.status === 204) {
		return null;
	}

	return response.json();
}

export const fetchArticles = ({ offset = 0, limit = 20, unread = false, feedId = null }) => {
	const params = new URLSearchParams({
		offset: offset.toString(),
		limit: limit.toString()
	});

	if (unread) params.append('unread', 'true');
	if (feedId) params.append('feed_id', feedId);

	return apiRequest(`/api/articles?${params}`);
};

export const fetchArticle = (articleId) => apiRequest(`/api/articles/${articleId}`);

export const fetchFeeds = () => apiRequest('/api/feeds');

export const createFeed = (feedData) =>
	apiRequest('/api/feeds', {
		method: 'POST',
		body: JSON.stringify(feedData)
	});

export const deleteFeed = (feedId) =>
	apiRequest(`/api/feeds/${feedId}`, {
		method: 'DELETE'
	});

export const refreshFeed = (feedId) =>
	apiRequest(`/api/feeds/${feedId}/refresh`, {
		method: 'POST'
	});

export const markAsRead = (articleId, isRead) =>
	apiRequest(`/api/articles/${articleId}/read`, {
		method: 'PUT',
		body: JSON.stringify({ is_read: isRead })
	});

export const starArticle = (articleId, isStarred) =>
	apiRequest(`/api/articles/${articleId}/star`, {
		method: 'PUT',
		body: JSON.stringify({ is_starred: isStarred })
	});

export const markAllRead = (feedId = null) => {
	const params = feedId ? `?feed_id=${feedId}` : '';
	return apiRequest(`/api/articles/mark-all-read${params}`, {
		method: 'POST'
	});
};

export const fetchStats = () => apiRequest('/api/stats');

export const discoverFeeds = (websiteUrl) => {
	const params = new URLSearchParams({ url: websiteUrl });
	return apiRequest(`/api/discover?${params}`);
};

export const resolveYouTubeChannel = (input) => {
	const params = new URLSearchParams({ input });
	return apiRequest(`/api/youtube/resolve?${params}`);
};
