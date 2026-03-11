package sqlite

import (
	"time"
)

func dateRange(year int, month int) (string, string) {
	loc, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		loc = time.FixedZone("CET", 1*60*60)
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)
	return start.Format(time.RFC3339), end.Format(time.RFC3339)
}

func (s *SQLiteStore) CountLogsByMonth(year int, month int) (int, error) {
	start, end := dateRange(year, month)
	var count int
	err := s.readDB.QueryRow(
		"SELECT COUNT(*) FROM log WHERE timestamp >= ? AND timestamp < ?",
		start, end,
	).Scan(&count)
	return count, err
}

func (s *SQLiteStore) CountLogsBySeverity(year int, month int) (map[string]int, error) {
	start, end := dateRange(year, month)
	rows, err := s.readDB.Query(`
		SELECT severity_text, COUNT(*)
		FROM log
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY severity_text`, start, end)
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
	start, end := dateRange(year, month)
	rows, err := s.readDB.Query(`
		SELECT CAST(strftime('%d', timestamp) AS INTEGER) AS day, COUNT(*)
		FROM log
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY day
		ORDER BY day`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var day int
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		result[day] = count
	}
	return result, rows.Err()
}

func (s *SQLiteStore) CountLogsByService(year int, month int) (map[string]int, error) {
	start, end := dateRange(year, month)
	rows, err := s.readDB.Query(`
		SELECT service_name, COUNT(*)
		FROM log
		WHERE timestamp >= ? AND timestamp < ? AND service_name IS NOT NULL
		GROUP BY service_name`, start, end)
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
	start, end := dateRange(year, month)
	rows, err := s.readDB.Query(`
		SELECT key, COUNT(DISTINCT log_id)
		FROM attribute
		WHERE log_id IN (
			SELECT id FROM log WHERE timestamp >= ? AND timestamp < ?
		)
		GROUP BY key`, start, end)
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
