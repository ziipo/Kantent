package worker

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ziipo/Kantent/models"
	"github.com/ziipo/Kantent/services"
)

func StartFeedWorker(db *sql.DB) {
	// Get fetch interval from environment or use default (30 minutes)
	intervalStr := os.Getenv("FEED_FETCH_INTERVAL")
	interval := 1800 // Default 30 minutes in seconds
	if intervalStr != "" {
		if i, err := strconv.Atoi(intervalStr); err == nil {
			interval = i
		}
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	log.Printf("Feed worker started (interval: %d seconds)", interval)

	// Initial fetch on startup (with delay to let server start)
	time.Sleep(5 * time.Second)
	fetchAllFeeds(db)

	// Periodic fetches
	for range ticker.C {
		fetchAllFeeds(db)
	}
}

func fetchAllFeeds(db *sql.DB) {
	log.Println("Starting feed fetch cycle...")
	feeds, err := services.GetAllFeeds(db)
	if err != nil {
		log.Printf("Error getting feeds: %v", err)
		return
	}

	if len(feeds) == 0 {
		log.Println("No feeds to fetch")
		return
	}

	log.Printf("Fetching %d feeds...", len(feeds))
	for _, feed := range feeds {
		// Fetch each feed in a goroutine for concurrent fetching
		go func(f models.Feed) {
			if err := services.FetchFeed(db, f.ID, f.URL); err != nil {
				log.Printf("Failed to fetch feed %d (%s): %v", f.ID, f.Title, err)
			}
		}(feed)
	}
}
