package session_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/session"
)

func TestFormatDir(t *testing.T) {
	home := config.RealHomeDir()

	tests := []struct {
		name     string
		dir      string
		maxLen   int
		expected string
	}{
		{
			name:     "empty dir",
			dir:      "",
			maxLen:   16,
			expected: "-",
		},
		{
			name:     "whitespace dir",
			dir:      "   ",
			maxLen:   16,
			expected: "-",
		},
		{
			name:     "short path within limit",
			dir:      "/workspace/app",
			maxLen:   16,
			expected: "/workspace/app",
		},
		{
			name:     "home path replaces with tilde",
			dir:      filepath.Join(home, "repo"),
			maxLen:   16,
			expected: "~/repo",
		},
		{
			name:     "two trailing segments fit",
			dir:      filepath.Join(home, "Projects", "my-projects", "aim"),
			maxLen:   16,
			expected: "my-projects/aim",
		},
		{
			name:     "fallback to base when two segments exceed maxLen",
			dir:      filepath.Join(home, "Projects", "very-long-category-name", "cxstatusline"),
			maxLen:   16,
			expected: "cxstatusline",
		},
		{
			name:     "truncate base when base exceeds maxLen",
			dir:      "/very/long/path/extremely-long-repository-name-that-is-way-too-big",
			maxLen:   16,
			expected: "extremely-lon...",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := session.FormatDir(tc.dir, tc.maxLen)
			if !strings.EqualFold(got, tc.expected) && got != tc.expected {
				t.Errorf("FormatDir(%q, %d) = %q; expected %q", tc.dir, tc.maxLen, got, tc.expected)
			}
			if len(got) > tc.maxLen {
				t.Errorf("FormatDir result %q exceeds maxLen %d", got, tc.maxLen)
			}
		})
	}
}
