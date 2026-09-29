package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/logger"
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

// A key the splicer refuses (an inline plugins table) must not undo the strip
// of the keys before it: the seed would then carry every host server, secrets
// and all.
func TestCopyHostConfig_RefusedKeyKeepsEarlierStrips(t *testing.T) {
	var warned strings.Builder
	logger.SetWarnOutput(&warned)
	defer logger.Reset()
	home := t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, ".codex"), 0o700)
	_ = os.WriteFile(filepath.Join(home, ".codex", "config.toml"),
		[]byte("model = \"x\"\nplugins = { \"p@q\" = { enabled = true } }\n\n[mcp_servers.a]\ncommand = \"t\"\nenv = { TOKEN = \"secret\" }\n"), 0o600)
	prof := t.TempDir()
	copyHostConfig(home, prof)
	b, _ := os.ReadFile(filepath.Join(prof, "config.toml"))
	s := string(b)
	if strings.Contains(s, "mcp_servers") || strings.Contains(s, "secret") || !strings.Contains(s, `model = "x"`) {
		t.Fatalf("the mcp_servers strip must survive a later refusal:\n%s", s)
	}
	if !strings.Contains(warned.String(), "plugins") {
		t.Fatalf("the refused key must be a visible warning, got %q", warned.String())
	}
}
