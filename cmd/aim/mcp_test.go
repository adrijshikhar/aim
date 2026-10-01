package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

type mcpMockAdapter struct {
	mockAdapter
	servers []agents.MCPServerInfo
}

func (m *mcpMockAdapter) ListMCPServers(ctx context.Context, profileName, profileDir string) ([]agents.MCPServerInfo, error) {
	return m.servers, nil
}

func TestMCPCmd_TableOutput(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	mock := &mcpMockAdapter{
		mockAdapter: mockAdapter{name: "mock", binaryPath: "/bin/sh"},
		servers: []agents.MCPServerInfo{
			{Name: "atlassian", Type: "http", Status: "enabled", Auth: "OAuth", Target: "https://mcp.atlassian.com/v1"},
			{Name: "playwright", Type: "stdio", Status: "enabled", Auth: "unsupported", Target: "npx @playwright/mcp@latest"},
			{Name: "disabled-tool", Type: "stdio", Status: "disabled", Auth: "unsupported", Target: "echo"},
		},
	}

	reg := agents.NewRegistry()
	reg.Register(mock)

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"mcp", "list", "mock", "work"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	for _, expected := range []string{
		"=== Configured MCP Servers (mock: work) ===",
		"NAME",
		"STATUS",
		"AUTH",
		"TYPE",
		"TARGET / COMMAND",
		"atlassian",
		"playwright",
		"disabled-tool",
		"OAuth",
		"enabled",
		"disabled",
		"Total: 3 server(s) configured (2 enabled, 1 disabled)",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("output missing expected text %q:\n%s", expected, out)
		}
	}
}

func TestMCPCmd_JSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	mock := &mcpMockAdapter{
		mockAdapter: mockAdapter{name: "mock", binaryPath: "/bin/sh"},
		servers: []agents.MCPServerInfo{
			{Name: "atlassian", Type: "http", Status: "enabled", Auth: "OAuth", Target: "https://mcp.atlassian.com/v1"},
		},
	}

	reg := agents.NewRegistry()
	reg.Register(mock)

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(reg, pm)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"mcp", "list", "mock", "work", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []agents.MCPServerInfo
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, stdout.String())
	}

	if len(parsed) != 1 || parsed[0].Name != "atlassian" || parsed[0].Auth != "OAuth" {
		t.Fatalf("parsed servers mismatch: %+v", parsed)
	}
}

func TestRunCmd_MCPListInterception(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	mock := &mcpMockAdapter{
		mockAdapter: mockAdapter{
			name:       "mock",
			binaryPath: "/bin/sh",
			args:       []string{"-c", "echo 'NATIVE_CLI_WAS_CALLED'"},
		},
		servers: []agents.MCPServerInfo{
			{Name: "intercepted-srv", Type: "stdio", Status: "enabled", Auth: "none", Target: "tool"},
		},
	}

	reg := agents.NewRegistry()
	reg.Register(mock)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"run", "mock", "work", "mcp", "list"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(out, "NATIVE_CLI_WAS_CALLED") {
		t.Errorf("native binary should NOT have been called on plain mcp list: %s", out)
	}
	if !strings.Contains(out, "intercepted-srv") {
		t.Errorf("expected intercepted server in output: %s", out)
	}
}

func TestRunCmd_MCPListPassthroughOnJSON(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	mock := &mcpMockAdapter{
		mockAdapter: mockAdapter{
			name:       "mock",
			binaryPath: "/bin/echo",
			args:       nil,
		},
		servers: []agents.MCPServerInfo{
			{Name: "intercepted-srv", Type: "stdio", Status: "enabled", Auth: "none", Target: "tool"},
		},
	}

	reg := agents.NewRegistry()
	reg.Register(mock)

	out, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"run", "mock", "work", "mcp", "list", "--json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// echo should output "mcp list --json"
	if !strings.Contains(out, "mcp list --json") {
		t.Errorf("expected passthrough to /bin/echo, got: %s", out)
	}
}
