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
	if localName != "alice:work" {
		t.Errorf("expected alice:work, got %s", localName)
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

func TestLeaseManagerPathTraversalRejection(t *testing.T) {
	profilesDir := t.TempDir()
	mgr := NewLeaseManager(profilesDir)

	// Attempt path traversal using ../../evil.txt
	resp := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_evil",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				"../../evil.txt": base64.StdEncoding.EncodeToString([]byte("pwned")),
			},
		},
	}

	_, err := mgr.InstallLease("attacker", "work", resp)
	if err == nil {
		t.Fatalf("expected InstallLease to reject path traversal, but succeeded")
	}
	if mgr.HasLease("attacker:work") {
		t.Fatalf("expected .aim-lease.json NOT to exist after rejected install")
	}

	// Attempt unauthorized file write (e.g. .ssh/id_rsa)
	resp2 := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_ssh",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".ssh/authorized_keys": base64.StdEncoding.EncodeToString([]byte("pwned")),
			},
		},
	}

	_, err = mgr.InstallLease("attacker", "work", resp2)
	if err == nil {
		t.Fatalf("expected InstallLease to reject non-credential file .ssh/authorized_keys, but succeeded")
	}
	if mgr.HasLease("attacker:work") {
		t.Fatalf("expected .aim-lease.json NOT to exist after rejected unauthorized file write")
	}

	// Verify that a failed install does not leave a directory blocking subsequent valid retries
	respValidRetry := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_valid_retry",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"valid"}`)),
			},
		},
	}
	localName, err := mgr.InstallLease("attacker", "work", respValidRetry)
	if err != nil {
		t.Fatalf("expected subsequent valid install to succeed after failed attempts, but got: %v", err)
	}
	if !mgr.HasLease(localName) {
		t.Fatalf("expected lease to be active for %s after retry", localName)
	}
	_ = mgr.ReleaseLease(localName)

	// Attempt sibling file injection (e.g. .claude_backdoor.sh)
	resp3 := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_sibling",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude_backdoor.sh": base64.StdEncoding.EncodeToString([]byte("#!/bin/sh")),
			},
		},
	}
	_, err = mgr.InstallLease("attacker", "work", resp3)
	if err == nil {
		t.Fatalf("expected InstallLease to reject sibling file injection .claude_backdoor.sh")
	}

	// Attempt overwriting a permanent profile that has no .aim-lease.json
	permDir := filepath.Join(profilesDir, "permanent_user")
	_ = os.MkdirAll(permDir, 0700)
	_ = os.WriteFile(filepath.Join(permDir, "user_data.txt"), []byte("vital data"), 0600)

	respPerm := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_perm",
		ProfileName: "permanent_user",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/auth.json": base64.StdEncoding.EncodeToString([]byte("{}")),
			},
		},
	}
	_, err = mgr.InstallLease("", "permanent_user", respPerm)
	if err == nil {
		t.Fatalf("expected InstallLease to reject overwriting permanent user profile")
	}
}

func TestLeaseManagerReleaseSafety(t *testing.T) {
	profilesDir := t.TempDir()
	mgr := NewLeaseManager(profilesDir)

	// Permanent profile (no .aim-lease.json)
	permDir := filepath.Join(profilesDir, "work")
	_ = os.MkdirAll(permDir, 0700)
	_ = os.WriteFile(filepath.Join(permDir, "config.json"), []byte("{}"), 0600)

	// ReleaseLease MUST fail on permanent profile
	err := mgr.ReleaseLease("work")
	if err == nil {
		t.Fatalf("expected ReleaseLease on permanent profile without .aim-lease.json to fail, but succeeded")
	}
	// Verify permanent profile was NOT deleted
	if _, err := os.Stat(filepath.Join(permDir, "config.json")); err != nil {
		t.Errorf("permanent profile file was deleted! %v", err)
	}

	// ReleaseLease on path traversal names
	if err := mgr.ReleaseLease("../profiles"); err == nil {
		t.Fatalf("expected ReleaseLease with traversal to fail")
	}
	if err := mgr.ReleaseLease("."); err == nil {
		t.Fatalf("expected ReleaseLease with . to fail")
	}
}

func TestLeaseManagerCleanExpired(t *testing.T) {
	profilesDir := t.TempDir()
	mgr := NewLeaseManager(profilesDir)

	// Leased profile 1: expired
	expiredResp := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_exp",
		ProfileName: "old",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(-10 * time.Minute), // expired
	}
	p1, err := mgr.InstallLease("peer1", "old", expiredResp)
	if err != nil {
		t.Fatalf("failed to install expired lease: %v", err)
	}

	// Leased profile 2: active
	activeResp := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_act",
		ProfileName: "current",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(2 * time.Hour), // active
	}
	p2, err := mgr.InstallLease("peer2", "current", activeResp)
	if err != nil {
		t.Fatalf("failed to install active lease: %v", err)
	}

	cleaned, err := mgr.CleanExpiredLeases()
	if err != nil {
		t.Fatalf("CleanExpiredLeases failed: %v", err)
	}
	if cleaned != 1 {
		t.Errorf("expected 1 cleaned lease, got %d", cleaned)
	}

	// p1 should be deleted, p2 should still exist
	if _, err := os.Stat(filepath.Join(profilesDir, p1)); !os.IsNotExist(err) {
		t.Errorf("expected expired profile %s to be deleted", p1)
	}
	if _, err := os.Stat(filepath.Join(profilesDir, p2)); err != nil {
		t.Errorf("expected active profile %s to still exist: %v", p2, err)
	}
}
