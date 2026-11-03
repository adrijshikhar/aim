package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
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
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/mcp"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/share"
)

// MA-1 (Codex Isolation):
// Verifies that preparing environment for a leased codex profile leaves
// host '~/.gemini/config/hooks.json' and host plugins 100% untouched.
func TestP2PGuard_MA1_CodexIsolation(t *testing.T) {
	tempDir := t.TempDir()
	mockHost := filepath.Join(tempDir, "host")
	t.Setenv("AIM_REAL_HOME", mockHost)
	t.Setenv("HOME", mockHost)
	t.Setenv("AIM_HOME", tempDir)

	// Set up host Gemini config and plugins
	hostGeminiCfg := filepath.Join(mockHost, ".gemini", "config")
	if err := os.MkdirAll(filepath.Join(hostGeminiCfg, "plugins"), 0755); err != nil {
		t.Fatal(err)
	}
	hostGeminiHooksPath := filepath.Join(hostGeminiCfg, "hooks.json")
	origGeminiHooks := `{"host_gemini_hook": true}`
	if err := os.WriteFile(hostGeminiHooksPath, []byte(origGeminiHooks), 0644); err != nil {
		t.Fatal(err)
	}
	hostPluginFile := filepath.Join(hostGeminiCfg, "plugins", "host_plugin.json")
	if err := os.WriteFile(hostPluginFile, []byte(`{"plugin": "gemini"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Set up host Codex hooks and plugins
	hostCodexDir := filepath.Join(mockHost, ".codex")
	if err := os.MkdirAll(filepath.Join(hostCodexDir, "plugins"), 0755); err != nil {
		t.Fatal(err)
	}
	hostCodexHooksPath := filepath.Join(hostCodexDir, "hooks.json")
	origCodexHooks := `{"host_codex_hook": true}`
	if err := os.WriteFile(hostCodexHooksPath, []byte(origCodexHooks), 0644); err != nil {
		t.Fatal(err)
	}
	hostCodexPluginFile := filepath.Join(hostCodexDir, "plugins", "host_codex_plugin.json")
	if err := os.WriteFile(hostCodexPluginFile, []byte(`{"plugin": "codex"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare environment for a leased codex profile
	profileName := "peer1:codex-leased"
	profileDir := filepath.Join(tempDir, "profiles", profileName)
	if err := os.MkdirAll(profileDir, 0700); err != nil {
		t.Fatal(err)
	}

	adapter := codex.NewAdapter()
	ctx := context.Background()
	_, err := adapter.PrepareEnv(ctx, profileName, profileDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	// 1. Verify host ~/.gemini/config/hooks.json is 100% untouched
	geminiHooksContent, err := os.ReadFile(hostGeminiHooksPath)
	if err != nil {
		t.Fatalf("failed to read host gemini hooks: %v", err)
	}
	if string(geminiHooksContent) != origGeminiHooks {
		t.Errorf("host gemini hooks modified! got %s, expected %s", string(geminiHooksContent), origGeminiHooks)
	}

	// 2. Verify host Gemini plugins untouched
	pluginContent, err := os.ReadFile(hostPluginFile)
	if err != nil || string(pluginContent) != `{"plugin": "gemini"}` {
		t.Errorf("host gemini plugins were modified or missing")
	}

	// 3. Verify host ~/.codex/hooks.json untouched
	codexHooksContent, err := os.ReadFile(hostCodexHooksPath)
	if err != nil || string(codexHooksContent) != origCodexHooks {
		t.Errorf("host codex hooks modified! got %s, expected %s", string(codexHooksContent), origCodexHooks)
	}

	// 4. Verify host Codex plugins untouched
	codexPluginContent, err := os.ReadFile(hostCodexPluginFile)
	if err != nil || string(codexPluginContent) != `{"plugin": "codex"}` {
		t.Errorf("host codex plugins were modified or missing")
	}

	// 5. Verify profile's .codex/hooks.json is a real file (not symlink) and contains aim hook guard
	profileHooksPath := filepath.Join(profileDir, ".codex", "hooks.json")
	fi, err := os.Lstat(profileHooksPath)
	if err != nil {
		t.Fatalf("profile hooks.json missing: %v", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("profile hooks.json must NOT be a symlink to host")
	}
	profileHooksContent, _ := os.ReadFile(profileHooksPath)
	if !strings.Contains(string(profileHooksContent), "aim hook guard") {
		t.Errorf("profile hooks.json should contain 'aim hook guard', got: %s", string(profileHooksContent))
	}
	if strings.Contains(string(profileHooksContent), "host_codex_hook") || strings.Contains(string(profileHooksContent), "host_gemini_hook") {
		t.Errorf("host hooks leaked into leased profile hooks.json")
	}

	// 6. Verify host plugins were not bridged or symlinked into profile
	profilePlugins := filepath.Join(profileDir, ".codex", "plugins")
	if _, err := os.Lstat(profilePlugins); err == nil {
		t.Errorf("leased profile should not have plugins directory symlinked or copied from host")
	}
}

// MA-2 (Claude Isolation):
// Verifies that preparing environment for a leased claude profile leaves
// host '~/.codex/hooks.json' and host settings untouched.
func TestP2PGuard_MA2_ClaudeIsolation(t *testing.T) {
	tempDir := t.TempDir()
	mockHost := filepath.Join(tempDir, "host")
	t.Setenv("AIM_REAL_HOME", mockHost)
	t.Setenv("HOME", mockHost)
	t.Setenv("AIM_HOME", tempDir)

	// Set up host Codex hooks
	hostCodexDir := filepath.Join(mockHost, ".codex")
	if err := os.MkdirAll(hostCodexDir, 0755); err != nil {
		t.Fatal(err)
	}
	hostCodexHooksPath := filepath.Join(hostCodexDir, "hooks.json")
	origCodexHooks := `{"host_codex": true}`
	if err := os.WriteFile(hostCodexHooksPath, []byte(origCodexHooks), 0644); err != nil {
		t.Fatal(err)
	}

	// Set up host Claude settings and ~/.claude.json
	hostClaudeDir := filepath.Join(mockHost, ".claude")
	if err := os.MkdirAll(hostClaudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	hostClaudeSettings := filepath.Join(hostClaudeDir, "settings.json")
	origClaudeSettings := `{"host_claude_setting": "preserve"}`
	if err := os.WriteFile(hostClaudeSettings, []byte(origClaudeSettings), 0644); err != nil {
		t.Fatal(err)
	}
	hostClaudeJSON := filepath.Join(mockHost, ".claude.json")
	origClaudeJSON := `{"host_mcp_servers": {}}`
	if err := os.WriteFile(hostClaudeJSON, []byte(origClaudeJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare environment for a leased claude profile
	profileName := "peer1:claude-leased"
	profileDir := filepath.Join(tempDir, "profiles", profileName)
	if err := os.MkdirAll(profileDir, 0700); err != nil {
		t.Fatal(err)
	}

	adapter := claude.NewAdapter()
	ctx := context.Background()
	_, err := adapter.PrepareEnv(ctx, profileName, profileDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	// 1. Verify host ~/.codex/hooks.json is 100% untouched
	codexHooksContent, err := os.ReadFile(hostCodexHooksPath)
	if err != nil {
		t.Fatalf("failed to read host codex hooks: %v", err)
	}
	if string(codexHooksContent) != origCodexHooks {
		t.Errorf("host codex hooks modified! got %s, expected %s", string(codexHooksContent), origCodexHooks)
	}

	// 2. Verify host Claude settings untouched
	claudeSettingsContent, err := os.ReadFile(hostClaudeSettings)
	if err != nil || string(claudeSettingsContent) != origClaudeSettings {
		t.Errorf("host claude settings modified! got %s, expected %s", string(claudeSettingsContent), origClaudeSettings)
	}

	// 3. Verify host ~/.claude.json untouched
	claudeJSONContent, err := os.ReadFile(hostClaudeJSON)
	if err != nil || string(claudeJSONContent) != origClaudeJSON {
		t.Errorf("host ~/.claude.json modified! got %s, expected %s", string(claudeJSONContent), origClaudeJSON)
	}

	// 4. Verify profile's .claude/settings.json is a real file (not symlink) and contains aim hook guard
	profileSettingsPath := filepath.Join(profileDir, ".claude", "settings.json")
	fi, err := os.Lstat(profileSettingsPath)
	if err != nil {
		t.Fatalf("profile settings.json missing: %v", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("profile settings.json must NOT be a symlink to host")
	}
	profileSettingsContent, _ := os.ReadFile(profileSettingsPath)
	if !strings.Contains(string(profileSettingsContent), "aim hook guard") {
		t.Errorf("profile settings.json should contain 'aim hook guard', got: %s", string(profileSettingsContent))
	}
	if strings.Contains(string(profileSettingsContent), "host_claude_setting") {
		t.Errorf("host claude settings leaked into leased profile settings.json")
	}
}

// MA-3 (Antigravity Schema Integrity):
// Verifies that preparing environment for a leased agy profile generates
// the flat hook schema without Claude-style nested matchers, ensuring
// zero 'denied by pre-tool hook' errors.
func TestP2PGuard_MA3_AntigravitySchemaIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	mockHost := filepath.Join(tempDir, "host")
	t.Setenv("AIM_REAL_HOME", mockHost)
	t.Setenv("HOME", mockHost)
	t.Setenv("AIM_HOME", tempDir)

	profileName := "peer1:agy-leased"
	profileDir := filepath.Join(tempDir, "profiles", profileName)
	if err := os.MkdirAll(profileDir, 0700); err != nil {
		t.Fatal(err)
	}

	adapter := agy.NewAdapter()
	ctx := context.Background()
	_, err := adapter.PrepareEnv(ctx, profileName, profileDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	hooksPath := filepath.Join(profileDir, ".gemini", "config", "hooks.json")
	data, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("failed to read agy hooks.json: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("agy hooks.json is not valid JSON: %v", err)
	}

	// 1. Verify top-level "aim-guard" exists
	aimGuardRaw, ok := parsed["aim-guard"]
	if !ok {
		t.Fatalf("expected top-level 'aim-guard' in agy hooks.json, got: %s", string(data))
	}
	aimGuard, ok := aimGuardRaw.(map[string]any)
	if !ok {
		t.Fatalf("'aim-guard' should be a JSON object, got %T", aimGuardRaw)
	}

	// 2. Verify PreInvocation and PreToolUse exist as flat slices of hook commands
	for _, event := range []string{"PreInvocation", "PreToolUse"} {
		eventRaw, ok := aimGuard[event]
		if !ok {
			t.Fatalf("missing event %s under aim-guard", event)
		}
		hooksList, ok := eventRaw.([]any)
		if !ok || len(hooksList) == 0 {
			t.Fatalf("expected non-empty array for %s, got %T", event, eventRaw)
		}
		entry, ok := hooksList[0].(map[string]any)
		if !ok {
			t.Fatalf("expected object in %s list", event)
		}
		cmdStr, _ := entry["command"].(string)
		if !strings.Contains(cmdStr, "aim hook guard") || !strings.Contains(cmdStr, "--agent agy") {
			t.Errorf("expected 'aim hook guard ... --agent agy' in command, got %s", cmdStr)
		}
		typStr, _ := entry["type"].(string)
		if typStr != "command" {
			t.Errorf("expected type 'command', got %s", typStr)
		}
	}

	// 3. Verify schema integrity: NO Claude-style nested matchers or nested "hooks" arrays
	rawStr := string(data)
	if strings.Contains(rawStr, `"matcher"`) {
		t.Errorf("CRITICAL: agy hook schema must NOT contain Claude-style 'matcher'")
	}
	for _, event := range []string{"PreInvocation", "PreToolUse"} {
		eventRaw := aimGuard[event].([]any)
		entry := eventRaw[0].(map[string]any)
		if _, hasNestedHooks := entry["hooks"]; hasNestedHooks {
			t.Errorf("CRITICAL: agy hook schema has nested 'hooks' field in %s", event)
		}
	}
}

// MA-4 (Non-Destructive Freeze):
// Simulates lender revocation while session is active; 'aim hook guard'
// execution on prompt submission outputs freeze banner and exits with code 1.
func TestP2PGuard_MA4_NonDestructiveFreeze(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	profName := "peer1:freeze-test"
	profDir := filepath.Join(tempDir, "profiles", profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	exp := time.Now().Add(2 * time.Hour)
	expStr := exp.Format(time.RFC3339)
	sig := share.SignLease(priv, "ls_freeze_1", "peer1", "borrower1", "freeze-test", "codex", expStr)

	meta := share.LeaseMeta{
		LeaseID:        "ls_freeze_1",
		SourcePeer:     "peer1",
		SourceProfile:  "freeze-test",
		Agent:          "codex",
		ExpiresAt:      exp,
		LastVerifiedAt: time.Now(),
		Revoked:        false,
		Signature:      sig,
		BorrowerID:     "borrower1",
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	leasePath := filepath.Join(profDir, ".aim-lease.json")
	if err := os.WriteFile(leasePath, data, 0600); err != nil {
		t.Fatal(err)
	}

	// Save pinned lender public key
	store := share.NewStore(tempDir)
	_ = store.SavePeer(share.PeerInfo{
		ID:        "peer1",
		HostName:  "peer1",
		PublicKey: pub,
	})
	_ = store.Close()

	// 1. Active lease: hook guard passes
	buf := new(bytes.Buffer)
	if err := executeHookGuardForDir(tempDir, profDir, profName, nil, buf); err != nil {
		t.Fatalf("expected active lease to pass, got: %v", err)
	}

	// 2. Simulate lender revocation while session is active
	meta.Revoked = true
	revokedData, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(leasePath, revokedData, 0600); err != nil {
		t.Fatal(err)
	}

	// 3. On prompt submission, hook guard executes
	buf.Reset()
	err := executeHookGuardForDir(tempDir, profDir, profName, nil, buf)
	if err == nil {
		t.Fatalf("expected revoked lease to fail with exit error")
	}

	exitErr, ok := err.(*ExitError)
	if !ok || exitErr.Code != 1 {
		t.Errorf("expected exit error with code 1, got %v", err)
	}

	outStr := buf.String()
	if !strings.Contains(outStr, "Revoked by Lender") {
		t.Errorf("expected freeze message to mention 'Revoked by Lender', got: %s", outStr)
	}
	if !strings.Contains(outStr, "Prompt & tool execution paused") {
		t.Errorf("expected freeze message to mention paused execution, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Run 'aim share renew") {
		t.Errorf("expected renewal hint in freeze banner, got: %s", outStr)
	}

	// 4. Non-destructive check: profile directory and files are intact
	if fi, err := os.Stat(profDir); err != nil || !fi.IsDir() {
		t.Errorf("profile directory was removed or damaged during freeze")
	}
	if _, err := os.Stat(leasePath); err != nil {
		t.Errorf(".aim-lease.json was removed during freeze")
	}
}

// MA-5 (In-Chat MCP Extension):
// Calls 'aim_request_lease' on ShareMCPServer while lease is active,
// verifying successful lease renewal and in-place credential installation.
func TestP2PGuard_MA5_InChatMCPExtension(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	profilesDir := filepath.Join(tempDir, "profiles")
	profName := "lender-host:mcp-prof"
	profDir := filepath.Join(profilesDir, profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	initialExp := time.Now().Add(30 * time.Minute)
	initialMeta := share.LeaseMeta{
		LeaseID:        "ls_initial",
		SourcePeer:     "lender-host",
		SourceProfile:  "mcp-prof",
		Agent:          "codex",
		GrantedAt:      time.Now(),
		ExpiresAt:      initialExp,
		LastVerifiedAt: time.Now(),
		Revoked:        false,
	}
	data, _ := json.MarshalIndent(initialMeta, "", "  ")
	if err := os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	// Add initial credential file
	codexDir := filepath.Join(profDir, ".codex")
	if err := os.MkdirAll(codexDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(`{"token":"old-secret"}`), 0600); err != nil {
		t.Fatal(err)
	}

	// Add custom borrower file that must be preserved
	borrowerCustomFile := filepath.Join(profDir, "notes.txt")
	if err := os.WriteFile(borrowerCustomFile, []byte("important session notes"), 0600); err != nil {
		t.Fatal(err)
	}

	renewedExp := time.Now().Add(4 * time.Hour)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/share/request" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(share.LeaseResponse{
				Status:      "approved",
				LeaseID:     "ls_renewed_mcp",
				ProfileName: "mcp-prof",
				Agent:       "codex",
				ExpiresAt:   renewedExp,
				Payload: &share.SharePayload{
					Files: map[string]string{
						".codex/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"new-secret-renewed"}`)),
					},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	server := mcp.NewShareMCPServer(tempDir, profilesDir, profName)
	client := share.NewClient(4319)
	client.SetBaseURLOverride(mockServer.URL)
	server.SetClient(client)

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"aim_request_lease","arguments":{"hours":4}}}` + "\n"
	out := new(bytes.Buffer)

	ctx := context.Background()
	if err := server.HandleMessage(ctx, []byte(input), out); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}

	var rpcResp mcp.JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &rpcResp); err != nil {
		t.Fatalf("failed to decode response: %v, raw: %s", err, out.String())
	}
	if rpcResp.Error != nil {
		t.Fatalf("expected nil RPC error, got: %v", rpcResp.Error)
	}
	resMap, ok := rpcResp.Result.(map[string]any)
	if !ok || resMap["isError"] == true {
		t.Fatalf("tool execution error: %s", out.String())
	}

	if !strings.Contains(out.String(), "Successfully renewed lease") {
		t.Errorf("expected success message in response, got: %s", out.String())
	}

	// Verify .aim-lease.json was updated with new lease ID and extended expiration
	updatedData, err := os.ReadFile(filepath.Join(profDir, ".aim-lease.json"))
	if err != nil {
		t.Fatalf("failed to read updated lease file: %v", err)
	}
	var updatedMeta share.LeaseMeta
	if err := json.Unmarshal(updatedData, &updatedMeta); err != nil {
		t.Fatalf("failed to parse updated lease metadata: %v", err)
	}
	if updatedMeta.LeaseID != "ls_renewed_mcp" {
		t.Errorf("expected updated lease ID 'ls_renewed_mcp', got: %s", updatedMeta.LeaseID)
	}
	if updatedMeta.ExpiresAt.Before(initialExp.Add(2 * time.Hour)) {
		t.Errorf("expected expiration to be extended by ~4 hours, got: %v", updatedMeta.ExpiresAt)
	}

	// Verify in-place credential installation
	credData, err := os.ReadFile(filepath.Join(codexDir, "auth.json"))
	if err != nil || !strings.Contains(string(credData), "new-secret-renewed") {
		t.Errorf("expected updated credentials in .codex/auth.json, got: %s (err: %v)", string(credData), err)
	}

	// Verify borrower custom file was preserved
	preservedNotes, err := os.ReadFile(borrowerCustomFile)
	if err != nil || string(preservedNotes) != "important session notes" {
		t.Errorf("expected borrower custom file to be preserved, got %s (err: %v)", string(preservedNotes), err)
	}
}

// MA-6 (Out-of-Band CLI Renewal):
// Runs 'aim share renew' while lease is frozen/revoked; verifies new signed
// manifest is installed and subsequent 'aim hook guard' call immediately unfreezes (exit code 0).
func TestP2PGuard_MA6_OutOfBandCLIRenewal(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	borrowerHost, _ := os.Hostname()
	if borrowerHost == "" {
		borrowerHost = "borrower"
	}

	mockLender := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/share/request" {
			newExp := time.Now().Add(2 * time.Hour)
			newExpStr := newExp.Format(time.RFC3339)
			sig := share.SignLease(priv, "ls_renewed_cli", "lender-node", borrowerHost, "cli-prof", "all", newExpStr)

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(share.LeaseResponse{
				Status:       "approved",
				LeaseID:      "ls_renewed_cli",
				ProfileName:  "cli-prof",
				LenderNodeID: "lender-node",
				BorrowerID:   borrowerHost,
				Agent:        "all",
				ExpiresAt:    newExp,
				Signature:    sig,
				Payload: &share.SharePayload{
					Files: map[string]string{
						".codex/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"secret-cli-renewed"}`)),
					},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockLender.Close()

	// Pin lender public key in peers.json
	store := share.NewStore(tempDir)
	_ = store.SavePeer(share.PeerInfo{
		ID:        "lender-node",
		HostName:  mockLender.Listener.Addr().String(),
		VPNIP:     mockLender.Listener.Addr().String(),
		PublicKey: pub,
	})
	_ = store.Close()

	// Install initial lease that is currently revoked
	mgr := share.NewLeaseManager(pm.ProfilesRoot())
	initialResp := &share.LeaseResponse{
		Status:       "approved",
		LeaseID:      "ls_initial_cli",
		ProfileName:  "cli-prof",
		LenderNodeID: "lender-node",
		BorrowerID:   borrowerHost,
		Agent:        "all",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Signature:    share.SignLease(priv, "ls_initial_cli", "lender-node", borrowerHost, "cli-prof", "all", time.Now().Add(1*time.Hour).Format(time.RFC3339)),
		Payload: &share.SharePayload{
			Files: map[string]string{
				".codex/auth.json": base64.StdEncoding.EncodeToString([]byte(`{"token":"secret-cli-old"}`)),
			},
		},
	}
	installedName, err := mgr.InstallLease(mockLender.Listener.Addr().String(), "cli-prof", initialResp)
	if err != nil {
		t.Fatalf("failed to install initial lease: %v", err)
	}

	profDir := filepath.Join(pm.ProfilesRoot(), installedName)

	// Mark lease revoked to simulate freeze
	leaseMeta, err := mgr.GetLeaseMeta(installedName)
	if err != nil || leaseMeta == nil {
		t.Fatalf("failed to get lease metadata: %v", err)
	}
	leaseMeta.Revoked = true
	revokedData, _ := json.MarshalIndent(leaseMeta, "", "  ")
	if err := os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), revokedData, 0600); err != nil {
		t.Fatal(err)
	}

	// 1. Hook guard fails with exit code 1 because lease is revoked
	buf := new(bytes.Buffer)
	err = executeHookGuardForDir(tempDir, profDir, installedName, nil, buf)
	if err == nil {
		t.Fatalf("expected revoked lease to be blocked by hook guard")
	}
	exitErr, ok := err.(*ExitError)
	if !ok || exitErr.Code != 1 {
		t.Errorf("expected exit error code 1, got %v", err)
	}
	if !strings.Contains(buf.String(), "inactive") {
		t.Errorf("expected freeze message in output: %s", buf.String())
	}

	// 2. Run 'aim share renew' out-of-band in CLI
	cmd := newShareCmd(reg, pm)
	cliBuf := new(bytes.Buffer)
	cmd.SetOut(cliBuf)
	cmd.SetErr(cliBuf)
	cmd.SetArgs([]string{"renew", installedName})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("aim share renew failed: %v, out: %s", err, cliBuf.String())
	}
	if !strings.Contains(cliBuf.String(), "Successfully renewed lease") {
		t.Errorf("expected success output from renew command: %s", cliBuf.String())
	}

	// 3. Subsequent 'aim hook guard' call immediately unfreezes (exit code 0)
	buf.Reset()
	err = executeHookGuardForDir(tempDir, profDir, installedName, nil, buf)
	if err != nil {
		t.Fatalf("expected session to unfreeze after renewal, got err: %v (buf: %s)", err, buf.String())
	}
}

// MA-7 (Tamper: Edit Expiration):
// Modifies 'expires_at' in '.aim-lease.json'; verifies Ed25519 signature
// verification fails and 'aim hook guard' rejects execution.
func TestP2PGuard_MA7_TamperEditExpiration(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	profName := "peer1:tamper-exp"
	profDir := filepath.Join(tempDir, "profiles", profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	borrowerHost, _ := os.Hostname()
	if borrowerHost == "" {
		borrowerHost = "borrower"
	}

	originalExp := time.Now().Add(2 * time.Hour)
	expStr := originalExp.Format(time.RFC3339)
	sig := share.SignLease(priv, "ls_tamper", "peer1", borrowerHost, "tamper-exp", "codex", expStr)

	meta := share.LeaseMeta{
		LeaseID:        "ls_tamper",
		SourcePeer:     "peer1",
		SourceProfile:  "tamper-exp",
		LenderNodeID:   "peer1",
		BorrowerID:     borrowerHost,
		Agent:          "codex",
		ExpiresAt:      originalExp,
		LastVerifiedAt: time.Now(),
		Revoked:        false,
		Signature:      sig,
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	leaseFile := filepath.Join(profDir, ".aim-lease.json")
	if err := os.WriteFile(leaseFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	store := share.NewStore(tempDir)
	_ = store.SavePeer(share.PeerInfo{
		ID:        "peer1",
		HostName:  "peer1",
		PublicKey: pub,
	})
	_ = store.Close()

	// Initial valid signature check passes
	buf := new(bytes.Buffer)
	if err := executeHookGuardForDir(tempDir, profDir, profName, nil, buf); err != nil {
		t.Fatalf("expected valid signature to pass: %v", err)
	}

	// Tamper: modify expires_at by 48 hours without lender signature
	tamperedExp := originalExp.Add(48 * time.Hour)
	meta.ExpiresAt = tamperedExp
	tamperedData, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(leaseFile, tamperedData, 0600); err != nil {
		t.Fatal(err)
	}

	// Verify hook guard detects tamper and rejects execution with exit code 1
	buf.Reset()
	err := executeHookGuardForDir(tempDir, profDir, profName, nil, buf)
	if err == nil {
		t.Fatalf("expected tampered expiration to be rejected")
	}

	exitErr, ok := err.(*ExitError)
	if !ok || exitErr.Code != 1 {
		t.Errorf("expected exit error with code 1, got: %v", err)
	}

	if !strings.Contains(buf.String(), "Cryptographic Verification Failed") {
		t.Errorf("expected 'Cryptographic Verification Failed' in freeze banner, got: %s", buf.String())
	}
}

// MA-8 (Tamper: Delete Hooks File):
// Spawns 'startLeaseSupervisor'; deletes or truncates the profile's 'hooks.json';
// verifies supervisor immediately calls cancel() to terminate child process.
func TestP2PGuard_MA8_TamperDeleteHooksFile(t *testing.T) {
	t.Run("DeleteHookFile", func(t *testing.T) {
		profDir := t.TempDir()
		hookFile := filepath.Join(profDir, "hooks.json")
		if err := os.WriteFile(hookFile, []byte(`{"command": "aim hook guard p2p-test"}`), 0600); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		startLeaseSupervisor(ctx, profDir, "p2p-test", cancel)
		time.Sleep(50 * time.Millisecond)

		// Delete hook file
		if err := os.Remove(hookFile); err != nil {
			t.Fatal(err)
		}

		select {
		case <-ctx.Done():
			// Pass: supervisor caught deletion and cancelled child process context
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for supervisor to cancel context after hook deletion")
		}
	})

	t.Run("TruncateHookFile", func(t *testing.T) {
		profDir := t.TempDir()
		hookFile := filepath.Join(profDir, "hooks.json")
		if err := os.WriteFile(hookFile, []byte(`{"command": "aim hook guard p2p-test"}`), 0600); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		startLeaseSupervisor(ctx, profDir, "p2p-test", cancel)
		time.Sleep(50 * time.Millisecond)

		// Truncate/tamper with hook file
		if err := os.WriteFile(hookFile, []byte(`{}`), 0600); err != nil {
			t.Fatal(err)
		}

		select {
		case <-ctx.Done():
			// Pass: supervisor caught truncation and cancelled context
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for supervisor to cancel context after hook truncation")
		}
	})
}

// MA-9 (Tamper: Pull Network Cable / Offline Stale):
// Advances 'last_verified_at' past 5 minutes (>5m ago); verifies 'aim hook guard'
// rejects prompt submission with stale heartbeat notice.
func TestP2PGuard_MA9_TamperOfflineStale(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	profName := "peer1:offline-stale"
	profDir := filepath.Join(tempDir, "profiles", profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	meta := share.LeaseMeta{
		LeaseID:        "ls_offline_1",
		SourcePeer:     "peer1",
		SourceProfile:  "offline-stale",
		Agent:          "codex",
		GrantedAt:      time.Now().Add(-10 * time.Minute),
		ExpiresAt:      time.Now().Add(2 * time.Hour),    // Still 2h left on lease
		LastVerifiedAt: time.Now().Add(-6 * time.Minute), // Stale (>5m ago)
		Revoked:        false,
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	err := executeHookGuardForDir(tempDir, profDir, profName, nil, buf)
	if err == nil {
		t.Fatalf("expected stale lease (>5m without heartbeat) to fail")
	}

	exitErr, ok := err.(*ExitError)
	if !ok || exitErr.Code != 1 {
		t.Errorf("expected exit error with code 1, got %v", err)
	}

	outStr := buf.String()
	if !strings.Contains(outStr, "Heartbeat Stale (>5m disconnected from VPN)") {
		t.Errorf("expected 'Heartbeat Stale (>5m disconnected from VPN)' in banner, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Prompt & tool execution paused") {
		t.Errorf("expected execution paused notice in banner, got: %s", outStr)
	}
}

// MA-10 (Subagent Trapping):
// Verifies that a tool call payload (like bash/read_file) sent via stdin to
// 'aim hook guard' is denied with exit code 1 when lease is inactive, but
// whitelisted tools ('aim_request_lease', 'aim_lease_status') pass.
func TestP2PGuard_MA10_SubagentTrapping(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	profName := "peer1:subagent-prof"
	profDir := filepath.Join(tempDir, "profiles", profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	meta := share.LeaseMeta{
		LeaseID:        "ls_subagent",
		SourcePeer:     "peer1",
		SourceProfile:  "subagent-prof",
		Agent:          "codex",
		ExpiresAt:      time.Now().Add(2 * time.Hour),
		LastVerifiedAt: time.Now(),
		Revoked:        true, // Revoked!
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(profDir, ".aim-lease.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	// Subagent tool executions that MUST be denied
	deniedPayloads := []struct {
		desc    string
		payload string
	}{
		{"bash tool call", `{"tool_name": "bash", "arguments": {"command": "cat /etc/passwd"}}`},
		{"read_file tool call", `{"name": "read_file", "arguments": {"path": "main.go"}}`},
		{"write_file tool call", `{"tool": "write_file", "arguments": {"path": "test.txt"}}`},
		{"subagent generic tool", `{"toolName": "python", "arguments": {"code": "print(1)"}}`},
		{"nested tool_call format", `{"tool_call": {"name": "execute_command"}}`},
	}

	for _, tc := range deniedPayloads {
		t.Run("Denied_"+tc.desc, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := executeHookGuardForDir(tempDir, profDir, profName, []byte(tc.payload), buf)
			if err == nil {
				t.Fatalf("expected subagent tool %s to be denied with exit code 1", tc.desc)
			}
			exitErr, ok := err.(*ExitError)
			if !ok || exitErr.Code != 1 {
				t.Errorf("expected ExitError with Code 1, got %v", err)
			}
			if !strings.Contains(buf.String(), "inactive") {
				t.Errorf("expected freeze banner in output, got: %s", buf.String())
			}
		})
	}

	// Whitelisted tools that MUST pass even when revoked
	allowedPayloads := []struct {
		desc    string
		payload string
	}{
		{"aim_request_lease via tool_name", `{"tool_name": "aim_request_lease", "arguments": {"hours": 2}}`},
		{"aim_lease_status via name", `{"name": "aim_lease_status"}`},
		{"aim_request_lease via tool", `{"tool": "aim_request_lease"}`},
		{"nested aim_request_lease tool_call", `{"tool_call": {"name": "aim_request_lease"}}`},
		{"nested aim_lease_status tool_call", `{"tool_call": {"name": "aim_lease_status"}}`},
	}

	for _, tc := range allowedPayloads {
		t.Run("Allowed_"+tc.desc, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := executeHookGuardForDir(tempDir, profDir, profName, []byte(tc.payload), buf)
			if err != nil {
				t.Fatalf("expected whitelisted tool %s to pass, got err: %v (buf: %s)", tc.desc, err, buf.String())
			}
			if buf.Len() > 0 {
				t.Errorf("expected silent success for whitelisted tool, got output: %s", buf.String())
			}
		})
	}
}

// MA-11 (Concurrent Dual-Agent Freeze):
// Sets up a shared leased profile for two simulated agent sessions;
// marks lease revoked; verifies both fail hook guard; runs renew once;
// verifies both sessions unfreeze.
func TestP2PGuard_MA11_ConcurrentDualAgentFreeze(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	profName := "peer1:dual-agent"
	profDir := filepath.Join(tempDir, "profiles", profName)
	if err := os.MkdirAll(profDir, 0700); err != nil {
		t.Fatal(err)
	}

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	borrowerHost, _ := os.Hostname()
	if borrowerHost == "" {
		borrowerHost = "borrower"
	}

	exp1 := time.Now().Add(1 * time.Hour)
	expStr1 := exp1.Format(time.RFC3339)
	sig1 := share.SignLease(priv, "ls_dual_1", "peer1", borrowerHost, "dual-agent", "all", expStr1)

	meta := share.LeaseMeta{
		LeaseID:        "ls_dual_1",
		SourcePeer:     "peer1",
		SourceProfile:  "dual-agent",
		LenderNodeID:   "peer1",
		BorrowerID:     borrowerHost,
		Agent:          "all",
		ExpiresAt:      exp1,
		LastVerifiedAt: time.Now(),
		Revoked:        false,
		Signature:      sig1,
	}
	leasePath := filepath.Join(profDir, ".aim-lease.json")
	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(leasePath, data, 0600); err != nil {
		t.Fatal(err)
	}

	store := share.NewStore(tempDir)
	_ = store.SavePeer(share.PeerInfo{
		ID:        "peer1",
		HostName:  "peer1",
		PublicKey: pub,
	})
	_ = store.Close()

	// 1. Both agents verify successfully while lease is active
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			buf := new(bytes.Buffer)
			errs[idx] = executeHookGuardForDir(tempDir, profDir, profName, nil, buf)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("agent %d failed initial active check: %v", i+1, err)
		}
	}

	// 2. Lender revokes lease
	meta.Revoked = true
	revokedData, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(leasePath, revokedData, 0600); err != nil {
		t.Fatal(err)
	}

	// Concurrently both agent sessions run prompt/tool execution and get frozen
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			buf := new(bytes.Buffer)
			err := executeHookGuardForDir(tempDir, profDir, profName, nil, buf)
			if err == nil {
				errs[idx] = fmt.Errorf("agent %d expected hook guard error when revoked, got nil", idx+1)
			} else if !strings.Contains(buf.String(), "inactive") {
				errs[idx] = fmt.Errorf("agent %d expected freeze banner in output, got: %s", idx+1, buf.String())
			} else {
				errs[idx] = nil
			}
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatalf("revoked check failed: %v", err)
		}
	}

	// 3. Perform renewal once
	exp2 := time.Now().Add(3 * time.Hour)
	expStr2 := exp2.Format(time.RFC3339)
	sig2 := share.SignLease(priv, "ls_dual_renewed", "peer1", borrowerHost, "dual-agent", "all", expStr2)
	meta.LeaseID = "ls_dual_renewed"
	meta.ExpiresAt = exp2
	meta.Revoked = false
	meta.Signature = sig2
	renewedData, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(leasePath, renewedData, 0600); err != nil {
		t.Fatal(err)
	}

	// 4. Both agent sessions immediately unfreeze concurrently on next execution
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			buf := new(bytes.Buffer)
			errs[idx] = executeHookGuardForDir(tempDir, profDir, profName, nil, buf)
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatalf("agent failed to unfreeze after renewal: %v", err)
		}
	}
}
