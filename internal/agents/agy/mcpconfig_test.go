package agy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegularizeMCPConfig_SymlinkBecomesFile(t *testing.T) {
	shared := t.TempDir()
	prof := t.TempDir()
	_ = os.WriteFile(filepath.Join(shared, "mcp_config.json"), []byte(`{"mcpServers":{"a":{}}}`), 0o644)
	_ = os.Symlink(filepath.Join(shared, "mcp_config.json"), filepath.Join(prof, "mcp_config.json"))
	if err := regularizeMCPConfig(prof); err != nil {
		t.Fatal(err)
	}
	fi, _ := os.Lstat(filepath.Join(prof, "mcp_config.json"))
	if fi.Mode()&os.ModeSymlink != 0 || fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", fi.Mode())
	}
	if b, _ := os.ReadFile(filepath.Join(prof, "mcp_config.json")); string(b) != `{"mcpServers":{"a":{}}}` {
		t.Fatalf("copy = %s", b)
	}
	if b, _ := os.ReadFile(filepath.Join(shared, "mcp_config.json")); string(b) != `{"mcpServers":{"a":{}}}` {
		t.Fatal("the link target must not change")
	}
}

func TestRegularizeMCPConfig_DanglingAndEmptyBecomeEmptyMap(t *testing.T) {
	for name, setup := range map[string]func(p string){
		"dangling": func(p string) { _ = os.Symlink("/nonexistent/x.json", p) },
		"empty":    func(p string) { _ = os.WriteFile(p, nil, 0o644) },
		"absent":   func(p string) {},
	} {
		prof := t.TempDir()
		p := filepath.Join(prof, "mcp_config.json")
		setup(p)
		if err := regularizeMCPConfig(prof); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if b, _ := os.ReadFile(p); string(b) != `{"mcpServers":{}}` {
			t.Fatalf("%s: got %q", name, b)
		}
	}
}

func TestBridgeSharedState_RegularMCPConfigSurvives(t *testing.T) {
	// T6 regression: a profile's own regular mcp_config.json is never deleted
	// or re-linked by bridging.
	t.Setenv("AIM_HOME", t.TempDir())
	realHome := t.TempDir()
	hostCfg := filepath.Join(realHome, ".gemini", "config")
	_ = os.MkdirAll(hostCfg, 0o755)
	_ = os.WriteFile(filepath.Join(hostCfg, "mcp_config.json"), []byte(`{"mcpServers":{"host":{"command":"h"}}}`), 0o644)
	profileDir := t.TempDir()
	profCfg := filepath.Join(profileDir, ".gemini", "config")
	_ = os.MkdirAll(profCfg, 0o755)
	own := filepath.Join(profCfg, "mcp_config.json")
	_ = os.WriteFile(own, []byte(`{"mcpServers":{"own":{"command":"o"}}}`), 0o600)
	for i := 0; i < 2; i++ {
		if err := bridgeSharedState(realHome, profileDir); err != nil {
			t.Fatal(err)
		}
	}
	fi, err := os.Lstat(own)
	if err != nil || fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("profile mcp_config.json must stay a regular file: %v %v", fi, err)
	}
	if b, _ := os.ReadFile(own); !strings.Contains(string(b), `"own"`) || strings.Contains(string(b), `"host"`) {
		t.Fatalf("profile servers lost or host copied in: %s", b)
	}
}
