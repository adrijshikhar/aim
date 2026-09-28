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
	return agents.ArgsMatch(args, []string{"--bg", "--background"}, []string{"agents", "attach", "respawn"})
}

// IsSession: -v/--version and -h/--help as `claude --help` lists them, plus
// -V, which prints the version too. There is no help or version subcommand:
// `claude version` is a prompt. A native `mcp …` subcommand is a session, so
// `mcp list` shows the merged set.
func (a *Adapter) IsSession(args []string) bool {
	return !agents.ArgsMatch(args, []string{"-v", "-V", "--version", "-h", "--help"}, nil)
}
