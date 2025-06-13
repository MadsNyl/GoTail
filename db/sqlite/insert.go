package sqlite

import (
	"encoding/json"
	"gotail/models"
)

func (s *SQLiteStore) InsertLog(entry models.LogEntry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	var (
		observedTs   interface{}
		traceFlags   interface{}
		resourceJson interface{}
		scopeJson    interface{}
	)

	if entry.ObservedTimestamp != nil {
		observedTs = entry.ObservedTimestamp.UTC()
	}
	if entry.TraceFlags != nil {
		traceFlags = *entry.TraceFlags
	}

	// Extract flat resource fields
	var serviceName, serviceVersion, serviceInstanceID, hostName string
	if entry.Resource != nil {
		if b, err := json.Marshal(entry.Resource); err == nil {
			resourceJson = string(b)
		} else {
			tx.Rollback()
			return err
		}
		serviceName = entry.Resource["service.name"]
		serviceVersion = entry.Resource["service.version"]
		serviceInstanceID = entry.Resource["service.instance.id"]
		hostName = entry.Resource["host.name"]
	}

	// Extract flat scope fields
	var scopeName, scopeVersion string
	if entry.InstrumentationScope != nil {
		if b, err := json.Marshal(entry.InstrumentationScope); err == nil {
			scopeJson = string(b)
		} else {
			tx.Rollback()
			return err
		}
		scopeName = entry.InstrumentationScope["name"]
		scopeVersion = entry.InstrumentationScope["version"]
	}

	_, err = tx.Exec(`
		INSERT INTO logs (
			id, timestamp, observed_ts, trace_id, span_id, trace_flags,
			severity_text, severity_number, body,
			resource, instr_scope,
			service_name, service_version, service_instance_id, host_name,
			scope_name, scope_version
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID,
		entry.Timestamp.UTC(),
		observedTs,
		entry.TraceID,
		entry.SpanID,
		traceFlags,
		entry.SeverityText,
		entry.SeverityNumber,
		entry.Body,
		resourceJson,
		scopeJson,
		serviceName,
		serviceVersion,
		serviceInstanceID,
		hostName,
		scopeName,
		scopeVersion,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO log_attributes (log_id, key, value) VALUES (?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for k, v := range entry.Attributes {
		if _, err := stmt.Exec(entry.ID, k, v); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
