package diagnostics

import (
	"fmt"
	"net/url"
	"regexp"
	"runtime"
	"strings"
)

var (
	// tokenPatterns matches common authorization tokens, bearer headers, and API keys.
	tokenPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(bearer\s+|token[=:\s]+|key[=:\s]+|secret[=:\s]+)[a-zA-Z0-9_\-\.]{8,}`),
		regexp.MustCompile(`(ya29\.[a-zA-Z0-9_\-]+)`),
		regexp.MustCompile(`(gh[pousr]_[a-zA-Z0-9]{36,})`),
		regexp.MustCompile(`(sk-[a-zA-Z0-9_\-]{20,})`),
		regexp.MustCompile(`(ph[cx]_[a-zA-Z0-9_\-]{20,})`),
	}
)

// SanitizeText removes personal home paths and sensitive credentials/tokens from arbitrary strings.
func SanitizeText(text string, homeDirs ...string) string {
	for _, homeDir := range homeDirs {
		if homeDir != "" && homeDir != "/" {
			text = strings.ReplaceAll(text, homeDir, "~")
		}
	}

	for _, re := range tokenPatterns {
		text = re.ReplaceAllStringFunc(text, func(m string) string {
			if idx := strings.IndexAny(m, " =:"); idx != -1 {
				prefix := m[:idx+1]
				return prefix + "[REDACTED]"
			}
			return "[REDACTED]"
		})
	}

	return text
}

// SanitizeStack removes personal home paths and sensitive credentials/tokens from stack traces.
func SanitizeStack(stack string, homeDir string) string {
	return SanitizeText(stack, homeDir)
}

// BuildGitHubIssueURL constructs a pre-filled GitHub issue URL with sanitized panic details.
func BuildGitHubIssueURL(title, body string) string {
	baseURL := "https://github.com/adrijshikhar/aim/issues/new"

	// URLs exceeding ~6,000 chars risk 414 Request-URI Too Large in browsers
	if len(body) > 4000 {
		body = body[:4000] + "\n\n... (stack trace truncated, see full trace in ~/.aim/logs/aim-debug.log)"
	}

	params := url.Values{}
	params.Set("title", title)
	params.Set("body", body)
	params.Set("labels", "bug,crash")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// HandlePanic processes an unhandled panic, scrubs sensitive paths and tokens,
// and returns both a GitHub issue creation URL and a formatted diagnostic string.
func HandlePanic(recovered any, stack []byte, homeDir string, version, commit string) (issueURL string, report string) {
	panicMsg := SanitizeText(fmt.Sprintf("%v", recovered), homeDir)
	sanitizedStack := SanitizeStack(string(stack), homeDir)

	report = fmt.Sprintf(`### Crash Details
**AIM Version:** %s (commit: %s)
**Environment:** %s/%s (%s)

**Panic:**
`+"```"+`
%s
`+"```"+`

**Sanitized Stack Trace:**
`+"```"+`
%s
`+"```"+`
`, version, commit, runtime.GOOS, runtime.GOARCH, runtime.Version(), panicMsg, sanitizedStack)

	issueTitle := fmt.Sprintf("Crash: %s", truncate(panicMsg, 60))
	issueURL = BuildGitHubIssueURL(issueTitle, report)

	return issueURL, report
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
