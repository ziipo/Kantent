import { useQuery } from '@tanstack/react-query';
import { fetchStats } from '../api/client';

export default function Header({ onToggleFeedManager, onToggleUnread, filterUnread }) {
  const { data: stats } = useQuery({
    queryKey: ['stats'],
    queryFn: fetchStats,
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  return (
    <header className="bg-white shadow-sm sticky top-0 z-40">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-2xl font-bold text-gray-900">Kantent</h1>
            {stats && (
              <div className="flex gap-4 text-sm text-gray-600">
                <span>{stats.total_feeds} feeds</span>
                <span>•</span>
                <span>{stats.unread_count} unread</span>
              </div>
            )}
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={onToggleUnread}
              className={`px-4 py-2 rounded-lg transition-colors ${
                filterUnread
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {filterUnread ? 'All Articles' : 'Unread Only'}
            </button>
            <button
              onClick={onToggleFeedManager}
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
            >
              Manage Feeds
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}
