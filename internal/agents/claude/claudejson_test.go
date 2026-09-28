package claude

import (
	"context"
	"os"
	"path/filepath"
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
