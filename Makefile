run:
	@go run main.go

air-build:
	@templ generate
	@go build -o ./tmp/main.exe .