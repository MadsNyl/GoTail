# GoTail – Lightweight Centralized Logging System in Go

GoTail is a self-hosted, centralized logging service written in Go. It supports both **SQLite** and **PostgreSQL** backends via a pluggable database abstraction layer, and exposes an HTTP API to receive logs in structured JSON format.

## 🚀 Features

- Lightweight and dependency-free by default (SQLite)
- PostgreSQL support for scaling
- Normalized schema: logs + context as key-value pairs
- API key protection for log ingestion
- Simple, modular code structure

## 🗂 Project Structure

```
gotail/
├── main.go                 # Entry point
├── .env.example            # Sample env config
├── db/
│   ├── db.go               # Interface and factory
│   ├── sqlite.go           # SQLite implementation
│   └── postgres.go         # PostgreSQL implementation
├── handlers/
│   └── log.go              # HTTP handler for /log
├── middleware/
│   └── auth.go             # API key middleware
```

## ⚙️ Setup

### 1. Clone the repo

```bash
git clone https://github.com/yourusername/gotail.git
cd gotail
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Create a `.env` file

```bash
cp .env.example .env
```

### 4. Run the server

```bash
go run main.go
```

## 🧪 API Usage

### Log Ingestion

```http
POST /log
Headers:
  Content-Type: application/json
  X-API-Key: your-api-key

Body:
{
  "level": "info",
  "message": "User logged in",
  "context": {
    "userId": "123",
    "route": "/login"
  }
}
```

## 🧾 Environment Variables

See `.env.example`:

```env
API_KEY=supersecretkey

# SQLite
DB_DRIVER=sqlite
DB_DSN=logs.db

# PostgreSQL
# DB_DRIVER=postgres
# DB_DSN=postgres://user:pass@localhost:5432/logs?sslmode=disable
```
