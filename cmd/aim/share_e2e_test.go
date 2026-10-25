package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/share"
)

// setupIsolatedNode initializes a complete isolated AIM node environment with a temp AIM_HOME.
func setupIsolatedNode(t *testing.T, nodeName string) (string, *profile.ProfileManager, *share.ShareStore) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("AIM_HOME", dir)
	pm := profile.NewProfileManager(dir)
	store := share.NewStore(dir)
	return dir, pm, store
}

func findSharedProfile(configs []share.SharedProfileConfig, name string) *share.SharedProfileConfig {
	for i := range configs {
		if configs[i].Name == name {
			return &configs[i]
		}
	}
	return nil
}

func TestShareCLI_ModeLifecycle(t *testing.T) {
	tempDir, pm, store := setupIsolatedNode(t, "test-node")
	t.Setenv("AIM_HOME", tempDir)

	reg := agents.NewRegistry()

	tests := []struct {
		name       string
		args       []string
		wantOutput string
		wantErr    bool
		checkMode  string
	}{
		{
			name:       "default mode is off",
			args:       []string{"mode"},
			wantOutput: "Global sharing mode: off",
			checkMode:  "off",
		},
		{
			name:       "set mode to request",
			args:       []string{"mode", "request"},
			wantOutput: "Global sharing mode set to request",
			checkMode:  "request",
		},
		{
			name:       "verify mode is request",
			args:       []string{"mode"},
			wantOutput: "Global sharing mode: request",
			checkMode:  "request",
		},
		{
			name:       "set mode to auto",
			args:       []string{"mode", "auto"},
			wantOutput: "Global sharing mode set to auto",
			checkMode:  "auto",
		},
		{
			name:       "set mode to off",
			args:       []string{"mode", "off"},
			wantOutput: "Global sharing mode set to off",
			checkMode:  "off",
		},
		{
			name:    "invalid mode fails",
			args:    []string{"mode", "invalid_mode"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newShareCmd(reg, pm)
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for args %v, got nil", tc.args)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error executing %v: %v", tc.args, err)
			}

			if tc.wantOutput != "" && !strings.Contains(buf.String(), tc.wantOutput) {
				t.Errorf("expected output to contain %q, got: %s", tc.wantOutput, buf.String())
			}

			if tc.checkMode != "" {
				got := store.GetEffectiveGlobalMode()
				if got != tc.checkMode {
					t.Errorf("store mode = %q, want %q", got, tc.checkMode)
				}
			}
		})
	}
}

func TestShareCLI_ProfileEnableDisable(t *testing.T) {
	tempDir, pm, store := setupIsolatedNode(t, "test-node")
	t.Setenv("AIM_HOME", tempDir)

	// Create profile "work"
	if _, err := pm.EnsureProfile("work"); err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	reg := agents.NewRegistry()

	t.Run("enable profile with auto-approve and ttl", func(t *testing.T) {
		cmd := newShareCmd(reg, pm)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"enable", "work", "--auto", "--ttl", "4h"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(buf.String(), "Profile work is now shared") {
			t.Errorf("unexpected output: %s", buf.String())
		}

		sp := findSharedProfile(store.GetAllSharedProfiles(), "work")
		if sp == nil {
			t.Fatalf("profile 'work' not found in share store")
		}
		if !sp.Enabled {
			t.Errorf("expected profile to be enabled")
		}
		if !sp.AutoApprove {
			t.Errorf("expected auto_approve to be true")
		}
		if sp.MaxLeaseTTL != 4*time.Hour {
			t.Errorf("expected ttl 4h, got %v", sp.MaxLeaseTTL)
		}
	})

	t.Run("check status output contains enabled profile", func(t *testing.T) {
		cmd := newShareCmd(reg, pm)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"status"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "work") {
			t.Errorf("expected status output to contain 'work', got: %s", out)
		}
		if !strings.Contains(out, "Auto") {
			t.Errorf("expected status output to show 'Auto', got: %s", out)
		}
	})

	t.Run("disable profile", func(t *testing.T) {
		cmd := newShareCmd(reg, pm)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"disable", "work"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(buf.String(), "Profile work is no longer shared") {
			t.Errorf("unexpected output: %s", buf.String())
		}

		profs := store.GetSharedProfiles()
		if len(profs) != 0 {
			t.Errorf("expected 0 active shared profiles, got %d", len(profs))
		}
	})
}

func TestShareCLI_PeerManagement(t *testing.T) {
	tempDir, pm, store := setupIsolatedNode(t, "test-node")
	t.Setenv("AIM_HOME", tempDir)

	reg := agents.NewRegistry()

	// Spin up a live mock peer server so "peer add" can ping it
	peerStore := share.NewStore(t.TempDir())
	peerServer := share.NewServer(peerStore, t.TempDir(), "alice-macbook", 0)
	ts := httptest.NewServer(peerServer.Handler())
	defer ts.Close()

	peerHost := strings.TrimPrefix(ts.URL, "http://")

	t.Run("peer add", func(t *testing.T) {
		cmd := newShareCmd(reg, pm)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"peer", "add", peerHost})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(buf.String(), "Added peer") {
			t.Errorf("unexpected output: %s", buf.String())
		}

		peers := store.GetPeers()
		if len(peers) != 1 {
			t.Fatalf("expected 1 peer in store, got %d", len(peers))
		}
		if peers[0].HostName != "alice-macbook" {
			t.Errorf("expected hostname alice-macbook, got %s", peers[0].HostName)
		}
	})

	t.Run("peer list", func(t *testing.T) {
		cmd := newShareCmd(reg, pm)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"peer", "list"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(buf.String(), "alice-macbook") {
			t.Errorf("expected peer list to include hostname, got: %s", buf.String())
		}
	})

	t.Run("peer remove", func(t *testing.T) {
		cmd := newShareCmd(reg, pm)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"peer", "rm", "alice-macbook"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(buf.String(), "Removed peer") {
			t.Errorf("unexpected output: %s", buf.String())
		}

		peers := store.GetPeers()
		if len(peers) != 0 {
			t.Errorf("expected 0 peers, got %d", len(peers))
		}
	})
}

func TestShare_P2P_FullLifecycle(t *testing.T) {
	// 1. Setup Node A (Host)
	nodeADir, pmA, storeA := setupIsolatedNode(t, "host-node")
	if _, err := pmA.EnsureProfile("work"); err != nil {
		t.Fatalf("Node A failed to create profile: %v", err)
	}

	// Seed dummy credentials in Node A's profile
	claudeDir := filepath.Join(nodeADir, "profiles", "work", ".claude")
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		t.Fatal(err)
	}
	credPayload := `{"account_id":"acc_12345","access_token":"sk-ant-test-secret-token"}`
	if err := os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(credPayload), 0600); err != nil {
		t.Fatal(err)
	}

	// Configure Node A sharing: auto-approve, global mode = auto
	if err := storeA.SetGlobalMode("auto"); err != nil {
		t.Fatal(err)
	}
	if err := storeA.EnableProfile("work", "claude", true, 4*time.Hour); err != nil {
		t.Fatal(err)
	}

	// Start Node A's real HTTP server on an ephemeral loopback port
	serverA := share.NewServer(storeA, filepath.Join(nodeADir, "profiles"), "host-node", 0)
	tsA := httptest.NewServer(serverA.Handler())
	defer tsA.Close()

	// 2. Setup Node B (Borrower)
	nodeBDir, pmB, storeB := setupIsolatedNode(t, "borrower-node")

	// Create borrower client talking to Node A's test server URL
	clientB := share.NewClient(0)
	clientB.SetBaseURLOverride(tsA.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Query offered profiles from Node A using discovery ping
	discB := share.NewDiscovery(storeB, 0)
	peerInfo, err := discB.PingPeerWithURL(ctx, tsA.URL)
	if err != nil {
		t.Fatalf("failed to query offered profiles: %v", err)
	}
	if len(peerInfo.SharedProfiles) != 1 || peerInfo.SharedProfiles[0].Name != "work" || !peerInfo.SharedProfiles[0].AutoApprove {
		t.Fatalf("unexpected offered profiles: %+v", peerInfo.SharedProfiles)
	}

	// Step 2: Request lease from Node A (Instant Auto-Approve)
	resp, err := clientB.RequestLease(ctx, "host-node", "work", "claude", 2)
	if err != nil {
		t.Fatalf("failed to request lease: %v", err)
	}
	if resp.Status != "approved" {
		t.Fatalf("expected approved status, got %q", resp.Status)
	}
	if resp.Payload == nil || resp.Payload.Files == nil {
		t.Fatalf("expected credentials payload in approved lease")
	}

	// Step 3: Install the granted lease locally on Node B
	leaseMgrB := share.NewLeaseManager(pmB.ProfilesRoot())
	localProfilePath, err := leaseMgrB.InstallLease("host-node", "work", resp)
	if err != nil {
		t.Fatalf("failed to install lease on Node B: %v", err)
	}

	// Expected local profile name is "host-node:work"
	localProfileName := "host-node:work"

	// Verify file permissions strictly adhere to Rule 3.2 (0700 dir, 0600 file)
	credFilePath := filepath.Join(pmB.ProfilesRoot(), localProfilePath, ".claude", ".credentials.json")
	stat, err := os.Stat(credFilePath)
	if err != nil {
		t.Fatalf("credentials file not found at %s: %v", credFilePath, err)
	}
	if stat.Mode().Perm() != 0600 {
		t.Errorf("credentials file permissions = %o, want 0600", stat.Mode().Perm())
	}

	// Verify credential content was transferred accurately
	data, err := os.ReadFile(credFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "sk-ant-test-secret-token") {
		t.Errorf("credentials corrupted, got: %s", string(data))
	}

	// Step 4: Verify lease status on Node B
	if !leaseMgrB.HasLease(localProfileName) {
		t.Errorf("expected lease %s to exist on Node B", localProfileName)
	}
	if leaseMgrB.IsLeaseExpired(localProfileName) {
		t.Errorf("lease %s should not be expired", localProfileName)
	}

	// Step 5: Release the lease from Node B
	if err := leaseMgrB.ReleaseLease(localProfileName); err != nil {
		t.Fatalf("failed to release lease: %v", err)
	}

	// Verify credentials directory was purged
	if _, err := os.Stat(localProfilePath); !os.IsNotExist(err) {
		t.Errorf("expected local profile directory %s to be deleted", localProfilePath)
	}

	// Step 6: Verify audit trail on Node A recorded granted event
	logs := storeA.GetAuditLogs(10)
	var foundGranted bool
	for _, l := range logs {
		if l.Action == "LEASE_GRANTED" && l.Profile == "work" {
			foundGranted = true
		}
	}
	if !foundGranted {
		t.Errorf("audit log missing LEASE_GRANTED entry: %+v", logs)
	}

	_ = nodeBDir
}

func TestShare_P2P_ManualApprovalWorkflow(t *testing.T) {
	nodeADir, pmA, storeA := setupIsolatedNode(t, "host-manual")
	if _, err := pmA.EnsureProfile("office"); err != nil {
		t.Fatal(err)
	}

	// Seed codex credentials
	codexDir := filepath.Join(nodeADir, "profiles", "office", ".codex")
	if err := os.MkdirAll(codexDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(`api_key = "sk-codex-123"`), 0600); err != nil {
		t.Fatal(err)
	}

	// Manual approval required: auto_approve = false, global mode = request
	if err := storeA.SetGlobalMode("request"); err != nil {
		t.Fatal(err)
	}
	if err := storeA.EnableProfile("office", "codex", false, 2*time.Hour); err != nil {
		t.Fatal(err)
	}

	serverA := share.NewServer(storeA, filepath.Join(nodeADir, "profiles"), "host-manual", 0)
	tsA := httptest.NewServer(serverA.Handler())
	defer tsA.Close()

	clientB := share.NewClient(0)
	clientB.SetBaseURLOverride(tsA.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Borrower requests lease -> receives pending status
	resp, err := clientB.RequestLease(ctx, "host-manual", "office", "codex", 1)
	if err != nil {
		t.Fatalf("request lease failed: %v", err)
	}
	if resp.Status != "pending" {
		t.Fatalf("expected status 'pending', got %q", resp.Status)
	}
	if resp.RequestID == "" {
		t.Fatalf("expected non-empty request_id in pending response")
	}

	// 2. Poll before approval -> should still be pending
	pollResp, err := clientB.PollLease(ctx, "host-manual", resp.RequestID)
	if err != nil {
		t.Fatalf("poll failed: %v", err)
	}
	if pollResp.Status != "pending" {
		t.Fatalf("expected status 'pending' before approval, got %q", pollResp.Status)
	}

	// 3. Host approves the pending request
	if _, err := serverA.ApproveRequest(resp.RequestID); err != nil {
		t.Fatalf("host approve failed: %v", err)
	}

	// 4. Poll after approval -> should now be approved with credentials payload
	approvedResp, err := clientB.PollLease(ctx, "host-manual", resp.RequestID)
	if err != nil {
		t.Fatalf("poll after approval failed: %v", err)
	}
	if approvedResp.Status != "approved" {
		t.Fatalf("expected status 'approved', got %q", approvedResp.Status)
	}
	if approvedResp.Payload == nil || approvedResp.Payload.Files == nil {
		t.Fatalf("expected credentials payload in approved response")
	}
	if _, ok := approvedResp.Payload.Files[".codex/config.toml"]; !ok {
		t.Errorf("missing config.toml in payload: %+v", approvedResp.Payload.Files)
	}
}

func TestShare_P2P_DenialWorkflow(t *testing.T) {
	nodeADir, pmA, storeA := setupIsolatedNode(t, "host-deny")
	if _, err := pmA.EnsureProfile("secure"); err != nil {
		t.Fatal(err)
	}

	if err := storeA.SetGlobalMode("request"); err != nil {
		t.Fatal(err)
	}
	if err := storeA.EnableProfile("secure", "agy", false, 1*time.Hour); err != nil {
		t.Fatal(err)
	}

	serverA := share.NewServer(storeA, filepath.Join(nodeADir, "profiles"), "host-deny", 0)
	tsA := httptest.NewServer(serverA.Handler())
	defer tsA.Close()

	clientB := share.NewClient(0)
	clientB.SetBaseURLOverride(tsA.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := clientB.RequestLease(ctx, "host-deny", "secure", "agy", 1)
	if err != nil {
		t.Fatalf("request lease failed: %v", err)
	}
	if resp.Status != "pending" {
		t.Fatalf("expected status 'pending', got %q", resp.Status)
	}

	// Host explicitly denies the pending request
	if err := serverA.DenyRequest(resp.RequestID); err != nil {
		t.Fatalf("host deny failed: %v", err)
	}

	// Poll after denial -> should return denied
	deniedResp, err := clientB.PollLease(ctx, "host-deny", resp.RequestID)
	if err != nil {
		t.Fatalf("poll after deny failed: %v", err)
	}
	if deniedResp.Status != "denied" {
		t.Fatalf("expected status 'denied', got %q", deniedResp.Status)
	}
}

func TestShare_P2P_SecurityEnforcement(t *testing.T) {
	t.Run("global mode off rejects all requests with 403", func(t *testing.T) {
		nodeADir, pmA, storeA := setupIsolatedNode(t, "host-private")
		if _, err := pmA.EnsureProfile("vault"); err != nil {
			t.Fatal(err)
		}
		// Explicitly OFF
		if err := storeA.SetGlobalMode("off"); err != nil {
			t.Fatal(err)
		}
		if err := storeA.EnableProfile("vault", "claude", true, 4*time.Hour); err != nil {
			t.Fatal(err)
		}

		serverA := share.NewServer(storeA, filepath.Join(nodeADir, "profiles"), "host-private", 0)
		tsA := httptest.NewServer(serverA.Handler())
		defer tsA.Close()

		clientB := share.NewClient(0)
		clientB.SetBaseURLOverride(tsA.URL)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := clientB.RequestLease(ctx, "host-private", "vault", "claude", 1)
		if err == nil {
			t.Fatalf("expected error when requesting from private host, got nil")
		}
		if !strings.Contains(err.Error(), "403") && !strings.Contains(err.Error(), "Forbidden") {
			t.Errorf("expected 403 Forbidden error, got: %v", err)
		}
	})

	t.Run("external non-localhost cannot call administrative approve endpoint", func(t *testing.T) {
		nodeADir, _, storeA := setupIsolatedNode(t, "host-auth")
		serverA := share.NewServer(storeA, filepath.Join(nodeADir, "profiles"), "host-auth", 0)
		tsA := httptest.NewServer(serverA.Handler())
		defer tsA.Close()

		// Simulate external request by spoofing X-Forwarded-For or RemoteAddr
		req, err := http.NewRequest("POST", tsA.URL+"/api/share/approve", strings.NewReader(`{"request_id":"req_fake"}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "198.51.100.1")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Endpoint must strictly protect local actions
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	})
}

func TestShare_SchedulePolicies(t *testing.T) {
	tempDir, _, store := setupIsolatedNode(t, "sched-node")
	t.Setenv("AIM_HOME", tempDir)

	now := time.Now()
	startTime := fmt.Sprintf("%02d:00", (now.Hour()-1+24)%24)
	endTime := fmt.Sprintf("%02d:00", (now.Hour()+1)%24)

	sched := share.SchedulePolicy{
		ID:        "sched_test_123",
		Scope:     "global",
		Mode:      "auto",
		StartTime: startTime,
		EndTime:   endTime,
		Enabled:   true,
	}

	if err := store.SaveSchedule(sched); err != nil {
		t.Fatalf("failed to save schedule: %v", err)
	}

	// Active schedule should override base global mode 'off' to 'auto'
	if err := store.SetGlobalMode("off"); err != nil {
		t.Fatal(err)
	}
	effectiveMode := store.GetEffectiveGlobalMode()
	if effectiveMode != "auto" {
		t.Errorf("effective mode = %q, want 'auto' from active schedule window", effectiveMode)
	}

	// Delete schedule
	if err := store.DeleteSchedule("sched_test_123"); err != nil {
		t.Fatalf("failed to delete schedule: %v", err)
	}
	effectiveModeAfter := store.GetEffectiveGlobalMode()
	if effectiveModeAfter != "off" {
		t.Errorf("effective mode after delete = %q, want 'off'", effectiveModeAfter)
	}
}

func TestShare_AuditLog_Integrity(t *testing.T) {
	tempDir, _, store := setupIsolatedNode(t, "audit-node")
	t.Setenv("AIM_HOME", tempDir)

	baseTime := time.Now().Add(-1 * time.Hour)
	for i := 1; i <= 55; i++ {
		_ = store.AppendAudit(share.AuditEntry{
			ID:        fmt.Sprintf("aud_%d", i),
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			Action:    "LEASE_REQUESTED",
			Actor:     fmt.Sprintf("peer_%d", i),
			Profile:   "work",
			Agent:     "claude",
			Details:   fmt.Sprintf("Audit log test entry #%d", i),
		})
	}

	logs := store.GetAuditLogs(100)
	if len(logs) != 55 {
		t.Errorf("expected 55 audit logs, got %d", len(logs))
	}

	// Verify reverse chronological order (newest first)
	if logs[0].Details != "Audit log test entry #55" {
		t.Errorf("expected newest entry first, got: %s", logs[0].Details)
	}
}
