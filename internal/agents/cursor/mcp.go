package cursor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
)

var (
	_ agents.MCPConfigManager = (*Adapter)(nil)
	_ agents.MCPProvider      = (*Adapter)(nil)
)

// IsBackground reports whether the native args start a background session.
// Cursor CLI opens windows directly or via IPC and does not have a background daemon mode.
func (a *Adapter) IsBackground(profileArgs, args []string) bool {
	return false
}

// IsSession returns true if the invocation starts a Cursor editing session rather than printing help or version.
func (a *Adapter) IsSession(profileArgs, args []string) bool {
	return !agents.ArgsMatch(profileArgs, args, []string{"-v", "--version", "-h", "--help"}, nil)
}

// MCPCollections specifies how Cursor MCP config merges between host and profile.
func (a *Adapter) MCPCollections(profileDir, realHome string) []merge.Collection {
	host := filepath.Join(realHome, ".cursor", "mcp.json")
	prof := filepath.Join(profileDir, ".cursor", "mcp.json")
	return []merge.Collection{{
		Agent: "cursor", Name: "mcpServers", Format: merge.JSON, Key: "mcpServers",
		HostPath: host, ProfilePath: prof,
		Group: "mcp", Noun: "server",
	}}
}

// HostMCPConfigPath returns the path to Cursor's host mcp.json.
func HostMCPConfigPath(realHome string) string {
	primary := filepath.Join(realHome, ".cursor", "mcp.json")
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	storagePath := filepath.Join(realHome, "Library", "Application Support", "Cursor", "User", "globalStorage", "cursor.mcp", "mcp.json")
	if _, err := os.Stat(storagePath); err == nil {
		return storagePath
	}
	return primary
}

// ProfileMCPConfigPath returns the path to Cursor's profile-scoped mcp.json.
func ProfileMCPConfigPath(profileDir string) string {
	return filepath.Join(profileDir, ".cursor", "mcp.json")
}

// ListMCPServers returns all MCP servers configured for Cursor across host and profile.
func (a *Adapter) ListMCPServers(ctx context.Context, profileName, profileDir string) ([]agents.MCPServerInfo, error) {
	serverMap := make(map[string]agents.MCPServerInfo)
	realHome := config.RealHomeDir()

	// 1. Read host configuration
	hostPath := HostMCPConfigPath(realHome)
	if entries, ok, err := merge.ReadJSONKey(hostPath, "mcpServers"); err == nil && ok {
		for _, name := range entries.Order {
			serverMap[name] = agents.ParseMCPServerMap(name, entries.Values[name], "host")
		}
	}

	// 2. Read profile configuration (overrides host)
	if profileDir != "" {
		profPath := ProfileMCPConfigPath(profileDir)
		if entries, ok, err := merge.ReadJSONKey(profPath, "mcpServers"); err == nil && ok {
			for _, name := range entries.Order {
				serverMap[name] = agents.ParseMCPServerMap(name, entries.Values[name], "profile")
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

// AddHostMCPServer registers or updates an MCP server in Cursor's host mcp.json.
func (a *Adapter) AddHostMCPServer(ctx context.Context, name string, cfg agents.MCPServerConfig) error {
	realHome := config.RealHomeDir()
	path := HostMCPConfigPath(realHome)
	return writeCursorMCPServer(path, name, cfg)
}

// RemoveHostMCPServer removes an MCP server from Cursor's host mcp.json.
func (a *Adapter) RemoveHostMCPServer(ctx context.Context, name string) error {
	realHome := config.RealHomeDir()
	path := HostMCPConfigPath(realHome)
	return deleteCursorMCPServer(path, name)
}

// SetHostMCPServerState toggles enabled/disabled state in Cursor's host mcp.json.
func (a *Adapter) SetHostMCPServerState(ctx context.Context, name string, enabled bool) error {
	realHome := config.RealHomeDir()
	path := HostMCPConfigPath(realHome)
	return toggleCursorMCPServer(path, name, enabled)
}

// AddProfileMCPServer registers or updates an MCP server in Cursor's profile mcp.json.
func (a *Adapter) AddProfileMCPServer(ctx context.Context, profileName, profileDir, name string, cfg agents.MCPServerConfig) error {
	path := ProfileMCPConfigPath(profileDir)
	return writeCursorMCPServer(path, name, cfg)
}

// RemoveProfileMCPServer removes an MCP server from Cursor's profile mcp.json.
func (a *Adapter) RemoveProfileMCPServer(ctx context.Context, profileName, profileDir, name string) error {
	path := ProfileMCPConfigPath(profileDir)
	return deleteCursorMCPServer(path, name)
}

// SetProfileMCPServerState toggles enabled/disabled state in Cursor's profile mcp.json.
func (a *Adapter) SetProfileMCPServerState(ctx context.Context, profileName, profileDir, name string, enabled bool) error {
	path := ProfileMCPConfigPath(profileDir)
	return toggleCursorMCPServer(path, name, enabled)
}

// DetectInstalled checks if cursor executable is in PATH or if Cursor configuration directories exist.
func (a *Adapter) DetectInstalled(realHome string) bool {
	if _, err := exec.LookPath("cursor"); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(realHome, ".cursor")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(realHome, "Library", "Application Support", "Cursor")); err == nil {
		return true
	}
	return false
}

func writeCursorMCPServer(path, name string, cfg agents.MCPServerConfig) error {
	root := make(map[string]any)
	data, err := os.ReadFile(path)
	if err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &root)
	}

	servers, ok := root["mcpServers"].(map[string]any)
	if !ok || servers == nil {
		servers = make(map[string]any)
	}

	entry := make(map[string]any)
	if cfg.URL != "" {
		entry["url"] = cfg.URL
		if len(cfg.Headers) > 0 {
			entry["headers"] = cfg.Headers
		}
	} else {
		if cfg.Command != "" {
			entry["command"] = cfg.Command
		}
		if len(cfg.Args) > 0 {
			entry["args"] = cfg.Args
		}
		if len(cfg.Env) > 0 {
			entry["env"] = cfg.Env
		}
	}
	if cfg.Disabled {
		entry["disabled"] = true
	}

	servers[name] = entry
	root["mcpServers"] = servers

	return atomicWriteJSONFile(path, root)
}

func deleteCursorMCPServer(path, name string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	servers, ok := root["mcpServers"].(map[string]any)
	if !ok || servers == nil {
		return nil
	}
	delete(servers, name)
	root["mcpServers"] = servers
	return atomicWriteJSONFile(path, root)
}

func toggleCursorMCPServer(path, name string, enabled bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	servers, ok := root["mcpServers"].(map[string]any)
	if !ok || servers == nil {
		return fmt.Errorf("server %q not found in %s", name, path)
	}
	serverEntry, exists := servers[name]
	if !exists {
		return fmt.Errorf("server %q not found in %s", name, path)
	}
	entryMap, ok := serverEntry.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid server entry format for %q", name)
	}
	if enabled {
		delete(entryMap, "disabled")
	} else {
		entryMap["disabled"] = true
	}
	servers[name] = entryMap
	root["mcpServers"] = servers
	return atomicWriteJSONFile(path, root)
}

func atomicWriteJSONFile(path string, data any) error {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(dir, ".mcp-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmpFile.Write(encoded); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
