package db

import (
	"database/sql"
	"strings"

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

func (s *SQLiteStore) GetLogsFiltered(page int, limit int, severity string, attrKey string, attrValue string) ([]LogEntry, int, error) {
	offset := (page - 1) * limit

	var (
		whereClauses []string
		args         []interface{}
	)

	query := `
		SELECT DISTINCT l.id, l.ts, l.severity_text, l.severity_number, l.body, l.trace_id, l.span_id
		FROM logs l`

	// Add join if filtering on attribute
	if attrKey != "" && attrValue != "" {
		query += `
			INNER JOIN log_attributes a ON a.log_id = l.id`
		whereClauses = append(whereClauses, "a.key = ? AND a.value LIKE ?")
		args = append(args, attrKey, "%"+attrValue+"%")
	}

	if severity != "" {
		whereClauses = append(whereClauses, "l.severity_text = ?")
		args = append(args, severity)
	}

	where := ""
	if len(whereClauses) > 0 {
		where = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += where + " ORDER BY l.ts DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Count query
	countQuery := `SELECT COUNT(DISTINCT l.id) FROM logs l`
	if attrKey != "" && attrValue != "" {
		countQuery += ` INNER JOIN log_attributes a ON a.log_id = l.id`
	}
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var count int
	countArgs := args[:len(args)-2] // exclude limit and offset
	if err := s.db.QueryRow(countQuery, countArgs...).Scan(&count); err != nil {
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
		var traceID, spanID sql.NullString
		
		// Scan into sql.NullString for nullable columns
		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.SeverityText,
			&entry.SeverityNumber, &entry.Body, &traceID, &spanID)
		if err != nil {
			return nil, 0, err
		}

		// Convert sql.NullString to *string
		if traceID.Valid {
			entry.TraceID = traceID.String
		}
		if spanID.Valid {
			entry.SpanID = spanID.String
		}

		attrRows, err := s.db.Query(`SELECT key, value FROM log_attributes WHERE log_id = ?`, entry.ID)
		if err != nil {
			return nil, 0, err
		}
		defer attrRows.Close()

		entry.Attributes = map[string]string{}
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


func (s *SQLiteStore) GetAttributeKeys() ([]string, error) {
    // Example implementation, adjust according to your schema
    rows, err := s.db.Query("SELECT DISTINCT key FROM log_attributes")
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
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return keys, nil
}