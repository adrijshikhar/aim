package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/session"
)

type rolloutMetaHeader struct {
	Type    string `json:"type"`
	Payload struct {
		ID          string `json:"id"`
		HistoryBase *struct {
			ThreadID string `json:"thread_id"`
		} `json:"history_base"`
	} `json:"payload"`
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
		if s.StoragePath != "" {
			if f, err := os.Open(s.StoragePath); err == nil {
				if tc, err := session.ReadTurnCounts(f); err == nil && tc > 0 {
					s.Progress = fmt.Sprintf("%d turns", tc)
				}
				_ = f.Close()
			}
		}
		sessions = append(sessions, s)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan session index at %s: %w", indexPath, err)
	}

	return sessions, nil
}

func (p *Provider) getFromJSONL(indexPath, idOrPrefix, profileName string, isHost bool) (*session.Session, error) {
	sessions, err := p.listFromJSONL(indexPath, profileName, isHost)
	if err != nil {
		return nil, err
	}
	var matches []*session.Session
	for _, s := range sessions {
		if strings.HasPrefix(s.ID, idOrPrefix) {
			match := s
			matches = append(matches, &match)
		}
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

func (p *Provider) findRolloutPath(ctx context.Context, codexDir, sessionID string) string {
	if p.sqliteBin != "" {
		dbPath := filepath.Join(codexDir, "state_5.sqlite")
		if _, err := os.Stat(dbPath); err == nil {
			query := fmt.Sprintf("SELECT rollout_path FROM threads WHERE id = '%s' LIMIT 1;\n", escapeSQL(sessionID))
			if out, err := session.RunBoundedSQLiteWithBin(ctx, p.sqliteBin, dbPath, query); err == nil {
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

func (p *Provider) findBaseRolloutPath(ctx context.Context, codexDir, parentID, currentRolloutPath string) string {
	cand := p.findRolloutPath(ctx, codexDir, parentID)
	if cand != "" && cand != currentRolloutPath {
		return cand
	}
	// Fallback: search sessions dir for another rollout containing parentID
	sessionsDir := filepath.Join(codexDir, "sessions")
	var fallbackMatch string
	_ = filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if path != currentRolloutPath && strings.HasSuffix(info.Name(), ".jsonl") && strings.Contains(info.Name(), parentID) {
			fallbackMatch = path
			if strings.HasSuffix(info.Name(), "-"+parentID+".jsonl") {
				return filepath.SkipAll
			}
		}
		return nil
	})
	return fallbackMatch
}

func extractParentThreadID(rolloutPath string) string {
	f, err := os.Open(rolloutPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for i := 0; i < 5; i++ {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			var meta rolloutMetaHeader
			if jsonErr := json.Unmarshal(line, &meta); jsonErr == nil {
				if meta.Payload.HistoryBase != nil && isValidSessionID(meta.Payload.HistoryBase.ThreadID) {
					return meta.Payload.HistoryBase.ThreadID
				}
			}
		}
		if err != nil {
			break
		}
	}
	return ""
}
