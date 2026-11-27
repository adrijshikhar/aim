package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
)

func TestMCPCmd_Help(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "--help"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, sub := range []string{"serve", "doctor", "install", "uninstall", "list", "add", "remove", "enable", "disable", "login"} {
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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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

func TestMCPServeShareCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"

	inBuf := bytes.NewBufferString(input)
	outBuf := new(bytes.Buffer)

	cmd := newRootCmd(reg, pm)
	cmd.SetIn(inBuf)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"mcp", "serve-share", "--profile", "peer1:testprof"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("mcp serve-share failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 responses, got %d: %s", len(lines), outBuf.String())
	}

	if !strings.Contains(lines[0], "aim-share") {
		t.Errorf("expected serverInfo.name 'aim-share', got %s", lines[0])
	}
	if !strings.Contains(lines[1], "aim_request_lease") || !strings.Contains(lines[1], "aim_lease_status") {
		t.Errorf("expected tools list to contain aim_request_lease and aim_lease_status, got %s", lines[1])
	}
}

func TestMCPInstallCmd_ClaudeGlobal(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
		cmd.SetArgs([]string{"mcp", "install", "unknown"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Auto-install is currently supported for 'claude', 'codex', 'agy', 'cursor'") {
		t.Errorf("expected guidance for non-supported agent, got: %s", out)
	}
}

func TestMCPInstallCmd_Cursor(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "install", "cursor", "--profile", "work"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Installed aim MCP server") {
		t.Errorf("expected confirmation in output: %s", out)
	}

	cursorPath := filepath.Join(pm.ProfileDir("work"), ".cursor", "mcp.json")
	data, err := os.ReadFile(cursorPath)
	if err != nil {
		t.Fatalf("failed to read profile cursor json at %s: %v", cursorPath, err)
	}

	var parsed struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal profile cursor json: %v", err)
	}

	if _, ok := parsed.MCPServers["aim"]; !ok {
		t.Errorf("expected 'aim' entry in mcpServers: %s", string(data))
	}
}

func TestMCPServeCmd_WhoamiTool(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

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

func TestMCPAddCmd_HTTPURLTarget(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "add", "remote-sse", "https://example.com/sse", "--clients", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add HTTP URL failed: %v", err)
		}
	})

	if !strings.Contains(out, "Added MCP server 'remote-sse' to Codex CLI") {
		t.Errorf("expected confirmation output, got: %s", out)
	}

	codexCfg := filepath.Join(tempDir, ".codex", "config.toml")
	data, err := os.ReadFile(codexCfg)
	if err != nil {
		t.Fatalf("failed to read codex config: %v", err)
	}
	cfgStr := string(data)
	if !strings.Contains(cfgStr, `url = `) || !strings.Contains(cfgStr, `https://example.com/sse`) {
		t.Errorf("expected url in codex config, got: %s", cfgStr)
	}
	if strings.Contains(cfgStr, "command =") {
		t.Errorf("expected no command field in codex config for URL server, got: %s", cfgStr)
	}
}

func TestMCPRemoveCmd_NonexistentServer(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	// Add an empty config so codex manager is considered installed
	codexDir := filepath.Join(tempDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)
	_ = os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte("[mcp_servers]\n"), 0600)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "remove", "missing-server", "--clients", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp remove failed: %v", err)
		}
	})

	if !strings.Contains(out, "Server 'missing-server' was not found in selected agent configs.") {
		t.Errorf("expected not found message, got: %s", out)
	}
}

func TestMCPAddAndRemoveCmd_UnknownClient(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	addCmd := newRootCmd(reg, pm)
	addCmd.SetArgs([]string{"mcp", "add", "test", "echo", "1", "--clients", "unknown-client"})
	err := addCmd.Execute()
	if err == nil {
		t.Fatal("expected error with unknown client in mcp add, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported or unconfigurable agent client") {
		t.Errorf("unexpected error message: %v", err)
	}

	rmCmd := newRootCmd(reg, pm)
	rmCmd.SetArgs([]string{"mcp", "remove", "test", "--clients", "unknown-client"})
	err = rmCmd.Execute()
	if err == nil {
		t.Fatal("expected error with unknown client in mcp remove, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported or unconfigurable agent client") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMCPCommands_ProfilePathTraversalProhibited(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	traversalProfiles := []string{
		"../../evil",
		"../escape",
		"profile/subdir",
		"profile\\backslash",
		"bad@name!",
	}

	for _, badProfile := range traversalProfiles {
		subcommands := [][]string{
			{"mcp", "install", "--profile", badProfile},
			{"mcp", "uninstall", "--profile", badProfile},
			{"mcp", "list", "--profile", badProfile},
			{"mcp", "add", "dummy", "echo", "hi", "--profile", badProfile},
			{"mcp", "remove", "dummy", "--profile", badProfile},
		}

		for _, args := range subcommands {
			testName := badProfile + "_" + strings.Join(args, "_")
			t.Run(testName, func(t *testing.T) {
				cmd := newRootCmd(reg, pm)
				cmd.SetArgs(args)
				err := cmd.Execute()
				if err == nil {
					t.Fatalf("expected error for invalid profile traversal '%s' with %v, got nil", badProfile, args)
				}
			})
		}
	}
}

func TestResolveTargetScope(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("dev")
	_, _ = pm.EnsureProfile("prod")

	oldInteractive := mcpIsInteractive
	defer func() { mcpIsInteractive = oldInteractive }()

	t.Run("explicit host flag", func(t *testing.T) {
		t.Setenv("AIM_PROFILE", "dev")
		targets, err := ResolveTargetScope(pm, "", true, false, strings.NewReader(""), &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if targets != nil {
			t.Fatalf("expected nil targets for host, got %v", targets)
		}
	})

	t.Run("explicit profile flag", func(t *testing.T) {
		t.Setenv("AIM_PROFILE", "dev")
		targets, err := ResolveTargetScope(pm, "prod", false, false, strings.NewReader(""), &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(targets) != 1 || targets[0].Name != "prod" {
			t.Fatalf("expected target prod, got %v", targets)
		}
	})

	t.Run("explicit all flag", func(t *testing.T) {
		targets, err := ResolveTargetScope(pm, "", false, true, strings.NewReader(""), &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(targets) != 2 {
			t.Fatalf("expected 2 targets, got %d", len(targets))
		}
	})

	t.Run("outside session without flags", func(t *testing.T) {
		t.Setenv("AIM_PROFILE", "")
		targets, err := ResolveTargetScope(pm, "", false, false, strings.NewReader(""), &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if targets != nil {
			t.Fatalf("expected nil targets for host outside session, got %v", targets)
		}
	})

	t.Run("inside session non-interactive", func(t *testing.T) {
		t.Setenv("AIM_PROFILE", "dev")
		mcpIsInteractive = func(io.Reader) bool { return false }
		targets, err := ResolveTargetScope(pm, "", false, false, strings.NewReader(""), &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(targets) != 1 || targets[0].Name != "dev" {
			t.Fatalf("expected dev profile target, got %v", targets)
		}
	})

	t.Run("inside session interactive responses", func(t *testing.T) {
		t.Setenv("AIM_PROFILE", "dev")
		mcpIsInteractive = func(io.Reader) bool { return true }

		// Affirmative responses
		for _, input := range []string{"\n", "y\n", "yes\n", "Y\n"} {
			var out bytes.Buffer
			targets, err := ResolveTargetScope(pm, "", false, false, strings.NewReader(input), &out)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", input, err)
			}
			if len(targets) != 1 || targets[0].Name != "dev" {
				t.Fatalf("expected dev profile target for %q, got %v", input, targets)
			}
		}

		// Host responses
		for _, input := range []string{"host\n", "h\n", "HOST\n"} {
			var out bytes.Buffer
			targets, err := ResolveTargetScope(pm, "", false, false, strings.NewReader(input), &out)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", input, err)
			}
			if targets != nil {
				t.Fatalf("expected nil targets for %q, got %v", input, targets)
			}
		}

		// Cancellation responses
		for _, input := range []string{"n\n", "no\n"} {
			var out bytes.Buffer
			_, err := ResolveTargetScope(pm, "", false, false, strings.NewReader(input), &out)
			if err == nil || !strings.Contains(err.Error(), "operation cancelled") {
				t.Fatalf("expected cancellation error for %q, got %v", input, err)
			}
		}

		// Invalid responses
		var out bytes.Buffer
		_, err := ResolveTargetScope(pm, "", false, false, strings.NewReader("invalid\n"), &out)
		if err == nil || !strings.Contains(err.Error(), "invalid response") {
			t.Fatalf("expected invalid response error, got %v", err)
		}
	})
}

func TestMCPAddCmd_HeaderAndEnvFlags(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{
			"mcp", "add", "posthog", "https://mcp.posthog.com/mcp",
			"--clients", "codex",
			"--header", "Authorization: Bearer my-secret-token",
			"-H", "X-Custom: test-val",
			"--env", "API_KEY=xyz",
			"-e", "DEBUG=true",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}
	})

	if !strings.Contains(out, "Added MCP server 'posthog' to Codex CLI") {
		t.Fatalf("expected success message, got: %s", out)
	}

	codexCfg := filepath.Join(tempDir, ".codex", "config.toml")
	data, err := os.ReadFile(codexCfg)
	if err != nil {
		t.Fatalf("failed to read codex config: %v", err)
	}
	cfgStr := string(data)
	if !strings.Contains(cfgStr, "Bearer my-secret-token") {
		t.Errorf("expected Authorization header in codex config: %s", cfgStr)
	}
	if !strings.Contains(cfgStr, "test-val") {
		t.Errorf("expected X-Custom header in codex config: %s", cfgStr)
	}
	if !strings.Contains(cfgStr, "API_KEY") || !strings.Contains(cfgStr, "DEBUG") {
		t.Errorf("expected env vars in codex config: %s", cfgStr)
	}
}

func TestMCPEnableAndDisableCmd(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	// Add server to codex and claude
	addOut, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{
			"mcp", "add", "toggle-srv", "https://example.com/mcp",
			"--clients", "codex,claude",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}
	})
	if !strings.Contains(addOut, "Added MCP server 'toggle-srv'") {
		t.Fatalf("expected add confirmation, got: %s", addOut)
	}

	// Disable server
	disOut, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{
			"mcp", "disable", "toggle-srv",
			"--clients", "codex,claude",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp disable failed: %v", err)
		}
	})
	if !strings.Contains(disOut, "Disabled MCP server 'toggle-srv'") {
		t.Fatalf("expected disable confirmation, got: %s", disOut)
	}

	// Verify codex config has enabled = false
	codexCfg := filepath.Join(tempDir, ".codex", "config.toml")
	cData, _ := os.ReadFile(codexCfg)
	if !strings.Contains(string(cData), "enabled = false") {
		t.Errorf("expected enabled = false in codex config: %s", string(cData))
	}

	// Verify claude config has disabled: true
	claudeCfg := filepath.Join(tempDir, ".claude.json")
	clData, _ := os.ReadFile(claudeCfg)
	var clMap map[string]any
	_ = json.Unmarshal(clData, &clMap)
	mcpServers, _ := clMap["mcpServers"].(map[string]any)
	srv, _ := mcpServers["toggle-srv"].(map[string]any)
	if disabled, ok := srv["disabled"].(bool); !ok || !disabled {
		t.Errorf("expected disabled=true in claude config: %s", string(clData))
	}

	// Enable server
	enOut, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{
			"mcp", "enable", "toggle-srv",
			"--clients", "codex,claude",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp enable failed: %v", err)
		}
	})
	if !strings.Contains(enOut, "Enabled MCP server 'toggle-srv'") {
		t.Fatalf("expected enable confirmation, got: %s", enOut)
	}

	// Verify codex config has enabled = true
	cData2, _ := os.ReadFile(codexCfg)
	if strings.Contains(string(cData2), "enabled = false") {
		t.Errorf("expected enabled = true in codex config: %s", string(cData2))
	}
}

func TestMCPCmd_ActiveSessionNotice(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	// Create an active session lock on profile "work" for agent "codex"
	store := pm.MergeStateStore()
	_ = os.MkdirAll(store.Dir, 0700)
	lock, err := merge.LockShared(store.SessionsLockPath("work", "codex"))
	if err != nil {
		t.Fatalf("failed to take shared lock: %v", err)
	}
	defer lock.Unlock()

	_, stderr := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{
			"mcp", "add", "session-test", "https://example.com/mcp",
			"--profile", "work",
			"--clients", "codex",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}
	})

	if !strings.Contains(stderr, "Notice: Active session detected for profile \"work\"") {
		t.Errorf("expected active session warning in stderr, got: %q", stderr)
	}
}

func TestMCPLoginCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)

	// 1. Server not found error
	t.Run("missing server name", func(t *testing.T) {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "login", "missing-server", "--host"})
		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "could not find remote URL") {
			t.Fatalf("expected could not find remote URL error, got %v", err)
		}
	})

	// 2. Server does not require OAuth (200 OK)
	t.Run("non-oauth endpoint", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		out, _ := captureOutput(t, func() {
			cmd := newRootCmd(reg, pm)
			cmd.SetArgs([]string{"mcp", "login", ts.URL, "--host"})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("mcp login failed: %v", err)
			}
		})

		if !strings.Contains(out, "Server does not require OAuth authentication") {
			t.Errorf("expected no oauth required message, got %s", out)
		}
	})

	// 3. Configured server by name that points to URL
	t.Run("configured server URL resolution", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// Register server first
		addCmd := newRootCmd(reg, pm)
		addCmd.SetArgs([]string{"mcp", "add", "my-remote", ts.URL, "--clients", "codex", "--host"})
		if err := addCmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}

		out, _ := captureOutput(t, func() {
			cmd := newRootCmd(reg, pm)
			cmd.SetArgs([]string{"mcp", "login", "my-remote", "--clients", "codex", "--host"})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("mcp login by name failed: %v", err)
			}
		})

		if !strings.Contains(out, "Discovering authentication requirements") || !strings.Contains(out, "Server does not require OAuth authentication") {
			t.Errorf("expected discovery and no oauth message, got %s", out)
		}
	})
}

func TestSaveScopedMCPOAuthTokens_IsolatedPerProfile(t *testing.T) {
	base := t.TempDir()
	homeDir := filepath.Join(base, "home")
	workDir := filepath.Join(base, "profiles", "work")
	personalDir := filepath.Join(base, "profiles", "personal")

	tok := map[string]any{"access_token": "work-secret-token"}
	targets := []ProfileTarget{{Name: "work", Dir: workDir}}
	if err := saveScopedMCPOAuthTokens(homeDir, targets, "https://mcp.example.com/mcp", tok); err != nil {
		t.Fatalf("saveScopedMCPOAuthTokens failed: %v", err)
	}

	workRoot, err := os.OpenRoot(filepath.Join(workDir, ".mcp-auth"))
	if err != nil {
		t.Fatalf("open work .mcp-auth root: %v", err)
	}
	defer func() { _ = workRoot.Close() }()
	data, err := workRoot.ReadFile("tokens.json")
	if err != nil {
		t.Fatalf("read work tokens.json: %v", err)
	}
	if !strings.Contains(string(data), "work-secret-token") {
		t.Fatalf("expected token persisted in work profile, got: %s", string(data))
	}
	fi, err := workRoot.Stat("tokens.json")
	if err != nil {
		t.Fatalf("stat work tokens.json: %v", err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600 permissions on tokens.json, got %v", fi.Mode().Perm())
	}

	if _, err := os.Stat(filepath.Join(personalDir, ".mcp-auth", "tokens.json")); !os.IsNotExist(err) {
		t.Fatalf("LEAK: personal profile received work MCP token, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".mcp-auth", "tokens.json")); !os.IsNotExist(err) {
		t.Fatalf("LEAK: host home received work MCP token, err=%v", err)
	}
}

func TestResolveTargetScope_AllSkipsLeasedAndRejectsTraversalAIMProfile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	leasedDir := pm.ProfileDir("peer1:borrowed")
	if err := os.MkdirAll(leasedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(leasedDir, ".aim-lease.json"), []byte(`{"lease_id":"l1"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	targets, err := ResolveTargetScope(pm, "", false, true, strings.NewReader(""), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ResolveTargetScope(--all) failed: %v", err)
	}
	if len(targets) != 1 || targets[0].Name != "work" {
		t.Fatalf("expected only configured profile 'work', got %+v", targets)
	}

	t.Setenv("AIM_PROFILE", "../escape")
	if _, err := ResolveTargetScope(pm, "", false, false, strings.NewReader(""), &bytes.Buffer{}); err == nil {
		t.Fatalf("expected error for traversal AIM_PROFILE")
	}
}

func TestWriteMCPOAuthTokenFile_RejectsSymlinkAndCorruptJSON(t *testing.T) {
	base := t.TempDir()
	targetDir := filepath.Join(base, "target")
	if err := os.MkdirAll(targetDir, 0o700); err != nil {
		t.Fatal(err)
	}
	symlinkAuth := filepath.Join(base, "prof", ".mcp-auth")
	if err := os.MkdirAll(filepath.Dir(symlinkAuth), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetDir, symlinkAuth); err != nil {
		t.Fatal(err)
	}
	if err := writeMCPOAuthTokenFile(filepath.Join(symlinkAuth, "tokens.json"), "https://example.com", map[string]any{"a": "b"}); err == nil {
		t.Fatalf("expected writeMCPOAuthTokenFile to reject symlinked .mcp-auth dir")
	}

	corruptDir := filepath.Join(base, "corrupt", ".mcp-auth")
	if err := os.MkdirAll(corruptDir, 0o700); err != nil {
		t.Fatal(err)
	}
	corruptFile := filepath.Join(corruptDir, "tokens.json")
	if err := os.WriteFile(corruptFile, []byte("{not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeMCPOAuthTokenFile(corruptFile, "https://example.com", map[string]any{"a": "b"}); err == nil {
		t.Fatalf("expected writeMCPOAuthTokenFile to reject corrupt tokens.json")
	}

	cmd := newRootCmd(defaultRegistry(), profile.NewProfileManager(base))
	cmd.SetArgs([]string{"mcp", "add", "srv", "uvx", "--profile", "work", "--host"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when both --profile and --host are set")
	}
}
