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
- Open read connection: DSN + `?_busy_timeout=5000` and execute `PRAGMA query_only=ON` (since `mode=ro` may not be supported by `modernc.org/sqlite`)
- Execute `PRAGMA journal_mode=WAL` on write connection
- Execute `PRAGMA synchronous=NORMAL` (safe with WAL, better write performance)
- Execute `PRAGMA foreign_keys=ON` (SQLite does not enforce by default)
- `writeDB.SetMaxOpenConns(1)`
- Remove broken busy_timeout logic from `main.go` (line 44 checks `dsn == "sqlite"` instead of `driver == "sqlite"` — the timeout is handled inside `NewSQLiteStore` now)

**`Close()` method** must close both `writeDB` and `readDB`.

**Read methods** (`GetLogsFiltered`, `GetAttributeKeys`, `GetTotalLogs`, `GetServices`, all stats methods including `CountLogsByAttribute`):
- Switch from `s.db` to `s.readDB`
- Multi-query reads (`GetLogsFiltered`) wrapped in a read-only transaction for snapshot consistency
- Single-query reads (`GetTotalLogs`, `GetAttributeKeys`, `GetServices`, individual stats methods) do not need transaction wrapping — a single query is already atomic

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
- Auth method (basic auth / API key / either auth)
- Requested path

Also update the `EitherAuth` middleware's own failure path in `api_key_auth.go`.

No credentials are logged. Uses existing `log.Printf` pattern.

### 5. Stats Date Filtering Fix

**Files:** `db/sqlite/stats.go`

Replace `LIKE` pattern matching with proper date range queries:

- `CountLogsByMonth`: `WHERE timestamp >= ? AND timestamp < ?` using first/last day of month
- `CountLogsPerDay`: Same range filter, group by `DATE(timestamp)` instead of `substr()`
- `CountLogsByService` and `CountLogsBySeverity`: Same range-based filtering
- `CountLogsByAttribute`: Same range-based filtering (also uses `LIKE` today)

Date range boundaries computed in Go:
```go
start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
end := start.AddDate(0, 1, 0)
// WHERE timestamp >= start AND timestamp < end
```

Timestamps are stored as RFC 3339 strings which support lexicographic comparison.

Enables proper use of `idx_log_ts` index.

### 6. Endpoint & Concurrency Tests

**Files:** `db/sqlite/store_test.go` (new), `handlers/integration_test.go` (new)

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

**Auth in tests:** Tests inject handlers directly (bypassing auth middleware) for endpoint tests. Concurrency tests may use a test helper that sets up valid API keys.

## Out of Scope

- Host header injection in docs.go (deferred)
- Rate limiting
- Open CORS on documentation endpoints
- N+1 attribute query optimization in `GetLogsFiltered` (future)
- Graceful HTTP server shutdown
