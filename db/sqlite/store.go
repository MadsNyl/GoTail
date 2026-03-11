package sqlite

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	writeDB *sql.DB
	readDB  *sql.DB
	mu      sync.Mutex
}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	writeDSN := dsn + "?_busy_timeout=5000"
	readDSN := dsn + "?_busy_timeout=5000"

	writeDB, err := sql.Open("sqlite", writeDSN)
	if err != nil {
		return nil, fmt.Errorf("open write db: %w", err)
	}
	writeDB.SetMaxOpenConns(1)

	// Enable WAL mode
	if _, err := writeDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := writeDB.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("set synchronous: %w", err)
	}
	if _, err := writeDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	readDB, err := sql.Open("sqlite", readDSN)
	if err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("open read db: %w", err)
	}
	readDB.SetMaxOpenConns(8)

	// Make read connection read-only
	if _, err := readDB.Exec("PRAGMA query_only=ON"); err != nil {
		writeDB.Close()
		readDB.Close()
		return nil, fmt.Errorf("set query_only: %w", err)
	}

	return &SQLiteStore{writeDB: writeDB, readDB: readDB}, nil
}

func (s *SQLiteStore) Close() error {
	writeErr := s.writeDB.Close()
	readErr := s.readDB.Close()
	if writeErr != nil {
		return writeErr
	}
	return readErr
}

func (s *SQLiteStore) ExecRaw(sql string) error {
	_, err := s.writeDB.Exec(sql)
	return err
}
