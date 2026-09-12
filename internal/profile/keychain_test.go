package profile

import (
	"runtime"
	"testing"
)

func TestGetIgnoredKeychainEntries(t *testing.T) {
	entries := GetIgnoredKeychainEntries()
	if len(entries) < 10 {
		t.Errorf("expected at least 10 default ignored keychain entries, got %d", len(entries))
	}

	hasAgy := false
	hasClaude := false
	hasCodex := false

	for _, e := range entries {
		if e.Agent == "agy" && e.Service == "gemini" && e.Account == "antigravity" {
			hasAgy = true
		}
		if e.Agent == "claude" && e.Service == "claude-code" {
			hasClaude = true
		}
		if e.Agent == "codex" && e.Service == "codex" {
			hasCodex = true
		}
	}

	if !hasAgy {
		t.Errorf("missing default entry for agy (gemini/antigravity)")
	}
	if !hasClaude {
		t.Errorf("missing default entry for claude (claude-code)")
	}
	if !hasCodex {
		t.Errorf("missing default entry for codex (codex)")
	}

	// Test with custom services
	custom := GetIgnoredKeychainEntries("custom-tool", "another-agent")
	if len(custom) != len(entries)+2 {
		t.Errorf("expected %d entries with custom, got %d", len(entries)+2, len(custom))
	}
}

func TestPurgeIgnoredKeychains_NonDarwinOrMissing(t *testing.T) {
	// Calling PurgeIgnoredKeychains should not error out even if items do not exist
	err := PurgeIgnoredKeychains("non-existent-agent")
	if err != nil {
		t.Errorf("unexpected error purging non-existent agent keychains: %v", err)
	}

	// Purge all should succeed (items missing will return exit code 44, handled cleanly)
	err = PurgeIgnoredKeychains("")
	if err != nil {
		t.Errorf("unexpected error purging keychains: %v", err)
	}
}

func TestFindIgnoredKeychains_NonDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		found := FindIgnoredKeychains("")
		if found != nil {
			t.Errorf("expected nil on non-darwin OS, got %v", found)
		}
	}
}
