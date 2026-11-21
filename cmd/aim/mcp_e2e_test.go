package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/service"
	"github.com/aim-cli/aim/internal/web"
)

func TestMCP_EndToEnd_CLI_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)
	profName := "work"
	_, err := pm.EnsureProfile(profName)
	if err != nil {
		t.Fatalf("failed to ensure profile %s: %v", profName, err)
	}

	// 1. Add remote server with headers to profile "work"
	out, errOut := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{
			"mcp", "add", "posthog", "https://mcp.posthog.com/mcp",
			"--profile", profName,
			"--clients", "claude,codex,agy,cursor",
			"--header", "Authorization: Bearer phx_secret_token",
			"-H", "X-Custom: test-val",
			"--env", "ENV_VAR=1",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp add failed: %v", err)
		}
	})
	if !strings.Contains(out, "Added MCP server 'posthog'") {
		t.Fatalf("expected add confirmation, got out=%q, errOut=%q", out, errOut)
	}

	// Verify Claude profile config
	pDir := pm.ProfileDir(profName)
	claudePath := filepath.Join(pDir, ".claude", ".claude.json")
	cData, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read claude profile config: %v", err)
	}
	if !strings.Contains(string(cData), "Bearer phx_secret_token") {
		t.Errorf("claude config missing Bearer token: %s", string(cData))
	}

	// Verify Codex profile config
	codexPath := filepath.Join(pDir, ".codex", "config.toml")
	cxData, err := os.ReadFile(codexPath)
	if err != nil {
		t.Fatalf("failed to read codex profile config: %v", err)
	}
	if !strings.Contains(string(cxData), "Bearer phx_secret_token") || !strings.Contains(string(cxData), "test-val") {
		t.Errorf("codex config missing headers: %s", string(cxData))
	}

	// Verify Antigravity profile config
	agyPath := filepath.Join(pDir, ".gemini", "config", "mcp_config.json")
	aData, err := os.ReadFile(agyPath)
	if err != nil {
		t.Fatalf("failed to read agy profile config: %v", err)
	}
	if !strings.Contains(string(aData), "Bearer phx_secret_token") {
		t.Errorf("agy config missing Bearer token: %s", string(aData))
	}

	// Verify Cursor profile config
	cursorPath := filepath.Join(pDir, ".cursor", "mcp.json")
	curData, err := os.ReadFile(cursorPath)
	if err != nil {
		t.Fatalf("failed to read cursor profile config: %v", err)
	}
	if !strings.Contains(string(curData), "Bearer phx_secret_token") || !strings.Contains(string(curData), "test-val") {
		t.Errorf("cursor config missing headers: %s", string(curData))
	}

	// 2. Run mcp list for profile "work"
	listOut, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "list", "--profile", profName})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp list failed: %v", err)
		}
	})
	if !strings.Contains(listOut, "posthog") || !strings.Contains(listOut, "enabled") {
		t.Errorf("expected posthog enabled in list output: %s", listOut)
	}

	// 3. Disable posthog on profile "work"
	disOut, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "disable", "posthog", "--profile", profName})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp disable failed: %v", err)
		}
	})
	if !strings.Contains(disOut, "Disabled MCP server 'posthog'") {
		t.Errorf("expected disable confirmation: %s", disOut)
	}

	// Verify disabled status in list output
	listDisOut, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "list", "--profile", profName})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp list failed: %v", err)
		}
	})
	if !strings.Contains(listDisOut, "disabled") {
		t.Errorf("expected disabled status in list output: %s", listDisOut)
	}

	// 4. Enable posthog on profile "work"
	enOut, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "enable", "posthog", "--profile", profName})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp enable failed: %v", err)
		}
	})
	if !strings.Contains(enOut, "Enabled MCP server 'posthog'") {
		t.Errorf("expected enable confirmation: %s", enOut)
	}

	// 5. Test active session warning on modification
	store := pm.MergeStateStore()
	_ = os.MkdirAll(store.Dir, 0700)
	lock, err := merge.LockShared(store.SessionsLockPath(profName, "claude"))
	if err != nil {
		t.Fatalf("failed to take shared lock: %v", err)
	}
	defer lock.Unlock()

	_, sessWarnErr := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "disable", "posthog", "--profile", profName})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp disable with active session failed: %v", err)
		}
	})
	if !strings.Contains(sessWarnErr, "Notice: Active session detected for profile \"work\"") {
		t.Errorf("expected active session warning in stderr: %s", sessWarnErr)
	}

	// 6. Remove posthog from profile "work"
	rmOut, _ := captureOutput(t, func() {
		cmd := newRootCmd(reg, pm)
		cmd.SetArgs([]string{"mcp", "remove", "posthog", "--profile", profName})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("mcp remove failed: %v", err)
		}
	})
	if !strings.Contains(rmOut, "Removed MCP server 'posthog'") {
		t.Errorf("expected remove confirmation: %s", rmOut)
	}
}

func TestMCP_EndToEnd_WebAPI_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	t.Setenv("AIM_REAL_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("dev")

	profileSvc := service.NewProfileService(pm, reg)
	sessionSvc := service.NewSessionService(nil, nil)
	launcherSvc := service.NewLauncherService()
	mcpSvc := service.NewMCPService(pm, reg)

	srv := web.NewServer(profileSvc, sessionSvc, launcherSvc, 0, false, mcpSvc)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	client := ts.Client()

	// 1. POST /api/mcp (Register new remote server)
	addPayload := service.AddMCPServerRequest{
		Name:    "api-tool",
		URL:     "https://mcp.example.com/sse",
		Headers: map[string]string{"Authorization": "Bearer secret-token"},
		Profile: "dev",
		Clients: []string{"claude", "codex", "cursor"},
	}
	body, _ := json.Marshal(addPayload)
	req, _ := http.NewRequestWithContext(context.Background(), "POST", ts.URL+"/api/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/mcp failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	// 2. GET /api/mcp?profile=dev
	getReq, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL+"/api/mcp?profile=dev", nil)
	getResp, err := client.Do(getReq)
	if err != nil {
		t.Fatalf("GET /api/mcp failed: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", getResp.StatusCode)
	}

	var servers []service.MCPServerDTO
	if err := json.NewDecoder(getResp.Body).Decode(&servers); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	var targetSrv *service.MCPServerDTO
	for _, s := range servers {
		if s.Name == "api-tool" {
			sc := s
			targetSrv = &sc
			break
		}
	}
	if targetSrv == nil {
		t.Fatalf("expected api-tool in servers list")
	}
	if targetSrv.Status != "enabled" || targetSrv.Type != "sse" || targetSrv.Auth != "connected" {
		t.Errorf("unexpected target server dto: %+v", targetSrv)
	}

	// 3. PATCH /api/mcp/api-tool/state (Toggle to disabled)
	patchPayload := service.SetMCPServerStateRequest{
		Enabled: false,
		Profile: "dev",
		Clients: []string{"claude", "codex", "cursor"},
	}
	patchBody, _ := json.Marshal(patchPayload)
	patchReq, _ := http.NewRequestWithContext(context.Background(), "PATCH", ts.URL+"/api/mcp/api-tool/state", bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := client.Do(patchReq)
	if err != nil {
		t.Fatalf("PATCH failed: %v", err)
	}
	defer patchResp.Body.Close()
	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", patchResp.StatusCode)
	}

	// Verify disabled in GET
	getReq2, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL+"/api/mcp?profile=dev", nil)
	getResp2, err := client.Do(getReq2)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer getResp2.Body.Close()
	var servers2 []service.MCPServerDTO
	_ = json.NewDecoder(getResp2.Body).Decode(&servers2)
	for _, s := range servers2 {
		if s.Name == "api-tool" && s.Status != "disabled" {
			t.Errorf("expected api-tool to be disabled, got %s", s.Status)
		}
	}

	// 4. DELETE /api/mcp/api-tool?profile=dev
	delReq, _ := http.NewRequestWithContext(context.Background(), "DELETE", ts.URL+"/api/mcp/api-tool?profile=dev&clients=claude,codex,cursor", nil)
	delResp, err := client.Do(delReq)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", delResp.StatusCode)
	}

	// Verify gone in GET
	getReq3, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL+"/api/mcp?profile=dev", nil)
	getResp3, err := client.Do(getReq3)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer getResp3.Body.Close()
	var servers3 []service.MCPServerDTO
	_ = json.NewDecoder(getResp3.Body).Decode(&servers3)
	for _, s := range servers3 {
		if s.Name == "api-tool" {
			t.Errorf("expected api-tool to be removed")
		}
	}
}
