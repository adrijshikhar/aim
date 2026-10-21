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

// CachedUsageProvider is optionally implemented by adapters that extract quota
// and rate-limit telemetry directly from local profile files on disk.
// Contract: strictly read-only, zero network calls, zero child processes.
type CachedUsageProvider interface {
	GetCachedUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error)
}

// LiveUsageProvider is optionally implemented by adapters that query live quota
// via provider CLI subcommands or authenticated endpoints.
type LiveUsageProvider interface {
	GetLiveUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error)
}

// CredentialHarvester is optionally implemented by adapters that need to discover,
// extract, or synchronize credentials from host stores (macOS Keychain, token files,
// system auth) into an isolated profile directory.
type CredentialHarvester interface {
	HarvestCredentials(ctx context.Context, profileName, profileDir string) error
}

// CredentialPurger is optionally implemented by adapters to clean up external or OS-level
// credentials (e.g. macOS Keychain items) when a profile is purged or deleted.
type CredentialPurger interface {
	PurgeCredentials(ctx context.Context, profileName, profileDir string) error
}

// SessionResumer is optionally implemented by adapters that define the command-line
// arguments required to resume an existing conversation thread.
type SessionResumer interface {
	ResumeArgs(sessionID string, extraArgs []string) []string
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

// RunnableFlag describes a supported flag that can be passed directly to the underlying agent CLI.
type RunnableFlag struct {
	Flag        string   `json:"flag"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"` // "bool", "string", "choice"
	Default     string   `json:"default,omitempty"`
	Choices     []string `json:"choices,omitempty"`
	Dangerous   bool     `json:"dangerous,omitempty"`
}

// RunnableFlagsProvider is optionally implemented by adapters that expose supported upstream CLI flags.
type RunnableFlagsProvider interface {
	RunnableFlags() []RunnableFlag
}
