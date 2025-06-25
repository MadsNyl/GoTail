package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"gotail/db"
	"gotail/middleware"
	"gotail/handlers/html"
	"gotail/handlers/logging"
)

func main() {
	// Load environment variables from .env
	_ = godotenv.Load()

	// Get DB driver and connection string from env
	driver := os.Getenv("DB_DRIVER") // e.g., "sqlite" or "postgres"
	dsn := os.Getenv("DB_DSN")       // e.g., "logs.db" or Postgres URL
	user := os.Getenv("UI_USERNAME") // Username for UI authentication
  	pass := os.Getenv("UI_PASSWORD") // Password for UI authentication

	if user == "" || pass == "" {
		log.Fatal("UI_USER and UI_PASS must be set")
	}

	if driver == "" || dsn == "" {
		log.Fatal("DB_DRIVER and DB_DSN must be set in .env")
	}

	// Initialize the correct DB store based on driver
	store, err := db.New(driver, dsn)
	if err != nil {
		log.Fatal("Failed to create DB store:", err)
	}
	defer store.Close()

	// Create log handler with store dependency
	handler := &logging.LogHandler{Store: store}
	// Create HTML handler with store dependency
	htmlHandler := &html.HTMLHandler{Store: store}

	// Route for submitting logs (POST)
	http.Handle("/log", middleware.BasicAuth(user, pass)(http.HandlerFunc(handler.HandleLogInsert)))

	// Route for HTML page
	http.Handle("/", middleware.BasicAuth(user, pass)(http.HandlerFunc(htmlHandler.HandleLogsPage)))


	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
