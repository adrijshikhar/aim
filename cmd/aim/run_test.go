package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
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

