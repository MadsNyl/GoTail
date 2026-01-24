setup:
	@echo "Installing development tools..."
	brew install goose
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest
	@echo "Setup complete! Run 'make db-init' to initialize the database."

run:
	@go run main.go

air-build:
	@templ generate
	@go build -o ./tmp/main.exe .

test:
	@go test ./...

test-v:
	@go test ./... -v

test-cover:
	@go test ./... -cover

test-middleware:
	@go test ./middleware -v

test-api:
	@go test ./handlers/api -v

db-init:
	@goose -dir migrations up

db-down:
	@goose -dir migrations down

db-status:
	@goose -dir migrations status

db-delete:
	@rm -f ./logs.db
	@echo "Database deleted"

db-reset: db-delete db-init
	@echo "Database reset complete"