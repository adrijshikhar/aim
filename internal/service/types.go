package service

import (
	"context"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

// QuotaWindowDTO represents an individual rate limit window (e.g. 5-Hour or Weekly).
type QuotaWindowDTO struct {
	Category     string     `json:"category,omitempty"`
	Name         string     `json:"name"`
	RemainingPct int        `json:"remaining_pct"`
	ResetsAt     *time.Time `json:"resets_at,omitempty"`
	ResetsInSec  int64      `json:"resets_in_sec,omitempty"`
}

// QuotaDTO represents the current quota state for a profile.
type QuotaDTO struct {
	BottleneckPct int              `json:"bottleneck_pct"`
	Summary       string           `json:"summary"`
	IsExhausted   bool             `json:"is_exhausted"`
	Windows       []QuotaWindowDTO `json:"windows,omitempty"`
	Credits       string           `json:"credits,omitempty"`
	FetchedAt     *time.Time       `json:"fetched_at,omitempty"`
}

// AdapterInfo represents adapter-specific identity, credentials, and quota within a profile.
type AdapterInfo struct {
	Agent          string               `json:"agent"`
	HasCredentials bool                 `json:"has_credentials"`
	Account        *profile.AccountInfo `json:"account,omitempty"`
	Quota          *QuotaDTO            `json:"quota,omitempty"`
}

// AdapterDTO represents a registered agent adapter and its supported runnable CLI flags.
type AdapterDTO struct {
	Name          string                `json:"name"`
	DisplayName   string                `json:"display_name"`
	BinaryName    string                `json:"binary_name"`
	Aliases       []string              `json:"aliases"`
	RunnableFlags []agents.RunnableFlag `json:"runnable_flags"`
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
	IsLeased       bool                 `json:"is_leased,omitempty"`
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
	PID       int       `json:"pid,omitempty"`
}

// SessionFilter provides query parameters for listing sessions.
type SessionFilter struct {
	Agent   string `json:"agent,omitempty"`
	Profile string `json:"profile,omitempty"`
	Cwd     string `json:"cwd,omitempty"`
	Query   string `json:"query,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Force   bool   `json:"force,omitempty"`
}

// ToolCallDTO represents an invoked tool action within an assistant turn.
type ToolCallDTO struct {
	ID        string `json:"id,omitempty"`
	ToolName  string `json:"tool_name"`
	Arguments string `json:"arguments,omitempty"`
	Output    string `json:"output,omitempty"`
	Status    string `json:"status,omitempty"` // "ok", "error"
}

// SessionTurnDTO represents a single conversational turn (user, assistant, or system).
type SessionTurnDTO struct {
	TurnIndex int           `json:"turn_index"`
	Role      string        `json:"role"` // "user", "assistant", "system"
	Timestamp time.Time     `json:"timestamp,omitempty"`
	Content   string        `json:"content"`
	Thinking  string        `json:"thinking,omitempty"`
	ToolCalls []ToolCallDTO `json:"tool_calls,omitempty"`
}

// SessionTranscriptDTO represents the full conversation replay and metadata for a session.
type SessionTranscriptDTO struct {
	SessionID string           `json:"session_id"`
	Agent     string           `json:"agent"`
	Profile   string           `json:"profile"`
	Title     string           `json:"title"`
	Cwd       string           `json:"cwd"`
	Goal      string           `json:"goal"`
	Turns     []SessionTurnDTO `json:"turns"`
}

// ResumeRequest defines parameters to resume an existing session in a terminal.
type ResumeRequest struct {
	Agent        string   `json:"agent"`
	Profile      string   `json:"profile"`
	SessionID    string   `json:"session_id"`
	Terminal     string   `json:"terminal,omitempty"`      // Target terminal emulator: Ghostty, iTerm2, Terminal, etc.
	Flags        []string `json:"flags,omitempty"`         // Legacy combined fallback
	AimFlags     []string `json:"aim_flags,omitempty"`     // Orchestrator flags: --exact, --catalyst, --fork, --force
	AdapterFlags []string `json:"adapter_flags,omitempty"` // Runnable adapter flags: e.g. --dangerously-skip-permissions
	CustomFlags  string   `json:"custom_flags,omitempty"`  // Legacy custom flags string
	CustomArgs   string   `json:"custom_args,omitempty"`   // Custom adapter arguments string
}

// LauncherService defines the contract for launching interactive terminal windows.
type LauncherService interface {
	LaunchTerminal(ctx context.Context, cmdStr string) error
	LaunchInTerminal(ctx context.Context, term, cmdStr string) error
	AvailableTerminals() []string
}

// ProfileService encapsulates all profile management business logic.
type ProfileService interface {
	ListProfiles(ctx context.Context, agent string) ([]ProfileDTO, error)
	ListAdapters(ctx context.Context) ([]AdapterDTO, error)
	CreateProfile(ctx context.Context, req CreateProfileRequest) (*ProfileDTO, error)
	RemoveProfile(ctx context.Context, agent, name string) error
	RenameProfile(ctx context.Context, agent, oldName, newName string) error
	UpdateProfileConfig(ctx context.Context, name string, mcpGlobal, pluginsGlobal *bool) error
	RefreshUsage(ctx context.Context, agent, profileName string) ([]ProfileDTO, error)
}

// SessionService encapsulates session discovery and resumption.
type SessionService interface {
	ListSessions(ctx context.Context, filter SessionFilter) ([]SessionDTO, error)
	ResumeSessionInTerminal(ctx context.Context, req ResumeRequest) error
	GetSessionTranscript(ctx context.Context, agent, sessionID string) (*SessionTranscriptDTO, error)
	InvalidateCache()
}

// MCPServerDTO represents a configured MCP server.
type MCPServerDTO struct {
	Name    string            `json:"name"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Scope   string            `json:"scope"`            // global or profile-specific
	Type    string            `json:"type,omitempty"`   // "stdio", "http", "sse"
	Status  string            `json:"status,omitempty"` // "enabled", "disabled"
	Auth    string            `json:"auth,omitempty"`   // "OAuth", "connected", "none", etc.
	Client  string            `json:"client,omitempty"` // "claude", "codex", "agy", etc.
}

// AddMCPServerRequest defines parameters for registering an MCP server.
type AddMCPServerRequest struct {
	Name    string            `json:"name"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Clients []string          `json:"clients,omitempty"`
	Profile string            `json:"profile,omitempty"` // empty = global/host
}

// SetMCPServerStateRequest defines parameters for enabling or disabling an MCP server.
type SetMCPServerStateRequest struct {
	Enabled bool     `json:"enabled"`
	Clients []string `json:"clients,omitempty"`
	Profile string   `json:"profile,omitempty"` // empty = global/host
}

// MCPService provides discovery and lifecycle management of configured MCP servers.
type MCPService interface {
	ListServers(ctx context.Context, profileName string) ([]MCPServerDTO, error)
	AddServer(ctx context.Context, req AddMCPServerRequest) error
	SetServerState(ctx context.Context, name string, req SetMCPServerStateRequest) error
	RemoveServer(ctx context.Context, name string, profileName string, clients []string) error
}

// DaemonDTO describes the background daemon operational state.
type DaemonDTO struct {
	Installed      bool      `json:"installed"`
	Active         bool      `json:"active"`
	Label          string    `json:"label"`
	ConfigPath     string    `json:"config_path"`
	IntervalSec    int       `json:"interval_sec"`
	BinaryPath     string    `json:"binary_path"`
	LogPath        string    `json:"log_path"`
	LastRun        time.Time `json:"last_run,omitempty"`
	LastRunMessage string    `json:"last_run_message,omitempty"`
}

// DaemonLogsDTO represents recent log records from the daemon log file.
type DaemonLogsDTO struct {
	LogPath    string   `json:"log_path"`
	Lines      []string `json:"lines"`
	TotalLines int      `json:"total_lines"`
}

// DaemonService provides management of the host OS background daemon service.
type DaemonService interface {
	GetStatus(ctx context.Context) (*DaemonDTO, error)
	GetLogs(ctx context.Context, maxLines, offset int) (*DaemonLogsDTO, error)
	Install(ctx context.Context, binaryPath string) (*DaemonDTO, error)
	Uninstall(ctx context.Context) error
	RunOnce(ctx context.Context) error
}

// DesktopStatusDTO represents the process and profile state of a graphical desktop application.
type DesktopStatusDTO struct {
	Agent       string `json:"agent"`
	DisplayName string `json:"display_name"`
	PID         int    `json:"pid"`
	Profile     string `json:"profile"`
	Running     bool   `json:"running"`
}

// DesktopLaunchRequest defines parameters to launch a sandboxed desktop application instance.
type DesktopLaunchRequest struct {
	Agent   string `json:"agent"`
	Profile string `json:"profile"`
	Force   bool   `json:"force"`
}

// DesktopService provides management and status inspection of desktop applications.
type DesktopService interface {
	ListStatus(ctx context.Context) ([]DesktopStatusDTO, error)
	Launch(ctx context.Context, req DesktopLaunchRequest) error
	Quit(ctx context.Context, agent string) error
	Focus(ctx context.Context, agent string) error
}
