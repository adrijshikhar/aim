package codex

import (
	"path/filepath"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/merge"
)

func (a *Adapter) MCPCollections(profileDir, realHome string) []merge.Collection {
	return []merge.Collection{{
		Agent: "codex", Name: "mcp_servers", Format: merge.TOML, Key: "mcp_servers",
		HostPath:    filepath.Join(realHome, ".codex", "config.toml"),
		ProfilePath: filepath.Join(profileDir, ".codex", "config.toml"),
	}}
}

func (a *Adapter) IsBackground(args []string) bool {
	return agents.BackgroundMatch(args, nil, []string{"app-server", "remote-control"})
}
