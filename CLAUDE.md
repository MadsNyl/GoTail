# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoTail is a lightweight log aggregation and viewing tool written in Go. It provides a REST API for log ingestion and a server-rendered web UI for viewing/filtering logs.

## Build & Development Commands

```bash
# Run directly
go run main.go

# Development with hot-reload (uses templ generate + go build)
air

# Build via Makefile
make run          # Runs go run main.go
make air-build    # Generates templ files and builds binary to ./tmp/main.exe

# Generate Templ templates manually
templ generate

# Docker
docker build -t gotail .
docker run -p 8080:8080 -v $(pwd)/data:/app gotail
```

## Architecture

**Stack:** Go 1.23, Templ (type-safe HTML templating), SQLite, standard net/http

**Key directories:**
- `db/` - Database abstraction with `LogStore` interface (`db/db.go`) and SQLite implementation (`db/sqlite/`)
- `handlers/` - HTTP handlers split by concern: `logging/` for API, `html/` for UI pages
- `models/` - Data structures (`Log`, `LogEntry`, `Attribute`)
- `ui/` - Templ templates and components for server-side rendering
- `middleware/` - Basic auth middleware with SHA256 credential hashing
- `migrations/` - Goose SQL migrations

**Request flow:** HTTP request → middleware (auth) → handler → db layer → SQLite

**Database:** SQLite with two tables: `log` (main entries) and `attribute` (key-value pairs per log). Interface-based design allows adding other backends.

## Environment Variables

Required in `.env`:
- `DB_DRIVER` - Database driver (currently only `sqlite`)
- `DB_DSN` - Database connection string (e.g., `logs.db`)
- `UI_USERNAME` / `UI_PASSWORD` - Basic auth credentials for web UI

## API

**POST /log** - Ingest logs (JSON body with severity, body, service info, attributes)

**GET /** - Web UI for viewing logs (protected by basic auth)

**GET /stats** - Statistics page

## Notes

- Templ files (`.templ`) generate `*_templ.go` files - don't edit generated files directly
- SQLite store uses mutex for concurrent write safety
- Cleanup cron job deletes INFO logs older than 30 days (see `cleanup.sh`)
- Timezone is hardcoded to Europe/Oslo in `db/sqlite/get.go`
