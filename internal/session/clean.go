package session

import (
	"regexp"
	"strings"
)

var (
	// Matches screenshot file paths like:
	// /var/folders/.../Screenshot 2026-10-02 at 2.26.50 PM.png
	// /tmp/Screenshot 2026-10-02 at 2.26.50 PM.png
	// /tmp/.../screenshot.png
	// or [Image #1], [Image: /path]
	screenshotPathRegex = regexp.MustCompile(`(?i)(?:^|\s)(?:\[Image(?:\s*#\d+|:\s*[^\]]+)?\]|/(?:private/)?(?:var/folders|tmp)/.*?Screenshot[^\n]+?\.(?:png|jpg|jpeg|gif|webp)|/(?:private/)?(?:var/folders|tmp)/[^\s\n]+\.(?:png|jpg|jpeg|gif|webp)|Screenshot\s+[\d\-]+\s+at\s+[\d\.\sAPMapm]+\.(?:png|jpg|jpeg|gif|webp))(?:\s+|$)`)

	// Matches local-command-caveat tags and their content
	localCaveatRegex = regexp.MustCompile(`(?s)<local-command-caveat>.*?</local-command-caveat>`)

	// Matches additional metadata tags and their content
	additionalMetadataRegex = regexp.MustCompile(`(?s)<ADDITIONAL_METADATA>.*?</ADDITIONAL_METADATA>`)

	// Matches user information tags and their content
	userInformationRegex = regexp.MustCompile(`(?s)<user_information>.*?</user_information>`)

	// Matches user settings tags and their content
	userSettingsRegex = regexp.MustCompile(`(?s)<USER_SETTINGS>.*?</USER_SETTINGS>`)

	// Matches system tags
	systemTagsRegex = regexp.MustCompile(`(?s)<SYSTEM>.*?</SYSTEM>`)

	// Matches user request wrappers to extract inner content
	userRequestOpenRegex  = regexp.MustCompile(`(?i)<USER_REQUEST>\s*`)
	userRequestCloseRegex = regexp.MustCompile(`(?i)\s*</USER_REQUEST>`)

	// Matches summary tags
	summaryOpenRegex  = regexp.MustCompile(`(?i)<summary>\s*`)
	summaryCloseRegex = regexp.MustCompile(`(?i)\s*</summary>`)

	// Multiple spaces/newlines
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// CleanPromptText cleans raw prompts, stripping out screenshot paths,
// XML/command caveat wrappers, system tags, and normalizes whitespace.
func CleanPromptText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// 1. Remove known wrapped metadata blocks
	raw = localCaveatRegex.ReplaceAllString(raw, "")
	raw = additionalMetadataRegex.ReplaceAllString(raw, "")
	raw = userInformationRegex.ReplaceAllString(raw, "")
	raw = userSettingsRegex.ReplaceAllString(raw, "")
	raw = systemTagsRegex.ReplaceAllString(raw, "")

	// 2. Extract content from <summary>...</summary> if present
	if strings.Contains(raw, "<summary>") && strings.Contains(raw, "</summary>") {
		start := strings.Index(raw, "<summary>") + len("<summary>")
		end := strings.Index(raw, "</summary>")
		if end > start {
			raw = raw[start:end]
		}
	}

	// 3. Extract content from <USER_REQUEST>...</USER_REQUEST> if present
	if strings.Contains(raw, "<USER_REQUEST>") && strings.Contains(raw, "</USER_REQUEST>") {
		start := strings.Index(raw, "<USER_REQUEST>") + len("<USER_REQUEST>")
		end := strings.Index(raw, "</USER_REQUEST>")
		if end > start {
			raw = raw[start:end]
		}
	} else {
		raw = userRequestOpenRegex.ReplaceAllString(raw, "")
		raw = userRequestCloseRegex.ReplaceAllString(raw, "")
	}

	// 4. Strip screenshot file paths and image markers
	raw = screenshotPathRegex.ReplaceAllString(raw, " ")

	// 5. Normalize whitespace
	raw = whitespaceRegex.ReplaceAllString(raw, " ")
	raw = strings.TrimSpace(raw)

	// 6. Cap to a sensible length for session titles/summaries
	if len(raw) > 500 {
		raw = strings.TrimSpace(raw[:500]) + "..."
	}

	return raw
}
