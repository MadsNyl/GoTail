package db

import (
	"errors"

	"gotail/db/sqlite"
	"gotail/models"
)

type LogStore interface {
	// Initialize the database
	Close() error

	// Insertion
	InsertLog(entry models.LogEntry) error

	// Logs overview
	GetLogsFiltered(page int, limit int, severity string, attrKey string, attrValue string, service string) ([]models.LogEntry, int, error)
	GetAttributeKeys() ([]string, error)
	GetTotalLogs() (int, error)
	GetServices() ([]string, error)

	// Monthly statistics
	CountLogsByMonth(year int, month int) (int, error)
	CountLogsBySeverity(year int, month int) (map[string]int, error)
	CountLogsPerDay(year int, month int) (map[int]int, error)
	CountLogsByService(year int, month int) (map[string]int, error)
	CountLogsByAttribute(year int, month int) (map[string]int, error)
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

