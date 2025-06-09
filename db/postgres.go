package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		ts TIMESTAMPTZ NOT NULL,
		severity_text TEXT NOT NULL,
		severity_number INT,
		body TEXT NOT NULL,
		trace_id TEXT,
		span_id TEXT
		);

		CREATE TABLE IF NOT EXISTS log_attributes (
		id SERIAL PRIMARY KEY,
		log_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		FOREIGN KEY (log_id) REFERENCES logs(id)
		);

		CREATE INDEX IF NOT EXISTS idx_attr_key_value ON log_attributes(key, value);
	`)
	return err
}

func (s *PostgresStore) InsertLog(entry LogEntry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO logs (id, ts, severity_text, severity_number, body, trace_id, span_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.ID, entry.Timestamp, entry.SeverityText, entry.SeverityNumber,
		entry.Body, entry.TraceID, entry.SpanID)
	if err != nil {
		tx.Rollback()
		return err
	}
	stmt, _ := tx.Prepare(`INSERT INTO log_attributes (log_id, key, value) VALUES ($1, $2, $3)`)
	for k, v := range entry.Attributes {
		_, _ = stmt.Exec(entry.ID, k, v)
	}
	stmt.Close()
	return tx.Commit()
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) GetRecentLogs(limit int) ([]LogEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, ts, severity_text, severity_number, body, trace_id, span_id
		FROM logs
		ORDER BY ts DESC
		LIMIT $1`, limit)
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

		// Load attributes
		attrRows, err := s.db.Query(`SELECT key, value FROM log_attributes WHERE log_id = $1`, entry.ID)
		if err != nil {
			return nil, err
		}
		defer attrRows.Close()

		entry.Attributes = map[string]string{}
		for attrRows.Next() {
			var key, value string
			if err := attrRows.Scan(&key, &value); err != nil {
				return nil, err
			}
			entry.Attributes[key] = value
		}

		logs = append(logs, entry)
	}

	return logs, nil
}
