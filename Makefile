run:
	@go run main.go

build-logfactory: ## Build the log factory binary
	@echo "Building log factory..."
	@go build -o bin/logfactory ./cmd/factory

generate-logs: build-logfactory ## Generate logs using the log factory
	@echo "Generating 1000 logs in logs.db..."
	@./bin/logfactory -db=logs.db -count=1000