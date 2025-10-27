import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';
import Header from './components/Header';
import MasonryGrid from './components/MasonryGrid';
import FeedManager from './components/FeedManager';
import ArticleModal from './components/ArticleModal';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  const [selectedArticle, setSelectedArticle] = useState(null);
  const [showFeedManager, setShowFeedManager] = useState(false);
  const [filterUnread, setFilterUnread] = useState(false);
  const [selectedFeed, setSelectedFeed] = useState(null);

  return (
    <QueryClientProvider client={queryClient}>
      <div className="min-h-screen bg-gray-50">
        <Header
          onToggleFeedManager={() => setShowFeedManager(!showFeedManager)}
          onToggleUnread={() => setFilterUnread(!filterUnread)}
          filterUnread={filterUnread}
        />

        {showFeedManager && (
          <FeedManager
            onClose={() => setShowFeedManager(false)}
            onSelectFeed={setSelectedFeed}
          />
        )}

        <MasonryGrid
          onArticleClick={setSelectedArticle}
          filterUnread={filterUnread}
          selectedFeed={selectedFeed}
        />

        {selectedArticle && (
          <ArticleModal
            article={selectedArticle}
            onClose={() => setSelectedArticle(null)}
          />
        )}
      </div>
    </QueryClientProvider>
  );
}

export default App;
