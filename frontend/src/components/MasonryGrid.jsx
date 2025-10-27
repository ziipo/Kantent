import { useEffect } from 'react';
import { useInView } from 'react-intersection-observer';
import Masonry from 'react-masonry-css';
import ArticleCard from './ArticleCard';
import { useArticles } from '../hooks/useArticles';

const breakpointColumns = {
  default: 4,
  1536: 3,
  1024: 2,
  640: 1
};

export default function MasonryGrid({ onArticleClick, filterUnread, selectedFeed }) {
  const { ref, inView } = useInView();

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    status,
  } = useArticles({
    unread: filterUnread,
    feedId: selectedFeed,
  });

  useEffect(() => {
    if (inView && hasNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, fetchNextPage]);

  if (status === 'pending') {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {[...Array(8)].map((_, i) => (
            <div key={i} className="bg-gray-200 animate-pulse rounded-lg h-64"></div>
          ))}
        </div>
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center text-gray-500">
          <p>Error loading articles. Please try again later.</p>
        </div>
      </div>
    );
  }

  const articles = data?.pages.flatMap(page => page) ?? [];

  if (articles.length === 0) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center text-gray-500 py-12">
          <h3 className="text-xl font-semibold mb-2">No articles yet</h3>
          <p>Add some feeds to get started!</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Masonry
        breakpointCols={breakpointColumns}
        className="flex -ml-4 w-auto"
        columnClassName="pl-4 bg-clip-padding"
      >
        {articles.map((article) => (
          <ArticleCard
            key={article.id}
            article={article}
            onClick={() => onArticleClick(article)}
          />
        ))}
      </Masonry>

      {hasNextPage && (
        <div ref={ref} className="text-center py-8">
          {isFetchingNextPage ? (
            <div className="text-gray-500">Loading more...</div>
          ) : (
            <div className="text-gray-400">Scroll for more</div>
          )}
        </div>
      )}
    </div>
  );
}
