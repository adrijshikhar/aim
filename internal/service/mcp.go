package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
)

type mcpService struct {
	pm  *profile.ProfileManager
	reg *agents.Registry
}

// NewMCPService creates a new MCPService backed by ProfileManager and Registry.
func NewMCPService(pm *profile.ProfileManager, reg ...*agents.Registry) MCPService {
	var r *agents.Registry
	if len(reg) > 0 && reg[0] != nil {
		r = reg[0]
	} else {
		r = agents.NewRegistry()
	}
	return &mcpService{pm: pm, reg: r}
}

type rawMCPEntry struct {
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	ServerURL string            `json:"serverUrl"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Disabled  bool              `json:"disabled"`
	Enabled   *bool             `json:"enabled"`
}

type rawMCPConfig struct {
	MCPServers map[string]rawMCPEntry `json:"mcpServers"`
}

func parseMCPJSONFile(path, scope, client string) map[string]MCPServerDTO {
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
		status := "enabled"
		if entry.Disabled || (entry.Enabled != nil && !*entry.Enabled) {
			status = "disabled"
		}

		targetURL := entry.ServerURL
		if targetURL == "" {
			targetURL = entry.URL
		}

		cmd := entry.Command
		args := entry.Args
		typ := "stdio"
		auth := "none"

		if targetURL != "" {
			typ = "http"
			if strings.Contains(targetURL, "/sse") {
				typ = "sse"
			}
			if cmd == "" {
				cmd = "remote"
				args = []string{targetURL}
			}
			if strings.Contains(strings.ToLower(targetURL), "oauth") || strings.Contains(strings.ToLower(name), "oauth") {
				auth = "OAuth"
			} else if len(entry.Headers) > 0 {
				auth = "connected"
			}
		} else {
			for ek := range entry.Env {
				if strings.Contains(ek, "KEY") || strings.Contains(ek, "TOKEN") {
					auth = "connected"
					break
				}
			}
		}

		res[name] = MCPServerDTO{
			Name:    name,
			Command: cmd,
			Args:    args,
			Env:     entry.Env,
			URL:     targetURL,
			Headers: entry.Headers,
			Scope:   scope,
			Type:    typ,
			Status:  status,
			Auth:    auth,
			Client:  client,
		}
	}
	return res
}

func parseCodexTOMLFile(path, scope string) map[string]MCPServerDTO {
	entries, ok, err := merge.ReadTOMLKey(path, "mcp_servers")
	if err != nil || !ok || len(entries.Order) == 0 {
		return nil
	}

	res := make(map[string]MCPServerDTO, len(entries.Order))
	for _, name := range entries.Order {
		raw := entries.Values[name]
		if raw == nil {
			continue
		}

		status := "enabled"
		if en, ok := raw["enabled"].(bool); ok && !en {
			status = "disabled"
		}

		cmd, _ := raw["command"].(string)
		urlStr, _ := raw["url"].(string)

		var args []string
		if rawArgs, ok := raw["args"].([]any); ok {
			for _, a := range rawArgs {
				if s, ok := a.(string); ok {
					args = append(args, s)
				}
			}
		}

		env := make(map[string]string)
		if rawEnv, ok := raw["env"].(map[string]any); ok {
			for k, v := range rawEnv {
				env[k] = fmt.Sprint(v)
			}
		}

		headers := make(map[string]string)
		if rawHeaders, ok := raw["http_headers"].(map[string]any); ok {
			for k, v := range rawHeaders {
				headers[k] = fmt.Sprint(v)
			}
		}

		typ := "stdio"
		auth := "none"
		if urlStr != "" {
			typ = "http"
			if strings.Contains(urlStr, "/sse") {
				typ = "sse"
			}
			if cmd == "" {
				cmd = "remote"
				args = []string{urlStr}
			}
			if strings.Contains(strings.ToLower(urlStr), "oauth") || strings.Contains(strings.ToLower(name), "oauth") {
				auth = "OAuth"
			} else if len(headers) > 0 {
				auth = "connected"
			}
		} else {
			for ek := range env {
				if strings.Contains(ek, "KEY") || strings.Contains(ek, "TOKEN") {
					auth = "connected"
					break
				}
			}
		}

		res[name] = MCPServerDTO{
			Name:    name,
			Command: cmd,
			Args:    args,
			Env:     env,
			URL:     urlStr,
			Headers: headers,
			Scope:   scope,
			Type:    typ,
			Status:  status,
			Auth:    auth,
			Client:  "codex",
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
		globalSources := []struct {
			path   string
			client string
			isTOML bool
		}{
			{filepath.Join(homeDir, ".gemini", "config", "mcp_config.json"), "agy", false},
			{filepath.Join(homeDir, ".gemini", "antigravity-cli", "mcp_config.json"), "agy", false},
			{filepath.Join(homeDir, ".claude.json"), "claude", false},
			{filepath.Join(homeDir, ".claude", "mcp_config.json"), "claude", false},
			{filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), "claude", false},
			{filepath.Join(homeDir, ".codex", "config.toml"), "codex", true},
		}

		for _, src := range globalSources {
			var servers map[string]MCPServerDTO
			if src.isTOML {
				servers = parseCodexTOMLFile(src.path, "global")
			} else {
				servers = parseMCPJSONFile(src.path, "global", src.client)
			}
			for k, v := range servers {
				if _, exists := allServers[k]; !exists {
					allServers[k] = v
				}
			}
		}
	}

	// 2. Collect profile-scoped MCP servers
	var pDir string
	if profileName != "" && profileName != "all" && s.pm != nil {
		pDir = s.pm.ProfileDir(profileName)
		profileSources := []struct {
			path   string
			client string
			isTOML bool
		}{
			{filepath.Join(pDir, ".gemini", "config", "mcp_config.json"), "agy", false},
			{filepath.Join(pDir, ".gemini", "antigravity-cli", "mcp_config.json"), "agy", false},
			{filepath.Join(pDir, ".claude.json"), "claude", false},
			{filepath.Join(pDir, ".claude", ".claude.json"), "claude", false},
			{filepath.Join(pDir, ".claude", "mcp_config.json"), "claude", false},
			{filepath.Join(pDir, ".codex", "config.toml"), "codex", true},
		}

		for _, src := range profileSources {
			var servers map[string]MCPServerDTO
			if src.isTOML {
				servers = parseCodexTOMLFile(src.path, profileName)
			} else {
				servers = parseMCPJSONFile(src.path, profileName, src.client)
			}
			for k, v := range servers {
				allServers[k] = v // Profile overrides global
			}
		}
	}

	result := make([]MCPServerDTO, 0, len(allServers))
	for _, v := range allServers {
		// Recheck live OAuth status for remote servers if auth is OAuth
		if v.URL != "" && (v.Auth == "OAuth" || v.Auth == "none") {
			status := agents.InspectRemoteOAuthStatus(v.URL, pDir)
			if status == "connected" || status == "expired" {
				v.Auth = status
			}
		}
		result = append(result, v)
	}

	// Sort deterministically by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (s *mcpService) getTargetAdapters(clients []string) []agents.MCPConfigManager {
	homeDir := config.RealHomeDir()
	var targets []agents.MCPConfigManager
	wantAll := len(clients) == 0
	clientMap := make(map[string]bool)
	for _, c := range clients {
		if c == "all" {
			wantAll = true
			break
		}
		clientMap[strings.ToLower(strings.TrimSpace(c))] = true
	}

	if s.reg != nil {
		for _, a := range s.reg.All() {
			mgr, ok := a.(agents.MCPConfigManager)
			if !ok {
				continue
			}
			if wantAll {
				if mgr.DetectInstalled(homeDir) {
					targets = append(targets, mgr)
				}
			} else if clientMap[strings.ToLower(a.Name())] {
				targets = append(targets, mgr)
			}
		}
	}
	return targets
}

func (s *mcpService) AddServer(ctx context.Context, req AddMCPServerRequest) error {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return fmt.Errorf("server name is required")
	}

	cfg := agents.MCPServerConfig{
		Command: strings.TrimSpace(req.Command),
		Args:    req.Args,
		Env:     req.Env,
		URL:     strings.TrimSpace(req.URL),
		Headers: req.Headers,
	}

	if cfg.Command == "" && cfg.URL == "" {
		return fmt.Errorf("command or URL is required")
	}

	targets := s.getTargetAdapters(req.Clients)
	if len(targets) == 0 {
		return fmt.Errorf("no matching or installed MCP clients found")
	}

	var lastErr error
	added := 0
	for _, adapter := range targets {
		if req.Profile == "" || req.Profile == "global" {
			if err := adapter.AddHostMCPServer(ctx, name, cfg); err == nil {
				added++
			} else {
				lastErr = err
			}
		} else {
			if s.pm == nil || !s.pm.ProfileExists(req.Profile) {
				return fmt.Errorf("profile %q does not exist", req.Profile)
			}
			pDir := s.pm.ProfileDir(req.Profile)
			if err := adapter.AddProfileMCPServer(ctx, req.Profile, pDir, name, cfg); err == nil {
				added++
			} else {
				lastErr = err
			}
		}
	}

	if added == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}

func (s *mcpService) SetServerState(ctx context.Context, name string, req SetMCPServerStateRequest) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("server name is required")
	}

	targets := s.getTargetAdapters(req.Clients)
	if len(targets) == 0 {
		return fmt.Errorf("no matching or installed MCP clients found")
	}

	var lastErr error
	updated := 0
	for _, adapter := range targets {
		if req.Profile == "" || req.Profile == "global" {
			if err := adapter.SetHostMCPServerState(ctx, name, req.Enabled); err == nil {
				updated++
			} else if !errors.Is(err, os.ErrNotExist) {
				lastErr = err
			}
		} else {
			if s.pm == nil || !s.pm.ProfileExists(req.Profile) {
				return fmt.Errorf("profile %q does not exist", req.Profile)
			}
			pDir := s.pm.ProfileDir(req.Profile)
			if err := adapter.SetProfileMCPServerState(ctx, req.Profile, pDir, name, req.Enabled); err == nil {
				updated++
			} else if !errors.Is(err, os.ErrNotExist) {
				lastErr = err
			}
		}
	}

	if updated == 0 && lastErr != nil {
		return lastErr
	}
	if updated == 0 {
		return fmt.Errorf("server %q was not found in selected agent configs", name)
	}
	return nil
}

func (s *mcpService) RemoveServer(ctx context.Context, name string, profileName string, clients []string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("server name is required")
	}

	targets := s.getTargetAdapters(clients)
	if len(targets) == 0 {
		return fmt.Errorf("no matching or installed MCP clients found")
	}

	var lastErr error
	removed := 0
	for _, adapter := range targets {
		if profileName == "" || profileName == "global" {
			if err := adapter.RemoveHostMCPServer(ctx, name); err == nil {
				removed++
			} else if !errors.Is(err, os.ErrNotExist) {
				lastErr = err
			}
		} else {
			if s.pm == nil || !s.pm.ProfileExists(profileName) {
				return fmt.Errorf("profile %q does not exist", profileName)
			}
			pDir := s.pm.ProfileDir(profileName)
			if err := adapter.RemoveProfileMCPServer(ctx, profileName, pDir, name); err == nil {
				removed++
			} else if !errors.Is(err, os.ErrNotExist) {
				lastErr = err
			}
		}
	}

	if removed == 0 && lastErr != nil {
		return lastErr
	}
	if removed == 0 {
		return fmt.Errorf("server %q was not found in selected agent configs", name)
	}
	return nil
}
