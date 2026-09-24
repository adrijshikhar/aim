package codex

import (
	"bufio"
	"bytes"
	"context"
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
	query := "SELECT id, title, preview, updated_at, rollout_path FROM threads ORDER BY updated_at DESC LIMIT 50;\n"
	cmd := exec.CommandContext(ctx, p.sqliteBin, dbPath, "-separator", "|||")
	cmd.Stdin = strings.NewReader(query)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite db at %s: %w", dbPath, err)
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
		return nil, fmt.Errorf("failed to open session index at %s: %w", indexPath, err)
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

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan session index at %s: %w", indexPath, err)
	}

	return sessions, nil
}

func (p *Provider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	sessions, err := p.ListSessions(ctx, profileDir, isHost)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions for codex: %w", err)
	}

	var matches []*session.Session
	for _, s := range sessions {
		if strings.HasPrefix(s.ID, idOrPrefix) {
			match := s
			matches = append(matches, &match)
		}
	}

	if len(matches) > 1 {
		var ids []string
		for _, m := range matches {
			ids = append(ids, m.ShortID)
		}
		return nil, fmt.Errorf("ambiguous prefix %q matches multiple sessions in %s: %s", idOrPrefix, profileDir, strings.Join(ids, ", "))
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	return nil, nil
}

func (p *Provider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	if srcSession == nil {
		return "", fmt.Errorf("source session is nil")
	}

	targetID := srcSession.ID
	if fork {
		var err error
		targetID, err = session.GenerateUUID()
		if err != nil {
			return "", fmt.Errorf("failed to generate uuid for forked session: %w", err)
		}
	}

	if !isValidSessionID(targetID) {
		return "", fmt.Errorf("invalid session ID %q", targetID)
	}
	if !isValidSessionID(srcSession.ID) {
		return "", fmt.Errorf("invalid source session ID %q", srcSession.ID)
	}

	targetCodexDir := filepath.Join(destProfileDir, ".codex")
	if err := os.MkdirAll(targetCodexDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create target codex directory %s: %w", targetCodexDir, err)
	}

	targetSessionsDir := filepath.Join(targetCodexDir, "sessions")
	if err := os.MkdirAll(targetSessionsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target sessions directory %s: %w", targetSessionsDir, err)
	}

	srcCodexDir := p.resolveSrcCodexDir(srcSession)

	// 1. Determine and copy rollout file
	srcRolloutPath := p.resolveSrcRolloutPath(ctx, srcCodexDir, srcSession)
	var targetRolloutPath string
	if srcRolloutPath != "" {
		var err error
		targetRolloutPath, err = p.hydrateRolloutFile(srcCodexDir, targetSessionsDir, srcRolloutPath, srcSession.ID, targetID, fork)
		if err != nil {
			return "", err
		}

		// 2. Recursively copy ancestor/parent sessions
		if srcCodexDir != "" {
			if err := p.hydrateAncestorSessions(ctx, srcCodexDir, targetCodexDir, targetSessionsDir, srcRolloutPath, targetID); err != nil {
				logger.Debug("[session/codex] error copying ancestor sessions: %v", err)
			}
		}
	}

	// 3. Hydrate state and history databases for the leaf session
	if p.sqliteBin != "" {
		if err := p.hydrateDatabases(ctx, srcCodexDir, targetCodexDir, srcSession.ID, targetID, targetRolloutPath, destProfileDir, srcSession); err != nil {
			logger.Debug("[session/codex] error hydrating leaf databases: %v", err)
		}
	}

	// 4. Update session_index.jsonl
	if err := p.appendSessionIndex(targetCodexDir, targetID, srcSession.Title, targetRolloutPath); err != nil {
		return "", err
	}

	return targetID, nil
}

func (p *Provider) resolveSrcRolloutPath(ctx context.Context, srcCodexDir string, srcSession *session.Session) string {
	srcRolloutPath := srcSession.StoragePath
	if srcRolloutPath != "" {
		if fi, err := os.Stat(srcRolloutPath); err != nil || fi.IsDir() {
			srcRolloutPath = ""
		}
	}
	if srcRolloutPath == "" && srcCodexDir != "" {
		srcRolloutPath = p.findRolloutPath(ctx, srcCodexDir, srcSession.ID)
	}
	return srcRolloutPath
}

func (p *Provider) hydrateRolloutFile(srcCodexDir, targetSessionsDir, srcRolloutPath, srcID, targetID string, fork bool) (string, error) {
	var srcSessionsDir string
	if srcCodexDir != "" {
		srcSessionsDir = filepath.Join(srcCodexDir, "sessions")
	}

	var relPath string
	if srcSessionsDir != "" {
		if rel, err := filepath.Rel(srcSessionsDir, srcRolloutPath); err == nil && !strings.HasPrefix(rel, "..") {
			relPath = rel
		}
	}
	if relPath == "" {
		relPath = filepath.Base(srcRolloutPath)
	}

	if fork {
		origBase := filepath.Base(relPath)
		var newBase string
		if strings.Contains(origBase, srcID) {
			newBase = strings.Replace(origBase, srcID, targetID, 1)
		} else {
			newBase = fmt.Sprintf("rollout-%s.jsonl", targetID)
		}
		relPath = filepath.Join(filepath.Dir(relPath), newBase)
	}

	destPath := filepath.Join(targetSessionsDir, relPath)
	oldID := ""
	newID := ""
	if fork {
		oldID = srcID
		newID = targetID
	}
	if err := copyFileWithReplace(srcRolloutPath, destPath, oldID, newID); err != nil {
		return "", fmt.Errorf("failed to copy rollout file to %s: %w", destPath, err)
	}
	return destPath, nil
}

func (p *Provider) hydrateAncestorSessions(ctx context.Context, srcCodexDir, targetCodexDir, targetSessionsDir, srcRolloutPath, targetID string) error {
	srcSessionsDir := filepath.Join(srcCodexDir, "sessions")
	visitedAncestors := map[string]bool{targetID: true}
	currentParentID := extractParentThreadID(srcRolloutPath)

	for currentParentID != "" && !visitedAncestors[currentParentID] {
		visitedAncestors[currentParentID] = true
		parentRollout := p.findRolloutPath(ctx, srcCodexDir, currentParentID)
		if parentRollout == "" {
			break
		}

		var destParentRel string
		if rel, err := filepath.Rel(srcSessionsDir, parentRollout); err == nil && !strings.HasPrefix(rel, "..") {
			destParentRel = rel
		}
		if destParentRel == "" {
			destParentRel = filepath.Base(parentRollout)
		}
		destParentRollout := filepath.Join(targetSessionsDir, destParentRel)

		srcStat, srcErr := os.Stat(parentRollout)
		destStat, destErr := os.Stat(destParentRollout)
		shouldCopy := destErr != nil || (srcErr == nil && (srcStat.Size() > destStat.Size() || srcStat.ModTime().After(destStat.ModTime())))
		if shouldCopy {
			if err := copyFileWithReplace(parentRollout, destParentRollout, "", ""); err != nil {
				logger.Debug("[session/codex] failed to copy parent rollout %s: %v", parentRollout, err)
			}
		}

		if p.sqliteBin != "" {
			srcDB := filepath.Join(srcCodexDir, "state_5.sqlite")
			targetDB := filepath.Join(targetCodexDir, "state_5.sqlite")
			if _, err := os.Stat(srcDB); err == nil {
				if err := p.copyThreadInStateDB(ctx, srcDB, targetDB, currentParentID, currentParentID, destParentRollout); err != nil {
					logger.Debug("[session/codex] failed to copy ancestor thread %s to %s: %v", currentParentID, targetDB, err)
				}
			}

			srcHistoryDB := filepath.Join(srcCodexDir, "thread_history_1.sqlite")
			targetHistoryDB := filepath.Join(targetCodexDir, "thread_history_1.sqlite")
			if _, err := os.Stat(srcHistoryDB); err == nil {
				if err := p.copyThreadHistoryDB(ctx, srcHistoryDB, targetHistoryDB, currentParentID, currentParentID); err != nil {
					logger.Debug("[session/codex] failed to copy ancestor thread history %s to %s: %v", currentParentID, targetHistoryDB, err)
				}
			}
		}

		currentParentID = extractParentThreadID(parentRollout)
	}
	return nil
}

func (p *Provider) hydrateDatabases(ctx context.Context, srcCodexDir, targetCodexDir, srcID, targetID, targetRolloutPath, destProfileDir string, srcSession *session.Session) error {
	targetDB := filepath.Join(targetCodexDir, "state_5.sqlite")
	var srcDB string
	if srcCodexDir != "" {
		srcDB = filepath.Join(srcCodexDir, "state_5.sqlite")
	}

	copied := false
	if srcDB != "" {
		if _, err := os.Stat(srcDB); err == nil {
			if err := p.copyThreadInStateDB(ctx, srcDB, targetDB, srcID, targetID, targetRolloutPath); err == nil {
				copied = true
			} else {
				logger.Debug("[session/codex] copyThreadInStateDB error, falling back: %v", err)
			}
		}
	}

	if !copied {
		p.fallbackInsertThread(ctx, targetDB, targetID, targetRolloutPath, srcSession, destProfileDir)
	}

	// Copy thread_history_1.sqlite for leaf session
	if srcCodexDir != "" {
		srcHistoryDB := filepath.Join(srcCodexDir, "thread_history_1.sqlite")
		targetHistoryDB := filepath.Join(targetCodexDir, "thread_history_1.sqlite")
		if _, err := os.Stat(srcHistoryDB); err == nil {
			if err := p.copyThreadHistoryDB(ctx, srcHistoryDB, targetHistoryDB, srcID, targetID); err != nil {
				logger.Debug("[session/codex] copyThreadHistoryDB error: %v", err)
			}
		}
	}
	return nil
}

func (p *Provider) appendSessionIndex(targetCodexDir, targetID, title, targetRolloutPath string) error {
	targetIndex := filepath.Join(targetCodexDir, "session_index.jsonl")
	rec := map[string]interface{}{
		"id":          targetID,
		"thread_name": title,
		"updated_at":  time.Now().Format(time.RFC3339),
		"file_path":   targetRolloutPath,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("failed to marshal session index record: %w", err)
	}

	f, err := os.OpenFile(targetIndex, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open session index file at %s: %w", targetIndex, err)
	}
	defer f.Close()

	if _, err := f.Write(append(b, '\n')); err != nil {
		return fmt.Errorf("failed to write to session index file at %s: %w", targetIndex, err)
	}
	return nil
}

func (p *Provider) resolveSrcCodexDir(srcSession *session.Session) string {
	if srcSession.IsHost {
		return filepath.Join(config.RealHomeDir(), ".codex")
	}
	if srcSession.Profile != "" && srcSession.Profile != "<host>" && isValidProfileName(srcSession.Profile) {
		baseProfilesDir := filepath.Clean(filepath.Join(config.BaseDir(), "profiles"))
		cand := filepath.Clean(filepath.Join(baseProfilesDir, srcSession.Profile, ".codex"))
		if strings.HasPrefix(cand, baseProfilesDir+string(filepath.Separator)) {
			if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
				return cand
			}
		}
	}
	if srcSession.StoragePath != "" {
		marker := string(filepath.Separator) + ".codex" + string(filepath.Separator)
		if idx := strings.Index(srcSession.StoragePath, marker); idx != -1 {
			cand := filepath.Clean(srcSession.StoragePath[:idx+len(string(filepath.Separator)+".codex")])
			if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
				return cand
			}
		}
	}
	return ""
}

func (p *Provider) findRolloutPath(ctx context.Context, codexDir, sessionID string) string {
	if p.sqliteBin != "" {
		dbPath := filepath.Join(codexDir, "state_5.sqlite")
		if _, err := os.Stat(dbPath); err == nil {
			query := fmt.Sprintf("SELECT rollout_path FROM threads WHERE id = '%s' LIMIT 1;\n", escapeSQL(sessionID))
			cmd := exec.CommandContext(ctx, p.sqliteBin, dbPath)
			cmd.Stdin = strings.NewReader(query)
			if out, err := cmd.Output(); err == nil {
				path := strings.TrimSpace(string(out))
				if path != "" {
					if _, statErr := os.Stat(path); statErr == nil {
						return path
					}
				}
			}
		}
	}
	// Fallback to globbing in sessions dir
	sessionsDir := filepath.Join(codexDir, "sessions")
	var match string
	walkErr := filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.Contains(info.Name(), sessionID) && strings.HasSuffix(info.Name(), ".jsonl") {
			match = path
			return filepath.SkipAll
		}
		return nil
	})
	if walkErr != nil {
		logger.Debug("[session/codex] error walking sessions directory %s: %v", sessionsDir, walkErr)
	}
	return match
}

func (p *Provider) getTableColumns(ctx context.Context, dbPath, table string) ([]string, error) {
	if !isValidIdentifier(table) {
		return nil, fmt.Errorf("invalid table name %q", table)
	}
	query := fmt.Sprintf("PRAGMA table_info(%s);\n", table)
	cmd := exec.CommandContext(ctx, p.sqliteBin, dbPath, "-separator", "|")
	cmd.Stdin = strings.NewReader(query)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query table info for %s in %s: %w", table, dbPath, err)
	}
	var cols []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			colName := strings.TrimSpace(parts[1])
			if colName != "" && isValidIdentifier(colName) {
				cols = append(cols, colName)
			}
		}
	}
	return cols, nil
}

func (p *Provider) ensureDBInitialized(ctx context.Context, srcDB, destDB string) error {
	if _, err := os.Stat(destDB); err == nil {
		return nil // already exists
	}
	if _, err := os.Stat(srcDB); err != nil {
		return nil // srcDB does not exist, cannot clone
	}

	cmd := exec.CommandContext(ctx, p.sqliteBin, srcDB, fmt.Sprintf("VACUUM INTO '%s';", escapeSQL(destDB)))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to vacuum clone db %s to %s: %s: %w", srcDB, destDB, string(out), err)
	}

	clearSQL := "PRAGMA foreign_keys = OFF;\nBEGIN TRANSACTION;\n"
	tblCmd := exec.CommandContext(ctx, p.sqliteBin, destDB, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name != '_sqlx_migrations';")
	out, err := tblCmd.Output()
	if err == nil {
		for _, tbl := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			tbl = strings.TrimSpace(tbl)
			if tbl != "" && isValidIdentifier(tbl) {
				clearSQL += fmt.Sprintf("DELETE FROM %s;\n", tbl)
			}
		}
	}
	clearSQL += "COMMIT;\n"

	execCmd := exec.CommandContext(ctx, p.sqliteBin, destDB)
	execCmd.Stdin = strings.NewReader(clearSQL)
	if out, err := execCmd.CombinedOutput(); err != nil {
		logger.Debug("[session/codex] failed to clear cloned db %s: %s: %v", destDB, string(out), err)
	}
	return nil
}

func (p *Provider) copyThreadInStateDB(ctx context.Context, srcDB, destDB, srcID, targetID, targetRolloutPath string) error {
	if err := p.ensureDBInitialized(ctx, srcDB, destDB); err != nil {
		logger.Debug("[session/codex] ensureDBInitialized error: %v", err)
	}

	checkCmd := exec.CommandContext(ctx, p.sqliteBin, srcDB, fmt.Sprintf("SELECT COUNT(*) FROM threads WHERE id = '%s';", escapeSQL(srcID)))
	out, err := checkCmd.Output()
	if err != nil || strings.TrimSpace(string(out)) == "0" {
		return fmt.Errorf("thread %s not found in source db %s", srcID, srcDB)
	}

	destCols, err := p.getTableColumns(ctx, destDB, "threads")
	if err != nil || len(destCols) == 0 {
		return fmt.Errorf("destination threads table not found or empty columns: %w", err)
	}

	srcCols, err := p.getTableColumns(ctx, srcDB, "threads")
	if err != nil || len(srcCols) == 0 {
		return fmt.Errorf("source threads table not found or empty columns: %w", err)
	}

	srcColSet := make(map[string]bool)
	for _, sc := range srcCols {
		srcColSet[sc] = true
	}

	var insertCols []string
	var selectExprs []string
	for _, col := range destCols {
		if !srcColSet[col] {
			continue
		}
		insertCols = append(insertCols, fmt.Sprintf(`"%s"`, col))
		switch col {
		case "id":
			selectExprs = append(selectExprs, fmt.Sprintf("'%s'", escapeSQL(targetID)))
		case "rollout_path":
			selectExprs = append(selectExprs, fmt.Sprintf("'%s'", escapeSQL(targetRolloutPath)))
		default:
			selectExprs = append(selectExprs, fmt.Sprintf(`"%s"`, col))
		}
	}

	if len(insertCols) == 0 {
		return fmt.Errorf("no common columns found for threads table")
	}

	sql := fmt.Sprintf(`
PRAGMA foreign_keys = OFF;
ATTACH DATABASE '%s' AS src;
BEGIN TRANSACTION;
INSERT OR REPLACE INTO threads (%s)
SELECT %s
FROM src.threads WHERE id = '%s';
COMMIT;
DETACH DATABASE src;
`, escapeSQL(srcDB), strings.Join(insertCols, ", "), strings.Join(selectExprs, ", "), escapeSQL(srcID))

	cmd := exec.CommandContext(ctx, p.sqliteBin, destDB)
	cmd.Stdin = strings.NewReader(sql)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy thread %s to %s: %s: %w", srcID, destDB, string(out), err)
	}
	return nil
}

func (p *Provider) copyThreadHistoryDB(ctx context.Context, srcHistoryDB, destHistoryDB, srcID, targetID string) error {
	if _, err := os.Stat(srcHistoryDB); err != nil {
		return nil // No source history DB, nothing to copy
	}

	if err := p.ensureDBInitialized(ctx, srcHistoryDB, destHistoryDB); err != nil {
		logger.Debug("[session/codex] ensureDBInitialized error for history DB: %v", err)
	}

	tables := []string{"thread_turns", "thread_items", "thread_history_projection_state", "thread_realtime_items"}
	var sqlParts []string

	for _, tbl := range tables {
		destCols, err := p.getTableColumns(ctx, destHistoryDB, tbl)
		if err != nil || len(destCols) == 0 {
			continue
		}
		srcCols, err := p.getTableColumns(ctx, srcHistoryDB, tbl)
		if err != nil || len(srcCols) == 0 {
			continue
		}

		srcColSet := make(map[string]bool)
		for _, sc := range srcCols {
			srcColSet[sc] = true
		}

		var insertCols []string
		var selectExprs []string
		hasThreadID := false

		for _, col := range destCols {
			if !srcColSet[col] {
				continue
			}
			insertCols = append(insertCols, fmt.Sprintf(`"%s"`, col))
			if col == "thread_id" {
				hasThreadID = true
				selectExprs = append(selectExprs, fmt.Sprintf("'%s'", escapeSQL(targetID)))
			} else {
				selectExprs = append(selectExprs, fmt.Sprintf(`"%s"`, col))
			}
		}

		if len(insertCols) == 0 || !hasThreadID {
			continue
		}

		part := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s)\nSELECT %s\nFROM src.%s WHERE thread_id = '%s';",
			tbl, strings.Join(insertCols, ", "), strings.Join(selectExprs, ", "), tbl, escapeSQL(srcID))
		sqlParts = append(sqlParts, part)
	}

	if len(sqlParts) == 0 {
		return nil
	}

	fullSQL := fmt.Sprintf(`
PRAGMA foreign_keys = OFF;
ATTACH DATABASE '%s' AS src;
BEGIN TRANSACTION;
%s
COMMIT;
DETACH DATABASE src;
`, escapeSQL(srcHistoryDB), strings.Join(sqlParts, "\n"))

	cmd := exec.CommandContext(ctx, p.sqliteBin, destHistoryDB)
	cmd.Stdin = strings.NewReader(fullSQL)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy thread history for %s: %s: %w", srcID, string(out), err)
	}
	return nil
}

func (p *Provider) fallbackInsertThread(ctx context.Context, targetDB, targetID, targetRolloutPath string, srcSession *session.Session, destProfileDir string) {
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
	schemaCmd := exec.CommandContext(ctx, p.sqliteBin, targetDB)
	schemaCmd.Stdin = strings.NewReader(schema)
	if err := schemaCmd.Run(); err != nil {
		logger.Debug("[session/codex] failed to initialize schema at %s: %v", targetDB, err)
	}

	unixNow := time.Now().Unix()
	insertQuery := fmt.Sprintf(
		"INSERT OR REPLACE INTO threads (id, rollout_path, created_at, updated_at, title, preview, source, model_provider, cwd, sandbox_policy, approval_mode) VALUES ('%s', '%s', %d, %d, '%s', '%s', 'cli', 'openai', '%s', 'workspace-write', 'ask');\n",
		escapeSQL(targetID), escapeSQL(targetRolloutPath), unixNow, unixNow, escapeSQL(srcSession.Title), escapeSQL(srcSession.Summary), escapeSQL(destProfileDir),
	)
	insertCmd := exec.CommandContext(ctx, p.sqliteBin, targetDB)
	insertCmd.Stdin = strings.NewReader(insertQuery)
	if err := insertCmd.Run(); err != nil {
		logger.Debug("[session/codex] failed to insert thread into %s: %v", targetDB, err)
	}
}

type rolloutMetaHeader struct {
	Type    string `json:"type"`
	Payload struct {
		ID          string `json:"id"`
		HistoryBase *struct {
			ThreadID string `json:"thread_id"`
		} `json:"history_base"`
	} `json:"payload"`
}

func extractParentThreadID(rolloutPath string) string {
	f, err := os.Open(rolloutPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	firstLine, err := reader.ReadBytes('\n')
	if err != nil && len(firstLine) == 0 {
		return ""
	}

	var meta rolloutMetaHeader
	if err := json.Unmarshal(firstLine, &meta); err != nil {
		return ""
	}
	if meta.Payload.HistoryBase != nil && isValidSessionID(meta.Payload.HistoryBase.ThreadID) {
		return meta.Payload.HistoryBase.ThreadID
	}
	return ""
}

func copyFileWithReplace(src, dst, oldID, newID string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source %s: %w", src, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for destination %s: %w", dst, err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination %s: %w", dst, err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close destination %s: %w", dst, closeErr)
		}
	}()

	if oldID == "" || newID == "" || oldID == newID {
		if _, err = io.Copy(out, in); err != nil {
			return fmt.Errorf("failed to copy data from %s to %s: %w", src, dst, err)
		}
	} else {
		oldBytes := []byte(oldID)
		newBytes := []byte(newID)
		reader := bufio.NewReader(in)
		writer := bufio.NewWriter(out)
		for {
			line, readErr := reader.ReadBytes('\n')
			if len(line) > 0 {
				replaced := bytes.ReplaceAll(line, oldBytes, newBytes)
				if _, wErr := writer.Write(replaced); wErr != nil {
					return fmt.Errorf("failed writing destination %s: %w", dst, wErr)
				}
			}
			if readErr != nil {
				if readErr != io.EOF {
					return fmt.Errorf("failed reading source %s: %w", src, readErr)
				}
				break
			}
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("failed flushing destination %s: %w", dst, err)
		}
	}

	if err = out.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination %s: %w", dst, err)
	}
	return nil
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

func escapeSQL(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.ReplaceAll(s, "'", "''")
}

func copyFile(src, dst string) (err error) {
	return copyFileWithReplace(src, dst, "", "")
}
