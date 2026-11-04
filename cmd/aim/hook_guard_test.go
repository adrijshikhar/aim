package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/share"
)

func TestHookGuard_ActiveAndRevoked(t *testing.T) {
	tempDir := t.TempDir()
	profDir := filepath.Join(tempDir, "p2p-test-work")
	_ = os.MkdirAll(profDir, 0700)

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	expStr := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	sig := share.SignLease(priv, "ls_1", "lender1", "borrower1", "work", "codex", expStr)

	meta := share.LeaseMeta{
		LeaseID:        "ls_1",
		SourcePeer:     "lender1",
		SourceProfile:  "work",
		Agent:          "codex",
		ExpiresAt:      time.Now().Add(2 * time.Hour),
		LastVerifiedAt: time.Now(),
		Revoked:        false,
		Signature:      sig,
		BorrowerID:     "borrower1",
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), data, 0600)

	// Save pinned lender public key in peers.json
	store := share.NewStore(tempDir)
	_ = store.SavePeer(share.PeerInfo{
		ID:        "lender1",
		HostName:  "lender1",
		PublicKey: pub,
	})
	_ = store.Close()

	// 1. Active lease check -> exit code 0
	buf := new(bytes.Buffer)
	err := executeHookGuardForDir(tempDir, profDir, "p2p-test-work", nil, buf)
	if err != nil {
		t.Fatalf("expected active lease to pass, got err: %v", err)
	}

	// 2. Revoked lease check -> non-zero error & freeze banner
	meta.Revoked = true
	dataRevoked, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), dataRevoked, 0600)

	buf.Reset()
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", nil, buf)
	if err == nil {
		t.Fatalf("expected revoked lease to fail with exit error")
	}
	if !strings.Contains(buf.String(), "inactive") {
		t.Errorf("expected freeze message in output, got: %s", buf.String())
	}

	// 3. In-chat renewal tool whitelist while revoked -> passes!
	buf.Reset()
	renewalPayload := []byte(`{"tool_name": "aim_request_lease"}`)
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", renewalPayload, buf)
	if err != nil {
		t.Errorf("expected aim_request_lease to be whitelisted even when revoked, got err: %v", err)
	}

	// 4. In-chat status tool whitelist while revoked -> passes!
	buf.Reset()
	statusPayload := []byte(`{"name": "aim_lease_status"}`)
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", statusPayload, buf)
	if err != nil {
		t.Errorf("expected aim_lease_status to be whitelisted even when revoked, got err: %v", err)
	}

	// Also check nested tool_call format
	buf.Reset()
	nestedPayload := []byte(`{"tool_call": {"name": "aim_request_lease"}}`)
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", nestedPayload, buf)
	if err != nil {
		t.Errorf("expected nested tool_call aim_request_lease to be whitelisted, got: %v", err)
	}

	// Non-whitelisted tool while revoked -> halts with exit error
	buf.Reset()
	otherToolPayload := []byte(`{"tool_name": "write_file"}`)
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", otherToolPayload, buf)
	if err == nil {
		t.Fatalf("expected non-whitelisted tool to fail when revoked")
	}

	// 5. Expired lease check -> fails with freeze message
	meta.Revoked = false
	meta.ExpiresAt = time.Now().Add(-1 * time.Hour)
	dataExpired, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), dataExpired, 0600)

	buf.Reset()
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", nil, buf)
	if err == nil {
		t.Fatalf("expected expired lease to fail with exit error")
	}
	if !strings.Contains(buf.String(), "inactive") {
		t.Errorf("expected freeze message for expired lease, got: %s", buf.String())
	}

	// 6. Stale heartbeat check (> 5m) -> fails with freeze message
	meta.ExpiresAt = time.Now().Add(2 * time.Hour)
	meta.LastVerifiedAt = time.Now().Add(-6 * time.Minute)
	dataStale, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), dataStale, 0600)

	buf.Reset()
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", nil, buf)
	if err == nil {
		t.Fatalf("expected stale heartbeat lease to fail with exit error")
	}
	if !strings.Contains(buf.String(), "inactive") {
		t.Errorf("expected freeze message for stale heartbeat lease, got: %s", buf.String())
	}

	// 7. Cryptographic signature check (tampered signature) -> fails with freeze message
	meta.LastVerifiedAt = time.Now()
	meta.Signature = "bad-sig-123"
	dataTampered, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), dataTampered, 0600)

	buf.Reset()
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", nil, buf)
	if err == nil {
		t.Fatalf("expected tampered signature lease to fail with exit error")
	}
	if !strings.Contains(buf.String(), "inactive") {
		t.Errorf("expected freeze message for tampered signature lease, got: %s", buf.String())
	}

	// 8. Missing signature when pinned public key exists -> fails with freeze message
	meta.Signature = ""
	dataNoSig, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), dataNoSig, 0600)

	buf.Reset()
	err = executeHookGuardForDir(tempDir, profDir, "p2p-test-work", nil, buf)
	if err == nil {
		t.Fatalf("expected missing signature to fail when lender is pinned")
	}
	if !strings.Contains(buf.String(), "inactive") {
		t.Errorf("expected freeze message for missing signature, got: %s", buf.String())
	}
}

func TestHookGuard_MissingAndCorruptFiles(t *testing.T) {
	tempDir := t.TempDir()
	emptyProfDir := filepath.Join(tempDir, "missing-profile")
	_ = os.MkdirAll(emptyProfDir, 0700)

	buf := new(bytes.Buffer)
	// Missing .aim-lease.json
	err := executeHookGuardForDir(tempDir, emptyProfDir, "missing-profile", nil, buf)
	if err == nil || !strings.Contains(err.Error(), "lease file missing") {
		t.Fatalf("expected 'lease file missing' error, got: %v", err)
	}

	// Corrupt .aim-lease.json
	_ = os.WriteFile(filepath.Join(emptyProfDir, ".aim-lease.json"), []byte("{not-valid-json"), 0600)
	err = executeHookGuardForDir(tempDir, emptyProfDir, "missing-profile", nil, buf)
	if err == nil || !strings.Contains(err.Error(), "corrupt lease metadata") {
		t.Fatalf("expected 'corrupt lease metadata' error, got: %v", err)
	}
}

func TestHookGuard_CommandAndFlags(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("AIM_HOME", tempHome)

	profDir := filepath.Join(tempHome, "profiles", "test-lease")
	_ = os.MkdirAll(profDir, 0700)

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	store := share.NewStore(tempHome)
	_ = store.SavePeer(share.PeerInfo{
		ID:        "pinned-peer",
		HostName:  "pinned-peer",
		PublicKey: pub,
	})
	_ = store.Close()

	expiresAt := time.Now().Add(1 * time.Hour)
	expStr := expiresAt.UTC().Format(time.RFC3339)
	h, _ := os.Hostname()
	sig := share.SignLease(priv, "ls_2", "pinned-peer", h, "work", "codex", expStr)

	meta := share.LeaseMeta{
		LeaseID:        "ls_2",
		SourcePeer:     "pinned-peer",
		SourceProfile:  "work",
		Agent:          "codex",
		ExpiresAt:      expiresAt,
		LastVerifiedAt: time.Now(),
		Revoked:        false,
		Signature:      sig,
		BorrowerID:     h,
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), data, 0600)

	cmd := newHookCmd()
	cmd.SetArgs([]string{"guard", "test-lease", "--agent", "codex"})
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected hook guard command to execute successfully, got: %v", err)
	}

	// Calling runHookGuard with no args returns error
	if err := runHookGuard(nil, nil, nil); err == nil {
		t.Errorf("expected error when args is empty")
	}
}

func BenchmarkHookGuard_Active(b *testing.B) {
	tempHome := b.TempDir()
	profDir := filepath.Join(tempHome, "profiles", "bench-prof")
	_ = os.MkdirAll(profDir, 0700)

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	store := share.NewStore(tempHome)
	_ = store.SavePeer(share.PeerInfo{
		ID:        "bench-peer",
		HostName:  "bench-peer",
		PublicKey: pub,
	})
	_ = store.Close()

	expiresAt := time.Now().Add(2 * time.Hour)
	expStr := expiresAt.UTC().Format(time.RFC3339)
	h, _ := os.Hostname()
	sig := share.SignLease(priv, "ls_bench", "bench-peer", h, "bench-prof", "codex", expStr)

	meta := share.LeaseMeta{
		LeaseID:        "ls_bench",
		SourcePeer:     "bench-peer",
		SourceProfile:  "bench-prof",
		Agent:          "codex",
		ExpiresAt:      expiresAt,
		LastVerifiedAt: time.Now(),
		Revoked:        false,
		Signature:      sig,
		BorrowerID:     h,
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), data, 0600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := executeHookGuardForDir(tempHome, profDir, "bench-prof", nil, io.Discard)
		if err != nil {
			b.Fatalf("hook guard benchmark failed: %v", err)
		}
	}
}
