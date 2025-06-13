package sqlite

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
			id               	TEXT PRIMARY KEY,
			timestamp        	DATETIME NOT NULL,
			observed_ts      	DATETIME,
			trace_id         	TEXT,
			span_id          	TEXT,
			trace_flags      	INTEGER,
			severity_text    	TEXT NOT NULL,
			severity_number  	INT,
			body             	TEXT NOT NULL,
			resource         	TEXT,
			instr_scope      	TEXT,
			service_name     	TEXT,
			service_version  	TEXT,
			service_instance_id TEXT,
			host_name        	TEXT,
			scope_name      	TEXT,
			scope_version   	TEXT,
			created_at       	DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS log_attributes (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			log_id      TEXT NOT NULL,
			key         TEXT NOT NULL,
			value       TEXT NOT NULL,
			FOREIGN KEY (log_id) REFERENCES logs(id)
		);

		CREATE INDEX IF NOT EXISTS idx_attr_key_value ON log_attributes(key, value);
		CREATE INDEX IF NOT EXISTS idx_log_ts ON logs(timestamp);
		CREATE INDEX IF NOT EXISTS idx_log_severity ON logs(severity_text, severity_number);
		CREATE INDEX IF NOT EXISTS idx_log_trace_span ON logs(trace_id, span_id);
		CREATE INDEX IF NOT EXISTS idx_service_name ON logs(service_name);
		CREATE INDEX IF NOT EXISTS idx_host_name ON logs(host_name);
		CREATE INDEX IF NOT EXISTS idx_scope_name ON logs(scope_name);
	`)

	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}