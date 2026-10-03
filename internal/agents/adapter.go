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
	BinaryPath   string
	Args         []string
	Env          map[string]string
	WorkingDir   string
	PostLauncher PostLauncher
}

// PostLauncher is optionally implemented by adapters that need to perform background work
// (such as credentials synchronization or keychain harvesting) during the active lifetime
// of the running process.
type PostLauncher interface {
	PostLaunch(ctx context.Context, profileName, profileDir string)
}

// AgentAdapter represents the minimal core interface required for an AI agent adapter.
type AgentAdapter interface {
	Name() string
	DisplayName() string
	Aliases() []string
	BinaryName() string
	HasCredentials(profileDir string) bool
	PrepareEnv(ctx context.Context, profileName, profileDir string) (LaunchEnv, error)
}

// Authenticator is optionally implemented by adapters that support interactive or automated login.
type Authenticator interface {
	Login(ctx context.Context, profileName, profileDir string) error
}

// Diagnostician is optionally implemented by adapters that can evaluate profile health and connectivity.
type Diagnostician interface {
	Doctor(ctx context.Context, profileName, profileDir string) []DiagnosticResult
}

// UsageProvider is optionally implemented by adapters that report quota and usage telemetry.
type UsageProvider interface {
	GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error)
}

// AccountInfoProvider is optionally implemented by adapters that report authenticated account details.
type AccountInfoProvider interface {
	GetAccountInfo(profileDir string) (email, authMethod, projectID string)
}

// FullAdapter represents an adapter that implements all core and standard optional capabilities.
type FullAdapter interface {
	AgentAdapter
	Authenticator
	Diagnostician
	UsageProvider
	AccountInfoProvider
}
