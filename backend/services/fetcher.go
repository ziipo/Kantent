package services

import (
	"database/sql"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/ziipo/Kantent/models"
)

func FetchFeed(db *sql.DB, feedID int, feedURL string) error {
	parser := gofeed.NewParser()
	parsedFeed, err := parser.ParseURL(feedURL)
	if err != nil {
		updateFeedError(db, feedID, err.Error())
		return err
	}

	// Update feed metadata if we got a title
	if parsedFeed.Title != "" {
		db.Exec("UPDATE feeds SET title = ?, description = ?, site_url = ? WHERE id = ?",
			parsedFeed.Title, parsedFeed.Description, parsedFeed.Link, feedID)
	}

	for _, item := range parsedFeed.Items {
		article := models.Article{
			FeedID:      feedID,
			GUID:        getGUID(item),
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

	log.Printf("Fetched feed %d: %s (%d items)", feedID, parsedFeed.Title, len(parsedFeed.Items))
	return nil
}

func getGUID(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	// Fallback to link if no GUID
	return item.Link
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
	cleaned := strings.TrimSpace(re.ReplaceAllString(html, " "))
	// Remove multiple spaces
	re = regexp.MustCompile(`\s+`)
	cleaned = re.ReplaceAllString(cleaned, " ")
	return cleaned
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
	log.Printf("Error fetching feed %d: %s", feedID, errMsg)
}

func GetAllFeeds(db *sql.DB) ([]models.Feed, error) {
	query := `SELECT id, title, url, site_url, description, fetch_interval,
                     last_fetched, last_error, created_at, updated_at
              FROM feeds`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var f models.Feed
		err := rows.Scan(&f.ID, &f.Title, &f.URL, &f.SiteURL, &f.Description,
			&f.FetchInterval, &f.LastFetched, &f.LastError, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning feed: %v", err)
			continue
		}
		feeds = append(feeds, f)
	}

	return feeds, nil
}
