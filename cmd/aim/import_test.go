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

func TestImportCmd_ArgValidation(t *testing.T) {
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
	cmd.SetArgs([]string{"sessions", "import"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires <agent> and <target-profile>") {
		t.Errorf("expected error requiring agent and target-profile, got: %v", err)
	}

	// Missing session ID and missing --all
	buf.Reset()
	cmd.SetArgs([]string{"sessions", "import", "codex", "work"})
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "must specify session-id or --all") {
		t.Errorf("expected error requiring session-id or --all, got: %v", err)
	}
}

func TestImportCmd_HydrateSession(t *testing.T) {
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
	_, _ = pm.EnsureProfile("work")

	// Create host mock codex session
	hostCodexDir := filepath.Join(tempDir, ".codex")
	_ = os.MkdirAll(hostCodexDir, 0755)
	hostDB := filepath.Join(hostCodexDir, "state_5.sqlite")

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
VALUES ('12345678-abcd-ef01-2345-6789abcdef01', '/tmp/fake.jsonl', 1700000000, 1700000000, 'Import Test Thread', 'Preview');
`
	if err := exec.Command(sqliteBin, hostDB, schema).Run(); err != nil {
		t.Fatalf("failed to seed host codex DB: %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"sessions", "import", "codex", "work", "12345678"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing sessions import: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Imported session") {
		t.Errorf("expected output to mention imported session, got: %s", out)
	}

	// Verify thread was hydrated into target profile's state_5.sqlite
	targetDB := filepath.Join(tempDir, "profiles", "work", ".codex", "state_5.sqlite")
	checkCmd := exec.Command(sqliteBin, targetDB, "SELECT title FROM threads WHERE id='12345678-abcd-ef01-2345-6789abcdef01';")
	checkOut, err := checkCmd.Output()
	if err != nil || !strings.Contains(string(checkOut), "Import Test Thread") {
		t.Errorf("expected hydrated thread in target DB, got: %s (err: %v)", string(checkOut), err)
	}
}
