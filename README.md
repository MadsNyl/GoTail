# GoTail – Lightweight Log Collector & Viewer

GoTail is a lightweight, self-hosted logging solution written in Go. It offers a simple API for log ingestion and a clean, server-rendered web UI for viewing logs and statistics. The project is designed to be easy to deploy (via Docker), minimal in resources, and friendly for both developers and non-technical users.

---

## 💡 Why GoTail?

GoTail was built to be a **cheap**, **lightweight**, and **self-hostable** logging app with a clear UI anyone can understand. It’s ideal for small teams, internal tools, or developers who want observability without vendor lock-in.

---

## ✨ Features

- ✅ **Log ingestion via `POST /log`**
- ✅ **Lightweight UI built with [Templ](https://templ.guide)**
- ✅ **Dockerized with SQLite volume support**
- ✅ **Goose-based database migrations**
- ✅ **Basic authentication for both UI and log endpoint**
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

Basic authentication is enforced for:

- The `POST /log` endpoint
- The web UI

Credentials are configured via environment variables (see below).

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

# Basic Auth for UI and /log
UI_USERNAME=admin
UI_PASSWORD=your_secure_password
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