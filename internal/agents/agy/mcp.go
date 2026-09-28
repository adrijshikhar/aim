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

func (a *Adapter) IsBackground(profileArgs, args []string) bool {
	return agents.ArgsMatch(profileArgs, args, nil, []string{"remote-control"})
}

// IsSession: agy parses Go-style flags, so -version and -help work beside
// --version, -h and --help; `help` is a subcommand. -v is its log-verbosity
// flag, not a version one, and there is no version subcommand.
func (a *Adapter) IsSession(profileArgs, args []string) bool {
	return !agents.ArgsMatch(profileArgs, args, []string{"--version", "-version", "-h", "--help", "-help"}, []string{"help"})
}
