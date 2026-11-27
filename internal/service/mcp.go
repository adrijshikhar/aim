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

const (
	scopeAll            = "all"
	scopeGlobal         = "global"
	agentAgy            = "agy"
	agentClaude         = "claude"
	agentCodex          = "codex"
	authConnected       = "connected"
	authNone            = "none"
	hostProfileSentinel = "<host>"
)

type mcpServerKey struct {
	client string
	scope  string
	name   string
}

func parseMCPJSONFromRoot(root *os.Root, relPath, scope, client string) map[string]MCPServerDTO {
	if root == nil {
		return nil
	}
	data, err := root.ReadFile(relPath)
	if err != nil {
		return nil
	}
	return parseMCPJSONBytes(data, scope, client)
}

func parseMCPJSONFile(path, scope, client string) map[string]MCPServerDTO {
	//nolint:gosec // G304: host/global MCP configuration path constructed from fixed host home locations.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return parseMCPJSONBytes(data, scope, client)
}

func parseMCPJSONBytes(data []byte, scope, client string) map[string]MCPServerDTO {

	var parsed rawMCPConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil
	}

	if len(parsed.MCPServers) == 0 {
		return nil
	}

	res := make(map[string]MCPServerDTO, len(parsed.MCPServers))
	for name, entry := range parsed.MCPServers {
		res[name] = buildMCPServerDTO(name, entry, scope, client)
	}
	return res
}

func buildMCPServerDTO(name string, entry rawMCPEntry, scope, client string) MCPServerDTO {
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
	auth := detectMCPAuth(name, targetURL, entry.Headers, entry.Env)

	if targetURL != "" {
		typ = "http"
		if strings.Contains(targetURL, "/sse") {
			typ = "sse"
		}
		if cmd == "" {
			cmd = "remote"
			args = []string{targetURL}
		}
	}

	return MCPServerDTO{
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

func detectMCPAuth(name, targetURL string, headers, env map[string]string) string {
	if targetURL != "" {
		if strings.Contains(strings.ToLower(targetURL), "oauth") || strings.Contains(strings.ToLower(name), "oauth") {
			return "OAuth"
		}
		if len(headers) > 0 {
			return authConnected
		}
		return authNone
	}
	for ek := range env {
		if strings.Contains(ek, "KEY") || strings.Contains(ek, "TOKEN") {
			return authConnected
		}
	}
	return authNone
}

func parseCodexTOMLFile(path, scope string) map[string]MCPServerDTO {
	entries, ok, err := merge.ReadTOMLKey(path, "mcp_servers")
	if err != nil || !ok || len(entries.Order) == 0 {
		return nil
	}
	return parseCodexTOMLEntries(entries, scope)
}

func parseCodexTOMLData(data []byte, scope string) map[string]MCPServerDTO {
	entries, ok, err := merge.ReadTOMLKeyFromBytes(data, "mcp_servers")
	if err != nil || !ok || len(entries.Order) == 0 {
		return nil
	}
	return parseCodexTOMLEntries(entries, scope)
}

func parseCodexTOMLEntries(entries merge.Entries, scope string) map[string]MCPServerDTO {
	res := make(map[string]MCPServerDTO, len(entries.Order))
	for _, name := range entries.Order {
		raw := entries.Values[name]
		if raw == nil {
			continue
		}
		res[name] = parseCodexTOMLEntry(name, raw, scope)
	}
	return res
}

func parseCodexTOMLEntry(name string, raw map[string]any, scope string) MCPServerDTO {
	status := "enabled"
	if en, ok := raw["enabled"].(bool); ok && !en {
		status = "disabled"
	}

	cmd, _ := raw["command"].(string)
	urlStr, _ := raw["url"].(string)
	args := extractTOMLStringSlice(raw["args"])
	env := extractTOMLStringMap(raw["env"])
	headers := extractTOMLStringMap(raw["http_headers"])

	typ := "stdio"
	if urlStr != "" {
		typ = "http"
		if strings.Contains(urlStr, "/sse") {
			typ = "sse"
		}
		if cmd == "" {
			cmd = "remote"
			args = []string{urlStr}
		}
	}

	return MCPServerDTO{
		Name:    name,
		Command: cmd,
		Args:    args,
		Env:     env,
		URL:     urlStr,
		Headers: headers,
		Scope:   scope,
		Type:    typ,
		Status:  status,
		Auth:    detectMCPAuth(name, urlStr, headers, env),
		Client:  "codex",
	}
}

func extractTOMLStringSlice(v any) []string {
	rawArgs, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, a := range rawArgs {
		if s, ok := a.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func extractTOMLStringMap(v any) map[string]string {
	rawMap, ok := v.(map[string]any)
	out := make(map[string]string)
	if !ok {
		return out
	}
	for k, val := range rawMap {
		out[k] = fmt.Sprint(val)
	}
	return out
}

func collectGlobalMCPServers(homeDir string, dst map[mcpServerKey]MCPServerDTO) {
	if homeDir == "" {
		return
	}
	globalSources := []struct {
		path   string
		client string
		isTOML bool
	}{
		{filepath.Join(homeDir, ".gemini", "config", "mcp_config.json"), agentAgy, false},
		{filepath.Join(homeDir, ".gemini", "antigravity-cli", "mcp_config.json"), agentAgy, false},
		{filepath.Join(homeDir, ".claude.json"), agentClaude, false},
		{filepath.Join(homeDir, ".claude", "mcp_config.json"), agentClaude, false},
		{filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), agentClaude, false},
		{filepath.Join(homeDir, ".codex", "config.toml"), agentCodex, true},
	}

	for _, src := range globalSources {
		var servers map[string]MCPServerDTO
		if src.isTOML {
			servers = parseCodexTOMLFile(src.path, scopeGlobal)
		} else {
			servers = parseMCPJSONFile(src.path, scopeGlobal, src.client)
		}
		for _, v := range servers {
			key := mcpServerKey{client: v.Client, scope: v.Scope, name: v.Name}
			if _, exists := dst[key]; !exists {
				dst[key] = v
			}
		}
	}
}

func (s *mcpService) collectProfileMCPServers(profileName string, dst map[mcpServerKey]MCPServerDTO) string {
	if profileName == "" || profileName == scopeAll || profileName == scopeGlobal || s.pm == nil {
		return ""
	}
	pDir := s.pm.ProfileDir(profileName)
	profilesRoot, err := os.OpenRoot(s.pm.ProfilesRoot())
	if err != nil {
		return pDir
	}
	defer func() { _ = profilesRoot.Close() }()
	profRoot, err := profilesRoot.OpenRoot(profileName)
	if err != nil {
		return pDir
	}
	defer func() { _ = profRoot.Close() }()

	profileSources := []struct {
		relPath string
		client  string
		isTOML  bool
	}{
		{filepath.Join(".gemini", "config", "mcp_config.json"), agentAgy, false},
		{filepath.Join(".gemini", "antigravity-cli", "mcp_config.json"), agentAgy, false},
		{".claude.json", agentClaude, false},
		{filepath.Join(".claude", ".claude.json"), agentClaude, false},
		{filepath.Join(".claude", "mcp_config.json"), agentClaude, false},
		{filepath.Join(".codex", "config.toml"), agentCodex, true},
	}

	for _, src := range profileSources {
		var servers map[string]MCPServerDTO
		if src.isTOML {
			if data, readErr := profRoot.ReadFile(src.relPath); readErr == nil {
				servers = parseCodexTOMLData(data, profileName)
			}
		} else {
			servers = parseMCPJSONFromRoot(profRoot, src.relPath, profileName, src.client)
		}
		for _, v := range servers {
			key := mcpServerKey{client: v.Client, scope: v.Scope, name: v.Name}
			dst[key] = v
		}
	}
	return pDir
}

func shouldIncludeGlobalMCP(profileName string) (bool, error) {
	if profileName == "" || profileName == scopeAll || profileName == scopeGlobal {
		return true, nil
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return false, fmt.Errorf("load config: %w", err)
	}
	if cfg == nil || cfg.Profiles == nil {
		return true, nil
	}
	profCfg, ok := cfg.Profiles[profileName]
	return !ok || profCfg.MCPGlobal == nil || *profCfg.MCPGlobal, nil
}

func (s *mcpService) ListServers(_ context.Context, profileName string) ([]MCPServerDTO, error) {
	homeDir := config.RealHomeDir()
	allServers := make(map[mcpServerKey]MCPServerDTO)

	profileName = strings.TrimSpace(profileName)
	if profileName != "" && profileName != scopeAll && profileName != scopeGlobal {
		if err := profile.ValidateProfileName(profileName); err != nil {
			return nil, err
		}
	}

	includeGlobal, err := shouldIncludeGlobalMCP(profileName)
	if err != nil {
		return nil, err
	}
	if includeGlobal {
		collectGlobalMCPServers(homeDir, allServers)
	}
	pDir := s.collectProfileMCPServers(profileName, allServers)

	result := make([]MCPServerDTO, 0, len(allServers))
	for _, v := range allServers {
		oauthDir := pDir
		if v.Scope == scopeGlobal && homeDir != "" {
			oauthDir = homeDir
		}
		result = append(result, enrichMCPServerOAuth(v, oauthDir))
	}
	sortMCPServers(result)
	return result, nil
}

func enrichMCPServerOAuth(v MCPServerDTO, pDir string) MCPServerDTO {
	if v.URL != "" && (v.Auth == "OAuth" || v.Auth == authNone) {
		switch agents.InspectRemoteOAuthStatus(v.URL, pDir) {
		case authConnected:
			v.Auth = authConnected
		case "auth_required", "expired":
			v.Auth = "auth required"
		}
	}
	return v
}

func sortMCPServers(result []MCPServerDTO) {
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		if result[i].Client != result[j].Client {
			return result[i].Client < result[j].Client
		}
		return result[i].Scope < result[j].Scope
	})
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

func (s *mcpService) resolveMCPMutationTargets(rawProfile string, clients []string) (string, []agents.MCPConfigManager, error) {
	profileName := strings.TrimSpace(rawProfile)
	if profileName != "" && profileName != scopeGlobal {
		if err := profile.ValidateProfileName(profileName); err != nil {
			return "", nil, err
		}
		if s.pm == nil || !s.pm.ProfileExists(profileName) {
			return "", nil, fmt.Errorf("profile %q does not exist", profileName)
		}
	}
	targets := s.getTargetAdapters(clients)
	if len(targets) == 0 {
		return "", nil, errors.New("no matching or installed MCP clients found")
	}
	return profileName, targets, nil
}

func (s *mcpService) applyMCPMutation(
	profileName string,
	targets []agents.MCPConfigManager,
	ignoreNotExist bool,
	hostFn func(agents.MCPConfigManager) error,
	profileFn func(agents.MCPConfigManager, string) error,
) (int, error) {
	isHost := profileName == "" || profileName == scopeGlobal
	var pDir string
	if !isHost {
		pDir = s.pm.ProfileDir(profileName)
	}
	var errs []error
	applied := 0
	for _, adapter := range targets {
		var err error
		if isHost {
			err = hostFn(adapter)
		} else {
			err = profileFn(adapter, pDir)
		}
		if err == nil {
			applied++
		} else if !ignoreNotExist || !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	if applied == 0 && len(errs) > 0 {
		return 0, errors.Join(errs...)
	}
	return applied, nil
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
	profileName, targets, err := s.resolveMCPMutationTargets(req.Profile, req.Clients)
	if err != nil {
		return err
	}
	_, err = s.applyMCPMutation(profileName, targets, false,
		func(a agents.MCPConfigManager) error { return a.AddHostMCPServer(ctx, name, cfg) },
		func(a agents.MCPConfigManager, pDir string) error {
			return a.AddProfileMCPServer(ctx, profileName, pDir, name, cfg)
		},
	)
	return err
}

func (s *mcpService) SetServerState(ctx context.Context, name string, req SetMCPServerStateRequest) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("server name is required")
	}
	profileName, targets, err := s.resolveMCPMutationTargets(req.Profile, req.Clients)
	if err != nil {
		return err
	}
	updated, err := s.applyMCPMutation(profileName, targets, true,
		func(a agents.MCPConfigManager) error { return a.SetHostMCPServerState(ctx, name, req.Enabled) },
		func(a agents.MCPConfigManager, pDir string) error {
			return a.SetProfileMCPServerState(ctx, profileName, pDir, name, req.Enabled)
		},
	)
	if err != nil {
		return err
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
	trimmedProfile, targets, err := s.resolveMCPMutationTargets(profileName, clients)
	if err != nil {
		return err
	}
	removed, err := s.applyMCPMutation(trimmedProfile, targets, true,
		func(a agents.MCPConfigManager) error { return a.RemoveHostMCPServer(ctx, name) },
		func(a agents.MCPConfigManager, pDir string) error {
			return a.RemoveProfileMCPServer(ctx, trimmedProfile, pDir, name)
		},
	)
	if err != nil {
		return err
	}
	if removed == 0 {
		return fmt.Errorf("server %q was not found in selected agent configs", name)
	}
	return nil
}
