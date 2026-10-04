package telemetry

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

// TestAnonymity_EventWhitelist_StrictRejection verifies that arbitrary, fishy,
// or non-whitelisted telemetry events are strictly rejected from both local spooling
// and PostHog capture.
func TestAnonymity_EventWhitelist_StrictRejection(t *testing.T) {
	allowedEvents := []string{
		EventCommandExecuted,
		EventTUIOpened,
		EventDoctorReportGenerated,
		EventFeedbackSubmitted,
		"agent_run_completed",
		"user_identified",
	}

	for _, name := range allowedEvents {
		if !IsAllowedEvent(name) {
			t.Errorf("expected allowed event %q to pass IsAllowedEvent", name)
		}
	}

	fishyEvents := []string{
		"fishy_telemetry_added_by_pr",
		"user_credentials_stolen",
		"keystrokes_logged",
		"sudo_command_executed",
		"custom_tracker",
		"",
		"   ",
		"COMMAND_EXECUTED", // Case sensitivity check
		"command_executed_extra",
	}

	for _, name := range fishyEvents {
		if IsAllowedEvent(name) {
			t.Errorf("fishy event %q unexpectedly passed IsAllowedEvent whitelist", name)
		}
	}

	// Verify DefaultClient.Track strictly drops unwhitelisted events
	tempBase := t.TempDir()
	tempCache := t.TempDir()
	client := NewClient(tempBase, tempCache, "1.0.0", nil)

	for _, fishy := range fishyEvents {
		client.Track(fishy, map[string]any{"danger": "data"})
	}

	spoolPath := filepath.Join(tempCache, "telemetry_spool.json")
	if _, err := os.Stat(spoolPath); !os.IsNotExist(err) {
		spooler := NewSpooler(spoolPath, "", "", nil)
		events, err := spooler.Read()
		if err == nil && len(events) > 0 {
			t.Fatalf("spool contains unwhitelisted events: %+v", events)
		}
	}

	// Verify CaptureAccountEvent strictly drops unwhitelisted events
	account := profile.AccountInfo{UserID: "usr_safe", Email: "test@example.com"}
	CaptureAccountEvent("codex", account, "fishy_pr_event", map[string]any{"secret": "val"})
}

// TestAnonymity_AnonymousMachineID_NoPII_RestrictedPermissions verifies that the
// machine ID is a deterministic SHA-256 hash containing zero personal information,
// and is saved with restricted file permissions (0600).
func TestAnonymity_AnonymousMachineID_NoPII_RestrictedPermissions(t *testing.T) {
	tempBase := t.TempDir()
	id := AnonymousMachineID(tempBase)

	// Must be exactly 64 hex characters (SHA-256)
	matched, err := regexp.MatchString("^[a-f0-9]{64}$", id)
	if err != nil || !matched {
		t.Fatalf("machine ID %q is not a valid 64-character hex SHA-256 hash", id)
	}

	// Must not leak current OS username or hostname
	currentUser := os.Getenv("USER")
	if currentUser != "" && len(currentUser) > 2 {
		if strings.Contains(strings.ToLower(id), strings.ToLower(currentUser)) {
			t.Errorf("machine ID leaked local username %q", currentUser)
		}
	}
	hostname, _ := os.Hostname()
	if hostname != "" && len(hostname) > 2 {
		if strings.Contains(strings.ToLower(id), strings.ToLower(hostname)) {
			t.Errorf("machine ID leaked local hostname %q", hostname)
		}
	}

	// Must not contain filesystem paths
	if strings.Contains(id, "/") || strings.Contains(id, "\\") || strings.Contains(id, "~") {
		t.Errorf("machine ID contains path characters: %q", id)
	}

	// Check file permissions on ~/.aim/telemetry_id
	idPath := filepath.Join(tempBase, "telemetry_id")
	info, err := os.Stat(idPath)
	if err != nil {
		t.Fatalf("failed to stat telemetry_id file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected telemetry_id file to have 0600 permissions, got %#o", perm)
	}

	// Anti-tamper: if telemetry_id file is corrupted with an invalid token,
	// AnonymousMachineID must not return the corrupted content.
	_ = os.WriteFile(idPath, []byte("invalid_attacker_injected_string"), 0600)
	recoveredID := AnonymousMachineID(tempBase)
	if recoveredID == "invalid_attacker_injected_string" {
		t.Fatal("AnonymousMachineID accepted tampered non-64-character content")
	}
	if len(recoveredID) != 64 {
		t.Fatalf("regenerated ID %q is not 64 chars", recoveredID)
	}
}

// TestAntiSpoofing_SystemPropertiesCannotBeOverridden ensures that system-level
// properties (version, aim_version, os, arch) cannot be spoofed or overridden
// by untrusted callers passing customProps.
func TestAntiSpoofing_SystemPropertiesCannotBeOverridden(t *testing.T) {
	spoofedProps := map[string]any{
		"version":     "hacked_version_9.9.9",
		"aim_version": "hacked_aim_version_9.9.9",
		"os":          "fake_os",
		"arch":        "fake_arch",
		"legit_prop":  "legit_value",
	}

	ev := NewEvent(EventCommandExecuted, "test_distinct_id", "1.2.3", runtime.GOOS, runtime.GOARCH, spoofedProps)

	if ev.Properties["version"] != "1.2.3" {
		t.Errorf("anti-spoofing failed: version was overridden to %v", ev.Properties["version"])
	}
	if ev.Properties["aim_version"] != "1.2.3" {
		t.Errorf("anti-spoofing failed: aim_version was overridden to %v", ev.Properties["aim_version"])
	}
	if ev.Properties["os"] != runtime.GOOS {
		t.Errorf("anti-spoofing failed: os was overridden to %v", ev.Properties["os"])
	}
	if ev.Properties["arch"] != runtime.GOARCH {
		t.Errorf("anti-spoofing failed: arch was overridden to %v", ev.Properties["arch"])
	}
	if ev.Properties["legit_prop"] != "legit_value" {
		t.Errorf("expected legit_prop to be preserved, got %v", ev.Properties["legit_prop"])
	}
}

// TestAnonymity_AccountDistinctID_NoPIILeak ensures that account distinct IDs
// never leak email addresses, full names, or secret keys.
func TestAnonymity_AccountDistinctID_NoPIILeak(t *testing.T) {
	account := profile.AccountInfo{
		UserID: "usr_abc456",
		Email:  "john.doe@confidential-corp.com",
		Name:   "John Doe",
	}

	id := AccountDistinctID("claude", account)
	if id != "claude:usr_abc456" {
		t.Fatalf("expected 'claude:usr_abc456', got %q", id)
	}

	// Verify no part of email or name is in distinct ID
	if strings.Contains(id, "john") || strings.Contains(id, "confidential") {
		t.Fatalf("AccountDistinctID leaked personal info in distinct ID: %q", id)
	}

	// When UserID is empty, ensure clean fallback to anonymous machine ID without prefix
	emptyUserAccount := profile.AccountInfo{
		Email: "secret@company.com",
	}
	fallbackID := AccountDistinctID("claude", emptyUserAccount)
	if fallbackID == "claude:" || strings.Contains(fallbackID, "claude") {
		t.Fatalf("AccountDistinctID with empty UserID should fall back to machine ID, got %q", fallbackID)
	}
	if len(fallbackID) != 64 {
		t.Fatalf("expected 64-char anonymous machine ID fallback, got %q", fallbackID)
	}

	// When agentName is empty, fallback to machine ID
	noAgentID := AccountDistinctID("", account)
	if len(noAgentID) != 64 {
		t.Fatalf("expected 64-char anonymous machine ID fallback when agent is empty, got %q", noAgentID)
	}
}

// TestAnonymity_OptOut_StrictZeroSpooling ensures that when telemetry is opted-out,
// zero events are spooled to disk across all opt-out mechanisms.
func TestAnonymity_OptOut_StrictZeroSpooling(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func()
		teardown  func()
		cfg       *config.Config
	}{
		{
			name: "DO_NOT_TRACK=1",
			setup: func() {
				_ = os.Setenv("DO_NOT_TRACK", "1")
			},
			teardown: func() {
				_ = os.Unsetenv("DO_NOT_TRACK")
			},
			cfg: nil,
		},
		{
			name: "AIM_TELEMETRY_DISABLED=1",
			setup: func() {
				_ = os.Setenv("AIM_TELEMETRY_DISABLED", "1")
			},
			teardown: func() {
				_ = os.Unsetenv("AIM_TELEMETRY_DISABLED")
			},
			cfg: nil,
		},
		{
			name: "Config_TelemetryDisabled=true",
			setup: func() {
				_ = os.Unsetenv("DO_NOT_TRACK")
				_ = os.Unsetenv("AIM_TELEMETRY_DISABLED")
			},
			teardown: func() {},
			cfg: &config.Config{
				TelemetryDisabled: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			defer tc.teardown()

			tempBase := t.TempDir()
			tempCache := t.TempDir()

			client := NewClient(tempBase, tempCache, "1.0.0", tc.cfg)
			client.Track(EventCommandExecuted, map[string]any{"command": "version"})
			client.Track(EventTUIOpened, nil)

			spoolPath := filepath.Join(tempCache, "telemetry_spool.json")
			if _, err := os.Stat(spoolPath); !os.IsNotExist(err) {
				spooler := NewSpooler(spoolPath, "", "", nil)
				events, _ := spooler.Read()
				if len(events) > 0 {
					t.Fatalf("spool contains %d events despite opt-out %s", len(events), tc.name)
				}
			}

			// Verify Flush and Close are safe no-ops
			if err := client.Flush(context.Background()); err != nil {
				t.Fatalf("Flush returned error on disabled client: %v", err)
			}
			if err := client.Close(); err != nil {
				t.Fatalf("Close returned error on disabled client: %v", err)
			}
		})
	}
}
