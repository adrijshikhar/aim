package agents_test

import (
	"testing"

	"github.com/aim-cli/aim/internal/agents"
)

func TestParseMCPServerMap_URLBased(t *testing.T) {
	cases := []struct {
		name     string
		input    map[string]any
		expected agents.MCPServerInfo
	}{
		{
			name: "http oauth in url",
			input: map[string]any{
				"url": "https://api.example.com/oauth/mcp",
			},
			expected: agents.MCPServerInfo{
				Name:   "server1",
				Type:   "http",
				Status: "enabled",
				Auth:   "OAuth",
				Target: "https://api.example.com/oauth/mcp",
				Origin: "profile",
			},
		},
		{
			name: "agy serverUrl with headers",
			input: map[string]any{
				"serverUrl": "https://mcp.internal.net",
				"headers": map[string]any{
					"Authorization": "Bearer 123",
				},
				"disabled": true,
			},
			expected: agents.MCPServerInfo{
				Name:   "server2",
				Type:   "http",
				Status: "disabled",
				Auth:   "connected",
				Target: "https://mcp.internal.net",
				Origin: "host",
			},
		},
		{
			name: "explicit sse type and disabled via enabled: false",
			input: map[string]any{
				"type":    "sse",
				"url":     "https://sse.example.com",
				"enabled": false,
			},
			expected: agents.MCPServerInfo{
				Name:   "server3",
				Type:   "sse",
				Status: "disabled",
				Auth:   "unsupported",
				Target: "https://sse.example.com",
				Origin: "plugin:test",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := agents.ParseMCPServerMap(tc.expected.Name, tc.input, tc.expected.Origin)
			if got != tc.expected {
				t.Fatalf("mismatch:\nexpected: %+v\ngot:      %+v", tc.expected, got)
			}
		})
	}
}

func TestParseMCPServerMap_CommandBased(t *testing.T) {
	cases := []struct {
		name     string
		input    map[string]any
		expected agents.MCPServerInfo
	}{
		{
			name: "stdio with env token",
			input: map[string]any{
				"command": "npx",
				"args":    []any{"-y", "@modelcontextprotocol/server-postgres"},
				"env": map[string]any{
					"PG_API_KEY": "secret",
				},
			},
			expected: agents.MCPServerInfo{
				Name:   "postgres",
				Type:   "stdio",
				Status: "enabled",
				Auth:   "connected",
				Target: "npx -y @modelcontextprotocol/server-postgres",
				Origin: "profile",
			},
		},
		{
			name: "stdio with auth_status",
			input: map[string]any{
				"command":     "python",
				"args":        []string{"-m", "server"},
				"auth_status": "connected",
			},
			expected: agents.MCPServerInfo{
				Name:   "py-server",
				Type:   "stdio",
				Status: "enabled",
				Auth:   "OAuth",
				Target: "python -m server",
				Origin: "profile",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := agents.ParseMCPServerMap(tc.expected.Name, tc.input, tc.expected.Origin)
			if got != tc.expected {
				t.Fatalf("mismatch:\nexpected: %+v\ngot:      %+v", tc.expected, got)
			}
		})
	}
}
