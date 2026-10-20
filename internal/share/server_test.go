package share

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServerAutoApproveRequest(t *testing.T) {
	aimDir := t.TempDir()
	store := NewStore(aimDir)
	_ = store.SetGlobalMode("auto")
	_ = store.EnableProfile("work", "claude", true, 4*time.Hour)

	// Seed dummy credentials in profile
	profilesDir := filepath.Join(aimDir, "profiles")
	workDir := filepath.Join(profilesDir, "work", ".claude")
	_ = os.MkdirAll(workDir, 0700)
	_ = os.WriteFile(filepath.Join(workDir, ".credentials.json"), []byte(`{"token":"secret-token"}`), 0600)

	server := NewServer(store, profilesDir, "alice-node", 4319)
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	client := NewClient(4319)
	client.baseURLOverride = ts.URL

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.RequestLease(ctx, "127.0.0.1", "work", "claude", 2)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.Status != "approved" {
		t.Fatalf("expected status approved, got %s", resp.Status)
	}
	if resp.Payload == nil || resp.Payload.Files == nil {
		t.Fatalf("expected payload files, got nil")
	}
	b64 := resp.Payload.Files[".claude/.credentials.json"]
	decoded, _ := base64.StdEncoding.DecodeString(b64)
	if string(decoded) != `{"token":"secret-token"}` {
		t.Errorf("expected secret-token, got %s", string(decoded))
	}
}

func TestServerManualPendingRequestAndApproval(t *testing.T) {
	aimDir := t.TempDir()
	store := NewStore(aimDir)
	_ = store.SetGlobalMode("request")
	_ = store.EnableProfile("work", "claude", false, 4*time.Hour) // autoApprove = false

	profilesDir := filepath.Join(aimDir, "profiles")
	workDir := filepath.Join(profilesDir, "work", ".claude")
	_ = os.MkdirAll(workDir, 0700)
	_ = os.WriteFile(filepath.Join(workDir, ".credentials.json"), []byte(`{"token":"secret-token"}`), 0600)

	server := NewServer(store, profilesDir, "alice-node", 4319)
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	client := NewClient(4319)
	client.baseURLOverride = ts.URL

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.RequestLease(ctx, "127.0.0.1", "work", "claude", 2)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.Status != "pending" || resp.RequestID == "" {
		t.Fatalf("expected status pending with request ID, got: %+v", resp)
	}

	// Verify request is listed in pending
	pending := server.GetPendingRequests()
	if len(pending) != 1 || pending[0].RequestID != resp.RequestID {
		t.Fatalf("expected 1 pending request, got %+v", pending)
	}

	// Approve request
	approvedResp, err := server.ApproveRequest(resp.RequestID)
	if err != nil {
		t.Fatalf("failed to approve: %v", err)
	}
	if approvedResp.Status != "approved" || approvedResp.Payload == nil {
		t.Errorf("expected approved response with payload: %+v", approvedResp)
	}
}

func TestServerCodexCredentials(t *testing.T) {
	aimDir := t.TempDir()
	store := NewStore(aimDir)
	_ = store.SetGlobalMode("auto")
	_ = store.EnableProfile("office", "codex", true, 4*time.Hour)

	profilesDir := filepath.Join(aimDir, "profiles")
	officeCodexDir := filepath.Join(profilesDir, "office", ".codex")
	_ = os.MkdirAll(officeCodexDir, 0700)
	_ = os.WriteFile(filepath.Join(officeCodexDir, "auth.json"), []byte(`{"access_token":"codex-tok"}`), 0600)
	_ = os.WriteFile(filepath.Join(officeCodexDir, "config.toml"), []byte(`model = "o3-mini"`), 0600)

	server := NewServer(store, profilesDir, "bob-node", 4319)
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	client := NewClient(4319)
	client.baseURLOverride = ts.URL

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.RequestLease(ctx, "127.0.0.1", "office", "codex", 2)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.Status != "approved" || resp.Payload == nil {
		t.Fatalf("expected approved response with payload: %+v", resp)
	}
	if _, ok := resp.Payload.Files[".codex/auth.json"]; !ok {
		t.Errorf("expected .codex/auth.json in files, got: %+v", resp.Payload.Files)
	}
	if _, ok := resp.Payload.Files[".codex/config.toml"]; !ok {
		t.Errorf("expected .codex/config.toml in files, got: %+v", resp.Payload.Files)
	}
}

func TestServerGlobalModeOffEnforcement(t *testing.T) {
	aimDir := t.TempDir()
	store := NewStore(aimDir)
	_ = store.SetGlobalMode("off") // Explicitly OFF
	_ = store.EnableProfile("office", "codex", true, 4*time.Hour)

	profilesDir := filepath.Join(aimDir, "profiles")
	server := NewServer(store, profilesDir, "bob-node", 4319)
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	client := NewClient(4319)
	client.baseURLOverride = ts.URL

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Should be rejected with 403 Forbidden
	_, err := client.RequestLease(ctx, "127.0.0.1", "office", "codex", 2)
	if err == nil {
		t.Fatalf("expected lease request to fail when global mode is off, but succeeded")
	}

	// Should advertise 0 profiles via ping
	disc := NewDiscovery(store, 4319)
	peer, err := disc.PingPeerWithURL(ctx, ts.URL)
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
	if len(peer.Profiles) != 0 || len(peer.SharedProfiles) != 0 {
		t.Fatalf("expected 0 profiles when global mode is off, got %d profiles, %d shared", len(peer.Profiles), len(peer.SharedProfiles))
	}
}

func TestSchedulePolicyActive(t *testing.T) {
	now := time.Now()
	// Schedule from 1 hour ago to 1 hour in the future
	startH := (now.Hour() - 1 + 24) % 24
	endH := (now.Hour() + 1) % 24
	sp := SchedulePolicy{
		Enabled:   true,
		Scope:     "global",
		Mode:      "auto",
		StartTime: fmt.Sprintf("%02d:00", startH),
		EndTime:   fmt.Sprintf("%02d:00", endH),
	}
	if !sp.IsActive(now) {
		t.Errorf("expected schedule to be active at %v, got false", now)
	}

	// Inactive schedule (1 hour in future to 2 hours in future)
	futureStart := (now.Hour() + 1) % 24
	futureEnd := (now.Hour() + 2) % 24
	spFuture := SchedulePolicy{
		Enabled:   true,
		Scope:     "global",
		Mode:      "auto",
		StartTime: fmt.Sprintf("%02d:00", futureStart),
		EndTime:   fmt.Sprintf("%02d:00", futureEnd),
	}
	if spFuture.IsActive(now) {
		t.Errorf("expected future schedule to be inactive, got true")
	}
}
