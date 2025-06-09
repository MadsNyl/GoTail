package db

import (
	"errors"
	"time"
)

type LogEntry struct {
	ID             string            `json:"id"`
	Timestamp      time.Time         `json:"timestamp"`
	SeverityText   string            `json:"severity_text"`
	SeverityNumber int               `json:"severity_number"`
	Body           string            `json:"body"`
	TraceID        string            `json:"trace_id"`
	SpanID         string            `json:"span_id"`
	Attributes     map[string]string `json:"attributes"`
}

type LogStore interface {
	Init() error
	InsertLog(entry LogEntry) error
	Close() error
	GetRecentLogs(limit int) ([]LogEntry, error)
}

var ErrUnsupportedDriver = errors.New("unsupported driver")

func New(driver string, dsn string) (LogStore, error) {
	switch driver {
	case "sqlite":
		return NewSQLiteStore(dsn)
	case "postgres":
		return NewPostgresStore(dsn)
	default:
		return nil, ErrUnsupportedDriver
	}
}

