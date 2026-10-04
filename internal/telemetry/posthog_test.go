package telemetry

import (
	"context"
	"os"
	"testing"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

func TestAccountDistinctID(t *testing.T) {
	// 1. With agent and userID
	account := profile.AccountInfo{
		UserID: "user_abc123",
		Email:  "test@example.com",
	}
	id := AccountDistinctID("codex", account)
	if id != "codex:user_abc123" {
		t.Fatalf("expected 'codex:user_abc123', got %q", id)
	}

	// 2. Without userID (falls back to anonymous machine ID)
	emptyAccount := profile.AccountInfo{Email: "test@example.com"}
	anonID := AccountDistinctID("claude", emptyAccount)
	if anonID == "" {
		t.Fatalf("expected non-empty anonymous ID fallback, got empty string")
	}
	if anonID == "claude:" {
		t.Fatalf("did not expect 'claude:', expected anonymous machine ID")
	}
}

func TestPostHog_DisabledWhenTelemetryDisabled(t *testing.T) {
	origDNT := os.Getenv("DO_NOT_TRACK")
	defer func() { _ = os.Setenv("DO_NOT_TRACK", origDNT) }()
	_ = os.Setenv("DO_NOT_TRACK", "1")

	InitPostHog(nil)
	if PostHogClient() != nil {
		t.Fatalf("expected PostHog client to be nil when DO_NOT_TRACK=1")
	}

	InitPostHogLogs(nil)
	if posthogLogProvider != nil {
		t.Fatalf("expected posthogLogProvider to be nil when DO_NOT_TRACK=1")
	}

	// Identify and capture should be safe no-ops
	IdentifyAccount("codex", profile.AccountInfo{UserID: "123"})
	CaptureAccountEvent("codex", profile.AccountInfo{UserID: "123"}, "test_event", nil)
	LogPostHogInfo(context.Background(), "test_log", nil)

	if err := ClosePostHog(); err != nil {
		t.Fatalf("ClosePostHog failed: %v", err)
	}
	if err := ClosePostHogLogs(); err != nil {
		t.Fatalf("ClosePostHogLogs failed: %v", err)
	}
}

func TestPostHog_TestIsolation(t *testing.T) {
	// In test execution, isRunningInTest() returns true.
	// Unless POSTHOG_PROJECT_TOKEN and POSTHOG_HOST are set explicitly,
	// InitPostHog and InitPostHogLogs should not create clients to avoid remote spam.
	_ = os.Unsetenv("POSTHOG_PROJECT_TOKEN")
	_ = os.Unsetenv("POSTHOG_HOST")
	_ = os.Unsetenv("AIM_TELEMETRY_API_KEY")

	cfg := config.NewDefaultConfig()
	InitPostHog(cfg)
	if PostHogClient() != nil {
		t.Fatalf("expected PostHog client to remain nil in tests without explicit env vars")
	}

	InitPostHogLogs(cfg)
	if posthogLogProvider != nil {
		t.Fatalf("expected PostHog log provider to remain nil in tests without explicit env vars")
	}
}
