package agy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/session"
)

// Ensure Provider implements session.SessionProvider.
var _ session.SessionProvider = (*Provider)(nil)

const untitledSessionTitle = "Untitled Session"

// Provider implements session.SessionProvider for Antigravity CLI (agy).
type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
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

type agyConvRow struct {
	convID       string
	title        string
	preview      string
	rawTime      any
	rawWorkspace string
	stepCount    int
}

func queryAgyConvRows(ctx context.Context, dbPath, idPrefix string) ([]agyConvRow, error) {
	db, err := session.OpenReadOnlySQLite(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	where := ""
	limit := "LIMIT 50"
	var args []any
	if idPrefix != "" {
		where = ` WHERE conversation_id LIKE ? ESCAPE '\'`
		limit = "LIMIT 2"
		args = append(args, strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(idPrefix)+"%")
	}

	queries := []struct {
		sql          string
		hasWorkspace bool
		hasSteps     bool
	}{
		{"SELECT conversation_id, COALESCE(title,''), COALESCE(preview,''), last_modified_time, COALESCE(workspace_uris,''), COALESCE(step_count,0) FROM conversation_summaries" + where + " ORDER BY last_modified_time DESC " + limit, true, true},
		{"SELECT conversation_id, COALESCE(title,''), COALESCE(preview,''), last_modified_time, COALESCE(workspace_uris,'') FROM conversation_summaries" + where + " ORDER BY last_modified_time DESC " + limit, true, false},
		{"SELECT conversation_id, COALESCE(title,''), COALESCE(preview,''), last_modified_time FROM conversation_summaries" + where + " ORDER BY last_modified_time DESC " + limit, false, false},
	}

	var rows *sql.Rows
	var chosenHasWorkspace, chosenHasSteps bool
	var queryErr error
	for _, q := range queries {
		rows, queryErr = db.QueryContext(ctx, q.sql, args...)
		if queryErr == nil {
			chosenHasWorkspace = q.hasWorkspace
			chosenHasSteps = q.hasSteps
			break
		}
	}
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { _ = rows.Close() }()

	var out []agyConvRow
	for rows.Next() {
		var r agyConvRow
		var scanErr error
		switch {
		case chosenHasWorkspace && chosenHasSteps:
			scanErr = rows.Scan(&r.convID, &r.title, &r.preview, &r.rawTime, &r.rawWorkspace, &r.stepCount)
		case chosenHasWorkspace:
			scanErr = rows.Scan(&r.convID, &r.title, &r.preview, &r.rawTime, &r.rawWorkspace)
		default:
			scanErr = rows.Scan(&r.convID, &r.title, &r.preview, &r.rawTime)
		}
		if scanErr != nil {
			continue
		}
		r.convID = strings.TrimSpace(r.convID)
		if r.convID != "" {
			out = append(out, r)
		}
	}
	return out, rows.Err()
}

func parseAgyModTime(raw any) time.Time {
	switch v := raw.(type) {
	case time.Time:
		return v
	case string:
		return parseTimeString(strings.TrimSpace(v))
	case []byte:
		return parseTimeString(strings.TrimSpace(string(v)))
	default:
		return time.Time{}
	}
}

func (p *Provider) buildAgySession(row agyConvRow, profileDir, profileName string, isHost, cleanTitle bool) session.Session {
	convID := row.convID
	title := strings.TrimSpace(row.title)
	if cleanTitle {
		title = session.CleanPromptText(title)
	}
	preview := strings.TrimSpace(row.preview)
	brainDir := p.brainDir(profileDir, isHost, convID)
	cwd := parseWorkspaceURIs(strings.TrimSpace(row.rawWorkspace))
	if cwd == "" {
		cwd = extractTranscriptCwd(brainDir)
	}

	modTime := parseAgyModTime(row.rawTime)
	s := session.NewSession(convID, title, "agy", profileName, isHost, modTime)
	s.StoragePath = brainDir
	s.Cwd = cwd
	s.Summary = preview
	s.MessageCount = row.stepCount
	if s.MessageCount == 0 && brainDir != "" {
		s.MessageCount = countAgyBrainTurns(brainDir)
	}
	if s.MessageCount > 0 {
		s.Progress = fmt.Sprintf("%d turns", s.MessageCount)
	}
	return s
}

func countAgyBrainTurns(brainDir string) int {
	root, err := os.OpenRoot(brainDir)
	if err != nil {
		return 0
	}
	defer func() { _ = root.Close() }()
	f, err := root.Open(filepath.Join(".system_generated", "logs", "transcript.jsonl"))
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()
	tc, err := session.ReadTurnCounts(f)
	if err != nil || tc <= 0 {
		return 0
	}
	return tc
}

func applyAgyResolvedSummary(s *session.Session, sum session.SessionSummary) {
	s.Goal = sum.Goal
	s.Progress = sum.Progress
	s.Recent = sum.RecentActivity
	if sum.Text() != "" {
		s.Summary = sum.Text()
	}
	if s.Title != "" && s.Title != untitledSessionTitle {
		return
	}
	if sum.Goal != "" {
		s.Title = sum.Goal
	} else if sum.RecentActivity != "" {
		s.Title = sum.RecentActivity
	}
}

func (p *Provider) applyAgySummaryAndTitle(ctx context.Context, s *session.Session, preview string, fallbackToGoalRecent bool) {
	if sum, err := p.ResolveSummary(ctx, s); err == nil {
		applyAgyResolvedSummary(s, sum)
	}
	if s.Title != "" && s.Title != untitledSessionTitle {
		return
	}
	if preview != "" {
		if cleanPrev := session.CleanPromptText(strings.Split(preview, "\n")[0]); cleanPrev != "" {
			s.Title = cleanPrev
			return
		}
	}
	if fallbackToGoalRecent && s.Goal != "" {
		s.Title = s.Goal
		return
	}
	if fallbackToGoalRecent && s.Recent != "" {
		s.Title = s.Recent
		return
	}
	s.Title = untitledSessionTitle
}

func hasNonEmptyAgyDB(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Size() > 0
}

// ListSessions lists Antigravity sessions stored in conversation_summaries.db.
func (p *Provider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	db := p.dbPath(profileDir, isHost)
	if !hasNonEmptyAgyDB(db) {
		return nil, nil
	}

	convRows, err := queryAgyConvRows(ctx, db, "")
	if err != nil {
		logger.Debug("[session/agy] query error on %s: %v", db, err)
		return nil, fmt.Errorf("failed to query sqlite db at %s: %w", db, err)
	}

	profileName := filepath.Base(profileDir)
	if isHost {
		profileName = "<host>"
	}

	sessions := make([]session.Session, 0, len(convRows))
	for _, row := range convRows {
		s := p.buildAgySession(row, profileDir, profileName, isHost, true)
		p.applyAgySummaryAndTitle(ctx, &s, strings.TrimSpace(row.preview), true)
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (p *Provider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	if idOrPrefix == "" || !isValidSessionID(idOrPrefix) {
		return nil, nil
	}

	db := p.dbPath(profileDir, isHost)
	if !hasNonEmptyAgyDB(db) {
		return nil, nil
	}

	convRows, err := queryAgyConvRows(ctx, db, idOrPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to query session %s from %s: %w", idOrPrefix, db, err)
	}
	if len(convRows) == 0 {
		return nil, nil
	}
	if len(convRows) > 1 {
		var ids []string
		for _, r := range convRows {
			id := r.convID
			if len(id) >= 8 {
				id = id[:8]
			}
			ids = append(ids, id)
		}
		return nil, fmt.Errorf("ambiguous prefix %q matches multiple sessions in %s: %s", idOrPrefix, profileDir, strings.Join(ids, ", "))
	}

	profileName := filepath.Base(profileDir)
	if isHost {
		profileName = "<host>"
	}

	preview := strings.TrimSpace(convRows[0].preview)
	s := p.buildAgySession(convRows[0], profileDir, profileName, isHost, false)

	if sum, err := p.ResolveSummary(ctx, &s); err == nil {
		s.Goal = sum.Goal
		s.Progress = sum.Progress
		s.Recent = sum.RecentActivity
		if sum.Text() != "" {
			s.Summary = sum.Text()
		}
		if s.Title == "" || s.Title == untitledSessionTitle {
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
			s.Title = untitledSessionTitle
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

	// 1. Check sqlite conversation_summaries.db via parameterized read-only query
	dbPath := p.dbPath(profileDir, s.IsHost)
	if cwd := lookupAgyWorkspaceCwd(ctx, dbPath, s.ID); cwd != "" {
		return cwd, nil
	}

	// 2. Fallback to transcript.jsonl
	brainDir := s.StoragePath
	if brainDir == "" {
		brainDir = p.brainDir(profileDir, s.IsHost, s.ID)
	}
	if cwd := extractTranscriptCwd(brainDir); cwd != "" {
		return cwd, nil
	}

	return "", nil
}

func lookupAgyWorkspaceCwd(ctx context.Context, dbPath, convID string) string {
	if _, err := os.Stat(dbPath); err != nil {
		return ""
	}
	db, err := session.OpenReadOnlySQLite(ctx, dbPath)
	if err != nil {
		return ""
	}
	defer func() { _ = db.Close() }()
	var wsURIs sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT workspace_uris FROM conversation_summaries WHERE conversation_id = ? LIMIT 1", convID).Scan(&wsURIs); err != nil {
		return ""
	}
	return parseWorkspaceURIs(wsURIs.String)
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
