import { formatDistanceToNow } from 'date-fns';

export default function ArticleModal({ article, onClose }) {
  if (!article) return null;

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6 border-b flex items-start justify-between">
          <div className="flex-1">
            <h2 className="text-2xl font-bold mb-2">{article.title}</h2>
            <div className="flex items-center gap-4 text-sm text-gray-500">
              <span>{article.feed_title}</span>
              <span>•</span>
              <span>
                {formatDistanceToNow(new Date(article.published_at), {
                  addSuffix: true,
                })}
              </span>
              {article.author && (
                <>
                  <span>•</span>
                  <span>By {article.author}</span>
                </>
              )}
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 text-2xl leading-none ml-4"
          >
            ×
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {article.image_url && (
            <img
              src={article.image_url}
              alt={article.title}
              className="w-full h-auto rounded-lg mb-6"
              onError={(e) => {
                e.target.style.display = 'none';
              }}
            />
          )}

          {article.content ? (
            <div
              className="prose prose-lg max-w-none"
              dangerouslySetInnerHTML={{ __html: article.content }}
            />
          ) : (
            <div className="prose prose-lg max-w-none">
              <p>{article.description}</p>
            </div>
          )}
        </div>

        <div className="p-6 border-t bg-gray-50">
          <a
            href={article.url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-block px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
          >
            Read Original Article →
          </a>
        </div>
      </div>
    </div>
  );
}
