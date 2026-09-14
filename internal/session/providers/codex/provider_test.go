package codex_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/session/providers/codex"
)

func setupMockCodex(t *testing.T) string {
	t.Helper()
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	codexDir := filepath.Join(tmpDir, ".codex")
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		t.Fatalf("failed to create mock codex dir: %v", err)
	}

	dbPath := filepath.Join(codexDir, "state_5.sqlite")
	schema := `
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT ''
);
INSERT INTO threads (id, rollout_path, created_at, updated_at, title, preview)
VALUES 
('01a09eb7-2f6c-7c52-895f-218f9ac9eecd', '/tmp/rollout1.jsonl', 1726300000, 1726300000, 'Refactor router middleware', 'Clean up routes'),
('01a04baf-4b95-7681-82bd-81df4f995452', '/tmp/rollout2.jsonl', 1726200000, 1726200000, 'Update docs', 'Fix outdated markdown');
`
	if err := exec.Command(sqliteBin, dbPath, schema).Run(); err != nil {
		t.Fatalf("failed to seed mock sqlite DB: %v", err)
	}

	// Also write a mock session_index.jsonl
	indexFile := filepath.Join(codexDir, "session_index.jsonl")
	indexContent := `{"id":"01a09eb7-2f6c-7c52-895f-218f9ac9eecd","thread_name":"Refactor router middleware","updated_at":"2026-09-14T07:12:52Z"}
{"id":"01a04baf-4b95-7681-82bd-81df4f995452","thread_name":"Update docs","updated_at":"2026-09-13T04:03:40Z"}
`
	_ = os.WriteFile(indexFile, []byte(indexContent), 0644)

	return tmpDir
}

func TestProvider_ListSessions(t *testing.T) {
	mockProfileDir := setupMockCodex(t)
	p := codex.NewProvider()

	if p.Agent() != "codex" {
		t.Fatalf("expected agent 'codex', got %q", p.Agent())
	}

	ctx := context.Background()
	sessions, err := p.ListSessions(ctx, mockProfileDir, false)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	if sessions[0].ID != "01a09eb7-2f6c-7c52-895f-218f9ac9eecd" {
		t.Errorf("expected first session 01a09eb7..., got %s", sessions[0].ID)
	}
	if sessions[0].ShortID != "01a09eb7" {
		t.Errorf("expected ShortID 01a09eb7, got %s", sessions[0].ShortID)
	}
	if sessions[0].Title != "Refactor router middleware" {
		t.Errorf("expected title 'Refactor router middleware', got %q", sessions[0].Title)
	}
}

func TestProvider_GetSession(t *testing.T) {
	mockProfileDir := setupMockCodex(t)
	p := codex.NewProvider()
	ctx := context.Background()

	s, err := p.GetSession(ctx, "01a09eb7", mockProfileDir, false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if s == nil || s.ID != "01a09eb7-2f6c-7c52-895f-218f9ac9eecd" {
		t.Fatalf("expected session 01a09eb7..., got %+v", s)
	}
}

func TestProvider_Hydrate(t *testing.T) {
	srcDir := setupMockCodex(t)
	_ = srcDir
	targetDir := t.TempDir()

	p := codex.NewProvider()
	ctx := context.Background()

	// Create a dummy rollout file
	dummyRollout := filepath.Join(t.TempDir(), "dummy-rollout.jsonl")
	_ = os.WriteFile(dummyRollout, []byte(`{"turn": 1}`), 0644)

	srcSession := session.NewSession("01a09eb7-2f6c-7c52-895f-218f9ac9eecd", "Test Hydrate", "codex", "<host>", true, time.Now())
	srcSession.StoragePath = dummyRollout
	srcSession.Summary = "Preview text"

	newID, err := p.Hydrate(ctx, &srcSession, targetDir, false)
	if err != nil {
		t.Fatalf("Hydrate failed: %v", err)
	}
	if newID != srcSession.ID {
		t.Errorf("expected newID %s, got %s", srcSession.ID, newID)
	}

	// Verify target has session file and sqlite entry
	targetDB := filepath.Join(targetDir, ".codex", "state_5.sqlite")
	if _, err := os.Stat(targetDB); err != nil {
		t.Errorf("expected target state_5.sqlite to exist: %v", err)
	}
}

func TestProvider_Hydrate_Fork(t *testing.T) {
	setupMockCodex(t)
	targetDir := t.TempDir()

	p := codex.NewProvider()
	ctx := context.Background()

	dummyRollout := filepath.Join(t.TempDir(), "dummy-rollout.jsonl")
	_ = os.WriteFile(dummyRollout, []byte(`{"turn": 1}`), 0644)

	srcSession := session.NewSession("01a09eb7-2f6c-7c52-895f-218f9ac9eecd", "Fork Test", "codex", "<host>", true, time.Now())
	srcSession.StoragePath = dummyRollout

	forkedID, err := p.Hydrate(ctx, &srcSession, targetDir, true)
	if err != nil {
		t.Fatalf("Hydrate fork failed: %v", err)
	}
	if forkedID == srcSession.ID {
		t.Errorf("expected forked ID to be different from source ID, got %s", forkedID)
	}
	if len(forkedID) != 36 {
		t.Errorf("expected 36-character UUID for forked ID, got %s", forkedID)
	}

	// Verify rollout file was copied with forked ID
	forkedRollout := filepath.Join(targetDir, ".codex", "sessions", "rollout-"+forkedID+".jsonl")
	if _, err := os.Stat(forkedRollout); err != nil {
		t.Errorf("expected forked rollout file to exist: %v", err)
	}
}

func TestProvider_SQLInjectionSafety(t *testing.T) {
	mockProfileDir := setupMockCodex(t)
	p := codex.NewProvider()
	ctx := context.Background()

	// 1. Malicious session ID in Hydrate
	maliciousSession := session.NewSession("'; DROP TABLE threads; --", "Hacked Title", "codex", "<host>", true, time.Now())
	targetDir := t.TempDir()
	_, err := p.Hydrate(ctx, &maliciousSession, targetDir, false)
	if err == nil {
		t.Errorf("expected error for malicious session ID in Hydrate, got nil")
	}

	// 2. Querying with malicious prefix
	s, err := p.GetSession(ctx, "'; DROP TABLE threads; --", mockProfileDir, false)
	if s != nil {
		t.Errorf("malicious prefix unexpectedly returned a session")
	}
	_ = err
}
