package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"gotail/db"
	"gotail/handlers/api"
	"gotail/handlers/html"
	"gotail/handlers/logging"
	"gotail/middleware"
)

func main() {
	// Load environment variables from .env
	_ = godotenv.Load()

	// Get DB driver and connection string from env
	driver := os.Getenv("DB_DRIVER")     // e.g., "sqlite" or "postgres"
	dsn := os.Getenv("DB_DSN")           // e.g., "logs.db" or Postgres URL
	user := os.Getenv("UI_USERNAME")     // Username for UI authentication
	pass := os.Getenv("UI_PASSWORD")     // Password for UI authentication
	apiKeys := os.Getenv("GOTAIL_API_KEYS")
	headlessMode := strings.ToLower(os.Getenv("HEADLESS_MODE")) == "true"

	// Reload API keys after godotenv.Load() since init() runs before env is loaded
	middleware.ReloadAPIKeys()

	// Only require UI credentials if not in headless mode
	if !headlessMode && (user == "" || pass == "") {
		log.Fatal("UI_USERNAME and UI_PASSWORD must be set (or enable HEADLESS_MODE)")
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

	// Create handlers with store dependency
	logHandler := &logging.LogHandler{Store: store}
	htmlHandler := &html.HTMLHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	// API documentation (no auth required)
	http.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		api.HandleDocs(w, r)
	})
	http.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.ServeFile(w, r, "openapi.yaml")
	})

	// Route for submitting logs (POST) - accepts both API key and basic auth
	if user != "" && pass != "" {
		http.Handle("/log", middleware.EitherAuth(user, pass)(http.HandlerFunc(logHandler.HandleLogInsert)))
	} else if apiKeys != "" {
		// Headless mode with only API key auth for /log
		http.Handle("/log", middleware.APIKeyAuth()(http.HandlerFunc(logHandler.HandleLogInsert)))
	}

	// API routes (available if API keys are configured)
	if apiKeys != "" {
		http.Handle("/api/logs", middleware.APIKeyAuth()(http.HandlerFunc(apiHandler.HandleLogsAPI)))
		http.Handle("/api/stats", middleware.APIKeyAuth()(http.HandlerFunc(apiHandler.HandleStatsAPI)))
		http.Handle("/api/attributes", middleware.APIKeyAuth()(http.HandlerFunc(apiHandler.HandleAttributeKeysAPI)))
		http.Handle("/api/services", middleware.APIKeyAuth()(http.HandlerFunc(apiHandler.HandleServicesAPI)))
	}

	// UI routes (only if not in headless mode)
	if !headlessMode {
		http.Handle("/", middleware.BasicAuth(user, pass)(http.HandlerFunc(htmlHandler.HandleLogsPage)))
		http.Handle("/stats", middleware.BasicAuth(user, pass)(http.HandlerFunc(htmlHandler.HandleLogStatsPage)))
	}

	log.Println("Listening on :8080...")
	if headlessMode {
		log.Println("Running in headless mode (UI disabled)")
	}
	log.Fatal(http.ListenAndServe(":8080", nil))
}
