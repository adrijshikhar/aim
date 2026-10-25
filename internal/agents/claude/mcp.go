package claude

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
)

var _ agents.MCPConfigManager = (*Adapter)(nil)

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
			info := agents.ParseMCPServerMap(name, hostEntries.Values[name], "host")
			serverMap[name] = info
		}
	}

	// 2. Read profile .claude/.claude.json (overrides host)
	if profileDir != "" {
		profJSON := filepath.Join(profileDir, ".claude", ".claude.json")
		if profEntries, ok, err := merge.ReadJSONKey(profJSON, "mcpServers"); err == nil && ok {
			for _, name := range profEntries.Order {
				info := agents.ParseMCPServerMap(name, profEntries.Values[name], "profile")
				serverMap[name] = info
			}
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

func configToMap(cfg agents.MCPServerConfig) map[string]any {
	m := make(map[string]any)
	if cfg.Command != "" {
		m["command"] = cfg.Command
	}
	if len(cfg.Args) > 0 {
		m["args"] = cfg.Args
	}
	if len(cfg.Env) > 0 {
		m["env"] = cfg.Env
	}
	if cfg.URL != "" {
		m["url"] = cfg.URL
	}
	if cfg.Disabled {
		m["disabled"] = true
	}
	return m
}

// AddHostMCPServer registers or updates an MCP server in ~/.claude.json.
func (a *Adapter) AddHostMCPServer(ctx context.Context, name string, cfg agents.MCPServerConfig) error {
	realHome := config.RealHomeDir()
	path := filepath.Join(realHome, ".claude.json")
	entries, ok, err := merge.ReadJSONKey(path, "mcpServers")
	if err != nil || !ok {
		entries = merge.NewEntries()
	}
	entries.Set(name, configToMap(cfg), nil)
	return merge.WriteJSONKey(path, "mcpServers", entries, 0o600)
}

// RemoveHostMCPServer unregisters an MCP server from ~/.claude.json.
func (a *Adapter) RemoveHostMCPServer(ctx context.Context, name string) error {
	realHome := config.RealHomeDir()
	path := filepath.Join(realHome, ".claude.json")
	entries, ok, err := merge.ReadJSONKey(path, "mcpServers")
	if err != nil {
		return err
	}
	if !ok || !entries.Has(name) {
		return os.ErrNotExist
	}
	entries.Delete(name)
	return merge.WriteJSONKey(path, "mcpServers", entries, 0o600)
}

// AddProfileMCPServer registers or updates an MCP server in <profileDir>/.claude/.claude.json.
func (a *Adapter) AddProfileMCPServer(ctx context.Context, profileName, profileDir, name string, cfg agents.MCPServerConfig) error {
	if profileDir == "" {
		return fmt.Errorf("profileDir is required")
	}
	path := filepath.Join(profileDir, ".claude", ".claude.json")
	entries, ok, err := merge.ReadJSONKey(path, "mcpServers")
	if err != nil || !ok {
		entries = merge.NewEntries()
	}
	entries.Set(name, configToMap(cfg), nil)
	return merge.WriteJSONKey(path, "mcpServers", entries, 0o600)
}

// RemoveProfileMCPServer unregisters an MCP server from <profileDir>/.claude/.claude.json.
func (a *Adapter) RemoveProfileMCPServer(ctx context.Context, profileName, profileDir, name string) error {
	if profileDir == "" {
		return fmt.Errorf("profileDir is required")
	}
	path := filepath.Join(profileDir, ".claude", ".claude.json")
	entries, ok, err := merge.ReadJSONKey(path, "mcpServers")
	if err != nil {
		return err
	}
	if !ok || !entries.Has(name) {
		return os.ErrNotExist
	}
	entries.Delete(name)
	return merge.WriteJSONKey(path, "mcpServers", entries, 0o600)
}

// DetectInstalled checks if claude executable is in PATH or if Claude configuration directories exist.
func (a *Adapter) DetectInstalled(realHome string) bool {
	if _, err := exec.LookPath("claude"); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(realHome, ".claude.json")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(realHome, ".claude")); err == nil {
		return true
	}
	return false
}
