import { useInfiniteQuery } from '@tanstack/react-query';
import { fetchArticles } from '../api/client';

export function useArticles(filters = {}) {
  return useInfiniteQuery({
    queryKey: ['articles', filters],
    queryFn: ({ pageParam = 0 }) =>
      fetchArticles({
        offset: pageParam,
        limit: 20,
        ...filters,
      }),
    getNextPageParam: (lastPage, allPages) => {
      if (!lastPage || lastPage.length < 20) return undefined;
      return allPages.length * 20;
    },
    initialPageParam: 0,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}
