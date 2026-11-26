package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
)

// GetSessionTranscript resolves a session and parses its full turn replay from storage.
func (s *sessionService) GetSessionTranscript(ctx context.Context, agent, sessionID string) (*SessionTranscriptDTO, error) {
	agent = strings.TrimSpace(agent)
	if agent == "" {
		return nil, errors.New("agent is required")
	}
	if !safeIdentifierPattern.MatchString(agent) {
		return nil, fmt.Errorf("invalid agent %q: contains illegal characters", agent)
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}
	if !safeIdentifierPattern.MatchString(sessionID) {
		return nil, fmt.Errorf("invalid session ID %q: contains illegal characters", sessionID)
	}

	if s.mgr == nil {
		return nil, errors.New("session manager unavailable")
	}

	sess, err := s.mgr.ResolveSession(ctx, agent, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve session %s: %w", sessionID, err)
	}
	if sess == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	cwd := sess.Cwd
	if cwd == "" {
		cwd, _ = s.mgr.ResolveCwd(ctx, sess)
	}

	goal := sess.Goal
	if goal == "" {
		if sum, err := s.mgr.ResolveSummary(ctx, sess); err == nil {
			goal = sum.Goal
			if goal == "" {
				goal = sum.RecentActivity
			}
		}
	}
	if goal == "" {
		goal = sess.Summary
	}
	if goal == "" {
		goal = sess.Title
	}

	transcript := &SessionTranscriptDTO{
		SessionID: sess.ID,
		Agent:     sess.Agent,
		Profile:   sess.Profile,
		Title:     sess.Title,
		Cwd:       cwd,
		Goal:      goal,
		Turns:     []SessionTurnDTO{},
	}

	// Try reading native transcript by agent type
	switch strings.ToLower(sess.Agent) {
	case "claude":
		turns, err := parseClaudeTranscript(sess, s.profilesRoot)
		if len(turns) > 0 {
			transcript.Turns = turns
			return transcript, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			_ = err
		}
	case agentCodex:
		turns, err := parseCodexTranscript(sess, s.profilesRoot)
		if len(turns) > 0 {
			transcript.Turns = turns
			return transcript, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			_ = err
		}
	case "agy":
		turns, err := parseAgyTranscript(sess)
		if len(turns) > 0 {
			transcript.Turns = turns
			return transcript, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			_ = err
		}
	}

	// Fallback for sessions where native transcript parser didn't find turns or for other providers
	transcript.Turns = buildFallbackTurns(sess, goal)
	return transcript, nil
}

const (
	roleUser      = "user"
	roleAssistant = "assistant"
)

func resolveTranscriptProfilesRoot(injected string) string {
	if strings.TrimSpace(injected) != "" {
		return injected
	}
	return filepath.Join(config.BaseDir(), "profiles")
}

func openStorageInProfileRoot(root *os.Root, profDir, storagePath, expectedExt string) (*os.File, bool) {
	if storagePath == "" {
		return nil, false
	}
	if expectedExt != "" && filepath.Ext(storagePath) != expectedExt {
		return nil, false
	}
	rel, err := filepath.Rel(profDir, storagePath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return nil, false
	}
	f, err := root.Open(rel)
	return f, err == nil
}

func openIsolatedSessionFile(sess *session.Session, profilesRoot, expectedExt string, findRel func(*os.Root) (string, error)) (*os.File, error) {
	if err := profile.ValidateProfileName(sess.Profile); err != nil {
		return nil, err
	}
	profDir := filepath.Join(resolveTranscriptProfilesRoot(profilesRoot), sess.Profile)
	root, err := os.OpenRoot(profDir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	if f, ok := openStorageInProfileRoot(root, profDir, sess.StoragePath, expectedExt); ok {
		return f, nil
	}
	if findRel != nil {
		if rel, findErr := findRel(root); findErr == nil && rel != "" {
			return root.Open(rel)
		}
	}
	return nil, os.ErrNotExist
}

func openDirectSessionFile(filePath, expectedExt string) (*os.File, bool) {
	if filePath == "" || (expectedExt != "" && filepath.Ext(filePath) != expectedExt) {
		return nil, false
	}
	//nolint:gosec // G304: host StoragePath validated from resolved host session object.
	f, err := os.Open(filePath)
	return f, err == nil
}

func openScopedSessionFile(sess *session.Session, profilesRoot, agentRelBase, expectedExt string, findRel func(*os.Root) (string, error)) (*os.File, error) {
	if !sess.IsHost && sess.Profile != "" && sess.Profile != hostProfileSentinel {
		return openIsolatedSessionFile(sess, profilesRoot, expectedExt, findRel)
	}
	if f, ok := openDirectSessionFile(sess.StoragePath, expectedExt); ok {
		return f, nil
	}
	if !sess.IsHost || agentRelBase == "" || findRel == nil {
		return nil, os.ErrNotExist
	}
	root, err := os.OpenRoot(config.RealHomeDir())
	if err != nil {
		return nil, os.ErrNotExist
	}
	defer func() { _ = root.Close() }()
	rel, findErr := findRel(root)
	if findErr != nil || rel == "" {
		return nil, os.ErrNotExist
	}
	return root.Open(rel)
}

type rawClaudeRecord struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

func parseClaudeTranscriptLine(line []byte, turnIdx int) (SessionTurnDTO, bool) {
	var rec rawClaudeRecord
	if err := json.Unmarshal(line, &rec); err != nil {
		return SessionTurnDTO{}, false
	}

	role := rec.Type
	if role == "" {
		role = rec.Message.Role
	}
	if role != roleUser && role != roleAssistant {
		return SessionTurnDTO{}, false
	}

	var parsedTime time.Time
	if rec.Timestamp != "" {
		parsedTime, _ = time.Parse(time.RFC3339, rec.Timestamp)
	}

	blocks, _ := session.ParseClaudeContent(rec.Message.Content)
	content, thinking, toolCalls := collectClaudeBlockFields(blocks)
	if role == roleUser {
		content = cleanUserPrompt(content)
	}
	if strings.TrimSpace(content) == "" && strings.TrimSpace(thinking) == "" && len(toolCalls) == 0 {
		return SessionTurnDTO{}, false
	}

	return SessionTurnDTO{
		TurnIndex: turnIdx,
		Role:      role,
		Timestamp: parsedTime,
		Content:   content,
		Thinking:  thinking,
		ToolCalls: toolCalls,
	}, true
}

func collectClaudeBlockFields(blocks []session.ClaudeContentBlock) (string, string, []ToolCallDTO) {
	var textParts []string
	var thinkingParts []string
	var toolCalls []ToolCallDTO

	for _, b := range blocks {
		switch b.Type {
		case "text":
			if strings.TrimSpace(b.Text) != "" {
				textParts = append(textParts, b.Text)
			}
		case "thinking":
			if strings.TrimSpace(b.Thinking) != "" {
				thinkingParts = append(thinkingParts, b.Thinking)
			}
		case "tool_use":
			argsStr := string(b.Input)
			if argsStr == "" {
				argsStr = "{}"
			}
			toolCalls = append(toolCalls, ToolCallDTO{
				ID:        b.ToolUseID,
				ToolName:  b.Name,
				Arguments: argsStr,
				Status:    "ok",
			})
		}
	}
	return strings.Join(textParts, "\n\n"), strings.Join(thinkingParts, "\n\n"), toolCalls
}

func parseClaudeTranscript(sess *session.Session, profilesRoot string) ([]SessionTurnDTO, error) {
	f, err := openScopedSessionFile(sess, profilesRoot, filepath.Join(".claude", "projects"), ".jsonl", func(root *os.Root) (string, error) {
		projRoot, openErr := root.OpenRoot(filepath.Join(".claude", "projects"))
		if openErr != nil {
			return "", openErr
		}
		defer func() { _ = projRoot.Close() }()
		entries, readErr := os.ReadDir(projRoot.Name())
		if readErr != nil {
			return "", readErr
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			rel := filepath.Join(".claude", "projects", e.Name(), sess.ID+".jsonl")
			if _, statErr := root.Stat(rel); statErr == nil {
				return rel, nil
			}
		}
		return "", os.ErrNotExist
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var turns []SessionTurnDTO
	scanner := bufio.NewScanner(f)
	// Claude lines can be large due to embedded code or file content
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if turn, ok := parseClaudeTranscriptLine(line, len(turns)+1); ok {
			turns = append(turns, turn)
		}
	}

	return turns, scanner.Err()
}

type codexContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type codexRolloutLine struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Message   json.RawMessage `json:"message,omitempty"`
	Payload   struct {
		Type      string          `json:"type"`
		Role      string          `json:"role"`
		Content   json.RawMessage `json:"content,omitempty"`
		Message   string          `json:"message,omitempty"`
		Name      string          `json:"name,omitempty"`
		Arguments string          `json:"arguments,omitempty"`
		CallID    string          `json:"call_id,omitempty"`
	} `json:"payload"`
}

func extractCodexMessageText(raw json.RawMessage) string {
	var parts []codexContentPart
	if err := json.Unmarshal(raw, &parts); err == nil && len(parts) > 0 {
		var texts []string
		for _, p := range parts {
			if strings.TrimSpace(p.Text) != "" {
				texts = append(texts, p.Text)
			}
		}
		return strings.Join(texts, "\n\n")
	}
	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return plain
	}
	return ""
}

func findCodexRolloutRel(root *os.Root, sessionID string) (string, error) {
	sessionsPath := filepath.Join(root.Name(), ".codex", "sessions")
	var match string
	_ = filepath.WalkDir(sessionsPath, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return walkErr
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") && strings.Contains(d.Name(), sessionID) {
			if rel, relErr := filepath.Rel(root.Name(), p); relErr == nil {
				match = rel
				return filepath.SkipAll
			}
		}
		return nil
	})
	if match == "" {
		return "", os.ErrNotExist
	}
	return match, nil
}

func appendCodexTurn(dst *[]SessionTurnDTO, role string, ts time.Time, text string, calls []ToolCallDTO) {
	if role == roleUser {
		text = cleanUserPrompt(text)
	} else {
		text = cleanAssistantResponse(text)
	}
	if strings.TrimSpace(text) == "" && len(calls) == 0 {
		return
	}
	*dst = append(*dst, SessionTurnDTO{
		TurnIndex: len(*dst) + 1,
		Role:      role,
		Timestamp: ts,
		Content:   text,
		ToolCalls: calls,
	})
}

func processCodexResponseItem(rec codexRolloutLine, ts time.Time, responseTurns *[]SessionTurnDTO) {
	switch rec.Payload.Type {
	case "message":
		if rec.Payload.Role == roleUser || rec.Payload.Role == roleAssistant {
			appendCodexTurn(responseTurns, rec.Payload.Role, ts, extractCodexMessageText(rec.Payload.Content), nil)
		}
	case "function_call":
		appendCodexTurn(responseTurns, roleAssistant, ts, "", []ToolCallDTO{{
			ID:        rec.Payload.CallID,
			ToolName:  rec.Payload.Name,
			Arguments: rec.Payload.Arguments,
			Status:    "ok",
		}})
	}
}

func processCodexRolloutLine(line []byte, responseTurns, eventTurns *[]SessionTurnDTO) {
	var rec codexRolloutLine
	if err := json.Unmarshal(line, &rec); err != nil {
		return
	}
	var ts time.Time
	if rec.Timestamp != "" {
		ts, _ = time.Parse(time.RFC3339Nano, rec.Timestamp)
	}

	switch rec.Type {
	case "response_item":
		processCodexResponseItem(rec, ts, responseTurns)
	case "event_msg":
		switch rec.Payload.Type {
		case "user_message":
			appendCodexTurn(eventTurns, roleUser, ts, rec.Payload.Message, nil)
		case "agent_message":
			appendCodexTurn(eventTurns, roleAssistant, ts, rec.Payload.Message, nil)
		}
	case roleUser, roleAssistant, "USER_INPUT", "PLANNER_RESPONSE":
		role := roleUser
		if rec.Type == roleAssistant || rec.Type == "PLANNER_RESPONSE" {
			role = roleAssistant
		}
		_, txt := session.ParseClaudeContent(rec.Message)
		appendCodexTurn(eventTurns, role, ts, txt, nil)
	}
}

func parseCodexTranscript(sess *session.Session, profilesRoot string) ([]SessionTurnDTO, error) {
	f, err := openScopedSessionFile(sess, profilesRoot, filepath.Join(".codex", "sessions"), ".jsonl", func(root *os.Root) (string, error) {
		return findCodexRolloutRel(root, sess.ID)
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var responseTurns []SessionTurnDTO
	var eventTurns []SessionTurnDTO

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) > 0 {
			processCodexRolloutLine(line, &responseTurns, &eventTurns)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(responseTurns) > 0 {
		return responseTurns, nil
	}
	return eventTurns, nil
}

func parseAgyTranscript(sess *session.Session) ([]SessionTurnDTO, error) {
	logPath := sess.StoragePath
	if !strings.HasSuffix(logPath, "transcript.jsonl") {
		logPath = filepath.Join(logPath, ".system_generated", "logs", "transcript.jsonl")
	}

	f, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var turns []SessionTurnDTO
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	type rawAgyToolCall struct {
		Name string                 `json:"name"`
		Args map[string]interface{} `json:"args"`
	}

	type rawAgyStep struct {
		StepIndex int              `json:"step_index"`
		Source    string           `json:"source"`
		Type      string           `json:"type"`
		Content   string           `json:"content"`
		Thinking  string           `json:"thinking"`
		CreatedAt string           `json:"created_at"`
		ToolCalls []rawAgyToolCall `json:"tool_calls"`
	}

	turnIdx := 1
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var step rawAgyStep
		if err := json.Unmarshal(line, &step); err != nil {
			continue
		}

		var role string
		switch {
		case step.Source == "USER_EXPLICIT" || step.Type == "USER_INPUT" || step.Type == "SYSTEM_MESSAGE":
			role = "user"
		case step.Type == "PLANNER_RESPONSE" || step.Source == "MODEL":
			role = "assistant"
		default:
			continue
		}

		var parsedTime time.Time
		if step.CreatedAt != "" {
			parsedTime, _ = time.Parse(time.RFC3339, step.CreatedAt)
		}

		var toolCalls []ToolCallDTO
		for _, tc := range step.ToolCalls {
			argsBytes, _ := json.Marshal(tc.Args)
			toolCalls = append(toolCalls, ToolCallDTO{
				ToolName:  tc.Name,
				Arguments: string(argsBytes),
				Status:    "ok",
			})
		}

		content := strings.TrimSpace(step.Content)
		if role == "user" {
			content = cleanUserPrompt(content)
		} else {
			content = cleanAssistantResponse(content)
		}

		if content == "" && step.Thinking == "" && len(toolCalls) == 0 {
			continue
		}

		turns = append(turns, SessionTurnDTO{
			TurnIndex: turnIdx,
			Role:      role,
			Timestamp: parsedTime,
			Content:   content,
			Thinking:  step.Thinking,
			ToolCalls: toolCalls,
		})
		turnIdx++
	}

	return turns, scanner.Err()
}

var (
	userRequestPattern    = regexp.MustCompile(`(?s)<USER_REQUEST>\s*(.*?)\s*</USER_REQUEST>`)
	contextSummaryPattern = regexp.MustCompile(`(?s)<CONTEXT_SUMMARY>.*?</CONTEXT_SUMMARY>`)
	systemPromptPattern   = regexp.MustCompile(`(?s)<(identity|user_information|user_rules|skills|subagents|slash_commands|conversation_transcript|artifacts|guidelines|communication_style)>.*?</(identity|user_information|user_rules|skills|subagents|slash_commands|conversation_transcript|artifacts|guidelines|communication_style)>`)
	claudeReminderPattern = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>`)
)

func cleanUserPrompt(s string) string {
	s = strings.TrimSpace(s)
	if matches := userRequestPattern.FindStringSubmatch(s); len(matches) > 1 {
		req := strings.TrimSpace(matches[1])
		if req != "" {
			return req
		}
	}
	if strings.Contains(s, "<CONTEXT_SUMMARY>") {
		cleaned := contextSummaryPattern.ReplaceAllString(s, "")
		cleaned = systemPromptPattern.ReplaceAllString(cleaned, "")
		cleaned = claudeReminderPattern.ReplaceAllString(cleaned, "")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned != "" {
			return cleaned
		}
		return "[Session Resumed — Context Summary Checkpoint]"
	}
	if strings.Contains(s, "<SYSTEM_MESSAGE>") {
		if idx := strings.Index(s, "content="); idx != -1 {
			inner := s[idx+len("content="):]
			if endIdx := strings.Index(inner, "</SYSTEM_MESSAGE>"); endIdx != -1 {
				inner = inner[:endIdx]
			}
			inner = strings.TrimSpace(inner)
			if inner != "" {
				s = inner
			}
		}
	}
	s = systemPromptPattern.ReplaceAllString(s, "")
	s = claudeReminderPattern.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func cleanAssistantResponse(s string) string {
	s = strings.TrimSpace(s)
	return s
}

func buildFallbackTurns(sess *session.Session, goal string) []SessionTurnDTO {
	turns := []SessionTurnDTO{}

	// User Turn (Initial Goal or Request)
	userContent := goal
	if userContent == "" {
		userContent = sess.Title
	}
	if userContent == "" {
		userContent = "Session initialized: " + sess.ID
	}

	turns = append(turns, SessionTurnDTO{
		TurnIndex: 1,
		Role:      "user",
		Timestamp: sess.StartedAt,
		Content:   userContent,
	})

	// Assistant Turn (Summary / Status)
	assistantContent := sess.Summary
	if assistantContent == "" {
		assistantContent = sess.Recent
	}
	if assistantContent == "" {
		assistantContent = fmt.Sprintf("Session active in %s with %d total turns recorded.", sess.Cwd, sess.MessageCount)
	}

	turns = append(turns, SessionTurnDTO{
		TurnIndex: 2,
		Role:      "assistant",
		Timestamp: sess.LastActiveAt,
		Content:   assistantContent,
	})

	return turns
}
