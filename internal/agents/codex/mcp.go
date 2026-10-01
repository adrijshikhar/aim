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
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/merge"
)

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
	configPath := filepath.Join(profileDir, ".codex", "config.toml")
	entries, ok, err := merge.ReadTOMLKey(configPath, "mcp_servers")
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read mcp_servers from %s: %w", configPath, err)
	}

	serverMap := make(map[string]agents.MCPServerInfo)
	if ok {
		for _, name := range entries.Order {
			v := entries.Values[name]
			status := "enabled"
			if en, ok := v["enabled"].(bool); ok && !en {
				status = "disabled"
			}
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
			typ := "stdio"
			auth := "unsupported"
			if urlStr != "" {
				typ = "http"
				if strings.Contains(strings.ToLower(urlStr), "oauth") || strings.Contains(strings.ToLower(name), "oauth") {
					auth = "OAuth"
				}
			} else {
				target = agents.CollapseCommand(cmdStr, argsSlice)
			}

			serverMap[name] = agents.MCPServerInfo{
				Name:   name,
				Type:   typ,
				Status: status,
				Auth:   auth,
				Target: target,
				Origin: "profile",
			}
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
					status := "enabled"
					if en, ok := sv["enabled"].(bool); ok && !en {
						status = "disabled"
					}
					urlStr, _ := sv["url"].(string)
					cmdStr, _ := sv["command"].(string)
					var argsSlice []string
					if rawArgs, ok := sv["args"].([]any); ok {
						for _, arg := range rawArgs {
							if as, ok := arg.(string); ok {
								argsSlice = append(argsSlice, as)
							}
						}
					}
					target := urlStr
					typ := "stdio"
					auth := "unsupported"
					if urlStr != "" {
						typ = "http"
						if strings.Contains(strings.ToLower(urlStr), "oauth") || strings.Contains(strings.ToLower(sName), "oauth") {
							auth = "OAuth"
						}
					} else {
						target = agents.CollapseCommand(cmdStr, argsSlice)
					}
					serverMap[sName] = agents.MCPServerInfo{
						Name:   sName,
						Type:   typ,
						Status: status,
						Auth:   auth,
						Target: target,
						Origin: "plugin:" + pluginName,
					}
				}
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

