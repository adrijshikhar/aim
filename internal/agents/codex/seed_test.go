package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyHostConfig_SeedsWithoutServersOrPlugins(t *testing.T) {
	home := t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, ".codex"), 0o700)
	_ = os.WriteFile(filepath.Join(home, ".codex", "config.toml"),
		[]byte("model = \"x\"\n\n[mcp_servers.a]\ncommand = \"t\"\n\n[plugins.\"p@q\"]\nenabled = true\n\n[plugins.\"hevo@hevo\".mcp_servers.atlassian]\nurl = \"https://x\"\n\n[hooks.state.\"p@q:stop\"]\nenabled = false\n"), 0o600)
	prof := t.TempDir()
	copyHostConfig(home, prof)
	b, _ := os.ReadFile(filepath.Join(prof, "config.toml"))
	s := string(b)
	if strings.Contains(s, "[mcp_servers.a]") || !strings.Contains(s, `model = "x"`) {
		t.Fatalf("host servers must not be seeded:\n%s", s)
	}
	// Host plugin enablement reaches sessions through the merge too (G5); the
	// per-plugin hook state is not part of it and is seeded as before.
	if strings.Contains(s, "[plugins.") || !strings.Contains(s, `[hooks.state."p@q:stop"]`) {
		t.Fatalf("plugin tables must not be seeded, hook state must:\n%s", s)
	}
	if left, _ := filepath.Glob(filepath.Join(prof, ".aim-seed-*")); len(left) != 0 {
		t.Fatalf("temp files left behind: %v", left)
	}
}
