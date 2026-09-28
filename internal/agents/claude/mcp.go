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

// MCPCollections: MCP servers live in .claude.json; which plugins are on, and
// the marketplaces they come from, in settings.json. Plugin content is shared.
func (a *Adapter) MCPCollections(profileDir, realHome string) []merge.Collection {
	hostSettings := filepath.Join(realHome, ".claude", "settings.json")
	profSettings := filepath.Join(profileDir, ".claude", "settings.json")
	return []merge.Collection{{
		Agent: "claude", Name: "mcpServers", Format: merge.JSON, Key: "mcpServers",
		HostPath:    filepath.Join(realHome, ".claude.json"),
		ProfilePath: filepath.Join(profileDir, ".claude", ".claude.json"),
		Normalise:   claudeNormalise,
		Group:       "mcp",
		Noun:        "server",
	}, {
		Agent: "claude", Name: "enabledPlugins", Format: merge.JSON, Key: "enabledPlugins",
		HostPath: hostSettings, ProfilePath: profSettings,
		Group: "plugins", Noun: "plugin",
	}, {
		Agent: "claude", Name: "extraKnownMarketplaces", Format: merge.JSON, Key: "extraKnownMarketplaces",
		HostPath: hostSettings, ProfilePath: profSettings,
		Group: "plugins", Noun: "marketplace",
	}}
}

// IsBackground: `--bg`/`--background` anywhere before "--", or the background
// session subcommands.
func (a *Adapter) IsBackground(profileArgs, args []string) bool {
	return agents.ArgsMatch(profileArgs, args, []string{"--bg", "--background"}, []string{"agents", "attach", "respawn"})
}

// IsSession: -v/--version and -h/--help as `claude --help` lists them, plus
// -V, which prints the version too. There is no help or version subcommand:
// `claude version` is a prompt. A native `mcp …` subcommand is a session, so
// `mcp list` shows the merged set.
func (a *Adapter) IsSession(profileArgs, args []string) bool {
	return !agents.ArgsMatch(profileArgs, args, []string{"-v", "-V", "--version", "-h", "--help"}, nil)
}
