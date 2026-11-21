package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type mockMCPManager struct{}

func (m *mockMCPManager) ListMCPServers(ctx context.Context, profileName, profileDir string) ([]MCPServerInfo, error) {
	return nil, nil
}

func (m *mockMCPManager) AddHostMCPServer(ctx context.Context, name string, cfg MCPServerConfig) error {
	return nil
}

func (m *mockMCPManager) RemoveHostMCPServer(ctx context.Context, name string) error {
	return nil
}

func (m *mockMCPManager) AddProfileMCPServer(ctx context.Context, profileName, profileDir, name string, cfg MCPServerConfig) error {
	return nil
}

func (m *mockMCPManager) RemoveProfileMCPServer(ctx context.Context, profileName, profileDir, name string) error {
	return nil
}

func (m *mockMCPManager) SetHostMCPServerState(ctx context.Context, name string, enabled bool) error {
	return nil
}

func (m *mockMCPManager) SetProfileMCPServerState(ctx context.Context, profileName, profileDir, name string, enabled bool) error {
	return nil
}

func (m *mockMCPManager) DetectInstalled(realHome string) bool {
	return true
}

func TestMCPConfigManager_InterfaceAssertion(t *testing.T) {
	var _ MCPConfigManager = (*mockMCPManager)(nil)
}

func TestMCPServerConfig_Headers(t *testing.T) {
	cfg := MCPServerConfig{
		URL: "https://mcp.posthog.com/mcp",
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
	}
	if cfg.Headers["Authorization"] != "Bearer test-token" {
		t.Errorf("expected header set, got %v", cfg.Headers)
	}
}

func TestMCPConfigManager_StateToggles(t *testing.T) {
	var mgr MCPConfigManager
	// Compile-time check for method existence on the interface
	type stateToggler interface {
		SetHostMCPServerState(ctx context.Context, name string, enabled bool) error
		SetProfileMCPServerState(ctx context.Context, profileName, profileDir, name string, enabled bool) error
	}
	var _ stateToggler = mgr
}

func TestArgsMatch(t *testing.T) {
	flags := []string{"--bg", "--background"}
	first := []string{"agents", "attach", "respawn"}
	opus := []string{"--model", "opus"}
	cases := []struct {
		profile, args []string
		want          bool
	}{
		{nil, []string{"--bg"}, true},
		{nil, []string{"--model", "opus", "--background"}, true},
		{nil, []string{"agents"}, true},
		{nil, []string{"attach", "id"}, true},
		{nil, []string{"--resume", "x"}, false},
		{nil, []string{"-p", "agents"}, false},      // a prompt that happens to say "agents"
		{nil, []string{"fix", "--", "--bg"}, false}, // after "--" belongs to the agent
		{nil, nil, false},
		// profiles.<p>.args precede the CLI args in the launched command
		{opus, []string{"agents"}, true}, // the CLI's first word, after profile flags
		{opus, []string{"--bg"}, true},
		{[]string{"--bg"}, nil, true},
		{[]string{"agents"}, nil, true},                  // the profile's first word
		{[]string{"agents"}, []string{"--resume"}, true}, // the profile's first word leads
		{opus, nil, false},
		{opus, []string{"fix", "agents"}, false},         // not a first word of either
		{[]string{"--", "x"}, []string{"--bg"}, false},   // the profile's "--" ends the scan
		{[]string{"--", "x"}, []string{"agents"}, false}, // …and makes the CLI args the agent's
	}
	for _, c := range cases {
		if got := ArgsMatch(c.profile, c.args, flags, first); got != c.want {
			t.Errorf("%v + %v: got %v want %v", c.profile, c.args, got, c.want)
		}
	}
}

func TestNormalizeAuth(t *testing.T) {
	cases := map[string]string{
		"o_auth":        "OAuth",
		"OAuth":         "OAuth",
		"connected":     "OAuth",
		"not_logged_in": "auth required",
		"auth_required": "auth required",
		"unsupported":   "unsupported",
		"none":          "unsupported",
		"":              "unsupported",
		"custom":        "custom",
	}
	for in, want := range cases {
		if got := NormalizeAuth(in); got != want {
			t.Errorf("NormalizeAuth(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCollapseCommand(t *testing.T) {
	cases := []struct {
		cmd  string
		args []string
		want string
	}{
		{
			cmd:  "npx",
			args: []string{"@playwright/mcp@latest"},
			want: "npx @playwright/mcp@latest",
		},
		{
			cmd:  "sh",
			args: []string{"-c", `url="${LOCAL_GRAFANA_URL:-http://localhost:3000}"; exec env GRAFANA_URL="$url" uvx mcp-grafana`},
			want: "uvx mcp-grafana",
		},
		{
			cmd:  "sh",
			args: []string{"-c", `key="${CORALOGIX_API_KEY:?required}"; exec npx -y mcp-remote@latest https://api.eu1.coralogix.com/mgmt/api/v1/mcp --header "Authorization: Bearer $key"`},
			want: "npx (api.eu1.coralogix.com)",
		},
		{
			cmd:  "sh",
			args: []string{"-c", `url="${SKRULL_BASE_URL:-https://skrull.me}"; exec env SKRULL_BASE_URL="$url" uvx --from 'skrull[mcp] @ git+https://github.com/hevoio/skrull@dev' skrull-mcp`},
			want: "uvx skrull-mcp",
		},
		{
			cmd:  "node",
			args: []string{"-e", `const f=require('fs'); ... if(f.existsSync(p.join(r,'scripts','mcp-server.cjs'))) ...`},
			want: "node .../mcp-server.cjs",
		},
		{
			cmd:  "/Applications/ChatGPT.app/Contents/Resources/cua_node/bin/node_repl",
			args: nil,
			want: "node_repl",
		},
		{
			cmd:  "/Users/nemesis/.caveman/bin/caveman-mcp",
			args: nil,
			want: "caveman-mcp",
		},
		{
			cmd:  "python",
			args: []string{"-c", `import hevo_mcp.server; server.main()`},
			want: "python -m hevo_mcp.server",
		},
	}

	for _, c := range cases {
		got := CollapseCommand(c.cmd, c.args)
		if got != c.want {
			t.Errorf("CollapseCommand(%q, %v) = %q, want %q", c.cmd, c.args, got, c.want)
		}
	}
}

func TestInspectRemoteOAuthStatus(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("AIM_REAL_HOME", tempHome)

	profDir := filepath.Join(tempHome, "profiles", "dev")
	_ = os.MkdirAll(profDir, 0700)

	// 1. None when files don't exist
	if status := InspectRemoteOAuthStatus("https://mcp.posthog.com/mcp", profDir); status != "none" {
		t.Errorf("expected 'none', got %q", status)
	}

	// 2. Connected when valid unexpired token exists in .mcp-auth/tokens.json
	mcpAuthDir := filepath.Join(profDir, ".mcp-auth")
	_ = os.MkdirAll(mcpAuthDir, 0700)
	futureExpiry := time.Now().Add(24 * time.Hour).Unix()
	tokensJSON := fmt.Sprintf(`{
		"https://mcp.posthog.com/mcp": {
			"access_token": "valid-token",
			"expires_at": %d
		}
	}`, futureExpiry)
	_ = os.WriteFile(filepath.Join(mcpAuthDir, "tokens.json"), []byte(tokensJSON), 0600)

	if status := InspectRemoteOAuthStatus("https://mcp.posthog.com/mcp", profDir); status != "connected" {
		t.Errorf("expected 'connected', got %q", status)
	}

	// 3. Auth required when expired
	pastExpiry := time.Now().Add(-1 * time.Hour).Unix()
	tokensExpiredJSON := fmt.Sprintf(`{
		"https://mcp.posthog.com/mcp": {
			"access_token": "expired-token",
			"expires_at": %d
		}
	}`, pastExpiry)
	_ = os.WriteFile(filepath.Join(mcpAuthDir, "tokens.json"), []byte(tokensExpiredJSON), 0600)

	if status := InspectRemoteOAuthStatus("https://mcp.posthog.com/mcp", profDir); status != "auth_required" {
		t.Errorf("expected 'auth_required', got %q", status)
	}

	// 4. Antigravity nested format test
	agyTokenDir := filepath.Join(profDir, ".gemini", "antigravity-cli")
	_ = os.MkdirAll(agyTokenDir, 0700)
	agyTokensJSON := `{
		"https://mcp.cloudflare.com/mcp": {
			"client_id": "cf_id",
			"token": {
				"access_token": "cf_tok",
				"expiry": "2099-01-01T00:00:00Z"
			}
		}
	}`
	_ = os.WriteFile(filepath.Join(agyTokenDir, "mcp_oauth_tokens.json"), []byte(agyTokensJSON), 0600)

	if status := InspectRemoteOAuthStatus("https://mcp.cloudflare.com/mcp", profDir); status != "connected" {
		t.Errorf("expected 'connected' for cloudflare agy token, got %q", status)
	}
}
