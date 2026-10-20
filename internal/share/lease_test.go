package share

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLeaseManagerInstallAndExpiry(t *testing.T) {
	profilesDir := t.TempDir()
	mgr := NewLeaseManager(profilesDir)

	resp := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_test",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/.credentials.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"valid"}`)),
			},
		},
	}

	localName, err := mgr.InstallLease("alice", "work", resp)
	if err != nil {
		t.Fatalf("failed to install lease: %v", err)
	}
	if localName != "p2p-alice-work" {
		t.Errorf("expected p2p-alice-work, got %s", localName)
	}

	credsFile := filepath.Join(profilesDir, localName, ".claude", ".credentials.json")
	data, err := os.ReadFile(credsFile)
	if err != nil || string(data) != `{"token":"valid"}` {
		t.Errorf("credential file not written properly: %v", err)
	}

	if mgr.IsLeaseExpired(localName) {
		t.Errorf("expected lease not expired")
	}

	err = mgr.ReleaseLease(localName)
	if err != nil {
		t.Fatalf("failed to release: %v", err)
	}
	if _, err := os.Stat(filepath.Join(profilesDir, localName)); !os.IsNotExist(err) {
		t.Errorf("expected directory deleted after release")
	}
}
