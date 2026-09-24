package runner_test

import (
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
