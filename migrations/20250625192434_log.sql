-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS log (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    severity_text TEXT NOT NULL,
    severity_number INTEGER NOT NULL,
    body TEXT NOT NULL,
    service_name TEXT,
    service_version TEXT,
    service_instance_id TEXT,
    host_name TEXT,
    scope_name TEXT,
    scope_version TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attribute (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (log_id) REFERENCES log(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_attr_key_value ON attribute(key, value);
CREATE INDEX IF NOT EXISTS idx_log_ts ON log(timestamp);
CREATE INDEX IF NOT EXISTS idx_log_severity ON log(severity_text, severity_number);
CREATE INDEX IF NOT EXISTS idx_service_name ON log(service_name);
CREATE INDEX IF NOT EXISTS idx_host_name ON log(host_name);
CREATE INDEX IF NOT EXISTS idx_scope_name ON log(scope_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_attr_key_value;
DROP INDEX IF EXISTS idx_log_ts;
DROP INDEX IF EXISTS idx_log_severity;
DROP INDEX IF EXISTS idx_service_name;
DROP INDEX IF EXISTS idx_host_name;
DROP INDEX IF EXISTS idx_scope_name;
DROP TABLE IF EXISTS attribute;
DROP TABLE IF EXISTS log;
-- +goose StatementEnd
