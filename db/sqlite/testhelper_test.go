package sqlite_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotail/db/sqlite"
)

func newTestStore(t *testing.T) *sqlite.SQLiteStore {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	store, err := sqlite.NewSQLiteStore(dsn)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Apply migration SQL directly
	migrationSQL, err := os.ReadFile("../../migrations/20250625192434_log.sql")
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	// Extract only the Up migration (between +goose Up and +goose Down)
	sql := extractUpMigration(string(migrationSQL))

	if err := store.ExecRaw(sql); err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	t.Cleanup(func() {
		store.Close()
	})

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
