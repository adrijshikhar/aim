package codex

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/session"
)

// Ensure Provider implements session.SessionProvider.
var _ session.SessionProvider = (*Provider)(nil)

// Provider implements session.SessionProvider for OpenAI Codex CLI (codex).
type Provider struct {
	sqliteBin string
}

func NewProvider() *Provider {
	bin, err := exec.LookPath("sqlite3")
	if err != nil {
		for _, fallback := range []string{"/usr/bin/sqlite3", "/opt/homebrew/bin/sqlite3", "/usr/local/bin/sqlite3"} {
			if _, statErr := os.Stat(fallback); statErr == nil {
				bin = fallback
				break
			}
		}
	}
	return &Provider{sqliteBin: bin}
}

func (p *Provider) Agent() string {
	return "codex"
}

func (p *Provider) codexDir(profileDir string, isHost bool) string {
	if isHost {
		return filepath.Join(config.RealHomeDir(), ".codex")
	}
	return filepath.Join(profileDir, ".codex")
}

func (p *Provider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	cDir := p.codexDir(profileDir, isHost)
	profileName := filepath.Base(profileDir)
	if isHost {
		profileName = "<host>"
	}

	// 1. Try SQLite state_5.sqlite first
	dbPath := filepath.Join(cDir, "state_5.sqlite")
	if p.sqliteBin != "" {
		if fi, err := os.Stat(dbPath); err == nil && fi.Size() > 0 {
			sessions, err := p.listFromSQLite(ctx, dbPath, profileName, isHost)
			if err == nil && len(sessions) > 0 {
				return sessions, nil
			}
		}
	}

	// 2. Fallback to session_index.jsonl
	indexPath := filepath.Join(cDir, "session_index.jsonl")
	if fi, err := os.Stat(indexPath); err == nil && fi.Size() > 0 {
		return p.listFromJSONL(indexPath, profileName, isHost)
	}

	return nil, nil
}

func (p *Provider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	if idOrPrefix == "" {
		return nil, nil
	}
	cDir := p.codexDir(profileDir, isHost)
	profileName := filepath.Base(profileDir)
	if isHost {
		profileName = "<host>"
	}

	dbPath := filepath.Join(cDir, "state_5.sqlite")
	if p.sqliteBin != "" {
		if fi, err := os.Stat(dbPath); err == nil && fi.Size() > 0 {
			s, err := p.getFromSQLite(ctx, dbPath, idOrPrefix, profileName, isHost)
			if err != nil {
				return nil, err
			}
			if s != nil {
				if sum, err := p.ResolveSummary(ctx, s); err == nil {
					s.Goal = sum.Goal
					s.Progress = sum.Progress
					s.Recent = sum.RecentActivity
					if sum.Text() != "" {
						s.Summary = sum.Text()
					}
				}
				return s, nil
			}
		}
	}

	indexPath := filepath.Join(cDir, "session_index.jsonl")
	if fi, err := os.Stat(indexPath); err == nil && fi.Size() > 0 {
		s, err := p.getFromJSONL(indexPath, idOrPrefix, profileName, isHost)
		if err != nil {
			return nil, err
		}
		if s != nil {
			if sum, err := p.ResolveSummary(ctx, s); err == nil {
				s.Goal = sum.Goal
				s.Progress = sum.Progress
				s.Recent = sum.RecentActivity
				if sum.Text() != "" {
					s.Summary = sum.Text()
				}
			}
			return s, nil
		}
	}

	return nil, nil
}

func isValidProfileName(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	if s == "." || s == ".." {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func isValidIdentifier(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func isValidSessionID(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func isValidSessionQuery(s string) bool {
	if s == "" || len(s) > 256 {
		return false
	}
	if strings.ContainsAny(s, "\x00;'\"\n\r") {
		return false
	}
	return true
}

func escapeSQL(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.ReplaceAll(s, "'", "''")
}
