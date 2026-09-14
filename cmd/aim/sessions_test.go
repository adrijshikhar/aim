package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
)

func setupMockSessionEnv(t *testing.T) (string, *agents.Registry, *profile.ProfileManager) {
	t.Helper()
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	// Set up mock Antigravity session in profile "work"
	agyDbDir := filepath.Join(tempDir, "profiles", "work", ".gemini", "antigravity-cli")
	_ = os.MkdirAll(agyDbDir, 0755)
	dbPath := filepath.Join(agyDbDir, "conversation_summaries.db")

	schema := `
CREATE TABLE conversation_summaries (
	conversation_id TEXT PRIMARY KEY,
	title TEXT NOT NULL DEFAULT '',
	preview TEXT NOT NULL DEFAULT '',
	last_modified_time DATETIME NOT NULL
);
INSERT INTO conversation_summaries (conversation_id, title, preview, last_modified_time)
VALUES 
('775e6ada-1595-4e7e-84fa-ce0ea71e3007', 'Resume Handoff Request', 'Summary of handoff', '2026-09-14 10:00:00'),
('deadbeef-1234-5678-90ab-cdef12345678', 'Past Idle Task', 'Past summary', '2026-09-13 10:00:00');
`
	if err := exec.Command(sqliteBin, dbPath, schema).Run(); err != nil {
		t.Fatalf("failed to seed mock sqlite DB: %v", err)
	}

	return tempDir, reg, pm
}

func TestSessionsCmd_Empty(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No sessions found") {
		t.Errorf("expected output to contain 'No sessions found', got: %s", out)
	}
}

func TestSessionsCmd_JSON(t *testing.T) {
	_, reg, pm := setupMockSessionEnv(t)

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sessions []session.Session
	if err := json.Unmarshal(buf.Bytes(), &sessions); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v, raw: %s", err, buf.String())
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions in JSON, got %d", len(sessions))
	}
}

func TestSessionsCmd_Table(t *testing.T) {
	_, reg, pm := setupMockSessionEnv(t)

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "RECENT SESSIONS") {
		t.Errorf("expected output to have 'RECENT SESSIONS', got: %s", out)
	}
	if !strings.Contains(out, "775e6ada") {
		t.Errorf("expected output to contain short ID '775e6ada', got: %s", out)
	}
	if !strings.Contains(out, "Resume Handoff Request") {
		t.Errorf("expected output to contain title, got: %s", out)
	}
}

func TestSessionsCmd_Filters(t *testing.T) {
	_, reg, pm := setupMockSessionEnv(t)

	// Filter by nonexistent profile
	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions", "--profile", "nonexistent"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No sessions found") {
		t.Errorf("expected no sessions found for nonexistent profile, got: %s", buf.String())
	}

	// Filter by agent arg
	buf.Reset()
	cmd = newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions", "agy", "--json"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sessions []session.Session
	if err := json.Unmarshal(buf.Bytes(), &sessions); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 agy sessions, got %d", len(sessions))
	}
}

func TestSessionsShowCmd(t *testing.T) {
	_, reg, pm := setupMockSessionEnv(t)

	// 1. Show by short prefix
	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions", "show", "deadbeef"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Conversation Session Preview") {
		t.Errorf("expected output to contain 'Conversation Session Preview', got: %s", out)
	}
	if !strings.Contains(out, "Past Idle Task") {
		t.Errorf("expected output to contain title, got: %s", out)
	}
	if !strings.Contains(out, "Past summary") {
		t.Errorf("expected output to contain summary preview, got: %s", out)
	}
	if !strings.Contains(out, "aim resume agy work deadbeef") {
		t.Errorf("expected output to contain quick resume tip, got: %s", out)
	}

	// 2. Show JSON
	buf.Reset()
	cmd = newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions", "show", "--json", "deadbeef"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sess session.Session
	if err := json.Unmarshal(buf.Bytes(), &sess); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if sess.ShortID != "deadbeef" {
		t.Errorf("expected ShortID deadbeef, got %s", sess.ShortID)
	}
	if sess.Summary != "Past summary" {
		t.Errorf("expected summary 'Past summary', got %s", sess.Summary)
	}

	// 3. Show non-existent
	buf.Reset()
	cmd = newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions", "show", "nonexistent"})

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for non-existent session, got nil")
	}
}
