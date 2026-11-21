package cursor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
)

func TestCursorMCP_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	hostHome := filepath.Join(tempDir, "home")
	profDir := filepath.Join(tempDir, "profile")
	_ = os.MkdirAll(hostHome, 0700)
	_ = os.MkdirAll(profDir, 0700)

	t.Setenv("AIM_REAL_HOME", hostHome)
	t.Setenv("HOME", hostHome)

	a := NewAdapter()
	ctx := context.Background()

	// 1. Add host MCP server
	err := a.AddHostMCPServer(ctx, "host-tool", agents.MCPServerConfig{
		Command: "npx",
		Args:    []string{"-y", "host-mcp-tool"},
	})
	if err != nil {
		t.Fatalf("AddHostMCPServer failed: %v", err)
	}

	// 2. Add profile MCP server
	err = a.AddProfileMCPServer(ctx, "dev", profDir, "profile-tool", agents.MCPServerConfig{
		URL:     "https://mcp.example.com/sse",
		Headers: map[string]string{"Authorization": "Bearer test-tok"},
	})
	if err != nil {
		t.Fatalf("AddProfileMCPServer failed: %v", err)
	}

	// 3. List servers and verify both are returned with proper origin/type
	servers, err := a.ListMCPServers(ctx, "dev", profDir)
	if err != nil {
		t.Fatalf("ListMCPServers failed: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d: %+v", len(servers), servers)
	}

	serverMap := make(map[string]agents.MCPServerInfo)
	for _, s := range servers {
		serverMap[s.Name] = s
	}

	hostSrv, ok := serverMap["host-tool"]
	if !ok || hostSrv.Origin != "host" || hostSrv.Type != "stdio" {
		t.Errorf("unexpected host-tool: %+v", hostSrv)
	}

	profSrv, ok := serverMap["profile-tool"]
	if !ok || profSrv.Origin != "profile" || profSrv.Type != "http" {
		t.Errorf("unexpected profile-tool: %+v", profSrv)
	}

	// 4. Toggle state
	err = a.SetProfileMCPServerState(ctx, "dev", profDir, "profile-tool", false)
	if err != nil {
		t.Fatalf("SetProfileMCPServerState disable failed: %v", err)
	}

	servers2, _ := a.ListMCPServers(ctx, "dev", profDir)
	for _, s := range servers2 {
		if s.Name == "profile-tool" && s.Status != "disabled" {
			t.Errorf("expected profile-tool to be disabled, got %s", s.Status)
		}
	}

	err = a.SetProfileMCPServerState(ctx, "dev", profDir, "profile-tool", true)
	if err != nil {
		t.Fatalf("SetProfileMCPServerState enable failed: %v", err)
	}

	// 5. Remove servers
	err = a.RemoveProfileMCPServer(ctx, "dev", profDir, "profile-tool")
	if err != nil {
		t.Fatalf("RemoveProfileMCPServer failed: %v", err)
	}
	err = a.RemoveHostMCPServer(ctx, "host-tool")
	if err != nil {
		t.Fatalf("RemoveHostMCPServer failed: %v", err)
	}

	serversAfter, _ := a.ListMCPServers(ctx, "dev", profDir)
	if len(serversAfter) != 0 {
		t.Errorf("expected 0 servers after removal, got %d", len(serversAfter))
	}
}

func TestCursor_DetectInstalled(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("PATH", tempDir)

	a := NewAdapter()
	if a.DetectInstalled(tempDir) {
		t.Errorf("expected cursor not installed in empty temp dir")
	}

	cursorDir := filepath.Join(tempDir, ".cursor")
	_ = os.MkdirAll(cursorDir, 0700)
	if !a.DetectInstalled(tempDir) {
		t.Errorf("expected cursor detected when .cursor dir exists")
	}
}
