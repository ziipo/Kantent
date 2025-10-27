package models

import "time"

type Feed struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	URL           string    `json:"url"`
	SiteURL       string    `json:"site_url"`
	Description   string    `json:"description"`
	FetchInterval int       `json:"fetch_interval"`
	LastFetched   *time.Time `json:"last_fetched"`
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
