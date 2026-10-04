package feedback

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPDispatcher_SubmitSuccess(t *testing.T) {
	var received Submission
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content-type, got %s", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	d := &HTTPDispatcher{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	sub := Submission{
		Category:      CategoryBug,
		Message:       "Crash occurred on startup",
		IncludeDoctor: true,
		DoctorReport:  "# AIM Diagnostic Report",
		AIMVersion:    "0.12.1",
		OS:            "darwin",
		Arch:          "arm64",
		Timestamp:     time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := d.Submit(ctx, sub); err != nil {
		t.Fatalf("expected Submit to succeed, got: %v", err)
	}

	if received.Category != CategoryBug {
		t.Errorf("expected category bug, got %s", received.Category)
	}
	if received.Message != "Crash occurred on startup" {
		t.Errorf("expected message 'Crash occurred on startup', got %q", received.Message)
	}
	if !received.IncludeDoctor || received.DoctorReport != "# AIM Diagnostic Report" {
		t.Errorf("expected doctor report in payload, got %q", received.DoctorReport)
	}
}

func TestHTTPDispatcher_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	d := &HTTPDispatcher{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	sub := Submission{
		Category: CategoryGeneral,
		Message:  "Great tool!",
	}

	err := d.Submit(context.Background(), sub)
	if err == nil {
		t.Fatal("expected error on 500 response, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status code 500 in error, got: %v", err)
	}
}

func TestHTTPDispatcher_EmptyEndpointGeneratesFallbackIssueURL(t *testing.T) {
	d := &HTTPDispatcher{
		Endpoint: "",
	}

	sub := Submission{
		Category: CategoryFeature,
		Message:  "Please support OpenCode agent",
	}

	fallbackURL := d.FallbackURL(sub)
	if !strings.Contains(fallbackURL, "github.com/adrijshikhar/aim/issues/new") {
		t.Errorf("expected GitHub issues URL, got %s", fallbackURL)
	}
	if !strings.Contains(fallbackURL, "OpenCode") {
		t.Errorf("expected message content in fallback URL, got %s", fallbackURL)
	}
}
