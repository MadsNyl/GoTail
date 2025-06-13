package sqlite

import (
	"database/sql"
	"encoding/json"
	"sort"
	"strings"

	"gotail/models"
)

func (s *SQLiteStore) GetLatestLogs(limit int) ([]models.LogEntry, error) {
	query := `
		SELECT 
			l.id, l.timestamp, l.severity_text, l.severity_number, l.body,
			l.trace_id, l.span_id, l.trace_flags, l.observed_ts,
			l.resource, l.instr_scope, l.created_at,
			l.service_name, l.service_version, l.service_instance_id,
			l.host_name, l.scope_name, l.scope_version,
			a.key, a.value
		FROM (
			SELECT * FROM logs ORDER BY timestamp DESC LIMIT ?
		) l
		LEFT JOIN log_attributes a ON l.id = a.log_id
		ORDER BY l.timestamp DESC`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logMap := make(map[string]*models.LogEntry)

	for rows.Next() {
		var (
			id           string
			entry        models.LogEntry
			resourceJSON sql.NullString
			scopeJSON    sql.NullString
			observedTs   sql.NullTime
			traceFlags   sql.NullInt64
			attrKey      sql.NullString
			attrVal      sql.NullString
		)

		err := rows.Scan(
			&id, &entry.Timestamp, &entry.SeverityText, &entry.SeverityNumber, &entry.Body,
			&entry.TraceID, &entry.SpanID, &traceFlags, &observedTs,
			&resourceJSON, &scopeJSON, &entry.CreatedAt,
			&entry.ServiceName, &entry.ServiceVersion, &entry.ServiceInstanceID,
			&entry.HostName, &entry.ScopeName, &entry.ScopeVersion,
			&attrKey, &attrVal,
		)
		if err != nil {
			return nil, err
		}

		entry.ID = id

		if traceFlags.Valid {
			v := uint8(traceFlags.Int64)
			entry.TraceFlags = &v
		}
		if observedTs.Valid {
			entry.ObservedTimestamp = &observedTs.Time
		}
		if resourceJSON.Valid {
			var res map[string]string
			if err := json.Unmarshal([]byte(resourceJSON.String), &res); err == nil {
				entry.Resource = res
			}
		}
		if scopeJSON.Valid {
			var scope map[string]string
			if err := json.Unmarshal([]byte(scopeJSON.String), &scope); err == nil {
				entry.InstrumentationScope = scope
			}
		}

		if existing, ok := logMap[id]; ok {
			if attrKey.Valid && attrVal.Valid {
				existing.Attributes[attrKey.String] = attrVal.String
			}
		} else {
			entry.Attributes = make(map[string]string)
			if attrKey.Valid && attrVal.Valid {
				entry.Attributes[attrKey.String] = attrVal.String
			}
			logMap[id] = &entry
		}
	}

	logs := make([]models.LogEntry, 0, len(logMap))
	for _, v := range logMap {
		logs = append(logs, *v)
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	return logs, nil
}


func (s *SQLiteStore) GetLogsFiltered(page int, limit int, severity string, attrKey string, attrValue string) ([]models.LogEntry, int, error) {
	offset := (page - 1) * limit

	var (
		whereClauses []string
		args         []interface{}
	)

	query := `
		SELECT DISTINCT l.id, l.timestamp, l.severity_text, l.severity_number, l.body, l.trace_id, l.span_id
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

	query += where + " ORDER BY l.timestamp DESC LIMIT ? OFFSET ?"
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

	var logs []models.LogEntry
	for rows.Next() {
		var entry models.LogEntry
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
