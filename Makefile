# === Config ===
APP_NAME=gotail
API_IMAGE=gotail-api

# SQLite (named Docker volume, not stored on host)
SQLITE_VOLUME=gotail-sqlite-data
SQLITE_DB_PATH=logs.db

# PostgreSQL
POSTGRES_VOLUME=gotail-pgdata:/var/lib/postgresql/data
POSTGRES_CONTAINER=gotail-postgres
POSTGRES_DB=logs
POSTGRES_USER=gotail
POSTGRES_PASSWORD=secret

# === Targets ===

# Build the Go API Docker image
build:
	docker build -t $(API_IMAGE) .

# Build the Go API locally
go:
	go build -o gotail .

# Run the API with SQLite (no local file, volume is managed by Docker)
run-sqlite:
	docker run --rm -p 8080:8080 \
		-v $(SQLITE_VOLUME):/app/data \
		--env DB_DRIVER=sqlite \
		--env DB_DSN=/app/data/logs.db \
		--env-file .env \
		--name $(APP_NAME)-sqlite \
		$(API_IMAGE)

# Create and run PostgreSQL container with persistent volume
postgres:
	docker run --rm -d \
		--name $(POSTGRES_CONTAINER) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-v $(POSTGRES_VOLUME) \
		-p 5432:5432 \
		postgres:15

# Stop PostgreSQL container
postgres-stop:
	docker stop $(POSTGRES_CONTAINER)

# Remove named SQLite volume (DANGER: deletes all logs)
clean-sqlite:
	docker volume rm $(SQLITE_VOLUME)

# Remove PostgreSQL volume (DANGER: deletes all logs)
clean-postgres:
	docker volume rm gotail-pgdata

# View logs from the API container
logs:
	docker logs -f $(APP_NAME)-sqlite

run:
	docker run --rm -p 8080:8080 \
		-v $(SQLITE_VOLUME):/app/data \
		--env DB_DRIVER=sqlite \
		--env DB_DSN=/app/data/logs.db \
		--env-file .env \
		--name $(APP_NAME)-sqlite \
		$(API_IMAGE)

build-logfactory: ## Build the log factory binary
	@echo "Building log factory..."
	@go build -o bin/logfactory ./cmd/factory

generate-logs: build-logfactory ## Generate mock logs (use LOG_COUNT and DB_PATH env vars)
	@echo "Generating 1000 logs in $(SQLITE_DB_PATH)..."
	@./bin/logfactory -db=$(SQLITE_DB_PATH) -count=1000