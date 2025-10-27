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
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("Warning: Could not enable WAL mode: %v", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		log.Printf("Warning: Could not set synchronous mode: %v", err)
	}
	if _, err := db.Exec("PRAGMA cache_size=-64000"); err != nil {
		log.Printf("Warning: Could not set cache size: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		log.Printf("Warning: Could not enable foreign keys: %v", err)
	}

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
