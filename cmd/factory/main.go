package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type LogEntry struct {
	ID             string
	Timestamp      time.Time
	SeverityText   string
	SeverityNumber int
	Body           string
	TraceID        *string
	SpanID         *string
	Attributes     map[string]string
}

type LogAttribute struct {
	LogID string
	Key   string
	Value string
}

var (
	severityLevels = []struct {
		text   string
		number int
	}{
		{"TRACE", 1},
		{"DEBUG", 5},
		{"INFO", 9},
		{"WARN", 13},
		{"ERROR", 17},
		{"FATAL", 21},
	}

	logMessages = []string{
		"User authentication successful",
		"Database connection established",
		"Request processing completed",
		"Cache miss for key: %s",
		"API rate limit exceeded",
		"File upload completed successfully",
		"Payment processing initiated",
		"Email notification sent",
		"Background job started",
		"Configuration loaded",
		"Health check passed",
		"Metrics collection updated",
		"Session expired for user",
		"Connection timeout occurred",
		"Invalid request format",
		"Resource not found",
		"Permission denied",
		"Service unavailable",
		"Internal server error",
		"Data validation failed",
	}

	attributeKeys = []string{
		"service.name",
		"service.version",
		"http.method",
		"http.status_code",
		"user.id",
		"request.id",
		"database.name",
		"error.type",
		"environment",
		"region",
	}

	attributeValues = map[string][]string{
		"service.name":     {"auth-service", "user-service", "payment-service", "notification-service"},
		"service.version":  {"1.0.0", "1.1.0", "2.0.0", "2.1.0"},
		"http.method":      {"GET", "POST", "PUT", "DELETE", "PATCH"},
		"http.status_code": {"200", "201", "400", "401", "403", "404", "500", "503"},
		"user.id":          {"user-123", "user-456", "user-789", "user-101"},
		"request.id":       {},
		"database.name":    {"users", "orders", "products", "sessions"},
		"error.type":       {"ValidationError", "ConnectionError", "TimeoutError", "AuthError"},
		"environment":      {"development", "staging", "production"},
		"region":           {"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"},
	}
)

func main() {
	var (
		dbPath = flag.String("db", "logs.db", "Database file path")
		count  = flag.Int("count", 100, "Number of logs to generate")
		help   = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("Log Factory - Generate mock logs for testing")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *count <= 0 {
		log.Fatal("Count must be greater than 0")
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	if err := generateLogs(db, *count); err != nil {
		log.Fatalf("Failed to generate logs: %v", err)
	}

	fmt.Printf("Successfully generated %d mock logs in %s\n", *count, *dbPath)
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		ts DATETIME NOT NULL,
		severity_text TEXT NOT NULL,
		severity_number INT,
		body TEXT NOT NULL,
		trace_id TEXT,
		span_id TEXT
		);

		CREATE TABLE IF NOT EXISTS log_attributes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		log_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		FOREIGN KEY (log_id) REFERENCES logs(id)
		);

		CREATE INDEX IF NOT EXISTS idx_attr_key_value ON log_attributes(key, value);
	`)

	return err
}

func generateLogs(db *sql.DB, count int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	logStmt, err := tx.Prepare(`
		INSERT INTO logs (id, ts, severity_text, severity_number, body, trace_id, span_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare log statement: %w", err)
	}
	defer logStmt.Close()

	attrStmt, err := tx.Prepare(`
		INSERT INTO log_attributes (log_id, key, value)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare attribute statement: %w", err)
	}
	defer attrStmt.Close()

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		logEntry := generateLogEntry()
		
		// Insert log entry
		_, err := logStmt.Exec(
			logEntry.ID,
			logEntry.Timestamp,
			logEntry.SeverityText,
			logEntry.SeverityNumber,
			logEntry.Body,
			logEntry.TraceID,
			logEntry.SpanID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert log entry: %w", err)
		}

		// Insert attributes directly from the logEntry.Attributes map
		for key, value := range logEntry.Attributes {
			_, err := attrStmt.Exec(logEntry.ID, key, value)
			if err != nil {
				return fmt.Errorf("failed to insert log attribute: %w", err)
			}
		}

		if (i+1)%1000 == 0 {
			fmt.Printf("Generated %d/%d logs...\n", i+1, count)
		}
	}

	return tx.Commit()
}

func generateLogEntry() LogEntry {
	severity := severityLevels[rand.Intn(len(severityLevels))]
	
	// Generate timestamp within the last 30 days
	now := time.Now()
	randomHours := rand.Intn(24 * 30)
	timestamp := now.Add(-time.Duration(randomHours) * time.Hour)
	
	// Generate message
	message := logMessages[rand.Intn(len(logMessages))]
	if message == "Cache miss for key: %s" {
		message = fmt.Sprintf(message, uuid.New().String()[:8])
	}

	logEntry := LogEntry{
		ID:             uuid.New().String(),
		Timestamp:      timestamp,
		SeverityText:   severity.text,
		SeverityNumber: severity.number,
		Body:           message,
		Attributes:     make(map[string]string),
	}

	// 70% chance to have trace_id and span_id
	if rand.Float32() < 0.7 {
		traceID := uuid.New().String()
		spanID := uuid.New().String()[:16]
		logEntry.TraceID = &traceID
		logEntry.SpanID = &spanID
	}

	// Generate attributes directly into the map
	generateLogAttributes(&logEntry)

	return logEntry
}

func generateLogAttributes(logEntry *LogEntry) {
	// Generate 2-5 attributes per log
	numAttrs := rand.Intn(4) + 2
	usedKeys := make(map[string]bool)
	
	for i := 0; i < numAttrs; i++ {
		key := attributeKeys[rand.Intn(len(attributeKeys))]
		
		// Avoid duplicate keys
		if usedKeys[key] {
			continue
		}
		usedKeys[key] = true
		
		var value string
		if values, exists := attributeValues[key]; exists && len(values) > 0 {
			value = values[rand.Intn(len(values))]
		} else {
			// Generate random value for keys without predefined values
			value = uuid.New().String()[:8]
		}
		
		logEntry.Attributes[key] = value
	}
}