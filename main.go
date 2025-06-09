package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gotail/db"
	"gotail/handlers"
	"gotail/middleware"
)

func main() {
	// Load environment variables from .env
	_ = godotenv.Load()

	// Get DB driver and connection string from env
	driver := os.Getenv("DB_DRIVER") // e.g., "sqlite" or "postgres"
	dsn := os.Getenv("DB_DSN")       // e.g., "logs.db" or Postgres URL

	if driver == "" || dsn == "" {
		log.Fatal("DB_DRIVER and DB_DSN must be set in .env")
	}

	// Initialize the correct DB store based on driver
	store, err := db.New(driver, dsn)
	if err != nil {
		log.Fatal("Failed to create DB store:", err)
	}
	defer store.Close()

	// Create schema
	if err := store.Init(); err != nil {
		log.Fatal("DB schema creation failed:", err)
	}

	// Create log handler with store dependency
	handler := &handlers.LogHandler{Store: store}

	// Register route with API key protection
	http.Handle("/log", middleware.RequireAPIKey(http.HandlerFunc(handler.HandleLog)))
	http.Handle("/logs", middleware.RequireAPIKey(http.HandlerFunc(handler.HandleGetLogs)))

	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
