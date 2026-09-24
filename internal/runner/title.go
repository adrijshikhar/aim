package runner

import (
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-isatty"
)

var isTerminalFunc = func(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func isTerminal(w io.Writer) bool {
	if f, ok := w.(interface{ Fd() uintptr }); ok {
		return isTerminalFunc(f.Fd())
	}
	// Non-file io.Writer (e.g. bytes.Buffer in tests) is treated as interactive
	return true
}

var titleSanitizer = strings.NewReplacer("\r", "", "\n", "", "\007", "", "\033", "")

func sanitizeTitle(title string) string {
	return titleSanitizer.Replace(title)
}

// SetTerminalTitle emits an OSC 0 escape sequence to set the terminal tab/window title.
// If the destination writer is a non-interactive file/pipe, it is a no-op.
func SetTerminalTitle(w io.Writer, title string) error {
	if w == nil || !isTerminal(w) {
		return nil
	}
	cleanTitle := sanitizeTitle(title)
	_, err := fmt.Fprintf(w, "\033]0;%s\007", cleanTitle)
	return err
}

// ResetTerminalTitle emits an OSC 0 escape sequence with an empty title to reset the terminal title.
// If the destination writer is a non-interactive file/pipe, it is a no-op.
func ResetTerminalTitle(w io.Writer) error {
	if w == nil || !isTerminal(w) {
		return nil
	}
	_, err := fmt.Fprintf(w, "\033]0;\007")
	return err
}

// FormatTitle returns a formatted terminal title string for an agent, profile, and session.
// Format: "AIM: [<agent>] <profile> (<session-id>)" if session-id is present,
// or "AIM: [<agent>] <profile>" if no session-id is present.
func FormatTitle(agent, profile, sessionID string) string {
	var base string
	if agent != "" && profile != "" {
		base = fmt.Sprintf("AIM: [%s] %s", agent, profile)
	} else if agent != "" {
		base = fmt.Sprintf("AIM: [%s]", agent)
	} else if profile != "" {
		base = fmt.Sprintf("AIM: %s", profile)
	} else {
		base = "AIM"
	}
	if sessionID != "" {
		return fmt.Sprintf("%s (%s)", base, sessionID)
	}
	return base
}

// ExtractSessionID inspects command line arguments to detect a resumed session ID or conversation.
func ExtractSessionID(args []string) string {
	_, value, _ := sessionArgument(args, 0)
	return value
}

// ReplaceSessionID replaces the first matching session argument value without
// disturbing the spelling or position of unrelated arguments.
func ReplaceSessionID(args []string, oldID, newID string) []string {
	if oldID == "" || newID == "" || oldID == newID {
		return args
	}
	result := append([]string(nil), args...)
	for start := 0; start < len(result); {
		index, value, prefix := sessionArgument(result, start)
		if index < 0 {
			break
		}
		if value == oldID {
			if prefix == "" {
				result[index] = newID
			} else {
				result[index] = prefix + newID
			}
			break
		}
		start = index + 1
	}
	return result
}

// sessionArgument returns a session value's argument index, value, and its
// attached-value prefix (empty for a separate argument), starting at start.
func sessionArgument(args []string, start int) (int, string, string) {
	for i := start; i < len(args); i++ {
		for _, flag := range []string{"resume", "--conversation", "-c", "--resume", "-r", "--session-id"} {
			if args[i] == flag && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				return i + 1, args[i+1], ""
			}
			prefix := flag + "="
			if flag != "resume" && strings.HasPrefix(args[i], prefix) && len(args[i]) > len(prefix) {
				return i, strings.TrimPrefix(args[i], prefix), prefix
			}
		}
	}
	return -1, "", ""
}
