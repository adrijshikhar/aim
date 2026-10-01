package claude

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
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

// ListMCPServers returns configured MCP servers for Claude Code in the profile.
// It inspects both the profile's .claude/.claude.json and host ~/.claude.json.
func (a *Adapter) ListMCPServers(ctx context.Context, profileName, profileDir string) ([]agents.MCPServerInfo, error) {
	serverMap := make(map[string]agents.MCPServerInfo)

	// 1. Read host ~/.claude.json first
	realHome := config.RealHomeDir()
	hostJSON := filepath.Join(realHome, ".claude.json")
	if hostEntries, ok, err := merge.ReadJSONKey(hostJSON, "mcpServers"); err == nil && ok {
		for _, name := range hostEntries.Order {
			info := parseClaudeMCPServer(name, hostEntries.Values[name], "host")
			serverMap[name] = info
		}
	}

	// 2. Read profile .claude/.claude.json (overrides host)
	profJSON := filepath.Join(profileDir, ".claude", ".claude.json")
	if profEntries, ok, err := merge.ReadJSONKey(profJSON, "mcpServers"); err == nil && ok {
		for _, name := range profEntries.Order {
			info := parseClaudeMCPServer(name, profEntries.Values[name], "profile")
			serverMap[name] = info
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

func parseClaudeMCPServer(name string, v map[string]any, origin string) agents.MCPServerInfo {
	status := "enabled"
	if dis, ok := v["disabled"].(bool); ok && dis {
		status = "disabled"
	}

	typ, _ := v["type"].(string)
	urlStr, _ := v["url"].(string)
	cmdStr, _ := v["command"].(string)

	var argsSlice []string
	if rawArgs, ok := v["args"].([]any); ok {
		for _, arg := range rawArgs {
			if as, ok := arg.(string); ok {
				argsSlice = append(argsSlice, as)
			}
		}
	}

	target := urlStr
	auth := "unsupported"
	if urlStr != "" {
		if typ == "" {
			typ = "http"
		}
		if strings.Contains(strings.ToLower(urlStr), "oauth") || strings.Contains(strings.ToLower(name), "oauth") {
			auth = "OAuth"
		} else if headers, ok := v["headers"].(map[string]any); ok && len(headers) > 0 {
			auth = "connected"
		}
	} else {
		if typ == "" {
			typ = "stdio"
		}
		target = agents.CollapseCommand(cmdStr, argsSlice)
		if env, ok := v["env"].(map[string]any); ok {
			for ek := range env {
				if strings.Contains(ek, "KEY") || strings.Contains(ek, "TOKEN") {
					auth = "connected"
					break
				}
			}
		}
	}

	return agents.MCPServerInfo{
		Name:   name,
		Type:   typ,
		Status: status,
		Auth:   auth,
		Target: target,
		Origin: origin,
	}
}
