package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/share"
)

func TestShareCommandHelp(t *testing.T) {
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(t.TempDir())
	cmd := newShareCmd(reg, pm)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Share AIM profiles across teammates")) {
		t.Errorf("expected share help text, got %s", buf.String())
	}
}

func TestShareSubcommandsExist(t *testing.T) {
	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(t.TempDir())
	cmd := newShareCmd(reg, pm)

	subcommands := []string{"enable", "disable", "status", "list", "request", "renew", "release", "peer", "serve"}
	for _, sub := range subcommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == sub {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to exist", sub)
		}
	}
}

func TestShareRenewCommand_Validation(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()

	// 1. Missing args
	cmd := newShareCmd(reg, pm)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"renew"})
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error when no profile specified for renew")
	}

	// 2. Invalid profile without existing lease
	cmd = newShareCmd(reg, pm)
	buf.Reset()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"renew", "nonexistent-profile"})
	err = cmd.Execute()
	if err == nil {
		t.Errorf("expected error for nonexistent profile without lease metadata")
	}
	if !strings.Contains(err.Error(), "could not resolve lease peer for profile") {
		t.Errorf("unexpected error message: %v", err)
	}

	// 3. Invalid hours flag
	cmd = newShareCmd(reg, pm)
	buf.Reset()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"renew", "peer:work", "--hours", "0"})
	err = cmd.Execute()
	if err == nil {
		t.Errorf("expected error for --hours 0")
	}
	if !strings.Contains(err.Error(), "requested hours must be between 1 and 72") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestShareRenewCommand_WithMockServer(t *testing.T) {
	var mu sync.Mutex
	requestCount := 0
	pollCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.URL.Path == "/api/share/request" {
			requestCount++
			var req share.LeaseRequest
			_ = json.NewDecoder(r.Body).Decode(&req)

			if req.ProfileName == "pending-work" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(share.LeaseResponse{
					Status:    "pending",
					RequestID: "req_pending_123",
				})
				return
			}

			if req.ProfileName == "denied-work" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(share.LeaseResponse{
					Status:  "denied",
					Message: "lender rejected renewal",
				})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(share.LeaseResponse{
				Status:      "approved",
				LeaseID:     "ls_renewed_456",
				ProfileName: req.ProfileName,
				Agent:       "claude",
				ExpiresAt:   time.Now().Add(time.Duration(req.RequestedHours) * time.Hour),
				Payload: &share.SharePayload{
					Files: map[string]string{
						".claude/.credentials.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"secret-renewed"}`)),
					},
				},
			})
			return
		}

		if r.URL.Path == "/api/share/poll" {
			pollCount++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(share.LeaseResponse{
				Status:    "approved",
				LeaseID:   "ls_poll_approved",
				Agent:     "claude",
				ExpiresAt: time.Now().Add(2 * time.Hour),
				Payload: &share.SharePayload{
					Files: map[string]string{
						".claude/.credentials.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"secret-polled"}`)),
					},
				},
			})
			return
		}

		http.NotFound(w, r)
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()

	// Install an initial lease using LeaseManager
	mgr := share.NewLeaseManager(pm.ProfilesRoot())
	respInitial := &share.LeaseResponse{
		Status:    "approved",
		LeaseID:   "ls_initial_111",
		Agent:     "claude",
		ExpiresAt: time.Now().Add(-10 * time.Minute),
		Payload: &share.SharePayload{
			Files: map[string]string{
				".claude/.credentials.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"secret-old"}`)),
			},
		},
	}
	installedName, err := mgr.InstallLease(ts.Listener.Addr().String(), "work", respInitial)
	if err != nil {
		t.Fatalf("failed to install initial lease: %v", err)
	}

	// Add a borrower custom file that must be preserved across renewal
	customNotePath := filepath.Join(pm.ProfilesRoot(), installedName, "conversation.json")
	if err := os.WriteFile(customNotePath, []byte(`{"conversation_id":"keep_me"}`), 0600); err != nil {
		t.Fatalf("failed to write custom conversation file: %v", err)
	}

	// 1. Plain profile name resolving peer and profile from existing .aim-lease.json
	cmd := newShareCmd(reg, pm)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"renew", installedName})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected successful renewal of %s, got err: %v", installedName, err)
	}

	output := buf.String()
	if !strings.Contains(output, "Successfully renewed lease") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Paused sessions have been automatically unfrozen") {
		t.Errorf("expected unfreeze message, got: %s", output)
	}

	// Verify custom borrower file was preserved
	if preservedData, err := os.ReadFile(customNotePath); err != nil || !strings.Contains(string(preservedData), "keep_me") {
		t.Errorf("expected borrower conversation file to be preserved across renewal, got err: %v", err)
	}

	// Verify credentials updated in-place
	credPath := filepath.Join(pm.ProfilesRoot(), installedName, ".claude", ".credentials.json")
	if credData, err := os.ReadFile(credPath); err != nil || !strings.Contains(string(credData), "secret-renewed") {
		t.Errorf("expected renewed credentials, got err=%v data=%s", err, string(credData))
	}

	// 2. Direct <peer>:<profile> renewal with --hours flag
	buf.Reset()
	cmd = newShareCmd(reg, pm)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	peerArg := fmt.Sprintf("%s:work", ts.Listener.Addr().String())
	cmd.SetArgs([]string{"renew", peerArg, "--hours", "3"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected successful renewal for %s, got err: %v", peerArg, err)
	}
	if !strings.Contains(buf.String(), "Successfully renewed lease") {
		t.Errorf("expected success message, got: %s", buf.String())
	}

	// 3. Pending poll approval flow
	buf.Reset()
	cmd = newShareCmd(reg, pm)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	pendingArg := fmt.Sprintf("%s:pending-work", ts.Listener.Addr().String())
	cmd.SetArgs([]string{"renew", pendingArg})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected successful renewal with pending poll, got err: %v", err)
	}
	if !strings.Contains(buf.String(), "Successfully renewed lease") {
		t.Errorf("expected success message, got: %s", buf.String())
	}
	if pollCount == 0 {
		t.Errorf("expected poll endpoint to be queried")
	}

	// 4. Plain profile name "work" resolving to matching *:<profArg>
	buf.Reset()
	cmd = newShareCmd(reg, pm)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"renew", "work"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected successful renewal resolving 'work', got err: %v", err)
	}
	if !strings.Contains(buf.String(), "Successfully renewed lease") {
		t.Errorf("expected success message, got: %s", buf.String())
	}

	// 5. Denied renewal request returns error
	buf.Reset()
	cmd = newShareCmd(reg, pm)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	deniedArg := fmt.Sprintf("%s:denied-work", ts.Listener.Addr().String())
	cmd.SetArgs([]string{"renew", deniedArg})

	if err := cmd.Execute(); err == nil {
		t.Errorf("expected error for denied renewal request, got nil")
	} else if !strings.Contains(err.Error(), "denied") {
		t.Errorf("expected error containing 'denied', got: %v", err)
	}

	// 6. Slash format <peer>/<profile>
	buf.Reset()
	cmd = newShareCmd(reg, pm)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	slashArg := fmt.Sprintf("%s/work", ts.Listener.Addr().String())
	cmd.SetArgs([]string{"renew", slashArg})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected successful renewal with slash format, got err: %v", err)
	}
	if !strings.Contains(buf.String(), "Successfully renewed lease") {
		t.Errorf("expected success message, got: %s", buf.String())
	}
}
