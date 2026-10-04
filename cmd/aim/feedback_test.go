package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
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
