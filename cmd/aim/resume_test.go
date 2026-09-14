package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/profile"
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
