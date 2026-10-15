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

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "install", "codex"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Auto-install is currently supported for 'claude'") {
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

func TestMCPInstallAndUninstallCmd_MultiClient(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	// Create dummy Cursor dir
	cursorDir := filepath.Join(tempDir, ".cursor")
	_ = os.MkdirAll(cursorDir, 0700)

	// 1. Install to cursor
	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "install", "cursor"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp install cursor failed: %v", err)
		}
	})

	if !strings.Contains(out, "Cursor IDE") {
		t.Errorf("expected Cursor confirmation: %s", out)
	}

	cursorConfig := filepath.Join(cursorDir, "mcp.json")
	data, err := os.ReadFile(cursorConfig)
	if err != nil {
		t.Fatalf("failed to read cursor mcp.json: %v", err)
	}
	if !strings.Contains(string(data), `"aim"`) {
		t.Errorf("expected aim entry in cursor mcp.json: %s", string(data))
	}

	// 2. Uninstall from cursor
	out, _ = captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "uninstall", "cursor"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp uninstall cursor failed: %v", err)
		}
	})

	if !strings.Contains(out, "Removed aim") {
		t.Errorf("expected uninstall confirmation: %s", out)
	}

	data, _ = os.ReadFile(cursorConfig)
	if strings.Contains(string(data), `"aim"`) {
		t.Errorf("aim should have been removed from cursor mcp.json: %s", string(data))
	}
}

func TestMCPListCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cursorDir := filepath.Join(tempDir, ".cursor")
	_ = os.MkdirAll(cursorDir, 0700)
	cursorCfg := filepath.Join(cursorDir, "mcp.json")
	_ = os.WriteFile(cursorCfg, []byte(`{"mcpServers":{"demo":{"command":"echo","args":["hello"]}}}`), 0600)

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

func TestMCPAddAndRemoveCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)

	cursorDir := filepath.Join(tempDir, ".cursor")
	_ = os.MkdirAll(cursorDir, 0700)

	// 1. Add server
	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "add", "fetcher", "uvx", "mcp-server-fetch", "--clients", "cursor"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}
	})

	if !strings.Contains(out, "Added MCP server 'fetcher'") {
		t.Errorf("expected add confirmation: %s", out)
	}

	cursorCfg := filepath.Join(cursorDir, "mcp.json")
	data, _ := os.ReadFile(cursorCfg)
	if !strings.Contains(string(data), `"fetcher"`) {
		t.Errorf("expected fetcher in cursor config: %s", string(data))
	}

	// 2. Remove server
	out, _ = captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "remove", "fetcher", "--clients", "cursor"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp remove failed: %v", err)
		}
	})

	if !strings.Contains(out, "Removed MCP server 'fetcher'") {
		t.Errorf("expected remove confirmation: %s", out)
	}

	data, _ = os.ReadFile(cursorCfg)
	if strings.Contains(string(data), `"fetcher"`) {
		t.Errorf("fetcher should be removed from cursor config: %s", string(data))
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
