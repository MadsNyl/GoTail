# Concurrency Fix & Security Hardening — Design Spec

**Date:** 2026-03-11
**Issue:** #9 — Conflict between parallel endpoint requests
**Branch:** 9-conflict-between-parallell-endpoint-requests

## Problem

SQLite fails with `cannot start a transaction within a transaction` when concurrent read and write requests hit the database. The root cause is Go's `database/sql` connection pooling sharing a single connection between a write transaction (InsertLog) and concurrent read queries. After this error, no further insertions work.

Additional vulnerabilities identified during investigation: missing input validation, no panic recovery, no auth failure logging, and incorrect date filtering in stats queries.

## Changes

### 1. WAL Mode + Dual Connection Pools

**Files:** `db/sqlite/store.go`, `db/sqlite/get.go`, `db/sqlite/insert.go`, `db/sqlite/stats.go`, `main.go`

The `SQLiteStore` gets two separate connection pools:

```go
type SQLiteStore struct {
    writeDB *sql.DB    // MaxOpenConns(1), mutex-protected
    readDB  *sql.DB    // read-only connection pool
    mu      sync.Mutex // protects writeDB only
}
```

**Initialization:**
- Open write connection: DSN + `?_busy_timeout=5000`
- Open read connection: DSN + `?mode=ro&_busy_timeout=5000`
- Execute `PRAGMA journal_mode=WAL` on write connection
- `writeDB.SetMaxOpenConns(1)`

**Read methods** (`GetLogsFiltered`, `GetAttributeKeys`, `GetTotalLogs`, `GetServices`, all stats methods):
- Switch from `s.db` to `s.readDB`
- Wrap queries in a read-only transaction (`s.readDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})`) for snapshot consistency
- `GetLogsFiltered` benefits most: count query, data query, and per-log attribute queries all see the same snapshot

**Write methods** (`InsertLog`):
- Continue using `s.writeDB` with mutex protection
- No change to transaction logic

### 2. Input Validation

**Files:** `middleware/max_body_size.go` (new), `handlers/logging/insert.go`

**Request body size limit:**
- New `MaxBodySize` middleware using `http.MaxBytesReader` with 1MB limit
- Applied to POST /log route only

**Field validation in insert handler (after JSON parsing):**
- `Body`: max 64KB (65,536 bytes)
- `SeverityText`, `ServiceName`, `HostName`, `ScopeName`: max 256 chars
- `Attributes`: max 50 entries per log, key max 256 chars, value max 4KB (4,096 bytes)
- Returns 400 with JSON error describing which field exceeded limits

### 3. Panic Recovery Middleware

**Files:** `middleware/recovery.go` (new), `main.go`

- Wraps handler in deferred `recover()`
- On panic: logs stack trace to stderr, returns HTTP 500 with JSON `{"error": "internal server error"}`
- Applied as outermost middleware on all routes in `main.go`

### 4. Failed Auth Attempt Logging

**Files:** `middleware/basic_auth.go`, `middleware/api_key_auth.go`

Log failed authentication attempts to stderr with:
- Timestamp (via `log.Printf`)
- Client IP (`r.RemoteAddr`)
- Auth method (basic auth / API key)
- Requested path

No credentials are logged. Uses existing `log.Printf` pattern.

### 5. Stats Date Filtering Fix

**Files:** `db/sqlite/stats.go`

Replace `LIKE` pattern matching with proper date range queries:

- `CountLogsByMonth`: `WHERE timestamp >= ? AND timestamp < ?` using first/last day of month
- `CountLogsPerDay`: Same range filter, group by `DATE(timestamp)` instead of `substr()`
- `CountLogsByService` and `CountLogsBySeverity`: Same range-based filtering

Enables proper use of `idx_log_ts` index.

### 6. Endpoint & Concurrency Tests

**Files:** New test files

**Test database setup:**
- Each test creates a temporary SQLite DB via `t.TempDir()`
- Goose migrations applied programmatically
- DB cleaned up automatically after test

**Endpoint tests:**
- POST /log — valid input, invalid input, oversized body, field validation
- GET /api/logs — with filters, pagination
- GET /api/stats — with year/month params
- GET /api/attributes — returns attribute keys
- GET /api/services — returns service list

**Concurrency tests using `sync.WaitGroup`:**
- Concurrent writes (multiple POST /log)
- Concurrent reads (multiple GET /api/logs)
- Mixed concurrent reads and writes
- Assert: no errors, data consistency, no deadlocks

## Out of Scope

- Host header injection in docs.go (deferred)
- Rate limiting
- Open CORS on documentation endpoints
