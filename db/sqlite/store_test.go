package sqlite_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"gotail/models"
)

func makeLogEntry(id string, svc string) models.LogEntry {
	s := svc
	return models.LogEntry{
		ID:             id,
		Timestamp:      time.Now().UTC(),
		SeverityText:   "INFO",
		SeverityNumber: 9,
		Body:           "test log body",
		ServiceName:    &s,
		Attributes:     map[string]any{"env": "test"},
	}
}

func TestInsertAndGetLogs(t *testing.T) {
	store := newTestStore(t)

	entry := makeLogEntry("test-1", "svc-a")
	if err := store.InsertLog(entry); err != nil {
		t.Fatalf("InsertLog failed: %v", err)
	}

	logs, count, err := store.GetLogsFiltered(1, 10, "", "", "", "")
	if err != nil {
		t.Fatalf("GetLogsFiltered failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
	if logs[0].ID != "test-1" {
		t.Errorf("expected ID test-1, got %s", logs[0].ID)
	}
}

func TestGetAttributeKeys(t *testing.T) {
	store := newTestStore(t)

	entry := makeLogEntry("test-1", "svc-a")
	entry.Attributes = map[string]any{"env": "prod", "region": "eu"}
	store.InsertLog(entry)

	keys, err := store.GetAttributeKeys()
	if err != nil {
		t.Fatalf("GetAttributeKeys failed: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestGetServices(t *testing.T) {
	store := newTestStore(t)

	store.InsertLog(makeLogEntry("test-1", "svc-a"))
	store.InsertLog(makeLogEntry("test-2", "svc-b"))

	services, err := store.GetServices()
	if err != nil {
		t.Fatalf("GetServices failed: %v", err)
	}
	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}
}

func TestCountLogsByMonth(t *testing.T) {
	store := newTestStore(t)

	now := time.Now().UTC()
	entry := makeLogEntry("test-1", "svc-a")
	entry.Timestamp = now
	store.InsertLog(entry)

	count, err := store.CountLogsByMonth(now.Year(), int(now.Month()))
	if err != nil {
		t.Fatalf("CountLogsByMonth failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestConcurrentWrites(t *testing.T) {
	store := newTestStore(t)
	var wg sync.WaitGroup
	errs := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			entry := makeLogEntry(fmt.Sprintf("concurrent-%d", i), "svc-a")
			if err := store.InsertLog(entry); err != nil {
				errs <- fmt.Errorf("insert %d: %w", i, err)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}

	total, err := store.GetTotalLogs()
	if err != nil {
		t.Fatalf("GetTotalLogs failed: %v", err)
	}
	if total != 50 {
		t.Errorf("expected 50 logs, got %d", total)
	}
}

func TestConcurrentReads(t *testing.T) {
	store := newTestStore(t)

	// Insert test data
	for i := 0; i < 10; i++ {
		store.InsertLog(makeLogEntry(fmt.Sprintf("read-%d", i), "svc-a"))
	}

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, count, err := store.GetLogsFiltered(1, 10, "", "", "", "")
			if err != nil {
				errs <- fmt.Errorf("GetLogsFiltered: %w", err)
				return
			}
			if count != 10 {
				errs <- fmt.Errorf("expected 10, got %d", count)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	store := newTestStore(t)
	var wg sync.WaitGroup
	errs := make(chan error, 100)

	// 25 concurrent writers
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			entry := makeLogEntry(fmt.Sprintf("mixed-%d", i), "svc-a")
			if err := store.InsertLog(entry); err != nil {
				errs <- fmt.Errorf("insert %d: %w", i, err)
			}
		}(i)
	}

	// 25 concurrent readers
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := store.GetLogsFiltered(1, 10, "", "", "", "")
			if err != nil {
				errs <- fmt.Errorf("GetLogsFiltered: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
