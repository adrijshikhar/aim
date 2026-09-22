package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
)

func TestResumeCmd_ArgValidation(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	reg.Register(codex.NewAdapter())
	pm := profile.NewProfileManager(tempDir)

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Missing args
	cmd.SetArgs([]string{"resume"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires <agent> and <profile>") {
		t.Errorf("expected error requiring agent and profile, got: %v", err)
	}

	// Missing session ID
	buf.Reset()
	cmd.SetArgs([]string{"resume", "agy", "work"})
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "session ID") {
		t.Errorf("expected error requiring session ID, got: %v", err)
	}
}

func TestResumeCmd_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"resume", "agy", "work", "nonexistent-session-id"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "session not found") {
		t.Errorf("expected session not found error, got: %v", err)
	}
}

func TestResumeCmd_CatalystBriefCreation(t *testing.T) {
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 not found in PATH")
	}

	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found in PATH")
	}

	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	// Initialize git repo in tempDir
	_ = exec.Command(gitBin, "-C", tempDir, "init").Run()
	_ = exec.Command(gitBin, "-C", tempDir, "checkout", "-b", "feat/my-feature").Run()

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	// Create mock Antigravity session in profile "work"
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
VALUES ('775e6ada-1595-4e7e-84fa-ce0ea71e3007', 'Resume Handoff Request', 'Summary of handoff', '2026-09-14 10:00:00');
`
	if err := exec.Command(sqliteBin, dbPath, schema).Run(); err != nil {
		t.Fatalf("failed to seed mock sqlite DB: %v", err)
	}

	// Also create mock fake binary for agy so executeRun doesn't fail on LookPath
	fakeBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(fakeBinDir, 0755)
	fakeAgy := filepath.Join(fakeBinDir, "agy")
	_ = os.WriteFile(fakeAgy, []byte("#!/bin/sh\nexit 0\n"), 0755)
	t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"resume", "agy", "work", "775e6ada", "--catalyst", "--force"})

	// Change working dir to git repo
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	_ = cmd.Execute()

	// Verify Catalyst handoff brief was created
	handoffPath := filepath.Join(tempDir, ".catalyst", "handoffs", "feat-my-feature.json")
	if fi, err := os.Stat(handoffPath); err != nil || fi.Size() == 0 {
		t.Errorf("expected handoff brief at %s, err: %v", handoffPath, err)
	}
}

func TestResume_SyncsLatestSession(t *testing.T) {
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 not found in PATH")
	}

	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(codex.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	pDirA, _ := pm.EnsureProfile("profile-a")
	_, _ = pm.EnsureProfile("profile-b")

	sessionID := "12345678-abcd-ef01-2345-6789abcdef01"

	// Profile A has a stale copy of the session (updated_at = 1700000000)
	codexDirA := filepath.Join(tempDir, "profiles", "profile-a", ".codex")
	_ = os.MkdirAll(codexDirA, 0755)
	dbA := filepath.Join(codexDirA, "state_5.sqlite")
	schemaA := fmt.Sprintf(`
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT ''
);
INSERT INTO threads (id, rollout_path, created_at, updated_at, title, preview)
VALUES ('%s', '/tmp/fake-a.jsonl', 1700000000, 1700000000, 'Stale Title Profile A', 'Preview A');
`, sessionID)
	if err := exec.Command(sqliteBin, dbA, schemaA).Run(); err != nil {
		t.Fatalf("failed to seed profile A codex DB: %v", err)
	}

	// Profile B has a newer copy of the session (updated_at = 1800000000)
	codexDirB := filepath.Join(tempDir, "profiles", "profile-b", ".codex")
	_ = os.MkdirAll(codexDirB, 0755)
	dbB := filepath.Join(codexDirB, "state_5.sqlite")
	schemaB := fmt.Sprintf(`
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT ''
);
INSERT INTO threads (id, rollout_path, created_at, updated_at, title, preview)
VALUES ('%s', '/tmp/fake-b.jsonl', 1700000000, 1800000000, 'Updated Title Profile B', 'Preview B');
`, sessionID)
	if err := exec.Command(sqliteBin, dbB, schemaB).Run(); err != nil {
		t.Fatalf("failed to seed profile B codex DB: %v", err)
	}

	// Fake codex binary so runner exits cleanly
	fakeBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(fakeBinDir, 0755)
	fakeCodex := filepath.Join(fakeBinDir, "codex")
	_ = os.WriteFile(fakeCodex, []byte("#!/bin/sh\nexit 0\n"), 0755)
	t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	mgr := defaultSessionManager()

	// Initial stale session pointing to profile-a
	sessStale := &session.Session{
		ID:           sessionID,
		ShortID:      session.ComputeShortID(sessionID),
		Title:        "Stale Title Profile A",
		Agent:        "codex",
		Profile:      "profile-a",
		IsHost:       false,
		LastActiveAt: time.Unix(1700000000, 0),
		Status:       session.StatusIdle,
	}

	// Execute exact resume into profile-a with the stale session
	err = executeExactResume(cmd, reg, pm, mgr, "codex", "profile-a", pDirA, sessStale, false, nil)
	if err != nil {
		t.Fatalf("executeExactResume failed: %v", err)
	}

	// Verify profile-a's state_5.sqlite now has the updated title from profile B
	checkCmd := exec.Command(sqliteBin, dbA, fmt.Sprintf("SELECT title FROM threads WHERE id='%s';", sessionID))
	checkOut, err := checkCmd.Output()
	if err != nil {
		t.Fatalf("failed to query profile A DB: %v", err)
	}
	if !strings.Contains(string(checkOut), "Updated Title Profile B") {
		t.Errorf("expected hydrated thread with title 'Updated Title Profile B' in profile A, got: %q", string(checkOut))
	}
}
