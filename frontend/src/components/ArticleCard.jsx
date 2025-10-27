import { useState } from 'react';
import { formatDistanceToNow } from 'date-fns';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { markAsRead } from '../api/client';

export default function ArticleCard({ article, onClick }) {
  const [imageError, setImageError] = useState(false);
  const queryClient = useQueryClient();

  const markReadMutation = useMutation({
    mutationFn: (isRead) => markAsRead(article.id, isRead),
    onMutate: async (isRead) => {
      // Optimistic update
      await queryClient.cancelQueries({ queryKey: ['articles'] });

      const previousData = queryClient.getQueryData(['articles']);

      queryClient.setQueryData(['articles'], (old) => {
        if (!old) return old;
        return {
          ...old,
          pages: old.pages.map(page =>
            page.map(a => a.id === article.id ? { ...a, is_read: isRead } : a)
          ),
        };
      });

      return { previousData };
    },
    onError: (err, variables, context) => {
      if (context?.previousData) {
        queryClient.setQueryData(['articles'], context.previousData);
      }
    },
  });

  const handleClick = () => {
    if (!article.is_read) {
      markReadMutation.mutate(true);
    }
    onClick();
  };

  return (
    <div
      className={`mb-4 break-inside-avoid cursor-pointer group transition-opacity ${
        article.is_read ? 'opacity-60 hover:opacity-80' : 'hover:opacity-90'
      }`}
      onClick={handleClick}
    >
      <div className="bg-white rounded-lg shadow-sm hover:shadow-lg transition-shadow overflow-hidden">
        {!imageError && article.image_url && (
          <div className="relative overflow-hidden bg-gray-100">
            <img
              src={article.image_url}
              alt=""
              className="w-full h-auto object-cover group-hover:scale-105 transition-transform duration-300"
              loading="lazy"
              onError={() => setImageError(true)}
            />
          </div>
        )}

        <div className="p-4">
          <h3 className="font-semibold text-lg mb-2 line-clamp-3 text-gray-900">
            {article.title}
          </h3>

          {article.description && (
            <p className="text-gray-600 text-sm mb-3 line-clamp-2">
              {article.description}
            </p>
          )}

          <div className="flex items-center justify-between text-xs text-gray-500">
            <span className="truncate mr-2">{article.feed_title}</span>
            <span className="whitespace-nowrap">
              {formatDistanceToNow(new Date(article.published_at), {
                addSuffix: true,
              })}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
