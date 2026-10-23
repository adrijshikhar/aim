package share

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

	// Continuous calendar interval spanning multiple days (e.g. yesterday 09:00 to tomorrow 18:00)
	spCalendarSpan := SchedulePolicy{
		Enabled:   true,
		Scope:     "global",
		Mode:      "auto",
		StartDate: now.Add(-24 * time.Hour).Format("2006-01-02"),
		EndDate:   now.Add(24 * time.Hour).Format("2006-01-02"),
		StartTime: "09:00",
		EndTime:   "18:00",
	}
	if !spCalendarSpan.IsActive(now) {
		t.Errorf("expected multi-day calendar span to remain active continuously, got false")
	}

	// Past calendar interval
	spCalendarExpired := SchedulePolicy{
		Enabled:   true,
		Scope:     "global",
		Mode:      "auto",
		StartDate: now.Add(-48 * time.Hour).Format("2006-01-02"),
		EndDate:   now.Add(-24 * time.Hour).Format("2006-01-02"),
		StartTime: "09:00",
		EndTime:   "18:00",
	}
	if spCalendarExpired.IsActive(now) {
		t.Errorf("expected past calendar span to be inactive, got true")
	}
}

func TestServerApproveRequiresLocalhost(t *testing.T) {
	aimDir := t.TempDir()
	store := NewStore(aimDir)
	_ = store.SetGlobalMode("request")
	_ = store.EnableProfile("work", "claude", false, 4*time.Hour)

	profilesDir := filepath.Join(aimDir, "profiles")
	server := NewServer(store, profilesDir, "alice-node", 4319)

	// Simulate remote request by setting RemoteAddr to a non-localhost IP
	req := httptest.NewRequest("POST", "/api/share/approve", strings.NewReader(`{"request_id":"req_123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.100:54321" // Remote peer on network
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for remote caller to /api/share/approve, got %d", w.Code)
	}

	// CSRF rejection: local caller with text/plain (simple CORS) must be rejected
	csrfReq := httptest.NewRequest("POST", "/api/share/approve", strings.NewReader(`{"request_id":"req_123"}`))
	csrfReq.Header.Set("Content-Type", "text/plain")
	csrfReq.RemoteAddr = "127.0.0.1:54321"
	csrfW := httptest.NewRecorder()
	server.Handler().ServeHTTP(csrfW, csrfReq)
	if csrfW.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for text/plain CSRF bypass, got %d", csrfW.Code)
	}

	// CSRF rejection: cross-origin Origin header must be rejected
	originReq := httptest.NewRequest("POST", "/api/share/approve", strings.NewReader(`{"request_id":"req_123"}`))
	originReq.Header.Set("Content-Type", "application/json")
	originReq.Header.Set("Origin", "https://malicious-website.com")
	originReq.RemoteAddr = "127.0.0.1:54321"
	originW := httptest.NewRecorder()
	server.Handler().ServeHTTP(originW, originReq)
	if originW.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for malicious Origin, got %d", originW.Code)
	}

	// CSRF rejection: cross-site Sec-Fetch-Site must be rejected
	sfsReq := httptest.NewRequest("POST", "/api/share/approve", strings.NewReader(`{"request_id":"req_123"}`))
	sfsReq.Header.Set("Content-Type", "application/json")
	sfsReq.Header.Set("Sec-Fetch-Site", "cross-site")
	sfsReq.RemoteAddr = "127.0.0.1:54321"
	sfsW := httptest.NewRecorder()
	server.Handler().ServeHTTP(sfsW, sfsReq)
	if sfsW.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for cross-site Sec-Fetch-Site, got %d", sfsW.Code)
	}

	// Verify legitimate local caller is allowed to attempt approve
	localReq := httptest.NewRequest("POST", "/api/share/approve", strings.NewReader(`{"request_id":"req_123"}`))
	localReq.Header.Set("Content-Type", "application/json")
	localReq.RemoteAddr = "127.0.0.1:54321"
	localW := httptest.NewRecorder()
	server.Handler().ServeHTTP(localW, localReq)
	// Not found (404) because req_123 doesn't exist, but NOT forbidden (403)!
	if localW.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for local caller, got %d", localW.Code)
	}
}

func TestServerPollLeaseEndpoint(t *testing.T) {
	aimDir := t.TempDir()
	store := NewStore(aimDir)
	_ = store.SetGlobalMode("request")
	_ = store.EnableProfile("work", "claude", false, 4*time.Hour)

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

	// 1. Borrower requests lease (returns pending)
	resp, err := client.RequestLease(ctx, "127.0.0.1", "work", "claude", 2)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.Status != "pending" {
		t.Fatalf("expected pending, got %s", resp.Status)
	}

	// 2. Poll while still pending
	pollResp, err := client.PollLease(ctx, "127.0.0.1", resp.RequestID)
	if err != nil {
		t.Fatalf("poll failed: %v", err)
	}
	if pollResp.Status != "pending" {
		t.Errorf("expected poll status pending, got %s", pollResp.Status)
	}

	// 3. Host approves request
	approved, err := server.ApproveRequest(resp.RequestID)
	if err != nil {
		t.Fatalf("failed to approve: %v", err)
	}
	if approved.Status != "approved" {
		t.Fatalf("expected approved, got %s", approved.Status)
	}

	// 4. Poll after approval -> returns approved lease with credentials!
	pollApproved, err := client.PollLease(ctx, "127.0.0.1", resp.RequestID)
	if err != nil {
		t.Fatalf("poll failed after approval: %v", err)
	}
	if pollApproved.Status != "approved" || pollApproved.Payload == nil {
		t.Fatalf("expected approved with payload, got: %+v", pollApproved)
	}
	if _, ok := pollApproved.Payload.Files[".claude/.credentials.json"]; !ok {
		t.Errorf("expected credential file in poll payload")
	}
}

func TestServerDenyRequestRejectionDoesNotHideApproval(t *testing.T) {
	aimDir := t.TempDir()
	store := NewStore(aimDir)
	_ = store.SetGlobalMode("request")
	profilesDir := t.TempDir()

	profileDir := filepath.Join(profilesDir, "work", ".claude")
	_ = os.MkdirAll(profileDir, 0700)
	_ = os.WriteFile(filepath.Join(profileDir, ".credentials.json"), []byte(`{"session":"valid"}`), 0600)

	_ = store.EnableProfile("work", "claude", false, 4*time.Hour)

	server := NewServer(store, profilesDir, "host-node", 0)
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	client := NewClient(0)
	client.baseURLOverride = ts.URL
	ctx := context.Background()

	// 1. Borrower requests lease (pending)
	resp, err := client.RequestLease(ctx, "127.0.0.1", "work", "claude", 2)
	if err != nil || resp.Status != "pending" {
		t.Fatalf("request failed: %v, resp: %+v", err, resp)
	}

	// 2. Host approves request
	approved, err := server.ApproveRequest(resp.RequestID)
	if err != nil || approved.Status != "approved" {
		t.Fatalf("approval failed: %v", err)
	}

	// 3. Simulate daemon restart with fresh memory state pointing to the same store
	server2 := NewServer(store, profilesDir, "host-node", 0)
	ts2 := httptest.NewServer(server2.Handler())
	defer ts2.Close()

	client2 := NewClient(0)
	client2.baseURLOverride = ts2.URL

	// 4. Calling DenyRequest on already-approved request must fail
	err = server2.DenyRequest(resp.RequestID)
	if err == nil {
		t.Fatalf("expected DenyRequest to fail on already approved request, but it succeeded")
	}

	// 5. Poll on server2 must still return approved, not falsely marked denied
	pollResp, err := client2.PollLease(ctx, "127.0.0.1", resp.RequestID)
	if err != nil {
		t.Fatalf("poll failed: %v", err)
	}
	if pollResp.Status != "approved" {
		t.Fatalf("expected poll status approved after rejected denial, got %s", pollResp.Status)
	}
}

func TestIsPrivateOrVPNIP(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{"127.0.0.1:4319", true},
		{"[::1]:4319", true},
		{"localhost:4319", true},
		{"10.8.19.251:4319", true},
		{"10.8.16.9:54321", true},
		{"192.168.1.50:4319", true},
		{"172.16.0.10:4319", true},
		{"100.64.1.2:4319", true},
		{"[fe80::1]:4319", true},
		{"[fe80::1%en0]:4319", true},
		{"[fe80::20c:29ff:fe4a:3b21%utun4]:51234", true},
		{"[fc00::1]:4319", true},
		{"8.8.8.8:4319", false},
		{"1.1.1.1:80", false},
		{"[2001:4860:4860::8888]:4319", false},
	}

	for _, tt := range tests {
		got := isPrivateOrVPNIP(tt.addr)
		if got != tt.want {
			t.Errorf("isPrivateOrVPNIP(%q) = %v, want %v", tt.addr, got, tt.want)
		}
	}
}

func TestPackageProfileCredentials_AgentScoping(t *testing.T) {
	profilesDir := t.TempDir()
	profDir := filepath.Join(profilesDir, "multi-agent")

	// Set up credentials for Claude, Codex, and Agy
	_ = os.MkdirAll(filepath.Join(profDir, ".claude"), 0700)
	_ = os.WriteFile(filepath.Join(profDir, ".claude", ".credentials.json"), []byte("claude-secret"), 0600)

	_ = os.MkdirAll(filepath.Join(profDir, ".codex"), 0700)
	_ = os.WriteFile(filepath.Join(profDir, ".codex", "auth.json"), []byte("codex-secret"), 0600)

	_ = os.MkdirAll(filepath.Join(profDir, ".gemini", "antigravity-cli"), 0700)
	_ = os.WriteFile(filepath.Join(profDir, ".gemini", "antigravity-cli", "antigravity-oauth-token"), []byte("agy-secret"), 0600)

	// 1. When agent == "claude" is requested, ONLY Claude credentials should be packaged
	payloadClaude, err := PackageProfileCredentials(profilesDir, "multi-agent", "claude")
	if err != nil {
		t.Fatalf("PackageProfileCredentials claude failed: %v", err)
	}
	if _, ok := payloadClaude.Files[".claude/.credentials.json"]; !ok {
		t.Errorf("expected claude credentials in payload")
	}
	if _, ok := payloadClaude.Files[".codex/auth.json"]; ok {
		t.Errorf("CRIT-1 failure: codex credentials leaked when only claude was requested!")
	}
	if _, ok := payloadClaude.Files[".gemini/antigravity-cli/antigravity-oauth-token"]; ok {
		t.Errorf("CRIT-1 failure: agy credentials leaked when only claude was requested!")
	}

	// 2. When agent == "codex" is requested, ONLY Codex credentials should be packaged
	payloadCodex, err := PackageProfileCredentials(profilesDir, "multi-agent", "codex")
	if err != nil {
		t.Fatalf("PackageProfileCredentials codex failed: %v", err)
	}
	if _, ok := payloadCodex.Files[".codex/auth.json"]; !ok {
		t.Errorf("expected codex credentials in payload")
	}
	if _, ok := payloadCodex.Files[".claude/.credentials.json"]; ok {
		t.Errorf("CRIT-1 failure: claude credentials leaked when only codex was requested!")
	}

	// 3. When agent == "all" is requested, ALL credentials should be packaged
	payloadAll, err := PackageProfileCredentials(profilesDir, "multi-agent", "all")
	if err != nil {
		t.Fatalf("PackageProfileCredentials all failed: %v", err)
	}
	if _, ok := payloadAll.Files[".claude/.credentials.json"]; !ok {
		t.Errorf("expected claude credentials in 'all' payload")
	}
	if _, ok := payloadAll.Files[".codex/auth.json"]; !ok {
		t.Errorf("expected codex credentials in 'all' payload")
	}
	if _, ok := payloadAll.Files[".gemini/antigravity-cli/antigravity-oauth-token"]; !ok {
		t.Errorf("expected agy credentials in 'all' payload")
	}
}
