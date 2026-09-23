package claude

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
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
		newUUID, err := generateUUID()
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

	var title string
	var earliestTime time.Time
	var latestTime time.Time
	messageCount := 0

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
			logger.Debug("[session/claude] skipping malformed JSON line in %s: %v", filePath, err)
			continue
		}

		// Timestamp tracking
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

		// User prompt extraction for title
		if title == "" && (ev.Type == "user" || isUserMessage(ev.Message)) {
			promptText := extractPromptText(ev)
			if promptText != "" {
				title = promptText
			}
		}

		// Message count tracking
		if ev.Type == "user" || ev.Type == "assistant" || isUserOrAssistantMessage(ev.Message) {
			messageCount++
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Debug("[session/claude] scanner error on %s: %v", filePath, err)
	}

	if title == "" {
		title = "Untitled Session"
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
	s.Summary = title
	s.StoragePath = filePath
	s.Cwd = SlugToPath(slug)
	s.MessageCount = messageCount

	return &s, nil
}

func isUserMessage(rawMsg json.RawMessage) bool {
	if len(rawMsg) == 0 {
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

func generateUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("failed to read random bytes for UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10xx
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
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
