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

	if !strings.Contains(out, "serve") {
		t.Errorf("expected 'serve' subcommand in help output: %s", out)
	}
	if !strings.Contains(out, "install") {
		t.Errorf("expected 'install' subcommand in help output: %s", out)
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
