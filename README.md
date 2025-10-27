# Kantent - Self-Hosted RSS Reader

A beautiful, self-hosted RSS reader with a Pinterest-style visual layout. Built with Go and React.

## Features

- 📱 **Pinterest-style masonry grid layout** - Visual, card-based interface
- 🔄 **Automatic feed updates** - Background worker fetches feeds every 30 minutes
- 📖 **Read/unread tracking** - Keep track of what you've read
- ⭐ **Star articles** - Save your favorite articles
- 🎨 **Clean, modern UI** - Built with Tailwind CSS
- 🐳 **Docker support** - Easy deployment with Docker Compose
- 💾 **SQLite database** - Lightweight and portable
- 🖼️ **Direct image loading** - No proxy, images load directly from sources

## Tech Stack

### Backend
- **Go 1.21+** - Fast, efficient backend
- **Chi Router** - Lightweight HTTP router
- **SQLite** - Embedded database with WAL mode
- **gofeed** - RSS/Atom feed parser

### Frontend
- **React 18** - Modern UI framework
- **Vite** - Fast build tool
- **TanStack Query** - Data fetching and caching
- **Tailwind CSS** - Utility-first styling
- **react-masonry-css** - Pinterest-style grid layout

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/ziipo/Kantent.git
cd Kantent
```

2. Start with Docker Compose:
```bash
docker-compose up -d
```

3. Open your browser and navigate to `http://localhost:8080`

### Development Setup

#### Prerequisites
- Go 1.21 or higher
- Node.js 18 or higher
- SQLite3

#### Backend

```bash
cd backend
go mod download
go run main.go
```

The backend will start on `http://localhost:8080`

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:5173`

## Configuration

### Environment Variables

- `DATABASE_PATH` - Path to SQLite database (default: `./data/rss-reader.db`)
- `PORT` - Server port (default: `8080`)
- `FEED_FETCH_INTERVAL` - Feed fetch interval in seconds (default: `1800` / 30 minutes)

### Docker Volumes

The Docker setup mounts `./data` to persist the SQLite database:

```yaml
volumes:
  - ./data:/data
```

## Usage

### Adding Feeds

1. Click "Manage Feeds" in the header
2. Enter an RSS/Atom feed URL
3. Click "Add Feed"
4. The feed will be fetched immediately and updated every 30 minutes

### Reading Articles

- Click on any article card to open the full article modal
- Click "Read Original Article" to open the source in a new tab
- Articles are automatically marked as read when opened

### Filtering

- **Unread Only** - Toggle to show only unread articles
- **By Feed** - Click on a feed in the Feed Manager to filter by that feed

## API Endpoints

### Feeds
- `GET /api/feeds` - List all feeds
- `POST /api/feeds` - Add new feed
- `GET /api/feeds/:id` - Get feed details
- `PUT /api/feeds/:id` - Update feed
- `DELETE /api/feeds/:id` - Delete feed
- `POST /api/feeds/:id/refresh` - Manually refresh feed

### Articles
- `GET /api/articles` - List articles (with pagination and filters)
- `GET /api/articles/:id` - Get single article
- `PUT /api/articles/:id/read` - Mark as read/unread
- `PUT /api/articles/:id/star` - Star/unstar article
- `POST /api/articles/mark-all-read` - Mark all as read

### Stats
- `GET /api/stats` - Get statistics (total feeds, articles, unread count)

## Project Structure

```
Kantent/
├── backend/
│   ├── api/           # HTTP handlers
│   ├── db/            # Database layer
│   ├── models/        # Data models
│   ├── services/      # Business logic (RSS fetching)
│   ├── worker/        # Background worker
│   └── main.go        # Entry point
├── frontend/
│   ├── src/
│   │   ├── api/       # API client
│   │   ├── components/# React components
│   │   ├── hooks/     # Custom hooks
│   │   └── App.jsx    # Main app
│   └── package.json
├── data/              # SQLite database (created on first run)
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Building from Source

### Build Docker Image

```bash
docker build -t kantent:latest .
```

### Build Standalone Binary

```bash
# Build frontend
cd frontend
npm run build
cd ..

# Build backend (includes frontend)
cd backend
CGO_ENABLED=1 go build -o kantent .
```

## Deployment

### Unraid

1. Copy the project to your Unraid server
2. Navigate to the project directory
3. Run `docker-compose up -d`
4. Access at `http://[UNRAID-IP]:8080`

### Self-Hosted Server

The Docker image works on any server with Docker installed:

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e DATABASE_PATH=/data/rss-reader.db \
  --name kantent \
  kantent:latest
```

## Troubleshooting

### Database locked errors

The SQLite database uses WAL mode for better concurrency. If you encounter locking issues:

1. Make sure only one instance is running
2. Check file permissions on the data directory
3. Restart the container: `docker-compose restart`

### Feeds not updating

Check the feed fetch interval and logs:

```bash
docker-compose logs -f
```

### Images not loading

Direct image loading may fail due to:
- CORS restrictions from the source
- HTTPS/HTTP mixed content
- Dead links

This is expected behavior for some feeds.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with inspiration from Feedly, Inoreader, and Pinterest
- RSS parsing powered by [gofeed](https://github.com/mmcdole/gofeed)
- UI components styled with [Tailwind CSS](https://tailwindcss.com/)
