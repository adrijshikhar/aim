package usage_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/usage"
)

func TestCacheStorePutGetAndExpiry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	ttl := 100 * time.Millisecond
	store := usage.NewCacheStore(tmpDir, ttl)

	// Initially empty
	_, found := store.Get("agy", "work")
	if found {
		t.Fatal("expected cache miss on empty store")
	}

	resetsAt := time.Now().Add(2 * time.Hour)
	rep := usage.Report{
		Agent:     "agy",
		Profile:   "work",
		Status:    usage.StatusOK,
		FetchedAt: time.Now(),
		Windows: []usage.LimitWindow{
			{Name: "5-Hour", RemainingPct: 90, ResetsAt: resetsAt},
		},
	}

	if err := store.Put(rep); err != nil {
		t.Fatalf("store.Put failed: %v", err)
	}

	// Should be found and fresh
	cached, found := store.Get("agy", "work")
	if !found {
		t.Fatal("expected cache hit")
	}
	if cached.Status != usage.StatusOK || len(cached.Windows) != 1 {
		t.Errorf("unexpected cached report: %+v", cached)
	}
	if !cached.FromCache {
		t.Errorf("expected FromCache to be true on Get()")
	}
	if cached.Windows[0].ResetsIn <= 0 {
		t.Errorf("expected ResetsIn to be dynamically computed, got %v", cached.Windows[0].ResetsIn)
	}

	// Reload from disk into new store instance to test atomic persistence
	store2 := usage.NewCacheStore(tmpDir, ttl)
	cached2, found2 := store2.Get("agy", "work")
	if !found2 || cached2.Status != usage.StatusOK {
		t.Fatalf("expected persistent cache hit after reload: found=%v, %+v", found2, cached2)
	}

	// Wait for TTL expiration
	time.Sleep(150 * time.Millisecond)
	_, foundExpired := store.Get("agy", "work")
	if foundExpired {
		t.Fatal("expected expired entry to be treated as cache miss")
	}
}

func TestCacheStoreConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-race-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := usage.NewCacheStore(tmpDir, time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			_ = store.Put(usage.Report{
				Agent:     "agy",
				Profile:   "work",
				Status:    usage.StatusOK,
				FetchedAt: time.Now(),
				Windows: []usage.LimitWindow{
					{Name: "5-Hour", RemainingPct: 90, ResetsAt: time.Now().Add(2 * time.Hour)},
					{Name: "Weekly", RemainingPct: 80, ResetsAt: time.Now().Add(5 * 24 * time.Hour)},
				},
			})
		}(i)

		go func() {
			defer wg.Done()
			_, _ = store.Get("agy", "work")
		}()
	}

	wg.Wait()
}

func TestCacheStoreFlush(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-flush-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := usage.NewCacheStore(tmpDir, time.Minute)
	_ = store.Put(usage.Report{Agent: "agy", Profile: "p1", Status: usage.StatusOK})
	_ = store.Put(usage.Report{Agent: "agy", Profile: "p2", Status: usage.StatusOK})
	_ = store.Put(usage.Report{Agent: "gemini", Profile: "p1", Status: usage.StatusOK})

	// Flush only agy
	store.Flush("agy")

	if _, found := store.Get("agy", "p1"); found {
		t.Errorf("expected agy:p1 to be flushed")
	}
	if _, found := store.Get("agy", "p2"); found {
		t.Errorf("expected agy:p2 to be flushed")
	}
	if _, found := store.Get("gemini", "p1"); !found {
		t.Errorf("expected gemini:p1 to remain cached")
	}

	// Flush all
	store.Flush("")
	if _, found := store.Get("gemini", "p1"); found {
		t.Errorf("expected gemini:p1 to be flushed on empty agent")
	}
}

func TestCacheStoreDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-delete-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := usage.NewCacheStore(tmpDir, time.Minute)
	_ = store.Put(usage.Report{Agent: "agy", Profile: "p1", Status: usage.StatusOK})
	_ = store.Put(usage.Report{Agent: "agy", Profile: "p2", Status: usage.StatusOK})

	store.Delete("agy", "p1")

	if _, found := store.Get("agy", "p1"); found {
		t.Errorf("expected agy:p1 to be deleted")
	}
	if _, found := store.Get("agy", "p2"); !found {
		t.Errorf("expected agy:p2 to remain cached")
	}
}

func TestCacheStoreRename(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-rename-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := usage.NewCacheStore(tmpDir, time.Minute)
	_ = store.Put(usage.Report{Agent: "agy", Profile: "p1", Status: usage.StatusOK})
	_ = store.Put(usage.Report{Agent: "gemini", Profile: "p1", Status: usage.StatusOK})
	_ = store.Put(usage.Report{Agent: "agy", Profile: "other", Status: usage.StatusOK})

	store.Rename("p1", "p1-renamed")

	if _, found := store.Get("agy", "p1"); found {
		t.Errorf("expected agy:p1 to be removed after rename")
	}
	if _, found := store.Get("gemini", "p1"); found {
		t.Errorf("expected gemini:p1 to be removed after rename")
	}

	repAgy, foundAgy := store.Get("agy", "p1-renamed")
	if !foundAgy || repAgy.Profile != "p1-renamed" {
		t.Errorf("expected agy:p1-renamed to exist with updated profile name")
	}

	repGem, foundGem := store.Get("gemini", "p1-renamed")
	if !foundGem || repGem.Profile != "p1-renamed" {
		t.Errorf("expected gemini:p1-renamed to exist with updated profile name")
	}

	if _, found := store.Get("agy", "other"); !found {
		t.Errorf("expected agy:other to remain untouched")
	}

	// Verify persistence
	store2 := usage.NewCacheStore(tmpDir, time.Minute)
	if _, found := store2.Get("agy", "p1-renamed"); !found {
		t.Errorf("expected agy:p1-renamed to persist on disk")
	}
}

func TestRefreshAsync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-engine-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := usage.NewCacheStore(tmpDir, 5*time.Minute)

	// Pre-seed cached profile
	_ = store.Put(usage.Report{
		Agent:     "agy",
		Profile:   "cached_prof",
		Status:    usage.StatusOK,
		FetchedAt: time.Now(),
	})

	var callCount int32
	targets := []usage.TargetProfile{
		{
			Agent:   "agy",
			Profile: "cached_prof",
			GetUsageFn: func(ctx context.Context, p, dir string) (*usage.Report, error) {
				atomic.AddInt32(&callCount, 1)
				return &usage.Report{Agent: "agy", Profile: p, Status: usage.StatusOK}, nil
			},
		},
		{
			Agent:   "agy",
			Profile: "fresh_prof",
			GetUsageFn: func(ctx context.Context, p, dir string) (*usage.Report, error) {
				atomic.AddInt32(&callCount, 1)
				return &usage.Report{
					Agent:   "agy",
					Profile: p,
					Status:  usage.StatusOK,
					Windows: []usage.LimitWindow{
						{Name: "5-Hour", RemainingPct: 75},
					},
				}, nil
			},
		},
		{
			Agent:   "gemini",
			Profile: "error_prof",
			GetUsageFn: func(ctx context.Context, p, dir string) (*usage.Report, error) {
				return nil, errors.New("network failure")
			},
		},
	}

	ctx := context.Background()
	reportsChan := usage.RefreshAsync(ctx, targets, store)

	var received []usage.Report
	for rep := range reportsChan {
		received = append(received, rep)
	}

	if len(received) != 3 {
		t.Fatalf("expected 3 reports, got %d", len(received))
	}

	// cached_prof should NOT have invoked GetUsageFn
	if calls := atomic.LoadInt32(&callCount); calls != 1 {
		t.Errorf("expected GetUsageFn to be called once for fresh_prof, called %d times", calls)
	}

	// Check that fresh_prof was saved in store
	cachedFresh, found := store.Get("agy", "fresh_prof")
	if !found {
		t.Errorf("expected fresh_prof to be stored in cache")
	} else if cachedFresh.Status != usage.StatusOK {
		t.Errorf("unexpected status for fresh_prof in cache: %s", cachedFresh.Status)
	}

	// Check error report
	var errReport *usage.Report
	for _, r := range received {
		if r.Profile == "error_prof" {
			errReport = &r
			break
		}
	}
	if errReport == nil {
		t.Fatalf("missing report for error_prof")
	}
	if errReport.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown for error_prof, got %s", errReport.Status)
	}
	if errReport.Error != "network failure" {
		t.Errorf("expected error 'network failure', got %q", errReport.Error)
	}
}

func TestRefreshAsyncContextCancel(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-cancel-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := usage.NewCacheStore(tmpDir, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	targets := []usage.TargetProfile{
		{
			Agent:   "agy",
			Profile: "prof1",
			GetUsageFn: func(ctx context.Context, p, dir string) (*usage.Report, error) {
				time.Sleep(100 * time.Millisecond)
				return &usage.Report{Agent: "agy", Profile: p}, nil
			},
		},
	}

	reportsChan := usage.RefreshAsync(ctx, targets, store)
	var count int
	for range reportsChan {
		count++
	}
	// With context cancelled upfront, channel should close quickly and not block
	if count > 1 {
		t.Errorf("expected at most 1 report on cancelled context, got %d", count)
	}

	// Verify cancelled probe did NOT poison the cache with fallback Report
	if _, found := store.Get("agy", "prof1"); found {
		t.Errorf("cancelled context should not poison cache with fallback report")
	}
}

func TestCacheStoreDeepCopyIsolation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-deepcopy-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := usage.NewCacheStore(tmpDir, time.Hour)

	windows := []usage.LimitWindow{
		{Name: "5-Hour", RemainingPct: 90},
	}
	rep := usage.Report{
		Agent:   "agy",
		Profile: "work",
		Status:  usage.StatusOK,
		Windows: windows,
	}

	if err := store.Put(rep); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Mutate local slice passed to Put; cache should remain unaffected
	windows[0].RemainingPct = 10
	got1, _ := store.Get("agy", "work")
	if got1.Windows[0].RemainingPct != 90 {
		t.Errorf("cache was corrupted by mutating original slice: expected 90, got %d", got1.Windows[0].RemainingPct)
	}

	// Mutate slice returned by Get; cache should remain unaffected
	got1.Windows[0].RemainingPct = 20
	got2, _ := store.Get("agy", "work")
	if got2.Windows[0].RemainingPct != 90 {
		t.Errorf("cache was corrupted by mutating Get() slice: expected 90, got %d", got2.Windows[0].RemainingPct)
	}
}

func TestNilReceiverGuards(t *testing.T) {
	var rep *usage.Report
	if rep.PrimaryWindow() != nil {
		t.Errorf("expected nil for nil receiver PrimaryWindow()")
	}
	if rep.WeeklyWindow() != nil {
		t.Errorf("expected nil for nil receiver WeeklyWindow()")
	}
}

func TestCacheStoreEdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aim-usage-edge-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := usage.NewCacheStore(tmpDir, time.Hour)

	// Test past ResetsAt clamps ResetsIn to 0
	pastTime := time.Now().Add(-10 * time.Minute)
	err = store.Put(usage.Report{
		Agent:   "agy",
		Profile: "p1",
		Windows: []usage.LimitWindow{
			{Name: "5-Hour", RemainingPct: 50, ResetsAt: pastTime},
		},
	})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	cached, found := store.Get("agy", "p1")
	if !found {
		t.Fatalf("expected cache hit")
	}
	if cached.Windows[0].ResetsIn != 0 {
		t.Errorf("expected ResetsIn to be 0 for past reset time, got %v", cached.Windows[0].ResetsIn)
	}

	// Verify file permissions 0600
	cacheFilePath := filepath.Join(tmpDir, "cache", "usage.json")
	fi, err := os.Stat(cacheFilePath)
	if err != nil {
		t.Fatalf("failed to stat cache file: %v", err)
	}
	perm := fi.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected permissions 0600, got %#o", perm)
	}

	// Test loading corrupt cache file
	corruptDir, err := os.MkdirTemp("", "aim-usage-corrupt-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(corruptDir)

	corruptFilePath := filepath.Join(corruptDir, "cache", "usage.json")
	if err := os.MkdirAll(filepath.Dir(corruptFilePath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corruptFilePath, []byte("{invalid json"), 0600); err != nil {
		t.Fatal(err)
	}

	// Should not panic, should initialize empty
	corruptStore := usage.NewCacheStore(corruptDir, time.Hour)
	if _, found := corruptStore.Get("agy", "p1"); found {
		t.Errorf("expected miss on corrupt cache file")
	}
}

func TestRefreshAsyncEdgeCases(t *testing.T) {
	// Empty targets
	ch := usage.RefreshAsync(context.Background(), nil, nil)
	var reports []usage.Report
	for r := range ch {
		reports = append(reports, r)
	}
	if len(reports) != 0 {
		t.Errorf("expected 0 reports for nil targets, got %d", len(reports))
	}

	// nil GetUsageFn in TargetProfile
	targets := []usage.TargetProfile{
		{
			Agent:      "agy",
			Profile:    "p_nil_fn",
			GetUsageFn: nil,
		},
	}
	ch2 := usage.RefreshAsync(context.Background(), targets, nil)
	var reports2 []usage.Report
	for r := range ch2 {
		reports2 = append(reports2, r)
	}
	if len(reports2) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports2))
	}
	if reports2[0].Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown for nil GetUsageFn, got %s", reports2[0].Status)
	}
}

func TestCacheStoreIsStale(t *testing.T) {
	tmpDir := t.TempDir()
	store := usage.NewCacheStore(tmpDir, time.Hour)

	// Missing entry is stale
	if !store.IsStale("agy", "missing", 5*time.Minute) {
		t.Errorf("expected missing entry to be stale")
	}

	// Fresh entry is not stale
	_ = store.Put(usage.Report{
		Agent:     "agy",
		Profile:   "fresh",
		FetchedAt: time.Now(),
	})
	if store.IsStale("agy", "fresh", 5*time.Minute) {
		t.Errorf("expected freshly added entry NOT to be stale")
	}

	// Old entry is stale
	_ = store.Put(usage.Report{
		Agent:     "agy",
		Profile:   "old",
		FetchedAt: time.Now().Add(-10 * time.Minute),
	})
	if !store.IsStale("agy", "old", 5*time.Minute) {
		t.Errorf("expected old entry to be stale")
	}
}

func TestCanPrewarm(t *testing.T) {
	tmpDir := t.TempDir()

	// First call should succeed
	if !usage.CanPrewarm(tmpDir, time.Minute) {
		t.Errorf("expected first call to CanPrewarm to return true")
	}

	// Immediate second call should be throttled
	if usage.CanPrewarm(tmpDir, time.Minute) {
		t.Errorf("expected immediate second call to CanPrewarm to be throttled")
	}

	// Call with zero/negative cooldown should succeed
	if !usage.CanPrewarm(tmpDir, 0) {
		t.Errorf("expected call with 0 cooldown to succeed")
	}
}
