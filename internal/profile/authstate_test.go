package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLogoutMarkerLifecycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		agent string
	}{
		{name: "claude", agent: "claude"},
		{name: "codex", agent: "codex"},
		{name: "agy", agent: "agy"},
		{name: "gemini", agent: "gemini"},
		{name: "cursor", agent: "cursor"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			profDir := t.TempDir()

			if IsAgentLoggedOut(profDir, tc.agent) {
				t.Fatalf("new profile should not start logged out")
			}
			if err := MarkAgentLoggedOut(profDir, tc.agent); err != nil {
				t.Fatalf("MarkAgentLoggedOut failed: %v", err)
			}
			if !IsAgentLoggedOut(profDir, tc.agent) {
				t.Fatalf("expected IsAgentLoggedOut=true after marking")
			}

			fi, err := os.Stat(LogoutMarkerPath(profDir, tc.agent))
			if err != nil {
				t.Fatalf("stat marker: %v", err)
			}
			if fi.Mode().Perm() != 0o600 {
				t.Errorf("marker mode = %#o, want 0600", fi.Mode().Perm())
			}

			if err := ClearAgentLoggedOut(profDir, tc.agent); err != nil {
				t.Fatalf("ClearAgentLoggedOut failed: %v", err)
			}
			if IsAgentLoggedOut(profDir, tc.agent) {
				t.Fatalf("expected IsAgentLoggedOut=false after clearing")
			}
		})
	}
}

func TestPurgeLocalAgentCredentials_Isolated(t *testing.T) {
	t.Parallel()

	p1 := t.TempDir()
	p2 := t.TempDir()

	writeFile := func(path string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(`{"token":"secret"}`), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	p1Creds := filepath.Join(p1, ".codex", "auth.json")
	p2Creds := filepath.Join(p2, ".codex", "auth.json")
	writeFile(p1Creds)
	writeFile(p2Creds)

	if err := PurgeLocalAgentCredentials(p1, "codex"); err != nil {
		t.Fatalf("PurgeLocalAgentCredentials failed: %v", err)
	}
	if _, err := os.Stat(p1Creds); !os.IsNotExist(err) {
		t.Errorf("expected p1 credentials removed, got %v", err)
	}
	if _, err := os.Stat(p2Creds); err != nil {
		t.Errorf("expected p2 credentials untouched, got %v", err)
	}
}
