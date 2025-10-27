# Refined RSS Reader Project Plan
## Tech Stack & Deployment Strategy

**Selected Technologies:**
- **Backend:** Go 1.21+
- **Database:** SQLite
- **Frontend:** React 18 + Vite
- **Images:** Direct links (no proxy)
- **Deployment:** Docker Compose (Unraid) + Vercel (demo)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                 React Frontend (SPA)                     │
│  - Masonry Grid (react-masonry-css)                     │
│  - TanStack Query for API calls                         │
│  - Infinite Scroll                                      │
│  - Direct image loading from sources                    │
└────────────────┬────────────────────────────────────────┘
                 │ REST API
┌────────────────▼────────────────────────────────────────┐
│              Go Backend API Server                       │
│  - Chi Router (lightweight)                             │
│  - Feed Management CRUD                                 │
│  - Article API with pagination                          │
│  - Background goroutine for feed fetching               │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│                  SQLite Database                         │
│  - Single file: rss-reader.db                           │
│  - Embedded in Go binary                                │
│  - WAL mode for better concurrency                      │
└─────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
rss-reader/
├── backend/
│   ├── main.go                 # Entry point
│   ├── api/
│   │   ├── handlers.go         # HTTP handlers
│   │   ├── middleware.go       # CORS, logging, etc.
│   │   └── routes.go           # Route definitions
│   ├── db/
│   │   ├── sqlite.go           # SQLite connection & setup
│   │   ├── migrations.go       # Schema migrations
│   │   └── queries.go          # Database queries
│   ├── models/
│   │   ├── feed.go
│   │   └── article.go
│   ├── services/
│   │   ├── fetcher.go          # RSS feed fetching logic
│   │   └── parser.go           # Feed parsing with gofeed
│   ├── worker/
│   │   └── scheduler.go        # Background feed updates
│   └── go.mod
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── ArticleCard.jsx     # Individual article card
│   │   │   ├── MasonryGrid.jsx     # Pinterest-style layout
│   │   │   ├── ArticleModal.jsx    # Full article view
│   │   │   ├── FeedManager.jsx     # Add/remove feeds
│   │   │   └── Header.jsx          # App header/nav
│   │   ├── hooks/
│   │   │   ├── useArticles.js      # TanStack Query hook
│   │   │   └── useFeeds.js
│   │   ├── api/
│   │   │   └── client.js           # API client setup
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── package.json
│   └── vite.config.js
│
├── docker-compose.yml          # Unraid deployment
├── Dockerfile                  # Go backend container
├── Dockerfile.demo             # Demo-specific container
├── vercel.json                 # Vercel configuration
├── scripts/
│   ├── build-binary.sh         # Build standalone binary
│   └── seed-demo-data.sh       # Populate demo database
└── README.md
```

---

## Database Schema

```sql
-- feeds table
CREATE TABLE feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    site_url TEXT,
    description TEXT,
    fetch_interval INTEGER DEFAULT 1800,  -- seconds
    last_fetched TIMESTAMP,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- articles table
CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    guid TEXT UNIQUE NOT NULL,              -- Unique identifier from RSS
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT,                        -- Short summary
    content TEXT,                            -- Full content if available
    author TEXT,
    published_at TIMESTAMP NOT NULL,
    fetched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_read BOOLEAN DEFAULT FALSE,
    is_starred BOOLEAN DEFAULT FALSE,
    image_url TEXT,                          -- Featured image URL (direct link)
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_articles_feed_id ON articles(feed_id);
CREATE INDEX idx_articles_published_at ON articles(published_at DESC);
CREATE INDEX idx_articles_is_read ON articles(is_read);
CREATE INDEX idx_articles_guid ON articles(guid);
CREATE INDEX idx_feeds_last_fetched ON feeds(last_fetched);

-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=-64000;  -- 64MB cache
```

---

## API Endpoints

### Feeds
```
GET    /api/feeds              List all feeds
POST   /api/feeds              Add new feed
GET    /api/feeds/:id          Get feed details
PUT    /api/feeds/:id          Update feed
DELETE /api/feeds/:id          Delete feed
POST   /api/feeds/:id/refresh  Manually refresh feed
```

### Articles
```
GET    /api/articles           List articles (paginated, filtered)
                               ?limit=20&offset=0&feed_id=1&unread=true
GET    /api/articles/:id       Get single article
PUT    /api/articles/:id/read  Mark as read/unread
PUT    /api/articles/:id/star  Star/unstar article
POST   /api/articles/mark-all-read  Mark all as read
```

### Import/Export
```
POST   /api/opml/import        Import feeds from OPML
GET    /api/opml/export        Export feeds to OPML
```

### Statistics
```
GET    /api/stats              Get stats (total feeds, articles, unread count)
```

---

## Deployment Strategy

### Deployment 1: Unraid Server (Docker Compose)

**docker-compose.yml:**
```yaml
version: '3.8'

services:
  rss-reader:
    build: .
    container_name: rss-reader
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data              # SQLite database location
      - ./config.yaml:/app/config.yaml
    environment:
      - DATABASE_PATH=/data/rss-reader.db
      - PORT=8080
      - FEED_FETCH_INTERVAL=1800  # 30 minutes
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

**Dockerfile:**
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Build binary with SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o rss-reader .

# Frontend build stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy backend binary
COPY --from=builder /app/rss-reader .

# Copy frontend build (Go will serve these)
COPY --from=frontend-builder /app/dist ./frontend/dist

# Create data directory
RUN mkdir -p /data

EXPOSE 8080

CMD ["./rss-reader"]
```

**Unraid Setup Instructions:**
```bash
# On Unraid, via SSH or terminal:

# 1. Navigate to appdata directory
cd /mnt/user/appdata
mkdir rss-reader
cd rss-reader

# 2. Clone or copy your repository
git clone https://github.com/yourusername/rss-reader.git .

# 3. Create data directory
mkdir data

# 4. Start with Docker Compose
docker-compose up -d

# 5. Access at http://unraid-ip:8080
```

**Unraid Community Apps Template** (optional):
```xml
<?xml version="1.0"?>
<Container version="2">
  <Name>RSS-Reader</Name>
  <Repository>yourusername/rss-reader</Repository>
  <Registry>https://hub.docker.com/r/yourusername/rss-reader/</Registry>
  <Network>bridge</Network>
  <Privileged>false</Privileged>
  <Support>https://github.com/yourusername/rss-reader</Support>
  <Project>https://github.com/yourusername/rss-reader</Project>
  <Overview>Self-hosted RSS reader with Pinterest-style visual feed</Overview>
  <Category>Network:Other Status:Stable</Category>
  <WebUI>http://[IP]:[PORT:8080]</WebUI>
  <Icon>https://raw.githubusercontent.com/yourusername/rss-reader/main/icon.png</Icon>
  
  <Config Name="WebUI" Target="8080" Default="8080" Mode="tcp" Description="Container Port: 8080" Type="Port" Display="always" Required="true" Mask="false">8080</Config>
  
  <Config Name="Database" Target="/data" Default="/mnt/user/appdata/rss-reader/data" Mode="rw" Description="Database storage" Type="Path" Display="advanced" Required="true" Mask="false">/mnt/user/appdata/rss-reader/data</Config>
</Container>
```

---

### Deployment 2: Vercel Demo

**Challenges with Vercel:**
1. Serverless functions are stateless
2. SQLite needs persistent filesystem
3. Background workers don't exist in serverless

**Solution: Demo Mode with Embedded Data**

Create a special demo build:
- Pre-populate SQLite database with sample feeds/articles
- Embed the SQLite database in the binary (read-only)
- No write operations (or use in-memory SQLite for session)
- Frontend deploys normally to Vercel
- Backend deploys as Vercel serverless functions

**vercel.json:**
```json
{
  "version": 2,
  "builds": [
    {
      "src": "frontend/package.json",
      "use": "@vercel/static-build",
      "config": {
        "distDir": "dist"
      }
    },
    {
      "src": "backend/api/*.go",
      "use": "@vercel/go"
    }
  ],
  "routes": [
    {
      "src": "/api/(.*)",
      "dest": "/backend/api/$1"
    },
    {
      "src": "/(.*)",
      "dest": "/frontend/$1"
    }
  ],
  "env": {
    "DEMO_MODE": "true"
  }
}
```

**Demo Mode Implementation:**

```go
// backend/main.go
package main

import (
    "embed"
    "os"
)

//go:embed demo.db
var demoDatabase embed.FS

func main() {
    var dbPath string
    
    if os.Getenv("DEMO_MODE") == "true" {
        // Extract embedded demo database to temp location
        dbPath = extractDemoDatabase()
    } else {
        // Use persistent database path
        dbPath = os.Getenv("DATABASE_PATH")
    }
    
    db := initDatabase(dbPath)
    startServer(db)
}
```

**Alternative: Vercel Frontend + Separate Demo Backend**

Even simpler approach:

```
Vercel: Frontend only (React SPA)
         ↓ API calls
Digital Ocean/Fly.io: Demo backend (small droplet, same Docker image)
```

**vercel.json (frontend only):**
```json
{
  "version": 2,
  "builds": [
    {
      "src": "frontend/package.json",
      "use": "@vercel/static-build",
      "config": {
        "distDir": "dist"
      }
    }
  ],
  "rewrites": [
    {
      "source": "/api/:path*",
      "destination": "https://demo-api.yourdomain.com/api/:path*"
    }
  ]
}
```

**Recommended Approach for Vercel Demo:**

Deploy demo backend to **Fly.io** (free tier):
- Same Docker image as Unraid
- Pre-seeded with interesting demo feeds
- Read-only mode or reset daily
- Costs: $0/month on free tier

```bash
# Deploy demo backend to Fly.io
fly launch
fly deploy
fly scale count 1

# Get URL: https://rss-reader-demo.fly.dev

# Update frontend .env for Vercel
VITE_API_URL=https://rss-reader-demo.fly.dev
```

Then Vercel just hosts the static frontend.

---

## Go Backend Implementation Details

### Main.go Structure

```go
package main

import (
    "embed"
    "log"
    "net/http"
    "os"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
    // Initialize database
    db := initDatabase(os.Getenv("DATABASE_PATH"))
    defer db.Close()
    
    // Start background worker
    go startFeedWorker(db)
    
    // Setup router
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Accept", "Content-Type"},
        AllowCredentials: false,
    }))
    
    // API routes
    r.Route("/api", func(r chi.Router) {
        r.Get("/feeds", handleListFeeds(db))
        r.Post("/feeds", handleCreateFeed(db))
        r.Get("/feeds/{id}", handleGetFeed(db))
        r.Put("/feeds/{id}", handleUpdateFeed(db))
        r.Delete("/feeds/{id}", handleDeleteFeed(db))
        r.Post("/feeds/{id}/refresh", handleRefreshFeed(db))
        
        r.Get("/articles", handleListArticles(db))
        r.Get("/articles/{id}", handleGetArticle(db))
        r.Put("/articles/{id}/read", handleMarkRead(db))
        r.Put("/articles/{id}/star", handleStarArticle(db))
        r.Post("/articles/mark-all-read", handleMarkAllRead(db))
        
        r.Post("/opml/import", handleImportOPML(db))
        r.Get("/opml/export", handleExportOPML(db))
        
        r.Get("/stats", handleGetStats(db))
    })
    
    // Health check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    // Serve frontend
    r.Handle("/*", http.FileServer(http.FS(frontendFS)))
    
    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Starting server on :%s", port)
    http.ListenAndServe(":"+port, r)
}
```

### Feed Fetcher Worker

```go
// backend/worker/scheduler.go
package worker

import (
    "database/sql"
    "log"
    "time"
    
    "github.com/mmcdole/gofeed"
)

func StartFeedWorker(db *sql.DB) {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()
    
    // Initial fetch on startup
    fetchAllFeeds(db)
    
    // Periodic fetches
    for range ticker.C {
        fetchAllFeeds(db)
    }
}

func fetchAllFeeds(db *sql.DB) {
    feeds, err := getFeeds(db)
    if err != nil {
        log.Printf("Error getting feeds: %v", err)
        return
    }
    
    for _, feed := range feeds {
        go fetchFeed(db, feed)
    }
}

func fetchFeed(db *sql.DB, feed Feed) {
    parser := gofeed.NewParser()
    parsedFeed, err := parser.ParseURL(feed.URL)
    if err != nil {
        updateFeedError(db, feed.ID, err.Error())
        return
    }
    
    for _, item := range parsedFeed.Items {
        article := Article{
            FeedID:      feed.ID,
            GUID:        item.GUID,
            Title:       item.Title,
            URL:         item.Link,
            Description: item.Description,
            Content:     item.Content,
            Author:      getAuthor(item),
            PublishedAt: getPublishedTime(item),
            ImageURL:    extractImageURL(item),
        }
        
        insertArticle(db, article)
    }
    
    updateFeedLastFetched(db, feed.ID)
}

func extractImageURL(item *gofeed.Item) string {
    // Try item.Image first
    if item.Image != nil && item.Image.URL != "" {
        return item.Image.URL
    }
    
    // Try enclosures (common in podcasts/media feeds)
    for _, enc := range item.Enclosures {
        if strings.HasPrefix(enc.Type, "image/") {
            return enc.URL
        }
    }
    
    // Parse HTML content for first image
    if item.Content != "" {
        return extractFirstImageFromHTML(item.Content)
    }
    
    if item.Description != "" {
        return extractFirstImageFromHTML(item.Description)
    }
    
    return ""
}
```

---

## React Frontend Implementation Details

### Main Component Structure

```jsx
// frontend/src/App.jsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';
import Header from './components/Header';
import MasonryGrid from './components/MasonryGrid';
import FeedManager from './components/FeedManager';
import ArticleModal from './components/ArticleModal';

const queryClient = new QueryClient();

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
```

### Masonry Grid Component

```jsx
// frontend/src/components/MasonryGrid.jsx
import { useInfiniteQuery } from '@tanstack/react-query';
import { useInView } from 'react-intersection-observer';
import { useEffect } from 'react';
import Masonry from 'react-masonry-css';
import ArticleCard from './ArticleCard';
import { fetchArticles } from '../api/client';

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
  } = useInfiniteQuery({
    queryKey: ['articles', filterUnread, selectedFeed],
    queryFn: ({ pageParam = 0 }) => 
      fetchArticles({ 
        offset: pageParam, 
        limit: 20,
        unread: filterUnread,
        feedId: selectedFeed 
      }),
    getNextPageParam: (lastPage, pages) => {
      if (lastPage.length < 20) return undefined;
      return pages.length * 20;
    },
  });

  useEffect(() => {
    if (inView && hasNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, fetchNextPage]);

  if (status === 'loading') {
    return <LoadingSkeleton />;
  }

  const articles = data?.pages.flatMap(page => page) ?? [];

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
          {isFetchingNextPage ? 'Loading more...' : 'Load more'}
        </div>
      )}
    </div>
  );
}
```

### Article Card Component

```jsx
// frontend/src/components/ArticleCard.jsx
import { useState } from 'react';
import { formatDistanceToNow } from 'date-fns';
import { markAsRead } from '../api/client';

export default function ArticleCard({ article, onClick }) {
  const [isRead, setIsRead] = useState(article.is_read);
  
  const handleClick = async () => {
    if (!isRead) {
      await markAsRead(article.id, true);
      setIsRead(true);
    }
    onClick();
  };

  return (
    <div 
      className={`mb-4 break-inside-avoid cursor-pointer group ${
        isRead ? 'opacity-60' : ''
      }`}
      onClick={handleClick}
    >
      <div className="bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow overflow-hidden">
        {article.image_url && (
          <div className="relative overflow-hidden">
            <img
              src={article.image_url}
              alt={article.title}
              className="w-full h-auto object-cover group-hover:scale-105 transition-transform duration-300"
              loading="lazy"
              onError={(e) => {
                e.target.style.display = 'none';
              }}
            />
          </div>
        )}
        
        <div className="p-4">
          <h3 className="font-semibold text-lg mb-2 line-clamp-3">
            {article.title}
          </h3>
          
          {article.description && (
            <p className="text-gray-600 text-sm mb-3 line-clamp-3">
              {article.description}
            </p>
          )}
          
          <div className="flex items-center justify-between text-xs text-gray-500">
            <span>{article.feed_title}</span>
            <span>
              {formatDistanceToNow(new Date(article.published_at), { 
                addSuffix: true 
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

## Direct Image Links Strategy

Since you chose direct image links (no proxy), here are important considerations:

### Pros:
- ✅ No bandwidth cost to your server
- ✅ Simpler backend implementation
- ✅ Faster image loading (no proxy hop)
- ✅ Lower storage requirements

### Cons & Mitigations:

**1. Broken Images**
```jsx
// Handle broken images gracefully
<img
  src={article.image_url}
  alt={article.title}
  onError={(e) => {
    // Hide broken image
    e.target.style.display = 'none';
    // Or show placeholder
    // e.target.src = '/placeholder.png';
  }}
/>
```

**2. CORS Issues**
```jsx
// Add crossorigin attribute for better error handling
<img
  src={article.image_url}
  crossOrigin="anonymous"
  alt={article.title}
/>
```

**3. Mixed Content (HTTP images on HTTPS site)**
```go
// In backend: upgrade HTTP to HTTPS when possible
func sanitizeImageURL(url string) string {
    if strings.HasPrefix(url, "http://") {
        // Try HTTPS version
        httpsURL := strings.Replace(url, "http://", "https://", 1)
        if isImageAccessible(httpsURL) {
            return httpsURL
        }
    }
    return url
}
```

**4. Privacy Concerns**
Add option in UI to disable images:
```jsx
const [showImages, setShowImages] = useState(
  localStorage.getItem('showImages') !== 'false'
);

// In ArticleCard
{showImages && article.image_url && (
  <img src={article.image_url} ... />
)}
```

**5. Slow Loading**
```jsx
// Use lazy loading and blur placeholder
import { LazyLoadImage } from 'react-lazy-load-image-component';
import 'react-lazy-load-image-component/src/effects/blur.css';

<LazyLoadImage
  src={article.image_url}
  alt={article.title}
  effect="blur"
  threshold={100}
/>
```

---

## Implementation Roadmap (Revised)

### Week 1: Backend Core
**Goals:** Basic Go backend with SQLite, feed fetching
- [x] Project setup (Go modules, Chi router)
- [x] SQLite database setup with migrations
- [x] Feed CRUD endpoints
- [x] Basic article endpoints
- [x] RSS parsing with gofeed
- [x] Background worker for feed fetching
- [x] Image URL extraction
- [x] OPML import/export

**Deliverable:** Working API you can test with curl/Postman

### Week 2: Frontend Foundation
**Goals:** React app with basic masonry layout
- [x] Vite + React setup
- [x] TanStack Query integration
- [x] API client with fetch
- [x] Basic masonry grid with react-masonry-css
- [x] Article card component
- [x] Responsive layout
- [x] Tailwind CSS styling

**Deliverable:** Frontend displays articles in grid (even with mock data)

### Week 3: Core Features
**Goals:** Full user workflow
- [x] Feed management UI
- [x] Infinite scroll
- [x] Article modal/reader
- [x] Mark as read/unread
- [x] Star articles
- [x] Filter by feed
- [x] Filter unread only
- [x] Search/filter UI

**Deliverable:** Fully functional RSS reader

### Week 4: Docker & Deployment
**Goals:** Both deployments working
- [x] Dockerfile for production
- [x] Docker Compose for Unraid
- [x] Test on Unraid
- [x] Demo data seed script
- [x] Vercel deployment setup
- [x] Fly.io demo backend
- [x] Documentation

**Deliverable:** Running on Unraid + live demo on Vercel

### Week 5: Polish & Launch
**Goals:** Production ready
- [x] Error handling
- [x] Loading states
- [x] Empty states
- [x] Keyboard shortcuts
- [x] Dark mode
- [x] Performance optimization
- [x] Mobile testing
- [x] README with screenshots

**Deliverable:** Launch! 🚀

---

## Build & Deploy Scripts

### scripts/build-binary.sh
```bash
#!/bin/bash
set -e

echo "Building frontend..."
cd frontend
npm run build
cd ..

echo "Building Go binary..."
cd backend
CGO_ENABLED=1 go build -o ../rss-reader .
cd ..

echo "Build complete: ./rss-reader"
echo "Run with: ./rss-reader"
```

### scripts/build-docker.sh
```bash
#!/bin/bash
set -e

echo "Building Docker image..."
docker build -t rss-reader:latest .

echo "Testing image..."
docker run --rm -p 8080:8080 -v $(pwd)/data:/data rss-reader:latest &
sleep 5
curl -f http://localhost:8080/health || exit 1

echo "Docker image ready: rss-reader:latest"
```

### scripts/deploy-vercel.sh
```bash
#!/bin/bash
set -e

echo "Deploying demo backend to Fly.io..."
fly deploy --app rss-reader-demo

echo "Building frontend for Vercel..."
cd frontend
VITE_API_URL=https://rss-reader-demo.fly.dev npm run build

echo "Deploying to Vercel..."
vercel --prod

echo "Demo live at: https://rss-reader.vercel.app"
```

---

## Configuration Files

### config.yaml (Unraid)
```yaml
database:
  path: /data/rss-reader.db
  
server:
  port: 8080
  
worker:
  fetch_interval: 1800  # 30 minutes
  concurrent_fetches: 5
  
feeds:
  timeout: 30  # seconds
  max_articles_per_feed: 1000
  
cleanup:
  enabled: true
  delete_read_after_days: 30
```

### .env.example
```bash
# Backend
DATABASE_PATH=/data/rss-reader.db
PORT=8080
FEED_FETCH_INTERVAL=1800
DEMO_MODE=false

# Frontend
VITE_API_URL=http://localhost:8080
```

---

## Next Steps

1. **Initialize Go Project**
   ```bash
   mkdir rss-reader && cd rss-reader
   mkdir -p backend frontend
   cd backend && go mod init github.com/yourusername/rss-reader
   ```

2. **Set Up Dependencies**
   ```bash
   # Backend
   go get github.com/go-chi/chi/v5
   go get github.com/mmcdole/gofeed
   go get github.com/mattn/go-sqlite3
   
   # Frontend
   cd ../frontend
   npm create vite@latest . -- --template react
   npm install @tanstack/react-query
   npm install react-masonry-css
   npm install react-intersection-observer
   npm install date-fns
   ```

3. **Start Development**
   ```bash
   # Terminal 1: Backend
   cd backend && go run main.go
   
   # Terminal 2: Frontend
   cd frontend && npm run dev
   ```

Ready to start building? Want me to help you with any specific component first?
