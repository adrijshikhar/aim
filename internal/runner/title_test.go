package runner

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
)

func TestTerminalTitle_SetAndReset(t *testing.T) {
	var buf bytes.Buffer

	// Test SetTerminalTitle
	title := "AIM: [codex] work (1234abcd)"
	err := SetTerminalTitle(&buf, title)
	if err != nil {
		t.Fatalf("SetTerminalTitle failed: %v", err)
	}

	expectedSet := "\033]0;AIM: [codex] work (1234abcd)\007"
	if buf.String() != expectedSet {
		t.Errorf("expected %q, got %q", expectedSet, buf.String())
	}

	// Test ResetTerminalTitle
	buf.Reset()
	err = ResetTerminalTitle(&buf)
	if err != nil {
		t.Fatalf("ResetTerminalTitle failed: %v", err)
	}

	expectedReset := "\033]0;\007"
	if buf.String() != expectedReset {
		t.Errorf("expected %q, got %q", expectedReset, buf.String())
	}

	// Test nil writer safety
	if err := SetTerminalTitle(nil, title); err != nil {
		t.Errorf("expected nil error on nil writer, got %v", err)
	}
	if err := ResetTerminalTitle(nil); err != nil {
		t.Errorf("expected nil error on nil writer, got %v", err)
	}
}

func TestTerminalTitle_Format(t *testing.T) {
	tests := []struct {
		name      string
		agent     string
		profile   string
		sessionID string
		expected  string
	}{
		{
			name:      "agent and profile with session ID",
			agent:     "codex",
			profile:   "work",
			sessionID: "1234abcd",
			expected:  "AIM: [codex] work (1234abcd)",
		},
		{
			name:      "agent and profile without session ID",
			agent:     "agy",
			profile:   "default",
			sessionID: "",
			expected:  "AIM: [agy] default",
		},
		{
			name:      "only profile with session ID",
			agent:     "",
			profile:   "work",
			sessionID: "sess-1",
			expected:  "AIM: work (sess-1)",
		},
		{
			name:      "only agent without session ID",
			agent:     "codex",
			profile:   "",
			sessionID: "",
			expected:  "AIM: [codex]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTitle(tt.agent, tt.profile, tt.sessionID)
			if got != tt.expected {
				t.Errorf("FormatTitle(%q, %q, %q) = %q; want %q",
					tt.agent, tt.profile, tt.sessionID, got, tt.expected)
			}
		})
	}
}

func TestTerminalTitle_ExtractSessionID(t *testing.T) {
	tests := []struct {
		args     []string
		expected string
	}{
		{args: []string{"resume", "sess-123"}, expected: "sess-123"},
		{args: []string{"--conversation=conv-456"}, expected: "conv-456"},
		{args: []string{"--conversation", "conv-789"}, expected: "conv-789"},
		{args: []string{"-c=conv-abc"}, expected: "conv-abc"},
		{args: []string{"-c", "conv-def"}, expected: "conv-def"},
		{args: []string{"--resume", "sess-claude"}, expected: "sess-claude"},
		{args: []string{"--resume=sess-claude2"}, expected: "sess-claude2"},
		{args: []string{"-r", "sess-claude-r"}, expected: "sess-claude-r"},
		{args: []string{"-r=sess-claude-req"}, expected: "sess-claude-req"},
		{args: []string{"--session-id", "sess-sid"}, expected: "sess-sid"},
		{args: []string{"--session-id=sess-sid-eq"}, expected: "sess-sid-eq"},
		{args: []string{"--resume"}, expected: ""},
		{args: []string{"-r"}, expected: ""},
		{args: []string{"--session-id"}, expected: ""},
		{args: []string{"--resume", "--flag"}, expected: ""},
		{args: []string{"-r", "-v"}, expected: ""},
		{args: []string{"--session-id", "--other"}, expected: ""},
		{args: []string{"run", "something"}, expected: ""},
		{args: []string{"resume", "--flag"}, expected: ""},
		{args: []string{"resume=sess-123"}, expected: ""},
	}

	for _, tt := range tests {
		got := ExtractSessionID(tt.args)
		if got != tt.expected {
			t.Errorf("ExtractSessionID(%v) = %q; want %q", tt.args, got, tt.expected)
		}
	}
}

func TestReplaceSessionID(t *testing.T) {
	tests := []struct {
		name, oldID, newID string
		args, want         []string
	}{
		{"separate value", "old", "new", []string{"--resume", "old"}, []string{"--resume", "new"}},
		{"attached value", "old", "new", []string{"--conversation=old"}, []string{"--conversation=new"}},
		{"first matching value", "old", "new", []string{"--resume", "other", "-r=old", "--session-id", "old"}, []string{"--resume", "other", "-r=new", "--session-id", "old"}},
		{"unrelated args", "old", "new", []string{"--flag", "old", "--resume", "other"}, []string{"--flag", "old", "--resume", "other"}},
		{"flag value ignored", "old", "new", []string{"--resume", "--flag"}, []string{"--resume", "--flag"}},
		{"no mutation when IDs invalid", "", "new", []string{"--resume", "old"}, []string{"--resume", "old"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := slices.Clone(tt.args)
			got := ReplaceSessionID(tt.args, tt.oldID, tt.newID)
			if !slices.Equal(tt.args, original) {
				t.Errorf("ReplaceSessionID mutated input: got %v; want %v", tt.args, original)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("ReplaceSessionID(%v) = %v; want %v", tt.args, got, tt.want)
			}
			if tt.oldID == "" && len(tt.args) > 0 && &got[0] != &tt.args[0] {
				t.Error("expected unchanged arguments to retain their original slice")
			}
		})
	}
}

func TestBuildEnv_AIMSessionID(t *testing.T) {
	environ := []string{
		"PATH=/bin:/usr/bin",
		"AIM_SESSION_ID=host-session-should-be-filtered",
		"USER=testuser",
	}

	// Case 1: launchEnv does NOT have AIM_SESSION_ID -> host variable should be filtered out
	res1 := BuildEnv(environ, map[string]string{})
	for _, e := range res1 {
		if e == "AIM_SESSION_ID=host-session-should-be-filtered" {
			t.Error("host AIM_SESSION_ID was not filtered from environ when launchEnv is empty")
		}
	}

	// Case 2: launchEnv has explicit AIM_SESSION_ID -> should be included, host version skipped
	launchEnv := map[string]string{
		"AIM_SESSION_ID": "explicit-session-id",
	}

	res2 := BuildEnv(environ, launchEnv)
	var foundHostSess, foundExplicitSess bool
	for _, e := range res2 {
		if e == "AIM_SESSION_ID=host-session-should-be-filtered" {
			foundHostSess = true
		}
		if e == "AIM_SESSION_ID=explicit-session-id" {
			foundExplicitSess = true
		}
	}
	if foundHostSess {
		t.Error("host AIM_SESSION_ID was not filtered from environ")
	}
	if !foundExplicitSess {
		t.Error("explicit AIM_SESSION_ID in launchEnv was not included")
	}
}

func TestTerminalTitle_Sanitize(t *testing.T) {
	var buf bytes.Buffer
	err := SetTerminalTitle(&buf, "AIM: [codex]\r\nwork\007evil\033hacked\007")
	if err != nil {
		t.Fatalf("SetTerminalTitle failed: %v", err)
	}
	expected := "\033]0;AIM: [codex]workevilhacked\007"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestTerminalTitle_NonTerminalFileIgnored(t *testing.T) {
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	defer rPipe.Close()

	// wPipe is a pipe, so isatty is false by default
	err = SetTerminalTitle(wPipe, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = ResetTerminalTitle(wPipe)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = wPipe.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(rPipe)
	if buf.Len() != 0 {
		t.Errorf("expected no output for non-terminal file, got %q", buf.String())
	}
}

func TestRunnerRun_TitleSetAndReset(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	origIsTerminal := isTerminalFunc
	isTerminalFunc = func(fd uintptr) bool { return true }
	defer func() { isTerminalFunc = origIsTerminal }()

	r := NewRunner()
	env := agents.LaunchEnv{
		BinaryPath: sh,
		Env: map[string]string{
			"AIM_AGENT":      "codex",
			"AIM_PROFILE":    "work",
			"AIM_SESSION_ID": "sess-42",
		},
	}

	oldStdout := os.Stdout
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = wPipe
	defer func() { os.Stdout = oldStdout }()

	code, err := r.Run(context.Background(), env, []string{"-c", "exit 0"})

	_ = wPipe.Close()

	if err != nil || code != 0 {
		t.Fatalf("Run failed: %v (code %d)", err, code)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(rPipe)
	_ = rPipe.Close()

	out := buf.String()
	expectedSet := "\033]0;AIM: [codex] work (sess-42)\007"
	expectedReset := "\033]0;\007"

	if !strings.Contains(out, expectedSet) {
		t.Errorf("stdout does not contain title set sequence: %q (got %q)", expectedSet, out)
	}
	if !strings.Contains(out, expectedReset) {
		t.Errorf("stdout does not contain title reset sequence: %q (got %q)", expectedReset, out)
	}
}
