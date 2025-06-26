package sqlite

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
	mutex sync.Mutex
}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}