package sqlite

import (
	"strings"

	"gotail/models"
)

func (s *SQLiteStore) GetLogsFiltered(
	page int,
	limit int,
	severity string,
	attrKey string,
	attrValue string,
) ([]models.LogEntry, int, error) {
	offset := (page - 1) * limit

	var (
		whereClauses []string
		args         []interface{}
	)

	query := `
		SELECT l.*
		FROM log l`

	// Add join if filtering on attribute
	if attrKey != "" && attrValue != "" {
		query += `
			INNER JOIN attribute a ON a.log_id = l.id`
		whereClauses = append(whereClauses, "a.key = ? AND a.value LIKE ?")
		args = append(args, attrKey, "%"+attrValue+"%")
	}

	if severity != "" {
		whereClauses = append(whereClauses, "l.severity_text = ?")
		args = append(args, severity)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY l.timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Count query
	countQuery := `
		SELECT COUNT(DISTINCT l.id)
		FROM log l`
	if attrKey != "" && attrValue != "" {
		countQuery += `
			INNER JOIN attribute a ON a.log_id = l.id`
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

	var logs []models.LogEntry
	for rows.Next() {
		var entry models.LogEntry

		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.SeverityText,
			&entry.SeverityNumber,
			&entry.Body,
			&entry.ServiceName,
			&entry.ServiceVersion,
			&entry.ServiceInstanceID,
			&entry.HostName,
			&entry.ScopeName,
			&entry.ScopeVersion,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		attrRows, err := s.db.Query(`SELECT key, value FROM attribute WHERE log_id = ?`, entry.ID)
		if err != nil {
			return nil, 0, err
		}

		entry.Attributes = make(map[string]any)
		for attrRows.Next() {
			var k string
			var v any
			if err := attrRows.Scan(&k, &v); err != nil {
				attrRows.Close()
				return nil, 0, err
			}
			entry.Attributes[k] = v
		}
		attrRows.Close()

		logs = append(logs, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}


func (s *SQLiteStore) GetAttributeKeys() ([]string, error) {
    // Example implementation, adjust according to your schema
    rows, err := s.db.Query("SELECT DISTINCT key FROM attribute")
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

func (s *SQLiteStore) GetTotalLogs() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM log").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}