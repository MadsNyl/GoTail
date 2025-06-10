package db

import (
	"database/sql"
	"fmt"
	"strings"

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

func (s *PostgresStore) GetLogsFiltered(
	page int, limit int, severity string, attrKey string, attrValue string,
) ([]LogEntry, int, error) {
	offset := (page - 1) * limit

	var (
		whereClauses []string
		args         []interface{}
		argIndex     = 1
	)

	query := `
		SELECT l.id, l.ts, l.severity_text, l.severity_number, l.body, l.trace_id, l.span_id
		FROM logs l`

	// Optional JOIN if filtering by attribute
	if attrKey != "" && attrValue != "" {
		query += `
			JOIN log_attributes a ON a.log_id = l.id`
		whereClauses = append(whereClauses, fmt.Sprintf("a.key = $%d AND a.value ILIKE $%d", argIndex, argIndex+1))
		args = append(args, attrKey, "%"+attrValue+"%")
		argIndex += 2
	}

	if severity != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("l.severity_text = $%d", argIndex))
		args = append(args, severity)
		argIndex++
	}

	// Combine WHERE clause
	where := ""
	if len(whereClauses) > 0 {
		where = " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += where + fmt.Sprintf(" ORDER BY l.ts DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Total count query
	countQuery := `SELECT COUNT(DISTINCT l.id) FROM logs l`
	if attrKey != "" && attrValue != "" {
		countQuery += ` JOIN log_attributes a ON a.log_id = l.id`
	}
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var count int
	if err := s.db.QueryRow(countQuery, args[:argIndex-1]...).Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.SeverityText,
			&entry.SeverityNumber, &entry.Body, &entry.TraceID, &entry.SpanID)
		if err != nil {
			return nil, 0, err
		}

		attrRows, err := s.db.Query(`SELECT key, value FROM log_attributes WHERE log_id = $1`, entry.ID)
		if err != nil {
			return nil, 0, err
		}
		defer attrRows.Close()

		entry.Attributes = make(map[string]string)
		for attrRows.Next() {
			var k, v string
			if err := attrRows.Scan(&k, &v); err != nil {
				return nil, 0, err
			}
			entry.Attributes[k] = v
		}
		logs = append(logs, entry)
	}

	return logs, count, nil
}

func (s *PostgresStore) GetAttributeKeys() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT key FROM log_attributes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}