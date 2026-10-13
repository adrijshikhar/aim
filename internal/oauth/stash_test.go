package oauth_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/oauth"
)

func TestStashStore_SaveAndFind(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewStashStore(tempDir)

	spentToken := "old-refresh-token-12345"
	newAccess := "new-access-token"
	newRefresh := "new-refresh-token"
	expiresAt := time.Now().Add(1 * time.Hour).Truncate(time.Second)

	err := store.Save(spentToken, "claude", "work", newAccess, newRefresh, expiresAt)
	if err != nil {
		t.Fatalf("expected nil error on Save, got: %v", err)
	}

	// Verify directory permissions: 0700
	stashDir := store.Dir
	di, err := os.Stat(stashDir)
	if err != nil {
		t.Fatalf("expected stash dir to exist: %v", err)
	}
	if perm := di.Mode().Perm(); perm != 0700 {
		t.Errorf("expected 0700 dir permissions, got %o", perm)
	}

	// Verify file permissions: 0600
	stashPath := store.StashPath(spentToken)
	fi, err := os.Stat(stashPath)
	if err != nil {
		t.Fatalf("expected stash file to exist: %v", err)
	}
	if perm := fi.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600 file permissions, got %o", perm)
	}

	// Retrieve
	cred, err := store.Find(spentToken)
	if err != nil {
		t.Fatalf("expected nil error on Find, got: %v", err)
	}
	if cred == nil {
		t.Fatalf("expected non-nil cred")
	}
	if cred.AccessToken != newAccess || cred.RefreshToken != newRefresh {
		t.Errorf("token mismatch: got %+v", cred)
	}
	if cred.Agent != "claude" || cred.Profile != "work" {
		t.Errorf("agent/profile mismatch: got %+v", cred)
	}
	if cred.SpentTokenHash == "" {
		t.Errorf("expected non-empty SpentTokenHash")
	}

	// Adopt matching agent and profile
	adopted, err := store.Adopt(spentToken, "claude", "work")
	if err != nil {
		t.Fatalf("expected nil error on Adopt, got: %v", err)
	}
	if adopted == nil || adopted.AccessToken != newAccess {
		t.Errorf("expected valid adopted credential, got %+v", adopted)
	}

	// Adopt mismatch should fail
	_, err = store.Adopt(spentToken, "other-agent", "work")
	if err == nil {
		t.Fatalf("expected error on Adopt with agent mismatch, got nil")
	}

	// Remove after successful commit
	if err := store.Remove(spentToken); err != nil {
		t.Fatalf("expected nil error on Remove, got: %v", err)
	}

	credAfter, err := store.Find(spentToken)
	if err != nil {
		t.Fatalf("expected nil error on Find after remove, got: %v", err)
	}
	if credAfter != nil {
		t.Errorf("expected cred to be nil after remove")
	}

	// Remove again should be idempotent and succeed
	if err := store.Remove(spentToken); err != nil {
		t.Fatalf("expected nil error on second Remove, got: %v", err)
	}
}

func TestStashStore_Find_NonExistent(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewStashStore(tempDir)

	cred, err := store.Find("non-existent-token")
	if err != nil {
		t.Fatalf("expected nil error for non-existent token, got: %v", err)
	}
	if cred != nil {
		t.Errorf("expected nil cred for non-existent token, got: %+v", cred)
	}

	adopted, err := store.Adopt("non-existent-token", "claude", "work")
	if err != nil {
		t.Fatalf("expected nil error on Adopt for non-existent token, got: %v", err)
	}
	if adopted != nil {
		t.Errorf("expected nil adopted cred for non-existent token, got: %+v", adopted)
	}
}

func TestStashStore_Find_Corrupted(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewStashStore(tempDir)

	spentToken := "corrupt-tok"
	_ = os.MkdirAll(store.Dir, 0700)
	stashPath := store.StashPath(spentToken)
	_ = os.WriteFile(stashPath, []byte("invalid-json{"), 0600)

	_, err := store.Find(spentToken)
	if err == nil {
		t.Fatalf("expected error on corrupted stash record, got nil")
	}
}

func TestStashStore_List(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewStashStore(tempDir)

	// List when dir doesn't exist yet
	creds, err := store.List()
	if err != nil {
		t.Fatalf("expected nil error on List with missing dir, got: %v", err)
	}
	if len(creds) != 0 {
		t.Fatalf("expected 0 creds, got %d", len(creds))
	}

	// Save 2 creds
	_ = store.Save("tok-1", "agy", "prof1", "acc1", "ref1", time.Now().Add(time.Hour))
	_ = store.Save("tok-2", "claude", "prof2", "acc2", "ref2", time.Now().Add(2*time.Hour))

	// Write a stray non-json file and sub-directory
	_ = os.WriteFile(filepath.Join(store.Dir, "stray.tmp"), []byte("temp"), 0600)
	_ = os.MkdirAll(filepath.Join(store.Dir, "subdir"), 0700)

	creds, err = store.List()
	if err != nil {
		t.Fatalf("expected nil error on List, got: %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("expected 2 creds, got %d", len(creds))
	}
}

func TestStashStore_Prune(t *testing.T) {
	tempDir := t.TempDir()
	store := oauth.NewStashStore(tempDir)

	spentToken := "old-token-to-prune"
	_ = store.Save(spentToken, "claude", "work", "acc", "ref", time.Now())

	stashPath := store.StashPath(spentToken)
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	_ = os.Chtimes(stashPath, oldTime, oldTime)

	// Keep another fresh one
	freshToken := "fresh-token-to-keep"
	_ = store.Save(freshToken, "claude", "work", "acc2", "ref2", time.Now())

	pruned, err := store.Prune(7 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}
	if pruned != 1 {
		t.Errorf("expected 1 file pruned, got %d", pruned)
	}

	// Verify old one is gone and fresh one is still there
	oldCred, _ := store.Find(spentToken)
	if oldCred != nil {
		t.Errorf("expected old cred to be pruned")
	}

	freshCred, _ := store.Find(freshToken)
	if freshCred == nil {
		t.Errorf("expected fresh cred to be kept")
	}
}
