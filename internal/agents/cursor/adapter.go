package cursor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

// Compile-time assertions that Adapter implements the segregated agent interfaces.
var (
	_ agents.AgentAdapter          = (*Adapter)(nil)
	_ agents.Diagnostician         = (*Adapter)(nil)
	_ agents.AccountInfoProvider   = (*Adapter)(nil)
	_ agents.RunnableFlagsProvider = (*Adapter)(nil)
	_ agents.MCPConfigManager      = (*Adapter)(nil)
)

type Adapter struct {
	agents.BaseAdapter
}

func NewAdapter() *Adapter {
	return &Adapter{
		BaseAdapter: agents.NewBaseAdapter(
			"cursor",
			"Cursor",
			"cursor",
			[]string{"cur"},
			"/Applications/Cursor.app/Contents/Resources/app/bin/cursor",
		),
	}
}

func (a *Adapter) GetAccountInfo(profileDir string) (email, authMethod, projectID string) {
	acc := profile.GetProfileAccountInfoForAgent(profileDir, a.Name())
	return acc.Email, acc.AuthMethod, acc.ProjectID
}

func (a *Adapter) HasCredentials(profileDir string) bool {
	// Cursor can authenticate via API keys in env, or stored auth in ~/.cursor or user-data
	if os.Getenv("CURSOR_API_KEY") != "" || os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("OPENAI_API_KEY") != "" {
		return true
	}
	dbPath := filepath.Join(profileDir, ".cursor", "user-data", "User", "globalStorage", "state.vscdb")
	if fi, err := os.Stat(dbPath); err == nil && fi.Size() > 0 {
		return true
	}
	// Also check host state.vscdb
	realHome := config.RealHomeDir()
	hostDb := filepath.Join(realHome, "Library", "Application Support", "Cursor", "User", "globalStorage", "state.vscdb")
	if fi, err := os.Stat(hostDb); err == nil && fi.Size() > 0 {
		return true
	}
	return false
}

func (a *Adapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	binPath := a.ResolveBinary()

	userDataDir := filepath.Join(profileDir, ".cursor", "user-data")
	extensionsDir := filepath.Join(profileDir, ".cursor", "extensions")
	_ = os.MkdirAll(userDataDir, 0700)
	_ = os.MkdirAll(extensionsDir, 0700)

	envMap := a.BaseLaunchEnv(profileName, profileDir, map[string]string{
		"AIM_PROFILE": profileName,
	})

	args := []string{
		"--user-data-dir", userDataDir,
		"--extensions-dir", extensionsDir,
	}

	cwd, _ := os.Getwd()
	if cwd == "" {
		cwd = profileDir
	}

	return agents.LaunchEnv{
		BinaryPath: binPath,
		Args:       args,
		Env:        envMap,
		WorkingDir: cwd,
	}, nil
}

func (a *Adapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	var results []agents.DiagnosticResult

	binPath := a.ResolveBinary()
	if path, err := exec.LookPath(binPath); err == nil {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "OK",
			Message:  fmt.Sprintf("Found cursor at %s", path),
		})
	} else if fi, err := os.Stat(binPath); err == nil && !fi.IsDir() {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "OK",
			Message:  fmt.Sprintf("Found cursor at %s", binPath),
		})
	} else {
		results = append(results, agents.DiagnosticResult{
			Category: "Binary",
			Status:   "WARN",
			Message:  "Cursor not installed or not in PATH",
		})
	}

	// Check MCP configuration file
	mcpPath := filepath.Join(profileDir, ".cursor", "mcp.json")
	if fi, err := os.Stat(mcpPath); err == nil && !fi.IsDir() {
		results = append(results, agents.DiagnosticResult{
			Category: "Config",
			Status:   "OK",
			Message:  fmt.Sprintf(".cursor/mcp.json found (%d bytes)", fi.Size()),
		})
	}

	return results
}

func (a *Adapter) RunnableFlags() []agents.RunnableFlag {
	return []agents.RunnableFlag{
		{
			Flag:        "--new-window",
			Name:        "New Window",
			Description: "Force open a new Cursor window.",
			Type:        "bool",
		},
		{
			Flag:        "--reuse-window",
			Name:        "Reuse Window",
			Description: "Force open a file or folder in the last active window.",
			Type:        "bool",
		},
		{
			Flag:        "--wait",
			Name:        "Wait",
			Description: "Wait for the files to be closed before returning.",
			Type:        "bool",
		},
	}
}
