package claude

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// Provider implements session.SessionProvider for Claude Code (claude).
type Provider struct{}

// NewProvider constructs a new Claude Code session provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Agent returns the identifier for Claude Code ("claude").
func (p *Provider) Agent() string {
	return "claude"
}

// ListSessions discovers sessions across project directories under .claude/projects/<slug>/*.jsonl.
func (p *Provider) ListSessions(ctx context.Context, profileDir string, isHost bool) ([]session.Session, error) {
	var projectsDir string
	if isHost {
		projectsDir = filepath.Join(config.RealHomeDir(), ".claude", "projects")
	} else {
		projectsDir = filepath.Join(profileDir, ".claude", "projects")
	}

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read claude projects directory %s: %w", projectsDir, err)
	}

	var results []session.Session
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		projPath := filepath.Join(projectsDir, slug)
		files, err := os.ReadDir(projPath)
		if err != nil {
			logger.Debug("[session/claude] Error reading project dir %s: %v", projPath, err)
			continue
		}

		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			sessID := strings.TrimSuffix(f.Name(), ".jsonl")
			if sessID == "" {
				continue
			}
			filePath := filepath.Join(projPath, f.Name())
			sess, err := parseClaudeSessionFile(filePath, sessID, slug, profileDir, isHost)
			if err != nil {
				logger.Debug("[session/claude] Error parsing session file %s: %v", filePath, err)
				continue
			}
			if sess != nil {
				results = append(results, *sess)
			}
		}
	}

	return results, nil
}

// GetSession resolves a full ID or short ID prefix to a concrete Session.
func (p *Provider) GetSession(ctx context.Context, idOrPrefix string, profileDir string, isHost bool) (*session.Session, error) {
	if idOrPrefix == "" {
		return nil, nil
	}

	sessions, err := p.ListSessions(ctx, profileDir, isHost)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions for claude: %w", err)
	}

	var matches []session.Session
	for _, s := range sessions {
		if s.ID == idOrPrefix || strings.HasPrefix(s.ID, idOrPrefix) || strings.HasPrefix(s.ShortID, idOrPrefix) {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		return nil, nil
	}

	if len(matches) > 1 {
		for _, s := range matches {
			if s.ID == idOrPrefix {
				match := s
				return &match, nil
			}
		}
		return nil, fmt.Errorf("ambiguous session ID prefix %q matches %d sessions", idOrPrefix, len(matches))
	}

	return &matches[0], nil
}

// Hydrate copies a conversation session to destProfileDir, either verbatim (fork=false)
// or with a generated UUID and all session references replaced (fork=true).
func (p *Provider) Hydrate(ctx context.Context, srcSession *session.Session, destProfileDir string, fork bool) (string, error) {
	if srcSession == nil {
		return "", fmt.Errorf("source session is nil")
	}

	slug := ""
	if srcSession.StoragePath != "" {
		slug = filepath.Base(filepath.Dir(srcSession.StoragePath))
	}
	if slug == "" || slug == "." || slug == string(filepath.Separator) {
		slug = PathToSlug(srcSession.Cwd)
	}
	if slug == "" || slug == "-" {
		slug = "-default"
	}

	srcFile := ""
	if srcSession.StoragePath != "" {
		if fi, err := os.Stat(srcSession.StoragePath); err == nil && !fi.IsDir() {
			srcFile = srcSession.StoragePath
		}
	}

	if srcFile == "" {
		var baseProjectsDir string
		if srcSession.IsHost {
			baseProjectsDir = filepath.Join(config.RealHomeDir(), ".claude", "projects")
		} else if srcSession.Profile != "" {
			baseProjectsDir = filepath.Join(config.BaseDir(), "profiles", srcSession.Profile, ".claude", "projects")
		}

		if baseProjectsDir != "" {
			cand := filepath.Join(baseProjectsDir, slug, srcSession.ID+".jsonl")
			if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
				srcFile = cand
			} else {
				// Search across all slug folders in baseProjectsDir
				if entries, err := os.ReadDir(baseProjectsDir); err == nil {
					for _, entry := range entries {
						if entry.IsDir() {
							c := filepath.Join(baseProjectsDir, entry.Name(), srcSession.ID+".jsonl")
							if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
								srcFile = c
								slug = entry.Name()
								break
							}
						}
					}
				}
			}
		}
	}

	if srcFile == "" {
		return "", fmt.Errorf("source session file for %s not found: %w", srcSession.ID, os.ErrNotExist)
	}

	destDir := filepath.Join(destProfileDir, ".claude", "projects", slug)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination project directory %s: %w", destDir, err)
	}

	destID := srcSession.ID
	if fork {
		newUUID, err := session.GenerateUUID()
		if err != nil {
			return "", fmt.Errorf("failed to generate fork UUID: %w", err)
		}
		destID = newUUID
	}

	destFile := filepath.Join(destDir, destID+".jsonl")
	if !fork && filepath.Clean(srcFile) == filepath.Clean(destFile) {
		return destID, nil
	}
	if fork {
		if err := copyJSONLWithReplace(srcFile, destFile, srcSession.ID, destID); err != nil {
			return "", fmt.Errorf("failed to copy forked session to %s: %w", destFile, err)
		}
	} else {
		if err := copyFile(srcFile, destFile); err != nil {
			return "", fmt.Errorf("failed to copy session file to %s: %w", destFile, err)
		}
	}

	return destID, nil
}

type claudeLineEvent struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId"`
	Timestamp string          `json:"timestamp"`
	CreatedAt string          `json:"created_at"`
	Message   json.RawMessage `json:"message"`
	Content   json.RawMessage `json:"content"`
}

type claudeMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func parseClaudeSessionFile(filePath, sessionUUID, slug, profileDir string, isHost bool) (*session.Session, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file %s: %w", filePath, err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat session file %s: %w", filePath, err)
	}

	var firstUserPrompt string
	var latestUserPrompt string
	var title string
	var earliestTime time.Time
	var latestTime time.Time
	messageCount := 0

	const headSize = 64 * 1024
	fileSize := fi.Size()

	if fileSize <= headSize {
		// Small file: full scan
		scanner := bufio.NewScanner(file)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 10*1024*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}

			var ev claudeLineEvent
			if err := json.Unmarshal(line, &ev); err != nil {
				continue
			}

			rawTS := ev.Timestamp
			if rawTS == "" {
				rawTS = ev.CreatedAt
			}
			if rawTS != "" {
				if t, ok := parseTimestamp(rawTS); ok {
					if earliestTime.IsZero() || t.Before(earliestTime) {
						earliestTime = t
					}
					if latestTime.IsZero() || t.After(latestTime) {
						latestTime = t
					}
				}
			}

			if ev.Type == "user" || isUserMessage(ev.Message) {
				promptText := extractPromptText(ev)
				if promptText != "" {
					cleaned := session.CleanPromptText(promptText)
					if cleaned != "" {
						if firstUserPrompt == "" {
							firstUserPrompt = cleaned
						}
						latestUserPrompt = cleaned
					}
				}
			}

			if ev.Type == "user" || ev.Type == "assistant" || isUserOrAssistantMessage(ev.Message) {
				messageCount++
			}
		}
	} else {
		// Large file: fast O(1) head + tail seeking
		// 1. Read Head (first 64KB) for earliest time and first user prompt
		headBuf := make([]byte, headSize)
		if n, err := file.Read(headBuf); err == nil && n > 0 {
			lines := strings.Split(string(headBuf[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var ev claudeLineEvent
				if err := json.Unmarshal([]byte(line), &ev); err == nil {
					rawTS := ev.Timestamp
					if rawTS == "" {
						rawTS = ev.CreatedAt
					}
					if rawTS != "" {
						if t, ok := parseTimestamp(rawTS); ok && (earliestTime.IsZero() || t.Before(earliestTime)) {
							earliestTime = t
						}
					}
					if firstUserPrompt == "" && (ev.Type == "user" || isUserMessage(ev.Message)) {
						if pt := extractPromptText(ev); pt != "" {
							firstUserPrompt = session.CleanPromptText(pt)
						}
					}
				}
				if firstUserPrompt != "" && !earliestTime.IsZero() {
					break
				}
			}
		}

		// 2. Read Tail (last 64KB) for latest time and latest user prompt
		tailBuf := make([]byte, headSize)
		offset := fileSize - headSize
		if _, err := file.ReadAt(tailBuf, offset); err == nil {
			lines := strings.Split(string(tailBuf), "\n")
			if len(lines) > 0 {
				lines = lines[1:] // Discard partial first line
			}
			for i := len(lines) - 1; i >= 0; i-- {
				line := strings.TrimSpace(lines[i])
				if line == "" {
					continue
				}
				var ev claudeLineEvent
				if err := json.Unmarshal([]byte(line), &ev); err == nil {
					rawTS := ev.Timestamp
					if rawTS == "" {
						rawTS = ev.CreatedAt
					}
					if rawTS != "" && latestTime.IsZero() {
						if t, ok := parseTimestamp(rawTS); ok {
							latestTime = t
						}
					}
					if latestUserPrompt == "" && (ev.Type == "user" || isUserMessage(ev.Message)) {
						if pt := extractPromptText(ev); pt != "" {
							latestUserPrompt = session.CleanPromptText(pt)
						}
					}
				}
				if latestUserPrompt != "" && !latestTime.IsZero() {
					break
				}
			}
		}

		// 3. Fast count of user/assistant messages without full JSON parsing
		fastBuf := make([]byte, 256*1024)
		_, _ = file.Seek(0, io.SeekStart)
		for {
			n, err := file.Read(fastBuf)
			if n > 0 {
				chunk := fastBuf[:n]
				messageCount += bytes.Count(chunk, []byte(`"type":"user"`))
				messageCount += bytes.Count(chunk, []byte(`"type":"assistant"`))
			}
			if err != nil {
				break
			}
		}
	}

	if title == "" {
		if firstUserPrompt != "" {
			title = firstUserPrompt
		} else {
			title = "Untitled Session"
		}
	}

	createdAt := fi.ModTime()
	if !earliestTime.IsZero() {
		createdAt = earliestTime
	}

	lastActive := fi.ModTime()
	if !latestTime.IsZero() {
		lastActive = latestTime
	}

	profileName := filepath.Base(profileDir)
	if isHost {
		profileName = "<host>"
	}

	s := session.NewSession(sessionUUID, title, "claude", profileName, isHost, lastActive)
	s.StartedAt = createdAt
	s.CreatedAt = createdAt
	s.LastActiveAt = lastActive
	s.StoragePath = filePath
	s.Cwd = SlugToPath(slug)
	s.MessageCount = messageCount
	s.Goal = firstUserPrompt
	if messageCount > 0 {
		s.Progress = fmt.Sprintf("%d messages", messageCount)
	}
	s.Recent = latestUserPrompt

	sum := session.SessionSummary{
		Goal:           s.Goal,
		Progress:       s.Progress,
		RecentActivity: s.Recent,
		Raw:            title,
	}
	s.Summary = sum.Text()

	return &s, nil
}

// ResolveCwd resolves the working directory for a Claude Code conversation session.
func (p *Provider) ResolveCwd(ctx context.Context, s *session.Session) (string, error) {
	if s == nil {
		return "", nil
	}
	if s.Cwd != "" {
		return s.Cwd, nil
	}
	if s.StoragePath != "" {
		slug := filepath.Base(filepath.Dir(s.StoragePath))
		cwd := SlugToPath(slug)
		if cwd != "" {
			s.Cwd = cwd
			return cwd, nil
		}
	}
	return "", nil
}

// ResolveSummary resolves structured summary information for a Claude Code conversation session.
func (p *Provider) ResolveSummary(ctx context.Context, s *session.Session) (session.SessionSummary, error) {
	if s == nil {
		return session.SessionSummary{}, nil
	}
	if s.Goal != "" || s.Recent != "" {
		return session.SessionSummary{
			Goal:           s.Goal,
			Progress:       s.Progress,
			RecentActivity: s.Recent,
			Raw:            s.Summary,
		}, nil
	}
	if s.StoragePath != "" {
		slug := filepath.Base(filepath.Dir(s.StoragePath))
		if parsed, err := parseClaudeSessionFile(s.StoragePath, s.ID, slug, s.Profile, s.IsHost); err == nil && parsed != nil {
			return session.SessionSummary{
				Goal:           parsed.Goal,
				Progress:       parsed.Progress,
				RecentActivity: parsed.Recent,
				Raw:            parsed.Summary,
			}, nil
		}
	}
	return session.SessionSummary{
		Goal: s.Title,
		Raw:  s.Summary,
	}, nil
}

func isUserMessage(rawMsg json.RawMessage) bool {
	if len(rawMsg) == 0 {
		return false
	}
	if !bytes.Contains(rawMsg, []byte(`"user"`)) {
		return false
	}
	var msg claudeMessage
	if err := json.Unmarshal(rawMsg, &msg); err == nil {
		return msg.Role == "user"
	}
	return false
}

func isUserOrAssistantMessage(rawMsg json.RawMessage) bool {
	if len(rawMsg) == 0 {
		return false
	}
	if !bytes.Contains(rawMsg, []byte(`"user"`)) && !bytes.Contains(rawMsg, []byte(`"assistant"`)) {
		return false
	}
	var msg claudeMessage
	if err := json.Unmarshal(rawMsg, &msg); err == nil {
		return msg.Role == "user" || msg.Role == "assistant"
	}
	return false
}

func extractPromptText(ev claudeLineEvent) string {
	// 1. Try ev.Message
	if len(ev.Message) > 0 {
		var text string
		if err := json.Unmarshal(ev.Message, &text); err == nil && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}

		var msg claudeMessage
		if err := json.Unmarshal(ev.Message, &msg); err == nil {
			if txt := extractTextFromRaw(msg.Content); txt != "" {
				return txt
			}
		}
	}

	// 2. Try ev.Content
	if len(ev.Content) > 0 {
		if txt := extractTextFromRaw(ev.Content); txt != "" {
			return txt
		}
	}

	return ""
}

func extractTextFromRaw(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try plain string
	var str string
	if err := json.Unmarshal(raw, &str); err == nil && strings.TrimSpace(str) != "" {
		return strings.TrimSpace(str)
	}

	// Try array of content blocks
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		for _, b := range blocks {
			if strings.TrimSpace(b.Text) != "" {
				return strings.TrimSpace(b.Text)
			}
		}
	}

	// Try generic array of maps
	var genericBlocks []map[string]interface{}
	if err := json.Unmarshal(raw, &genericBlocks); err == nil {
		for _, b := range genericBlocks {
			if t, ok := b["text"].(string); ok && strings.TrimSpace(t) != "" {
				return strings.TrimSpace(t)
			}
		}
	}

	return ""
}

func parseTimestamp(ts string) (time.Time, bool) {
	if ts == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source %s: %w", src, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for destination %s: %w", dst, err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination %s: %w", dst, err)
	}
	defer func() {
		if cErr := out.Close(); cErr != nil && err == nil {
			err = fmt.Errorf("failed to close destination %s: %w", dst, cErr)
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy data from %s to %s: %w", src, dst, err)
	}
	return nil
}

func copyJSONLWithReplace(src, dst, oldUUID, newUUID string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source %s: %w", src, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for destination %s: %w", dst, err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination %s: %w", dst, err)
	}
	defer func() {
		if cErr := out.Close(); cErr != nil && err == nil {
			err = fmt.Errorf("failed to close destination %s: %w", dst, cErr)
		}
	}()

	oldBytes := []byte(oldUUID)
	newBytes := []byte(newUUID)

	reader := bufio.NewReaderSize(in, 64*1024)
	writer := bufio.NewWriterSize(out, 64*1024)

	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			replaced := line
			if bytes.Contains(line, oldBytes) {
				replaced = bytes.ReplaceAll(line, oldBytes, newBytes)
			}
			if _, wErr := writer.Write(replaced); wErr != nil {
				return fmt.Errorf("failed to write to %s: %w", dst, wErr)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("failed to read from %s: %w", src, readErr)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush destination %s: %w", dst, err)
	}

	return nil
}
