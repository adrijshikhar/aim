package claude

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolatedHome gives the adapter a fake real home, aim dirs under it, and a
// `security` binary that always fails, as TestAdapter_PrepareEnv does, so no
// test touches ~/.aim, ~/.claude or the login keychain.
func isolatedHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("AIM_HOME", filepath.Join(home, ".aim"))
	t.Setenv("AIM_REAL_HOME", home)
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	for _, kind := range []string{"CONFIG", "DATA", "CACHE", "STATE"} {
		t.Setenv("AIM_"+kind+"_DIR", "")
		t.Setenv("XDG_"+kind+"_HOME", filepath.Join(home, kind))
	}
	bin := filepath.Join(home, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "security"), []byte("#!/bin/sh\nexit 44\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return home
}

func TestPrepareEnvAndDoctor_NeverWriteRootClaudeJSON(t *testing.T) {
	home := isolatedHome(t)
	_ = os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"oauthAccount":{"email":"host"},"mcpServers":{}}`), 0o600)
	p := filepath.Join(home, "profiles", "x")
	_ = os.MkdirAll(p, 0o700)
	root := filepath.Join(p, ".claude.json")
	_ = os.WriteFile(root, []byte(`{"machineID":"own"}`), 0o600)
	a := &Adapter{}
	if _, err := a.PrepareEnv("x", p); err != nil {
		t.Fatal(err)
	}
	_ = a.Doctor(context.Background(), "x", p)
	if b, _ := os.ReadFile(root); string(b) != `{"machineID":"own"}` {
		t.Fatalf("root .claude.json overwritten: %s", b)
	}
	fresh := filepath.Join(home, "profiles", "fresh")
	_ = os.MkdirAll(fresh, 0o700)
	if _, err := a.PrepareEnv("fresh", fresh); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(fresh, ".claude.json")); !os.IsNotExist(err) {
		t.Fatal("PrepareEnv must not create <p>/.claude.json from the host file")
	}
}

func TestPrepareEnv_SeedsProfileClaudeJSON(t *testing.T) {
	home := isolatedHome(t)
	hostConfig := `{
  "oauthAccount": {"email": "host@example.com", "accountUuid": "acc-1"},
  "mcpServers": {"host-mcp": {"command": "npx"}},
  "hasCompletedOnboarding": true,
  "lastOnboardingVersion": "2.1.286",
  "theme": "dark",
  "projects": {
    "/Users/test/project": {
      "hasTrustDialogAccepted": true,
      "hasClaudeMdExternalIncludesApproved": true
    }
  }
}`
	_ = os.WriteFile(filepath.Join(home, ".claude.json"), []byte(hostConfig), 0o600)

	p := filepath.Join(home, "profiles", "dev")
	_ = os.MkdirAll(p, 0o700)
	a := &Adapter{}
	if _, err := a.PrepareEnv("dev", p); err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	dest := filepath.Join(p, ".claude", ".claude.json")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("expected %s to be created: %v", dest, err)
	}

	content := string(data)
	if strings.Contains(content, "oauthAccount") || strings.Contains(content, "host@example.com") {
		t.Errorf("seeded config must NOT contain oauthAccount: %s", content)
	}
	if strings.Contains(content, "mcpServers") || strings.Contains(content, "host-mcp") {
		t.Errorf("seeded config must NOT contain static mcpServers: %s", content)
	}
	if !strings.Contains(content, `"hasCompletedOnboarding":true`) && !strings.Contains(content, `"hasCompletedOnboarding": true`) {
		t.Errorf("expected hasCompletedOnboarding to be preserved: %s", content)
	}
	if !strings.Contains(content, "hasTrustDialogAccepted") || !strings.Contains(content, "hasClaudeMdExternalIncludesApproved") {
		t.Errorf("expected projects trust to be preserved: %s", content)
	}

	// Verify permissions are 0600
	fi, err := os.Stat(dest)
	if err != nil || fi.Mode().Perm() != 0o600 {
		t.Errorf("expected permissions 0600, got %v (err %v)", fi.Mode().Perm(), err)
	}

	// Verify existing .claude/.claude.json is NEVER overwritten
	customContent := `{"custom":"profile-value"}`
	_ = os.WriteFile(dest, []byte(customContent), 0o600)
	if _, err := a.PrepareEnv("dev", p); err != nil {
		t.Fatalf("second PrepareEnv failed: %v", err)
	}
	afterData, _ := os.ReadFile(dest)
	if string(afterData) != customContent {
		t.Errorf("existing profile config was overwritten: got %s, want %s", string(afterData), customContent)
	}
}
