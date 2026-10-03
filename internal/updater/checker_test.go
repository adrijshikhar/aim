package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"v0.13.0", "v0.12.1", true},
		{"0.13.0", "0.12.1", true},
		{"v1.0.0", "0.12.1", true},
		{"v0.12.2", "0.12.1", true},
		{"v0.12.1", "0.12.1", false},
		{"v0.12.0", "0.12.1", false},
		{"v0.11.9", "0.12.1", false},
		{"v0.13.0", "dev", false}, // dev builds do not report update available
		{"invalid", "v0.12.1", false},
	}

	for _, tt := range tests {
		got := IsNewerVersion(tt.latest, tt.current)
		if got != tt.want {
			t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestCheckForUpdate_FreshFetchAndCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"tag_name": "v0.13.0",
			"html_url": "https://github.com/adrijshikhar/aim/releases/tag/v0.13.0",
		})
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	checker := &Checker{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	info, err := checker.Check(ctx, "0.12.1", cacheDir, false)
	if err != nil {
		t.Fatalf("expected successful update check, got: %v", err)
	}

	if !info.UpdateAvailable {
		t.Errorf("expected update to be available, got false")
	}
	if info.LatestVersion != "v0.13.0" {
		t.Errorf("expected latest version v0.13.0, got %s", info.LatestVersion)
	}

	// Verify cache file was written
	cacheFile := filepath.Join(cacheDir, "update_check.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}
	var cached CacheData
	if err := json.Unmarshal(data, &cached); err != nil {
		t.Fatalf("failed to parse cache file: %v", err)
	}
	if cached.LatestVersion != "v0.13.0" {
		t.Errorf("expected cached latest version v0.13.0, got %s", cached.LatestVersion)
	}
}

func TestCheckForUpdate_UsesCachedWithin24Hours(t *testing.T) {
	cacheHitCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheHitCount++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"tag_name": "v0.14.0",
			"html_url": "https://github.com/adrijshikhar/aim/releases/tag/v0.14.0",
		})
	}))
	defer server.Close()

	cacheDir := t.TempDir()

	// Pre-populate cache with a check from 1 hour ago
	preCache := CacheData{
		LatestVersion: "v0.13.0",
		ReleaseURL:    "https://github.com/adrijshikhar/aim/releases/tag/v0.13.0",
		CheckedAt:     time.Now().Add(-1 * time.Hour),
	}
	data, _ := json.Marshal(preCache)
	_ = os.WriteFile(filepath.Join(cacheDir, "update_check.json"), data, 0644)

	checker := &Checker{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}

	info, err := checker.Check(context.Background(), "0.12.1", cacheDir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Server should NOT have been hit
	if cacheHitCount != 0 {
		t.Errorf("expected 0 HTTP requests due to 24h cache hit, got %d", cacheHitCount)
	}
	if info.LatestVersion != "v0.13.0" {
		t.Errorf("expected latest version from cache (v0.13.0), got %s", info.LatestVersion)
	}
	if !info.UpdateAvailable {
		t.Errorf("expected update to be available from cached comparison")
	}
}

func TestCheckForUpdate_GracefulDegradeOnNetworkFailure(t *testing.T) {
	cacheDir := t.TempDir()
	checker := &Checker{
		Endpoint:   "http://127.0.0.1:59999/unreachable",
		HTTPClient: &http.Client{Timeout: 100 * time.Millisecond},
	}

	// Should not panic, should return gracefully
	info, err := checker.Check(context.Background(), "0.12.1", cacheDir, false)
	if err != nil {
		t.Fatalf("expected nil error on network failure (graceful degradation), got: %v", err)
	}
	if info.UpdateAvailable {
		t.Errorf("expected UpdateAvailable to be false on network failure")
	}
}
