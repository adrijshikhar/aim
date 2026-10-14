package usage_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/usage"
)

func TestInspection_StatusConstants(t *testing.T) {
	if usage.StatusIdle != "IDLE" {
		t.Fatalf("expected StatusIdle == 'IDLE', got %q", usage.StatusIdle)
	}
	if usage.StatusTimeout != "TIMEOUT" {
		t.Fatalf("expected StatusTimeout == 'TIMEOUT', got %q", usage.StatusTimeout)
	}
	if usage.StatusOffline != "OFFLINE" {
		t.Fatalf("expected StatusOffline == 'OFFLINE', got %q", usage.StatusOffline)
	}
}

func TestInspection_ZeroFileModifications(t *testing.T) {
	tmpDir := t.TempDir()
	profDir := filepath.Join(tmpDir, "profiles", "codex-test")
	codexDir := filepath.Join(profDir, ".codex")
	if err := os.MkdirAll(codexDir, 0700); err != nil {
		t.Fatalf("failed to create codex dir: %v", err)
	}

	credPath := filepath.Join(codexDir, "auth.json")
	credContent := `{
  "tokens": {
    "access_token": "expired-access-token-12345",
    "refresh_token": "live-refresh-token-67890"
  }
}`
	if err := os.WriteFile(credPath, []byte(credContent), 0600); err != nil {
		t.Fatalf("failed to write credential file: %v", err)
	}

	// Set ModTime in the past to ensure any accidental touch/write changes it
	pastTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(credPath, pastTime, pastTime); err != nil {
		t.Fatalf("failed to set mtime: %v", err)
	}

	initialStat, err := os.Stat(credPath)
	if err != nil {
		t.Fatalf("failed to stat credential file: %v", err)
	}

	cacheStore := usage.NewCacheStore(tmpDir, usage.DefaultTTL)
	networkCalled := false

	target := usage.TargetProfile{
		Agent:      "codex",
		Profile:    "codex-test",
		ProfileDir: profDir,
		GetUsageFn: func(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
			networkCalled = true
			return nil, errors.New("network should not be called during inspection")
		},
	}

	report := usage.InspectProfile(context.Background(), target, cacheStore)

	finalStat, err := os.Stat(credPath)
	if err != nil {
		t.Fatalf("failed to stat credential file after inspection: %v", err)
	}

	if !finalStat.ModTime().Equal(initialStat.ModTime()) {
		t.Errorf("file modification time changed: initial=%v, final=%v", initialStat.ModTime(), finalStat.ModTime())
	}
	if finalStat.Size() != initialStat.Size() {
		t.Errorf("file size changed: initial=%d, final=%d", initialStat.Size(), finalStat.Size())
	}
	if networkCalled {
		t.Errorf("network call was made via GetUsageFn, inspection must strictly make zero network calls")
	}
	if report.Status != usage.StatusIdle {
		t.Errorf("expected report.Status == %q, got %q", usage.StatusIdle, report.Status)
	}
	if report.Summary != "Ready on use" {
		t.Errorf("expected report.Summary == %q, got %q", "Ready on use", report.Summary)
	}
}

func TestInspection_RefusalPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	profDir := filepath.Join(tmpDir, "profiles", "refused-prof")
	codexDir := filepath.Join(profDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)
	_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(`{"tokens":{"access_token":"tok","refresh_token":"ref"}}`), 0600)

	cacheStore := usage.NewCacheStore(tmpDir, usage.DefaultTTL)
	refStore := cacheStore.RefusalStore()
	if refStore == nil {
		t.Fatal("expected non-nil RefusalStore")
	}

	// Record active refusal
	if err := refStore.Record("codex", "refused-prof", oauth.RefusalRefreshTokenExpired, "refresh token is expired"); err != nil {
		t.Fatalf("failed to record refusal: %v", err)
	}

	// Also seed fresh cache to prove refusal takes precedence over fresh cache
	_ = cacheStore.Put(usage.Report{
		Agent:     "codex",
		Profile:   "refused-prof",
		Status:    usage.StatusOK,
		Summary:   "5h: 100%",
		FetchedAt: time.Now(),
	})

	networkCalled := false
	target := usage.TargetProfile{
		Agent:      "codex",
		Profile:    "refused-prof",
		ProfileDir: profDir,
		GetUsageFn: func(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
			networkCalled = true
			return nil, errors.New("network should not be called")
		},
	}

	report := usage.InspectProfile(context.Background(), target, cacheStore)

	if networkCalled {
		t.Errorf("network call was made during inspection")
	}
	if report.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown for refused profile, got %q", report.Status)
	}
	expectedErr := "re-login required (" + string(oauth.RefusalRefreshTokenExpired) + ")"
	if report.Error != expectedErr {
		t.Errorf("expected report.Error == %q, got %q", expectedErr, report.Error)
	}
}

func TestInspection_MissingCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	emptyProfDir := filepath.Join(tmpDir, "profiles", "empty-prof")
	_ = os.MkdirAll(emptyProfDir, 0700)

	cacheStore := usage.NewCacheStore(tmpDir, usage.DefaultTTL)
	networkCalled := false

	t.Run("empty profile directory", func(t *testing.T) {
		target := usage.TargetProfile{
			Agent:      "codex",
			Profile:    "empty-prof",
			ProfileDir: emptyProfDir,
			GetUsageFn: func(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
				networkCalled = true
				return nil, errors.New("network should not be called")
			},
		}

		report := usage.InspectProfile(context.Background(), target, cacheStore)

		if networkCalled {
			t.Errorf("network call was made during inspection")
		}
		if report.Status != usage.StatusUnknown {
			t.Errorf("expected StatusUnknown, got %q", report.Status)
		}
		if report.Error != "no credentials" {
			t.Errorf("expected Error == %q, got %q", "no credentials", report.Error)
		}
	})

	t.Run("non-existent profile directory", func(t *testing.T) {
		target := usage.TargetProfile{
			Agent:      "claude",
			Profile:    "nonexistent",
			ProfileDir: filepath.Join(tmpDir, "does-not-exist"),
		}

		report := usage.InspectProfile(context.Background(), target, cacheStore)

		if report.Status != usage.StatusUnknown {
			t.Errorf("expected StatusUnknown, got %q", report.Status)
		}
		if report.Error != "no credentials" {
			t.Errorf("expected Error == %q, got %q", "no credentials", report.Error)
		}
	})
}

func TestInspection_FreshCachePassthrough(t *testing.T) {
	tmpDir := t.TempDir()
	cacheStore := usage.NewCacheStore(tmpDir, usage.DefaultTTL)

	resetTime := time.Now().Add(3 * time.Hour)
	freshReport := usage.Report{
		Agent:   "codex",
		Profile: "active-prof",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Name: "5-Hour", RemainingPct: 85, ResetsAt: resetTime},
		},
		Summary:   "5h: 85%",
		FetchedAt: time.Now(),
	}
	if err := cacheStore.Put(freshReport); err != nil {
		t.Fatalf("failed to put cache: %v", err)
	}

	networkCalled := false
	target := usage.TargetProfile{
		Agent:   "codex",
		Profile: "active-prof",
		GetUsageFn: func(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
			networkCalled = true
			return nil, errors.New("network should not be called")
		},
	}

	report := usage.InspectProfile(context.Background(), target, cacheStore)

	if networkCalled {
		t.Errorf("network call was made during inspection")
	}
	if report.Status != usage.StatusOK {
		t.Errorf("expected StatusOK from fresh cache, got %q", report.Status)
	}
	if !report.FromCache {
		t.Errorf("expected FromCache == true")
	}
	if report.Summary != "5h: 85%" {
		t.Errorf("expected Summary == %q, got %q", "5h: 85%", report.Summary)
	}
	if len(report.Windows) != 1 || report.Windows[0].RemainingPct != 85 {
		t.Errorf("expected cached window to be preserved: %+v", report.Windows)
	}
}

func TestInspection_StaleCacheWithPriorStatusOK(t *testing.T) {
	tmpDir := t.TempDir()
	cacheStore := usage.NewCacheStore(tmpDir, 5*time.Minute)

	resetTime := time.Now().Add(3 * time.Hour)
	staleReport := usage.Report{
		Agent:   "codex",
		Profile: "stale-prof",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Name: "5-Hour", RemainingPct: 92, ResetsAt: resetTime},
		},
		Summary:   "5h: 92%",
		FetchedAt: time.Now().Add(-30 * time.Minute), // Older than 5 min TTL
	}
	if err := cacheStore.Put(staleReport); err != nil {
		t.Fatalf("failed to put cache: %v", err)
	}

	networkCalled := false
	target := usage.TargetProfile{
		Agent:   "codex",
		Profile: "stale-prof",
		GetUsageFn: func(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
			networkCalled = true
			return nil, errors.New("network should not be called")
		},
	}

	report := usage.InspectProfile(context.Background(), target, cacheStore)

	if networkCalled {
		t.Errorf("network call was made during inspection")
	}
	if report.Status != usage.StatusIdle {
		t.Errorf("expected StatusIdle for stale cache, got %q", report.Status)
	}
	if !report.FromCache {
		t.Errorf("expected FromCache == true")
	}
	if report.Summary != "5h: 92%" {
		t.Errorf("expected preserved Summary == %q, got %q", "5h: 92%", report.Summary)
	}
	if len(report.Windows) != 1 || report.Windows[0].RemainingPct != 92 {
		t.Errorf("expected cached window preserved: %+v", report.Windows)
	}

	// Test case where stale report had empty summary
	staleNoSummary := usage.Report{
		Agent:   "codex",
		Profile: "stale-no-summary",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Name: "5-Hour", RemainingPct: 80},
		},
		Summary:   "",
		FetchedAt: time.Now().Add(-30 * time.Minute),
	}
	_ = cacheStore.Put(staleNoSummary)
	repNoSum := usage.InspectProfile(context.Background(), usage.TargetProfile{
		Agent:   "codex",
		Profile: "stale-no-summary",
	}, cacheStore)
	if repNoSum.Status != usage.StatusIdle {
		t.Errorf("expected StatusIdle, got %q", repNoSum.Status)
	}
	if repNoSum.Summary != "Ready on use" {
		t.Errorf("expected fallback summary 'Ready on use', got %q", repNoSum.Summary)
	}
}

func TestInspection_ZeroByteAndCorruptCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	cacheStore := usage.NewCacheStore(tmpDir, usage.DefaultTTL)

	t.Run("zero-byte credential file", func(t *testing.T) {
		profDir := filepath.Join(tmpDir, "profiles", "zero-byte-prof")
		codexDir := filepath.Join(profDir, ".codex")
		_ = os.MkdirAll(codexDir, 0700)
		_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte{}, 0600)

		target := usage.TargetProfile{
			Agent:      "codex",
			Profile:    "zero-byte-prof",
			ProfileDir: profDir,
		}

		report := usage.InspectProfile(context.Background(), target, cacheStore)
		if report.Status != usage.StatusUnknown {
			t.Errorf("expected StatusUnknown for zero-byte credentials, got %q", report.Status)
		}
		if report.Error != "no credentials" {
			t.Errorf("expected Error == 'no credentials', got %q", report.Error)
		}
	})

	t.Run("corrupt JSON syntax", func(t *testing.T) {
		profDir := filepath.Join(tmpDir, "profiles", "corrupt-json-prof")
		claudeDir := filepath.Join(profDir, ".claude")
		_ = os.MkdirAll(claudeDir, 0700)
		_ = os.WriteFile(filepath.Join(claudeDir, "auth.json"), []byte(`{invalid-json`), 0600)

		target := usage.TargetProfile{
			Agent:      "claude",
			Profile:    "corrupt-json-prof",
			ProfileDir: profDir,
		}

		report := usage.InspectProfile(context.Background(), target, cacheStore)
		if report.Status != usage.StatusUnknown {
			t.Errorf("expected StatusUnknown for corrupt JSON, got %q", report.Status)
		}
		if report.Error != "no credentials" {
			t.Errorf("expected Error == 'no credentials', got %q", report.Error)
		}
	})

	t.Run("empty JSON object", func(t *testing.T) {
		profDir := filepath.Join(tmpDir, "profiles", "empty-json-prof")
		_ = os.MkdirAll(profDir, 0700)
		_ = os.WriteFile(filepath.Join(profDir, "auth.json"), []byte(`{}`), 0600)

		target := usage.TargetProfile{
			Agent:      "codex",
			Profile:    "empty-json-prof",
			ProfileDir: profDir,
		}

		report := usage.InspectProfile(context.Background(), target, cacheStore)
		if report.Status != usage.StatusUnknown {
			t.Errorf("expected StatusUnknown for empty JSON object, got %q", report.Status)
		}
		if report.Error != "no credentials" {
			t.Errorf("expected Error == 'no credentials', got %q", report.Error)
		}
	})

	t.Run("irrelevant JSON fields without credentials", func(t *testing.T) {
		profDir := filepath.Join(tmpDir, "profiles", "irrelevant-json-prof")
		_ = os.MkdirAll(profDir, 0700)
		_ = os.WriteFile(filepath.Join(profDir, "auth.json"), []byte(`{"version": 1, "name": "foo"}`), 0600)

		target := usage.TargetProfile{
			Agent:      "codex",
			Profile:    "irrelevant-json-prof",
			ProfileDir: profDir,
		}

		report := usage.InspectProfile(context.Background(), target, cacheStore)
		if report.Status != usage.StatusUnknown {
			t.Errorf("expected StatusUnknown for JSON without tokens, got %q", report.Status)
		}
		if report.Error != "no credentials" {
			t.Errorf("expected Error == 'no credentials', got %q", report.Error)
		}
	})
}

func TestInspection_NilCacheStore(t *testing.T) {
	tmpDir := t.TempDir()
	profDir := filepath.Join(tmpDir, "profiles", "valid-prof")
	codexDir := filepath.Join(profDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)
	_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(`{"tokens":{"access_token":"ok","refresh_token":"ref"}}`), 0600)

	target := usage.TargetProfile{
		Agent:      "codex",
		Profile:    "valid-prof",
		ProfileDir: profDir,
	}

	report := usage.InspectProfile(context.Background(), target, nil)
	if report.Status != usage.StatusIdle {
		t.Errorf("expected StatusIdle with nil cache, got %q", report.Status)
	}
	if report.Summary != "Ready on use" {
		t.Errorf("expected Summary == 'Ready on use', got %q", report.Summary)
	}
}

func TestInspection_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	target := usage.TargetProfile{
		Agent:   "codex",
		Profile: "any-prof",
	}

	report := usage.InspectProfile(ctx, target, nil)
	if report.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown when context is cancelled, got %q", report.Status)
	}
	if report.Error != context.Canceled.Error() {
		t.Errorf("expected Error == %q, got %q", context.Canceled.Error(), report.Error)
	}
}

func TestInspection_ExcessiveFileSizeCheck(t *testing.T) {
	tmpDir := t.TempDir()
	profDir := filepath.Join(tmpDir, "profiles", "huge-prof")
	codexDir := filepath.Join(profDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)

	// Create file larger than 1MB (1<<20 bytes)
	hugeFile := filepath.Join(codexDir, "auth.json")
	f, err := os.Create(hugeFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate((1 << 20) + 1024); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	target := usage.TargetProfile{
		Agent:      "codex",
		Profile:    "huge-prof",
		ProfileDir: profDir,
	}

	report := usage.InspectProfile(context.Background(), target, nil)
	if report.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown for >1MB file, got %q", report.Status)
	}
	if report.Error != "no credentials" {
		t.Errorf("expected 'no credentials' error for huge file, got %q", report.Error)
	}
}

func TestInspection_StaleCacheWithNonStatusOKAndWindows(t *testing.T) {
	tmpDir := t.TempDir()
	cacheStore := usage.NewCacheStore(tmpDir, 5*time.Minute)

	staleWarning := usage.Report{
		Agent:   "codex",
		Profile: "stale-warning",
		Status:  usage.StatusWarning,
		Windows: []usage.LimitWindow{
			{Name: "5-Hour", RemainingPct: 15},
		},
		Summary:   "5h: 15%",
		FetchedAt: time.Now().Add(-30 * time.Minute),
	}
	_ = cacheStore.Put(staleWarning)

	target := usage.TargetProfile{
		Agent:   "codex",
		Profile: "stale-warning",
	}

	report := usage.InspectProfile(context.Background(), target, cacheStore)
	if report.Status != usage.StatusIdle {
		t.Errorf("expected StatusIdle for stale non-StatusOK with windows, got %q", report.Status)
	}
	if report.Summary != "5h: 15%" {
		t.Errorf("expected summary preserved, got %q", report.Summary)
	}
	if len(report.Windows) != 1 || report.Windows[0].RemainingPct != 15 {
		t.Errorf("expected windows preserved, got %+v", report.Windows)
	}
}

func TestInspection_ClaudeAiOauthStructure(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("metadata only without tokens returns no credentials", func(t *testing.T) {
		profDir := filepath.Join(tmpDir, "profiles", "meta-only")
		claudeDir := filepath.Join(profDir, ".claude")
		_ = os.MkdirAll(claudeDir, 0700)
		_ = os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(`{"claudeAiOauth":{"subscriptionType":"team"}}`), 0600)

		target := usage.TargetProfile{
			Agent:      "claude",
			Profile:    "meta-only",
			ProfileDir: profDir,
		}

		report := usage.InspectProfile(context.Background(), target, nil)
		if report.Status != usage.StatusUnknown {
			t.Errorf("expected StatusUnknown for metadata-only claudeAiOauth, got %q", report.Status)
		}
		if report.Error != "no credentials" {
			t.Errorf("expected Error == 'no credentials', got %q", report.Error)
		}
	})

	t.Run("valid token in claudeAiOauth returns StatusIdle", func(t *testing.T) {
		profDir := filepath.Join(tmpDir, "profiles", "valid-claude")
		claudeDir := filepath.Join(profDir, ".claude")
		_ = os.MkdirAll(claudeDir, 0700)
		_ = os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(`{"claudeAiOauth":{"accessToken":"sk-ant-123","subscriptionType":"team"}}`), 0600)

		target := usage.TargetProfile{
			Agent:      "claude",
			Profile:    "valid-claude",
			ProfileDir: profDir,
		}

		report := usage.InspectProfile(context.Background(), target, nil)
		if report.Status != usage.StatusIdle {
			t.Errorf("expected StatusIdle for valid claudeAiOauth, got %q", report.Status)
		}
		if report.Summary != "Ready on use" {
			t.Errorf("expected Summary == 'Ready on use', got %q", report.Summary)
		}
	})
}
