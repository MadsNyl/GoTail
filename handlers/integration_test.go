package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gotail/db/sqlite"
	"gotail/handlers/api"
	"gotail/handlers/logging"
	"gotail/models"
)

func newTestStore(t *testing.T) *sqlite.SQLiteStore {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	store, err := sqlite.NewSQLiteStore(dsn)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	migrationSQL, err := os.ReadFile("../migrations/20250625192434_log.sql")
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	sql := extractUpMigration(string(migrationSQL))
	if err := store.ExecRaw(sql); err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	t.Cleanup(func() { store.Close() })
	return store
}

func extractUpMigration(content string) string {
	inUp := false
	inStatement := false
	var result string
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "+goose Up") {
			inUp = true
			continue
		}
		if strings.Contains(line, "+goose Down") {
			break
		}
		if inUp && strings.Contains(line, "+goose StatementBegin") {
			inStatement = true
			continue
		}
		if inUp && strings.Contains(line, "+goose StatementEnd") {
			inStatement = false
			continue
		}
		if inUp && inStatement {
			result += line + "\n"
		}
	}
	return result
}

func postLog(handler http.Handler, entry models.LogEntry) *httptest.ResponseRecorder {
	body, _ := json.Marshal(entry)
	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestPostLog_Valid(t *testing.T) {
	store := newTestStore(t)
	handler := &logging.LogHandler{Store: store}

	svc := "test-service"
	entry := models.LogEntry{
		Timestamp:      time.Now().UTC(),
		SeverityText:   "INFO",
		SeverityNumber: 9,
		Body:           "test message",
		ServiceName:    &svc,
		Attributes:     map[string]any{"env": "test"},
	}

	rr := postLog(http.HandlerFunc(handler.HandleLogInsert), entry)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	total, _ := store.GetTotalLogs()
	if total != 1 {
		t.Errorf("expected 1 log, got %d", total)
	}
}

func TestPostLog_InvalidJSON(t *testing.T) {
	store := newTestStore(t)
	handler := &logging.LogHandler{Store: store}

	req := httptest.NewRequest(http.MethodPost, "/log", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	handler.HandleLogInsert(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestPostLog_BodyTooLarge(t *testing.T) {
	store := newTestStore(t)
	handler := &logging.LogHandler{Store: store}

	entry := models.LogEntry{
		SeverityText:   "INFO",
		SeverityNumber: 9,
		Body:           strings.Repeat("x", 70000), // > 64KB
	}

	rr := postLog(http.HandlerFunc(handler.HandleLogInsert), entry)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPostLog_TooManyAttributes(t *testing.T) {
	store := newTestStore(t)
	handler := &logging.LogHandler{Store: store}

	attrs := make(map[string]any)
	for i := 0; i < 51; i++ {
		attrs[fmt.Sprintf("key-%d", i)] = "val"
	}

	entry := models.LogEntry{
		SeverityText:   "INFO",
		SeverityNumber: 9,
		Body:           "test",
		Attributes:     attrs,
	}

	rr := postLog(http.HandlerFunc(handler.HandleLogInsert), entry)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetLogsAPI(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	// Insert a log
	svc := "svc-a"
	postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
		SeverityText: "ERROR", SeverityNumber: 17, Body: "err", ServiceName: &svc,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/logs?severity=ERROR", nil)
	rr := httptest.NewRecorder()
	apiHandler.HandleLogsAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.LogsAPIResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
}

func TestGetStatsAPI(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	svc := "svc-a"
	postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
		SeverityText: "INFO", SeverityNumber: 9, Body: "msg", ServiceName: &svc,
	})

	now := time.Now()
	url := fmt.Sprintf("/api/stats?year=%d&month=%d", now.Year(), int(now.Month()))
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	apiHandler.HandleStatsAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetAttributeKeysAPI(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
		SeverityText: "INFO", SeverityNumber: 9, Body: "msg",
		Attributes: map[string]any{"env": "prod"},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/attributes", nil)
	rr := httptest.NewRecorder()
	apiHandler.HandleAttributeKeysAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetServicesAPI(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	svc := "my-service"
	postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
		SeverityText: "INFO", SeverityNumber: 9, Body: "msg", ServiceName: &svc,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rr := httptest.NewRecorder()
	apiHandler.HandleServicesAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestConcurrentHTTPReadsAndWrites(t *testing.T) {
	store := newTestStore(t)
	logHandler := &logging.LogHandler{Store: store}
	apiHandler := &api.APIHandler{Store: store}

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	// 25 concurrent POST /log
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			svc := "svc"
			rr := postLog(http.HandlerFunc(logHandler.HandleLogInsert), models.LogEntry{
				SeverityText: "INFO", SeverityNumber: 9,
				Body: fmt.Sprintf("msg-%d", i), ServiceName: &svc,
			})
			if rr.Code != http.StatusNoContent {
				errs <- fmt.Errorf("POST /log %d: got %d", i, rr.Code)
			}
		}(i)
	}

	// 25 concurrent GET /api/logs
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
			rr := httptest.NewRecorder()
			apiHandler.HandleLogsAPI(rr, req)
			if rr.Code != http.StatusOK {
				errs <- fmt.Errorf("GET /api/logs: got %d", rr.Code)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
