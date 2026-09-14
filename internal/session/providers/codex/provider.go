package codex

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
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

func (p *Provider) listFromSQLite(ctx context.Context, dbPath, profileName string, isHost bool) ([]session.Session, error) {
	query := "SELECT id, title, preview, updated_at, rollout_path FROM threads ORDER BY updated_at DESC LIMIT 50;"
	cmd := exec.CommandContext(ctx, p.sqliteBin, dbPath, "-separator", "|||", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var sessions []session.Session
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "|||")
		if len(parts) < 5 {
			continue
		}

		id := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])
		preview := strings.TrimSpace(parts[2])
		rawUnix := strings.TrimSpace(parts[3])
		rolloutPath := strings.TrimSpace(parts[4])

		if title == "" {
			if preview != "" {
				firstLine := strings.Split(preview, "\n")[0]
				title = strings.TrimSpace(firstLine)
			}
			if title == "" {
				title = "Untitled Session"
			}
		}

		sec, _ := strconv.ParseInt(rawUnix, 10, 64)
		modTime := time.Unix(sec, 0)

		s := session.NewSession(id, title, "codex", profileName, isHost, modTime)
		s.Summary = preview
		s.StoragePath = rolloutPath
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (p *Provider) listFromJSONL(indexPath, profileName string, isHost bool) ([]session.Session, error) {
	f, err := os.Open(indexPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var sessions []session.Session
	scanner := bufio.NewScanner(f)

	type indexRecord struct {
		ID         string `json:"id"`
		ThreadName string `json:"thread_name"`
		UpdatedAt  string `json:"updated_at"`
		FilePath   string `json:"file_path"`
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec indexRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil || rec.ID == "" {
			continue
		}

		title := rec.ThreadName
		if title == "" {
			title = "Untitled Session"
		}

		modTime, _ := time.Parse(time.RFC3339, rec.UpdatedAt)
		if modTime.IsZero() {
			modTime = time.Now()
		}

		s := session.NewSession(rec.ID, title, "codex", profileName, isHost, modTime)
		s.StoragePath = rec.FilePath
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (p *Provider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	sessions, err := p.ListSessions(ctx, profileDir, isHost)
	if err != nil {
		return nil, err
	}

	for _, s := range sessions {
		if strings.HasPrefix(s.ID, idOrPrefix) {
			match := s
			return &match, nil
		}
	}
	return nil, nil
}

func (p *Provider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	if srcSession == nil {
		return "", fmt.Errorf("source session is nil")
	}

	targetCodexDir := filepath.Join(destProfileDir, ".codex")
	if err := os.MkdirAll(targetCodexDir, 0700); err != nil {
		return "", err
	}

	targetSessionsDir := filepath.Join(targetCodexDir, "sessions")
	_ = os.MkdirAll(targetSessionsDir, 0755)

	targetID := srcSession.ID
	if fork {
		targetID = generateUUID()
	}

	targetRolloutPath := srcSession.StoragePath
	if srcSession.StoragePath != "" {
		if fi, err := os.Stat(srcSession.StoragePath); err == nil && !fi.IsDir() {
			baseName := filepath.Base(srcSession.StoragePath)
			if fork {
				baseName = fmt.Sprintf("rollout-%s.jsonl", targetID)
			}
			destPath := filepath.Join(targetSessionsDir, baseName)
			if err := copyFile(srcSession.StoragePath, destPath); err == nil {
				targetRolloutPath = destPath
			}
		}
	}

	targetDB := filepath.Join(targetCodexDir, "state_5.sqlite")
	if p.sqliteBin != "" {
		schema := `
CREATE TABLE IF NOT EXISTS threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	title TEXT NOT NULL,
	preview TEXT NOT NULL DEFAULT '',
	source TEXT NOT NULL DEFAULT 'cli',
	model_provider TEXT NOT NULL DEFAULT 'openai',
	cwd TEXT NOT NULL DEFAULT '',
	sandbox_policy TEXT NOT NULL DEFAULT 'workspace-write',
	approval_mode TEXT NOT NULL DEFAULT 'ask'
);`
		_ = exec.CommandContext(ctx, p.sqliteBin, targetDB, schema).Run()

		escapedID := strings.ReplaceAll(targetID, "'", "''")
		escapedPath := strings.ReplaceAll(targetRolloutPath, "'", "''")
		escapedTitle := strings.ReplaceAll(srcSession.Title, "'", "''")
		escapedPreview := strings.ReplaceAll(srcSession.Summary, "'", "''")
		escapedProfileDir := strings.ReplaceAll(destProfileDir, "'", "''")
		unixNow := time.Now().Unix()

		insertQuery := fmt.Sprintf(
			"INSERT OR REPLACE INTO threads (id, rollout_path, created_at, updated_at, title, preview, source, model_provider, cwd, sandbox_policy, approval_mode) VALUES ('%s', '%s', %d, %d, '%s', '%s', 'cli', 'openai', '%s', 'workspace-write', 'ask');",
			escapedID, escapedPath, unixNow, unixNow, escapedTitle, escapedPreview, escapedProfileDir,
		)
		if err := exec.CommandContext(ctx, p.sqliteBin, targetDB, insertQuery).Run(); err != nil {
			logger.Debug("[session/codex] failed to insert thread into %s: %v", targetDB, err)
		}
	}

	targetIndex := filepath.Join(targetCodexDir, "session_index.jsonl")
	rec := map[string]interface{}{
		"id":          targetID,
		"thread_name": srcSession.Title,
		"updated_at":  time.Now().Format(time.RFC3339),
		"file_path":   targetRolloutPath,
	}
	if b, err := json.Marshal(rec); err == nil {
		if f, err := os.OpenFile(targetIndex, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
			_, _ = f.Write(append(b, '\n'))
			_ = f.Close()
		}
	}

	return targetID, nil
}

func generateUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
