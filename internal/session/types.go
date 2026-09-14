package session

import (
	"context"
	"time"
)

// Status represents the runtime lifecycle state of a conversation session.
type Status string

const (
	// StatusActive indicates a live process is currently attached to this conversation.
	StatusActive Status = "ACTIVE"
	// StatusIdle indicates a dormant conversation saved on disk.
	StatusIdle Status = "IDLE"
)

// Session encapsulates metadata for an agent conversation.
type Session struct {
	ID           string    `json:"id"`
	ShortID      string    `json:"short_id"`
	Title        string    `json:"title"`
	Agent        string    `json:"agent"`
	Profile      string    `json:"profile"`
	IsHost       bool      `json:"is_host"`
	StartedAt    time.Time `json:"started_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	Status       Status    `json:"status"`
	Summary      string    `json:"summary"`
	StoragePath  string    `json:"storage_path"`
	PID          int       `json:"pid,omitempty"`
}

// ComputeShortID returns the 8-character prefix of an ID.
func ComputeShortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// NewSession constructs a Session struct with computed ShortID.
func NewSession(id, title, agent, profile string, isHost bool, lastActive time.Time) Session {
	return Session{
		ID:           id,
		ShortID:      ComputeShortID(id),
		Title:        title,
		Agent:        agent,
		Profile:      profile,
		IsHost:       isHost,
		LastActiveAt: lastActive,
		Status:       StatusIdle,
	}
}

// ActiveProcessInfo captures information about a running process attached to a conversation.
type ActiveProcessInfo struct {
	PID            int
	Agent          string
	Profile        string
	ConversationID string
}

// ProcessScanner defines the interface for discovering active processes.
type ProcessScanner interface {
	ScanActiveProcesses(ctx context.Context) (map[string]ActiveProcessInfo, error)
}

// SessionProvider is the interface implemented by each agent adapter for conversation management.
type SessionProvider interface {
	Agent() string
	ListSessions(ctx context.Context, profileDir string, isHost bool) ([]Session, error)
	GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*Session, error)
	Hydrate(ctx context.Context, srcSession *Session, destProfileDir string, fork bool) (string, error)
}
