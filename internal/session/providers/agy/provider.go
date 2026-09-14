package agy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/session"
)

// Ensure Provider implements session.SessionProvider.
var _ session.SessionProvider = (*Provider)(nil)

// Provider implements session.SessionProvider for Antigravity CLI (agy).
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
	return "agy"
}

func (p *Provider) dbPath(profileDir string, isHost bool) string {
	if isHost {
		return filepath.Join(config.RealHomeDir(), ".gemini", "antigravity-cli", "conversation_summaries.db")
	}
	return filepath.Join(profileDir, ".gemini", "antigravity-cli", "conversation_summaries.db")
}

func (p *Provider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	if p.sqliteBin == "" {
		return nil, nil
	}

	db := p.dbPath(profileDir, isHost)
	if fi, err := os.Stat(db); err != nil || fi.Size() == 0 {
		return nil, nil
	}

	query := "SELECT conversation_id, title, preview, last_modified_time FROM conversation_summaries ORDER BY last_modified_time DESC LIMIT 50;"
	cmd := exec.CommandContext(ctx, p.sqliteBin, db, "-separator", "|||", query)
	out, err := cmd.Output()
	if err != nil {
		logger.Debug("[session/agy] query error on %s: %v", db, err)
		return nil, nil
	}

	profileName := filepath.Base(profileDir)
	if isHost {
		profileName = "<host>"
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var sessions []session.Session

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "|||")
		if len(parts) < 4 {
			continue
		}

		convID := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])
		preview := strings.TrimSpace(parts[2])
		rawTime := strings.TrimSpace(parts[3])

		if title == "" {
			if preview != "" {
				firstLine := strings.Split(preview, "\n")[0]
				title = strings.TrimSpace(firstLine)
			}
			if title == "" {
				title = "Untitled Session"
			}
		}

		modTime := parseTimeString(rawTime)
		s := session.NewSession(convID, title, "agy", profileName, isHost, modTime)
		s.Summary = preview
		s.StoragePath = db
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (p *Provider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	if p.sqliteBin == "" || idOrPrefix == "" {
		return nil, nil
	}

	db := p.dbPath(profileDir, isHost)
	if fi, err := os.Stat(db); err != nil || fi.Size() == 0 {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT conversation_id, title, preview, last_modified_time FROM conversation_summaries WHERE conversation_id LIKE '%s%%' LIMIT 2;", idOrPrefix)
	cmd := exec.CommandContext(ctx, p.sqliteBin, db, "-separator", "|||", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return nil, nil
	}

	profileName := filepath.Base(profileDir)
	if isHost {
		profileName = "<host>"
	}

	parts := strings.Split(lines[0], "|||")
	if len(parts) < 4 {
		return nil, nil
	}

	convID := strings.TrimSpace(parts[0])
	title := strings.TrimSpace(parts[1])
	preview := strings.TrimSpace(parts[2])
	rawTime := strings.TrimSpace(parts[3])

	if title == "" {
		title = preview
	}

	modTime := parseTimeString(rawTime)
	s := session.NewSession(convID, title, "agy", profileName, isHost, modTime)
	s.Summary = preview
	s.StoragePath = db
	return &s, nil
}

func (p *Provider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	if srcSession == nil {
		return "", fmt.Errorf("source session is nil")
	}

	// For Antigravity, conversation files are already symlinked to the shared directory across profiles.
	// In the future with fork=true, we can clone the conversation JSON trajectory.
	return srcSession.ID, nil
}

func parseTimeString(raw string) time.Time {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05.999999+00:00",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, raw); err == nil {
			return t
		}
	}
	return time.Now()
}
