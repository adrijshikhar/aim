package codex

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/logger"
)

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

func (p *Provider) copyThreadHistoryDB(ctx context.Context, srcHistoryDB, destHistoryDB, srcID, targetID, targetRolloutPath string) error {
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

		part := fmt.Sprintf("DELETE FROM %s WHERE thread_id = '%s';\nINSERT OR REPLACE INTO %s (%s)\nSELECT %s\nFROM src.%s WHERE thread_id = '%s';",
			tbl, escapeSQL(targetID), tbl, strings.Join(insertCols, ", "), strings.Join(selectExprs, ", "), tbl, escapeSQL(srcID))
		sqlParts = append(sqlParts, part)
	}

	if len(sqlParts) == 0 {
		return nil
	}

	var rolloutSize int64 = -1
	if targetRolloutPath != "" {
		if fi, err := os.Stat(targetRolloutPath); err == nil {
			rolloutSize = fi.Size()
		}
	}

	if rolloutSize >= 0 {
		// If next_rollout_bytes_offset points beyond the actual rollout file,
		// the projection state is invalid. Purge it so Codex re-projects cleanly from offset 0.
		purgeSQL := fmt.Sprintf(`
DELETE FROM thread_items WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s' AND next_rollout_bytes_offset > %[2]d);
DELETE FROM thread_turns WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s' AND next_rollout_bytes_offset > %[2]d);
DELETE FROM thread_realtime_items WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s' AND next_rollout_bytes_offset > %[2]d);
DELETE FROM thread_history_projection_state WHERE thread_id = '%[1]s' AND next_rollout_bytes_offset > %[2]d;`,
			escapeSQL(targetID), rolloutSize)
		sqlParts = append(sqlParts, purgeSQL)
	}

	// Always prune any stale items or turns with rollout_ordinal >= next_rollout_ordinal
	// to prevent SQLite UNIQUE constraint failures during Codex thread resumption.
	pruneSQL := fmt.Sprintf(`
DELETE FROM thread_items WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s') AND rollout_ordinal >= (SELECT next_rollout_ordinal FROM thread_history_projection_state WHERE thread_id = '%[1]s');
DELETE FROM thread_turns WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s') AND rollout_ordinal >= (SELECT next_rollout_ordinal FROM thread_history_projection_state WHERE thread_id = '%[1]s');`,
		escapeSQL(targetID))
	sqlParts = append(sqlParts, pruneSQL)

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

func copyFile(src, dst string) (err error) {
	return copyFileWithReplace(src, dst, "", "")
}
