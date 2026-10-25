package share

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShareDB_InitAndMigrate(t *testing.T) {
	tempDir := t.TempDir()

	// Pre-create legacy audit file to test migration
	legacyFile := filepath.Join(tempDir, "share_audit.json")
	legacyEntries := []AuditEntry{
		{
			ID:        "aud_legacy_1",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Action:    "LEASE_GRANTED",
			Actor:     "remote-node",
			Profile:   "work",
			Agent:     "claude",
			Details:   "Legacy granted lease",
		},
	}
	data, err := json.Marshal(legacyEntries)
	if err != nil {
		t.Fatalf("failed to marshal legacy entries: %v", err)
	}
	if err := os.WriteFile(legacyFile, data, 0600); err != nil {
		t.Fatalf("failed to write legacy file: %v", err)
	}

	db, err := NewShareDB(tempDir)
	if err != nil {
		t.Fatalf("NewShareDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	logs, err := db.GetAuditLogs(ctx, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 migrated audit entry, got %d", len(logs))
	}
	if logs[0].ID != "aud_legacy_1" {
		t.Errorf("expected ID 'aud_legacy_1', got %s", logs[0].ID)
	}
}

func TestShareDB_RequestsLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	db, err := NewShareDB(tempDir)
	if err != nil {
		t.Fatalf("NewShareDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Save incoming request
	inReq := ShareRequestRecord{
		ID:             "req_in_1",
		Direction:      "incoming",
		PeerID:         "peer-laptop",
		PeerIP:         "10.8.19.100",
		ProfileName:    "prod",
		Agent:          "claude",
		RequestedHours: 4,
		Status:         "pending",
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	if err := db.SaveRequest(ctx, inReq); err != nil {
		t.Fatalf("SaveRequest failed: %v", err)
	}

	// 2. Save outgoing request
	outReq := ShareRequestRecord{
		ID:             "req_out_1",
		Direction:      "outgoing",
		PeerID:         "peer-laptop-2",
		PeerIP:         "10.8.19.101",
		ProfileName:    "dev",
		Agent:          "codex",
		RequestedHours: 2,
		Status:         "pending",
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	if err := db.SaveRequest(ctx, outReq); err != nil {
		t.Fatalf("SaveRequest failed: %v", err)
	}

	// 3. Query incoming pending
	inList, err := db.GetPendingIncomingRequests(ctx)
	if err != nil {
		t.Fatalf("GetPendingIncomingRequests failed: %v", err)
	}
	if len(inList) != 1 || inList[0].ID != "req_in_1" {
		t.Fatalf("expected [req_in_1], got %v", inList)
	}

	// 4. Query outgoing pending
	outList, err := db.GetPendingOutgoingRequests(ctx)
	if err != nil {
		t.Fatalf("GetPendingOutgoingRequests failed: %v", err)
	}
	if len(outList) != 1 || outList[0].ID != "req_out_1" {
		t.Fatalf("expected [req_out_1], got %v", outList)
	}

	// 5. Approve incoming request
	if err := db.UpdateRequestStatus(ctx, "req_in_1", "approved", "ls_123", ""); err != nil {
		t.Fatalf("UpdateRequestStatus failed: %v", err)
	}
	rec, err := db.GetRequest(ctx, "req_in_1")
	if err != nil || rec == nil {
		t.Fatalf("GetRequest failed: %v", err)
	}
	if rec.Status != "approved" || rec.LeaseID != "ls_123" {
		t.Errorf("expected approved with ls_123, got status=%s lease_id=%s", rec.Status, rec.LeaseID)
	}

	// Incoming list should now be empty
	inList2, _ := db.GetPendingIncomingRequests(ctx)
	if len(inList2) != 0 {
		t.Errorf("expected 0 pending incoming requests, got %d", len(inList2))
	}

	// 6. Cancel outgoing request
	cancelled, err := db.CancelRequest(ctx, "req_out_1")
	if err != nil {
		t.Fatalf("CancelRequest failed: %v", err)
	}
	if !cancelled {
		t.Fatalf("expected CancelRequest to succeed and return true")
	}
	recOut, err := db.GetRequest(ctx, "req_out_1")
	if err != nil || recOut == nil {
		t.Fatalf("GetRequest failed: %v", err)
	}
	if recOut.Status != "cancelled" {
		t.Errorf("expected cancelled, got %s", recOut.Status)
	}
}

func TestShareDB_TransitionRequestStatus(t *testing.T) {
	tempDir := t.TempDir()
	db, err := NewShareDB(tempDir)
	if err != nil {
		t.Fatalf("NewShareDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	req := ShareRequestRecord{
		ID:             "req_trans_1",
		Direction:      "incoming",
		PeerID:         "peer-laptop",
		PeerIP:         "10.8.19.100",
		ProfileName:    "prod",
		Agent:          "claude",
		RequestedHours: 4,
		Status:         "pending",
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	if err := db.SaveRequest(ctx, req); err != nil {
		t.Fatalf("SaveRequest failed: %v", err)
	}

	// First transition from pending to approved succeeds
	ok, err := db.TransitionRequestStatus(ctx, "req_trans_1", "pending", "approved", "ls_1", "")
	if err != nil {
		t.Fatalf("TransitionRequestStatus failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected transition to succeed")
	}

	// Concurrent/subsequent transition from pending fails
	ok2, err := db.TransitionRequestStatus(ctx, "req_trans_1", "pending", "approved", "ls_2", "")
	if err != nil {
		t.Fatalf("TransitionRequestStatus error: %v", err)
	}
	if ok2 {
		t.Fatalf("expected duplicate transition from pending to fail")
	}

	// Cancelling an already approved request does not overwrite it
	cancelled, err := db.CancelRequest(ctx, "req_trans_1")
	if err != nil {
		t.Fatalf("CancelRequest error: %v", err)
	}
	if cancelled {
		t.Fatalf("expected cancelling already approved request to return false")
	}
	rec, _ := db.GetRequest(ctx, "req_trans_1")
	if rec.Status != "approved" {
		t.Errorf("expected status to stay approved, got %s", rec.Status)
	}
}

func TestShareDB_AuditLogs(t *testing.T) {
	tempDir := t.TempDir()
	db, err := NewShareDB(tempDir)
	if err != nil {
		t.Fatalf("NewShareDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	entries := []AuditEntry{
		{Action: "LEASE_REQUESTED", Actor: "alice", Profile: "p1", Agent: "claude", Details: "req 1"},
		{Action: "LEASE_APPROVED", Actor: "local", Profile: "p1", Agent: "claude", Details: "app 1"},
		{Action: "LEASE_BORROWED", Actor: "bob", Profile: "p2", Agent: "codex", Details: "bor 1"},
		{Action: "LEASE_CANCELLED", Actor: "local", Profile: "p3", Agent: "agy", Details: "can 1"},
	}

	for _, e := range entries {
		if err := db.AppendAudit(ctx, e); err != nil {
			t.Fatalf("AppendAudit failed: %v", err)
		}
	}

	logs, err := db.GetAuditLogs(ctx, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs failed: %v", err)
	}
	if len(logs) != 4 {
		t.Fatalf("expected 4 logs, got %d", len(logs))
	}
	// First should be newest (LEASE_CANCELLED)
	if logs[0].Action != "LEASE_CANCELLED" {
		t.Errorf("expected newest log to be LEASE_CANCELLED, got %s", logs[0].Action)
	}
}

func TestShareDB_TimeFormatAndParse_Monotonicity(t *testing.T) {
	// 1. Zero time
	if got := formatDBTime(time.Time{}); got != "" {
		t.Errorf("expected empty string for zero time, got %q", got)
	}
	if got := parseDBTime(""); !got.IsZero() {
		t.Errorf("expected zero time for empty string, got %v", got)
	}

	// 2. Trailing zeros preserved (exactly 30 chars)
	t0 := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	s0 := formatDBTime(t0)
	if len(s0) != 30 || s0 != "2026-10-25T12:00:00.000000000Z" {
		t.Fatalf("expected 30 chars with 9 zeros, got %q (len %d)", s0, len(s0))
	}

	// 3. Strict monotonic string ordering vs nanosecond increments
	tA := t0.Add(100 * time.Millisecond) // .100000000 (ends in 8 trailing zeros)
	tB := tA.Add(1 * time.Nanosecond)    // .100000001
	sA := formatDBTime(tA)
	sB := formatDBTime(tB)

	if len(sA) != 30 || len(sB) != 30 {
		t.Fatalf("expected both strings to be 30 chars, got %d and %d", len(sA), len(sB))
	}
	if sA >= sB {
		t.Fatalf("expected sA < sB lexicographically, got %q vs %q", sA, sB)
	}

	// 4. Round-trip through parseDBTime
	parsedA := parseDBTime(sA)
	if !parsedA.Equal(tA) {
		t.Errorf("roundtrip failed: got %v, want %v", parsedA, tA)
	}
	parsedB := parseDBTime(sB)
	if !parsedB.Equal(tB) {
		t.Errorf("roundtrip failed: got %v, want %v", parsedB, tB)
	}
}

func TestShareDB_IdenticalTimestamps_TieBreaker(t *testing.T) {
	tempDir := t.TempDir()
	db, err := NewShareDB(tempDir)
	if err != nil {
		t.Fatalf("NewShareDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	fixedTime := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)

	for i := 1; i <= 5; i++ {
		err := db.AppendAudit(ctx, AuditEntry{
			ID:        fmt.Sprintf("aud_tie_%d", i),
			Timestamp: fixedTime,
			Action:    "LEASE_REQUESTED",
			Actor:     "local",
			Details:   fmt.Sprintf("Tie entry #%d", i),
		})
		if err != nil {
			t.Fatalf("AppendAudit failed: %v", err)
		}
	}

	logs, err := db.GetAuditLogs(ctx, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs failed: %v", err)
	}
	if len(logs) != 5 {
		t.Fatalf("expected 5 logs, got %d", len(logs))
	}

	// Must return reverse insertion order (LIFO) via rowid DESC
	for i, entry := range logs {
		expectedDetails := fmt.Sprintf("Tie entry #%d", 5-i)
		if entry.Details != expectedDetails {
			t.Errorf("logs[%d]: expected %q, got %q", i, expectedDetails, entry.Details)
		}
	}
}
