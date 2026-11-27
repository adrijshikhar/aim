package codex

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/session"
)

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

	matches, err := p.queryThreadsSQLite(ctx, dbPath, query, nil, profileName, isHost)
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

	likePattern := escapeSQLLike(idOrPrefix) + "%"
	var matchClauses []string
	var args []any
	matchClauses = append(matchClauses, "id = ?", `id LIKE ? ESCAPE '\'`)
	args = append(args, idOrPrefix, likePattern)
	if colSet["name"] {
		matchClauses = append(matchClauses, "name = ?", `name LIKE ? ESCAPE '\'`)
		args = append(args, idOrPrefix, likePattern)
	}
	if colSet["title"] {
		matchClauses = append(matchClauses, "title = ?")
		args = append(args, idOrPrefix)
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

	matches, err := p.queryThreadsSQLite(ctx, dbPath, query, args, profileName, isHost)
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

func escapeSQLLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

type codexThreadRecord struct {
	ID               string
	Name             string
	Title            string
	FirstUserMessage string
	Preview          string
	Cwd              string
	UpdatedSec       int64
	RolloutPath      string
}

type codexTurnInfo struct {
	maxTime   time.Time
	turnCount int
}

func parseUnixTimestampAny(raw any) int64 {
	switch v := raw.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case string:
		sec, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return sec
	case []byte:
		sec, _ := strconv.ParseInt(strings.TrimSpace(string(v)), 10, 64)
		return sec
	default:
		return 0
	}
}

func (p *Provider) scanCodexThreadRows(ctx context.Context, dbPath, query string, args []any) ([]codexThreadRecord, error) {
	db, err := session.OpenReadOnlySQLite(ctx, dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db %s: %w", dbPath, err)
	}
	defer func() { _ = db.Close() }()

	sqlRows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite db at %s: %w", dbPath, err)
	}
	defer func() { _ = sqlRows.Close() }()

	var recs []codexThreadRecord
	for sqlRows.Next() {
		var (
			id, name, title, firstMsg, preview, cwd, rollout sql.NullString
			updatedRaw                                       any
		)
		if err := sqlRows.Scan(&id, &name, &title, &firstMsg, &preview, &cwd, &updatedRaw, &rollout); err != nil {
			return nil, fmt.Errorf("scan sqlite row: %w", err)
		}
		recs = append(recs, codexThreadRecord{
			ID:               id.String,
			Name:             name.String,
			Title:            title.String,
			FirstUserMessage: firstMsg.String,
			Preview:          preview.String,
			Cwd:              cwd.String,
			UpdatedSec:       parseUnixTimestampAny(updatedRaw),
			RolloutPath:      rollout.String,
		})
	}
	if err := sqlRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sqlite rows: %w", err)
	}
	return recs, nil
}

func loadCodexTurnInfo(ctx context.Context, dbPath string, rows []codexThreadRecord) map[string]codexTurnInfo {
	turnsByThread := make(map[string]codexTurnInfo)
	historyDB := filepath.Join(filepath.Dir(dbPath), "thread_history_1.sqlite")
	fi, err := os.Stat(historyDB)
	if err != nil || fi.Size() == 0 || len(rows) == 0 {
		return turnsByThread
	}
	db, err := session.OpenReadOnlySQLite(ctx, historyDB)
	if err != nil {
		return turnsByThread
	}
	defer func() { _ = db.Close() }()

	stmt, err := db.PrepareContext(ctx, "SELECT MAX(COALESCE(completed_at, started_at, 0)), COUNT(*) FROM thread_turns WHERE thread_id = ?")
	if err != nil {
		return turnsByThread
	}
	defer func() { _ = stmt.Close() }()

	for _, r := range rows {
		if r.ID == "" {
			continue
		}
		var (
			maxRaw any
			cnt    int
		)
		if err := stmt.QueryRowContext(ctx, r.ID).Scan(&maxRaw, &cnt); err != nil || cnt == 0 {
			continue
		}
		var tTime time.Time
		if tSec := parseUnixTimestampAny(maxRaw); tSec > 0 {
			tTime = time.Unix(tSec, 0)
		}
		turnsByThread[r.ID] = codexTurnInfo{maxTime: tTime, turnCount: cnt}
	}
	return turnsByThread
}

func (p *Provider) queryThreadsSQLite(ctx context.Context, dbPath, query string, args []any, profileName string, isHost bool) ([]*session.Session, error) {
	rows, err := p.scanCodexThreadRows(ctx, dbPath, query, args)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	turnsByThread := loadCodexTurnInfo(ctx, dbPath, rows)

	sessions := make([]*session.Session, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, buildCodexSQLiteSession(row, turnsByThread[row.ID], profileName, isHost))
	}
	return sessions, nil
}

func buildCodexSQLiteSession(row codexThreadRecord, ti codexTurnInfo, profileName string, isHost bool) *session.Session {
	displayTitle := pickCodexDisplayTitle(row)
	cleanFirst := session.CleanPromptText(row.FirstUserMessage)
	if cleanFirst == "" {
		cleanFirst = session.CleanPromptText(row.Preview)
	}
	modTime, turnCount := resolveCodexTimingAndTurns(row, ti)

	s := session.NewSession(row.ID, displayTitle, "codex", profileName, isHost, modTime)
	s.Goal = session.CleanPromptText(displayTitle)
	if s.Goal == "Untitled Session" && cleanFirst != "" {
		s.Goal = cleanFirst
	}
	if turnCount > 0 {
		s.Progress = fmt.Sprintf("%d turns", turnCount)
		s.MessageCount = turnCount
	}
	s.StoragePath = row.RolloutPath
	s.Cwd = row.Cwd

	sum := session.SessionSummary{
		Goal:     s.Goal,
		Progress: s.Progress,
		Raw:      row.FirstUserMessage,
	}
	s.Summary = sum.Text()
	return &s
}

func resolveCodexTimingAndTurns(row codexThreadRecord, ti codexTurnInfo) (time.Time, int) {
	modTime := time.Unix(row.UpdatedSec, 0)
	if row.RolloutPath != "" {
		if fi, err := os.Stat(row.RolloutPath); err == nil && fi.ModTime().After(modTime) {
			modTime = fi.ModTime()
		}
	}
	turnCount := ti.turnCount
	if ti.maxTime.After(modTime) {
		modTime = ti.maxTime
	}
	if turnCount == 0 && row.RolloutPath != "" {
		if f, err := os.Open(row.RolloutPath); err == nil {
			if tc, err := session.ReadTurnCounts(f); err == nil && tc > 0 {
				turnCount = tc
			}
			_ = f.Close()
		}
	}
	return modTime, turnCount
}

func pickCodexDisplayTitle(row codexThreadRecord) string {
	if row.Name != "" {
		return row.Name
	}
	if row.Title != "" {
		return row.Title
	}
	if row.FirstUserMessage != "" {
		if t := session.CleanPromptText(strings.Split(row.FirstUserMessage, "\n")[0]); t != "" {
			return t
		}
	}
	if row.Preview != "" {
		if t := session.CleanPromptText(strings.Split(row.Preview, "\n")[0]); t != "" {
			return t
		}
	}
	return "Untitled Session"
}

func (p *Provider) getTableColumns(ctx context.Context, dbPath, table string) ([]string, error) {
	if !isValidIdentifier(table) {
		return nil, fmt.Errorf("invalid table name %q", table)
	}
	db, err := session.OpenReadOnlySQLite(ctx, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db %s: %w", dbPath, err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s);", table))
	if err != nil {
		return nil, fmt.Errorf("failed to query table info for %s in %s: %w", table, dbPath, err)
	}
	defer func() { _ = rows.Close() }()

	var cols []string
	for rows.Next() {
		var (
			cid     int
			name    string
			typ     string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			continue
		}
		colName := strings.TrimSpace(name)
		if colName != "" && isValidIdentifier(colName) {
			cols = append(cols, colName)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate table info for %s: %w", table, err)
	}
	return cols, nil
}
