package session

import (
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/config"
)

// FormatDir formats a working directory path for compact tabular display.
// It replaces the user's home directory with ~, extracts recognisable path components,
// and truncates if necessary to fit within maxLen characters.
func FormatDir(dir string, maxLen int) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "-"
	}
	home := config.RealHomeDir()
	if home != "" && strings.HasPrefix(dir, home) {
		dir = "~" + dir[len(home):]
	}
	dir = filepath.Clean(dir)
	if len(dir) <= maxLen {
		return dir
	}
	// Try showing the last 2 path segments (e.g. "my-projects/aim")
	parts := strings.Split(dir, string(filepath.Separator))
	if len(parts) >= 2 {
		short := parts[len(parts)-2] + "/" + parts[len(parts)-1]
		if len(short) <= maxLen {
			return short
		}
	}
	base := filepath.Base(dir)
	r := []rune(base)
	if len(r) <= maxLen {
		return base
	}
	if maxLen > 3 {
		return string(r[:maxLen-3]) + "..."
	}
	return string(r[:maxLen])
}
