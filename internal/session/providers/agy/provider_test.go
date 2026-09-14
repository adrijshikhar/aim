package agy_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/session/providers/agy"
)

func setupMockAgyDB(t *testing.T) string {
	t.Helper()
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	tokenDir := filepath.Join(tmpDir, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		t.Fatalf("failed to create mock token dir: %v", err)
	}

	dbPath := filepath.Join(tokenDir, "conversation_summaries.db")
	schema := `
CREATE TABLE conversation_summaries (
	conversation_id TEXT PRIMARY KEY,
	title TEXT NOT NULL DEFAULT '',
	preview TEXT NOT NULL DEFAULT '',
	last_modified_time DATETIME NOT NULL
);
INSERT INTO conversation_summaries (conversation_id, title, preview, last_modified_time)
VALUES 
('775e6ada-1595-4e7e-84fa-ce0ea71e3007', 'Aim MMVP Refactor', 'Let us build a TUI', '2026-09-14 10:00:00'),
('fcdbc2e0-2dc8-4ffa-9ee2-eb5aaa3e556f', 'Planning Binsight Release', 'Review PRs', '2026-09-14 09:00:00');
`
	cmd := exec.Command(sqliteBin, dbPath)
	cmd.Stdin = os.Stdin
	if err := exec.Command(sqliteBin, dbPath, schema).Run(); err != nil {
		t.Fatalf("failed to seed mock sqlite DB: %v", err)
	}

	return tmpDir
}

func TestProvider_ListSessions(t *testing.T) {
	mockProfileDir := setupMockAgyDB(t)
	p := agy.NewProvider()

	if p.Agent() != "agy" {
		t.Fatalf("expected agent 'agy', got %q", p.Agent())
	}

	ctx := context.Background()
	sessions, err := p.ListSessions(ctx, mockProfileDir, false)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	if sessions[0].ID != "775e6ada-1595-4e7e-84fa-ce0ea71e3007" {
		t.Errorf("expected first session 775e6ada..., got %s", sessions[0].ID)
	}
	if sessions[0].ShortID != "775e6ada" {
		t.Errorf("expected ShortID 775e6ada, got %s", sessions[0].ShortID)
	}
	if sessions[0].Title != "Aim MMVP Refactor" {
		t.Errorf("expected title 'Aim MMVP Refactor', got %q", sessions[0].Title)
	}
}

func TestProvider_GetSession(t *testing.T) {
	mockProfileDir := setupMockAgyDB(t)
	p := agy.NewProvider()
	ctx := context.Background()

	// Test prefix lookup
	s, err := p.GetSession(ctx, "775e6ada", mockProfileDir, false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if s == nil || s.ID != "775e6ada-1595-4e7e-84fa-ce0ea71e3007" {
		t.Fatalf("expected session 775e6ada..., got %+v", s)
	}

	// Test not found
	nonExistent, err := p.GetSession(ctx, "nonexistent-prefix", mockProfileDir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nonExistent != nil {
		t.Fatalf("expected nil for non-existent session, got %+v", nonExistent)
	}
}
