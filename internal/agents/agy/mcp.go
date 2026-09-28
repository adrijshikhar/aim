package agy

import (
	"path/filepath"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/merge"
)

// agyNormalise drops the `"disabled": false` agy adds when it rewrites an entry.
func agyNormalise(v map[string]any) map[string]any {
	v = merge.DropEmpty(v)
	if d, ok := v["disabled"].(bool); ok && !d {
		delete(v, "disabled")
	}
	return v
}

func (a *Adapter) MCPCollections(profileDir, realHome string) []merge.Collection {
	return []merge.Collection{{
		Agent: "agy", Name: "mcpServers", Format: merge.JSON, Key: "mcpServers",
		HostPath:    filepath.Join(sharedConfigDir(realHome), "mcp_config.json"),
		ProfilePath: filepath.Join(profileDir, ".gemini", "config", "mcp_config.json"),
		Normalise:   agyNormalise,
	}}
}

func (a *Adapter) IsBackground(args []string) bool {
	return agents.BackgroundMatch(args, nil, []string{"remote-control"})
}
