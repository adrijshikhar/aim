package usage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/oauth"
)

func TestParseRefusalFromError_HTTP401Boundaries(t *testing.T) {
	tests := []struct {
		name         string
		errStr       string
		expectedCode oauth.RefusalCode
	}{
		{
			name:         "port number with 401 should not match",
			errStr:       "connect to 127.0.0.1:4012: connection refused",
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "timestamp with 401 should not match",
			errStr:       "request failed at 14:01:00 UTC",
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "byte count with 401 should not match",
			errStr:       "read 4012 bytes from socket",
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "request ID with 401 should not match",
			errStr:       "request-id: abc-4012-xyz failed",
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "space separated 401",
			errStr:       "received 401 error from server",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "prefix 401",
			errStr:       "401 unauthorized",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "status 401",
			errStr:       "HTTP status 401",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "status: 401",
			errStr:       "status: 401",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "status code: 401",
			errStr:       "HTTP status code: 401",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "http 401",
			errStr:       "http 401 returned",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "JSON error with port 4012 should not trigger 401 refusal",
			errStr:       `{"error": "network_failure", "endpoint": "127.0.0.1:4012"}`,
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "trailing suffix 401",
			errStr:       "request failed with 401",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "colon separated 401",
			errStr:       "received 401: token revoked",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "parenthesized 401",
			errStr:       "HTTP error (401)",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "bracketed 401",
			errStr:       "[401] Token rejected",
			expectedCode: oauth.RefusalInvalidGrant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := parseRefusalFromError(tt.errStr)
			if code != tt.expectedCode {
				t.Fatalf("expected code %v, got %v for error %q", tt.expectedCode, code, tt.errStr)
			}
		})
	}
}

func TestInspectProfile_CachedUsage(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	credPath := filepath.Join(claudeDir, ".credentials.json")
	if err := os.WriteFile(credPath, []byte(`{"sessionKey":"test-key"}`), 0600); err != nil {
		t.Fatalf("failed to write credentials: %v", err)
	}

	t.Run("returns cached usage windows without calling live usage", func(t *testing.T) {
		liveUsageInvoked := false
		cachedUsageInvoked := false

		target := TargetProfile{
			Agent:      "claude",
			Profile:    "work",
			ProfileDir: tmpDir,
			GetUsageFn: func(ctx context.Context, profileName, profileDir string) (*Report, error) {
				liveUsageInvoked = true
				return nil, errors.New("live usage must never be called during inspection")
			},
			GetCachedUsageFn: func(ctx context.Context, profileName, profileDir string) (*Report, error) {
				cachedUsageInvoked = true
				return &Report{
					Status: StatusOK,
					Windows: []LimitWindow{
						{
							Name:         "Session (5-hour)",
							RemainingPct: 92,
						},
						{
							Name:         "Weekly (7-day)",
							RemainingPct: 45,
						},
					},
				}, nil
			},
		}

		rep := InspectProfile(context.Background(), target, nil)

		if liveUsageInvoked {
			t.Errorf("GetUsageFn was invoked, but InspectProfile must strictly not call live usage")
		}
		if !cachedUsageInvoked {
			t.Errorf("expected GetCachedUsageFn to be invoked, but it was not")
		}
		if !rep.FromCache {
			t.Errorf("expected rep.FromCache to be true, got %v", rep.FromCache)
		}
		if rep.Agent != "claude" {
			t.Errorf("expected rep.Agent to be 'claude', got %q", rep.Agent)
		}
		if rep.Profile != "work" {
			t.Errorf("expected rep.Profile to be 'work', got %q", rep.Profile)
		}
		if rep.Status != StatusOK {
			t.Errorf("expected rep.Status to be StatusOK, got %q", rep.Status)
		}
		if len(rep.Windows) != 2 {
			t.Fatalf("expected 2 windows, got %d", len(rep.Windows))
		}
		if rep.Windows[0].RemainingPct != 92 {
			t.Errorf("expected first window RemainingPct 92, got %v", rep.Windows[0].RemainingPct)
		}
	})

	t.Run("falls back to StatusIdle when GetCachedUsageFn returns error", func(t *testing.T) {
		target := TargetProfile{
			Agent:      "claude",
			Profile:    "work",
			ProfileDir: tmpDir,
			GetCachedUsageFn: func(ctx context.Context, profileName, profileDir string) (*Report, error) {
				return nil, errors.New("cache read error")
			},
		}

		rep := InspectProfile(context.Background(), target, nil)
		if rep.Status != StatusIdle {
			t.Errorf("expected StatusIdle fallback, got %q", rep.Status)
		}
		if rep.Summary != "Ready on use" {
			t.Errorf("expected 'Ready on use', got %q", rep.Summary)
		}
	})

	t.Run("falls back to StatusIdle when GetCachedUsageFn returns empty windows", func(t *testing.T) {
		target := TargetProfile{
			Agent:      "claude",
			Profile:    "work",
			ProfileDir: tmpDir,
			GetCachedUsageFn: func(ctx context.Context, profileName, profileDir string) (*Report, error) {
				return &Report{
					Status:  StatusOK,
					Windows: nil,
				}, nil
			},
		}

		rep := InspectProfile(context.Background(), target, nil)
		if rep.Status != StatusIdle {
			t.Errorf("expected StatusIdle fallback, got %q", rep.Status)
		}
		if rep.Summary != "Ready on use" {
			t.Errorf("expected 'Ready on use', got %q", rep.Summary)
		}
	})

	t.Run("falls back to StatusIdle when GetCachedUsageFn is nil", func(t *testing.T) {
		target := TargetProfile{
			Agent:      "claude",
			Profile:    "work",
			ProfileDir: tmpDir,
		}

		rep := InspectProfile(context.Background(), target, nil)
		if rep.Status != StatusIdle {
			t.Errorf("expected StatusIdle fallback, got %q", rep.Status)
		}
		if rep.Summary != "Ready on use" {
			t.Errorf("expected 'Ready on use', got %q", rep.Summary)
		}
	})
}
