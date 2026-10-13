package mcp_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	aimmcp "github.com/aim-cli/aim/internal/mcp"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
)

func TestMCPServer_InitializeAndListTools(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err := srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 JSON-RPC responses, got: %s", outBuf.String())
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
		t.Errorf("expected server name 'aim', got %q", initResp.Result.ServerInfo.Name)
	}

	var toolsResp struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &toolsResp); err != nil {
		t.Fatalf("failed to decode tools response: %v", err)
	}

	toolNames := make(map[string]bool)
	for _, tool := range toolsResp.Result.Tools {
		toolNames[tool.Name] = true
	}

	for _, expected := range []string{"aim_whoami", "aim_get_quota", "aim_list_profiles"} {
		if !toolNames[expected] {
			t.Errorf("expected tool %q in tools/list", expected)
		}
	}

	// Security assertion: NO write, switch, delete, or export tools exist
	for _, forbidden := range []string{
		"aim_switch_profile",
		"aim_delete_profile",
		"aim_create_profile",
		"aim_export_token",
		"aim_set_token",
	} {
		if toolNames[forbidden] {
			t.Fatalf("CRITICAL SECURITY VIOLATION: forbidden tool %q found in read-only MCP server", forbidden)
		}
	}
}

func TestMCPServer_NotificationsIgnored(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n" +
		`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err := srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly 1 response for ping, got %d: %s", len(lines), outBuf.String())
	}

	var pingResp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &pingResp); err != nil {
		t.Fatalf("failed to decode ping response: %v", err)
	}
	if pingResp.ID != 1 {
		t.Errorf("expected ID 1, got %d", pingResp.ID)
	}
}

func TestMCPServer_UnknownMethod(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":42,"method":"nonexistent/method"}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err := srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var resp struct {
		ID    int `json:"id"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error object, got nil")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected code -32601, got %d", resp.Error.Code)
	}
}

func TestMCPServer_ToolCall_Whoami(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_PROFILE", "work")
	t.Setenv("AIM_AGENT", "claude")

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, err := pm.EnsureProfile("work")
	if err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}

	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"aim_whoami","arguments":{}}}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err = srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var callResp struct {
		ID     int `json:"id"`
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &callResp); err != nil {
		t.Fatalf("failed to decode call response: %v", err)
	}
	if callResp.ID != 10 {
		t.Errorf("expected ID 10, got %d", callResp.ID)
	}
	if callResp.Result.IsError {
		t.Errorf("expected isError to be false, got true")
	}
	if len(callResp.Result.Content) == 0 {
		t.Fatalf("expected content block, got none")
	}

	text := callResp.Result.Content[0].Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("expected JSON formatted text output in content block, got %q: %v", text, err)
	}

	if data["profile"] != "work" {
		t.Errorf("expected profile 'work', got %v", data["profile"])
	}
	if data["agent"] != "claude" {
		t.Errorf("expected agent 'claude', got %v", data["agent"])
	}
}

func TestMCPServer_ToolCall_Whoami_StaleQuota(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_PROFILE", "work")
	t.Setenv("AIM_AGENT", "claude")

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	// Pre-seed an expired/stale report with short TTL
	cacheStore := usage.NewCacheStore(tempDir, time.Millisecond)
	_ = cacheStore.Put(usage.Report{
		Agent:     "claude",
		Profile:   "work",
		Status:    usage.StatusOK,
		Summary:   "Stale 50% remaining",
		FetchedAt: time.Now().Add(-time.Hour),
	})
	time.Sleep(10 * time.Millisecond) // expire TTL

	srv := aimmcp.NewServer(reg, pm, tempDir)
	input := `{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"aim_whoami","arguments":{}}}` + "\n"
	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	if err := srv.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var callResp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &callResp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(callResp.Result.Content) == 0 {
		t.Fatal("expected content")
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(callResp.Result.Content[0].Text), &data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if data["quota_summary"] != "Stale 50% remaining" {
		t.Errorf("expected stale quota %q, got %v", "Stale 50% remaining", data["quota_summary"])
	}
}

func TestMCPServer_ToolCall_GetQuota(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_PROFILE", "dev")
	t.Setenv("AIM_AGENT", "agy")

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("dev")

	// Pre-seed quota cache
	cacheStore := usage.NewCacheStore(tempDir, time.Hour)
	err := cacheStore.Put(usage.Report{
		Agent:   "agy",
		Profile: "dev",
		Status:  usage.StatusOK,
		Summary: "85% remaining (resets in 3h)",
		Windows: []usage.LimitWindow{
			{
				Name:         "5-hour limit",
				RemainingPct: 85,
				ResetsIn:     3 * time.Hour,
			},
		},
		FetchedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to seed cache: %v", err)
	}

	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":20,"method":"tools/call","params":{"name":"aim_get_quota","arguments":{"agent":"agy","profile":"dev"}}}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err = srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var callResp struct {
		ID     int `json:"id"`
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &callResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(callResp.Result.Content) == 0 {
		t.Fatalf("expected content block, got empty")
	}

	text := callResp.Result.Content[0].Text
	var rep usage.Report
	if err := json.Unmarshal([]byte(text), &rep); err != nil {
		t.Fatalf("failed to decode quota report from tool output: %v (text: %s)", err, text)
	}

	if rep.Status != usage.StatusOK {
		t.Errorf("expected status OK, got %v", rep.Status)
	}
	if len(rep.Windows) != 1 || rep.Windows[0].RemainingPct != 85 {
		t.Errorf("expected window with 85%% remaining, got %+v", rep.Windows)
	}
}

func TestMCPServer_ToolCall_ListProfiles(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("personal")

	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":30,"method":"tools/call","params":{"name":"aim_list_profiles","arguments":{}}}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err := srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var callResp struct {
		ID     int `json:"id"`
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &callResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(callResp.Result.Content) == 0 {
		t.Fatalf("expected content block")
	}

	text := callResp.Result.Content[0].Text
	var res struct {
		Profiles []string `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(text), &res); err != nil {
		t.Fatalf("failed to decode profile list json: %v (text: %s)", err, text)
	}

	if len(res.Profiles) < 2 {
		t.Errorf("expected at least 2 profiles, got %d: %v", len(res.Profiles), res.Profiles)
	}
}

func TestMCPServer_ToolCall_UnknownTool(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":40,"method":"tools/call","params":{"name":"nonexistent_tool","arguments":{}}}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err := srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var callResp struct {
		ID    int `json:"id"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &callResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if callResp.Error == nil || callResp.Error.Code != -32601 {
		t.Errorf("expected -32601 error for unknown tool, got %+v", callResp.Error)
	}
}

func TestMCPServer_ToolCall_WhoamiUnattached(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_PROFILE", "")
	t.Setenv("AIM_AGENT", "")
	t.Setenv("HOME", tempDir) // not in profiles dir

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":50,"method":"tools/call","params":{"name":"aim_whoami","arguments":{}}}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err := srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
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
		t.Fatalf("failed to decode response: %v", err)
	}
	if callResp.Result.IsError {
		t.Errorf("expected isError false, got true")
	}

	text := callResp.Result.Content[0].Text
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatalf("failed to unmarshal unattached payload: %v", err)
	}
	if payload["status"] != "unattached" {
		t.Errorf("expected status 'unattached', got %v", payload["status"])
	}
}

func TestMCPServer_MalformedJSONRecovery(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{malformed json}` + "\n" +
		`{"jsonrpc":"2.0","id":99,"method":"ping"}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	err := srv.Serve(&inBuf, &outBuf)
	if err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 responses (parse error + ping), got %d: %s", len(lines), outBuf.String())
	}

	var errResp struct {
		Error struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &errResp); err != nil {
		t.Fatalf("failed to decode parse error response: %v", err)
	}
	if errResp.Error.Code != -32700 {
		t.Errorf("expected -32700 for parse error, got %d", errResp.Error.Code)
	}

	var pingResp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &pingResp); err != nil {
		t.Fatalf("failed to decode ping response: %v", err)
	}
	if pingResp.ID != 99 {
		t.Errorf("expected ID 99, got %d", pingResp.ID)
	}
}

func TestMCPServer_AgentIdentityIsolation(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)
	pDir, _ := pm.EnsureProfile("multi-agent")

	// Set up Codex credentials in the profile
	codexDir := filepath.Join(pDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)
	_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(`{"email":"codex@example.com"}`), 0600)

	// Set up Claude credentials in the profile
	claudeDir := filepath.Join(pDir, ".claude")
	_ = os.MkdirAll(claudeDir, 0700)
	_ = os.WriteFile(filepath.Join(claudeDir, ".claude.json"), []byte(`{"oauthAccount":{"emailAddress":"claude@example.com"}}`), 0600)

	// Simulate active agent is Claude
	t.Setenv("AIM_PROFILE", "multi-agent")
	t.Setenv("AIM_AGENT", "claude")

	reg := agents.NewRegistry()
	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":101,"method":"tools/call","params":{"name":"aim_whoami","arguments":{}}}` + "\n"
	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	if err := srv.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var callResp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &callResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	text := callResp.Result.Content[0].Text
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatalf("failed to unmarshal whoami payload: %v", err)
	}

	if payload["account_email"] != "claude@example.com" {
		t.Errorf("expected claude email 'claude@example.com', got %v (isolation breach)", payload["account_email"])
	}
}

func TestMCPServer_NilProfileManager(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	srv := aimmcp.NewServer(reg, nil, tempDir)

	// whoami with nil pm
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"aim_whoami","arguments":{}}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"aim_list_profiles","arguments":{}}}` + "\n"

	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	if err := srv.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(lines))
	}

	for i, line := range lines {
		var resp struct {
			Result struct {
				IsError bool `json:"isError"`
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("line %d: decode error: %v", i, err)
		}
		if !resp.Result.IsError {
			t.Errorf("line %d: expected isError true when pm is nil", i)
		}
	}
}

func TestMCPServer_ToolCall_GetQuota_CaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	cacheStore := usage.NewCacheStore(tempDir, time.Hour)
	_ = cacheStore.Put(usage.Report{
		Agent:   "claude",
		Profile: "work",
		Status:  usage.StatusOK,
		Summary: "Available",
	})

	srv := aimmcp.NewServer(reg, pm, tempDir)

	// uppercase agent "Claude"
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"aim_get_quota","arguments":{"agent":"Claude","profile":"work"}}}` + "\n"
	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	if err := srv.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var resp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	var rep usage.Report
	if err := json.Unmarshal([]byte(resp.Result.Content[0].Text), &rep); err != nil {
		t.Fatalf("expected quota report for case-insensitive agent, got: %s", resp.Result.Content[0].Text)
	}
	if rep.Summary != "Available" {
		t.Errorf("expected summary 'Available', got %q", rep.Summary)
	}
}

func TestMCPServer_ToolCall_GetQuota_InvalidArgs(t *testing.T) {
	tempDir := t.TempDir()
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	srv := aimmcp.NewServer(reg, pm, tempDir)

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"aim_get_quota","arguments":"not-an-object"}}` + "\n"
	var inBuf bytes.Buffer
	inBuf.WriteString(input)
	var outBuf bytes.Buffer

	if err := srv.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var resp struct {
		Error *struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != -32602 {
		t.Errorf("expected -32602 for invalid arguments, got %+v", resp.Error)
	}
}
