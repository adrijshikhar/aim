package agy_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	last_modified_time DATETIME NOT NULL,
	workspace_uris TEXT NOT NULL DEFAULT ''
);
INSERT INTO conversation_summaries (conversation_id, title, preview, last_modified_time, workspace_uris)
VALUES 
('775e6ada-1595-4e7e-84fa-ce0ea71e3007', 'Aim MMVP Refactor', 'Let us build a TUI', '2026-09-14 10:00:00', '["file:///Users/mock/projects/aim"]'),
('fcdbc2e0-2dc8-4ffa-9ee2-eb5aaa3e556f', 'Planning Binsight Release', 'Review PRs', '2026-09-14 09:00:00', '["file:///Users/mock/projects/catalyst/.catalyst/tasks","file:///Users/mock/projects/catalyst"]');
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
	if sessions[0].Cwd != "/Users/mock/projects/aim" {
		t.Errorf("expected Cwd '/Users/mock/projects/aim', got %q", sessions[0].Cwd)
	}
	if sessions[1].Cwd != "/Users/mock/projects/catalyst" {
		t.Errorf("expected Cwd '/Users/mock/projects/catalyst', got %q", sessions[1].Cwd)
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
	if s.Cwd != "/Users/mock/projects/aim" {
		t.Errorf("expected Cwd '/Users/mock/projects/aim', got %q", s.Cwd)
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

func TestProvider_LegacyDBSchemaFallback(t *testing.T) {
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	tokenDir := filepath.Join(tmpDir, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		t.Fatalf("failed to create mock token dir: %v", err)
	}

	// Schema without workspace_uris column (legacy db)
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
('12345678-1234-1234-1234-123456789abc', 'Legacy Session', 'Legacy preview', '2026-09-14 10:00:00');
`
	if err := exec.Command(sqliteBin, dbPath, schema).Run(); err != nil {
		t.Fatalf("failed to seed legacy sqlite DB: %v", err)
	}

	p := agy.NewProvider()
	ctx := context.Background()

	sessions, err := p.ListSessions(ctx, tmpDir, false)
	if err != nil {
		t.Fatalf("ListSessions on legacy DB failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session from legacy DB, got %d", len(sessions))
	}
	if sessions[0].Cwd != "" {
		t.Errorf("expected empty Cwd on legacy DB, got %q", sessions[0].Cwd)
	}

	s, err := p.GetSession(ctx, "12345678", tmpDir, false)
	if err != nil {
		t.Fatalf("GetSession on legacy DB failed: %v", err)
	}
	if s == nil || s.ID != "12345678-1234-1234-1234-123456789abc" {
		t.Fatalf("expected session 12345678..., got %+v", s)
	}
	if s.Cwd != "" {
		t.Errorf("expected empty Cwd for legacy session, got %q", s.Cwd)
	}
}

func TestProvider_TranscriptSummaryExtraction(t *testing.T) {
	mockProfileDir := setupMockAgyDB(t)
	p := agy.NewProvider()
	ctx := context.Background()

	// Create a mock transcript for session 775e6ada
	brainDir := filepath.Join(mockProfileDir, ".gemini", "antigravity-cli", "brain", "775e6ada-1595-4e7e-84fa-ce0ea71e3007", ".system_generated", "logs")
	if err := os.MkdirAll(brainDir, 0755); err != nil {
		t.Fatalf("failed to create brainDir: %v", err)
	}

	transcriptPath := filepath.Join(brainDir, "transcript.jsonl")
	transcriptContent := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"<USER_REQUEST>\nYou are implementing Task 6: Adopt K9s-Style Contextual Help Overlay (?)\n\n## Task Description\nRead your task brief first.\n</USER_REQUEST>"}`
	if err := os.WriteFile(transcriptPath, []byte(transcriptContent), 0644); err != nil {
		t.Fatalf("failed to write mock transcript: %v", err)
	}

	s, err := p.GetSession(ctx, "775e6ada", mockProfileDir, false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if s == nil {
		t.Fatalf("expected session, got nil")
	}

	// s.Summary should now contain the extracted user request from the transcript
	if !strings.Contains(s.Summary, "Adopt K9s-Style Contextual Help Overlay") {
		t.Errorf("expected summary to contain user request, got: %s", s.Summary)
	}
	if !strings.Contains(s.Summary, "Read your task brief first.") {
		t.Errorf("expected summary to include multi-line description, got: %s", s.Summary)
	}

	// Session without transcript should still fall back to DB preview
	s2, err := p.GetSession(ctx, "fcdbc2e0", mockProfileDir, false)
	if err != nil {
		t.Fatalf("GetSession for fcdbc2e0 failed: %v", err)
	}
	if s2 == nil || s2.Summary != "Review PRs" {
		t.Errorf("expected fallback preview 'Review PRs', got: %+v", s2)
	}
}

func TestProvider_SQLInjectionSafety(t *testing.T) {
	mockProfileDir := setupMockAgyDB(t)
	p := agy.NewProvider()
	ctx := context.Background()

	maliciousInputs := []string{
		"'; DROP TABLE conversation_summaries; --",
		"775e6ada' OR '1'='1",
		"775e6ada; SELECT * FROM conversation_summaries;",
		"775e6ada\x00extra",
		"775e6ada%",
		"775e6ada_",
	}

	for _, input := range maliciousInputs {
		t.Run(input, func(t *testing.T) {
			s, err := p.GetSession(ctx, input, mockProfileDir, false)
			// Must either cleanly return nil without error (rejected by whitelist)
			// or error out without executing injection.
			if s != nil {
				t.Errorf("malicious input %q unexpectedly returned a session", input)
			}
			_ = err
		})
	}

	// Verify table was not dropped
	sessions, err := p.ListSessions(ctx, mockProfileDir, false)
	if err != nil || len(sessions) != 2 {
		t.Fatalf("table conversation_summaries was compromised or corrupted: err=%v, sessions=%d", err, len(sessions))
	}
}
