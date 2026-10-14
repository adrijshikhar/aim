package oauth_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/oauth"
)

func TestRefusalStore_RecordAndIsRefused(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewRefusalStore(tempDir)

	agent := "claude"
	profile := "work"
	code := oauth.RefusalInvalidGrant
	desc := "Invalid refresh token"

	// Initially not refused
	rec, ok := store.IsRefused(agent, profile)
	if ok || rec != nil {
		t.Fatalf("expected not refused initially, got ok=%v, rec=%v", ok, rec)
	}

	before := time.Now().UTC().Add(-1 * time.Second)
	if err := store.Record(agent, profile, code, desc); err != nil {
		t.Fatalf("unexpected error on Record: %v", err)
	}
	after := time.Now().UTC().Add(1 * time.Second)

	rec, ok = store.IsRefused(agent, profile)
	if !ok || rec == nil {
		t.Fatalf("expected refused record, got ok=%v, rec=%v", ok, rec)
	}

	if rec.Agent != agent {
		t.Errorf("expected Agent %q, got %q", agent, rec.Agent)
	}
	if rec.Profile != profile {
		t.Errorf("expected Profile %q, got %q", profile, rec.Profile)
	}
	if rec.Code != code {
		t.Errorf("expected Code %q, got %q", code, rec.Code)
	}
	if rec.Description != desc {
		t.Errorf("expected Description %q, got %q", desc, rec.Description)
	}
	if rec.RecordedAt.Before(before) || rec.RecordedAt.After(after) {
		t.Errorf("expected RecordedAt between %v and %v, got %v", before, after, rec.RecordedAt)
	}

	// Different agent/profile should not be refused
	recOther, okOther := store.IsRefused("openai", profile)
	if okOther || recOther != nil {
		t.Errorf("expected not refused for different agent, got ok=%v, rec=%v", okOther, recOther)
	}
	recOtherProf, okOtherProf := store.IsRefused(agent, "personal")
	if okOtherProf || recOtherProf != nil {
		t.Errorf("expected not refused for different profile, got ok=%v, rec=%v", okOtherProf, recOtherProf)
	}
}

func TestRefusalStore_Clear(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewRefusalStore(tempDir)

	agent := "anthropic"
	profile := "default"
	if err := store.Record(agent, profile, oauth.RefusalRefreshTokenReused, "Token consumed"); err != nil {
		t.Fatalf("failed to record: %v", err)
	}

	rec, ok := store.IsRefused(agent, profile)
	if !ok || rec == nil {
		t.Fatalf("expected refused before clear")
	}

	if err := store.Clear(agent, profile); err != nil {
		t.Fatalf("unexpected error on Clear: %v", err)
	}

	rec, ok = store.IsRefused(agent, profile)
	if ok || rec != nil {
		t.Fatalf("expected not refused after clear, got ok=%v, rec=%v", ok, rec)
	}

	// Clearing again should be idempotent and not error
	if err := store.Clear(agent, profile); err != nil {
		t.Fatalf("unexpected error on second Clear: %v", err)
	}
}

func TestRefusalStore_Expiry(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewRefusalStore(tempDir)

	filePath := filepath.Join(tempDir, "refusals.json")

	// Pre-populate with one expired record (25h old) and one active record (23h old)
	now := time.Now().UTC()
	expiredTime := now.Add(-25 * time.Hour)
	activeTime := now.Add(-23 * time.Hour)

	records := map[string]oauth.RefusalRecord{
		"claude:expired": {
			Agent:       "claude",
			Profile:     "expired",
			Code:        oauth.RefusalInvalidGrant,
			Description: "Expired refusal",
			RecordedAt:  expiredTime,
		},
		"claude:active": {
			Agent:       "claude",
			Profile:     "active",
			Code:        oauth.RefusalRefreshTokenExpired,
			Description: "Active refusal within 24h",
			RecordedAt:  activeTime,
		},
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal records: %v", err)
	}
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Expired should return false and nil
	recExpired, okExpired := store.IsRefused("claude", "expired")
	if okExpired || recExpired != nil {
		t.Errorf("expected expired record (>24h) to return false/nil, got ok=%v, rec=%v", okExpired, recExpired)
	}

	// Active should return true and record
	recActive, okActive := store.IsRefused("claude", "active")
	if !okActive || recActive == nil {
		t.Fatalf("expected active record (<24h) to return true/rec, got ok=%v, rec=%v", okActive, recActive)
	}
	if recActive.Code != oauth.RefusalRefreshTokenExpired {
		t.Errorf("expected code %q, got %q", oauth.RefusalRefreshTokenExpired, recActive.Code)
	}
}

func TestRefusalStore_NestedDirCreationAndFilePerms(t *testing.T) {
	tempDir := t.TempDir()
	nestedBase := filepath.Join(tempDir, "nested", "oauth", "store")
	store := oauth.NewRefusalStore(nestedBase)

	if err := store.Record("test-agent", "prof", oauth.RefusalInvalidGrant, "bad grant"); err != nil {
		t.Fatalf("failed to record in nested dir: %v", err)
	}

	filePath := filepath.Join(nestedBase, "refusals.json")
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat refusals.json: %v", err)
	}

	// File permission check: 0600
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file mode 0600, got %04o", perm)
	}

	// Check directory permission: 0700
	dirInfo, err := os.Stat(nestedBase)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}
	if dirInfo.Mode().Perm() != 0700 {
		t.Errorf("expected dir mode 0700, got %04o", dirInfo.Mode().Perm())
	}

	// Check for any leftover .tmp files
	entries, err := os.ReadDir(nestedBase)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("found leftover tmp file: %s", entry.Name())
		}
	}
}

func TestRefusalStore_PersistenceAcrossInstances(t *testing.T) {
	tempDir := t.TempDir()
	store1 := oauth.NewRefusalStore(tempDir)
	store2 := oauth.NewRefusalStore(tempDir)

	if err := store1.Record("agent1", "prof1", oauth.RefusalRefreshTokenInvalidated, "invalidated"); err != nil {
		t.Fatalf("store1 record error: %v", err)
	}

	rec, ok := store2.IsRefused("agent1", "prof1")
	if !ok || rec == nil {
		t.Fatalf("store2 expected to find record written by store1")
	}
	if rec.Code != oauth.RefusalRefreshTokenInvalidated {
		t.Errorf("expected %q, got %q", oauth.RefusalRefreshTokenInvalidated, rec.Code)
	}
}

func TestRefusalStore_CorruptedFileGraceful(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "refusals.json")
	if err := os.WriteFile(filePath, []byte("NOT_VALID_JSON{{{"), 0600); err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	store := oauth.NewRefusalStore(tempDir)
	rec, ok := store.IsRefused("agent", "prof")
	if ok || rec != nil {
		t.Errorf("expected false/nil for corrupted file, got ok=%v, rec=%v", ok, rec)
	}

	// Should be able to overwrite cleanly
	if err := store.Record("agent", "prof", oauth.RefusalInvalidGrant, "re-recorded"); err != nil {
		t.Fatalf("failed to record after corruption: %v", err)
	}

	rec, ok = store.IsRefused("agent", "prof")
	if !ok || rec == nil {
		t.Fatalf("expected successful retrieval after rewrite")
	}
}

func TestRefusalStore_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewRefusalStore(tempDir)

	var wg sync.WaitGroup
	workers := 16
	iterations := 25

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				agent := "agent"
				profile := "prof"

				if (workerID+i)%3 == 0 {
					_ = store.Record(agent, profile, oauth.RefusalInvalidGrant, "grant error")
				} else if (workerID+i)%3 == 1 {
					_, _ = store.IsRefused(agent, profile)
				} else {
					_ = store.Clear(agent, profile)
				}
			}
		}(w)
	}

	wg.Wait()
}

func TestRefusalStore_Prune(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewRefusalStore(tempDir)
	filePath := filepath.Join(tempDir, "refusals.json")

	now := time.Now().UTC()
	expiredTime := now.Add(-25 * time.Hour)
	activeTime := now.Add(-1 * time.Hour)

	records := map[string]oauth.RefusalRecord{
		"claude:expired": {
			Agent:       "claude",
			Profile:     "expired",
			Code:        oauth.RefusalInvalidGrant,
			Description: "Expired refusal",
			RecordedAt:  expiredTime,
		},
		"claude:active": {
			Agent:       "claude",
			Profile:     "active",
			Code:        oauth.RefusalRefreshTokenExpired,
			Description: "Active refusal",
			RecordedAt:  activeTime,
		},
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal records: %v", err)
	}
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := store.Prune(); err != nil {
		t.Fatalf("unexpected error on Prune: %v", err)
	}

	// Verify on-disk file no longer contains expired record
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read refusals file: %v", err)
	}
	var onDisk map[string]oauth.RefusalRecord
	if err := json.Unmarshal(fileData, &onDisk); err != nil {
		t.Fatalf("failed to unmarshal pruned file: %v", err)
	}

	if _, exists := onDisk["claude:expired"]; exists {
		t.Errorf("expected expired record to be purged from disk by Prune")
	}
	if _, exists := onDisk["claude:active"]; !exists {
		t.Errorf("expected active record to be preserved on disk by Prune")
	}

	// Calling Prune again when nothing to purge should be a no-op returning nil
	if err := store.Prune(); err != nil {
		t.Fatalf("unexpected error on second Prune: %v", err)
	}
}

