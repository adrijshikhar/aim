package share

import (
	"testing"
	"time"
)

func TestShareStoreEnableAndList(t *testing.T) {
	tempDir := t.TempDir()
	store := NewStore(tempDir)

	err := store.EnableProfile("work", "claude", true, 4*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profiles := store.GetSharedProfiles()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 shared profile, got %d", len(profiles))
	}
	if profiles[0].Name != "work" || !profiles[0].AutoApprove || profiles[0].MaxLeaseTTL != 4*time.Hour {
		t.Errorf("unexpected profile: %+v", profiles[0])
	}

	err = store.DisableProfile("work")
	if err != nil {
		t.Fatalf("failed to disable: %v", err)
	}
	if len(store.GetSharedProfiles()) != 0 {
		t.Errorf("expected 0 shared profiles after disable")
	}
}

func TestShareStorePeerPersistence(t *testing.T) {
	tempDir := t.TempDir()
	store := NewStore(tempDir)

	peer := PeerInfo{
		ID:       "peer-1",
		HostName: "bobs-mac",
		VPNIP:    "10.8.16.9",
		Port:     4319,
		LastSeen: time.Now(),
		Profiles: []string{"work"},
	}

	if err := store.SavePeer(peer); err != nil {
		t.Fatalf("failed to save peer: %v", err)
	}

	peers := store.GetPeers()
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	if peers[0].HostName != "bobs-mac" || peers[0].VPNIP != "10.8.16.9" {
		t.Errorf("unexpected peer data: %+v", peers[0])
	}

	// Update existing peer
	peer.HostName = "bobs-new-mac"
	if err := store.SavePeer(peer); err != nil {
		t.Fatalf("failed to update peer: %v", err)
	}
	peers = store.GetPeers()
	if len(peers) != 1 || peers[0].HostName != "bobs-new-mac" {
		t.Errorf("expected updated single peer, got: %+v", peers)
	}
}

func TestShareStoreGrantedLeasesPruning(t *testing.T) {
	tempDir := t.TempDir()
	store := NewStore(tempDir)

	now := time.Now()
	// Active lease (valid for 1 hour)
	activeLease := GrantedLease{
		LeaseID:      "ls_active",
		ProfileName:  "work",
		Agent:        "claude",
		BorrowerID:   "peer-1",
		GrantedAt:    now,
		ExpiresAt:    now.Add(1 * time.Hour),
		Revoked:      false,
		AutoApproved: true,
	}

	// Expired lease (expired 10 minutes ago)
	expiredLease := GrantedLease{
		LeaseID:      "ls_expired",
		ProfileName:  "office",
		Agent:        "codex",
		BorrowerID:   "peer-2",
		GrantedAt:    now.Add(-2 * time.Hour),
		ExpiresAt:    now.Add(-10 * time.Minute),
		Revoked:      false,
		AutoApproved: false,
	}

	if err := store.SaveGrantedLease(activeLease); err != nil {
		t.Fatalf("failed to save active lease: %v", err)
	}
	if err := store.SaveGrantedLease(expiredLease); err != nil {
		t.Fatalf("failed to save expired lease: %v", err)
	}

	// GetActiveGrantedLeases returns only active unexpired leases
	activeLeases := store.GetActiveGrantedLeases()
	if len(activeLeases) != 1 {
		t.Fatalf("expected 1 active lease, got %d", len(activeLeases))
	}
	if activeLeases[0].LeaseID != "ls_active" {
		t.Errorf("expected ls_active, got %s", activeLeases[0].LeaseID)
	}

	// GetGrantedLeases retains history for dashboard audit & reporting
	allLeases := store.GetGrantedLeases()
	if len(allLeases) != 2 {
		t.Fatalf("expected 2 leases in history, got %d", len(allLeases))
	}

	// Revoke the active lease
	revoked, err := store.RevokeGrantedLease("ls_active")
	if err != nil {
		t.Fatalf("failed to revoke lease: %v", err)
	}
	if !revoked.Revoked {
		t.Errorf("expected revoked.Revoked to be true")
	}

	// Now GetActiveGrantedLeases should return 0 active leases
	activeAfterRevoke := store.GetActiveGrantedLeases()
	if len(activeAfterRevoke) != 0 {
		t.Fatalf("expected 0 active leases after revoke, got %d", len(activeAfterRevoke))
	}

	// But history in GetGrantedLeases retains the revoked status
	historyAfterRevoke := store.GetGrantedLeases()
	if len(historyAfterRevoke) != 2 {
		t.Fatalf("expected 2 leases retained in history after revoke, got %d", len(historyAfterRevoke))
	}

	// PruneExpiredGrantedLeases explicitly purges inactive leases from disk
	pruned := store.PruneExpiredGrantedLeases()
	if pruned != 2 {
		t.Fatalf("expected 2 pruned leases, got %d", pruned)
	}
	remaining := store.GetGrantedLeases()
	if len(remaining) != 0 {
		t.Fatalf("expected 0 remaining leases after PruneExpiredGrantedLeases, got %d", len(remaining))
	}
}

func TestShareStoreActiveLeasesNotDroppedByHistoryCap(t *testing.T) {
	tempDir := t.TempDir()
	store := NewStore(tempDir)
	now := time.Now()

	// Add an active lease first
	activeLease := GrantedLease{
		LeaseID:     "ls_vital_active",
		ProfileName: "work",
		Agent:       "claude",
		GrantedAt:   now,
		ExpiresAt:   now.Add(2 * time.Hour),
		Revoked:     false,
	}
	if err := store.SaveGrantedLease(activeLease); err != nil {
		t.Fatalf("failed to save active lease: %v", err)
	}

	// Add 120 expired leases
	for i := 0; i < 120; i++ {
		expLease := GrantedLease{
			LeaseID:     "ls_exp_" + string(rune('a'+(i%26))) + string(rune('0'+(i/26))),
			ProfileName: "work",
			Agent:       "claude",
			GrantedAt:   now.Add(-2 * time.Hour),
			ExpiresAt:   now.Add(-1 * time.Hour),
			Revoked:     false,
		}
		if err := store.SaveGrantedLease(expLease); err != nil {
			t.Fatalf("failed to save expired lease: %v", err)
		}
	}

	// Active lease must NOT have been dropped by history truncation
	actives := store.GetActiveGrantedLeases()
	if len(actives) != 1 || actives[0].LeaseID != "ls_vital_active" {
		t.Fatalf("expected active lease to be retained, got: %+v", actives)
	}
}

func TestShareStoreRegisterProfile(t *testing.T) {
	tempDir := t.TempDir()
	store := NewStore(tempDir)

	if err := store.RegisterProfile("agy_work", "agy", false, true, 4*time.Hour); err != nil {
		t.Fatalf("RegisterProfile failed: %v", err)
	}

	all := store.GetAllSharedProfiles()
	if len(all) != 1 {
		t.Fatalf("expected 1 profile in GetAllSharedProfiles, got %d", len(all))
	}
	if all[0].Name != "agy_work" || all[0].Enabled != false || all[0].Agent != "agy" {
		t.Fatalf("unexpected profile registered: %+v", all[0])
	}

	// But GetSharedProfiles() should only return enabled profiles (0)
	if len(store.GetSharedProfiles()) != 0 {
		t.Fatalf("expected 0 profiles in GetSharedProfiles, got %d", len(store.GetSharedProfiles()))
	}
}

func TestShareStoreDBUnavailableReturnsError(t *testing.T) {
	// A store with nil database must not report success on request operations
	store := &ShareStore{}

	req := ShareRequestRecord{
		ID:          "req-test",
		PeerID:      "peer1",
		ProfileName: "work",
		Agent:       "claude",
		Direction:   "incoming",
		Status:      "pending",
	}

	if err := store.SaveShareRequest(req); err == nil {
		t.Fatalf("expected error when saving request with nil database, got nil")
	}

	if _, err := store.GetShareRequest("req-test"); err == nil {
		t.Fatalf("expected error when getting request with nil database, got nil")
	}

	if err := store.UpdateRequestStatus("req-test", "approved", "ls_1", ""); err == nil {
		t.Fatalf("expected error when updating request status with nil database, got nil")
	}

	if _, err := store.TransitionRequestStatus("req-test", "pending", "approved", "ls_1", ""); err == nil {
		t.Fatalf("expected error when transitioning request status with nil database, got nil")
	}

	if _, err := store.CancelShareRequest("req-test"); err == nil {
		t.Fatalf("expected error when cancelling request with nil database, got nil")
	}
}
