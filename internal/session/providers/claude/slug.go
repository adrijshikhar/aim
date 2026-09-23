package claude

import (
	"os"
	"path/filepath"
	"strings"
)

// PathToSlug converts an absolute filesystem path to Claude Code's directory slug format.
// e.g. /Users/nemesis/Projects/my-projects/aim -> -Users-nemesis-Projects-my-projects-aim
func PathToSlug(path string) string {
	if path == "" {
		return ""
	}

	// Ensure leading slash for consistency if a relative path was passed
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "\\") {
		path = "/" + path
	}

	cleaned := filepath.Clean(path)
	toSlash := filepath.ToSlash(cleaned)
	if toSlash == "/" {
		return "-"
	}

	return strings.ReplaceAll(toSlash, "/", "-")
}

// SlugToPath converts a Claude Code directory slug back to an absolute filesystem path.
// e.g. -Users-nemesis-Projects-aim -> /Users/nemesis/Projects/aim
// If directories exist on disk, filesystem resolution is used to disambiguate
// hyphens that were originally part of directory names (e.g. "my-projects").
func SlugToPath(slug string) string {
	if slug == "" {
		return ""
	}

	trimmed := strings.Trim(slug, "-")
	if trimmed == "" {
		return "/"
	}

	// Check if current working directory matches the slug first (O(1) fast-path).
	if wd, err := os.Getwd(); err == nil {
		slugWd := PathToSlug(wd)
		if slugWd == slug || slugWd == "-"+trimmed {
			return wd
		}
	}

	// Try resolving against the live filesystem.
	tokens := strings.Split(trimmed, "-")
	if resolved, ok := resolvePathFromTokens("/", tokens); ok {
		return resolved
	}

	// Default fallback: replace hyphens with directory separators.
	return filepath.Clean("/" + strings.ReplaceAll(trimmed, "-", "/"))
}

// resolvePathFromTokens recursively searches the filesystem to match token segments
// against real directories, properly handling directory names containing hyphens.
func resolvePathFromTokens(current string, tokens []string) (string, bool) {
	if len(tokens) == 0 {
		return current, true
	}

	for k := 1; k <= len(tokens); k++ {
		segment := strings.Join(tokens[:k], "-")
		candidate := filepath.Join(current, segment)
		fi, err := os.Stat(candidate)
		if err == nil && fi.IsDir() {
			if res, ok := resolvePathFromTokens(candidate, tokens[k:]); ok {
				return res, true
			}
		}
	}
	return "", false
}
