package agy

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
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
		Group:       "mcp",
		Noun:        "server",
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

// ListMCPServers returns configured MCP servers for Antigravity in the profile.
func (a *Adapter) ListMCPServers(ctx context.Context, profileName, profileDir string) ([]agents.MCPServerInfo, error) {
	serverMap := make(map[string]agents.MCPServerInfo)

	// 1. Read host mcp_config.json
	realHome := config.RealHomeDir()
	hostJSON := filepath.Join(sharedConfigDir(realHome), "mcp_config.json")
	if hostEntries, ok, err := merge.ReadJSONKey(hostJSON, "mcpServers"); err == nil && ok {
		for _, name := range hostEntries.Order {
			serverMap[name] = agents.ParseMCPServerMap(name, hostEntries.Values[name], "host")
		}
	}

	// 2. Read profile mcp_config.json (overrides host)
	profJSON := filepath.Join(profileDir, ".gemini", "config", "mcp_config.json")
	if profEntries, ok, err := merge.ReadJSONKey(profJSON, "mcpServers"); err == nil && ok {
		for _, name := range profEntries.Order {
			serverMap[name] = agents.ParseMCPServerMap(name, profEntries.Values[name], "profile")
		}
	}

	var res []agents.MCPServerInfo
	for _, s := range serverMap {
		res = append(res, s)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res, nil
}
