package share

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServerAutoApproveRequest(t *testing.T) {
	aimDir := t.TempDir()
	store := NewStore(aimDir)
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
