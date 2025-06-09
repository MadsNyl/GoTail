package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		ts DATETIME NOT NULL,
		severity_text TEXT NOT NULL,
		severity_number INT,
		body TEXT NOT NULL,
		trace_id TEXT,
		span_id TEXT
		);

		CREATE TABLE IF NOT EXISTS log_attributes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		log_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		FOREIGN KEY (log_id) REFERENCES logs(id)
		);

		CREATE INDEX IF NOT EXISTS idx_attr_key_value ON log_attributes(key, value);
	`)

	return err
}

func (s *SQLiteStore) InsertLog(entry LogEntry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO logs (id, ts, severity_text, severity_number, body, trace_id, span_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.Timestamp, entry.SeverityText, entry.SeverityNumber,
		entry.Body, entry.TraceID, entry.SpanID)
	if err != nil {
		tx.Rollback()
		return err
	}
	stmt, _ := tx.Prepare(`INSERT INTO log_attributes (log_id, key, value) VALUES (?, ?, ?)`)
	for k, v := range entry.Attributes {
		_, _ = stmt.Exec(entry.ID, k, v)
	}
	stmt.Close()
	return tx.Commit()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) GetRecentLogs(limit int) ([]LogEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, ts, severity_text, severity_number, body, trace_id, span_id
		FROM logs
		ORDER BY ts DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		if err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.SeverityText,
			&entry.SeverityNumber, &entry.Body, &entry.TraceID, &entry.SpanID); err != nil {
			return nil, err
		}

		// load attributes
		attrRows, err := s.db.Query(`SELECT key, value FROM log_attributes WHERE log_id = ?`, entry.ID)
		if err != nil {
			return nil, err
		}
		defer attrRows.Close()

		entry.Attributes = map[string]string{}
		for attrRows.Next() {
			var key, value string
			_ = attrRows.Scan(&key, &value)
			entry.Attributes[key] = value
		}

		logs = append(logs, entry)
	}
	return logs, nil
}

