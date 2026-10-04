package telemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/config"
)

func TestSpooler_AppendAndFlush(t *testing.T) {
	tempDir := t.TempDir()
	spoolFile := filepath.Join(tempDir, "telemetry_spool.json")

	var mu sync.Mutex
	var received []Event

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var batch []Event
		_ = json.NewDecoder(r.Body).Decode(&batch)
		mu.Lock()
		received = append(received, batch...)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	spooler := NewSpooler(spoolFile, server.URL, "", server.Client())

	ev1 := Event{EventName: "cmd1", DistinctID: "machine-1"}
	ev2 := Event{EventName: "cmd2", DistinctID: "machine-1"}

	if err := spooler.Append(ev1); err != nil {
		t.Fatalf("failed to append ev1: %v", err)
	}
	if err := spooler.Append(ev2); err != nil {
		t.Fatalf("failed to append ev2: %v", err)
	}

	// Verify events are in spool file
	events, err := spooler.Read()
	if err != nil {
		t.Fatalf("failed to read spool: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events in spool, got %d", len(events))
	}

	// Flush
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := spooler.Flush(ctx); err != nil {
		t.Fatalf("expected successful flush, got: %v", err)
	}

	// After flush, spool should be empty
	eventsAfter, _ := spooler.Read()
	if len(eventsAfter) != 0 {
		t.Errorf("expected 0 events after flush, got %d", len(eventsAfter))
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Errorf("expected server to receive 2 events, got %d", len(received))
	}
}

func TestSpooler_AppendAndFlush_WithPostHogBatchFormat(t *testing.T) {
	tempDir := t.TempDir()
	spoolFile := filepath.Join(tempDir, "telemetry_spool.json")

	var mu sync.Mutex
	var receivedBatch BatchPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload BatchPayload
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		receivedBatch = payload
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testAPIKey := "phc_test_key_123"
	spooler := NewSpooler(spoolFile, server.URL, testAPIKey, server.Client())

	ev1 := Event{EventName: "posthog_test_cmd", DistinctID: "machine-ph"}
	if err := spooler.Append(ev1); err != nil {
		t.Fatalf("failed to append ev1: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := spooler.Flush(ctx); err != nil {
		t.Fatalf("expected successful flush, got: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if receivedBatch.APIKey != testAPIKey {
		t.Errorf("expected API key %q, got %q", testAPIKey, receivedBatch.APIKey)
	}
	if len(receivedBatch.Batch) != 1 || receivedBatch.Batch[0].EventName != "posthog_test_cmd" {
		t.Errorf("expected 1 event in batch, got %+v", receivedBatch.Batch)
	}
}

func TestTelemetryClient_OptOutRespect(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.NewDefaultConfig()
	cfg.TelemetryDisabled = true

	client := NewClient(tempDir, tempDir, "0.12.1", cfg)
	client.Track("test_event", map[string]any{"key": "val"})

	// Spool file should not even be created
	spoolFile := filepath.Join(tempDir, "telemetry_spool.json")
	if _, err := os.Stat(spoolFile); err == nil {
		t.Errorf("expected no spool file when telemetry disabled")
	}
}

func TestTelemetryClient_ZeroLatencyOnTimeout(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.NewDefaultConfig()

	// Slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(tempDir, tempDir, "0.12.1", cfg)
	client.spooler.Endpoint = server.URL

	// Bounded flush should not hang
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = client.Flush(ctx)
	duration := time.Since(start)

	if duration > 300*time.Millisecond {
		t.Errorf("flush took %v, expected <= 300ms", duration)
	}
}
