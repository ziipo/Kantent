package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ziipo/Kantent/api"
	"github.com/ziipo/Kantent/db"
	"github.com/ziipo/Kantent/worker"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/rss-reader.db"
	}

	// Initialize database
	database := db.InitDatabase(dbPath)
	defer database.Close()

	// Start background worker for feed fetching
	go worker.StartFeedWorker(database)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Feed routes
		r.Get("/feeds", api.HandleListFeeds(database))
		r.Post("/feeds", api.HandleCreateFeed(database))
		r.Get("/feeds/{id}", api.HandleGetFeed(database))
		r.Put("/feeds/{id}", api.HandleUpdateFeed(database))
		r.Delete("/feeds/{id}", api.HandleDeleteFeed(database))
		r.Post("/feeds/{id}/refresh", api.HandleRefreshFeed(database))

		// Article routes
		r.Get("/articles", api.HandleListArticles(database))
		r.Get("/articles/{id}", api.HandleGetArticle(database))
		r.Put("/articles/{id}/read", api.HandleMarkRead(database))
		r.Put("/articles/{id}/star", api.HandleStarArticle(database))
		r.Post("/articles/mark-all-read", api.HandleMarkAllRead(database))

		// Stats
		r.Get("/stats", api.HandleGetStats(database))
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Serve frontend static files from filesystem (for development)
	frontendPath := "../frontend/dist"
	if _, err := os.Stat(frontendPath); err == nil {
		log.Printf("Serving frontend from: %s", frontendPath)
		r.Handle("/*", http.FileServer(http.Dir(frontendPath)))
	} else {
		log.Printf("Frontend not found at %s - API only mode", frontendPath)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	log.Printf("Database: %s", dbPath)
	log.Printf("API available at http://localhost:%s/api", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
