package codex

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/logger"
)

// SanitizeSession cleans up stale or desynchronized projection state in Codex's thread_history_1.sqlite.
// It ensures that no thread items or turns exist with rollout_ordinal >= next_rollout_ordinal,
// and resets broken projections where the rollout file is smaller than next_rollout_bytes_offset.
func (p *Provider) SanitizeSession(ctx context.Context, sessionID, profileDir string) error {
	if p.sqliteBin == "" || sessionID == "" || !isValidSessionID(sessionID) {
		return nil
	}
	historyDB := filepath.Join(p.codexDir(profileDir, false), "thread_history_1.sqlite")
	if _, err := os.Stat(historyDB); err != nil {
		return nil
	}

	rolloutPath := p.findRolloutPath(ctx, p.codexDir(profileDir, false), sessionID)
	var rolloutSize int64 = -1
	if rolloutPath != "" {
		if fi, err := os.Stat(rolloutPath); err == nil {
			rolloutSize = fi.Size()
		}
	}

	var sqlParts []string
	if rolloutSize >= 0 {
		purgeSQL := fmt.Sprintf(`
DELETE FROM thread_items WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s' AND next_rollout_bytes_offset > %[2]d);
DELETE FROM thread_turns WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s' AND next_rollout_bytes_offset > %[2]d);
DELETE FROM thread_realtime_items WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s' AND next_rollout_bytes_offset > %[2]d);
DELETE FROM thread_history_projection_state WHERE thread_id = '%[1]s' AND next_rollout_bytes_offset > %[2]d;`,
			escapeSQL(sessionID), rolloutSize)
		sqlParts = append(sqlParts, purgeSQL)
	}

	pruneSQL := fmt.Sprintf(`
DELETE FROM thread_items WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s') AND rollout_ordinal >= (SELECT next_rollout_ordinal FROM thread_history_projection_state WHERE thread_id = '%[1]s');
DELETE FROM thread_turns WHERE thread_id = '%[1]s' AND EXISTS (SELECT 1 FROM thread_history_projection_state WHERE thread_id = '%[1]s') AND rollout_ordinal >= (SELECT next_rollout_ordinal FROM thread_history_projection_state WHERE thread_id = '%[1]s');`,
		escapeSQL(sessionID))
	sqlParts = append(sqlParts, pruneSQL)

	fullSQL := fmt.Sprintf("BEGIN TRANSACTION;\n%s\nCOMMIT;\n", strings.Join(sqlParts, "\n"))
	cmd := exec.CommandContext(ctx, p.sqliteBin, historyDB)
	cmd.Stdin = strings.NewReader(fullSQL)
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Debug("[session/codex] SanitizeSession error for %s: %s: %v", sessionID, string(out), err)
		return err
	}
	return nil
}
