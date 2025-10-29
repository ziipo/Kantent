package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ziipo/Kantent/models"
	"github.com/ziipo/Kantent/services"
)

// HandleListArticles returns paginated articles with optional filters
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

// HandleGetArticle returns a single article by ID
func HandleGetArticle(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		var a models.Article
		err := db.QueryRow(`
            SELECT a.id, a.feed_id, f.title as feed_title, a.guid, a.title,
                   a.url, a.description, a.content, a.author, a.published_at,
                   a.is_read, a.is_starred, a.image_url
            FROM articles a
            JOIN feeds f ON a.feed_id = f.id
            WHERE a.id = ?
        `, id).Scan(
			&a.ID, &a.FeedID, &a.FeedTitle, &a.GUID, &a.Title,
			&a.URL, &a.Description, &a.Content, &a.Author, &a.PublishedAt,
			&a.IsRead, &a.IsStarred, &a.ImageURL,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Article not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(a)
	}
}

// HandleListFeeds returns all feeds
func HandleListFeeds(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feeds, err := services.GetAllFeeds(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(feeds)
	}
}

// HandleGetFeed returns a single feed by ID
func HandleGetFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		var f models.Feed
		err := db.QueryRow(`
            SELECT id, title, url, site_url, description, fetch_interval,
                   last_fetched, last_error, created_at, updated_at
            FROM feeds WHERE id = ?
        `, id).Scan(&f.ID, &f.Title, &f.URL, &f.SiteURL, &f.Description,
			&f.FetchInterval, &f.LastFetched, &f.LastError, &f.CreatedAt, &f.UpdatedAt)

		if err == sql.ErrNoRows {
			http.Error(w, "Feed not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(f)
	}
}

// HandleCreateFeed creates a new feed
func HandleCreateFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var feed models.Feed
		if err := json.NewDecoder(r.Body).Decode(&feed); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if feed.URL == "" {
			http.Error(w, "Feed URL is required", http.StatusBadRequest)
			return
		}

		// Set default title if not provided
		if feed.Title == "" {
			feed.Title = "New Feed"
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

		// Immediately fetch the feed to populate articles
		go services.FetchFeed(db, feed.ID, feed.URL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(feed)
	}
}

// HandleUpdateFeed updates an existing feed
func HandleUpdateFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		var feed models.Feed
		if err := json.NewDecoder(r.Body).Decode(&feed); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec(
			"UPDATE feeds SET title = ?, url = ?, site_url = ?, description = ?, fetch_interval = ? WHERE id = ?",
			feed.Title, feed.URL, feed.SiteURL, feed.Description, feed.FetchInterval, id,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// HandleDeleteFeed deletes a feed
func HandleDeleteFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		_, err := db.Exec("DELETE FROM feeds WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleRefreshFeed manually triggers a feed refresh
func HandleRefreshFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		feedID, _ := strconv.Atoi(id)

		var feedURL string
		err := db.QueryRow("SELECT url FROM feeds WHERE id = ?", id).Scan(&feedURL)
		if err == sql.ErrNoRows {
			http.Error(w, "Feed not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Trigger refresh in background
		go services.FetchFeed(db, feedID, feedURL)

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "refresh started"})
	}
}

// HandleMarkRead marks an article as read/unread
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

// HandleStarArticle stars/unstars an article
func HandleStarArticle(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		var req struct {
			IsStarred bool `json:"is_starred"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE articles SET is_starred = ? WHERE id = ?", req.IsStarred, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// HandleMarkAllRead marks all articles as read
func HandleMarkAllRead(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feedID := r.URL.Query().Get("feed_id")

		query := "UPDATE articles SET is_read = 1"
		args := []interface{}{}

		if feedID != "" {
			query += " WHERE feed_id = ?"
			args = append(args, feedID)
		}

		_, err := db.Exec(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// HandleGetStats returns statistics
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


// HandleDiscoverFeeds discovers RSS/Atom feeds from a website URL
func HandleDiscoverFeeds(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		websiteURL := r.URL.Query().Get("url")
		if websiteURL == "" {
			http.Error(w, "URL parameter is required", http.StatusBadRequest)
			return
		}

		candidates, err := services.DiscoverFeeds(websiteURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to discover feeds: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(candidates)
	}
}

// HandleResolveYouTubeChannel resolves a YouTube channel handle/URL to a channel ID
func HandleResolveYouTubeChannel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := r.URL.Query().Get("input")
		if input == "" {
			http.Error(w, "input parameter is required", http.StatusBadRequest)
			return
		}

		channelID, err := services.ResolveYouTubeChannelID(input)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to resolve YouTube channel: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"channel_id": channelID,
			"rss_url":    fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelID),
		})
	}
}
