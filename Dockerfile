# --- Builder Stage ---
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install templ and goose CLI
RUN apk add --no-cache git
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Cache and build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

RUN apk add --no-cache sqlite sqlite-libs tzdata ca-certificates busybox
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /go/bin/goose /usr/local/bin/

EXPOSE 8080

# Copy migration and cron files
COPY migrations/ /migrations
COPY cleanup.sh /usr/local/bin/cleanup.sh
COPY cleanup.cron /etc/crontabs/root

RUN chmod +x /usr/local/bin/cleanup.sh

CMD ["sh", "-c", "\
    crond && \
    goose -v up && \
    exec ./main \
"]

