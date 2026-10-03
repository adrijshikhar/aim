package gemini

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/usage"
)

// Compile-time assertion that Adapter implements agents.AgentAdapter.
var _ agents.AgentAdapter = (*Adapter)(nil)

type Adapter struct {
	agents.BaseAdapter
}

func NewAdapter() *Adapter {
	return &Adapter{
		BaseAdapter: agents.NewBaseAdapter("gemini", "Gemini CLI", "gemini", []string{"gemini-cli"}),
	}
}

func (a *Adapter) TokenPath(profileDir string) string {
	return filepath.Join(profileDir, ".gemini", "gemini-oauth-token")
}

func (a *Adapter) HasCredentials(profileDir string) bool {
	p := a.TokenPath(profileDir)
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir() && fi.Size() > 0
}

func (a *Adapter) Login(ctx context.Context, profileName, profileDir string) error {
	return fmt.Errorf("gemini interactive login adapter coming soon")
}

func (a *Adapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	bin := a.ResolveBinary()
	envMap := a.BaseLaunchEnv(profileName, profileDir, map[string]string{
		"GEMINI_CLI_HOME": filepath.Join(profileDir, ".gemini"),
	})
	cwd, _ := os.Getwd()
	return agents.LaunchEnv{
		BinaryPath: bin,
		Env:        envMap,
		WorkingDir: cwd,
	}, nil
}

func (a *Adapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	return []agents.DiagnosticResult{
		{Category: "Adapter", Status: "OK", Message: "Gemini adapter registered"},
	}
}

func (a *Adapter) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	if !a.HasCredentials(profileDir) {
		return &usage.Report{
			Agent:     a.Name(),
			Profile:   profileName,
			Status:    usage.StatusUnknown,
			FetchedAt: time.Now(),
			Error:     "no credentials",
		}, nil
	}

	return &usage.Report{
		Agent:     a.Name(),
		Profile:   profileName,
		Status:    usage.StatusOK,
		Summary:   "Active credentials",
		FetchedAt: time.Now(),
	}, nil
}
