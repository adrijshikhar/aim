package service

import (
	"context"
	"time"

	"github.com/aim-cli/aim/internal/profile"
)

// QuotaDTO represents the current quota state for a profile.
type QuotaDTO struct {
	BottleneckPct int    `json:"bottleneck_pct"`
	Summary       string `json:"summary"`
	IsExhausted   bool   `json:"is_exhausted"`
}

// AdapterInfo represents adapter-specific identity, credentials, and quota within a profile.
type AdapterInfo struct {
	Agent          string               `json:"agent"`
	HasCredentials bool                 `json:"has_credentials"`
	Account        *profile.AccountInfo `json:"account,omitempty"`
	Quota          *QuotaDTO            `json:"quota,omitempty"`
}

// ProfileDTO represents an agent profile, which may host one or more agent adapters.
type ProfileDTO struct {
	Agent          string               `json:"agent"`
	Name           string               `json:"name"`
	Path           string               `json:"path"`
	HasCredentials bool                 `json:"has_credentials"`
	Account        *profile.AccountInfo `json:"account,omitempty"`
	Quota          *QuotaDTO            `json:"quota,omitempty"`
	MCPGlobal      *bool                `json:"mcp_global,omitempty"`
	PluginsGlobal  *bool                `json:"plugins_global,omitempty"`
	Adapters       []AdapterInfo        `json:"adapters,omitempty"`
}

// CreateProfileRequest represents a request to create or scaffold a new profile.
type CreateProfileRequest struct {
	Agent     string `json:"agent"`
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	CloneFrom string `json:"clone_from,omitempty"`
}

// SessionDTO represents an active or historical conversation session.
type SessionDTO struct {
	ID        string    `json:"id"`
	Agent     string    `json:"agent"`
	Profile   string    `json:"profile"`
	Title     string    `json:"title"`
	Cwd       string    `json:"cwd"`
	Goal      string    `json:"goal"`
	Turns     int       `json:"turns"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

// SessionFilter provides query parameters for listing sessions.
type SessionFilter struct {
	Agent   string `json:"agent,omitempty"`
	Profile string `json:"profile,omitempty"`
	Query   string `json:"query,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// ResumeRequest defines parameters to resume an existing session in a terminal.
type ResumeRequest struct {
	Agent     string `json:"agent"`
	Profile   string `json:"profile"`
	SessionID string `json:"session_id"`
}

// LauncherService defines the contract for launching interactive terminal windows.
type LauncherService interface {
	LaunchTerminal(ctx context.Context, cmdStr string) error
	AvailableTerminals() []string
}

// ProfileService encapsulates all profile management business logic.
type ProfileService interface {
	ListProfiles(ctx context.Context, agent string) ([]ProfileDTO, error)
	CreateProfile(ctx context.Context, req CreateProfileRequest) (*ProfileDTO, error)
	RemoveProfile(ctx context.Context, agent, name string) error
	RenameProfile(ctx context.Context, agent, oldName, newName string) error
	UpdateProfileConfig(ctx context.Context, name string, mcpGlobal, pluginsGlobal *bool) error
}

// SessionService encapsulates session discovery and resumption.
type SessionService interface {
	ListSessions(ctx context.Context, filter SessionFilter) ([]SessionDTO, error)
	ResumeSessionInTerminal(ctx context.Context, req ResumeRequest) error
}

// MCPServerDTO represents a configured MCP server.
type MCPServerDTO struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
	Scope   string            `json:"scope"` // global or profile-specific
}

// MCPService provides discovery of configured MCP servers.
type MCPService interface {
	ListServers(ctx context.Context, profileName string) ([]MCPServerDTO, error)
}
