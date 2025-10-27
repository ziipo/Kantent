import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchFeeds, createFeed, deleteFeed, refreshFeed } from '../api/client';

export function useFeeds() {
  const queryClient = useQueryClient();

  const feedsQuery = useQuery({
    queryKey: ['feeds'],
    queryFn: fetchFeeds,
  });

  const createMutation = useMutation({
    mutationFn: createFeed,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] });
      // Wait a bit before refreshing articles to let the backend fetch the feed
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ['articles'] });
      }, 2000);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteFeed,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] });
      queryClient.invalidateQueries({ queryKey: ['articles'] });
    },
  });

  const refreshMutation = useMutation({
    mutationFn: refreshFeed,
    onSuccess: () => {
      // Wait a bit before refreshing articles to let the backend fetch the feed
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ['articles'] });
      }, 2000);
    },
  });

  return {
    feeds: feedsQuery.data || [],
    isLoading: feedsQuery.isLoading,
    createFeed: createMutation.mutate,
    deleteFeed: deleteMutation.mutate,
    refreshFeed: refreshMutation.mutate,
    isCreating: createMutation.isPending,
  };
}
