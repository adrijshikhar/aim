package session_test

import (
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/session"
)

func TestSession_ShortID(t *testing.T) {
	tests := []struct {
		name     string
		fullID   string
		expected string
	}{
		{
			name:     "36-char standard uuid",
			fullID:   "775e6ada-1595-4e7e-84fa-ce0ea71e3007",
			expected: "775e6ada",
		},
		{
			name:     "short id less than 8 chars",
			fullID:   "abc",
			expected: "abc",
		},
		{
			name:     "empty id",
			fullID:   "",
			expected: "",
		},
		{
			name:     "exact 8 chars",
			fullID:   "12345678",
			expected: "12345678",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := session.NewSession(tc.fullID, "title", "agy", "work", false, time.Now())
			if s.ShortID != tc.expected {
				t.Fatalf("expected ShortID %q, got %q", tc.expected, s.ShortID)
			}
		})
	}
}
