package codex

import (
	"context"
	"encoding/json"
	"fmt"
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

// ResolveCwd resolves the working directory for an OpenAI Codex CLI session.
func (p *Provider) ResolveCwd(ctx context.Context, s *session.Session) (string, error) {
	if s == nil {
		return "", nil
	}
	if s.Cwd != "" {
		return s.Cwd, nil
	}

	cDir := ""
	if s.StoragePath != "" {
		if idx := strings.Index(s.StoragePath, ".codex"); idx != -1 {
			cDir = s.StoragePath[:idx+len(".codex")]
		}
	}
	if cDir == "" {
		profileDir := ""
		if !s.IsHost && s.Profile != "" {
			profileDir = filepath.Join(config.BaseDir(), "profiles", s.Profile)
		}
		cDir = p.codexDir(profileDir, s.IsHost)
	}
	dbPath := filepath.Join(cDir, "state_5.sqlite")

	if p.sqliteBin != "" {
		query := fmt.Sprintf("SELECT cwd FROM threads WHERE id = '%s' LIMIT 1;\n", escapeSQL(s.ID))
		if out, err := session.RunBoundedSQLiteWithBin(ctx, p.sqliteBin, dbPath, query); err == nil {
			cwd := strings.TrimSpace(string(out))
			if cwd != "" {
				s.Cwd = cwd
				return cwd, nil
			}
		}
	}
	return "", nil
}

// ResolveSummary resolves structured summary information for an OpenAI Codex CLI session.
func (p *Provider) ResolveSummary(ctx context.Context, s *session.Session) (session.SessionSummary, error) {
	if s == nil {
		return session.SessionSummary{}, nil
	}

	cDir := ""
	if s.StoragePath != "" {
		if idx := strings.Index(s.StoragePath, ".codex"); idx != -1 {
			cDir = s.StoragePath[:idx+len(".codex")]
		}
	}
	if cDir == "" {
		profileDir := ""
		if !s.IsHost && s.Profile != "" {
			profileDir = filepath.Join(config.BaseDir(), "profiles", s.Profile)
		}
		cDir = p.codexDir(profileDir, s.IsHost)
	}
	historyDB := filepath.Join(cDir, "thread_history_1.sqlite")

	goal := s.Goal
	if goal == "" {
		goal = session.CleanPromptText(s.Title)
	}

	progress := s.Progress
	recent := s.Recent

	if p.sqliteBin != "" && historyDB != "" {
		if fi, err := os.Stat(historyDB); err == nil && fi.Size() > 0 {
			// Query turns count if not already populated
			if progress == "" {
				turnQuery := fmt.Sprintf("SELECT COUNT(*) FROM thread_turns WHERE thread_id = '%s';\n", escapeSQL(s.ID))
				if out, err := session.RunBoundedSQLiteWithBin(ctx, p.sqliteBin, historyDB, turnQuery); err == nil {
					cntStr := strings.TrimSpace(string(out))
					if cnt, err := strconv.Atoi(cntStr); err == nil && cnt > 0 {
						progress = fmt.Sprintf("%d turns", cnt)
					}
				}
			}

			// Query latest user message
			if recent == "" {
				itemQuery := fmt.Sprintf("SELECT item_json FROM thread_items WHERE thread_id = '%s' AND item_type = 'userMessage' ORDER BY rollout_ordinal DESC LIMIT 1;\n", escapeSQL(s.ID))
				if out, err := session.RunBoundedSQLiteWithBin(ctx, p.sqliteBin, historyDB, itemQuery); err == nil && len(out) > 0 {
					var item struct {
						Content []struct {
							Text string `json:"text"`
						} `json:"content"`
					}
					if err := json.Unmarshal(out, &item); err == nil && len(item.Content) > 0 {
						recent = session.CleanPromptText(item.Content[0].Text)
					}
				}
			}
		}
	}

	sum := session.SessionSummary{
		Goal:           goal,
		Progress:       progress,
		RecentActivity: recent,
		Raw:            s.Summary,
	}
	return sum, nil
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

	// 4. Hydrate child subagents and shell snapshots
	if srcCodexDir != "" {
		if err := p.hydrateChildSubagentSessions(ctx, srcCodexDir, targetCodexDir, targetSessionsDir, srcSession.ID); err != nil {
			logger.Debug("[session/codex] error copying child subagent sessions: %v", err)
		}
		if err := p.hydrateShellSnapshots(srcCodexDir, targetCodexDir); err != nil {
			logger.Debug("[session/codex] error copying shell snapshots: %v", err)
		}
	}

	// 5. Update session_index.jsonl
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
	visitedRollouts := map[string]bool{srcRolloutPath: true}
	currentRollout := srcRolloutPath

	for {
		parentID := extractParentThreadID(currentRollout)
		if parentID == "" {
			break
		}
		parentRollout := p.findBaseRolloutPath(ctx, srcCodexDir, parentID, currentRollout)
		if parentRollout == "" || visitedRollouts[parentRollout] {
			break
		}
		visitedRollouts[parentRollout] = true

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
				if err := p.copyThreadInStateDB(ctx, srcDB, targetDB, parentID, parentID, destParentRollout); err != nil {
					logger.Debug("[session/codex] failed to copy ancestor thread %s to %s: %v", parentID, targetDB, err)
				}
			}

			srcHistoryDB := filepath.Join(srcCodexDir, "thread_history_1.sqlite")
			targetHistoryDB := filepath.Join(targetCodexDir, "thread_history_1.sqlite")
			if _, err := os.Stat(srcHistoryDB); err == nil {
				if err := p.copyThreadHistoryDB(ctx, srcHistoryDB, targetHistoryDB, parentID, parentID, destParentRollout); err != nil {
					logger.Debug("[session/codex] failed to copy ancestor thread history %s to %s: %v", parentID, targetHistoryDB, err)
				}
			}
		}

		currentRollout = parentRollout
	}
	return nil
}

func (p *Provider) hydrateChildSubagentSessions(ctx context.Context, srcCodexDir, targetCodexDir, targetSessionsDir, parentID string) error {
	if p.sqliteBin == "" {
		return nil
	}
	srcDB := filepath.Join(srcCodexDir, "state_5.sqlite")
	targetDB := filepath.Join(targetCodexDir, "state_5.sqlite")
	if _, err := os.Stat(srcDB); err != nil {
		return nil
	}

	query := fmt.Sprintf("SELECT child_thread_id, COALESCE(status, '') FROM thread_spawn_edges WHERE parent_thread_id = '%s';\n", escapeSQL(parentID))
	cmd := exec.CommandContext(ctx, p.sqliteBin, "-list", "-separator", "|", srcDB)
	cmd.Stdin = strings.NewReader(query)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	srcSessionsDir := filepath.Join(srcCodexDir, "sessions")
	srcHistoryDB := filepath.Join(srcCodexDir, "thread_history_1.sqlite")
	targetHistoryDB := filepath.Join(targetCodexDir, "thread_history_1.sqlite")

	createEdgeTable := "CREATE TABLE IF NOT EXISTS thread_spawn_edges (parent_thread_id TEXT NOT NULL, child_thread_id TEXT NOT NULL PRIMARY KEY, status TEXT NOT NULL);\n"
	createCmd := exec.CommandContext(ctx, p.sqliteBin, targetDB)
	createCmd.Stdin = strings.NewReader(createEdgeTable)
	_ = createCmd.Run()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) == 0 {
			continue
		}
		childID := strings.TrimSpace(parts[0])
		if childID == "" {
			continue
		}
		childStatus := ""
		if len(parts) > 1 {
			childStatus = strings.TrimSpace(parts[1])
		}

		childRollout := p.findRolloutPath(ctx, srcCodexDir, childID)
		var destChildRollout string
		if childRollout != "" {
			var destChildRel string
			if rel, err := filepath.Rel(srcSessionsDir, childRollout); err == nil && !strings.HasPrefix(rel, "..") {
				destChildRel = rel
			}
			if destChildRel == "" {
				destChildRel = filepath.Base(childRollout)
			}
			destChildRollout = filepath.Join(targetSessionsDir, destChildRel)
			_ = os.MkdirAll(filepath.Dir(destChildRollout), 0755)
			_ = copyFile(childRollout, destChildRollout)
		}

		_ = p.copyThreadInStateDB(ctx, srcDB, targetDB, childID, childID, destChildRollout)
		if _, err := os.Stat(srcHistoryDB); err == nil {
			_ = p.copyThreadHistoryDB(ctx, srcHistoryDB, targetHistoryDB, childID, childID, destChildRollout)
		}

		// Insert edge into target thread_spawn_edges
		insertEdge := fmt.Sprintf("INSERT OR REPLACE INTO thread_spawn_edges (parent_thread_id, child_thread_id, status) VALUES ('%s', '%s', '%s');\n",
			escapeSQL(parentID), escapeSQL(childID), escapeSQL(childStatus))
		edgeCmd := exec.CommandContext(ctx, p.sqliteBin, targetDB)
		edgeCmd.Stdin = strings.NewReader(insertEdge)
		_ = edgeCmd.Run()
	}
	return nil
}

func (p *Provider) hydrateShellSnapshots(srcCodexDir, targetCodexDir string) error {
	srcShellDir := filepath.Join(srcCodexDir, "shell_snapshots")
	if _, err := os.Stat(srcShellDir); err != nil {
		return nil
	}
	targetShellDir := filepath.Join(targetCodexDir, "shell_snapshots")
	if err := os.MkdirAll(targetShellDir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(srcShellDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcFile := filepath.Join(srcShellDir, entry.Name())
		destFile := filepath.Join(targetShellDir, entry.Name())
		_ = copyFile(srcFile, destFile)
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
			if err := p.copyThreadHistoryDB(ctx, srcHistoryDB, targetHistoryDB, srcID, targetID, targetRolloutPath); err != nil {
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
