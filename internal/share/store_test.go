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
