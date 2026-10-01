package main

import (
	"strings"
	"testing"
)

func TestSanitizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean semver",
			input:    "0.10.1",
			expected: "0.10.1",
		},
		{
			name:     "clean semver with v prefix",
			input:    "v0.10.1",
			expected: "0.10.1",
		},
		{
			name:     "git describe with distance and hash",
			input:    "0.10.0-5-g97df544",
			expected: "0.10.0",
		},
		{
			name:     "git describe with v prefix distance and hash",
			input:    "v0.10.0-5-g97df544",
			expected: "0.10.0",
		},
		{
			name:     "git describe single commit ahead",
			input:    "0.10.1-1-g97df544",
			expected: "0.10.1",
		},
		{
			name:     "git describe hash only suffix",
			input:    "0.10.1-g97df544",
			expected: "0.10.1",
		},
		{
			name:     "git describe with dirty suffix",
			input:    "0.10.1-1-g97df544-dirty",
			expected: "0.10.1",
		},
		{
			name:     "standalone dirty suffix",
			input:    "0.10.1-dirty",
			expected: "0.10.1",
		},
		{
			name:     "next minor release semver",
			input:    "v0.11.0",
			expected: "0.11.0",
		},
		{
			name:     "raw git short commit hash only",
			input:    "97df544",
			expected: "dev",
		},
		{
			name:     "raw git full commit hash only",
			input:    "97df544bf28e9b9dff88171270f6222044a1a7f",
			expected: "dev",
		},
		{
			name:     "go module pseudo-version",
			input:    "v0.0.0-20261001152033-97df544abcdef",
			expected: "dev",
		},
		{
			name:     "dev fallback",
			input:    "dev",
			expected: "dev",
		},
		{
			name:     "empty string fallback",
			input:    "",
			expected: "dev",
		},
		{
			name:     "none string fallback",
			input:    "none",
			expected: "dev",
		},
		{
			name:     "unknown string fallback",
			input:    "unknown",
			expected: "dev",
		},
		{
			name:     "whitespace padded",
			input:    "  0.10.1 \n",
			expected: "0.10.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeVersion(tc.input)
			if got != tc.expected {
				t.Errorf("sanitizeVersion(%q) = %q; want %q", tc.input, got, tc.expected)
			}
			if strings.Contains(got, "-g") {
				t.Fatalf("CRITICAL: sanitizeVersion(%q) returned string containing '-g': %q", tc.input, got)
			}
			if strings.Contains(got, "-dirty") {
				t.Fatalf("CRITICAL: sanitizeVersion(%q) returned string containing '-dirty': %q", tc.input, got)
			}
		})
	}
}
