package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

type mcpService struct {
	pm *profile.ProfileManager
}

// NewMCPService creates a new MCPService backed by ProfileManager.
func NewMCPService(pm *profile.ProfileManager) MCPService {
	return &mcpService{pm: pm}
}

type rawMCPEntry struct {
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	ServerURL string            `json:"serverUrl"`
	URL       string            `json:"url"`
	Disabled  bool              `json:"disabled"`
}

type rawMCPConfig struct {
	MCPServers map[string]rawMCPEntry `json:"mcpServers"`
}

func parseMCPFile(path string, scope string) map[string]MCPServerDTO {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var parsed rawMCPConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil
	}

	if len(parsed.MCPServers) == 0 {
		return nil
	}

	res := make(map[string]MCPServerDTO, len(parsed.MCPServers))
	for name, entry := range parsed.MCPServers {
		if entry.Disabled {
			continue
		}

		cmd := entry.Command
		args := entry.Args
		if cmd == "" {
			targetURL := entry.ServerURL
			if targetURL == "" {
				targetURL = entry.URL
			}
			if targetURL != "" {
				cmd = "remote"
				args = []string{targetURL}
			}
		}

		res[name] = MCPServerDTO{
			Name:    name,
			Command: cmd,
			Args:    args,
			Env:     entry.Env,
			Scope:   scope,
		}
	}
	return res
}

func (s *mcpService) ListServers(ctx context.Context, profileName string) ([]MCPServerDTO, error) {
	homeDir := config.RealHomeDir()
	allServers := make(map[string]MCPServerDTO)

	profileName = strings.TrimSpace(profileName)
	includeGlobal := true

	if profileName != "" && profileName != "all" {
		cfg, _ := config.LoadConfig()
		if cfg != nil && cfg.Profiles != nil {
			if profCfg, ok := cfg.Profiles[profileName]; ok {
				if profCfg.MCPGlobal != nil && !*profCfg.MCPGlobal {
					includeGlobal = false
				}
			}
		}
	}

	// 1. Collect global MCP servers if enabled
	if includeGlobal && homeDir != "" {
		globalPaths := []string{
			filepath.Join(homeDir, ".gemini", "config", "mcp_config.json"),
			filepath.Join(homeDir, ".claude.json"),
			filepath.Join(homeDir, ".claude", "mcp_config.json"),
			filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
		}
		for _, gp := range globalPaths {
			servers := parseMCPFile(gp, "global")
			for k, v := range servers {
				if _, exists := allServers[k]; !exists {
					allServers[k] = v
				}
			}
		}
	}

	// 2. Collect profile-scoped MCP servers
	if profileName != "" && profileName != "all" && s.pm != nil {
		pDir := s.pm.ProfileDir(profileName)
		profilePaths := []string{
			filepath.Join(pDir, ".gemini", "config", "mcp_config.json"),
			filepath.Join(pDir, ".claude.json"),
			filepath.Join(pDir, ".claude", "mcp_config.json"),
		}
		for _, pp := range profilePaths {
			servers := parseMCPFile(pp, "profile")
			for k, v := range servers {
				// Profile servers override global servers
				allServers[k] = v
			}
		}
	}

	result := make([]MCPServerDTO, 0, len(allServers))
	for _, v := range allServers {
		result = append(result, v)
	}

	// Sort deterministically by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}
