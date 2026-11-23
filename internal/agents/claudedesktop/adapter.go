package claudedesktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
)

var (
	_ agents.AgentAdapter          = (*Adapter)(nil)
	_ agents.Diagnostician         = (*Adapter)(nil)
	_ agents.AccountInfoProvider   = (*Adapter)(nil)
	_ agents.RunnableFlagsProvider = (*Adapter)(nil)
	_ agents.MCPConfigManager      = (*Adapter)(nil)
	_ agents.DesktopApp            = (*Adapter)(nil)
)

type Adapter struct {
	agents.BaseAdapter
}

func NewAdapter() *Adapter {
	extraPaths := []string{
		"/Applications/Claude.app/Contents/MacOS/Claude",
		filepath.Join(config.RealHomeDir(), "Applications", "Claude.app", "Contents", "MacOS", "Claude"),
	}
	if runtime.GOOS == "linux" {
		extraPaths = []string{
			"/usr/bin/claude-desktop",
			"/opt/Claude/claude-desktop",
		}
	}
	return &Adapter{
		BaseAdapter: agents.NewBaseAdapter(
			"claude-desktop",
			"Claude Desktop",
			"claude-desktop",
			[]string{"clauded", "cd"},
			extraPaths...,
		),
	}
}

func (a *Adapter) IsDesktop() bool       { return true }
func (a *Adapter) AppBundleName() string { return "Claude" }

func (a *Adapter) UserDataDir(profileDir string) string {
	return filepath.Join(profileDir, ".claude-desktop", "user-data")
}

func (a *Adapter) GetAccountInfo(profileDir string) (email, authMethod, projectID string) {
	acc := profile.GetProfileAccountInfoForAgent(profileDir, a.Name())
	return acc.Email, acc.AuthMethod, acc.ProjectID
}

func (a *Adapter) HasCredentials(profileDir string) bool {
	userDataDir := a.UserDataDir(profileDir)
	localState := filepath.Join(userDataDir, "Local State")
	if fi, err := os.Stat(localState); err == nil && fi.Size() > 0 {
		return true
	}
	mcpPath := filepath.Join(profileDir, ".claude-desktop", "claude_desktop_config.json")
	if fi, err := os.Stat(mcpPath); err == nil && fi.Size() > 0 {
		return true
	}
	return false
}

func (a *Adapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	binPath := a.ResolveBinary()
	userDataDir := a.UserDataDir(profileDir)
	_ = os.MkdirAll(userDataDir, 0700)

	envMap := a.BaseLaunchEnv(profileName, profileDir, map[string]string{
		"AIM_PROFILE": profileName,
		"AIM_AGENT":   a.Name(),
	})

	args := []string{
		"--user-data-dir=" + userDataDir,
	}

	cwd, _ := os.Getwd()
	if cwd == "" {
		cwd = profileDir
	}

	return agents.LaunchEnv{
		BinaryPath: binPath,
		WorkingDir: cwd,
		Args:       args,
		Env:        envMap,
	}, nil
}

func (a *Adapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	var results []agents.DiagnosticResult
	binPath := a.ResolveBinary()
	if path, err := exec.LookPath(binPath); err == nil {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "OK",
			Message:  fmt.Sprintf("Found Claude Desktop at %s", path),
		})
	} else if fi, err := os.Stat(binPath); err == nil && !fi.IsDir() {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "OK",
			Message:  fmt.Sprintf("Found Claude Desktop at %s", binPath),
		})
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "WARN",
			Message:  "Claude Desktop not installed or not in PATH",
		})
	}

	if runtime.GOOS == "linux" {
		if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			results = append(results, agents.DiagnosticResult{
				Category: "Display",
				Status:   "FAIL",
				Message:  "Neither $DISPLAY nor $WAYLAND_DISPLAY is set",
			})
		}
	}

	mcpPath := filepath.Join(profileDir, ".claude-desktop", "claude_desktop_config.json")
	if fi, err := os.Stat(mcpPath); err == nil && !fi.IsDir() {
		results = append(results, agents.DiagnosticResult{
			Category: "Config",
			Status:   "OK",
			Message:  fmt.Sprintf(".claude-desktop/claude_desktop_config.json found (%d bytes)", fi.Size()),
		})
	}

	return results
}

func (a *Adapter) RunnableFlags() []agents.RunnableFlag {
	return []agents.RunnableFlag{}
}

func (a *Adapter) DetectInstalled(realHome string) bool {
	bin := a.ResolveBinary()
	if _, err := os.Stat(bin); err == nil {
		return true
	}
	_, err := exec.LookPath(a.BinaryName())
	return err == nil
}

func HostMCPConfigPath(realHome string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(realHome, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	}
	return filepath.Join(realHome, ".config", "Claude", "claude_desktop_config.json")
}

func ProfileMCPConfigPath(profileDir string) string {
	return filepath.Join(profileDir, ".claude-desktop", "claude_desktop_config.json")
}

func (a *Adapter) ListMCPServers(ctx context.Context, profileName, profileDir string) ([]agents.MCPServerInfo, error) {
	serverMap := make(map[string]agents.MCPServerInfo)
	realHome := config.RealHomeDir()

	hostPath := HostMCPConfigPath(realHome)
	if entries, ok, err := merge.ReadJSONKey(hostPath, "mcpServers"); err == nil && ok {
		for _, name := range entries.Order {
			serverMap[name] = agents.ParseMCPServerMap(name, entries.Values[name], "host")
		}
	}

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

func (a *Adapter) AddHostMCPServer(ctx context.Context, name string, cfg agents.MCPServerConfig) error {
	path := HostMCPConfigPath(config.RealHomeDir())
	return writeClaudeDesktopMCPServer(path, name, cfg)
}

func (a *Adapter) RemoveHostMCPServer(ctx context.Context, name string) error {
	path := HostMCPConfigPath(config.RealHomeDir())
	return deleteClaudeDesktopMCPServer(path, name)
}

func (a *Adapter) SetHostMCPServerState(ctx context.Context, name string, enabled bool) error {
	path := HostMCPConfigPath(config.RealHomeDir())
	return toggleClaudeDesktopMCPServer(path, name, enabled)
}

func (a *Adapter) AddProfileMCPServer(ctx context.Context, profileName, profileDir, name string, cfg agents.MCPServerConfig) error {
	path := ProfileMCPConfigPath(profileDir)
	return writeClaudeDesktopMCPServer(path, name, cfg)
}

func (a *Adapter) RemoveProfileMCPServer(ctx context.Context, profileName, profileDir, name string) error {
	path := ProfileMCPConfigPath(profileDir)
	return deleteClaudeDesktopMCPServer(path, name)
}

func (a *Adapter) SetProfileMCPServerState(ctx context.Context, profileName, profileDir, name string, enabled bool) error {
	path := ProfileMCPConfigPath(profileDir)
	return toggleClaudeDesktopMCPServer(path, name, enabled)
}

func writeClaudeDesktopMCPServer(path, name string, cfg agents.MCPServerConfig) error {
	root := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil {
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

func deleteClaudeDesktopMCPServer(path, name string) error {
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

func toggleClaudeDesktopMCPServer(path, name string, enabled bool) error {
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
	entry, ok := servers[name].(map[string]any)
	if !ok {
		return fmt.Errorf("server %q not found in %s", name, path)
	}
	if enabled {
		delete(entry, "disabled")
	} else {
		entry["disabled"] = true
	}
	servers[name] = entry
	root["mcpServers"] = servers
	return atomicWriteJSONFile(path, root)
}

func atomicWriteJSONFile(path string, data any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp.*.json")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("setting file permissions: %w", err)
	}

	if _, err := tmp.Write(buf.Bytes()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing to temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("syncing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming temp file to %s: %w", path, err)
	}
	return nil
}
