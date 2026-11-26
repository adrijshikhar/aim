// Package session provides cross-agent session discovery, turn parsing, and hydration helpers.
package session

import (
	"encoding/json"
	"strings"
)

// ClaudeContentBlock represents a structured content block inside Claude Code JSONL transcripts.
type ClaudeContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	Content   any             `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
}

// ParseClaudeContent decodes a Claude message content field which may be either
// a plain JSON string or an array of structured content blocks.
func ParseClaudeContent(raw json.RawMessage) ([]ClaudeContentBlock, string) {
	if len(raw) == 0 {
		return nil, ""
	}

	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		trimmed := strings.TrimSpace(plain)
		if trimmed == "" {
			return nil, ""
		}
		return []ClaudeContentBlock{{Type: "text", Text: plain}}, trimmed
	}

	var blocks []ClaudeContentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil && len(blocks) > 0 {
		for _, b := range blocks {
			if strings.TrimSpace(b.Text) != "" {
				return blocks, strings.TrimSpace(b.Text)
			}
		}
		return blocks, ""
	}

	var genericBlocks []map[string]any
	if err := json.Unmarshal(raw, &genericBlocks); err == nil {
		out := make([]ClaudeContentBlock, 0, len(genericBlocks))
		firstText := ""
		for _, gb := range genericBlocks {
			typ, _ := gb["type"].(string)
			txt, _ := gb["text"].(string)
			out = append(out, ClaudeContentBlock{Type: typ, Text: txt})
			if firstText == "" && strings.TrimSpace(txt) != "" {
				firstText = strings.TrimSpace(txt)
			}
		}
		return out, firstText
	}

	return nil, ""
}
