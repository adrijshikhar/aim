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
		turns, err := parseClaudeTranscript(sess)
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

func parseClaudeTranscript(sess *session.Session) ([]SessionTurnDTO, error) {
	filePath := sess.StoragePath
	if filePath == "" || filepath.Ext(filePath) != ".jsonl" {
		// Look up in profile or host .claude/projects directory
		var baseProjectsDir string
		if sess.IsHost {
			baseProjectsDir = filepath.Join(config.RealHomeDir(), ".claude", "projects")
		} else if sess.Profile != "" {
			baseProjectsDir = filepath.Join(config.BaseDir(), "profiles", sess.Profile, ".claude", "projects")
		}

		if baseProjectsDir != "" {
			// Find matches
			matches, _ := filepath.Glob(filepath.Join(baseProjectsDir, "*", sess.ID+".jsonl"))
			if len(matches) > 0 {
				filePath = matches[0]
			}
		}
	}

	if filePath == "" {
		return nil, os.ErrNotExist
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var turns []SessionTurnDTO
	scanner := bufio.NewScanner(f)
	// Claude lines can be large due to embedded code or file content
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	type rawClaudeBlock struct {
		Type      string          `json:"type"`
		Text      string          `json:"text,omitempty"`
		Thinking  string          `json:"thinking,omitempty"`
		Name      string          `json:"name,omitempty"`
		Input     json.RawMessage `json:"input,omitempty"`
		Content   interface{}     `json:"content,omitempty"`
		IsError   bool            `json:"is_error,omitempty"`
		ToolUseID string          `json:"tool_use_id,omitempty"`
	}

	type rawClaudeMessage struct {
		Role    string           `json:"role"`
		Content []rawClaudeBlock `json:"content"`
	}

	type rawClaudeRecord struct {
		Type      string           `json:"type"`
		Timestamp string           `json:"timestamp"`
		Message   rawClaudeMessage `json:"message"`
	}

	turnIdx := 1
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var rec rawClaudeRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			continue
		}

		role := rec.Type
		if role == "" {
			role = rec.Message.Role
		}

		if role != "user" && role != "assistant" {
			continue
		}

		var parsedTime time.Time
		if rec.Timestamp != "" {
			parsedTime, _ = time.Parse(time.RFC3339, rec.Timestamp)
		}

		var textParts []string
		var thinkingParts []string
		var toolCalls []ToolCallDTO

		for _, b := range rec.Message.Content {
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

		content := strings.Join(textParts, "\n\n")
		thinking := strings.Join(thinkingParts, "\n\n")

		if role == "user" {
			content = cleanUserPrompt(content)
		}

		// If user turn has no text content and no tool calls, skip internal noise
		if strings.TrimSpace(content) == "" && strings.TrimSpace(thinking) == "" && len(toolCalls) == 0 {
			continue
		}

		turns = append(turns, SessionTurnDTO{
			TurnIndex: turnIdx,
			Role:      role,
			Timestamp: parsedTime,
			Content:   content,
			Thinking:  thinking,
			ToolCalls: toolCalls,
		})
		turnIdx++
	}

	return turns, scanner.Err()
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
