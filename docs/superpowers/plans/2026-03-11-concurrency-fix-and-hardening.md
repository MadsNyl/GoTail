# Concurrency Fix & Security Hardening Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix SQLite concurrent request conflicts and add security hardening (input validation, panic recovery, auth logging, stats date filtering).

**Architecture:** Dual SQLite connection pools (read/write separation) with WAL mode. New middleware for body size limits and panic recovery. Field validation in insert handler. Proper date range queries in stats.

**Tech Stack:** Go 1.23, modernc.org/sqlite, standard net/http, sync.Mutex

**Spec:** `docs/superpowers/specs/2026-03-11-concurrency-fix-and-hardening-design.md`

---

## Chunk 1: SQLite Store Restructure

### Task 1: Restructure SQLiteStore with dual connection pools

**Files:**
- Modify: `db/sqlite/store.go`
- Modify: `main.go:42-46` (remove broken busy_timeout logic)

- [ ] **Step 1: Update SQLiteStore struct and constructor**

Replace `db/sqlite/store.go` entirely:

```go
package sqlite

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	writeDB *sql.DB
	readDB  *sql.DB
	mu      sync.Mutex
}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	writeDSN := dsn + "?_busy_timeout=5000"
	readDSN := dsn + "?_busy_timeout=5000"

	writeDB, err := sql.Open("sqlite", writeDSN)
	if err != nil {
		return nil, fmt.Errorf("open write db: %w", err)
	}
	writeDB.SetMaxOpenConns(1)

	// Enable WAL mode
	if _, err := writeDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := writeDB.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("set synchronous: %w", err)
	}
	if _, err := writeDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	readDB, err := sql.Open("sqlite", readDSN)
	if err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("open read db: %w", err)
	}
	readDB.SetMaxOpenConns(8)

	// Make read connection read-only
	if _, err := readDB.Exec("PRAGMA query_only=ON"); err != nil {
		writeDB.Close()
		readDB.Close()
		return nil, fmt.Errorf("set query_only: %w", err)
	}

	return &SQLiteStore{writeDB: writeDB, readDB: readDB}, nil
}

func (s *SQLiteStore) Close() error {
	writeErr := s.writeDB.Close()
	readErr := s.readDB.Close()
	if writeErr != nil {
		return writeErr
	}
	return readErr
}
```

- [ ] **Step 2: Remove broken busy_timeout logic from main.go**

In `main.go`, remove lines 42-46:

```go
	// Set busy timeout for SQLite if using SQLite
	// This is important to avoid database lock issues
	if dsn == "sqlite" {
		dsn = dsn + "?_busy_timeout=5000"
	}
```

- [ ] **Step 3: Update InsertLog to use writeDB**

In `db/sqlite/insert.go`, change line 13 from `s.db.Begin()` to `s.writeDB.Begin()`:

```go
func (s *SQLiteStore) InsertLog(entry models.LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.writeDB.Begin()
```

- [ ] **Step 4: Update GetLogsFiltered to use readDB with transaction**

In `db/sqlite/get.go`, add `"context"` and `"database/sql"` to imports, then wrap the multi-query read in a read-only transaction. Replace the method body:

```go
func (s *SQLiteStore) GetLogsFiltered(
	page int,
	limit int,
	severity string,
	attrKey string,
	attrValue string,
	service string,
) ([]models.LogEntry, int, error) {
	offset := (page - 1) * limit

	tx, err := s.readDB.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var (
		whereClauses []string
		args         []interface{}
	)

	query := `
		SELECT l.*
		FROM log l`

	if attrKey != "" && attrValue != "" {
		query += `
			INNER JOIN attribute a ON a.log_id = l.id`
		whereClauses = append(whereClauses, "a.key = ? AND a.value LIKE ?")
		args = append(args, attrKey, "%"+attrValue+"%")
	}

	if severity != "" {
		whereClauses = append(whereClauses, "l.severity_text = ?")
		args = append(args, severity)
	}

	if service != "" {
		whereClauses = append(whereClauses, "l.service_name = ?")
		args = append(args, service)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY l.timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Count query
	countQuery := `
		SELECT COUNT(DISTINCT l.id)
		FROM log l`
	if attrKey != "" && attrValue != "" {
		countQuery += `
			INNER JOIN attribute a ON a.log_id = l.id`
	}
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var count int
	countArgs := args[:len(args)-2]
	if err := tx.QueryRow(countQuery, countArgs...).Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	loc, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		loc = time.FixedZone("CET", 1*60*60)
	}

	var logs []models.LogEntry
	for rows.Next() {
		var entry models.LogEntry

		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.SeverityText,
			&entry.SeverityNumber,
			&entry.Body,
			&entry.ServiceName,
			&entry.ServiceVersion,
			&entry.ServiceInstanceID,
			&entry.HostName,
			&entry.ScopeName,
			&entry.ScopeVersion,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		entry.Timestamp = entry.Timestamp.In(loc)

		attrRows, err := tx.Query(`SELECT key, value FROM attribute WHERE log_id = ?`, entry.ID)
		if err != nil {
			return nil, 0, err
		}

		entry.Attributes = make(map[string]any)
		for attrRows.Next() {
			var k string
			var v any
			if err := attrRows.Scan(&k, &v); err != nil {
				attrRows.Close()
				return nil, 0, err
			}
			entry.Attributes[k] = v
		}
		attrRows.Close()

		logs = append(logs, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}
```

- [ ] **Step 5: Update simple read methods to use readDB**

In `db/sqlite/get.go`, update `GetAttributeKeys`, `GetTotalLogs`, `GetServices` — replace all `s.db.` with `s.readDB.` (no transaction needed for single queries).

- [ ] **Step 6: Update stats methods to use readDB**

In `db/sqlite/stats.go`, replace all `s.db.` with `s.readDB.` in all five methods.

- [ ] **Step 7: Verify it compiles**

Run: `cd /Users/madsnylund/Documents/gotail && go build ./...`
Expected: no errors

- [ ] **Step 8: Commit**

```bash
git add db/sqlite/store.go db/sqlite/insert.go db/sqlite/get.go db/sqlite/stats.go main.go
git commit -m "feat: WAL mode with dual read/write connection pools

Fixes #9 - separates read and write SQLite connections to prevent
transaction conflicts during concurrent requests."
```

---

## Chunk 2: Input Validation

### Task 2: Add MaxBodySize middleware

**Files:**
- Create: `middleware/max_body_size.go`

- [ ] **Step 1: Create the middleware**

```go
package middleware

import (
	"net/http"
)

func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 2: Apply middleware in main.go**

In `main.go`, wrap the `/log` handler with `MaxBodySize(1 << 20)` (1MB). Replace the route registration for `/log`:

Where the `/log` routes are registered (lines 79-84), wrap the handler:

```go
logHandlerWithLimit := middleware.MaxBodySize(1 << 20)(http.HandlerFunc(logHandler.HandleLogInsert))

if user != "" && pass != "" {
    http.Handle("/log", middleware.EitherAuth(user, pass)(logHandlerWithLimit))
} else if apiKeys != "" {
    http.Handle("/log", middleware.APIKeyAuth()(logHandlerWithLimit))
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /Users/madsnylund/Documents/gotail && go build ./...`

### Task 3: Add field validation to insert handler

**Files:**
- Modify: `handlers/logging/insert.go`

- [ ] **Step 1: Add validation after JSON parsing**

In `handlers/logging/insert.go`, add validation between the JSON unmarshal (line 34) and the ID assignment (line 36). Add a `validateLogEntry` function:

```go
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"gotail/db"
	"gotail/models"
)

const (
	maxBodyLen     = 65536 // 64KB
	maxStringLen   = 256
	maxAttributes  = 50
	maxAttrKeyLen  = 256
	maxAttrValueLen = 4096 // 4KB
)

type LogHandler struct {
	Store db.LogStore
}

func (h *LogHandler) HandleLogInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	var logEntry models.LogEntry
	if err := json.Unmarshal(body, &logEntry); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if msg := validateLogEntry(logEntry); msg != "" {
		writeJSONError(w, msg, http.StatusBadRequest)
		return
	}

	logEntry.ID = uuid.New().String()

	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now().UTC()
	}

	err = h.Store.InsertLog(logEntry)
	if err != nil {
		log.Printf("Failed to insert log: %v", err)
		writeJSONError(w, "Failed to insert log", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateLogEntry(e models.LogEntry) string {
	if len(e.Body) > maxBodyLen {
		return fmt.Sprintf("body exceeds maximum length of %d bytes", maxBodyLen)
	}
	if len(e.SeverityText) > maxStringLen {
		return fmt.Sprintf("severity_text exceeds maximum length of %d chars", maxStringLen)
	}
	if e.ServiceName != nil && len(*e.ServiceName) > maxStringLen {
		return fmt.Sprintf("service_name exceeds maximum length of %d chars", maxStringLen)
	}
	if e.HostName != nil && len(*e.HostName) > maxStringLen {
		return fmt.Sprintf("host_name exceeds maximum length of %d chars", maxStringLen)
	}
	if e.ScopeName != nil && len(*e.ScopeName) > maxStringLen {
		return fmt.Sprintf("scope_name exceeds maximum length of %d chars", maxStringLen)
	}
	if len(e.Attributes) > maxAttributes {
		return fmt.Sprintf("attributes exceeds maximum count of %d", maxAttributes)
	}
	for k, v := range e.Attributes {
		if len(k) > maxAttrKeyLen {
			return fmt.Sprintf("attribute key %q exceeds maximum length of %d chars", k, maxAttrKeyLen)
		}
		if str, ok := v.(string); ok && len(str) > maxAttrValueLen {
			return fmt.Sprintf("attribute value for key %q exceeds maximum length of %d bytes", k, maxAttrValueLen)
		}
	}
	return ""
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/madsnylund/Documents/gotail && go build ./...`

- [ ] **Step 3: Commit**

```bash
git add middleware/max_body_size.go handlers/logging/insert.go main.go
git commit -m "feat: add input validation and body size limits on POST /log"
```

---

## Chunk 3: Panic Recovery & Auth Logging

### Task 4: Add panic recovery middleware

**Files:**
- Create: `middleware/recovery.go`
- Modify: `main.go`

- [ ] **Step 1: Create recovery middleware**

```go
package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n%s", err, debug.Stack())
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 2: Apply recovery middleware in main.go**

In `main.go`, wrap the default mux with recovery. Replace the final `http.ListenAndServe` line:

```go
log.Fatal(http.ListenAndServe(":8080", middleware.Recovery(http.DefaultServeMux)))
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /Users/madsnylund/Documents/gotail && go build ./...`

- [ ] **Step 4: Commit**

```bash
git add middleware/recovery.go main.go
git commit -m "feat: add panic recovery middleware"
```

### Task 5: Add failed auth attempt logging

**Files:**
- Modify: `middleware/basic_auth.go`
- Modify: `middleware/api_key_auth.go`

- [ ] **Step 1: Add logging to BasicAuth**

In `middleware/basic_auth.go`, add `"log"` to imports and log on failed auth. The two failure paths are: `!ok` (no credentials) and failed comparison. Update both to log before calling `unauthorized(w)`:

```go
package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
)

func BasicAuth(expectedUser, expectedPass string) func(http.Handler) http.Handler {
	userHash := sha256.Sum256([]byte(expectedUser))
	passHash := sha256.Sum256([]byte(expectedPass))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok {
				log.Printf("AUTH FAILURE: basic_auth missing credentials, ip=%s path=%s", r.RemoteAddr, r.URL.Path)
				unauthorized(w)
				return
			}

			uHash := sha256.Sum256([]byte(u))
			pHash := sha256.Sum256([]byte(p))
			if subtle.ConstantTimeCompare(uHash[:], userHash[:]) == 1 &&
				subtle.ConstantTimeCompare(pHash[:], passHash[:]) == 1 {
				next.ServeHTTP(w, r)
				return
			}

			log.Printf("AUTH FAILURE: basic_auth invalid credentials, ip=%s path=%s", r.RemoteAddr, r.URL.Path)
			unauthorized(w)
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="GoTail UI", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
```

- [ ] **Step 2: Add logging to APIKeyAuth and EitherAuth**

In `middleware/api_key_auth.go`, add `"log"` to imports. Update `APIKeyAuth` and `EitherAuth` failure paths:

In `APIKeyAuth` handler (around lines 44-52):

```go
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    key := extractAPIKey(r)
    if key == "" {
        log.Printf("AUTH FAILURE: api_key missing, ip=%s path=%s", r.RemoteAddr, r.URL.Path)
        unauthorizedJSON(w, "API key required")
        return
    }

    if !validateAPIKey(key) {
        log.Printf("AUTH FAILURE: api_key invalid, ip=%s path=%s", r.RemoteAddr, r.URL.Path)
        unauthorizedJSON(w, "Invalid API key")
        return
    }

    next.ServeHTTP(w, r)
})
```

In `EitherAuth`, at the final failure (line 88):

```go
// Neither authentication method succeeded
log.Printf("AUTH FAILURE: either_auth no valid method, ip=%s path=%s", r.RemoteAddr, r.URL.Path)
unauthorized(w)
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /Users/madsnylund/Documents/gotail && go build ./...`

- [ ] **Step 4: Commit**

```bash
git add middleware/basic_auth.go middleware/api_key_auth.go
git commit -m "feat: log failed authentication attempts"
```

---

## Chunk 4: Stats Date Filtering Fix

### Task 6: Replace LIKE with date range queries in stats

**Files:**
- Modify: `db/sqlite/stats.go`

- [ ] **Step 1: Rewrite all stats methods with proper date ranges**

Replace `db/sqlite/stats.go` entirely:

```go
package sqlite

import (
	"time"
)

func dateRange(year int, month int) (string, string) {
	loc, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		loc = time.FixedZone("CET", 1*60*60)
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)
	return start.Format(time.RFC3339), end.Format(time.RFC3339)
}

func (s *SQLiteStore) CountLogsByMonth(year int, month int) (int, error) {
	start, end := dateRange(year, month)
	var count int
	err := s.readDB.QueryRow(
		"SELECT COUNT(*) FROM log WHERE timestamp >= ? AND timestamp < ?",
		start, end,
	).Scan(&count)
	return count, err
}

func (s *SQLiteStore) CountLogsBySeverity(year int, month int) (map[string]int, error) {
	start, end := dateRange(year, month)
	rows, err := s.readDB.Query(`
		SELECT severity_text, COUNT(*)
		FROM log
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY severity_text`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, err
		}
		result[severity] = count
	}
	return result, rows.Err()
}

func (s *SQLiteStore) CountLogsPerDay(year int, month int) (map[int]int, error) {
	start, end := dateRange(year, month)
	rows, err := s.readDB.Query(`
		SELECT CAST(strftime('%d', timestamp) AS INTEGER) AS day, COUNT(*)
		FROM log
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY day
		ORDER BY day`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var day int
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		result[day] = count
	}
	return result, rows.Err()
}

func (s *SQLiteStore) CountLogsByService(year int, month int) (map[string]int, error) {
	start, end := dateRange(year, month)
	rows, err := s.readDB.Query(`
		SELECT service_name, COUNT(*)
		FROM log
		WHERE timestamp >= ? AND timestamp < ? AND service_name IS NOT NULL
		GROUP BY service_name`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var service string
		var count int
		if err := rows.Scan(&service, &count); err != nil {
			return nil, err
		}
		result[service] = count
	}
	return result, rows.Err()
}

func (s *SQLiteStore) CountLogsByAttribute(year int, month int) (map[string]int, error) {
	start, end := dateRange(year, month)
	rows, err := s.readDB.Query(`
		SELECT key, COUNT(DISTINCT log_id)
		FROM attribute
		WHERE log_id IN (
			SELECT id FROM log WHERE timestamp >= ? AND timestamp < ?
		)
		GROUP BY key`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var key string
		var count int
		if err := rows.Scan(&key, &count); err != nil {
			return nil, err
		}
		result[key] = count
	}
	return result, rows.Err()
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/madsnylund/Documents/gotail && go build ./...`

- [ ] **Step 3: Commit**

```bash
git add db/sqlite/stats.go
git commit -m "fix: replace LIKE with date range queries in stats"
```

---

## Chunk 5: Integration & Concurrency Tests

### Task 7: Create test helpers

**Files:**
- Create: `db/sqlite/testhelper_test.go`

- [ ] **Step 1: Create test helper with DB setup/teardown**

This helper creates a temp SQLite DB, applies migrations from the SQL file, and returns a ready store.

```go
package sqlite_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotail/db/sqlite"
)

func newTestStore(t *testing.T) *sqlite.SQLiteStore {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	store, err := sqlite.NewSQLiteStore(dsn)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Apply migration SQL directly
	migrationSQL, err := os.ReadFile("../../migrations/20250625192434_log.sql")
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	// Extract only the Up migration (between +goose Up and +goose Down)
	sql := extractUpMigration(string(migrationSQL))

	if err := store.ExecRaw(sql); err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	t.Cleanup(func() {
		store.Close()
	})

	return store
}

func extractUpMigration(content string) string {
	inUp := false
	inStatement := false
	var result string
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "+goose Up") {
			inUp = true
			continue
		}
		if strings.Contains(line, "+goose Down") {
			break
		}
		if inUp && strings.Contains(line, "+goose StatementBegin") {
			inStatement = true
			continue
		}
		if inUp && strings.Contains(line, "+goose StatementEnd") {
			inStatement = false
			continue
		}
		if inUp && inStatement {
			result += line + "\n"
		}
	}
	return result
}
```

- [ ] **Step 2: Add ExecRaw method to SQLiteStore**

In `db/sqlite/store.go`, add a method for running raw SQL (used by tests for migrations):

```go
func (s *SQLiteStore) ExecRaw(sql string) error {
	_, err := s.writeDB.Exec(sql)
	return err
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /Users/madsnylund/Documents/gotail && go build ./...`

### Task 8: Write store-level tests

**Files:**
- Create: `db/sqlite/store_test.go`

- [ ] **Step 1: Write basic CRUD and concurrency tests**

```go
package sqlite_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"gotail/models"
)

func makeLogEntry(id string, svc string) models.LogEntry {
	s := svc
	return models.LogEntry{
		ID:           id,
		Timestamp:    time.Now().UTC(),
		SeverityText: "INFO",
		SeverityNumber: 9,
		Body:         "test log body",
		ServiceName:  &s,
		Attributes:   map[string]any{"env": "test"},
	}
}

func TestInsertAndGetLogs(t *testing.T) {
	store := newTestStore(t)

	entry := makeLogEntry("test-1", "svc-a")
	if err := store.InsertLog(entry); err != nil {
		t.Fatalf("InsertLog failed: %v", err)
	}

	logs, count, err := store.GetLogsFiltered(1, 10, "", "", "", "")
	if err != nil {
		t.Fatalf("GetLogsFiltered failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
	if logs[0].ID != "test-1" {
		t.Errorf("expected ID test-1, got %s", logs[0].ID)
	}
	// Check attributes were stored
	if logs[0].Attributes["env"] != "test" {
		t.Errorf("expected attribute env=test, got %v", logs[0].Attributes["env"])
	}
}

func TestGetAttributeKeys(t *testing.T) {
	store := newTestStore(t)

	entry := makeLogEntry("test-1", "svc-a")
	entry.Attributes = map[string]any{"env": "prod", "region": "eu"}
	store.InsertLog(entry)

	keys, err := store.GetAttributeKeys()
	if err != nil {
		t.Fatalf("GetAttributeKeys failed: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestGetServices(t *testing.T) {
	store := newTestStore(t)

	store.InsertLog(makeLogEntry("test-1", "svc-a"))
	store.InsertLog(makeLogEntry("test-2", "svc-b"))

	services, err := store.GetServices()
	if err != nil {
		t.Fatalf("GetServices failed: %v", err)
	}
	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}
}

func TestCountLogsByMonth(t *testing.T) {
	store := newTestStore(t)

	now := time.Now().UTC()
	entry := makeLogEntry("test-1", "svc-a")
	entry.Timestamp = now
	store.InsertLog(entry)

	count, err := store.CountLogsByMonth(now.Year(), int(now.Month()))
	if err != nil {
		t.Fatalf("CountLogsByMonth failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestConcurrentWrites(t *testing.T) {
	store := newTestStore(t)
	var wg sync.WaitGroup
	errs := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			entry := makeLogEntry(fmt.Sprintf("concurrent-%d", i), "svc-a")
			if err := store.InsertLog(entry); err != nil {
				errs <- fmt.Errorf("insert %d: %w", i, err)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}

	total, err := store.GetTotalLogs()
	if err != nil {
		t.Fatalf("GetTotalLogs failed: %v", err)
	}
	if total != 50 {
		t.Errorf("expected 50 logs, got %d", total)
	}
}

func TestConcurrentReads(t *testing.T) {
	store := newTestStore(t)

	// Insert test data
	for i := 0; i < 10; i++ {
		store.InsertLog(makeLogEntry(fmt.Sprintf("read-%d", i), "svc-a"))
	}

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, count, err := store.GetLogsFiltered(1, 10, "", "", "", "")
			if err != nil {
				errs <- fmt.Errorf("GetLogsFiltered: %w", err)
				return
			}
			if count != 10 {
				errs <- fmt.Errorf("expected 10, got %d", count)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	store := newTestStore(t)
	var wg sync.WaitGroup
	errs := make(chan error, 100)

	// 25 concurrent writers
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			entry := makeLogEntry(fmt.Sprintf("mixed-%d", i), "svc-a")
			if err := store.InsertLog(entry); err != nil {
				errs <- fmt.Errorf("insert %d: %w", i, err)
			}
		}(i)
	}

	// 25 concurrent readers
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := store.GetLogsFiltered(1, 10, "", "", "", "")
			if err != nil {
				errs <- fmt.Errorf("GetLogsFiltered: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
```

- [ ] **Step 2: Run tests**

Run: `cd /Users/madsnylund/Documents/gotail && go test ./db/sqlite/ -v -race`
Expected: all tests pass, no race conditions

- [ ] **Step 3: Commit**

```bash
git add db/sqlite/store.go db/sqlite/testhelper_test.go db/sqlite/store_test.go
git commit -m "test: add store-level and concurrency tests"
```

### Task 9: Write integration tests for HTTP endpoints

**Files:**
- Create: `handlers/integration_test.go`

- [ ] **Step 1: Write endpoint integration tests**

```go
package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gotail/db/sqlite"
	"gotail/handlers/api"
	"gotail/handlers/logging"
	"gotail/models"
)

func newTestStore(t *testing.T) *sqlite.SQLiteStore {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	store, err := sqlite.NewSQLiteStore(dsn)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	migrationSQL, err := os.ReadFile("../migrations/20250625192434_log.sql")
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	sql := extractUpMigration(string(migrationSQL))
	if err := store.ExecRaw(sql); err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	t.Cleanup(func() { store.Close() })
	return store
}

func extractUpMigration(content string) string {
	inUp := false
	inStatement := false
	var result string
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "+goose Up") {
			inUp = true
			continue
		}
		if strings.Contains(line, "+goose Down") {
			break
		}
		if inUp && strings.Contains(line, "+goose StatementBegin") {
			inStatement = true
			continue
		}
		if inUp && strings.Contains(line, "+goose StatementEnd") {
			inStatement = false
			continue
		}
		if inUp && inStatement {
			result += line + "\n"
		}
	}
	return result
}

func postLog(handler http.Handler, entry models.LogEntry) *httptest.ResponseRecorder {
	body, _ := json.Marshal(entry)
	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestPostLog_Valid(t *testing.T) {
	store := newTestStore(t)
	handler := &logging.LogHandler{Store: store}

	svc := "test-service"
	entry := models.LogEntry{
		Timestamp:      time.Now().UTC(),
		SeverityText:   "INFO",
		SeverityNumber: 9,
		Body:           "test message",
		ServiceName:    &svc,
		Attributes:     map[string]any{"env": "test"},
	}

	rr := postLog(http.HandlerFunc(handler.HandleLogInsert), entry)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	total, _ := store.GetTotalLogs()
	if total != 1 {
		t.Errorf("expected 1 log, got %d", total)
	}
}

func TestPostLog_InvalidJSON(t *testing.T) {
	store := newTestStore(t)
	handler := &logging.LogHandler{Store: store}

	req := httptest.NewRequest(http.MethodPost, "/log", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	handler.HandleLogInsert(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestPostLog_BodyTooLarge(t *testing.T) {
	store := newTestStore(t)
	handler := &logging.LogHandler{Store: store}

	entry := models.LogEntry{
		SeverityText:   "INFO",
		SeverityNumber: 9,
		Body:           strings.Repeat("x", 70000), // > 64KB
	}

	rr := postLog(http.HandlerFunc(handler.HandleLogInsert), entry)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPostLog_TooManyAttributes(t *testing.T) {
	store := newTestStore(t)
	handler := &logging.LogHandler{Store: store}

	attrs := make(map[string]any)
	for i := 0; i < 51; i++ {
		attrs[fmt.Sprintf("key-%d", i)] = "val"
	}

	entry := models.LogEntry{
		SeverityText:   "INFO",
		SeverityNumber: 9,
		Body:           "test",
		Attributes:     attrs,
	}

	rr := postLog(http.HandlerFunc(handler.HandleLogInsert), entry)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetLogsAPI(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	// Insert a log
	svc := "svc-a"
	postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
		SeverityText: "ERROR", SeverityNumber: 17, Body: "err", ServiceName: &svc,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/logs?severity=ERROR", nil)
	rr := httptest.NewRecorder()
	apiHandler.HandleLogsAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.LogsAPIResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
}

func TestGetStatsAPI(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	svc := "svc-a"
	postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
		SeverityText: "INFO", SeverityNumber: 9, Body: "msg", ServiceName: &svc,
	})

	now := time.Now()
	url := fmt.Sprintf("/api/stats?year=%d&month=%d", now.Year(), int(now.Month()))
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	apiHandler.HandleStatsAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetAttributeKeysAPI(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
		SeverityText: "INFO", SeverityNumber: 9, Body: "msg",
		Attributes: map[string]any{"env": "prod"},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/attributes", nil)
	rr := httptest.NewRecorder()
	apiHandler.HandleAttributeKeysAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetServicesAPI(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	svc := "my-service"
	postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
		SeverityText: "INFO", SeverityNumber: 9, Body: "msg", ServiceName: &svc,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rr := httptest.NewRecorder()
	apiHandler.HandleServicesAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestConcurrentHTTPReadsAndWrites(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	// 25 concurrent POST /log
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			svc := "svc"
			rr := postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
				SeverityText: "INFO", SeverityNumber: 9,
				Body: fmt.Sprintf("msg-%d", i), ServiceName: &svc,
			})
			if rr.Code != http.StatusNoContent {
				errs <- fmt.Errorf("POST /log %d: got %d", i, rr.Code)
			}
		}(i)
	}

	// 25 concurrent GET /api/logs
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
			rr := httptest.NewRecorder()
			apiHandler.HandleLogsAPI(rr, req)
			if rr.Code != http.StatusOK {
				errs <- fmt.Errorf("GET /api/logs: got %d", rr.Code)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
```

- [ ] **Step 2: Run all tests**

Run: `cd /Users/madsnylund/Documents/gotail && go test ./... -v -race`
Expected: all tests pass

- [ ] **Step 3: Commit**

```bash
git add handlers/integration_test.go
git commit -m "test: add endpoint integration and concurrent HTTP tests"
```
