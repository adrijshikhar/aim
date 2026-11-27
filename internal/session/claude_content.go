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

	return parseClaudeContentArrayElements(raw)
}

func parseClaudeContentArrayElements(raw json.RawMessage) ([]ClaudeContentBlock, string) {
	var rawElems []json.RawMessage
	if err := json.Unmarshal(raw, &rawElems); err != nil {
		return nil, ""
	}
	out := make([]ClaudeContentBlock, 0, len(rawElems))
	firstText := ""
	for _, elem := range rawElems {
		b, ok := decodeSingleClaudeBlock(elem)
		if !ok {
			continue
		}
		out = append(out, b)
		if firstText == "" && strings.TrimSpace(b.Text) != "" {
			firstText = strings.TrimSpace(b.Text)
		}
	}
	if len(out) == 0 {
		return nil, ""
	}
	return out, firstText
}

func decodeSingleClaudeBlock(elem json.RawMessage) (ClaudeContentBlock, bool) {
	var b ClaudeContentBlock
	if err := json.Unmarshal(elem, &b); err == nil {
		return b, true
	}
	var gb map[string]any
	if err := json.Unmarshal(elem, &gb); err == nil {
		typ, _ := gb["type"].(string)
		txt, _ := gb["text"].(string)
		return ClaudeContentBlock{Type: typ, Text: txt}, true
	}
	return ClaudeContentBlock{}, false
}
