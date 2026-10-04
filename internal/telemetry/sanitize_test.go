package telemetry

import (
	"strings"
	"testing"
	"time"
)

func TestDurationBucket(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{100 * time.Millisecond, "<500ms"},
		{499 * time.Millisecond, "<500ms"},
		{500 * time.Millisecond, "500ms-2s"},
		{1500 * time.Millisecond, "500ms-2s"},
		{2 * time.Second, "2s-10s"},
		{9999 * time.Millisecond, "2s-10s"},
		{10 * time.Second, ">10s"},
		{30 * time.Second, ">10s"},
	}

	for _, tt := range tests {
		got := DurationBucket(tt.d)
		if got != tt.want {
			t.Errorf("DurationBucket(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestSanitizeCommand_NoProfileNamesOrPathsOrFlags(t *testing.T) {
	tests := []struct {
		rawCmd    string
		rawArgs   []string
		wantCmd   string
		wantAgent string
	}{
		{"run", []string{"agy", "work-client-corp", "--prompt", "my secret prompt"}, "run", "agy"},
		{"login", []string{"claude", "personal@gmail.com"}, "login", "claude"},
		{"doctor", []string{"codex"}, "doctor", "codex"},
		{"list", []string{"--json"}, "list", ""},
		{"sessions", []string{"show", "019485..."}, "sessions", ""},
		{"feedback", []string{"-m", "secret message"}, "feedback", ""},
		{"unknown-custom", []string{"foo"}, "unknown-custom", ""},
	}

	for _, tt := range tests {
		cleanCmd, cleanAgent := SanitizeCommand(tt.rawCmd, tt.rawArgs)
		if cleanCmd != tt.wantCmd {
			t.Errorf("SanitizeCommand(%q, %v) cmd = %q, want %q", tt.rawCmd, tt.rawArgs, cleanCmd, tt.wantCmd)
		}
		if cleanAgent != tt.wantAgent {
			t.Errorf("SanitizeCommand(%q, %v) agent = %q, want %q", tt.rawCmd, tt.rawArgs, cleanAgent, tt.wantAgent)
		}
	}
}

func TestMachineID_DeterministicAndCached(t *testing.T) {
	tempDir := t.TempDir()
	id1 := AnonymousMachineID(tempDir)
	id2 := AnonymousMachineID(tempDir)

	if id1 == "" {
		t.Fatal("expected non-empty machine ID")
	}
	if id1 != id2 {
		t.Errorf("expected deterministic machine ID across calls, got %q vs %q", id1, id2)
	}
	if len(id1) != 64 { // SHA256 hex string
		t.Errorf("expected 64-char sha256 hex string, got %d chars: %q", len(id1), id1)
	}
	if strings.Contains(id1, " ") {
		t.Errorf("unexpected spaces in machine ID: %q", id1)
	}
}
