package mcp

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/share"
)

func TestShareMCPServer_ListTools(t *testing.T) {
	tempDir := t.TempDir()
	server := NewShareMCPServer(tempDir, tempDir, "Adrijs-MBP-bbrouter:office")

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}` + "\n"
	out := new(bytes.Buffer)

	ctx := context.Background()
	err := server.HandleMessage(ctx, []byte(input), out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := out.String()
	if !strings.Contains(resp, "aim_request_lease") || !strings.Contains(resp, "aim_lease_status") {
		t.Errorf("expected MCP server to list aim_request_lease and aim_lease_status tools, got: %s", resp)
	}
}

func TestShareMCPServer_InitializeAndPing(t *testing.T) {
	tempDir := t.TempDir()
	server := NewShareMCPServer(tempDir, tempDir, "host:prof")

	ctx := context.Background()

	// Initialize
	initInput := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	out := new(bytes.Buffer)
	if err := server.HandleMessage(ctx, []byte(initInput), out); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if !strings.Contains(out.String(), "aim-share") || !strings.Contains(out.String(), "2024-11-05") {
		t.Errorf("unexpected initialize response: %s", out.String())
	}

	// Ping
	out.Reset()
	pingInput := `{"jsonrpc":"2.0","id":2,"method":"ping","params":{}}` + "\n"
	if err := server.HandleMessage(ctx, []byte(pingInput), out); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
	var pingResp JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &pingResp); err != nil {
		t.Fatalf("failed to parse ping response: %v", err)
	}
	if pingResp.Error != nil {
		t.Errorf("expected nil error on ping, got: %v", pingResp.Error)
	}
}

func TestShareMCPServer_CallStatus(t *testing.T) {
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	profName := "peer1:testprof"
	profDir := filepath.Join(profilesDir, profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	exp := time.Now().Add(2 * time.Hour)
	meta := share.LeaseMeta{
		LeaseID:        "ls_987",
		SourcePeer:     "peer1",
		SourceProfile:  "testprof",
		Agent:          "codex",
		GrantedAt:      time.Now(),
		ExpiresAt:      exp,
		LastVerifiedAt: time.Now(),
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), metaBytes, 0600); err != nil {
		t.Fatal(err)
	}

	server := NewShareMCPServer(tempDir, profilesDir, profName)

	input := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"aim_lease_status","arguments":{}}}` + "\n"
	out := new(bytes.Buffer)

	ctx := context.Background()
	if err := server.HandleMessage(ctx, []byte(input), out); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}

	resp := out.String()
	if !strings.Contains(resp, "ls_987") {
		t.Errorf("expected response to contain lease ID ls_987, got: %s", resp)
	}
	if !strings.Contains(resp, "Remaining Duration") && !strings.Contains(resp, "remaining") {
		t.Errorf("expected response to contain remaining duration, got: %s", resp)
	}
	if !strings.Contains(resp, "peer1") {
		t.Errorf("expected response to contain peer1, got: %s", resp)
	}

	// Test missing lease file returns isError: true
	serverMissing := NewShareMCPServer(tempDir, profilesDir, "peer1:nonexistent")
	out.Reset()
	if err := serverMissing.HandleMessage(ctx, []byte(input), out); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	var errResp JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	resMap, ok := errResp.Result.(map[string]any)
	if !ok || resMap["isError"] != true {
		t.Errorf("expected isError: true for missing lease, got: %s", out.String())
	}
}

func TestShareMCPServer_CallRequestLease(t *testing.T) {
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	profName := "peer1:testprof"
	profDir := filepath.Join(profilesDir, profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	initialExp := time.Now().Add(1 * time.Hour)
	meta := share.LeaseMeta{
		LeaseID:        "ls_initial",
		SourcePeer:     "peer1",
		SourceProfile:  "testprof",
		Agent:          "codex",
		GrantedAt:      time.Now(),
		ExpiresAt:      initialExp,
		LastVerifiedAt: time.Now(),
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), metaBytes, 0600); err != nil {
		t.Fatal(err)
	}

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	borrowerHost, _ := os.Hostname()
	if borrowerHost == "" {
		borrowerHost = "borrower"
	}

	store := share.NewStore(tempDir)
	_ = store.SavePeer(share.PeerInfo{
		ID:        "peer1",
		HostName:  "peer1",
		PublicKey: pub,
	})
	_ = store.Close()

	newExp := time.Now().Add(4 * time.Hour).Truncate(time.Second)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/share/request" {
			var req share.LeaseRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.RequestedHours != 4 {
				t.Errorf("expected 4 requested hours, got %d", req.RequestedHours)
			}
			sig := share.SignLease(priv, "ls_renewed_456", "peer1", borrowerHost, "testprof", "codex", newExp.UTC().Format(time.RFC3339))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(share.LeaseResponse{
				Status:       "approved",
				LeaseID:      "ls_renewed_456",
				ProfileName:  "testprof",
				Agent:        "codex",
				ExpiresAt:    newExp,
				LenderNodeID: "peer1",
				BorrowerID:   borrowerHost,
				Signature:    sig,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := share.NewClient(4319)
	client.SetBaseURLOverride(mockServer.URL)

	server := NewShareMCPServer(tempDir, profilesDir, profName)
	server.SetClient(client)

	input := `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"aim_request_lease","arguments":{"hours":4}}}` + "\n"
	out := new(bytes.Buffer)

	ctx := context.Background()
	if err := server.HandleMessage(ctx, []byte(input), out); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v, raw: %s", err, out.String())
	}
	if resp.Error != nil {
		t.Fatalf("expected nil RPC error, got: %v", resp.Error)
	}

	resMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got %T", resp.Result)
	}
	if resMap["isError"] == true {
		t.Fatalf("tool execution returned error: %s", out.String())
	}

	respStr := out.String()
	if !strings.Contains(respStr, "Successfully renewed lease") {
		t.Errorf("expected success message in response, got: %s", respStr)
	}
	if !strings.Contains(respStr, "unfrozen") {
		t.Errorf("expected note about unfrozen sessions, got: %s", respStr)
	}

	// Verify .aim-lease.json was updated with new lease ID
	updatedData, err := os.ReadFile(filepath.Join(profDir, ".aim-lease.json"))
	if err != nil {
		t.Fatalf("failed to read updated lease file: %v", err)
	}
	var updatedMeta share.LeaseMeta
	if err := json.Unmarshal(updatedData, &updatedMeta); err != nil {
		t.Fatalf("failed to parse updated lease metadata: %v", err)
	}
	if updatedMeta.LeaseID != "ls_renewed_456" {
		t.Errorf("expected lease ID ls_renewed_456, got: %s", updatedMeta.LeaseID)
	}
}

func TestShareMCPServer_CallRequestLease_Validation(t *testing.T) {
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	profName := "peer1:testprof"

	server := NewShareMCPServer(tempDir, profilesDir, profName)

	ctx := context.Background()

	// hours = 0
	input0 := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"aim_request_lease","arguments":{"hours":0}}}` + "\n"
	out := new(bytes.Buffer)
	_ = server.HandleMessage(ctx, []byte(input0), out)
	if !strings.Contains(out.String(), `"isError":true`) {
		t.Errorf("expected hours=0 to fail validation, got: %s", out.String())
	}

	// hours = 73
	out.Reset()
	input73 := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"aim_request_lease","arguments":{"hours":73}}}` + "\n"
	_ = server.HandleMessage(ctx, []byte(input73), out)
	if !strings.Contains(out.String(), `"isError":true`) {
		t.Errorf("expected hours=73 to fail validation, got: %s", out.String())
	}
}

func TestShareMCPServer_CallRequestLease_PendingPoll(t *testing.T) {
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	profName := "peer1:testprof"
	profDir := filepath.Join(profilesDir, profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	meta := share.LeaseMeta{
		LeaseID:        "ls_initial",
		SourcePeer:     "peer1",
		SourceProfile:  "testprof",
		Agent:          "codex",
		GrantedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(1 * time.Hour),
		LastVerifiedAt: time.Now(),
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), metaBytes, 0600)

	pollPub, pollPriv, _ := ed25519.GenerateKey(rand.Reader)
	pollBorrowerHost, _ := os.Hostname()
	if pollBorrowerHost == "" {
		pollBorrowerHost = "borrower"
	}

	pollStore := share.NewStore(tempDir)
	_ = pollStore.SavePeer(share.PeerInfo{
		ID:        "peer1",
		HostName:  "peer1",
		PublicKey: pollPub,
	})
	_ = pollStore.Close()

	pollCount := 0
	pollExp := time.Now().Add(3 * time.Hour)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/share/request" {
			_ = json.NewEncoder(w).Encode(share.LeaseResponse{
				Status:    "pending",
				RequestID: "req_poll_123",
			})
			return
		}
		if r.URL.Path == "/api/share/poll" {
			pollCount++
			if pollCount < 2 {
				_ = json.NewEncoder(w).Encode(share.LeaseResponse{
					Status:    "pending",
					RequestID: "req_poll_123",
				})
				return
			}
			pollSig := share.SignLease(pollPriv, "ls_polled_999", "peer1", pollBorrowerHost, "testprof", "codex", pollExp.UTC().Format(time.RFC3339))
			_ = json.NewEncoder(w).Encode(share.LeaseResponse{
				Status:       "approved",
				LeaseID:      "ls_polled_999",
				ProfileName:  "testprof",
				Agent:        "codex",
				ExpiresAt:    pollExp,
				LenderNodeID: "peer1",
				BorrowerID:   pollBorrowerHost,
				Signature:    pollSig,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := share.NewClient(4319)
	client.SetBaseURLOverride(mockServer.URL)

	server := NewShareMCPServer(tempDir, profilesDir, profName)
	server.SetClient(client)

	input := `{"jsonrpc":"2.0","id":15,"method":"tools/call","params":{"name":"aim_request_lease","arguments":{"hours":3}}}` + "\n"
	out := new(bytes.Buffer)

	ctx := context.Background()
	if err := server.HandleMessage(ctx, []byte(input), out); err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	if !strings.Contains(out.String(), "Successfully renewed lease") {
		t.Errorf("expected success after poll, got: %s", out.String())
	}
}

func TestShareMCPServer_ServeStdio(t *testing.T) {
	tempDir := t.TempDir()
	server := NewShareMCPServer(tempDir, tempDir, "peer1:test")

	input := `{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"

	in := strings.NewReader(input)
	out := new(bytes.Buffer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.ServeStdio(ctx, in, out); err != nil {
		t.Fatalf("ServeStdio returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 response lines, got %d: %s", len(lines), out.String())
	}
}
