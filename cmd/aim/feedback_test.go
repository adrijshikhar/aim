package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/feedback"
	"github.com/aim-cli/aim/internal/profile"
)

func TestFeedbackCmd_FlagSubmissionToEndpoint(t *testing.T) {
	var received feedback.Submission
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	origEndpoint := os.Getenv("AIM_FEEDBACK_ENDPOINT")
	defer func() { _ = os.Setenv("AIM_FEEDBACK_ENDPOINT", origEndpoint) }()
	_ = os.Setenv("AIM_FEEDBACK_ENDPOINT", server.URL)

	tempBase := t.TempDir()
	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	out, errOut := captureOutput(t, func() {
		code := dispatch([]string{"feedback", "--type", "bug", "--message", "Test crash report", "--include-doctor"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected feedback command to exit 0, got %d", code)
		}
	})

	if errOut != "" {
		t.Errorf("expected clean stderr, got:\n%s", errOut)
	}
	if !strings.Contains(out, "submitted") && !strings.Contains(out, "Thank you") {
		t.Errorf("expected confirmation output, got:\n%s", out)
	}

	if received.Category != feedback.CategoryBug {
		t.Errorf("expected category bug, got %s", received.Category)
	}
	if received.Message != "Test crash report" {
		t.Errorf("expected message 'Test crash report', got %q", received.Message)
	}
	if !received.IncludeDoctor || !strings.Contains(received.DoctorReport, "# AIM Diagnostic Report") {
		t.Errorf("expected doctor report attached, got:\n%s", received.DoctorReport)
	}
}

func TestFeedbackCmd_FallbackToGitHubURLWhenNoEndpoint(t *testing.T) {
	origEndpoint := os.Getenv("AIM_FEEDBACK_ENDPOINT")
	defer func() { _ = os.Setenv("AIM_FEEDBACK_ENDPOINT", origEndpoint) }()
	_ = os.Setenv("AIM_FEEDBACK_ENDPOINT", "")

	tempBase := t.TempDir()
	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	out, _ := captureOutput(t, func() {
		code := dispatch([]string{"feedback", "--type", "feature", "--message", "Add fish shell completions"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected feedback command to exit 0, got %d", code)
		}
	})

	if !strings.Contains(out, "github.com/adrijshikhar/aim/issues/new") {
		t.Errorf("expected GitHub fallback issue URL in output, got:\n%s", out)
	}
}

func TestFeedbackCmd_RespectsTelemetryDisabledAndSkipsDuplicatePostHogOnCustomEndpoint(t *testing.T) {
	var posthogHits atomic.Int32
	var bodyMu sync.Mutex
	var bodies []string
	posthogSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posthogHits.Add(1)
		data, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		bodyMu.Lock()
		bodies = append(bodies, string(data))
		bodyMu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer posthogSrv.Close()

	var customHits atomic.Int32
	customSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		customHits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer customSrv.Close()

	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)
	t.Setenv("POSTHOG_HOST", posthogSrv.URL)
	t.Setenv("POSTHOG_PROJECT_TOKEN", "phc_test_token")
	t.Setenv("AIM_TELEMETRY_API_KEY", "phc_test_token")
	t.Setenv("AIM_FEEDBACK_ENDPOINT", "")

	cfg := config.NewDefaultConfig()
	cfg.TelemetryDisabled = true
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()
	if code := dispatch([]string{"feedback", "--type", "feature", "--message", "No telemetry"}, reg, pm); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if posthogHits.Load() != 0 {
		t.Fatalf("expected 0 PostHog requests when telemetry_disabled=true, got %d", posthogHits.Load())
	}

	cfg.TelemetryDisabled = false
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AIM_FEEDBACK_ENDPOINT", customSrv.URL)
	if code := dispatch([]string{"feedback", "--type", "bug", "--message", "Custom endpoint only"}, reg, pm); code != 0 {
		t.Fatalf("expected exit 0 for custom endpoint, got %d", code)
	}
	if customHits.Load() != 1 {
		t.Fatalf("expected 1 custom endpoint request, got %d", customHits.Load())
	}
	bodyMu.Lock()
	defer bodyMu.Unlock()
	for _, b := range bodies {
		if strings.Contains(b, "feedback_submitted") {
			t.Fatalf("expected no duplicate feedback_submitted telemetry when custom endpoint succeeded, got body: %s", b)
		}
	}
}
