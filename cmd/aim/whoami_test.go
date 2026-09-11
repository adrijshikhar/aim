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

func TestWhoami_NoActiveSession(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("personal")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"whoami"})
	err := cmd.Execute()

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "No active AIM session in this shell") {
		t.Errorf("expected output to mention no active session, got: %s", output)
	}
	if !strings.Contains(output, "work") || !strings.Contains(output, "personal") {
		t.Errorf("expected output to list configured profiles, got: %s", output)
	}
}

func TestWhoami_ActiveSession(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_PROFILE", "bby")
	t.Setenv("AIM_AGENT", "agy")
	profileHome := filepath.Join(tempDir, "profiles", "bby")
	t.Setenv("HOME", profileHome)
	t.Setenv("ANTIGRAVITY_SOURCE_METADATA", "")

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("bby")

	// Create mock presence lock file
	presenceDir := filepath.Join(profileHome, ".gemini", "antigravity-cli", "presence")
	_ = os.MkdirAll(presenceDir, 0755)
	_ = os.WriteFile(filepath.Join(presenceDir, "mock-conv-12345678.lock"), []byte(""), 0600)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"current"}) // test alias
	err := cmd.Execute()

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Active AIM Session") {
		t.Errorf("expected 'Active AIM Session', got: %s", output)
	}
	if !strings.Contains(output, "bby") {
		t.Errorf("expected profile 'bby', got: %s", output)
	}
	if !strings.Contains(output, "agy") {
		t.Errorf("expected agent 'agy', got: %s", output)
	}
	if !strings.Contains(output, "mock-con") {
		t.Errorf("expected session ID to appear, got: %s", output)
	}
}

func TestWhoami_ActiveSession_FromMetadata(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_PROFILE", "bby")
	t.Setenv("AIM_AGENT", "agy")
	profileHome := filepath.Join(tempDir, "profiles", "bby")
	t.Setenv("HOME", profileHome)
	t.Setenv("ANTIGRAVITY_SOURCE_METADATA", `{"tool":{"conversationId":"meta-conv-98765432"}}`)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("bby")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"whoami"})
	err := cmd.Execute()

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "meta-con") {
		t.Errorf("expected metadata conversation ID to appear, got: %s", output)
	}
}

func TestWhoami_InferFromHome(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")
	profileHome := filepath.Join(tempDir, "profiles", "inferred_prof")
	t.Setenv("HOME", profileHome)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("inferred_prof")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"whoami"})
	err := cmd.Execute()

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "inferred_prof") {
		t.Errorf("expected profile 'inferred_prof' inferred from HOME, got: %s", output)
	}
}
