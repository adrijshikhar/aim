package main

import (
	"context"
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
