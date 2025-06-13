package db

import (
	"errors"
	"gotail/models"

	"gotail/db/sqlite"
)

type LogStore interface {
	Init() error
	InsertLog(entry models.LogEntry) error
	Close() error
	GetLogsFiltered(page int, limit int, severity string, attrKey string, attrValue string) ([]models.LogEntry, int, error)
	GetAttributeKeys() ([]string, error)
	GetLatestLogs(limit int) ([]models.LogEntry, error)
}

var ErrUnsupportedDriver = errors.New("unsupported driver")

func New(driver string, dsn string) (LogStore, error) {
	switch driver {
	case "sqlite":
		return sqlite.NewSQLiteStore(dsn)
	default:
		return nil, ErrUnsupportedDriver
	}
}

