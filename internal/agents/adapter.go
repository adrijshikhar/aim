package agents

import (
	"context"

	"github.com/aim-cli/aim/internal/usage"
)

type DiagnosticResult struct {
	Category string
	Status   string // "OK", "WARN", "FAIL"
	Message  string
}

type LaunchEnv struct {
	BinaryPath string
	Args       []string
	Env        map[string]string
	WorkingDir string
}

type AgentAdapter interface {
	Name() string
	DisplayName() string
	Aliases() []string
	BinaryName() string
	HasCredentials(profileDir string) bool
	Login(ctx context.Context, profileName, profileDir string) error
	PrepareEnv(profileName, profileDir string) (LaunchEnv, error)
	Doctor(ctx context.Context, profileName, profileDir string) []DiagnosticResult
	GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error)
}
