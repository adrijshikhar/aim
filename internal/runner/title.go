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
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "resume" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--conversation=") {
			val := strings.TrimPrefix(arg, "--conversation=")
			if val != "" {
				return val
			}
		}
		if arg == "--conversation" && i+1 < len(args) {
			if !strings.HasPrefix(args[i+1], "-") {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "-c=") {
			val := strings.TrimPrefix(arg, "-c=")
			if val != "" {
				return val
			}
		}
		if arg == "-c" && i+1 < len(args) {
			if !strings.HasPrefix(args[i+1], "-") {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--resume=") {
			val := strings.TrimPrefix(arg, "--resume=")
			if val != "" {
				return val
			}
		}
		if arg == "--resume" && i+1 < len(args) {
			if !strings.HasPrefix(args[i+1], "-") {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "-r=") {
			val := strings.TrimPrefix(arg, "-r=")
			if val != "" {
				return val
			}
		}
		if arg == "-r" && i+1 < len(args) {
			if !strings.HasPrefix(args[i+1], "-") {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--session-id=") {
			val := strings.TrimPrefix(arg, "--session-id=")
			if val != "" {
				return val
			}
		}
		if arg == "--session-id" && i+1 < len(args) {
			if !strings.HasPrefix(args[i+1], "-") {
				return args[i+1]
			}
		}
	}
	return ""
}
