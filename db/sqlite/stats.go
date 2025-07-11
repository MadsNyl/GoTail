package sqlite

import (
	"fmt"
	"strconv"
)

func (s *SQLiteStore) CountLogsByMonth(year int, month int) (int, error) {
	datePrefix := fmt.Sprintf("%04d-%02d", year, month)
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM log WHERE strftime('%Y-%m', timestamp) = ?", datePrefix).Scan(&count)
	return count, err
}

func (s *SQLiteStore) CountLogsBySeverity(year int, month int) (map[string]int, error) {
    datePrefix := fmt.Sprintf("%04d-%02d", year, month)
    rows, err := s.db.Query("SELECT severity_text, COUNT(*) FROM log WHERE strftime('%Y-%m', timestamp) = ? GROUP BY severity_text", datePrefix)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make(map[string]int)
    for rows.Next() {
        var severity string
        var count int
        if err := rows.Scan(&severity, &count); err != nil {
            return nil, err
        }
        result[severity] = count
    }

    return result, rows.Err()
}

func (s *SQLiteStore) CountLogsPerDay(year int, month int) (map[int]int, error) {
    datePrefix := fmt.Sprintf("%04d-%02d", year, month)
    rows, err := s.db.Query("SELECT strftime('%d', timestamp) AS day, COUNT(*) FROM log WHERE strftime('%Y-%m', timestamp) = ? GROUP BY day ORDER BY day", datePrefix)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make(map[int]int)
    for rows.Next() {
        var dayStr string
        var count int
        if err := rows.Scan(&dayStr, &count); err != nil {
            return nil, err
        }
        day, _ := strconv.Atoi(dayStr)
        result[day] = count
    }

    return result, rows.Err()
}

func (s *SQLiteStore) CountLogsByService(year int, month int) (map[string]int, error) {
    datePrefix := fmt.Sprintf("%04d-%02d", year, month)
    rows, err := s.db.Query("SELECT service_name, COUNT(*) FROM log WHERE strftime('%Y-%m', timestamp) = ? AND service_name IS NOT NULL GROUP BY service_name", datePrefix)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make(map[string]int)
    for rows.Next() {
        var service string
        var count int
        if err := rows.Scan(&service, &count); err != nil {
            return nil, err
        }
        result[service] = count
    }

    return result, rows.Err()
}

func (s *SQLiteStore) CountLogsByAttribute(year int, month int) (map[string]int, error) {
    datePrefix := fmt.Sprintf("%04d-%02d", year, month)
    query := `
        SELECT key, COUNT(DISTINCT log_id) FROM attribute
        WHERE log_id IN (
            SELECT id FROM log WHERE strftime('%Y-%m', timestamp) = ?
        )
        GROUP BY key;
    `
    rows, err := s.db.Query(query, datePrefix)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make(map[string]int)
    for rows.Next() {
        var key string
        var count int
        if err := rows.Scan(&key, &count); err != nil {
            return nil, err
        }
        result[key] = count
    }

    return result, rows.Err()
}
