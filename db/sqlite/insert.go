package sqlite

import (
	"gotail/models"
)

func (s *SQLiteStore) InsertLog(entry models.LogEntry) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Insert into log table
	_, err = tx.Exec(`
        INSERT INTO log (
            id, timestamp, severity_text, severity_number, body,
            service_name, service_version, service_instance_id,
            host_name, scope_name, scope_version
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
		entry.ID,
		entry.Timestamp,
		entry.SeverityText,
		entry.SeverityNumber,
		entry.Body,
		entry.ServiceName,
		entry.ServiceVersion,
		entry.ServiceInstanceID,
		entry.HostName,
		entry.ScopeName,
		entry.ScopeVersion,
	)
	if err != nil {
		return err
	}

	// Insert attributes
	for k, v := range entry.Attributes {
		_, err := tx.Exec(`
            INSERT INTO attribute (log_id, key, value) VALUES (?, ?, ?)
        `, entry.ID, k, v)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
