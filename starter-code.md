# Starter Code Examples

Quick-start code snippets for the RSS Reader project.

---

## Backend: Database Setup

### backend/db/sqlite.go

```go
package db

import (
    "database/sql"
    "log"
    
    _ "github.com/mattn/go-sqlite3"
)

func InitDatabase(path string) *sql.DB {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        log.Fatal(err)
    }
    
    // Enable WAL mode for better concurrency
    db.Exec("PRAGMA journal_mode=WAL")
    db.Exec("PRAGMA synchronous=NORMAL")
    db.Exec("PRAGMA cache_size=-64000")
    db.Exec("PRAGMA foreign_keys=ON")
    
    // Run migrations
    runMigrations(db)
    
    return db
}

func runMigrations(db *sql.DB) {
    migrations := []string{
        `CREATE TABLE IF NOT EXISTS feeds (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            url TEXT UNIQUE NOT NULL,
            site_url TEXT,
            description TEXT,
            fetch_interval INTEGER DEFAULT 1800,
            last_fetched TIMESTAMP,
            last_error TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
        
        `CREATE TABLE IF NOT EXISTS articles (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            feed_id INTEGER NOT NULL,
            guid TEXT UNIQUE NOT NULL,
            title TEXT NOT NULL,
            url TEXT NOT NULL,
            description TEXT,
            content TEXT,
            author TEXT,
            published_at TIMESTAMP NOT NULL,
            fetched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            is_read BOOLEAN DEFAULT FALSE,
            is_starred BOOLEAN DEFAULT FALSE,
            image_url TEXT,
            FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
        )`,
        
        `CREATE INDEX IF NOT EXISTS idx_articles_feed_id ON articles(feed_id)`,
        `CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC)`,
        `CREATE INDEX IF NOT EXISTS idx_articles_is_read ON articles(is_read)`,
        `CREATE INDEX IF NOT EXISTS idx_articles_guid ON articles(guid)`,
        `CREATE INDEX IF NOT EXISTS idx_feeds_last_fetched ON feeds(last_fetched)`,
    }
    
    for _, migration := range migrations {
        if _, err := db.Exec(migration); err != nil {
            log.Fatalf("Migration failed: %v", err)
        }
    }
    
    log.Println("Database migrations completed")
}
```

---

## Backend: Models

### backend/models/models.go

```go
package models

import "time"

type Feed struct {
    ID            int       `json:"id"`
    Title         string    `json:"title"`
    URL           string    `json:"url"`
    SiteURL       string    `json:"site_url"`
    Description   string    `json:"description"`
    FetchInterval int       `json:"fetch_interval"`
    LastFetched   time.Time `json:"last_fetched"`
    LastError     string    `json:"last_error"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

type Article struct {
    ID          int       `json:"id"`
    FeedID      int       `json:"feed_id"`
    FeedTitle   string    `json:"feed_title,omitempty"`
    GUID        string    `json:"guid"`
    Title       string    `json:"title"`
    URL         string    `json:"url"`
    Description string    `json:"description"`
    Content     string    `json:"content"`
    Author      string    `json:"author"`
    PublishedAt time.Time `json:"published_at"`
    FetchedAt   time.Time `json:"fetched_at"`
    IsRead      bool      `json:"is_read"`
    IsStarred   bool      `json:"is_starred"`
    ImageURL    string    `json:"image_url"`
}

type Stats struct {
    TotalFeeds    int `json:"total_feeds"`
    TotalArticles int `json:"total_articles"`
    UnreadCount   int `json:"unread_count"`
}
```

---

## Backend: API Handlers

### backend/api/handlers.go

```go
package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/go-chi/chi/v5"
    "github.com/yourusername/rss-reader/models"
)

// List articles with pagination and filters
func HandleListArticles(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
        if limit == 0 || limit > 100 {
            limit = 20
        }
        
        offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
        feedID := r.URL.Query().Get("feed_id")
        unreadOnly := r.URL.Query().Get("unread") == "true"
        
        query := `
            SELECT a.id, a.feed_id, f.title as feed_title, a.guid, a.title, 
                   a.url, a.description, a.author, a.published_at, 
                   a.is_read, a.is_starred, a.image_url
            FROM articles a
            JOIN feeds f ON a.feed_id = f.id
            WHERE 1=1
        `
        
        args := []interface{}{}
        
        if feedID != "" {
            query += " AND a.feed_id = ?"
            args = append(args, feedID)
        }
        
        if unreadOnly {
            query += " AND a.is_read = 0"
        }
        
        query += " ORDER BY a.published_at DESC LIMIT ? OFFSET ?"
        args = append(args, limit, offset)
        
        rows, err := db.Query(query, args...)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        
        articles := []models.Article{}
        for rows.Next() {
            var a models.Article
            err := rows.Scan(
                &a.ID, &a.FeedID, &a.FeedTitle, &a.GUID, &a.Title,
                &a.URL, &a.Description, &a.Author, &a.PublishedAt,
                &a.IsRead, &a.IsStarred, &a.ImageURL,
            )
            if err != nil {
                continue
            }
            articles = append(articles, a)
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(articles)
    }
}

// Create a new feed
func HandleCreateFeed(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var feed models.Feed
        if err := json.NewDecoder(r.Body).Decode(&feed); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        result, err := db.Exec(
            "INSERT INTO feeds (title, url, site_url, description) VALUES (?, ?, ?, ?)",
            feed.Title, feed.URL, feed.SiteURL, feed.Description,
        )
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        id, _ := result.LastInsertId()
        feed.ID = int(id)
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(feed)
    }
}

// Mark article as read/unread
func HandleMarkRead(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := chi.URLParam(r, "id")
        
        var req struct {
            IsRead bool `json:"is_read"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        _, err := db.Exec("UPDATE articles SET is_read = ? WHERE id = ?", req.IsRead, id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.WriteHeader(http.StatusOK)
    }
}

// Get statistics
func HandleGetStats(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var stats models.Stats
        
        db.QueryRow("SELECT COUNT(*) FROM feeds").Scan(&stats.TotalFeeds)
        db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&stats.TotalArticles)
        db.QueryRow("SELECT COUNT(*) FROM articles WHERE is_read = 0").Scan(&stats.UnreadCount)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(stats)
    }
}
```

---

## Backend: Feed Fetcher

### backend/services/fetcher.go

```go
package services

import (
    "database/sql"
    "log"
    "regexp"
    "strings"
    "time"
    
    "github.com/mmcdole/gofeed"
    "github.com/yourusername/rss-reader/models"
)

func FetchFeed(db *sql.DB, feedID int, feedURL string) error {
    parser := gofeed.NewParser()
    parsedFeed, err := parser.ParseURL(feedURL)
    if err != nil {
        updateFeedError(db, feedID, err.Error())
        return err
    }
    
    for _, item := range parsedFeed.Items {
        article := models.Article{
            FeedID:      feedID,
            GUID:        item.GUID,
            Title:       item.Title,
            URL:         item.Link,
            Description: cleanHTML(item.Description),
            Content:     item.Content,
            Author:      getAuthor(item),
            PublishedAt: getPublishedTime(item),
            ImageURL:    extractImageURL(item),
        }
        
        // Insert only if doesn't exist (GUID is unique)
        _, err := db.Exec(`
            INSERT OR IGNORE INTO articles 
            (feed_id, guid, title, url, description, content, author, published_at, image_url)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, article.FeedID, article.GUID, article.Title, article.URL,
            article.Description, article.Content, article.Author,
            article.PublishedAt, article.ImageURL)
            
        if err != nil {
            log.Printf("Error inserting article: %v", err)
        }
    }
    
    // Update feed last fetched time
    db.Exec("UPDATE feeds SET last_fetched = ?, last_error = NULL WHERE id = ?", 
        time.Now(), feedID)
    
    return nil
}

func extractImageURL(item *gofeed.Item) string {
    // Try item.Image first
    if item.Image != nil && item.Image.URL != "" {
        return item.Image.URL
    }
    
    // Try enclosures
    for _, enc := range item.Enclosures {
        if strings.HasPrefix(enc.Type, "image/") {
            return enc.URL
        }
    }
    
    // Parse HTML content for first image
    if item.Content != "" {
        if img := extractFirstImageFromHTML(item.Content); img != "" {
            return img
        }
    }
    
    if item.Description != "" {
        if img := extractFirstImageFromHTML(item.Description); img != "" {
            return img
        }
    }
    
    return ""
}

func extractFirstImageFromHTML(html string) string {
    // Simple regex to find first <img src="...">
    re := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)
    matches := re.FindStringSubmatch(html)
    if len(matches) > 1 {
        return matches[1]
    }
    return ""
}

func cleanHTML(html string) string {
    // Remove HTML tags for description
    re := regexp.MustCompile(`<[^>]*>`)
    return strings.TrimSpace(re.ReplaceAllString(html, " "))
}

func getAuthor(item *gofeed.Item) string {
    if item.Author != nil {
        return item.Author.Name
    }
    return ""
}

func getPublishedTime(item *gofeed.Item) time.Time {
    if item.PublishedParsed != nil {
        return *item.PublishedParsed
    }
    if item.UpdatedParsed != nil {
        return *item.UpdatedParsed
    }
    return time.Now()
}

func updateFeedError(db *sql.DB, feedID int, errMsg string) {
    db.Exec("UPDATE feeds SET last_error = ? WHERE id = ?", errMsg, feedID)
}
```

---

## Frontend: API Client

### frontend/src/api/client.js

```javascript
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

async function apiRequest(endpoint, options = {}) {
  const response = await fetch(`${API_URL}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.statusText}`);
  }

  return response.json();
}

export const fetchArticles = ({ offset = 0, limit = 20, unread = false, feedId = null }) => {
  const params = new URLSearchParams({
    offset: offset.toString(),
    limit: limit.toString(),
  });
  
  if (unread) params.append('unread', 'true');
  if (feedId) params.append('feed_id', feedId);
  
  return apiRequest(`/api/articles?${params}`);
};

export const fetchFeeds = () => apiRequest('/api/feeds');

export const createFeed = (feedData) => 
  apiRequest('/api/feeds', {
    method: 'POST',
    body: JSON.stringify(feedData),
  });

export const deleteFeed = (feedId) =>
  apiRequest(`/api/feeds/${feedId}`, {
    method: 'DELETE',
  });

export const markAsRead = (articleId, isRead) =>
  apiRequest(`/api/articles/${articleId}/read`, {
    method: 'PUT',
    body: JSON.stringify({ is_read: isRead }),
  });

export const starArticle = (articleId, isStarred) =>
  apiRequest(`/api/articles/${articleId}/star`, {
    method: 'PUT',
    body: JSON.stringify({ is_starred: isStarred }),
  });

export const fetchStats = () => apiRequest('/api/stats');
```

---

## Frontend: Custom Hooks

### frontend/src/hooks/useArticles.js

```javascript
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
      if (lastPage.length < 20) return undefined;
      return allPages.length * 20;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}
```

### frontend/src/hooks/useFeeds.js

```javascript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchFeeds, createFeed, deleteFeed } from '../api/client';

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
      queryClient.invalidateQueries({ queryKey: ['articles'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteFeed,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] });
      queryClient.invalidateQueries({ queryKey: ['articles'] });
    },
  });

  return {
    feeds: feedsQuery.data || [],
    isLoading: feedsQuery.isLoading,
    createFeed: createMutation.mutate,
    deleteFeed: deleteMutation.mutate,
  };
}
```

---

## Frontend: Article Card (Optimized)

### frontend/src/components/ArticleCard.jsx

```jsx
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
      
      queryClient.setQueryData(['articles'], (old) => ({
        ...old,
        pages: old.pages.map(page =>
          page.map(a => a.id === article.id ? { ...a, is_read: isRead } : a)
        ),
      }));
      
      return { previousData };
    },
    onError: (err, variables, context) => {
      queryClient.setQueryData(['articles'], context.previousData);
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
```

---

## Frontend: Feed Manager Modal

### frontend/src/components/FeedManager.jsx

```jsx
import { useState } from 'react';
import { useFeeds } from '../hooks/useFeeds';

export default function FeedManager({ onClose, onSelectFeed }) {
  const { feeds, isLoading, createFeed, deleteFeed } = useFeeds();
  const [newFeedUrl, setNewFeedUrl] = useState('');
  const [isAdding, setIsAdding] = useState(false);

  const handleAddFeed = async (e) => {
    e.preventDefault();
    if (!newFeedUrl.trim()) return;

    setIsAdding(true);
    try {
      await createFeed({ url: newFeedUrl, title: 'Loading...' });
      setNewFeedUrl('');
    } catch (error) {
      alert('Failed to add feed: ' + error.message);
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-2xl w-full max-h-[80vh] overflow-hidden flex flex-col">
        <div className="p-6 border-b flex items-center justify-between">
          <h2 className="text-2xl font-bold">Manage Feeds</h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 text-2xl"
          >
            ×
          </button>
        </div>

        <div className="p-6 border-b">
          <form onSubmit={handleAddFeed} className="flex gap-2">
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
              disabled={isAdding}
              className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
            >
              {isAdding ? 'Adding...' : 'Add Feed'}
            </button>
          </form>
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
                  </div>
                  <button
                    onClick={() => {
                      if (confirm(`Delete feed "${feed.title}"?`)) {
                        deleteFeed(feed.id);
                      }
                    }}
                    className="ml-4 px-3 py-1 text-red-600 hover:bg-red-50 rounded"
                  >
                    Delete
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
```

---

## Docker Compose Final Configuration

### docker-compose.yml

```yaml
version: '3.8'

services:
  rss-reader:
    build: .
    container_name: rss-reader
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - DATABASE_PATH=/data/rss-reader.db
      - PORT=8080
      - FEED_FETCH_INTERVAL=1800
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

---

## Environment Setup Commands

### Initial setup:

```bash
# Create project structure
mkdir -p rss-reader/{backend,frontend}
cd rss-reader

# Backend setup
cd backend
go mod init github.com/yourusername/rss-reader
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go get github.com/mmcdole/gofeed
go get github.com/mattn/go-sqlite3

# Frontend setup
cd ../frontend
npm create vite@latest . -- --template react
npm install
npm install @tanstack/react-query
npm install react-masonry-css
npm install react-intersection-observer
npm install date-fns
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p

# Create data directory for Unraid
mkdir ../data

# Create .env files
echo "DATABASE_PATH=./data/rss-reader.db" > ../backend/.env
echo "VITE_API_URL=http://localhost:8080" > .env
```

### Development:

```bash
# Terminal 1: Backend
cd backend
go run main.go

# Terminal 2: Frontend
cd frontend
npm run dev
```

### Production build:

```bash
# Build everything
docker-compose build

# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

---

## Quick Test Script

### scripts/test-api.sh

```bash
#!/bin/bash

API_URL="http://localhost:8080"

echo "Testing RSS Reader API..."

# Test health
echo -e "\n1. Health check:"
curl -s $API_URL/health

# Add a feed
echo -e "\n\n2. Adding feed:"
curl -s -X POST $API_URL/api/feeds \
  -H "Content-Type: application/json" \
  -d '{"url":"https://xkcd.com/rss.xml","title":"XKCD"}' | jq

# List feeds
echo -e "\n\n3. Listing feeds:"
curl -s $API_URL/api/feeds | jq

# Wait for articles to be fetched
echo -e "\n\nWaiting 5 seconds for articles to be fetched..."
sleep 5

# List articles
echo -e "\n4. Listing articles:"
curl -s $API_URL/api/articles?limit=5 | jq

# Get stats
echo -e "\n\n5. Stats:"
curl -s $API_URL/api/stats | jq

echo -e "\n\nAPI tests complete!"
```

---

Ready to start coding! Which component would you like to build first?
