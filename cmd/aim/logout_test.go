package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/profile"
)

func TestLogoutCmd_IsolatesSiblingProfilesAndScopedKeychain(t *testing.T) {
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	baseDir := t.TempDir()
	t.Setenv("AIM_HOME", baseDir)
	realHome := t.TempDir()
	t.Setenv("AIM_REAL_HOME", realHome)

	pm := profile.NewProfileManager(baseDir)
	workDir, err := pm.EnsureProfile("work")
	if err != nil {
		t.Fatal(err)
	}
	otherDir, err := pm.EnsureProfile("other")
	if err != nil {
		t.Fatal(err)
	}

	workCreds := writeTestClaudeCreds(t, workDir)
	otherCreds := writeTestClaudeCreds(t, otherDir)

	var deletedServices []string
	restore := profile.SetKeychainPasswordDeleterForTest(func(service, _ string) error {
		deletedServices = append(deletedServices, service)
		return nil
	})
	defer restore()

	reg := agents.NewRegistry()
	reg.Register(claude.NewAdapter())
	reg.Register(codex.NewAdapter())

	cmd := newRootCmd(reg, pm)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"logout", "claude", "work"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("logout failed: %v (stderr: %s)", err, errOut.String())
	}
	if !strings.Contains(out.String(), "Logged out") {
		t.Errorf("unexpected stdout: %q", out.String())
	}

	// Only work's scoped keychain service should be deleted (never host or other profile's).
	assertDeletedOnlyService(t, deletedServices,
		profile.ClaudeScopedKeychainService(filepath.Join(workDir, ".claude")),
		profile.ClaudeScopedKeychainService(filepath.Join(otherDir, ".claude")),
	)

	// work creds should be gone & marked logged out; other creds untouched.
	if _, err := os.Stat(workCreds); !os.IsNotExist(err) {
		t.Errorf("expected work credentials removed, got %v", err)
	}
	if !profile.IsAgentLoggedOut(workDir, "claude") {
		t.Errorf("expected work profile to be marked logged out")
	}
	if _, err := os.Stat(otherCreds); err != nil {
		t.Errorf("sibling profile credentials altered: %v", err)
	}
	if profile.IsAgentLoggedOut(otherDir, "claude") {
		t.Errorf("sibling profile must not be marked logged out")
	}
}

func writeTestClaudeCreds(t *testing.T, profDir string) string {
	t.Helper()
	p := filepath.Join(profDir, ".claude", ".credentials.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(`{"claudeAiOauth":{"accessToken":"tok","refreshToken":"ref"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func assertDeletedOnlyService(t *testing.T, deleted []string, want, forbiddenOther string) {
	t.Helper()
	if len(deleted) != 1 || deleted[0] != want {
		t.Errorf("deletedServices = %v, want [%s]", deleted, want)
	}
	for _, svc := range deleted {
		if svc == forbiddenOther || svc == "Claude Code-credentials" || svc == "Claude Safe Storage" {
			t.Errorf("logout touched host/sibling keychain service %q", svc)
		}
	}
}

func TestLogoutCmd_RejectsUnknownProfileAndInvalidArgs(t *testing.T) {
	baseDir := t.TempDir()
	t.Setenv("AIM_HOME", baseDir)
	reg := agents.NewRegistry()
	reg.Register(codex.NewAdapter())
	pm := profile.NewProfileManager(baseDir)

	cmd := newRootCmd(reg, pm)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"logout", "codex", "missing"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error logging out non-existent profile")
	}
}

func TestLogoutCmd_PreventsHostReseedOnPrimaryProfile(t *testing.T) {
	baseDir := t.TempDir()
	t.Setenv("AIM_HOME", baseDir)
	realHome := t.TempDir()
	t.Setenv("AIM_REAL_HOME", realHome)

	// Host has ~/.codex/auth.json
	hostCodex := filepath.Join(realHome, ".codex", "auth.json")
	if err := os.MkdirAll(filepath.Dir(hostCodex), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hostCodex, []byte(`{"tokens":{"access_token":"host-tok"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	pm := profile.NewProfileManager(baseDir)
	personalDir, err := pm.EnsureProfile("personal")
	if err != nil {
		t.Fatal(err)
	}
	personalAuth := filepath.Join(personalDir, ".codex", "auth.json")
	if err := os.MkdirAll(filepath.Dir(personalAuth), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(personalAuth, []byte(`{"tokens":{"access_token":"prof-tok"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	reg := agents.NewRegistry()
	codexAd := codex.NewAdapter()
	reg.Register(codexAd)

	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"logout", "codex", "personal"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	if _, err := codexAd.PrepareEnv(context.Background(), "personal", personalDir); err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}
	if codexAd.HasCredentials(personalDir) {
		t.Fatalf("logged-out personal profile silently re-seeded host credentials")
	}
}
