package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ClientID identifies a supported AI coding agent or IDE.
type ClientID string

const (
	ClientClaudeCode    ClientID = "claude"
	ClientClaudeDesktop ClientID = "claude-desktop"
	ClientCursor        ClientID = "cursor"
	ClientWindsurf      ClientID = "windsurf"
)

// RawMCPEntry represents an MCP server entry in client configuration files.
type RawMCPEntry struct {
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	ServerURL string            `json:"serverUrl,omitempty"`
	URL       string            `json:"url,omitempty"`
	Disabled  bool              `json:"disabled,omitempty"`
}

// ClientConfigDriver handles MCP configuration detection, reading, and writing for a specific client.
type ClientConfigDriver interface {
	ID() ClientID
	Name() string
	Detect(homeDir string) bool
	ConfigPath(homeDir string, profileDir string) (string, error)
	ReadServers(path string) (map[string]RawMCPEntry, error)
	WriteServer(path string, name string, entry RawMCPEntry) error
	DeleteServer(path string, name string) error
}

var allDrivers = []ClientConfigDriver{
	&claudeCodeDriver{},
	&claudeDesktopDriver{},
	&cursorDriver{},
	&windsurfDriver{},
}

// AllDrivers returns all registered ClientConfigDriver implementations.
func AllDrivers() []ClientConfigDriver {
	drivers := make([]ClientConfigDriver, len(allDrivers))
	copy(drivers, allDrivers)
	return drivers
}

// GetDriver returns the driver matching the given ClientID, or false if not found.
func GetDriver(id ClientID) (ClientConfigDriver, bool) {
	for _, d := range allDrivers {
		if d.ID() == id {
			return d, true
		}
	}
	return nil, false
}

// DetectInstalledClients scans the environment and returns drivers for all installed clients.
func DetectInstalledClients(homeDir string) []ClientConfigDriver {
	var detected []ClientConfigDriver
	for _, d := range allDrivers {
		if d.Detect(homeDir) {
			detected = append(detected, d)
		}
	}
	return detected
}

// --- Shared Base Driver with Atomic JSON Storage ---

type baseJSONDriver struct {
	id   ClientID
	name string
}

func (b *baseJSONDriver) ID() ClientID {
	return b.id
}

func (b *baseJSONDriver) Name() string {
	return b.name
}

func (b *baseJSONDriver) ReadServers(path string) (map[string]RawMCPEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]RawMCPEntry), nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return make(map[string]RawMCPEntry), nil
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse JSON at %s: %w", path, err)
	}

	servers := make(map[string]RawMCPEntry)
	rawServers, ok := root["mcpServers"].(map[string]any)
	if !ok || rawServers == nil {
		return servers, nil
	}

	for sName, sVal := range rawServers {
		bytes, mErr := json.Marshal(sVal)
		if mErr != nil {
			continue
		}
		var entry RawMCPEntry
		if uErr := json.Unmarshal(bytes, &entry); uErr == nil {
			servers[sName] = entry
		}
	}

	return servers, nil
}

func (b *baseJSONDriver) WriteServer(path string, name string, entry RawMCPEntry) error {
	var root map[string]any
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing config at %s: %w", path, err)
	}
	if err == nil && len(data) > 0 {
		if jsonErr := json.Unmarshal(data, &root); jsonErr != nil {
			return fmt.Errorf("failed to parse existing config at %s: %w", path, jsonErr)
		}
	}
	if root == nil {
		root = make(map[string]any)
	}

	servers, ok := root["mcpServers"].(map[string]any)
	if !ok || servers == nil {
		servers = make(map[string]any)
	}

	servers[name] = entry
	root["mcpServers"] = servers

	return b.atomicWriteJSON(path, root)
}

func (b *baseJSONDriver) DeleteServer(path string, name string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse JSON at %s: %w", path, err)
	}

	servers, ok := root["mcpServers"].(map[string]any)
	if !ok || servers == nil {
		return nil
	}

	if _, exists := servers[name]; !exists {
		return nil
	}

	delete(servers, name)
	root["mcpServers"] = servers

	return b.atomicWriteJSON(path, root)
}

func (b *baseJSONDriver) atomicWriteJSON(path string, data any) error {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	encoded = append(encoded, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpFile := filepath.Join(dir, fmt.Sprintf(".%s.tmp.%d", filepath.Base(path), os.Getpid()))
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := f.Write(encoded); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpFile, path); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file to %s: %w", path, err)
	}

	return nil
}

// --- Specific Drivers ---

// Claude Code Driver
type claudeCodeDriver struct {
	baseJSONDriver
}

func init() {
	allDrivers[0] = &claudeCodeDriver{
		baseJSONDriver: baseJSONDriver{
			id:   ClientClaudeCode,
			name: "Claude Code",
		},
	}
	allDrivers[1] = &claudeDesktopDriver{
		baseJSONDriver: baseJSONDriver{
			id:   ClientClaudeDesktop,
			name: "Claude Desktop",
		},
	}
	allDrivers[2] = &cursorDriver{
		baseJSONDriver: baseJSONDriver{
			id:   ClientCursor,
			name: "Cursor IDE",
		},
	}
	allDrivers[3] = &windsurfDriver{
		baseJSONDriver: baseJSONDriver{
			id:   ClientWindsurf,
			name: "Windsurf IDE",
		},
	}
}

func (d *claudeCodeDriver) Detect(homeDir string) bool {
	if _, err := os.Stat(filepath.Join(homeDir, ".claude.json")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".claude")); err == nil {
		return true
	}
	if _, err := exec.LookPath("claude"); err == nil {
		return true
	}
	return false
}

func (d *claudeCodeDriver) ConfigPath(homeDir string, profileDir string) (string, error) {
	if profileDir != "" {
		return filepath.Join(profileDir, ".claude", ".claude.json"), nil
	}
	return filepath.Join(homeDir, ".claude.json"), nil
}

// Claude Desktop Driver
type claudeDesktopDriver struct {
	baseJSONDriver
}

func (d *claudeDesktopDriver) Detect(homeDir string) bool {
	cfg, err := d.ConfigPath(homeDir, "")
	if err == nil {
		if _, err := os.Stat(cfg); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Dir(cfg)); err == nil {
			return true
		}
	}
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat("/Applications/Claude.app"); err == nil {
			return true
		}
	}
	return false
}

func (d *claudeDesktopDriver) ConfigPath(homeDir string, profileDir string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	default:
		// Linux / other XDG
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(homeDir, ".config")
		}
		return filepath.Join(configDir, "Claude", "claude_desktop_config.json"), nil
	}
}

// Cursor Driver
type cursorDriver struct {
	baseJSONDriver
}

func (d *cursorDriver) Detect(homeDir string) bool {
	if _, err := os.Stat(filepath.Join(homeDir, ".cursor")); err == nil {
		return true
	}
	if _, err := exec.LookPath("cursor"); err == nil {
		return true
	}
	return false
}

func (d *cursorDriver) ConfigPath(homeDir string, profileDir string) (string, error) {
	return filepath.Join(homeDir, ".cursor", "mcp.json"), nil
}

// Windsurf Driver
type windsurfDriver struct {
	baseJSONDriver
}

func (d *windsurfDriver) Detect(homeDir string) bool {
	if _, err := os.Stat(filepath.Join(homeDir, ".codeium", "windsurf")); err == nil {
		return true
	}
	if _, err := exec.LookPath("windsurf"); err == nil {
		return true
	}
	return false
}

func (d *windsurfDriver) ConfigPath(homeDir string, profileDir string) (string, error) {
	return filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json"), nil
}
