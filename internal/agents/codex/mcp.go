package codex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
)

var _ agents.MCPConfigManager = (*Adapter)(nil)

// MCPCollections: servers and plugins share config.toml. A [plugins."<id>"]
// entry carries its enabled flag and any [plugins."<id>".mcp_servers.*]
// sub-tables, which move with the plugin, not with mcp_servers.
func (a *Adapter) MCPCollections(profileDir, realHome string) []merge.Collection {
	host := filepath.Join(realHome, ".codex", "config.toml")
	prof := filepath.Join(profileDir, ".codex", "config.toml")
	return []merge.Collection{{
		Agent: "codex", Name: "mcp_servers", Format: merge.TOML, Key: "mcp_servers",
		HostPath: host, ProfilePath: prof,
		Group: "mcp", Noun: "server",
	}, {
		Agent: "codex", Name: "plugins", Format: merge.TOML, Key: "plugins",
		HostPath: host, ProfilePath: prof,
		Group: "plugins", Noun: "plugin",
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

type codexTransportJSON struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	URL     string   `json:"url"`
}

type codexServerJSON struct {
	Name       string             `json:"name"`
	Enabled    bool               `json:"enabled"`
	Transport  codexTransportJSON `json:"transport"`
	AuthStatus string             `json:"auth_status"`
}

// ListMCPServers returns all MCP servers configured for the profile.
// It queries `codex mcp list --json` first for live auth status and enablement,
// and falls back to inspecting the merged config.toml if codex CLI is not present.
func (a *Adapter) ListMCPServers(ctx context.Context, profileName, profileDir string) ([]agents.MCPServerInfo, error) {
	codexDir := filepath.Join(profileDir, ".codex")
	bin := a.ResolveBinary()
	if bin != "" {
		timeoutCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()

		cmd := exec.CommandContext(timeoutCtx, bin, "mcp", "list", "--json")
		cmd.Dir = profileDir
		cmd.Env = append(os.Environ(),
			"HOME="+profileDir,
			"CODEX_HOME="+codexDir,
			"AIM_AGENT=codex",
			"AIM_PROFILE="+profileName,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil {
			var parsed []codexServerJSON
			if jsonErr := json.Unmarshal(stdout.Bytes(), &parsed); jsonErr == nil && len(parsed) > 0 {
				var res []agents.MCPServerInfo
				for _, s := range parsed {
					status := "enabled"
					if !s.Enabled {
						status = "disabled"
					}
					target := s.Transport.URL
					if target == "" {
						target = agents.CollapseCommand(s.Transport.Command, s.Transport.Args)
					}
					typ := s.Transport.Type
					if typ == "" {
						if s.Transport.URL != "" {
							typ = "http"
						} else {
							typ = "stdio"
						}
					}
					res = append(res, agents.MCPServerInfo{
						Name:   s.Name,
						Type:   typ,
						Status: status,
						Auth:   agents.NormalizeAuth(s.AuthStatus),
						Target: target,
					})
				}
				sort.Slice(res, func(i, j int) bool {
					return res[i].Name < res[j].Name
				})
				return res, nil
			}
		}
	}

	return a.listMCPServersFromConfig(profileDir)
}

func (a *Adapter) listMCPServersFromConfig(profileDir string) ([]agents.MCPServerInfo, error) {
	serverMap := make(map[string]agents.MCPServerInfo)

	// 1. Read host ~/.codex/config.toml
	realHome := config.RealHomeDir()
	hostConfig := filepath.Join(realHome, ".codex", "config.toml")
	readCodexConfigFile(hostConfig, "host", serverMap)

	// 2. Read profile .codex/config.toml (overrides host)
	if profileDir != "" {
		profConfig := filepath.Join(profileDir, ".codex", "config.toml")
		readCodexConfigFile(profConfig, "profile", serverMap)
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

func readCodexConfigFile(configPath string, scope string, serverMap map[string]agents.MCPServerInfo) {
	entries, ok, err := merge.ReadTOMLKey(configPath, "mcp_servers")
	if err == nil && ok {
		for _, name := range entries.Order {
			serverMap[name] = agents.ParseMCPServerMap(name, entries.Values[name], scope)
		}
	}

	// Read plugins table
	pluginEntries, ok, err := merge.ReadTOMLKey(configPath, "plugins")
	if err == nil && ok {
		for _, pluginName := range pluginEntries.Order {
			pv := pluginEntries.Values[pluginName]
			if pms, ok := pv["mcp_servers"].(map[string]any); ok {
				for sName, sVal := range pms {
					if _, exists := serverMap[sName]; exists {
						continue
					}
					sv, ok := sVal.(map[string]any)
					if !ok {
						continue
					}
					serverMap[sName] = agents.ParseMCPServerMap(sName, sv, "plugin:"+pluginName)
				}
			}
		}
	}
}

func codexConfigToMap(cfg agents.MCPServerConfig) map[string]any {
	m := make(map[string]any)
	if cfg.Command != "" {
		m["command"] = cfg.Command
	}
	if len(cfg.Args) > 0 {
		args := make([]any, len(cfg.Args))
		for i, a := range cfg.Args {
			args[i] = a
		}
		m["args"] = args
	}
	if len(cfg.Env) > 0 {
		env := make(map[string]any, len(cfg.Env))
		for k, v := range cfg.Env {
			env[k] = v
		}
		m["env"] = env
	}
	if cfg.URL != "" {
		m["url"] = cfg.URL
	}
	if cfg.Disabled {
		m["enabled"] = false
	}
	return m
}

// AddHostMCPServer registers or updates an MCP server in ~/.codex/config.toml.
func (a *Adapter) AddHostMCPServer(ctx context.Context, name string, cfg agents.MCPServerConfig) error {
	realHome := config.RealHomeDir()
	path := filepath.Join(realHome, ".codex", "config.toml")
	entries, ok, err := merge.ReadTOMLKey(path, "mcp_servers")
	if err != nil || !ok {
		entries = merge.NewEntries()
	}
	entries.Set(name, codexConfigToMap(cfg), nil)
	_, err = merge.WriteTOMLKey(path, "mcp_servers", entries, 0o600)
	return err
}

// RemoveHostMCPServer unregisters an MCP server from ~/.codex/config.toml.
func (a *Adapter) RemoveHostMCPServer(ctx context.Context, name string) error {
	realHome := config.RealHomeDir()
	path := filepath.Join(realHome, ".codex", "config.toml")
	entries, ok, err := merge.ReadTOMLKey(path, "mcp_servers")
	if err != nil || !ok {
		return nil
	}
	entries.Delete(name)
	_, err = merge.WriteTOMLKey(path, "mcp_servers", entries, 0o600)
	return err
}

// AddProfileMCPServer registers or updates an MCP server in <profileDir>/.codex/config.toml.
func (a *Adapter) AddProfileMCPServer(ctx context.Context, profileName, profileDir, name string, cfg agents.MCPServerConfig) error {
	if profileDir == "" {
		return fmt.Errorf("profileDir is required")
	}
	path := filepath.Join(profileDir, ".codex", "config.toml")
	entries, ok, err := merge.ReadTOMLKey(path, "mcp_servers")
	if err != nil || !ok {
		entries = merge.NewEntries()
	}
	entries.Set(name, codexConfigToMap(cfg), nil)
	_, err = merge.WriteTOMLKey(path, "mcp_servers", entries, 0o600)
	return err
}

// RemoveProfileMCPServer unregisters an MCP server from <profileDir>/.codex/config.toml.
func (a *Adapter) RemoveProfileMCPServer(ctx context.Context, profileName, profileDir, name string) error {
	if profileDir == "" {
		return fmt.Errorf("profileDir is required")
	}
	path := filepath.Join(profileDir, ".codex", "config.toml")
	entries, ok, err := merge.ReadTOMLKey(path, "mcp_servers")
	if err != nil || !ok {
		return nil
	}
	entries.Delete(name)
	_, err = merge.WriteTOMLKey(path, "mcp_servers", entries, 0o600)
	return err
}

// DetectInstalled reports whether codex is in PATH or ~/.codex/config.toml exists.
func (a *Adapter) DetectInstalled(realHome string) bool {
	if _, err := exec.LookPath("codex"); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(realHome, ".codex", "config.toml")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(realHome, ".codex")); err == nil {
		return true
	}
	return false
}
