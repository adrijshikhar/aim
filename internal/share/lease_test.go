package share

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func setupTestLeaseManager(t *testing.T) (*LeaseManager, string) {
	t.Helper()
	tempHome := t.TempDir()
	t.Setenv("AIM_HOME", tempHome)
	profilesDir := filepath.Join(tempHome, "profiles")
	if err := os.MkdirAll(profilesDir, 0700); err != nil {
		t.Fatalf("failed to create temp profiles dir: %v", err)
	}
	mgr := NewLeaseManager(profilesDir)
	mgr.NotifyFn = func(sourcePeer, leaseID string) {} // stub network notifications in tests
	return mgr, profilesDir
}

func TestLeaseManagerInstallAndExpiry(t *testing.T) {
	mgr, profilesDir := setupTestLeaseManager(t)

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
	mgr, profilesDir := setupTestLeaseManager(t)

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
	mgr, profilesDir := setupTestLeaseManager(t)

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
	mgr, profilesDir := setupTestLeaseManager(t)

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

func TestLeaseManagerRenewalPreservesHistory(t *testing.T) {
	mgr, profilesDir := setupTestLeaseManager(t)

	resp1 := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_initial",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"initial"}`)),
			},
		},
	}
	pName, err := mgr.InstallLease("peer1", "work", resp1)
	if err != nil {
		t.Fatalf("initial InstallLease failed: %v", err)
	}

	// Simulate borrower running sessions and generating local conversation transcripts
	profilePath := filepath.Join(profilesDir, pName)
	convoDir := filepath.Join(profilePath, ".claude", "conversations")
	_ = os.MkdirAll(convoDir, 0700)
	convoFile := filepath.Join(convoDir, "chat_123.json")
	_ = os.WriteFile(convoFile, []byte(`{"conversation_id":"123","messages":["hello"]}`), 0600)

	// Renew the lease with new token and updated expiry
	resp2 := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_renewed",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(4 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"renewed"}`)),
			},
		},
	}
	pName2, err := mgr.InstallLease("peer1", "work", resp2)
	if err != nil {
		t.Fatalf("renewal InstallLease failed: %v", err)
	}
	if pName2 != pName {
		t.Fatalf("expected profile name %s, got %s", pName, pName2)
	}

	// Verify conversation transcript was preserved across renewal
	convoData, err := os.ReadFile(convoFile)
	if err != nil {
		t.Fatalf("expected conversation file %s to be preserved after renewal, but got error: %v", convoFile, err)
	}
	if !strings.Contains(string(convoData), "conversation_id") || !strings.Contains(string(convoData), "hello") {
		t.Fatalf("conversation data corrupted after renewal: %s", string(convoData))
	}

	// Verify token was updated
	authData, _ := os.ReadFile(filepath.Join(profilePath, ".claude", "auth.json"))
	if !strings.Contains(string(authData), "renewed") {
		t.Fatalf("expected updated token in auth.json, got: %s", string(authData))
	}
}

func TestLeaseManagerRenewalAtomicRollbackOnFailure(t *testing.T) {
	mgr, profilesDir := setupTestLeaseManager(t)

	initialExpiry := time.Now().Add(1 * time.Hour).Truncate(time.Second)
	resp1 := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_initial",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   initialExpiry,
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"initial_safe"}`)),
			},
		},
	}
	pName, err := mgr.InstallLease("peer1", "work", resp1)
	if err != nil {
		t.Fatalf("initial InstallLease failed: %v", err)
	}

	profilePath := filepath.Join(profilesDir, pName)

	// Create an unwriteable directory to force failure when staging the second file
	codexDir := filepath.Join(profilePath, ".codex")
	if err := os.MkdirAll(codexDir, 0500); err != nil {
		t.Fatalf("failed to create codex dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(codexDir, 0700)
	})

	// Renewal payload contains two files: the first would succeed, but the second will fail in .codex
	resp2 := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_failed_renewal",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(10 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"uncommitted_change"}`)),
				".codex/auth.json":  base64.StdEncoding.EncodeToString([]byte(`{"token":"should_fail"}`)),
			},
		},
	}

	_, err = mgr.InstallLease("peer1", "work", resp2)
	if err == nil {
		t.Fatalf("expected InstallLease to fail on unwritable dir, but it succeeded")
	}

	// Verify original .claude/auth.json remains completely intact with original credentials
	authData, err := os.ReadFile(filepath.Join(profilePath, ".claude", "auth.json"))
	if err != nil {
		t.Fatalf("failed to read auth.json: %v", err)
	}
	if !strings.Contains(string(authData), "initial_safe") {
		t.Fatalf("expected auth.json to retain initial credentials, got: %s", string(authData))
	}
	if strings.Contains(string(authData), "uncommitted_change") {
		t.Fatalf("auth.json contains uncommitted renewal change after failure")
	}

	// Verify .aim-lease.json was NOT updated to the new expiry or ID
	metaData, err := os.ReadFile(filepath.Join(profilePath, ".aim-lease.json"))
	if err != nil {
		t.Fatalf("failed to read lease metadata: %v", err)
	}
	var meta LeaseMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		t.Fatalf("failed to unmarshal metadata: %v", err)
	}
	if meta.LeaseID != "ls_initial" {
		t.Fatalf("expected lease ID ls_initial, got %s", meta.LeaseID)
	}

	// Verify no temporary or backup files remain on disk after failed renewal
	_ = filepath.Walk(profilePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			if strings.Contains(info.Name(), ".tmp.") || strings.Contains(info.Name(), ".bak.") {
				t.Errorf("temporary/backup file was left behind after failed renewal: %s", path)
			}
		}
		return nil
	})
}

func TestLeaseManagerConcurrentRenewals(t *testing.T) {
	mgr, profilesDir := setupTestLeaseManager(t)

	resp1 := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_initial",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"initial"}`)),
			},
		},
	}
	pName, err := mgr.InstallLease("peer1", "work", resp1)
	if err != nil {
		t.Fatalf("initial InstallLease failed: %v", err)
	}

	const workers = 8
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r := &LeaseResponse{
				Status:      "approved",
				LeaseID:     fmt.Sprintf("ls_renew_%d", workerID),
				ProfileName: "work",
				Agent:       "claude",
				ExpiresAt:   time.Now().Add(time.Duration(workerID+1) * time.Hour),
				Payload: &SharePayload{
					Files: map[string]string{
						".claude/auth.json": base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"token":"token_%d"}`, workerID))),
					},
				},
			}
			_, err := mgr.InstallLease("peer1", "work", r)
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent renewal failed: %v", err)
	}

	// Verify lease metadata and profile are intact and readable
	metaData, err := os.ReadFile(filepath.Join(profilesDir, pName, ".aim-lease.json"))
	if err != nil {
		t.Fatalf("failed to read metadata after concurrent renewals: %v", err)
	}
	var meta LeaseMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		t.Fatalf("metadata corrupted after concurrent renewals: %v", err)
	}
	if !strings.HasPrefix(meta.LeaseID, "ls_renew_") {
		t.Fatalf("unexpected lease ID after renewals: %s", meta.LeaseID)
	}
}

func TestLeaseManagerScrubCredentials(t *testing.T) {
	mgr, profilesDir := setupTestLeaseManager(t)

	targetDir := filepath.Join(profilesDir, "test-scrub")
	_ = os.MkdirAll(filepath.Join(targetDir, ".claude"), 0700)
	_ = os.MkdirAll(filepath.Join(targetDir, ".codex"), 0700)
	_ = os.MkdirAll(filepath.Join(targetDir, ".gemini", "antigravity-cli"), 0700)

	claudeCreds := filepath.Join(targetDir, ".claude", ".credentials.json")
	codexCreds := filepath.Join(targetDir, ".codex", "auth.json")
	agyCreds := filepath.Join(targetDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	userFile := filepath.Join(targetDir, "user_notes.txt")

	_ = os.WriteFile(claudeCreds, []byte("super-secret-claude-token"), 0600)
	_ = os.WriteFile(codexCreds, []byte("super-secret-codex-token"), 0600)
	_ = os.WriteFile(agyCreds, []byte("super-secret-agy-token"), 0600)
	_ = os.WriteFile(userFile, []byte("keep-my-notes"), 0600)

	mgr.ScrubCredentials(targetDir)

	if _, err := os.Stat(claudeCreds); !os.IsNotExist(err) {
		t.Errorf("expected claude credentials to be scrubbed and removed")
	}
	if _, err := os.Stat(codexCreds); !os.IsNotExist(err) {
		t.Errorf("expected codex credentials to be scrubbed and removed")
	}
	if _, err := os.Stat(agyCreds); !os.IsNotExist(err) {
		t.Errorf("expected agy credentials to be scrubbed and removed")
	}
	// User files must NOT be deleted
	if data, err := os.ReadFile(userFile); err != nil || string(data) != "keep-my-notes" {
		t.Errorf("user notes were affected by ScrubCredentials: %v", err)
	}
}

func TestLeaseManagerRevokeLeaseAndScrub(t *testing.T) {
	mgr, profilesDir := setupTestLeaseManager(t)

	resp := &LeaseResponse{
		Status:      "approved",
		LeaseID:     "ls_revoke_test",
		ProfileName: "work",
		Agent:       "claude",
		ExpiresAt:   time.Now().Add(2 * time.Hour),
		Payload: &SharePayload{
			Files: map[string]string{
				".claude/.credentials.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"secret"}`)),
			},
		},
	}

	localName, err := mgr.InstallLease("peer1", "work", resp)
	if err != nil {
		t.Fatalf("InstallLease failed: %v", err)
	}

	credsPath := filepath.Join(profilesDir, localName, ".claude", ".credentials.json")
	if _, err := os.Stat(credsPath); err != nil {
		t.Fatalf("credentials not installed: %v", err)
	}

	// Revoke lease
	if err := mgr.RevokeLease(localName); err != nil {
		t.Fatalf("RevokeLease failed: %v", err)
	}

	// Check credentials are scrubbed
	if _, err := os.Stat(credsPath); !os.IsNotExist(err) {
		t.Errorf("expected credentials to be scrubbed on RevokeLease")
	}

	// Check metadata marked revoked
	meta, err := mgr.GetLeaseMeta(localName)
	if err != nil {
		t.Fatalf("GetLeaseMeta failed: %v", err)
	}
	if !meta.Revoked {
		t.Errorf("expected meta.Revoked to be true")
	}
	if !mgr.IsLeaseExpired(localName) {
		t.Errorf("expected IsLeaseExpired to be true for revoked lease")
	}
}
