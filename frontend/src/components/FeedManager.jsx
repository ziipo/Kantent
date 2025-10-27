import { useState } from 'react';
import { useFeeds } from '../hooks/useFeeds';

export default function FeedManager({ onClose, onSelectFeed }) {
  const { feeds, isLoading, createFeed, deleteFeed, refreshFeed, isCreating } = useFeeds();
  const [activeTab, setActiveTab] = useState('rss'); // 'rss', 'reddit', or 'youtube'

  // RSS feed state
  const [newFeedUrl, setNewFeedUrl] = useState('');

  // Reddit feed state
  const [subreddit, setSubreddit] = useState('');
  const [sortBy, setSortBy] = useState('hot');

  // YouTube feed state
  const [youtubeInput, setYoutubeInput] = useState('');

  const handleAddRssFeed = async (e) => {
    e.preventDefault();
    if (!newFeedUrl.trim()) return;

    createFeed({ url: newFeedUrl, title: 'Loading...' });
    setNewFeedUrl('');
  };

  const handleAddRedditFeed = async (e) => {
    e.preventDefault();
    if (!subreddit.trim()) return;

    // Clean up subreddit name (remove r/ prefix if present)
    const cleanSubreddit = subreddit.replace(/^r\//, '').trim();

    // Reddit RSS feed URL format
    const redditUrl = `https://www.reddit.com/r/${cleanSubreddit}/${sortBy}.rss`;
    const feedTitle = `r/${cleanSubreddit} (${sortBy})`;

    createFeed({ url: redditUrl, title: feedTitle });
    setSubreddit('');
  };

  const extractYouTubeChannelId = (input) => {
    const trimmed = input.trim();

    // If it looks like a channel ID already (starts with UC and is 24 chars)
    if (/^UC[\w-]{22}$/.test(trimmed)) {
      return trimmed;
    }

    // Extract from various YouTube URL formats
    const patterns = [
      // https://www.youtube.com/channel/UCxxxxxx
      /youtube\.com\/channel\/(UC[\w-]{22})/,
      // https://www.youtube.com/@channelname -> need to handle differently
      /youtube\.com\/@([\w-]+)/,
      // https://www.youtube.com/c/channelname -> need to handle differently
      /youtube\.com\/c\/([\w-]+)/,
      // https://www.youtube.com/user/username -> need to handle differently
      /youtube\.com\/user\/([\w-]+)/,
    ];

    for (const pattern of patterns) {
      const match = trimmed.match(pattern);
      if (match) {
        return match[1];
      }
    }

    // If no pattern matched, assume it's a channel handle or username
    return trimmed;
  };

  const handleAddYouTubeFeed = async (e) => {
    e.preventDefault();
    if (!youtubeInput.trim()) return;

    const channelId = extractYouTubeChannelId(youtubeInput);

    // YouTube RSS feed URL format
    const youtubeUrl = `https://www.youtube.com/feeds/videos.xml?channel_id=${channelId}`;
    const feedTitle = `YouTube: ${channelId}`;

    createFeed({ url: youtubeUrl, title: feedTitle });
    setYoutubeInput('');
  };

  const handleRefresh = (feedId) => {
    refreshFeed(feedId);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-2xl w-full max-h-[80vh] overflow-hidden flex flex-col">
        <div className="p-6 border-b flex items-center justify-between">
          <h2 className="text-2xl font-bold">Manage Feeds</h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 text-2xl leading-none"
          >
            ×
          </button>
        </div>

        <div className="border-b">
          <div className="flex">
            <button
              onClick={() => setActiveTab('rss')}
              className={`flex-1 px-4 py-3 font-medium transition-colors ${
                activeTab === 'rss'
                  ? 'text-blue-600 border-b-2 border-blue-600'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              RSS Feed
            </button>
            <button
              onClick={() => setActiveTab('reddit')}
              className={`flex-1 px-4 py-3 font-medium transition-colors ${
                activeTab === 'reddit'
                  ? 'text-blue-600 border-b-2 border-blue-600'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              Reddit
            </button>
            <button
              onClick={() => setActiveTab('youtube')}
              className={`flex-1 px-4 py-3 font-medium transition-colors ${
                activeTab === 'youtube'
                  ? 'text-blue-600 border-b-2 border-blue-600'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              YouTube
            </button>
          </div>
        </div>

        <div className="p-6 border-b">
          {activeTab === 'rss' ? (
            <form onSubmit={handleAddRssFeed} className="flex gap-2">
              <input
                type="url"
                placeholder="Enter feed URL (e.g., https://example.com/feed.xml)"
                value={newFeedUrl}
                onChange={(e) => setNewFeedUrl(e.target.value)}
                className="flex-1 px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
              <button
                type="submit"
                disabled={isCreating}
                className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isCreating ? 'Adding...' : 'Add Feed'}
              </button>
            </form>
          ) : activeTab === 'reddit' ? (
            <form onSubmit={handleAddRedditFeed} className="space-y-3">
              <div className="flex gap-2">
                <div className="flex-1">
                  <input
                    type="text"
                    placeholder="Subreddit name (e.g., technology or r/technology)"
                    value={subreddit}
                    onChange={(e) => setSubreddit(e.target.value)}
                    className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>
                <select
                  value={sortBy}
                  onChange={(e) => setSortBy(e.target.value)}
                  className="px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
                >
                  <option value="hot">Hot</option>
                  <option value="new">New</option>
                  <option value="top">Top</option>
                  <option value="rising">Rising</option>
                </select>
              </div>
              <button
                type="submit"
                disabled={isCreating}
                className="w-full px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isCreating ? 'Adding...' : 'Add Reddit Feed'}
              </button>
            </form>
          ) : (
            <form onSubmit={handleAddYouTubeFeed} className="space-y-3">
              <input
                type="text"
                placeholder="Channel URL or ID (e.g., https://youtube.com/@channelname or UCxxxxx)"
                value={youtubeInput}
                onChange={(e) => setYoutubeInput(e.target.value)}
                className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
              <p className="text-xs text-gray-500">
                Paste any YouTube channel URL, or just the channel ID (starts with UC)
              </p>
              <button
                type="submit"
                disabled={isCreating}
                className="w-full px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isCreating ? 'Adding...' : 'Add YouTube Feed'}
              </button>
            </form>
          )}
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {isLoading ? (
            <div className="text-center py-8 text-gray-500">Loading feeds...</div>
          ) : feeds.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No feeds yet. Add one above to get started!
            </div>
          ) : (
            <div className="space-y-2">
              {feeds.map((feed) => (
                <div
                  key={feed.id}
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50"
                >
                  <div
                    className="flex-1 cursor-pointer"
                    onClick={() => {
                      onSelectFeed(feed.id);
                      onClose();
                    }}
                  >
                    <h3 className="font-semibold">{feed.title}</h3>
                    <p className="text-sm text-gray-500 truncate">{feed.url}</p>
                    {feed.last_error && (
                      <p className="text-xs text-red-500 mt-1">Error: {feed.last_error}</p>
                    )}
                  </div>
                  <div className="flex gap-2 ml-4">
                    <button
                      onClick={() => handleRefresh(feed.id)}
                      className="px-3 py-1 text-blue-600 hover:bg-blue-50 rounded"
                      title="Refresh feed"
                    >
                      ↻
                    </button>
                    <button
                      onClick={() => {
                        if (window.confirm(`Delete feed "${feed.title}"?`)) {
                          deleteFeed(feed.id);
                        }
                      }}
                      className="px-3 py-1 text-red-600 hover:bg-red-50 rounded"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
