# GoTail – Lightweight Log Collector & Viewer

GoTail is a lightweight, self-hosted logging solution written in Go. It offers a simple API for log ingestion and a clean, server-rendered web UI for viewing logs and statistics. The project is designed to be easy to deploy (via Docker), minimal in resources, and friendly for both developers and non-technical users.

---

## 💡 Why GoTail?

GoTail was built to be a **cheap**, **lightweight**, and **self-hostable** logging app with a clear UI anyone can understand. It’s ideal for small teams, internal tools, or developers who want observability without vendor lock-in.

---

## ✨ Features

- ✅ **Log ingestion via `POST /log`**
- ✅ **JSON API endpoints (`GET /api/logs`, `GET /api/stats`)**
- ✅ **Lightweight UI built with [Templ](https://templ.guide)**
- ✅ **Dockerized with SQLite volume support**
- ✅ **Goose-based database migrations**
- ✅ **Basic authentication for UI**
- ✅ **API key authentication for programmatic access**
- ✅ **Headless mode for API-only deployments**
- ✅ **Built-in cron job for automatic log cleanup**
- ✅ **Easy to fork and customize**
- ❌ **No tracing support (yet)**

---

## 📦 POST /log

Logs are submitted via a POST request to `/log` with the following JSON body:

```json
{
  "timestamp": "2023-01-01T12:00:00Z",
  "severity_text": "INFO",
  "severity_number": 9,
  "body": "Something happened",
  "service_name": "api-service",
  "service_version": "1.0.0",
  "service_instance_id": "abc123",
  "host_name": "my-host",
  "scope_name": "main",
  "scope_version": "1.0.0",
  "attributes": {
    "user_id": 42,
    "env": "prod"
  }
}
```

---

## 📡 JSON API

GoTail provides JSON API endpoints for programmatic access, protected by API key authentication.

### `GET /api/logs`

Query parameters:
- `page` – Page number (default: 1)
- `limit` – Results per page (default: 20, max: 100)
- `severity` – Filter by severity (e.g., `INFO`, `ERROR`)
- `service` – Filter by service name
- `attr_key` / `attr_value` – Filter by attribute

```bash
curl -H "X-API-Key: gt_your_api_key_here_32chars" \
  "http://localhost:8080/api/logs?severity=ERROR&limit=50"
```

### `GET /api/stats`

Query parameters:
- `year` – Year (default: current year)
- `month` – Month 1-12 (default: current month)

```bash
curl -H "X-API-Key: gt_your_api_key_here_32chars" \
  "http://localhost:8080/api/stats?year=2026&month=1"
```

---

## 📄 API Documentation

GoTail includes built-in interactive API docs and an OpenAPI 3.0 spec.

**Interactive docs (no auth required):**
- `GET /docs` – Swagger UI interface
- `GET /openapi.yaml` – Raw OpenAPI spec

**Generate client SDKs:**
```bash
# Go client
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -package gotail openapi.yaml > client.go

# TypeScript client
npx openapi-typescript-codegen --input openapi.yaml --output ./client

# Python client
pip install openapi-python-client
openapi-python-client generate --path openapi.yaml
```

**AI Integration:** The OpenAPI spec is machine-readable - share `openapi.yaml` with AI tools to help implement integrations.

---

## 🖥️ UI

- Rendered server-side using Go Templ
- Protected with basic authentication
- Designed to be intuitive for non-technical users

---

## 🗃️ Database

- Uses **SQLite** with a mounted volume (for persistence)
- Migrations are managed using **Goose**

---

## 🔒 Authentication

GoTail supports two authentication methods:

### Basic Authentication
- Used for the **web UI** (`/` and `/stats`)
- Configured via `UI_USERNAME` and `UI_PASSWORD`

### API Key Authentication
- Used for **JSON API endpoints** (`/api/logs`, `/api/stats`)
- Keys provided via `X-API-Key` header or `Authorization: Bearer` header
- Configured via `GOTAIL_API_KEYS` (comma-separated for multiple keys)

### Log Ingestion (`POST /log`)
- Accepts **both** basic auth and API key authentication
- Allows services to submit logs using API keys while users access the UI with credentials

### API Key Format
Keys must be 35 characters: `gt_` prefix + 32 alphanumeric characters.

Generate a key:
```bash
echo "gt_$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)"
```

### Headless Mode
Set `HEADLESS_MODE=true` to disable the UI entirely and run GoTail as an API-only service. In this mode, `UI_USERNAME` and `UI_PASSWORD` are not required.

---

## 🧹 Cron Job Cleanup

GoTail includes a daily cron job that:

- Runs at **00:00 every day**
- Deletes logs with:
  - `severity_text = 'INFO'`
  - Older than **30 days**

You are encouraged to **fork this repo** and adjust the cron job or retention policy as needed.

---

## ⚙️ Environment Configuration

Sample `.env.example`:

```env
# === SQLite configuration ===
DB_DRIVER=sqlite
DB_DSN=logs.db

# === UI web basic auth ===
UI_USERNAME=admin
UI_PASSWORD=admin

# === API key authentication (comma-separated for multiple keys) ===
GOTAIL_API_KEYS=gt_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6

# === Headless mode (set to "true" to disable UI) ===
HEADLESS_MODE=false

# === Goose migration tool configuration ===
GOOSE_DRIVER=sqlite3
GOOSE_DBSTRING=./logs.db
GOOSE_MIGRATION_DIR=./migrations
```

---

## 🐳 Docker

The project is fully dockerized and mounts an SQLite volume for persistence.

To build and run:

```bash
docker build -t gotail .
docker run -p 8080:8080 -v $(pwd)/data:/app gotail
```

---

## 🛠 Development

Development is powered by [`air`](https://github.com/air-verse/air). To start developing:

```bash
air
```

---

## 🚀 Deployment on Railway

You can deploy GoTail to [Railway](https://railway.app) in just a few steps.

### 🔧 1. Create a new Railway project

- Go to [https://railway.app](https://railway.app)
- Click **New Project** → **Deploy from GitHub Repo**
- Select your forked copy of this repository

### 🛠️ 2. Set environment variables

In the **Environment** tab of your Railway project, configure the following variables:

```env
# Required for the database
DB_DRIVER=sqlite
DB_DSN=/app/logs.db

# Required for Goose migrations
GOOSE_DRIVER=sqlite3
GOOSE_DBSTRING=/app/logs.db
GOOSE_MIGRATION_DIR=./migrations

# Basic Auth for UI
UI_USERNAME=admin
UI_PASSWORD=your_secure_password

# API keys for programmatic access (optional, comma-separated)
GOTAIL_API_KEYS=gt_your_api_key_here_32chars

# Set to "true" for API-only mode without UI (optional)
HEADLESS_MODE=false
```

### 💾 3. Add a persistent volume

Railway supports volumes for SQLite persistence:

- Click the **"+ Create"** button
- Click **"Volume"**
- Mount it at `/app` (so it matches your DB_DSN)

> 🔒 Note: Only one service can mount a volume in Railway. This is why the cron job runs **inside the same container**.

### 🕒 4. Cron job setup

The Docker container includes a built-in cron job that:

- Runs daily at `00:00`
- Deletes all logs with `severity_text = 'INFO'` older than 30 days

You can modify the schedule or script by editing `cleanup.cron` and `cleanup.sh`.