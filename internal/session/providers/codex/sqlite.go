package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/session"
)

type sqliteThreadRow struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Title            string          `json:"title"`
	FirstUserMessage string          `json:"first_user_message"`
	Preview          string          `json:"preview"`
	Cwd              string          `json:"cwd"`
	UpdatedAt        json.RawMessage `json:"updated_at"`
	RolloutPath      string          `json:"rollout_path"`
}

func (p *Provider) listFromSQLite(ctx context.Context, dbPath, profileName string, isHost bool) ([]session.Session, error) {
	cols, err := p.getTableColumns(ctx, dbPath, "threads")
	if err != nil || len(cols) == 0 {
		return nil, fmt.Errorf("failed to inspect threads table at %s: %w", dbPath, err)
	}

	colSet := make(map[string]bool, len(cols))
	for _, c := range cols {
		colSet[c] = true
	}

	var selectCols []string
	selectCols = append(selectCols, "id")
	if colSet["name"] {
		selectCols = append(selectCols, "COALESCE(name, '') AS name")
	} else {
		selectCols = append(selectCols, "'' AS name")
	}
	if colSet["title"] {
		selectCols = append(selectCols, "COALESCE(title, '') AS title")
	} else {
		selectCols = append(selectCols, "'' AS title")
	}
	if colSet["first_user_message"] {
		selectCols = append(selectCols, "COALESCE(first_user_message, '') AS first_user_message")
	} else {
		selectCols = append(selectCols, "'' AS first_user_message")
	}
	if colSet["preview"] {
		selectCols = append(selectCols, "COALESCE(preview, '') AS preview")
	} else {
		selectCols = append(selectCols, "'' AS preview")
	}
	if colSet["cwd"] {
		selectCols = append(selectCols, "COALESCE(cwd, '') AS cwd")
	} else {
		selectCols = append(selectCols, "'' AS cwd")
	}
	selectCols = append(selectCols, "updated_at", "rollout_path")

	var whereClauses []string
	if colSet["thread_source"] {
		whereClauses = append(whereClauses, "(thread_source IS NULL OR thread_source != 'subagent')")
	}
	if colSet["archived"] {
		whereClauses = append(whereClauses, "(archived IS NULL OR archived = 0)")
	}

	// Filter out empty zombie sessions (0 tokens, no title, no user message or preview)
	var contentConditions []string
	if colSet["tokens_used"] {
		contentConditions = append(contentConditions, "(tokens_used IS NOT NULL AND tokens_used > 0)")
	}
	if colSet["name"] {
		contentConditions = append(contentConditions, "(name IS NOT NULL AND name != '')")
	}
	if colSet["title"] {
		contentConditions = append(contentConditions, "(title IS NOT NULL AND title != '')")
	}
	if colSet["first_user_message"] {
		contentConditions = append(contentConditions, "(first_user_message IS NOT NULL AND first_user_message != '')")
	}
	if colSet["preview"] {
		contentConditions = append(contentConditions, "(preview IS NOT NULL AND preview != '')")
	}
	if len(contentConditions) > 0 {
		whereClauses = append(whereClauses, "("+strings.Join(contentConditions, " OR ")+")")
	}

	query := fmt.Sprintf("SELECT %s FROM threads", strings.Join(selectCols, ", "))
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY updated_at DESC LIMIT 100;\n"

	matches, err := p.queryThreadsSQLite(ctx, dbPath, query, profileName, isHost)
	if err != nil {
		return nil, err
	}
	var sessions []session.Session
	for _, m := range matches {
		sessions = append(sessions, *m)
	}
	return sessions, nil
}

func (p *Provider) getFromSQLite(ctx context.Context, dbPath, idOrPrefix, profileName string, isHost bool) (*session.Session, error) {
	if !isValidSessionQuery(idOrPrefix) {
		return nil, nil
	}

	cols, err := p.getTableColumns(ctx, dbPath, "threads")
	if err != nil || len(cols) == 0 {
		return nil, nil
	}
	colSet := make(map[string]bool, len(cols))
	for _, c := range cols {
		colSet[c] = true
	}

	var selectCols []string
	selectCols = append(selectCols, "id")
	if colSet["name"] {
		selectCols = append(selectCols, "COALESCE(name, '') AS name")
	} else {
		selectCols = append(selectCols, "'' AS name")
	}
	if colSet["title"] {
		selectCols = append(selectCols, "COALESCE(title, '') AS title")
	} else {
		selectCols = append(selectCols, "'' AS title")
	}
	if colSet["first_user_message"] {
		selectCols = append(selectCols, "COALESCE(first_user_message, '') AS first_user_message")
	} else {
		selectCols = append(selectCols, "'' AS first_user_message")
	}
	if colSet["preview"] {
		selectCols = append(selectCols, "COALESCE(preview, '') AS preview")
	} else {
		selectCols = append(selectCols, "'' AS preview")
	}
	if colSet["cwd"] {
		selectCols = append(selectCols, "COALESCE(cwd, '') AS cwd")
	} else {
		selectCols = append(selectCols, "'' AS cwd")
	}
	selectCols = append(selectCols, "updated_at", "rollout_path")

	var matchClauses []string
	matchClauses = append(matchClauses, fmt.Sprintf("id = '%s'", escapeSQL(idOrPrefix)))
	matchClauses = append(matchClauses, fmt.Sprintf("id LIKE '%s%%'", escapeSQL(idOrPrefix)))
	if colSet["name"] {
		matchClauses = append(matchClauses, fmt.Sprintf("name = '%s'", escapeSQL(idOrPrefix)))
		matchClauses = append(matchClauses, fmt.Sprintf("name LIKE '%s%%'", escapeSQL(idOrPrefix)))
	}
	if colSet["title"] {
		matchClauses = append(matchClauses, fmt.Sprintf("title = '%s'", escapeSQL(idOrPrefix)))
	}
	whereClause := "(" + strings.Join(matchClauses, " OR ") + ")"
	if colSet["thread_source"] {
		whereClause += " AND (thread_source IS NULL OR thread_source != 'subagent')"
	}
	if colSet["archived"] {
		whereClause += " AND (archived IS NULL OR archived = 0)"
	}

	query := fmt.Sprintf("SELECT %s FROM threads WHERE %s ORDER BY updated_at DESC LIMIT 5;\n",
		strings.Join(selectCols, ", "), whereClause)

	matches, err := p.queryThreadsSQLite(ctx, dbPath, query, profileName, isHost)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite db at %s: %w", dbPath, err)
	}

	if len(matches) > 1 {
		for _, m := range matches {
			if m.ID == idOrPrefix {
				return m, nil
			}
		}
		var ids []string
		for _, m := range matches {
			ids = append(ids, m.ShortID)
		}
		return nil, fmt.Errorf("ambiguous prefix %q matches multiple sessions in %s: %s", idOrPrefix, profileName, strings.Join(ids, ", "))
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	return nil, nil
}

func (p *Provider) queryThreadsSQLite(ctx context.Context, dbPath, query, profileName string, isHost bool) ([]*session.Session, error) {
	out, err := session.RunBoundedSQLiteWithBin(ctx, p.sqliteBin, dbPath, query, "-json")
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite db at %s: %w", dbPath, err)
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "[]" {
		return nil, nil
	}

	var rows []sqliteThreadRow
	if err := json.Unmarshal([]byte(trimmed), &rows); err != nil {
		return nil, fmt.Errorf("failed to parse json output from sqlite: %w", err)
	}

	historyDB := filepath.Join(filepath.Dir(dbPath), "thread_history_1.sqlite")
	hasHistory := false
	if fi, err := os.Stat(historyDB); err == nil && fi.Size() > 0 {
		hasHistory = true
	}

	type turnInfo struct {
		maxTime   time.Time
		turnCount int
	}
	turnsByThread := make(map[string]turnInfo)
	if hasHistory && len(rows) > 0 {
		var ids []string
		for _, r := range rows {
			if r.ID != "" {
				ids = append(ids, fmt.Sprintf("'%s'", escapeSQL(r.ID)))
			}
		}
		if len(ids) > 0 {
			turnQuery := fmt.Sprintf("SELECT thread_id, MAX(COALESCE(completed_at, started_at, 0)), COUNT(*) FROM thread_turns WHERE thread_id IN (%s) GROUP BY thread_id;\n", strings.Join(ids, ","))
			if tOut, err := session.RunBoundedSQLiteWithBin(ctx, p.sqliteBin, historyDB, turnQuery, "-separator", "|"); err == nil {
				for _, line := range strings.Split(strings.TrimSpace(string(tOut)), "\n") {
					parts := strings.Split(strings.TrimSpace(line), "|")
					if len(parts) >= 3 {
						tID := parts[0]
						var tTime time.Time
						if tSec, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil && tSec > 0 {
							tTime = time.Unix(tSec, 0)
						}
						cnt, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
						turnsByThread[tID] = turnInfo{maxTime: tTime, turnCount: cnt}
					}
				}
			}
		}
	}

	var sessions []*session.Session
	for _, row := range rows {
		displayTitle := row.Name
		if displayTitle == "" {
			displayTitle = row.Title
		}
		if displayTitle == "" && row.FirstUserMessage != "" {
			displayTitle = session.CleanPromptText(strings.Split(row.FirstUserMessage, "\n")[0])
		}
		if displayTitle == "" && row.Preview != "" {
			displayTitle = session.CleanPromptText(strings.Split(row.Preview, "\n")[0])
		}
		if displayTitle == "" {
			displayTitle = "Untitled Session"
		}

		cleanFirst := session.CleanPromptText(row.FirstUserMessage)
		if cleanFirst == "" {
			cleanFirst = session.CleanPromptText(row.Preview)
		}

		var sec int64
		if len(row.UpdatedAt) > 0 {
			raw := strings.Trim(string(row.UpdatedAt), `"`)
			sec, _ = strconv.ParseInt(raw, 10, 64)
		}
		modTime := time.Unix(sec, 0)

		if row.RolloutPath != "" {
			if fi, err := os.Stat(row.RolloutPath); err == nil && fi.ModTime().After(modTime) {
				modTime = fi.ModTime()
			}
		}

		var turnCount int
		if ti, ok := turnsByThread[row.ID]; ok {
			if ti.maxTime.After(modTime) {
				modTime = ti.maxTime
			}
			turnCount = ti.turnCount
		}
		if turnCount == 0 && row.RolloutPath != "" {
			if f, err := os.Open(row.RolloutPath); err == nil {
				if tc, err := session.ReadTurnCounts(f); err == nil && tc > 0 {
					turnCount = tc
				}
				_ = f.Close()
			}
		}

		s := session.NewSession(row.ID, displayTitle, "codex", profileName, isHost, modTime)
		s.Goal = session.CleanPromptText(displayTitle)
		if s.Goal == "Untitled Session" && cleanFirst != "" {
			s.Goal = cleanFirst
		}
		if turnCount > 0 {
			s.Progress = fmt.Sprintf("%d turns", turnCount)
		}
		s.StoragePath = row.RolloutPath
		s.Cwd = row.Cwd

		sum := session.SessionSummary{
			Goal:     s.Goal,
			Progress: s.Progress,
			Raw:      row.FirstUserMessage,
		}
		s.Summary = sum.Text()
		sessions = append(sessions, &s)
	}

	return sessions, nil
}

func (p *Provider) getTableColumns(ctx context.Context, dbPath, table string) ([]string, error) {
	if !isValidIdentifier(table) {
		return nil, fmt.Errorf("invalid table name %q", table)
	}
	query := fmt.Sprintf("PRAGMA table_info(%s);\n", table)
	out, err := session.RunBoundedSQLiteWithBin(ctx, p.sqliteBin, dbPath, query, "-separator", "|")
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
