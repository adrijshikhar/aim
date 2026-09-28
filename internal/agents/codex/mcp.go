package codex

import (
	"path/filepath"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/merge"
)

// MCPCollections: servers and plugins share config.toml. A [plugins."<id>"]
// entry carries its enabled flag and any [plugins."<id>".mcp_servers.*]
// sub-tables, which move with the plugin, not with mcp_servers.
func (a *Adapter) MCPCollections(profileDir, realHome string) []merge.Collection {
	host := filepath.Join(realHome, ".codex", "config.toml")
	prof := filepath.Join(profileDir, ".codex", "config.toml")
	return []merge.Collection{{
		Agent: "codex", Name: "mcp_servers", Format: merge.TOML, Key: "mcp_servers",
		HostPath: host, ProfilePath: prof,
		Group: "mcp", Noun: "server",
	}, {
		Agent: "codex", Name: "plugins", Format: merge.TOML, Key: "plugins",
		HostPath: host, ProfilePath: prof,
		Group: "plugins", Noun: "plugin",
	}}
}

func (a *Adapter) IsBackground(args []string) bool {
	return agents.BackgroundMatch(args, nil, []string{"app-server", "remote-control"})
}
