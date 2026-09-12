package agy

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/usage"
)

func TestAgyMetadata(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "agy" {
		t.Errorf("expected Name to be 'agy', got '%s'", adapter.Name())
	}
	if adapter.DisplayName() != "Antigravity CLI" {
		t.Errorf("expected DisplayName to be 'Antigravity CLI', got '%s'", adapter.DisplayName())
	}
	if len(adapter.Aliases()) != 1 || adapter.Aliases()[0] != "antigravity" {
		t.Errorf("expected Aliases to contain 'antigravity', got %v", adapter.Aliases())
	}
	if adapter.BinaryName() != "agy" {
		t.Errorf("expected BinaryName to be 'agy', got '%s'", adapter.BinaryName())
	}
}

func TestAgyPrepareEnv(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-agy-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set GEMINI_CLI_HOME to verify it gets stripped by PrepareEnv
	t.Setenv("GEMINI_CLI_HOME", "/some/unwanted/path")

	adapter := NewAdapter()
	launchEnv, err := adapter.PrepareEnv("work", tempDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	if launchEnv.Env["HOME"] != tempDir {
		t.Errorf("expected HOME to be %s, got %s", tempDir, launchEnv.Env["HOME"])
	}
	if _, exists := launchEnv.Env["SSH_CONNECTION"]; exists {
		t.Errorf("expected SSH_CONNECTION to be omitted when credentials missing to enable browser auto-open")
	}

	// Now write token into tempDir and verify SSH_CONNECTION is set
	tokenFile := filepath.Join(tempDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	_ = os.MkdirAll(filepath.Dir(tokenFile), 0700)
	_ = os.WriteFile(tokenFile, []byte(`{"token":{"access_token":"mock"}}`), 0600)

	launchEnvWithCreds, err := adapter.PrepareEnv("work", tempDir)
	if err != nil {
		t.Fatalf("PrepareEnv with creds failed: %v", err)
	}
	if launchEnvWithCreds.Env["SSH_CONNECTION"] != "127.0.0.1 50000 127.0.0.1 22" {
		t.Errorf("expected SSH_CONNECTION to be set when credentials exist, got %s", launchEnvWithCreds.Env["SSH_CONNECTION"])
	}
	if _, exists := launchEnv.Env["SSH_CLIENT"]; exists {
		t.Errorf("expected SSH_CLIENT to be deleted from launch environment")
	}
	if _, exists := launchEnv.Env["SSH_TTY"]; exists {
		t.Errorf("expected SSH_TTY to be deleted from launch environment")
	}
	if launchEnv.Env["AIM_AGENT"] != "agy" {
		t.Errorf("expected AIM_AGENT to be 'agy', got '%s'", launchEnv.Env["AIM_AGENT"])
	}
	if launchEnv.Env["AIM_PROFILE"] != "work" {
		t.Errorf("expected AIM_PROFILE to be 'work', got '%s'", launchEnv.Env["AIM_PROFILE"])
	}
	if _, exists := launchEnv.Env["GEMINI_CLI_HOME"]; exists {
		t.Errorf("expected GEMINI_CLI_HOME to be deleted from launch environment")
	}

	tokenDir := filepath.Join(tempDir, ".gemini", "antigravity-cli")
	fi, err := os.Stat(tokenDir)
	if err != nil || !fi.IsDir() {
		t.Errorf("expected .gemini/antigravity-cli directory to be created")
	}
}

func TestAgyTokenPath(t *testing.T) {
	adapter := NewAdapter()
	expected := filepath.Join("/fake/profile", ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if got := adapter.TokenPath("/fake/profile"); got != expected {
		t.Errorf("expected TokenPath %s, got %s", expected, got)
	}
}

func TestAgyDoctor(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-doctor-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	adapter := NewAdapter()

	// Scenario 1: Missing token file
	results := adapter.Doctor(context.Background(), "work", tempDir)
	var tokenResultFound bool
	for _, r := range results {
		if r.Category == "Token" {
			tokenResultFound = true
			if r.Status != "FAIL" {
				t.Errorf("expected Token status FAIL for missing token, got %s", r.Status)
			}
			if !strings.Contains(r.Message, "Missing token file") {
				t.Errorf("unexpected message for missing token: %s", r.Message)
			}
		}
	}
	if !tokenResultFound {
		t.Fatalf("Token category not found in doctor results")
	}

	// Scenario 2: Corrupt token file (invalid json)
	tokFile := adapter.TokenPath(tempDir)
	if err := os.MkdirAll(filepath.Dir(tokFile), 0700); err != nil {
		t.Fatalf("failed to create token dir: %v", err)
	}
	if err := os.WriteFile(tokFile, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("failed to write corrupt token: %v", err)
	}

	results = adapter.Doctor(context.Background(), "work", tempDir)
	for _, r := range results {
		if r.Category == "Token" {
			if r.Status != "FAIL" {
				t.Errorf("expected FAIL for corrupt token JSON, got %s", r.Status)
			}
		}
	}

	// Scenario 3: Corrupt token file (valid json, but empty refresh_token)
	emptyRefresh := map[string]any{
		"token": map[string]any{
			"access_token":  "access123",
			"refresh_token": "",
		},
	}
	data, _ := json.Marshal(emptyRefresh)
	if err := os.WriteFile(tokFile, data, 0600); err != nil {
		t.Fatalf("failed to write empty refresh token: %v", err)
	}

	results = adapter.Doctor(context.Background(), "work", tempDir)
	for _, r := range results {
		if r.Category == "Token" {
			if r.Status != "FAIL" {
				t.Errorf("expected FAIL for missing refresh_token, got %s", r.Status)
			}
		}
	}

	// Scenario 4: Valid token but expired access token
	pastTime := time.Now().Add(-10 * time.Minute)
	validExpired := map[string]any{
		"token": map[string]any{
			"access_token":  "expired_acc",
			"refresh_token": "valid_ref",
			"expiry":        pastTime.Format(time.RFC3339),
		},
	}
	data, _ = json.Marshal(validExpired)
	if err := os.WriteFile(tokFile, data, 0600); err != nil {
		t.Fatalf("failed to write valid expired token: %v", err)
	}

	results = adapter.Doctor(context.Background(), "work", tempDir)
	for _, r := range results {
		if r.Category == "Token" {
			if r.Status != "OK" {
				t.Errorf("expected OK for expired access token with valid refresh_token, got %s", r.Status)
			}
			if !strings.Contains(r.Message, "expired") || !strings.Contains(r.Message, "ready for auto-refresh") {
				t.Errorf("unexpected message for expired token: %s", r.Message)
			}
		}
	}

	// Scenario 5: Valid token with future expiry
	futureTime := time.Now().Add(45 * time.Minute)
	validFuture := map[string]any{
		"token": map[string]any{
			"access_token":  "valid_acc",
			"refresh_token": "valid_ref",
			"expiry":        futureTime.Format(time.RFC3339),
		},
	}
	data, _ = json.Marshal(validFuture)
	if err := os.WriteFile(tokFile, data, 0600); err != nil {
		t.Fatalf("failed to write valid future token: %v", err)
	}

	results = adapter.Doctor(context.Background(), "work", tempDir)
	for _, r := range results {
		if r.Category == "Token" {
			if r.Status != "OK" {
				t.Errorf("expected OK for valid future token, got %s", r.Status)
			}
			if !strings.Contains(r.Message, "valid for") {
				t.Errorf("unexpected message for future token: %s", r.Message)
			}
		}
	}
}

func TestAntigravityAdapter_HasCredentials(t *testing.T) {
	adapter := NewAntigravityAdapter()
	tmpDir := t.TempDir()

	// 1. Not present
	if adapter.HasCredentials(tmpDir) {
		t.Errorf("expected HasCredentials to be false on empty profile")
	}

	// 2. Present but empty file
	tokenPath := adapter.TokenPath(tmpDir)
	_ = os.MkdirAll(filepath.Dir(tokenPath), 0755)
	_ = os.WriteFile(tokenPath, []byte(""), 0600)
	if adapter.HasCredentials(tmpDir) {
		t.Errorf("expected HasCredentials to be false on empty token file")
	}

	// 3. Valid token file
	_ = os.WriteFile(tokenPath, []byte(`{"access_token":"valid"}`), 0600)
	if !adapter.HasCredentials(tmpDir) {
		t.Errorf("expected HasCredentials to be true on valid token file")
	}

	// 4. Directory instead of file
	dirProfile := t.TempDir()
	dirTokenPath := adapter.TokenPath(dirProfile)
	_ = os.MkdirAll(dirTokenPath, 0755)
	if adapter.HasCredentials(dirProfile) {
		t.Errorf("expected HasCredentials to be false when token path is a directory")
	}

	// 5. Corrupt token file with trailing extra brace (auto-repaired)
	corruptProfile := t.TempDir()
	corruptTokenPath := adapter.TokenPath(corruptProfile)
	_ = os.MkdirAll(filepath.Dir(corruptTokenPath), 0755)
	_ = os.WriteFile(corruptTokenPath, []byte("{\"access_token\":\"valid\"}}\n"), 0600)
	if !adapter.HasCredentials(corruptProfile) {
		t.Errorf("expected HasCredentials to repair and return true for token file with trailing brace")
	}
	repairedContent, _ := os.ReadFile(corruptTokenPath)
	if string(repairedContent) != "{\"access_token\":\"valid\"}\n" {
		t.Errorf("expected file to be repaired on disk, got %q", string(repairedContent))
	}

	// 6. Google Application Default Credentials (ADC)
	adcProfile := t.TempDir()
	adcPath := filepath.Join(adcProfile, ".config", "gcloud", "application_default_credentials.json")
	_ = os.MkdirAll(filepath.Dir(adcPath), 0755)
	_ = os.WriteFile(adcPath, []byte(`{"client_id":"test","client_secret":"test"}`), 0600)
	if !adapter.HasCredentials(adcProfile) {
		t.Errorf("expected HasCredentials to be true when ADC exists")
	}

	// 7. Auto-seed personal profile from host token
	fakeHome := t.TempDir()
	t.Setenv("AIM_REAL_HOME", fakeHome)
	hostTokenPath := filepath.Join(fakeHome, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	_ = os.MkdirAll(filepath.Dir(hostTokenPath), 0755)
	_ = os.WriteFile(hostTokenPath, []byte(`{"token":{"access_token":"seeded_tok"}}`), 0600)

	personalProfileDir := filepath.Join(t.TempDir(), "personal")
	_ = os.MkdirAll(personalProfileDir, 0755)
	if !adapter.SeedDefaultCredentials("personal", personalProfileDir) {
		t.Errorf("expected SeedDefaultCredentials to return true for 'personal'")
	}
	if !adapter.HasCredentials(personalProfileDir) {
		t.Errorf("expected HasCredentials to be true after seeding")
	}
	seededToken, sErr := os.ReadFile(adapter.TokenPath(personalProfileDir))
	if sErr != nil || !strings.Contains(string(seededToken), "seeded_tok") {
		t.Errorf("expected token to be seeded into personal profile, got err: %v, content: %s", sErr, string(seededToken))
	}

	// 8. Auto-seed shorthand "p" profile via direct HasCredentials call
	pProfileDir := filepath.Join(t.TempDir(), "p")
	_ = os.MkdirAll(pProfileDir, 0755)
	if !adapter.HasCredentials(pProfileDir) {
		t.Errorf("expected HasCredentials to auto-seed and return true for 'p' profile")
	}
	pToken, pErr := os.ReadFile(adapter.TokenPath(pProfileDir))
	if pErr != nil || !strings.Contains(string(pToken), "seeded_tok") {
		t.Errorf("expected token to be seeded into 'p' profile, got err: %v, content: %s", pErr, string(pToken))
	}

	// 9. Profile not eligible for auto-seeding when multiple profiles exist
	fakeAimHome := t.TempDir()
	t.Setenv("AIM_HOME", fakeAimHome)
	multiCfg := config.NewDefaultConfig()
	multiCfg.Profiles["p"] = config.ProfileConfig{}
	multiCfg.Profiles["rs"] = config.ProfileConfig{}
	multiCfg.DefaultProfile = "p"
	_ = config.SaveConfig(multiCfg)

	rsProfileDir := filepath.Join(t.TempDir(), "rs")
	_ = os.MkdirAll(rsProfileDir, 0755)
	if adapter.SeedDefaultCredentials("rs", rsProfileDir) {
		t.Errorf("expected SeedDefaultCredentials to return false for non-default profile 'rs'")
	}
	if adapter.HasCredentials(rsProfileDir) {
		t.Errorf("expected HasCredentials to be false for unauthenticated 'rs' profile")
	}

	// 10. Auto-seeding from macOS Keychain when host token file is missing
	noFileHome := t.TempDir()
	t.Setenv("AIM_REAL_HOME", noFileHome)
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	keychainProfileDir := filepath.Join(t.TempDir(), "personal")
	_ = os.MkdirAll(keychainProfileDir, 0755)

	// Harvest will fail without keychain mock, but adapter shouldn't crash
	_ = adapter.SeedDefaultCredentials("personal", keychainProfileDir)
}

func TestAntigravityAdapter_DoctorADC(t *testing.T) {
	adapter := NewAdapter()
	adcProfile := t.TempDir()
	adcPath := filepath.Join(adcProfile, ".config", "gcloud", "application_default_credentials.json")
	_ = os.MkdirAll(filepath.Dir(adcPath), 0755)
	_ = os.WriteFile(adcPath, []byte(`{"client_id":"test"}`), 0600)

	results := adapter.Doctor(context.Background(), "adc_prof", adcProfile)
	foundADC := false
	for _, r := range results {
		if r.Category == "Token" && r.Status == "OK" && strings.Contains(r.Message, "ADC") {
			foundADC = true
			break
		}
	}
	if !foundADC {
		t.Errorf("expected Doctor to report OK for ADC credentials, got: %+v", results)
	}
}

func TestAntigravityAdapterGetUsageNoCredentials(t *testing.T) {
	adapter := NewAdapter()
	tmpDir, err := os.MkdirTemp("", "aim-agy-usage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	rep, err := adapter.GetUsage(context.Background(), "test_prof", tmpDir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rep.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown when no credentials exist, got %s", rep.Status)
	}
}

func TestParseAgyUsageTSV(t *testing.T) {
	tsv := "Gemini Models\tWeekly Limit Remaining\t80%\t2026-09-15T18:53:30Z\n" +
		"Gemini Models\tFive Hour Limit Remaining\t95%\t2026-09-09T13:37:30Z\n" +
		"Claude and GPT models\tWeekly Limit Remaining\t84%\t2026-09-16T05:20:19Z\n" +
		"Claude and GPT models\tFive Hour Limit Remaining\t100%\t2026-09-09T16:09:10Z\n"

	windows := ParseUsageTSV(tsv)
	if len(windows) != 4 {
		t.Fatalf("expected 4 windows, got %d", len(windows))
	}
	if windows[0].RemainingPct != 80 || windows[1].RemainingPct != 95 {
		t.Errorf("incorrect percentages parsed: %+v", windows)
	}
}

func TestParseAgyUsageTSV_EdgeCases(t *testing.T) {
	tsv := "\n" + // empty line
		"OnlyOneColumn\n" +
		"Two\tColumns\n" +
		"Bad\tPct\tnotanumber%\n" +
		"Valid\tNoDate\t50%\n" +
		"Valid\tInvalidDate\t60%\tnot-a-date\n" +
		"Valid\tPastDate\t70%\t2000-01-01T00:00:00Z\n" +
		"Gemini Models\tFive Hour Limit Remaining\tdisabled\t\n"

	windows := ParseUsageTSV(tsv)
	if len(windows) != 4 {
		t.Fatalf("expected 4 valid windows, got %d", len(windows))
	}

	if windows[0].Name != "NoDate" || windows[0].RemainingPct != 50 {
		t.Errorf("unexpected window 0: %+v", windows[0])
	}
	if windows[1].Name != "InvalidDate" || windows[1].RemainingPct != 60 {
		t.Errorf("unexpected window 1: %+v", windows[1])
	}
	if windows[2].Name != "PastDate" || windows[2].RemainingPct != 70 {
		t.Errorf("unexpected window 2: %+v", windows[2])
	}
	if windows[2].ResetsIn != 0 {
		t.Errorf("expected past date to have ResetsIn = 0, got %v", windows[2].ResetsIn)
	}
	if windows[3].Name != "Five Hour Limit Remaining" || windows[3].RemainingPct != 0 {
		t.Errorf("expected disabled window to have RemainingPct = 0, got %+v", windows[3])
	}
}

func TestAntigravityAdapterGetUsage_CommandFailureAndFallback(t *testing.T) {
	adapter := NewAdapter()
	tmpDir := t.TempDir()

	// Add valid token so HasCredentials is true
	tokenPath := adapter.TokenPath(tmpDir)
	_ = os.MkdirAll(filepath.Dir(tokenPath), 0700)
	_ = os.WriteFile(tokenPath, []byte(`{"access_token":"dummy"}`), 0600)

	// Create a dummy bin dir with a failing agy binary
	binDir := t.TempDir()
	failingScript := filepath.Join(binDir, "agy")
	scriptContent := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(failingScript, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	// Case 1: command fails, no conversation_summaries.db
	rep, err := adapter.GetUsage(context.Background(), "test_prof", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Status != usage.StatusUnknown {
		t.Errorf("expected StatusUnknown, got %s", rep.Status)
	}
	if rep.Summary != "Offline" {
		t.Errorf("expected summary 'Offline', got '%s'", rep.Summary)
	}
	if rep.Error == "" {
		t.Errorf("expected non-empty error")
	}

	// Case 2: command fails, conversation_summaries.db exists
	dbPath := filepath.Join(tmpDir, ".gemini", "antigravity-cli", "conversation_summaries.db")
	_ = os.WriteFile(dbPath, []byte("sqlite content"), 0644)

	rep2, err := adapter.GetUsage(context.Background(), "test_prof", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep2.Summary != "Offline (local session cache present)" {
		t.Errorf("expected 'Offline (local session cache present)', got '%s'", rep2.Summary)
	}
}

func TestAntigravityAdapterGetUsage_Success(t *testing.T) {
	adapter := NewAdapter()
	tmpDir := t.TempDir()

	// Add valid token
	tokenPath := adapter.TokenPath(tmpDir)
	_ = os.MkdirAll(filepath.Dir(tokenPath), 0700)
	_ = os.WriteFile(tokenPath, []byte(`{"access_token":"dummy"}`), 0600)

	// Create a dummy mock script for agy
	binDir := t.TempDir()
	mockScript := filepath.Join(binDir, "agy")
	scriptContent := `#!/bin/sh
if [ "$2" = "/usage" ]; then
    printf "Gemini Models\tFive Hour Limit Remaining\t85%%\t2030-01-01T00:00:00Z\n"
    printf "Claude and GPT models\tWeekly Limit Remaining\t92%%\t2030-01-07T00:00:00Z\n"
elif [ "$2" = "/credits" ]; then
    printf "Credits Remaining\t250\n"
fi
`
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	rep, err := adapter.GetUsage(context.Background(), "work", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rep.Agent != "agy" {
		t.Errorf("expected agent 'agy', got '%s'", rep.Agent)
	}
	if rep.Profile != "work" {
		t.Errorf("expected profile 'work', got '%s'", rep.Profile)
	}
	if rep.Status != usage.StatusOK {
		t.Errorf("expected status OK, got %s", rep.Status)
	}
	if len(rep.Windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(rep.Windows))
	}
	if rep.Credits != "250" {
		t.Errorf("expected 250 credits, got '%s'", rep.Credits)
	}
	if !strings.Contains(rep.Summary, "5h: 85%") {
		t.Errorf("expected summary to contain '5h: 85%%', got '%s'", rep.Summary)
	}
	if !strings.Contains(rep.Summary, "wk: 92%") {
		t.Errorf("expected summary to contain 'wk: 92%%', got '%s'", rep.Summary)
	}
}

func TestAntigravityAdapterGetUsage_ZeroWindows(t *testing.T) {
	adapter := NewAdapter()
	tmpDir := t.TempDir()

	// Write mock credentials
	tokenPath := adapter.TokenPath(tmpDir)
	if err := os.MkdirAll(filepath.Dir(tokenPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tokenPath, []byte(`{"access_token":"fake"}`), 0600); err != nil {
		t.Fatal(err)
	}

	binDir := t.TempDir()
	mockScript := filepath.Join(binDir, "agy")
	scriptContent := `#!/bin/sh
echo "Unrelated stdout"
`
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	rep, err := adapter.GetUsage(context.Background(), "work", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rep.Windows) != 0 {
		t.Fatalf("expected 0 windows, got %d", len(rep.Windows))
	}
	if rep.Summary != "0 windows parsed" {
		t.Errorf("expected summary '0 windows parsed', got %q (Error: %q, Status: %q, Windows: %d)", rep.Summary, rep.Error, rep.Status, len(rep.Windows))
	}
}

func TestAgyBridgeSharedState(t *testing.T) {
	fakeRealHome := t.TempDir()
	sharedAgyDir := filepath.Join(fakeRealHome, ".gemini", "antigravity-cli")
	_ = os.MkdirAll(filepath.Join(sharedAgyDir, "conversations"), 0755)
	_ = os.MkdirAll(filepath.Join(sharedAgyDir, "brain"), 0755)
	_ = os.WriteFile(filepath.Join(sharedAgyDir, "conversation_summaries.db"), []byte("shared-db"), 0644)
	_ = os.WriteFile(filepath.Join(sharedAgyDir, "history.jsonl"), []byte("shared-hist"), 0644)

	profileDir := t.TempDir()
	profAgyDir := filepath.Join(profileDir, ".gemini", "antigravity-cli")
	profConvDir := filepath.Join(profAgyDir, "conversations")
	_ = os.MkdirAll(profConvDir, 0755)
	_ = os.WriteFile(filepath.Join(profConvDir, "local-session.db"), []byte("local-session-content"), 0644)

	err := bridgeSharedState(fakeRealHome, profileDir)
	if err != nil {
		t.Fatalf("bridgeSharedState failed: %v", err)
	}

	// Verify conversations is now a symlink
	fi, err := os.Lstat(profConvDir)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected profile conversations to be a symlink")
	}

	// Verify pre-existing local session was migrated into shared store
	migrated := filepath.Join(sharedAgyDir, "conversations", "local-session.db")
	if _, err := os.Stat(migrated); os.IsNotExist(err) {
		t.Errorf("expected local-session.db to be migrated to shared conversations")
	}

	// Verify brain is a symlink
	profBrain := filepath.Join(profAgyDir, "brain")
	fi, err = os.Lstat(profBrain)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected profile brain to be a symlink")
	}

	// Verify conversation_summaries.db is a symlink
	profDb := filepath.Join(profAgyDir, "conversation_summaries.db")
	fi, err = os.Lstat(profDb)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected profile conversation_summaries.db to be a symlink")
	}

	// Verify history.jsonl is a symlink
	profHist := filepath.Join(profAgyDir, "history.jsonl")
	fi, err = os.Lstat(profHist)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected profile history.jsonl to be a symlink")
	}

	// Idempotency check: running again should not error or alter symlinks
	err = bridgeSharedState(fakeRealHome, profileDir)
	if err != nil {
		t.Errorf("second bridgeSharedState call failed: %v", err)
	}
}

func TestAgyPrepareEnv_AIMHome(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("AIM_HOME", tempHome)

	profileDir := t.TempDir()
	adapter := NewAdapter()
	env, err := adapter.PrepareEnv("testprof", profileDir)
	if err != nil {
		t.Fatalf("PrepareEnv failed: %v", err)
	}

	if env.Env["AIM_HOME"] != tempHome {
		t.Errorf("expected AIM_HOME to be %s, got %s", tempHome, env.Env["AIM_HOME"])
	}
}

func TestAgyLogin_NewProfileAndExistingProfile(t *testing.T) {
	// Create mock agy script
	binDir := t.TempDir()
	mockAgy := filepath.Join(binDir, "agy")
	mockScript := `#!/bin/sh
# Verify SSH variables are stripped so agy auto-opens the browser
if [ -n "$SSH_CONNECTION" ] || [ -n "$SSH_CLIENT" ] || [ -n "$SSH_TTY" ] || [ -n "$GEMINI_CLI_HOME" ]; then
    echo "ERROR: SSH variable present" >&2
    exit 1
fi
if [ -z "$HOME" ] || [ -z "$AIM_AGENT" ] || [ -z "$AIM_PROFILE" ]; then
    echo "ERROR: required AIM env variable missing" >&2
    exit 2
fi

# Simulate successful login by writing mock token to profile directory
mkdir -p "$HOME/.gemini/antigravity-cli"
echo '{"token":{"access_token":"mock_login_access_token"}}' > "$HOME/.gemini/antigravity-cli/antigravity-oauth-token"
exit 0
`
	if err := os.WriteFile(mockAgy, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	// Simulate host having SSH environment variables
	t.Setenv("SSH_CONNECTION", "192.168.1.100 54321 192.168.1.50 22")
	t.Setenv("SSH_CLIENT", "192.168.1.100 54321 22")
	t.Setenv("SSH_TTY", "/dev/ttys002")

	adapter := NewAdapter()

	// Scenario 1: Login for a brand new profile directory
	newProfDir := filepath.Join(t.TempDir(), "brand_new")
	if err := adapter.Login(context.Background(), "brand_new", newProfDir); err != nil {
		t.Fatalf("Login failed for brand new profile: %v", err)
	}
	if !adapter.HasCredentials(newProfDir) {
		t.Errorf("expected HasCredentials to be true after successful login for brand new profile")
	}

	// Scenario 2: Login for an existing profile directory where credentials were missing
	existingProfDir := t.TempDir()
	if adapter.HasCredentials(existingProfDir) {
		t.Fatalf("expected existing profile to have no credentials initially")
	}
	if err := adapter.Login(context.Background(), "existing_prof", existingProfDir); err != nil {
		t.Fatalf("Login failed for existing profile with missing creds: %v", err)
	}
	if !adapter.HasCredentials(existingProfDir) {
		t.Errorf("expected HasCredentials to be true after login for existing profile")
	}
}

