package db

import (
	"errors"
	"gotail/models"

	"gotail/db/sqlite"
)

type LogStore interface {
	Close() error
	InsertLog(entry models.LogEntry) error
	GetLogsFiltered(page int, limit int, severity string, attrKey string, attrValue string, service string) ([]models.LogEntry, int, error)
	GetAttributeKeys() ([]string, error)
	GetTotalLogs() (int, error)
	GetServices() ([]string, error)
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

