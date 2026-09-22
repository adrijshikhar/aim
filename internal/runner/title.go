package runner

import (
	"fmt"
	"io"
	"strings"
)

// SetTerminalTitle emits an OSC 0 escape sequence to set the terminal tab/window title.
func SetTerminalTitle(w io.Writer, title string) error {
	if w == nil {
		return nil
	}
	_, err := fmt.Fprintf(w, "\033]0;%s\007", title)
	return err
}

// ResetTerminalTitle emits an OSC 0 escape sequence with an empty title to reset the terminal title.
func ResetTerminalTitle(w io.Writer) error {
	if w == nil {
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

// extractSessionID inspects command line arguments to detect a resumed session ID or conversation.
func extractSessionID(args []string) string {
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
	}
	return ""
}
