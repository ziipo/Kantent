# Kantent - Getting Started

Kantent is a self-hosted RSS reader with a Pinterest-style visual layout. This guide will help you get it running on your system.

## Quick Start with Docker (Recommended)

The easiest way to run Kantent is using Docker Compose:

```bash
# From the project root directory
docker-compose up
```

**What this does:**
- Builds both the Go backend and React frontend
- Creates a SQLite database in `./data/rss-reader.db`
- Starts the application on `http://localhost:8080`
- Automatically refreshes feeds every 30 minutes

**To stop the application:**
```bash
docker-compose down
```

## Development Setup

If you want to develop or run without Docker, follow these steps:

### Prerequisites

- **Go 1.21 or higher** - [Download here](https://go.dev/dl/)
- **Node.js 18 or higher** - [Download here](https://nodejs.org/)
- **SQLite3** (usually pre-installed on macOS/Linux)

### Backend Setup

```bash
# Navigate to the backend directory
cd backend

# Install Go dependencies
go mod download

# Build the backend binary
go build -o kantent .

# Run the backend server
DATABASE_PATH=../data/rss-reader.db PORT=8080 ./kantent
```

**What these commands do:**

- `go mod download` - Downloads all Go dependencies listed in `go.mod`
- `go build -o kantent .` - Compiles the Go application into a binary named `kantent`
- `DATABASE_PATH=../data/rss-reader.db PORT=8080 ./kantent` - Runs the server with:
  - Database stored in `../data/rss-reader.db` (creates it if doesn't exist)
  - API server listening on port 8080
  - Background worker that fetches feeds every 30 minutes

The backend will be available at `http://localhost:8080/api`

### Frontend Setup

Open a **new terminal window** and run:

```bash
# Navigate to the frontend directory
cd frontend

# Install Node dependencies
npm install

# Start the development server
npm run dev
```

**What these commands do:**

- `npm install` - Downloads all JavaScript dependencies from `package.json`
- `npm run dev` - Starts the Vite development server with:
  - Hot module reloading (changes appear instantly)
  - Development server typically on `http://localhost:5173`
  - Proxies API requests to the backend at `http://localhost:8080`

### Building for Production

If you want to create optimized production builds:

```bash
# Build the frontend
cd frontend
npm run build

# This creates a 'dist' folder with optimized static files
# The backend will automatically serve these files from frontend/dist
```

```bash
# Build the backend
cd backend
go build -o kantent .

# Run in production mode
DATABASE_PATH=../data/rss-reader.db PORT=8080 ./kantent
```

## Using the Application

Once running, open your browser to `http://localhost:8080` (or `http://localhost:5173` if using dev mode).

### Adding Feeds

Click the **"Manage Feeds"** button in the header to open the feed manager. You have four options:

#### 1. Discover Feeds
- Enter any website URL (e.g., `https://www.theverge.com`)
- Click "Discover Feeds"
- The app will automatically find RSS/Atom feeds on that site
- Click "Add" next to any discovered feed to subscribe

#### 2. RSS Feed
- Enter the direct URL to an RSS or Atom feed
- Click "Add Feed"
- Examples: `https://xkcd.com/rss.xml`, `https://blog.golang.org/feed.atom`

#### 3. Reddit
- Enter a subreddit name (e.g., `technology` or `r/technology`)
- Select sort order: Hot, New, Top, or Rising
- Click "Add Reddit Feed"
- The app automatically constructs the RSS feed URL

#### 4. YouTube
- Paste any YouTube channel URL or channel ID
- Supported formats:
  - `https://youtube.com/@channelname`
  - `https://youtube.com/channel/UC...`
  - `UC...` (just the channel ID)
- Click "Add YouTube Feed"
- Videos will appear with thumbnails

### Managing Content

- **Click any article card** to open and read the full content
- **Click the checkmark icon** to mark as read/unread
- **Click the star icon** to bookmark articles
- **Refresh button** on each feed to manually fetch new content
- **Delete button** to remove a feed and all its articles

## Configuration

### Environment Variables

You can customize the application using these environment variables:

- `DATABASE_PATH` - Location of the SQLite database (default: `./data/rss-reader.db`)
- `PORT` - HTTP server port (default: `8080`)
- `FEED_FETCH_INTERVAL` - Seconds between automatic feed updates (default: `1800` = 30 minutes)

Example with custom settings:
```bash
DATABASE_PATH=/var/lib/kantent/db.sqlite PORT=3000 FEED_FETCH_INTERVAL=3600 ./kantent
```

### Docker Compose Configuration

Edit `docker-compose.yml` to change settings:

```yaml
environment:
  - DATABASE_PATH=/data/rss-reader.db
  - PORT=8080
  - FEED_FETCH_INTERVAL=1800  # 30 minutes

ports:
  - "8080:8080"  # Change left side to use different host port

volumes:
  - ./data:/data  # Database persists here
```

## Troubleshooting

### Port Already in Use

If you see `bind: address already in use`:

```bash
# Find what's using port 8080
lsof -i :8080

# Kill the process (replace PID with the actual process ID)
kill -9 PID

# Or use a different port
PORT=3000 ./kantent
```

### Database Locked Error

If you see `database is locked`:
- Make sure only one instance of Kantent is running
- Check that the database file isn't open in another application
- The database uses WAL mode which should prevent most locking issues

### Frontend Can't Connect to Backend

If the frontend shows connection errors:
- Verify the backend is running on port 8080
- Check `frontend/vite.config.js` - the proxy should point to `http://localhost:8080`
- Make sure CORS is enabled (it's configured by default in the backend)

### Feed Won't Update

If a feed shows an error or won't update:
- Check the feed URL is correct and accessible
- Some feeds may have rate limiting or require specific user agents
- Look at the "Error" message shown on the feed in the UI
- Try the "Refresh" button to manually trigger an update

## Data Location

All application data is stored in the SQLite database:
- Default location: `./data/rss-reader.db`
- Contains feeds, articles, read/starred status
- Back up this file to preserve your data
- Delete this file to completely reset the application

## Next Steps

- Add your favorite blogs, news sites, and YouTube channels
- Explore the Pinterest-style masonry grid layout
- Use filters to view unread articles or specific feeds
- Star important articles to find them later

Enjoy your ad-free, self-hosted reading experience!
