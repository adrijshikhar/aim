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

func (a *Adapter) IsBackground(profileArgs, args []string) bool {
	return agents.ArgsMatch(profileArgs, args, nil, []string{"app-server", "remote-control"})
}

// IsSession: -V/--version, -h/--help and the help subcommand. codex rejects
// -v and has no version subcommand, so both reach codex as a session attempt.
func (a *Adapter) IsSession(profileArgs, args []string) bool {
	return !agents.ArgsMatch(profileArgs, args, []string{"-V", "--version", "-h", "--help"}, []string{"help"})
}
