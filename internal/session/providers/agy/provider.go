package agy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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

	query := "SELECT conversation_id, title, preview, last_modified_time, workspace_uris FROM conversation_summaries ORDER BY last_modified_time DESC LIMIT 50;\n"
	cmd := exec.CommandContext(ctx, p.sqliteBin, db, "-separator", "|||")
	cmd.Stdin = strings.NewReader(query)
	out, err := cmd.Output()
	if err != nil {
		// Fallback for older schemas where workspace_uris might not exist
		fallback := "SELECT conversation_id, title, preview, last_modified_time FROM conversation_summaries ORDER BY last_modified_time DESC LIMIT 50;\n"
		cmdFallback := exec.CommandContext(ctx, p.sqliteBin, db, "-separator", "|||")
		cmdFallback.Stdin = strings.NewReader(fallback)
		out, err = cmdFallback.Output()
		if err != nil {
			logger.Debug("[session/agy] query error on %s: %v", db, err)
			return nil, fmt.Errorf("failed to query sqlite db at %s: %w", db, err)
		}
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
		title := session.CleanPromptText(strings.TrimSpace(parts[1]))
		preview := strings.TrimSpace(parts[2])
		rawTime := strings.TrimSpace(parts[3])
		rawWorkspace := ""
		if len(parts) >= 5 {
			rawWorkspace = strings.TrimSpace(parts[4])
		}

		brainDir := p.brainDir(profileDir, isHost, convID)
		cwd := parseWorkspaceURIs(rawWorkspace)
		if cwd == "" {
			cwd = extractTranscriptCwd(brainDir)
		}

		modTime := parseTimeString(rawTime)
		s := session.NewSession(convID, title, "agy", profileName, isHost, modTime)
		s.StoragePath = brainDir
		s.Cwd = cwd
		s.Summary = preview

		if sum, err := p.ResolveSummary(ctx, &s); err == nil {
			s.Goal = sum.Goal
			s.Progress = sum.Progress
			s.Recent = sum.RecentActivity
			if sum.Text() != "" {
				s.Summary = sum.Text()
			}
			if s.Title == "" || s.Title == "Untitled Session" {
				if sum.Goal != "" {
					s.Title = sum.Goal
				} else if sum.RecentActivity != "" {
					s.Title = sum.RecentActivity
				}
			}
		}

		if s.Title == "" || s.Title == "Untitled Session" {
			if preview != "" {
				cleanPrev := session.CleanPromptText(strings.Split(preview, "\n")[0])
				if cleanPrev != "" {
					s.Title = cleanPrev
				}
			}
			if s.Title == "" {
				if s.Goal != "" {
					s.Title = s.Goal
				} else if s.Recent != "" {
					s.Title = s.Recent
				} else {
					s.Title = "Untitled Session"
				}
			}
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (p *Provider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	if p.sqliteBin == "" || idOrPrefix == "" {
		return nil, nil
	}

	if !isValidSessionID(idOrPrefix) {
		return nil, nil
	}

	db := p.dbPath(profileDir, isHost)
	if fi, err := os.Stat(db); err != nil || fi.Size() == 0 {
		return nil, nil
	}

	prefixLen := len(idOrPrefix)
	query := fmt.Sprintf("SELECT conversation_id, title, preview, last_modified_time, workspace_uris FROM conversation_summaries WHERE substr(conversation_id, 1, %d) = '%s' LIMIT 2;\n", prefixLen, idOrPrefix)
	cmd := exec.CommandContext(ctx, p.sqliteBin, db, "-separator", "|||")
	cmd.Stdin = strings.NewReader(query)
	out, err := cmd.Output()
	if err != nil {
		// Fallback for older schemas where workspace_uris might not exist
		fallback := fmt.Sprintf("SELECT conversation_id, title, preview, last_modified_time FROM conversation_summaries WHERE substr(conversation_id, 1, %d) = '%s' LIMIT 2;\n", prefixLen, idOrPrefix)
		cmdFallback := exec.CommandContext(ctx, p.sqliteBin, db, "-separator", "|||")
		cmdFallback.Stdin = strings.NewReader(fallback)
		out, err = cmdFallback.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to query session %s from %s: %w", idOrPrefix, db, err)
		}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var validLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			validLines = append(validLines, l)
		}
	}
	if len(validLines) == 0 {
		return nil, nil
	}

	if len(validLines) > 1 {
		var ids []string
		for _, l := range validLines {
			p := strings.Split(l, "|||")
			if len(p) >= 1 {
				id := strings.TrimSpace(p[0])
				if len(id) >= 8 {
					id = id[:8]
				}
				ids = append(ids, id)
			}
		}
		return nil, fmt.Errorf("ambiguous prefix %q matches multiple sessions in %s: %s", idOrPrefix, profileDir, strings.Join(ids, ", "))
	}

	profileName := filepath.Base(profileDir)
	if isHost {
		profileName = "<host>"
	}

	parts := strings.Split(validLines[0], "|||")
	if len(parts) < 4 {
		return nil, nil
	}

	convID := strings.TrimSpace(parts[0])
	title := strings.TrimSpace(parts[1])
	preview := strings.TrimSpace(parts[2])
	rawTime := strings.TrimSpace(parts[3])
	rawWorkspace := ""
	if len(parts) >= 5 {
		rawWorkspace = strings.TrimSpace(parts[4])
	}

	brainDir := p.brainDir(profileDir, isHost, convID)
	cwd := parseWorkspaceURIs(rawWorkspace)
	if cwd == "" {
		cwd = extractTranscriptCwd(brainDir)
	}

	modTime := parseTimeString(rawTime)
	s := session.NewSession(convID, title, "agy", profileName, isHost, modTime)
	s.StoragePath = brainDir
	s.Cwd = cwd
	s.Summary = preview

	if sum, err := p.ResolveSummary(ctx, &s); err == nil {
		s.Goal = sum.Goal
		s.Progress = sum.Progress
		s.Recent = sum.RecentActivity
		if sum.Text() != "" {
			s.Summary = sum.Text()
		}
		if s.Title == "" || s.Title == "Untitled Session" {
			if sum.Goal != "" {
				s.Title = sum.Goal
			} else if sum.RecentActivity != "" {
				s.Title = sum.RecentActivity
			}
		}
	}

	if s.Title == "" {
		if preview != "" {
			s.Title = session.CleanPromptText(strings.Split(preview, "\n")[0])
		}
		if s.Title == "" {
			s.Title = "Untitled Session"
		}
	}
	return &s, nil
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

func (p *Provider) brainDir(profileDir string, isHost bool, convID string) string {
	if isHost {
		return filepath.Join(config.RealHomeDir(), ".gemini", "antigravity-cli", "brain", convID)
	}
	return filepath.Join(profileDir, ".gemini", "antigravity-cli", "brain", convID)
}

func extractTranscriptSummary(brainDir string) string {
	tPath := filepath.Join(brainDir, ".system_generated", "logs", "transcript.jsonl")
	f, err := os.Open(tPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	// Read up to 32KB to find the initial user request or summary
	buf := make([]byte, 32768)
	n, err := f.Read(buf)
	if n == 0 || (err != nil && err != io.EOF) {
		return ""
	}
	rawStr := string(buf[:n])
	line := rawStr
	if idx := strings.Index(rawStr, "\n"); idx != -1 {
		line = rawStr[:idx]
	}

	var step struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(line), &step); err != nil {
		return ""
	}

	raw := step.Content
	if raw == "" {
		return ""
	}

	// 1. Look for <summary>...</summary>
	if start := strings.Index(raw, "<summary>"); start != -1 {
		rest := raw[start+9:]
		if end := strings.Index(rest, "</summary>"); end != -1 {
			return session.CleanPromptText(rest[:end])
		}
	}

	// 2. Look for <USER_REQUEST>...</USER_REQUEST>
	if start := strings.Index(raw, "<USER_REQUEST>"); start != -1 {
		rest := raw[start+14:]
		if end := strings.Index(rest, "</USER_REQUEST>"); end != -1 {
			return session.CleanPromptText(rest[:end])
		}
		return session.CleanPromptText(rest)
	}

	return session.CleanPromptText(raw)
}

// ResolveCwd resolves the working directory for an Antigravity conversation session.
func (p *Provider) ResolveCwd(ctx context.Context, s *session.Session) (string, error) {
	if s == nil {
		return "", nil
	}
	if s.Cwd != "" {
		return s.Cwd, nil
	}

	profileDir := ""
	if !s.IsHost && s.Profile != "" {
		profileDir = filepath.Join(config.BaseDir(), "profiles", s.Profile)
	}

	// 1. Check sqlite conversation_summaries.db
	db := p.dbPath(profileDir, s.IsHost)
	if p.sqliteBin != "" {
		query := fmt.Sprintf("SELECT workspace_uris FROM conversation_summaries WHERE conversation_id = '%s' LIMIT 1;\n", escapeSQL(s.ID))
		cmd := exec.CommandContext(ctx, p.sqliteBin, db)
		cmd.Stdin = strings.NewReader(query)
		if out, err := cmd.Output(); err == nil {
			if cwd := parseWorkspaceURIs(string(out)); cwd != "" {
				s.Cwd = cwd
				return cwd, nil
			}
		}
	}

	// 2. Fallback to transcript.jsonl
	brainDir := s.StoragePath
	if brainDir == "" {
		brainDir = p.brainDir(profileDir, s.IsHost, s.ID)
	}
	if cwd := extractTranscriptCwd(brainDir); cwd != "" {
		s.Cwd = cwd
		return cwd, nil
	}

	return "", nil
}

// ResolveSummary resolves structured summary information (Goal, Progress, RecentActivity) for an Antigravity session.
func (p *Provider) ResolveSummary(ctx context.Context, s *session.Session) (session.SessionSummary, error) {
	if s == nil {
		return session.SessionSummary{}, nil
	}

	profileDir := ""
	if !s.IsHost && s.Profile != "" {
		profileDir = filepath.Join(config.BaseDir(), "profiles", s.Profile)
	}

	brainDir := s.StoragePath
	if brainDir == "" {
		brainDir = p.brainDir(profileDir, s.IsHost, s.ID)
	}

	sum := session.SessionSummary{
		Raw: s.Summary,
	}

	// 1. Goal: task.md.metadata.json summary
	metaPath := filepath.Join(brainDir, "task.md.metadata.json")
	if data, err := os.ReadFile(metaPath); err == nil {
		var meta struct {
			Summary string `json:"summary"`
		}
		if err := json.Unmarshal(data, &meta); err == nil && strings.TrimSpace(meta.Summary) != "" {
			sum.Goal = session.CleanPromptText(meta.Summary)
		}
	}

	// 2. Goal fallback: task.md heading
	taskPath := filepath.Join(brainDir, "task.md")
	if sum.Goal == "" {
		if data, err := os.ReadFile(taskPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, l := range lines {
				l = strings.TrimSpace(l)
				if strings.HasPrefix(l, "# ") {
					h := strings.TrimPrefix(l, "# ")
					h = strings.TrimPrefix(h, "Task Checklist:")
					h = strings.TrimPrefix(h, "Task:")
					sum.Goal = session.CleanPromptText(h)
					break
				}
			}
		}
	}

	// 3. Goal fallback: extract from transcript if present
	if sum.Goal == "" {
		if trSummary := extractTranscriptSummary(brainDir); trSummary != "" {
			sum.Goal = trSummary
		}
	}

	// 4. Progress: parse checklist in task.md
	if data, err := os.ReadFile(taskPath); err == nil {
		lines := strings.Split(string(data), "\n")
		done, total := 0, 0
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if strings.HasPrefix(l, "- [x]") || strings.HasPrefix(l, "- [X]") {
				done++
				total++
			} else if strings.HasPrefix(l, "- [ ]") {
				total++
			}
		}
		if total > 0 {
			pct := (done * 100) / total
			sum.Progress = fmt.Sprintf("%d/%d tasks completed (%d%%)", done, total, pct)
		}
	}

	// 5. RecentActivity: seek transcript.jsonl near end
	if recent := extractTranscriptRecent(brainDir); recent != "" {
		sum.RecentActivity = recent
	}

	return sum, nil
}

func extractTranscriptRecent(brainDir string) string {
	tPath := filepath.Join(brainDir, ".system_generated", "logs", "transcript.jsonl")
	f, err := os.Open(tPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil || fi.Size() == 0 {
		return ""
	}

	readSize := int64(65536)
	offset := int64(0)
	if fi.Size() > readSize {
		offset = fi.Size() - readSize
	}
	buf := make([]byte, fi.Size()-offset)
	if _, err := f.ReadAt(buf, offset); err != nil && err != io.EOF {
		return ""
	}

	lines := strings.Split(string(buf), "\n")
	if offset > 0 && len(lines) > 0 {
		lines = lines[1:] // Discard truncated partial first line
	}
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if !strings.Contains(line, `"type":"USER_INPUT"`) {
			continue
		}
		var step struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(line), &step); err == nil && step.Content != "" {
			cleaned := session.CleanPromptText(step.Content)
			if cleaned != "" {
				return cleaned
			}
		}
	}
	return ""
}

func extractTranscriptCwd(brainDir string) string {
	tPath := filepath.Join(brainDir, ".system_generated", "logs", "transcript.jsonl")
	f, err := os.Open(tPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, 32768)
	n, err := f.Read(buf)
	if n == 0 || (err != nil && err != io.EOF) {
		return ""
	}
	content := string(buf[:n])

	if idx := strings.Index(content, "Command Working Directory: "); idx != -1 {
		rest := content[idx+len("Command Working Directory: "):]
		if end := strings.IndexAny(rest, "\r\n<"); end != -1 {
			return strings.TrimSpace(rest[:end])
		}
	}
	if idx := strings.Index(content, "Active Workspaces:\n- "); idx != -1 {
		rest := content[idx+len("Active Workspaces:\n- "):]
		if end := strings.IndexAny(rest, "\r\n<"); end != -1 {
			return strings.TrimSpace(rest[:end])
		}
	}
	return ""
}

func escapeSQL(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.ReplaceAll(s, "'", "''")
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

func parseWorkspaceURIs(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == `""` || raw == "[]" {
		return ""
	}

	var uris []string
	if err := json.Unmarshal([]byte(raw), &uris); err == nil && len(uris) > 0 {
		for _, u := range uris {
			path := cleanURIPath(u)
			if path != "" && !strings.Contains(path, "/.") {
				return path
			}
		}
		for _, u := range uris {
			if path := cleanURIPath(u); path != "" {
				return path
			}
		}
		return ""
	}

	return cleanURIPath(raw)
}

func cleanURIPath(u string) string {
	u = strings.TrimSpace(u)
	u = strings.Trim(u, "\"")
	if u == "" {
		return ""
	}
	if strings.HasPrefix(u, "file://") {
		if parsed, err := url.Parse(u); err == nil && parsed.Path != "" {
			return filepath.Clean(parsed.Path)
		}
		u = strings.TrimPrefix(u, "file://")
	}
	if unescaped, err := url.PathUnescape(u); err == nil {
		u = unescaped
	}
	return filepath.Clean(u)
}
