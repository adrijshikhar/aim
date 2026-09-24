package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/agents/gemini"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/runner"
)

func TestAdapterEnvironmentIsolation(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIM_REAL_HOME", root)
	t.Setenv("AIM_HOME", filepath.Join(root, "aim"))
	t.Setenv("AIM_TEST_INHERITED", "keep-me")
	// PrepareEnv can inspect the macOS keychain; never consult host credentials.
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "security"), []byte("#!/bin/sh\nexit 44\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	managed := []string{
		"SSH_CONNECTION", "SSH_CLIENT", "SSH_TTY", "AIM_AGENT", "AIM_PROFILE", "AIM_SESSION_ID",
		"GEMINI_CLI_HOME", "CODEX_HOME", "CLAUDE_CONFIG_DIR",
		"CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_REFRESH_TOKEN",
	}
	for _, key := range managed {
		t.Setenv(key, "host-value")
	}
	for _, adapter := range []agents.AgentAdapter{agy.NewAdapter(), codex.NewAdapter(), claude.NewAdapter(), gemini.NewAdapter()} {
		t.Run(adapter.Name(), func(t *testing.T) {
			profileDir := filepath.Join(root, "profiles", adapter.Name())
			if err := os.MkdirAll(profileDir, 0700); err != nil {
				t.Fatal(err)
			}
			launch, err := adapter.PrepareEnv("work", profileDir)
			if err != nil {
				t.Fatal(err)
			}
			env := make(map[string]string)
			for _, entry := range runner.BuildEnv(os.Environ(), launch.Env) {
				key, value, _ := strings.Cut(entry, "=")
				env[key] = value
			}
			if env["AIM_TEST_INHERITED"] != "keep-me" || env["HOME"] != profileDir || env["AIM_AGENT"] != adapter.Name() || env["AIM_PROFILE"] != "work" {
				t.Fatal("benign inheritance or explicit profile identity was lost")
			}
			if env["AIM_HOME"] != os.Getenv("AIM_HOME") {
				t.Errorf("AIM_HOME = %q; want parent profile store %q", env["AIM_HOME"], os.Getenv("AIM_HOME"))
			}
			for _, key := range managed {
				if env[key] == "host-value" {
					t.Errorf("host %s leaked through adapter and runner", key)
				}
			}
			launch.Env["AIM_SESSION_ID"] = "explicit-session"
			if !slices.Contains(runner.BuildEnv(os.Environ(), launch.Env), "AIM_SESSION_ID=explicit-session") {
				t.Error("explicit session override was lost")
			}
		})
	}
}

// Login bypasses Runner.Run, so verify its native subprocess environment too.
// Provider binaries and Keychain are stubbed; this does not authenticate.
func TestAdapterLoginStorage(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIM_HOME", "")
	t.Setenv("AIM_REAL_HOME", root)
	for _, kind := range []string{"CONFIG", "DATA", "CACHE", "STATE"} {
		t.Setenv("AIM_"+kind+"_DIR", "")
		t.Setenv("XDG_"+kind+"_HOME", filepath.Join(root, kind))
	}
	t.Setenv("AIM_AGY_CLIENT_ID", "")
	t.Setenv("AIM_AGY_CLIENT_SECRET", "")
	clientID, clientSecret := agy.DefaultClientID, agy.DefaultClientSecret
	agy.DefaultClientID, agy.DefaultClientSecret = "", ""
	t.Cleanup(func() { agy.DefaultClientID, agy.DefaultClientSecret = clientID, clientSecret })
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "security"), []byte("#!/bin/sh\nexit 44\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for key, value := range config.StorageEnv() {
		t.Setenv("EXPECTED_"+key, value)
	}
	probe := `#!/bin/sh
set -eu
[ "$AIM_HOME" = "$EXPECTED_AIM_HOME" ]
[ "$AIM_REAL_HOME" = "$EXPECTED_AIM_REAL_HOME" ]
[ "$AIM_CONFIG_DIR" = "$EXPECTED_AIM_CONFIG_DIR" ]
[ "$AIM_DATA_DIR" = "$EXPECTED_AIM_DATA_DIR" ]
[ "$AIM_CACHE_DIR" = "$EXPECTED_AIM_CACHE_DIR" ]
[ "$AIM_STATE_DIR" = "$EXPECTED_AIM_STATE_DIR" ]
`
	for _, adapter := range []agents.AgentAdapter{agy.NewAdapter(), codex.NewAdapter(), claude.NewAdapter()} {
		t.Run(adapter.Name(), func(t *testing.T) {
			if err := os.WriteFile(filepath.Join(bin, adapter.BinaryName()), []byte(probe), 0700); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(root, "profiles", adapter.Name())
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := adapter.Login(context.Background(), "work", dir); err != nil {
				t.Fatal(err)
			}
		})
	}
}
