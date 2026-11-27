package profile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/aim-cli/aim/internal/config"
)

func TestGetIgnoredKeychainEntries(t *testing.T) {
	entries := GetIgnoredKeychainEntries()
	if len(entries) < 10 {
		t.Errorf("expected at least 10 default ignored keychain entries, got %d", len(entries))
	}

	hasAgy := false
	hasClaude := false
	hasCodex := false

	for _, e := range entries {
		if e.Agent == "agy" && e.Service == "gemini" && e.Account == "antigravity" {
			hasAgy = true
		}
		if e.Agent == "claude" && e.Service == "claude-code" {
			hasClaude = true
		}
		if e.Agent == "codex" && e.Service == "codex" {
			hasCodex = true
		}
	}

	if !hasAgy {
		t.Errorf("missing default entry for agy (gemini/antigravity)")
	}
	if !hasClaude {
		t.Errorf("missing default entry for claude (claude-code)")
	}
	if !hasCodex {
		t.Errorf("missing default entry for codex (codex)")
	}

	// Test with custom services
	custom := GetIgnoredKeychainEntries("custom-tool", "another-agent")
	if len(custom) != len(entries)+2 {
		t.Errorf("expected %d entries with custom, got %d", len(entries)+2, len(custom))
	}
}

func TestPurgeIgnoredKeychains_NonDarwinOrMissing(t *testing.T) {
	var (
		mu      sync.Mutex
		deleted []string
	)
	defer SetKeychainPasswordDeleterForTest(func(service, _ string) error {
		mu.Lock()
		deleted = append(deleted, service)
		mu.Unlock()
		return nil
	})()
	defer SetKeychainInternetDeleterForTest(func(_, _ string) error {
		return nil
	})()

	// Calling PurgeIgnoredKeychains should not error out even if items do not exist
	err := PurgeIgnoredKeychains("non-existent-agent")
	if err != nil {
		t.Errorf("unexpected error purging non-existent agent keychains: %v", err)
	}

	// Purge all should succeed
	err = PurgeIgnoredKeychains("")
	if err != nil {
		t.Errorf("unexpected error purging keychains: %v", err)
	}

	if runtime.GOOS == "darwin" && len(deleted) == 0 {
		t.Errorf("expected mock deletions on darwin, got 0")
	}
}

func TestFindIgnoredKeychains_NonDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		found := FindIgnoredKeychains("")
		if found != nil {
			t.Errorf("expected nil on non-darwin OS, got %v", found)
		}
	}
}

func TestDecodeKeychainPassword(t *testing.T) {
	// 1. go-keyring-base64 encoded JSON
	originalJSON := `{"token":{"access_token":"ya29.test","refresh_token":"1//test"}}`
	b64Encoded := "go-keyring-base64:eyJ0b2tlbiI6eyJhY2Nlc3NfdG9rZW4iOiJ5YTI5LnRlc3QiLCJyZWZyZXNoX3Rva2VuIjoiMS8vdGVzdCJ9fQ=="
	decoded := DecodeKeychainPassword(b64Encoded)
	if string(decoded) != originalJSON {
		t.Errorf("expected %s, got %s", originalJSON, string(decoded))
	}

	// 2. Plain JSON
	plainJSON := `{"token":"plain_test"}`
	decodedPlain := DecodeKeychainPassword(plainJSON)
	if string(decodedPlain) != plainJSON {
		t.Errorf("expected %s, got %s", plainJSON, string(decodedPlain))
	}

	// 3. Empty string
	if DecodeKeychainPassword("") != nil {
		t.Errorf("expected nil for empty string")
	}

	// 4. Whitespace
	if DecodeKeychainPassword("   \n\t") != nil {
		t.Errorf("expected nil for whitespace")
	}
}

func TestHarvestKeychainTokenToProfile(t *testing.T) {
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")

	mockTokenJSON := `{"token":{"access_token":"mock_harvested_token"}}`
	defer SetKeychainPasswordGetterForTest(func(service, account string) (string, error) {
		if service == "gemini" && account == "antigravity" {
			return "go-keyring-base64:eyJ0b2tlbiI6eyJhY2Nlc3NfdG9rZW4iOiJtb2NrX2hhcnZlc3RlZF90b2tlbiJ9fQ==", nil
		}
		return "", fmt.Errorf("not found")
	})()

	// 1. Non-seed profile (e.g. "bby") should NOT harvest from macOS Keychain
	nonSeedProfile := filepath.Join(t.TempDir(), "bby")
	if HarvestKeychainTokenToProfile("agy", nonSeedProfile) {
		t.Fatalf("expected HarvestKeychainTokenToProfile to return false for non-seed profile")
	}
	destFileNonSeed := filepath.Join(nonSeedProfile, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if _, err := os.Stat(destFileNonSeed); err == nil {
		t.Fatalf("expected token file not to exist for non-seed profile")
	}

	// 2. Primary seed profile (e.g. "personal") should harvest from macOS Keychain
	targetProfile := filepath.Join(t.TempDir(), "personal")
	if !HarvestKeychainTokenToProfile("agy", targetProfile) {
		t.Fatalf("expected HarvestKeychainTokenToProfile to return true for personal profile")
	}

	destFile := filepath.Join(targetProfile, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read harvested token file: %v", err)
	}
	if string(data) != mockTokenJSON {
		t.Errorf("expected token content %s, got %s", mockTokenJSON, string(data))
	}

	fi, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("failed to stat harvested token file: %v", err)
	}
	if fi.Mode().Perm() != 0600 {
		t.Errorf("expected file permissions 0600, got %#o", fi.Mode().Perm())
	}

	// Harvesting again should return true without re-reading
	if !HarvestKeychainTokenToProfile("agy", targetProfile) {
		t.Errorf("expected second harvest to return true")
	}
}

func TestClaudeScopedKeychainService(t *testing.T) {
	// Verify exact 8-hex sha256 prefix hash matching Claude Code's RD() implementation
	workService := ClaudeScopedKeychainService("/Users/nemesis/.aim/profiles/work/.claude")
	if workService != "Claude Code-credentials-a937c299" {
		t.Errorf("expected Claude Code-credentials-a937c299, got %s", workService)
	}

	officeService := ClaudeScopedKeychainService("/Users/nemesis/.aim/profiles/office/.claude")
	if officeService != "Claude Code-credentials-47cc6b7e" {
		t.Errorf("expected Claude Code-credentials-47cc6b7e, got %s", officeService)
	}
}

func TestHasClaudeCredentials(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", false},
		{"invalid json", "not-json", false},
		{"mcpOAuth only", `{"mcpOAuth":{"test":"token"}}`, false},
		{"empty object", `{}`, false},
		{"oauthAccount only (metadata)", `{"oauthAccount":{"email":"test@example.com"}}`, false},
		{"claudeAiOauth empty tokens", `{"claudeAiOauth":{"accessToken":"","refreshToken":""}}`, false},
		{"claudeAiOauth with accessToken", `{"claudeAiOauth":{"accessToken":"sk-ant-test"}}`, true},
		{"claudeAiOauth with refreshToken", `{"claudeAiOauth":{"refreshToken":"sk-ant-ref"}}`, true},
		{"claudeAiOauth complete team token", `{"claudeAiOauth":{"accessToken":"sk-ant-test","refreshToken":"ref","subscriptionType":"team"},"mcpOAuth":{"plugin":"abc"}}`, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HasClaudeCredentials([]byte(tc.input))
			if got != tc.expected {
				t.Errorf("HasClaudeCredentials(%q) = %v; expected %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestHarvestKeychainTokenToProfile_Claude(t *testing.T) {
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")

	targetProfile := t.TempDir()
	claudeDir := filepath.Join(targetProfile, ".claude")
	scopedService := ClaudeScopedKeychainService(claudeDir)
	destFile := filepath.Join(claudeDir, ".credentials.json")

	// 1. With only mcpOAuth in scoped keychain: should return false and NOT write file
	restore := SetKeychainPasswordGetterForTest(func(service, _ string) (string, error) {
		if service == scopedService {
			return `{"mcpOAuth":{"test":"token"}}`, nil
		}
		return "", fmt.Errorf("not found")
	})
	defer func() {
		if restore != nil {
			restore()
		}
	}()
	if HarvestKeychainTokenToProfile("claude", targetProfile) {
		t.Fatalf("expected HarvestKeychainTokenToProfile to return false when only mcpOAuth is in keychain")
	}
	if _, err := os.Stat(destFile); err == nil {
		t.Fatalf("expected credentials file not to be written when only mcpOAuth is in keychain")
	}

	// 2. With scoped service containing valid claudeAiOauth: should succeed and write file
	validTokenJSON := `{"claudeAiOauth":{"accessToken":"sk-ant-access-123","refreshToken":"sk-ant-refresh-123","subscriptionType":"team"},"mcpOAuth":{"plugin":"ok"}}`
	restore()
	restore = SetKeychainPasswordGetterForTest(func(service, _ string) (string, error) {
		if service == scopedService {
			return validTokenJSON, nil
		}
		return "", fmt.Errorf("not found")
	})

	if !HarvestKeychainTokenToProfile("claude", targetProfile) {
		t.Fatalf("expected HarvestKeychainTokenToProfile to return true with scoped claudeAiOauth")
	}

	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read harvested claude token file: %v", err)
	}
	if string(data) != validTokenJSON {
		t.Errorf("expected token content %s, got %s", validTokenJSON, string(data))
	}

	fi, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("failed to stat harvested claude token file: %v", err)
	}
	if fi.Mode().Perm() != 0600 {
		t.Errorf("expected file permissions 0600, got %#o", fi.Mode().Perm())
	}

	// 3. Re-harvesting when file already exists on disk should return true even if keychain is empty
	restore()
	restore = SetKeychainPasswordGetterForTest(func(_, _ string) (string, error) {
		return "", fmt.Errorf("not found")
	})
	if !HarvestKeychainTokenToProfile("claude", targetProfile) {
		t.Fatalf("expected HarvestKeychainTokenToProfile to return true from existing valid file")
	}
}

func TestReadScopedClaudeKeychainCredentials(t *testing.T) {
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")

	claudeDir := filepath.Join(t.TempDir(), ".claude")
	scopedService := ClaudeScopedKeychainService(claudeDir)

	validTokenJSON := `{"claudeAiOauth":{"accessToken":"sk-ant-test-456","refreshToken":"sk-ant-refresh-456"}}`
	restore := SetKeychainPasswordGetterForTest(func(service, _ string) (string, error) {
		if service == scopedService {
			return validTokenJSON, nil
		}
		// Return different token for any other service (e.g. host account)
		return `{"claudeAiOauth":{"accessToken":"sk-ant-host-999"}}`, nil
	})
	defer func() {
		if restore != nil {
			restore()
		}
	}()

	creds := ReadScopedClaudeKeychainCredentials(claudeDir)
	if string(creds) != validTokenJSON {
		t.Fatalf("expected scoped token %s, got %s", validTokenJSON, string(creds))
	}

	// For a different directory whose scoped service is not in keychain:
	otherDir := filepath.Join(t.TempDir(), ".claude")
	restore()
	restore = SetKeychainPasswordGetterForTest(func(service, _ string) (string, error) {
		if service == scopedService {
			return validTokenJSON, nil
		}
		return "", fmt.Errorf("not found")
	})
	if otherCreds := ReadScopedClaudeKeychainCredentials(otherDir); otherCreds != nil {
		t.Fatalf("expected nil for non-matching scoped service, got %s", string(otherCreds))
	}
}

func TestClaudeKeychainServices(t *testing.T) {
	services := KnownKeychainServices("claude")
	expected := map[string]bool{
		"Claude Safe Storage":     false,
		"Claude Code-credentials": false,
	}
	for _, s := range services {
		if _, ok := expected[s]; ok {
			expected[s] = true
		}
	}
	for s, found := range expected {
		if !found {
			t.Errorf("expected Claude keychain service %q not found in KnownKeychainServices", s)
		}
	}

	allServices := KnownKeychainServices()
	hasSafeStorage := false
	for _, s := range allServices {
		if s == "Claude Safe Storage" {
			hasSafeStorage = true
			break
		}
	}
	if !hasSafeStorage {
		t.Errorf("expected 'Claude Safe Storage' in KnownKeychainServices()")
	}

	if err := PurgeAgentKeychain("claude"); err != nil {
		t.Errorf("PurgeAgentKeychain('claude') returned unexpected error: %v", err)
	}
}

type mockCredentialPurger struct {
	calledWithProfile string
	calledWithDir     string
}

func (m *mockCredentialPurger) PurgeCredentials(ctx context.Context, profileName, profileDir string) error {
	m.calledWithProfile = profileName
	m.calledWithDir = profileDir
	return nil
}

func TestPurgeProfileKeychain_Delegation(t *testing.T) {
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	profileDir := filepath.Join(t.TempDir(), "claude-work")
	claudeDir := filepath.Join(profileDir, ".claude")
	_ = os.MkdirAll(claudeDir, 0700)
	scopedService := ClaudeScopedKeychainService(claudeDir)

	var deletedService string
	restore := SetKeychainPasswordDeleterForTest(func(service, account string) error {
		deletedService = service
		return nil
	})
	defer restore()

	// 1. Without purger: delegates cleanly to default keychain deletion
	if err := PurgeProfileKeychain("claude", profileDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedService != scopedService {
		t.Errorf("expected deleted service %q, got %q", scopedService, deletedService)
	}

	// 2. With CredentialPurger: invokes purger
	purger := &mockCredentialPurger{}
	if err := PurgeProfileKeychain("claude", profileDir, purger); err != nil {
		t.Fatalf("unexpected error with purger: %v", err)
	}
	if purger.calledWithProfile != "claude-work" || purger.calledWithDir != profileDir {
		t.Errorf("expected purger to be called with profile 'claude-work' and dir %s, got %s and %s",
			profileDir, purger.calledWithProfile, purger.calledWithDir)
	}
}

func TestProfileManager_RenameProfile_MigratesClaudeKeychain(t *testing.T) {
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	pm := NewProfileManager(t.TempDir())
	cfg := config.NewDefaultConfig()

	oldDir, err := pm.EnsureProfile("old")
	if err != nil {
		t.Fatalf("EnsureProfile old: %v", err)
	}
	cfg.AddProfileAgent("old", "claude")
	oldClaudeDir := filepath.Join(oldDir, ".claude")
	if err := os.MkdirAll(oldClaudeDir, 0o700); err != nil {
		t.Fatalf("mkdir old .claude: %v", err)
	}

	oldService := ClaudeScopedKeychainService(oldClaudeDir)
	const payload = `{"claudeAiOauth":{"accessToken":"sk-ant-oat01-migrated"}}`
	defer SetKeychainPasswordGetterForTest(func(service, _ string) (string, error) {
		if service == oldService {
			return payload, nil
		}
		return "", fmt.Errorf("not found: %s", service)
	})()

	var deleted []string
	defer SetKeychainPasswordDeleterForTest(func(service, _ string) error {
		deleted = append(deleted, service)
		return nil
	})()

	if err := pm.RenameProfile("old", "newname", cfg); err != nil {
		t.Fatalf("RenameProfile failed: %v", err)
	}

	newRoot, err := os.OpenRoot(filepath.Join(pm.ProfileDir("newname"), ".claude"))
	if err != nil {
		t.Fatalf("open new .claude root: %v", err)
	}
	defer func() { _ = newRoot.Close() }()
	got, err := newRoot.ReadFile(".credentials.json")
	if err != nil {
		t.Fatalf("expected migrated .credentials.json: %v", err)
	}
	if string(got) != payload {
		t.Fatalf("migrated payload mismatch: got %s, want %s", string(got), payload)
	}

	if !slices.Contains(deleted, oldService) {
		t.Fatalf("expected old scoped keychain service %q to be deleted; deleted=%v", oldService, deleted)
	}
}

func TestProfileManager_MoveAgent_ClaudeMigratesKeychain(t *testing.T) {
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	pm := NewProfileManager(t.TempDir())
	cfg := config.NewDefaultConfig()

	srcDir, err := pm.EnsureProfile("src")
	if err != nil {
		t.Fatalf("EnsureProfile src: %v", err)
	}
	cfg.AddProfileAgent("src", "claude")
	srcClaudeDir := filepath.Join(srcDir, ".claude")
	if err := os.MkdirAll(srcClaudeDir, 0o700); err != nil {
		t.Fatalf("mkdir src .claude: %v", err)
	}

	oldService := ClaudeScopedKeychainService(srcClaudeDir)
	const payload = `{"claudeAiOauth":{"accessToken":"sk-ant-oat01-moved"}}`
	defer SetKeychainPasswordGetterForTest(func(service, _ string) (string, error) {
		if service == oldService {
			return payload, nil
		}
		return "", fmt.Errorf("not found: %s", service)
	})()

	var deleted []string
	defer SetKeychainPasswordDeleterForTest(func(service, _ string) error {
		deleted = append(deleted, service)
		return nil
	})()

	if err := pm.MoveAgent("claude", "src", "dst", false, cfg, nil); err != nil {
		t.Fatalf("MoveAgent failed: %v", err)
	}

	dstRoot, err := os.OpenRoot(filepath.Join(pm.ProfileDir("dst"), ".claude"))
	if err != nil {
		t.Fatalf("open dst .claude root: %v", err)
	}
	defer func() { _ = dstRoot.Close() }()
	got, err := dstRoot.ReadFile(".credentials.json")
	if err != nil {
		t.Fatalf("expected migrated .credentials.json: %v", err)
	}
	if string(got) != payload {
		t.Fatalf("migrated payload mismatch: got %s, want %s", string(got), payload)
	}

	if !slices.Contains(deleted, oldService) {
		t.Fatalf("expected old scoped keychain service %q to be deleted; deleted=%v", oldService, deleted)
	}
}

func TestProfileManager_RemoveAgent_PurgesKeychainWhenOtherAgentsRemain(t *testing.T) {
	t.Setenv("AIM_MOCK_KEYCHAIN", "1")
	pm := NewProfileManager(t.TempDir())
	cfg := config.NewDefaultConfig()

	workDir, err := pm.EnsureProfile("work")
	if err != nil {
		t.Fatalf("EnsureProfile work: %v", err)
	}
	cfg.AddProfileAgent("work", "claude")
	cfg.AddProfileAgent("work", "codex")

	claudeDir := filepath.Join(workDir, ".claude")
	codexDir := filepath.Join(workDir, ".codex")
	for _, d := range []string{claudeDir, codexDir} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(`{"claudeAiOauth":{"accessToken":"secret"}}`), 0o600); err != nil {
		t.Fatalf("write claude creds: %v", err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(`{"tokens":{"access_token":"codex-keep"}}`), 0o600); err != nil {
		t.Fatalf("write codex auth: %v", err)
	}

	scopedService := ClaudeScopedKeychainService(claudeDir)
	var deleted []string
	defer SetKeychainPasswordDeleterForTest(func(service, _ string) error {
		deleted = append(deleted, service)
		return nil
	})()

	deletedProfile, err := pm.RemoveAgent("work", "claude", cfg)
	if err != nil {
		t.Fatalf("RemoveAgent failed: %v", err)
	}
	if deletedProfile {
		t.Fatal("expected profile to be preserved when codex agent remains")
	}
	if cfg.HasAgent("work", "claude") || !cfg.HasAgent("work", "codex") {
		t.Fatalf("unexpected agent membership: claude=%v codex=%v", cfg.HasAgent("work", "claude"), cfg.HasAgent("work", "codex"))
	}
	if _, err := os.Stat(filepath.Join(codexDir, "auth.json")); err != nil {
		t.Fatalf("expected codex auth.json preserved, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(claudeDir, ".credentials.json")); !os.IsNotExist(err) {
		t.Fatalf("expected claude credentials purged, got err=%v", err)
	}
	if !slices.Contains(deleted, scopedService) {
		t.Fatalf("expected scoped keychain service %q to be deleted; deleted=%v", scopedService, deleted)
	}
}
