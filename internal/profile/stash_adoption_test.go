package profile_test

import (
	"os"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/profile"
)

func TestProfileManager_AdoptStash(t *testing.T) {
	tempDir := t.TempDir()
	pm := profile.NewProfileManager(tempDir)

	store := pm.UnclaimedStore()
	spentToken := "spent-refresh-token-xyz"
	newAccess := "new-acc"
	newRefresh := "new-ref"
	expiresAt := time.Now().Add(2 * time.Hour).Truncate(time.Second)

	if err := store.Save(spentToken, "claude", "work", newAccess, newRefresh, expiresAt); err != nil {
		t.Fatalf("failed to save stash: %v", err)
	}

	// Successful adoption
	cred, err := pm.AdoptStash("claude", "work", spentToken)
	if err != nil {
		t.Fatalf("expected nil error on AdoptStash, got: %v", err)
	}
	if cred == nil || cred.AccessToken != newAccess || cred.RefreshToken != newRefresh {
		t.Fatalf("unexpected cred adopted: %+v", cred)
	}

	// Mismatched agent or profile
	_, err = pm.AdoptStash("agy", "work", spentToken)
	if err == nil {
		t.Fatalf("expected error on agent mismatch, got nil")
	}

	// FindUnclaimedCredentials
	unclaimed, err := pm.FindUnclaimedCredentials("claude", "work")
	if err != nil {
		t.Fatalf("expected nil error on FindUnclaimedCredentials, got: %v", err)
	}
	if len(unclaimed) != 1 || unclaimed[0].AccessToken != newAccess {
		t.Fatalf("unexpected unclaimed results: %+v", unclaimed)
	}

	unclaimedOther, err := pm.FindUnclaimedCredentials("claude", "personal")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(unclaimedOther) != 0 {
		t.Fatalf("expected 0 unclaimed for personal, got %d", len(unclaimedOther))
	}

	// PruneUnclaimed
	stashPath := store.StashPath(spentToken)
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	_ = os.Chtimes(stashPath, oldTime, oldTime)

	pruned, err := pm.PruneUnclaimed(0) // default 7 days
	if err != nil {
		t.Fatalf("PruneUnclaimed failed: %v", err)
	}
	if pruned != 1 {
		t.Errorf("expected 1 file pruned, got %d", pruned)
	}
}
