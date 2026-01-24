package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"gotail/middleware"
	"gotail/models"
)

// testStore implements db.LogStore for testing
type testStore struct {
	db *sql.DB
}

func newTestStore(t *testing.T) *testStore {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create tables
	schema := `
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
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return &testStore{db: db}
}

func (s *testStore) Close() error {
	return s.db.Close()
}

func (s *testStore) InsertLog(entry models.LogEntry) error {
	_, err := s.db.Exec(`
		INSERT INTO log (id, timestamp, severity_text, severity_number, body, service_name, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.Timestamp, entry.SeverityText, entry.SeverityNumber, entry.Body, entry.ServiceName, time.Now())
	if err != nil {
		return err
	}

	for k, v := range entry.Attributes {
		_, err := s.db.Exec(`INSERT INTO attribute (log_id, key, value) VALUES (?, ?, ?)`, entry.ID, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *testStore) GetLogsFiltered(page, limit int, severity, attrKey, attrValue, service string) ([]models.LogEntry, int, error) {
	query := `SELECT id, timestamp, severity_text, severity_number, body, service_name FROM log WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM log WHERE 1=1`
	args := []interface{}{}

	if severity != "" {
		query += ` AND severity_text = ?`
		countQuery += ` AND severity_text = ?`
		args = append(args, severity)
	}
	if service != "" {
		query += ` AND service_name = ?`
		countQuery += ` AND service_name = ?`
		args = append(args, service)
	}

	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query += ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var log models.LogEntry
		if err := rows.Scan(&log.ID, &log.Timestamp, &log.SeverityText, &log.SeverityNumber, &log.Body, &log.ServiceName); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

func (s *testStore) GetAttributeKeys() ([]string, error) {
	return []string{}, nil
}

func (s *testStore) GetTotalLogs() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM log`).Scan(&count)
	return count, err
}

func (s *testStore) GetServices() ([]string, error) {
	return []string{}, nil
}

func (s *testStore) CountLogsByMonth(year, month int) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM log
		WHERE strftime('%Y', timestamp) = ? AND strftime('%m', timestamp) = ?`,
		year, month).Scan(&count)
	return count, err
}

func (s *testStore) CountLogsBySeverity(year, month int) (map[string]int, error) {
	rows, err := s.db.Query(`
		SELECT severity_text, COUNT(*) FROM log
		WHERE strftime('%Y', timestamp) = ? AND strftime('%m', timestamp) = ?
		GROUP BY severity_text`, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var sev string
		var count int
		if err := rows.Scan(&sev, &count); err != nil {
			return nil, err
		}
		result[sev] = count
	}
	return result, nil
}

func (s *testStore) CountLogsPerDay(year, month int) (map[int]int, error) {
	return map[int]int{}, nil
}

func (s *testStore) CountLogsByService(year, month int) (map[string]int, error) {
	return map[string]int{}, nil
}

func (s *testStore) CountLogsByAttribute(year, month int) (map[string]int, error) {
	return map[string]int{}, nil
}

// seedTestData adds sample logs to the test store
func seedTestData(t *testing.T, store *testStore) {
	t.Helper()

	logs := []models.LogEntry{
		{ID: "1", Timestamp: time.Now(), SeverityText: "INFO", SeverityNumber: 9, Body: "Info log 1", ServiceName: ptr("api-service")},
		{ID: "2", Timestamp: time.Now(), SeverityText: "ERROR", SeverityNumber: 17, Body: "Error log 1", ServiceName: ptr("api-service")},
		{ID: "3", Timestamp: time.Now(), SeverityText: "INFO", SeverityNumber: 9, Body: "Info log 2", ServiceName: ptr("web-service")},
		{ID: "4", Timestamp: time.Now(), SeverityText: "WARN", SeverityNumber: 13, Body: "Warning log", ServiceName: ptr("api-service")},
		{ID: "5", Timestamp: time.Now(), SeverityText: "ERROR", SeverityNumber: 17, Body: "Error log 2", ServiceName: ptr("web-service")},
	}

	for _, log := range logs {
		if err := store.InsertLog(log); err != nil {
			t.Fatalf("failed to insert test log: %v", err)
		}
	}
}

func ptr(s string) *string {
	return &s
}

// Test Handlers

func TestHandleLogsAPI(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	seedTestData(t, store)

	handler := &APIHandler{Store: store}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		checkResponse  func(*testing.T, *models.LogsAPIResponse)
	}{
		{
			name:           "get all logs default pagination",
			query:          "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *models.LogsAPIResponse) {
				if resp.Total != 5 {
					t.Errorf("expected 5 total logs, got %d", resp.Total)
				}
				if resp.Page != 1 {
					t.Errorf("expected page 1, got %d", resp.Page)
				}
				if resp.Limit != 20 {
					t.Errorf("expected limit 20, got %d", resp.Limit)
				}
			},
		},
		{
			name:           "filter by severity",
			query:          "?severity=ERROR",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *models.LogsAPIResponse) {
				if resp.Total != 2 {
					t.Errorf("expected 2 ERROR logs, got %d", resp.Total)
				}
				for _, log := range resp.Logs {
					if log.SeverityText != "ERROR" {
						t.Errorf("expected ERROR severity, got %s", log.SeverityText)
					}
				}
			},
		},
		{
			name:           "filter by service",
			query:          "?service=api-service",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *models.LogsAPIResponse) {
				if resp.Total != 3 {
					t.Errorf("expected 3 api-service logs, got %d", resp.Total)
				}
			},
		},
		{
			name:           "custom pagination",
			query:          "?page=1&limit=2",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *models.LogsAPIResponse) {
				if len(resp.Logs) != 2 {
					t.Errorf("expected 2 logs on page, got %d", len(resp.Logs))
				}
				if resp.TotalPages != 3 {
					t.Errorf("expected 3 total pages, got %d", resp.TotalPages)
				}
			},
		},
		{
			name:           "method not allowed",
			query:          "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.name == "method not allowed" {
				method = http.MethodPost
			}

			req := httptest.NewRequest(method, "/api/logs"+tt.query, nil)
			rec := httptest.NewRecorder()

			handler.HandleLogsAPI(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.checkResponse != nil && rec.Code == http.StatusOK {
				var resp models.LogsAPIResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				tt.checkResponse(t, &resp)
			}
		})
	}
}

func TestHandleStatsAPI(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	handler := &APIHandler{Store: store}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "get stats for current month",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get stats with year and month",
			query:          "?year=2026&month=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid month",
			query:          "?year=2026&month=13",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid year",
			query:          "?year=1999&month=1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			query:          "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.name == "method not allowed" {
				method = http.MethodPost
			}

			req := httptest.NewRequest(method, "/api/stats"+tt.query, nil)
			rec := httptest.NewRecorder()

			handler.HandleStatsAPI(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if rec.Code == http.StatusOK {
				var resp models.StatsAPIResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Year == 0 || resp.Month == 0 {
					t.Error("response should have year and month set")
				}
			}
		})
	}
}

func TestAPIWithMiddleware(t *testing.T) {
	testKey := "gt_aaaabbbbccccddddeeeeffffgggghhhh"
	os.Setenv("GOTAIL_API_KEYS", testKey)
	middleware.ReloadAPIKeys()
	defer func() {
		os.Unsetenv("GOTAIL_API_KEYS")
		middleware.ReloadAPIKeys()
	}()

	store := newTestStore(t)
	defer store.Close()
	seedTestData(t, store)

	handler := &APIHandler{Store: store}
	protectedHandler := middleware.APIKeyAuth()(http.HandlerFunc(handler.HandleLogsAPI))

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "valid API key",
			apiKey:         testKey,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid API key",
			apiKey:         "gt_invalidkey12345678901234567890ab",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			rec := httptest.NewRecorder()
			protectedHandler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestLogsAPIResponseFormat(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	seedTestData(t, store)

	handler := &APIHandler{Store: store}

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	rec := httptest.NewRecorder()
	handler.HandleLogsAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Check Content-Type header
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Verify JSON structure
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	requiredFields := []string{"logs", "page", "limit", "total", "total_pages"}
	for _, field := range requiredFields {
		if _, ok := resp[field]; !ok {
			t.Errorf("response missing required field: %s", field)
		}
	}
}
