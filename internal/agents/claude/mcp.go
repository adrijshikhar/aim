package claude

import (
	"path/filepath"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/merge"
)

// claudeNormalise drops what Claude adds when it rewrites an entry (T2).
func claudeNormalise(v map[string]any) map[string]any {
	v = merge.DropEmpty(v)
	if v["type"] == "stdio" && v["command"] != nil {
		delete(v, "type")
	}
	return v
}

func (a *Adapter) MCPCollections(profileDir, realHome string) []merge.Collection {
	return []merge.Collection{{
		Agent: "claude", Name: "mcpServers", Format: merge.JSON, Key: "mcpServers",
		HostPath:    filepath.Join(realHome, ".claude.json"),
		ProfilePath: filepath.Join(profileDir, ".claude", ".claude.json"),
		Normalise:   claudeNormalise,
	}}
}

// IsBackground: `--bg`/`--background` anywhere before "--", or the background
// session subcommands.
func (a *Adapter) IsBackground(args []string) bool {
	return agents.BackgroundMatch(args, []string{"--bg", "--background"}, []string{"agents", "attach", "respawn"})
}
