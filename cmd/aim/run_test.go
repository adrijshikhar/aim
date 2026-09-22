package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
)

func TestRunCmd_ProfileExists_NoPrompt(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "mock", "work"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(stdout.String(), "does not exist") {
		t.Errorf("expected no missing profile prompt, got: %s", stdout.String())
	}
}

func TestRunCmd_ProfileDoesNotExist_Interactive_ConfirmYes(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})
	pm := profile.NewProfileManager(tempDir)

	oldInteractive := isInteractiveFunc
	isInteractiveFunc = func(r io.Reader) bool { return true }
	defer func() { isInteractiveFunc = oldInteractive }()

	var stdout, stderr bytes.Buffer
	stdin := bytes.NewBufferString("y\n")

	cmd := newRootCmd(reg, pm)
	cmd.SetIn(stdin)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "mock", "newprof"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `Profile "newprof" does not exist. Do you want to create it and start mock? [y/N]:`) {
		t.Errorf("expected creation prompt, got: %s", out)
	}

	// Verify profile directory was created
	pDir := filepath.Join(tempDir, "profiles", "newprof")
	if fi, err := os.Stat(pDir); err != nil || !fi.IsDir() {
		t.Errorf("expected profile directory %s to be created on confirm", pDir)
	}
}

func TestRunCmd_ProfileDoesNotExist_Interactive_ConfirmNo(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})
	pm := profile.NewProfileManager(tempDir)

	oldInteractive := isInteractiveFunc
	isInteractiveFunc = func(r io.Reader) bool { return true }
	defer func() { isInteractiveFunc = oldInteractive }()

	var stdout, stderr bytes.Buffer
	stdin := bytes.NewBufferString("n\n")

	cmd := newRootCmd(reg, pm)
	cmd.SetIn(stdin)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "mock", "canceledprof"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error on abort: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `Profile "canceledprof" does not exist`) {
		t.Errorf("expected creation prompt, got: %s", out)
	}
	if !strings.Contains(out, "Profile creation aborted.") {
		t.Errorf("expected abort message, got: %s", out)
	}

	// Verify profile directory was NOT created
	pDir := filepath.Join(tempDir, "profiles", "canceledprof")
	if _, err := os.Stat(pDir); !os.IsNotExist(err) {
		t.Errorf("expected profile directory %s to NOT exist after cancellation", pDir)
	}
}

func TestRunCmd_ProfileDoesNotExist_Interactive_TypoDashes(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name:       "codex",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})
	pm := profile.NewProfileManager(tempDir)

	oldInteractive := isInteractiveFunc
	isInteractiveFunc = func(r io.Reader) bool { return true }
	defer func() { isInteractiveFunc = oldInteractive }()

	var stdout, stderr bytes.Buffer
	stdin := bytes.NewBufferString("\n") // user hits Enter to cancel

	cmd := newRootCmd(reg, pm)
	cmd.SetIn(stdin)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "codex", "office--yolo"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error on abort: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `Warning: profile name "office--yolo" contains "--". Did you mean "office" with flag "--yolo"?`) {
		t.Errorf("expected dash typo warning, got: %s", out)
	}
	if !strings.Contains(out, `Profile "office--yolo" does not exist. Do you want to create it and start codex? [y/N]:`) {
		t.Errorf("expected prompt, got: %s", out)
	}
	if !strings.Contains(out, "Profile creation aborted.") {
		t.Errorf("expected abort message, got: %s", out)
	}

	// Verify office--yolo was NOT created
	pDir := filepath.Join(tempDir, "profiles", "office--yolo")
	if _, err := os.Stat(pDir); !os.IsNotExist(err) {
		t.Errorf("expected office--yolo to NOT be created at %s", pDir)
	}
}

func TestRunCmd_ProfileDoesNotExist_NonInteractive(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})
	pm := profile.NewProfileManager(tempDir)

	oldInteractive := isInteractiveFunc
	isInteractiveFunc = func(r io.Reader) bool { return false }
	defer func() { isInteractiveFunc = oldInteractive }()

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "mock", "nonexistent"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), `profile "nonexistent" does not exist`) {
		t.Errorf("expected error for non-existent profile in non-interactive mode, got: %v", err)
	}

	// Verify profile directory was NOT created
	pDir := filepath.Join(tempDir, "profiles", "nonexistent")
	if _, err := os.Stat(pDir); !os.IsNotExist(err) {
		t.Errorf("expected profile directory %s to NOT exist", pDir)
	}
}

func TestRunCmd_ProfileDoesNotExist_FlagYes(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})
	pm := profile.NewProfileManager(tempDir)

	oldInteractive := isInteractiveFunc
	isInteractiveFunc = func(r io.Reader) bool { return false }
	defer func() { isInteractiveFunc = oldInteractive }()

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "mock", "newprof", "-y"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error with -y flag: %v", err)
	}

	pDir := filepath.Join(tempDir, "profiles", "newprof")
	if _, err := os.Stat(pDir); os.IsNotExist(err) {
		t.Errorf("expected profile directory %s to be created with -y", pDir)
	}
}

func TestRunCmd_ProfileDoesNotExist_EnvAutoCreate(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_AUTO_CREATE", "1")

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})
	pm := profile.NewProfileManager(tempDir)

	oldInteractive := isInteractiveFunc
	isInteractiveFunc = func(r io.Reader) bool { return false }
	defer func() { isInteractiveFunc = oldInteractive }()

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "mock", "envprof"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error with AIM_AUTO_CREATE=1: %v", err)
	}

	pDir := filepath.Join(tempDir, "profiles", "envprof")
	if _, err := os.Stat(pDir); os.IsNotExist(err) {
		t.Errorf("expected profile directory %s to be created with AIM_AUTO_CREATE=1", pDir)
	}
}

func TestRunCmd_AutoHydrateOnResume(t *testing.T) {
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
	_, _ = pm.EnsureProfile("source_prof")
	_, _ = pm.EnsureProfile("target_prof")

	sessionID := "87654321-dcba-fe01-4321-abcdef012345"

	// Create session only in source_prof
	sourceCodexDir := filepath.Join(tempDir, "profiles", "source_prof", ".codex")
	_ = os.MkdirAll(sourceCodexDir, 0755)
	dbSource := filepath.Join(sourceCodexDir, "state_5.sqlite")
	schema := fmt.Sprintf(`
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT ''
);
INSERT INTO threads (id, rollout_path, created_at, updated_at, title, preview)
VALUES ('%s', '/tmp/fake-source.jsonl', 1700000000, 1700000000, 'Source Thread Title', 'Preview');
`, sessionID)
	if err := exec.Command(sqliteBin, dbSource, schema).Run(); err != nil {
		t.Fatalf("failed to seed source_prof codex DB: %v", err)
	}

	// Fake codex binary
	fakeBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(fakeBinDir, 0755)
	fakeCodex := filepath.Join(fakeBinDir, "codex")
	_ = os.WriteFile(fakeCodex, []byte("#!/bin/sh\nexit 0\n"), 0755)
	t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "codex", "target_prof", "resume", "87654321"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("aim run failed: %v", err)
	}

	// Verify target_prof now has the thread hydrated from source_prof
	targetDB := filepath.Join(tempDir, "profiles", "target_prof", ".codex", "state_5.sqlite")
	checkCmd := exec.Command(sqliteBin, targetDB, fmt.Sprintf("SELECT title FROM threads WHERE id='%s';", sessionID))
	checkOut, err := checkCmd.Output()
	if err != nil {
		t.Fatalf("failed to query target DB: %v", err)
	}
	if !strings.Contains(string(checkOut), "Source Thread Title") {
		t.Errorf("expected hydrated thread with title 'Source Thread Title' in target_prof, got: %q", string(checkOut))
	}
}

func TestRunCmd_AutoHydrateOnResume_SyncsNewer(t *testing.T) {
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
	_, _ = pm.EnsureProfile("source_prof")
	_, _ = pm.EnsureProfile("target_prof")

	sessionID := "99998888-abcd-ef01-4321-abcdef012345"

	// Create older session in target_prof
	targetCodexDir := filepath.Join(tempDir, "profiles", "target_prof", ".codex")
	_ = os.MkdirAll(targetCodexDir, 0755)
	dbTarget := filepath.Join(targetCodexDir, "state_5.sqlite")
	schemaTarget := fmt.Sprintf(`
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT ''
);
INSERT INTO threads (id, rollout_path, created_at, updated_at, title, preview)
VALUES ('%s', '/tmp/fake-target.jsonl', 1600000000, 1600000000, 'Older Target Thread Title', 'Preview');
`, sessionID)
	if err := exec.Command(sqliteBin, dbTarget, schemaTarget).Run(); err != nil {
		t.Fatalf("failed to seed target_prof codex DB: %v", err)
	}

	// Create newer session in source_prof
	sourceCodexDir := filepath.Join(tempDir, "profiles", "source_prof", ".codex")
	_ = os.MkdirAll(sourceCodexDir, 0755)
	dbSource := filepath.Join(sourceCodexDir, "state_5.sqlite")
	schemaSource := fmt.Sprintf(`
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT ''
);
INSERT INTO threads (id, rollout_path, created_at, updated_at, title, preview)
VALUES ('%s', '/tmp/fake-source.jsonl', 1600000000, 1700000000, 'Newer Source Thread Title', 'Preview');
`, sessionID)
	if err := exec.Command(sqliteBin, dbSource, schemaSource).Run(); err != nil {
		t.Fatalf("failed to seed source_prof codex DB: %v", err)
	}

	// Fake codex binary
	fakeBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(fakeBinDir, 0755)
	fakeCodex := filepath.Join(fakeBinDir, "codex")
	_ = os.WriteFile(fakeCodex, []byte("#!/bin/sh\nexit 0\n"), 0755)
	t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "codex", "target_prof", "resume", "99998888"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("aim run failed: %v", err)
	}

	// Verify target_prof now has the newer thread hydrated from source_prof
	checkCmd := exec.Command(sqliteBin, dbTarget, fmt.Sprintf("SELECT title FROM threads WHERE id='%s';", sessionID))
	checkOut, err := checkCmd.Output()
	if err != nil {
		t.Fatalf("failed to query target DB: %v", err)
	}
	if !strings.Contains(string(checkOut), "Newer Source Thread Title") {
		t.Errorf("expected updated thread with title 'Newer Source Thread Title' in target_prof, got: %q", string(checkOut))
	}
}

func TestRunCmd_NoSelfHydration(t *testing.T) {
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
	_, _ = pm.EnsureProfile("native_prof")

	sessionID := "77776666-abcd-ef01-4321-abcdef012345"

	// Session exists only in native_prof
	nativeCodexDir := filepath.Join(tempDir, "profiles", "native_prof", ".codex")
	_ = os.MkdirAll(nativeCodexDir, 0755)
	dbNative := filepath.Join(nativeCodexDir, "state_5.sqlite")
	schema := fmt.Sprintf(`
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT ''
);
INSERT INTO threads (id, rollout_path, created_at, updated_at, title, preview)
VALUES ('%s', '/tmp/fake-native.jsonl', 1700000000, 1700000000, 'Native Thread Title', 'Preview');
`, sessionID)
	if err := exec.Command(sqliteBin, dbNative, schema).Run(); err != nil {
		t.Fatalf("failed to seed native_prof codex DB: %v", err)
	}

	fakeBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(fakeBinDir, 0755)
	fakeCodex := filepath.Join(fakeBinDir, "codex")
	_ = os.WriteFile(fakeCodex, []byte("#!/bin/sh\nexit 0\n"), 0755)
	t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "codex", "native_prof", "resume", "77776666"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("aim run failed: %v", err)
	}

	// Output must NOT contain "Syncing latest" because it already belongs to native_prof
	if strings.Contains(stdout.String(), "Syncing latest") {
		t.Errorf("expected no syncing output for native session, got: %s", stdout.String())
	}
}

func TestRunAndResume_AIMSessionIDPropagation(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(codex.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	dumpFile := filepath.Join(tempDir, "env_dump.txt")
	fakeBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(fakeBinDir, 0755)
	fakeCodex := filepath.Join(fakeBinDir, "codex")
	fakeScript := fmt.Sprintf("#!/bin/sh\necho \"SESSION=$AIM_SESSION_ID\" > %q\nexit 0\n", dumpFile)
	_ = os.WriteFile(fakeCodex, []byte(fakeScript), 0755)
	t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

	// Case 1: aim run with resume argument
	_ = os.Remove(dumpFile)
	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "codex", "work", "resume", "test-session-123"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("aim run failed: %v", err)
	}

	content, err := os.ReadFile(dumpFile)
	if err != nil {
		t.Fatalf("failed to read env dump: %v", err)
	}
	if strings.TrimSpace(string(content)) != "SESSION=test-session-123" {
		t.Errorf("expected SESSION=test-session-123, got %q", string(content))
	}

	// Case 2: aim run without session ID -> AIM_SESSION_ID should be empty
	_ = os.Remove(dumpFile)
	stdout.Reset()
	stderr.Reset()
	cmd = newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "codex", "work"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("aim run without session failed: %v", err)
	}

	content, err = os.ReadFile(dumpFile)
	if err != nil {
		t.Fatalf("failed to read env dump: %v", err)
	}
	if strings.TrimSpace(string(content)) != "SESSION=" {
		t.Errorf("expected empty SESSION, got %q", string(content))
	}
}

type testRunProvider struct {
	agent   string
	session session.Session
}

func (p *testRunProvider) Agent() string { return p.agent }
func (p *testRunProvider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	return []session.Session{p.session}, nil
}
func (p *testRunProvider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	if strings.HasPrefix(p.session.ID, idOrPrefix) || strings.HasPrefix(p.session.ShortID, idOrPrefix) {
		copy := p.session
		return &copy, nil
	}
	return nil, nil
}
func (p *testRunProvider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	return srcSession.ID, nil
}

func TestRunCmd_ExpandsShortPrefixToFullUUID(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(codex.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	fullUUID := "01a0c8fe-ec30-7b22-bf66-1edcc7516885"
	shortID := "01a0c8fe"

	sess := session.Session{
		ID:           fullUUID,
		ShortID:      shortID,
		Agent:        "codex",
		Profile:      "work",
		LastActiveAt: time.Now(),
	}

	mockMgr := session.NewManager()
	mockMgr.RegisterProvider(&testRunProvider{
		agent:   "codex",
		session: sess,
	})
	oldMgrFunc := defaultSessionManager
	defaultSessionManager = func() *session.Manager {
		return mockMgr
	}
	t.Cleanup(func() {
		defaultSessionManager = oldMgrFunc
	})

	dumpFile := filepath.Join(tempDir, "args_dump.txt")
	fakeBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(fakeBinDir, 0755)
	fakeCodex := filepath.Join(fakeBinDir, "codex")
	fakeScript := fmt.Sprintf("#!/bin/sh\necho \"$@\" > %q\nexit 0\n", dumpFile)
	_ = os.WriteFile(fakeCodex, []byte(fakeScript), 0755)
	t.Setenv("PATH", fakeBinDir+":"+os.Getenv("PATH"))

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"run", "codex", "work", "resume", shortID})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("aim run failed: %v", err)
	}

	content, err := os.ReadFile(dumpFile)
	if err != nil {
		t.Fatalf("failed to read args dump: %v", err)
	}
	expected := "resume " + fullUUID + "\n"
	if string(content) != expected {
		t.Errorf("expected args %q, got %q", expected, string(content))
	}
}
