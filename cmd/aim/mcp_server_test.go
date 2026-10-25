package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

func TestMCPCmd_Help(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "--help"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, sub := range []string{"serve", "doctor", "install", "uninstall", "list", "add", "remove"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected %q subcommand in help output: %s", sub, out)
		}
	}
}

func TestMCPServeCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"ping"}` + "\n"

	inBuf := bytes.NewBufferString(input)
	outBuf := new(bytes.Buffer)

	cmd := newRootCmd(reg, pm)
	cmd.SetIn(inBuf)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"mcp", "serve", "--profile", "work", "--agent", "claude"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("mcp serve failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 responses, got %d: %s", len(lines), outBuf.String())
	}

	var initResp struct {
		Result struct {
			ServerInfo struct {
				Name string `json:"name"`
			} `json:"serverInfo"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &initResp); err != nil {
		t.Fatalf("failed to decode init response: %v", err)
	}
	if initResp.Result.ServerInfo.Name != "aim" {
		t.Errorf("expected serverInfo.name 'aim', got %q", initResp.Result.ServerInfo.Name)
	}
}

func TestMCPInstallCmd_ClaudeGlobal(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "install", "claude"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Installed aim MCP server") {
		t.Errorf("expected installation confirmation in output: %s", out)
	}

	claudePath := filepath.Join(tempDir, ".claude.json")
	data, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read .claude.json: %v", err)
	}

	var parsed struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal .claude.json: %v", err)
	}

	srv, ok := parsed.MCPServers["aim"]
	if !ok {
		t.Fatalf("expected 'aim' entry in mcpServers, got: %s", string(data))
	}
	if srv.Command != "aim" || len(srv.Args) != 2 || srv.Args[0] != "mcp" || srv.Args[1] != "serve" {
		t.Errorf("expected command 'aim' with args ['mcp', 'serve'], got command %q args %v", srv.Command, srv.Args)
	}
}

func TestMCPInstallCmd_ClaudeProfile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "install", "claude", "--profile", "work"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Installed aim MCP server") {
		t.Errorf("expected confirmation in output: %s", out)
	}

	claudePath := filepath.Join(pm.ProfileDir("work"), ".claude", ".claude.json")
	data, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read profile claude json at %s: %v", claudePath, err)
	}

	var parsed struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal profile claude json: %v", err)
	}

	if _, ok := parsed.MCPServers["aim"]; !ok {
		t.Errorf("expected 'aim' entry in mcpServers: %s", string(data))
	}
}

func TestMCPInstallCmd_OtherAgent(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "install", "cursor"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Auto-install is currently supported for 'claude', 'codex', 'agy'") {
		t.Errorf("expected guidance for non-supported agent, got: %s", out)
	}
}

func TestMCPServeCmd_WhoamiTool(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("production")

	input := `{"jsonrpc":"2.0","id":100,"method":"tools/call","params":{"name":"aim_whoami","arguments":{}}}` + "\n"

	inBuf := bytes.NewBufferString(input)
	outBuf := new(bytes.Buffer)

	cmd := newRootCmd(reg, pm)
	cmd.SetIn(inBuf)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"mcp", "serve", "--profile", "production", "--agent", "claude"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("mcp serve failed: %v", err)
	}

	var callResp struct {
		ID     int `json:"id"`
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &callResp); err != nil {
		t.Fatalf("failed to decode response: %v\noutput: %s", err, outBuf.String())
	}
	if callResp.ID != 100 {
		t.Errorf("expected ID 100, got %d", callResp.ID)
	}
	if len(callResp.Result.Content) == 0 {
		t.Fatalf("expected content block, got none")
	}

	text := callResp.Result.Content[0].Text
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatalf("expected JSON formatted text in content: %v", err)
	}
	if payload["profile"] != "production" {
		t.Errorf("expected profile 'production', got %v", payload["profile"])
	}
	if payload["agent"] != "claude" {
		t.Errorf("expected agent 'claude', got %v", payload["agent"])
	}
}

func TestMCPInstallCmd_ProfilePathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"mcp", "install", "claude", "--profile", "../../evil"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for path traversal in --profile, got nil")
	}
	if !strings.Contains(err.Error(), "path traversal") {
		t.Errorf("expected 'path traversal' in error, got: %v", err)
	}
}

func TestMCPInstallCmd_ProfileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"mcp", "install", "claude", "--profile", "nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent profile, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' in error, got: %v", err)
	}
}

func TestMCPServeCmd_InvalidProfile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"mcp", "serve", "--profile", "../../evil"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for path traversal in --profile, got nil")
	}
	if !strings.Contains(err.Error(), "path traversal") {
		t.Errorf("expected 'path traversal' in error, got: %v", err)
	}
}

func TestMCPServeCmd_NonexistentProfile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"mcp", "serve", "--profile", "nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent profile, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' in error, got: %v", err)
	}
}

func TestMCPServeCmd_NoArgsConstraint(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"mcp", "serve", "extra-arg"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when positional args provided to serve, got nil")
	}
}

func TestMCPInstallCmd_CorruptConfigRefusesOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	claudePath := filepath.Join(tempDir, ".claude.json")
	_ = os.WriteFile(claudePath, []byte(`{corrupt json`), 0600)

	cmd := newRootCmd(reg, pm)
	cmd.SetArgs([]string{"mcp", "install", "claude"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when .claude.json is corrupt, got nil")
	}

	content, _ := os.ReadFile(claudePath)
	if string(content) != "{corrupt json" {
		t.Errorf("file was modified despite corruption error: %s", string(content))
	}
}

func TestMCPDoctorCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "doctor"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Protocol:") {
		t.Errorf("expected Protocol section in doctor output: %s", out)
	}
	if !strings.Contains(out, "[OK]") {
		t.Errorf("expected passing status in doctor output: %s", out)
	}
}

func TestMCPInstallAndUninstallCmd_AdapterDriven(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	codexDir := filepath.Join(tempDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)

	// 1. Install to codex
	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "install", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp install codex failed: %v", err)
		}
	})

	if !strings.Contains(out, "Codex CLI") {
		t.Errorf("expected Codex confirmation: %s", out)
	}

	codexConfig := filepath.Join(codexDir, "config.toml")
	data, err := os.ReadFile(codexConfig)
	if err != nil {
		t.Fatalf("failed to read codex config: %v", err)
	}
	if !strings.Contains(string(data), `[mcp_servers.aim]`) {
		t.Errorf("expected [mcp_servers.aim] in codex config: %s", string(data))
	}

	// 2. Uninstall from codex
	out, _ = captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "uninstall", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp uninstall codex failed: %v", err)
		}
	})

	if !strings.Contains(out, "Removed aim") {
		t.Errorf("expected uninstall confirmation: %s", out)
	}

	data, _ = os.ReadFile(codexConfig)
	if strings.Contains(string(data), `[mcp_servers.aim]`) {
		t.Errorf("aim should have been removed from codex config: %s", string(data))
	}
}

func TestMCPListCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	claudeCfg := filepath.Join(tempDir, ".claude.json")
	_ = os.WriteFile(claudeCfg, []byte(`{"mcpServers":{"demo":{"command":"echo","args":["hello"]}}}`), 0600)

	// 1. Table output
	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "list"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp list failed: %v", err)
		}
	})

	if !strings.Contains(out, "demo") {
		t.Errorf("expected 'demo' in mcp list table: %s", out)
	}

	// 2. JSON output
	out, _ = captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "list", "--json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp list --json failed: %v", err)
		}
	})

	var parsed []map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("expected valid JSON array from mcp list --json: %v\nOutput: %s", err, out)
	}
	if len(parsed) == 0 {
		t.Errorf("expected at least 1 server in JSON output, got 0")
	}
}

func TestMCPListCmd_TruncatesLongCommand(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)
	t.Setenv("AIM_TERMINAL_WIDTH", "100")

	claudeCfg := filepath.Join(tempDir, ".claude.json")
	_ = os.WriteFile(claudeCfg, []byte(`{"mcpServers":{"wiremock":{"command":"java","args":["-cp","/very/long/path/to/some/dependency/wiremock-standalone.jar:/another/very/long/classpath/entry/here","wiremock.Run","--port","8080","--verbose","--debug"]}}}`), 0600)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "list"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp list failed: %v", err)
		}
	})

	if !strings.Contains(out, "wiremock") {
		t.Errorf("expected 'wiremock' in table output: %s", out)
	}
	if !strings.Contains(out, "…") && !strings.Contains(out, "...") {
		t.Errorf("expected truncated string with ellipsis for long command: %s", out)
	}
	if strings.Contains(out, "--debug") {
		t.Errorf("expected full command tail to be truncated away, but found in output: %s", out)
	}
}

func TestMCPAddAndRemoveCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	codexDir := filepath.Join(tempDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)

	// 1. Add server
	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "add", "fetcher", "uvx", "mcp-server-fetch", "--clients", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}
	})

	if !strings.Contains(out, "Added MCP server 'fetcher'") {
		t.Errorf("expected add confirmation: %s", out)
	}

	codexCfg := filepath.Join(codexDir, "config.toml")
	data, _ := os.ReadFile(codexCfg)
	if !strings.Contains(string(data), `[mcp_servers.fetcher]`) {
		t.Errorf("expected fetcher in codex config: %s", string(data))
	}

	// 2. Remove server
	out, _ = captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "remove", "fetcher", "--clients", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp remove failed: %v", err)
		}
	})

	if !strings.Contains(out, "Removed MCP server 'fetcher'") {
		t.Errorf("expected remove confirmation: %s", out)
	}

	data, _ = os.ReadFile(codexCfg)
	if strings.Contains(string(data), `fetcher`) {
		t.Errorf("fetcher should be removed from codex config: %s", string(data))
	}
}

func TestMCPAddCmd_AutoDetectInstalled(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("PATH", t.TempDir())

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	codexDir := filepath.Join(tempDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)
	codexCfg := filepath.Join(codexDir, "config.toml")
	_ = os.WriteFile(codexCfg, []byte("[mcp_servers]\n"), 0600)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "add", "my-server", "echo", "hello"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}
	})

	if !strings.Contains(out, "Added MCP server 'my-server' to Codex CLI") {
		t.Errorf("expected server added to detected Codex CLI, got: %s", out)
	}

	data, _ := os.ReadFile(codexCfg)
	if !strings.Contains(string(data), "[mcp_servers.my-server]") {
		t.Errorf("expected my-server in codex config: %s", string(data))
	}

	claudeCfg := filepath.Join(tempDir, ".claude.json")
	if _, err := os.Stat(claudeCfg); err == nil {
		t.Errorf("claude config should not have been created")
	}
}

func TestMCPAddCmd_MultiProfile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("staging")

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "add", "fetcher", "uvx", "mcp-server-fetch", "--profile", "work,staging", "--clients", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}
	})

	if !strings.Contains(out, "(profile work)") || !strings.Contains(out, "(profile staging)") {
		t.Errorf("expected output confirming addition to both profiles: %s", out)
	}

	for _, p := range []string{"work", "staging"} {
		cfgPath := filepath.Join(pm.ProfileDir(p), ".codex", "config.toml")
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatalf("failed to read profile %s codex config: %v", p, err)
		}
		if !strings.Contains(string(data), "[mcp_servers.fetcher]") {
			t.Errorf("profile %s missing fetcher server: %s", p, string(data))
		}
	}
}

func TestMCPAddCmd_AllProfiles(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("alpha")
	_, _ = pm.EnsureProfile("beta")

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "add", "shared-tool", "npx", "tool", "--all", "--clients", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add --all failed: %v", err)
		}
	})

	if !strings.Contains(out, "(profile alpha)") || !strings.Contains(out, "(profile beta)") {
		t.Errorf("expected output confirming addition to both profiles: %s", out)
	}

	for _, p := range []string{"alpha", "beta"} {
		cfgPath := filepath.Join(pm.ProfileDir(p), ".codex", "config.toml")
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatalf("failed to read profile %s codex config: %v", p, err)
		}
		if !strings.Contains(string(data), "[mcp_servers.shared-tool]") {
			t.Errorf("profile %s missing shared-tool: %s", p, string(data))
		}
	}
}

func TestMCPAddCmd_ExplicitClients(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "add", "codex-only", "echo", "target", "--clients", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add --clients codex failed: %v", err)
		}
	})

	if !strings.Contains(out, "Added MCP server 'codex-only' to Codex CLI") {
		t.Errorf("expected confirmation for Codex CLI, got: %s", out)
	}

	codexCfg := filepath.Join(tempDir, ".codex", "config.toml")
	data, err := os.ReadFile(codexCfg)
	if err != nil {
		t.Fatalf("failed to read codex config: %v", err)
	}
	if !strings.Contains(string(data), "[mcp_servers.codex-only]") {
		t.Errorf("expected codex-only in codex config: %s", string(data))
	}

	claudeCfg := filepath.Join(tempDir, ".claude.json")
	if _, err := os.Stat(claudeCfg); err == nil {
		t.Errorf("claude config should NOT have been created when --clients codex was given")
	}
}

func TestMCPListCmd_UnifiedView(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	// Create Claude server
	claudeCfg := filepath.Join(tempDir, ".claude.json")
	_ = os.WriteFile(claudeCfg, []byte(`{"mcpServers":{"claude-srv":{"command":"node","args":["server.js"]}}}`), 0600)

	// Create Codex server
	codexDir := filepath.Join(tempDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)
	codexCfg := filepath.Join(codexDir, "config.toml")
	_ = os.WriteFile(codexCfg, []byte("[mcp_servers.codex-srv]\ncommand = \"python\"\nargs = [\"main.py\"]\n"), 0600)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "list"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp list failed: %v", err)
		}
	})

	if !strings.Contains(out, "claude-srv") {
		t.Errorf("expected 'claude-srv' in unified list output: %s", out)
	}
	if !strings.Contains(out, "codex-srv") {
		t.Errorf("expected 'codex-srv' in unified list output: %s", out)
	}
}

func TestMCPCommands_InvalidProfileRefused(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	subcommands := [][]string{
		{"mcp", "install", "--profile", "nonexistent_prof"},
		{"mcp", "uninstall", "--profile", "nonexistent_prof"},
		{"mcp", "list", "--profile", "nonexistent_prof"},
		{"mcp", "add", "dummy", "echo", "hi", "--profile", "nonexistent_prof"},
		{"mcp", "remove", "dummy", "--profile", "nonexistent_prof"},
	}

	for _, args := range subcommands {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			cmd := newRootCmd(reg, pm)
			cmd.SetArgs(args)
			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error for nonexistent profile with %v, got nil", args)
			}
			if !strings.Contains(err.Error(), "does not exist") {
				t.Errorf("expected 'does not exist' error, got: %v", err)
			}
		})
	}
}
