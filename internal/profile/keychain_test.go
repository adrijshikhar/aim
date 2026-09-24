package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
	// Calling PurgeIgnoredKeychains should not error out even if items do not exist
	err := PurgeIgnoredKeychains("non-existent-agent")
	if err != nil {
		t.Errorf("unexpected error purging non-existent agent keychains: %v", err)
	}

	// Purge all should succeed (items missing will return exit code 44, handled cleanly)
	err = PurgeIgnoredKeychains("")
	if err != nil {
		t.Errorf("unexpected error purging keychains: %v", err)
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
	origFn := getGenericPasswordFn
	defer func() { getGenericPasswordFn = origFn }()

	mockTokenJSON := `{"token":{"access_token":"mock_harvested_token"}}`
	getGenericPasswordFn = func(service, account string) (string, error) {
		if service == "gemini" && account == "antigravity" {
			return "go-keyring-base64:eyJ0b2tlbiI6eyJhY2Nlc3NfdG9rZW4iOiJtb2NrX2hhcnZlc3RlZF90b2tlbiJ9fQ==", nil
		}
		return "", fmt.Errorf("not found")
	}

	targetProfile := t.TempDir()
	if !HarvestKeychainTokenToProfile("agy", targetProfile) {
		t.Fatalf("expected HarvestKeychainTokenToProfile to return true")
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
	origFn := getGenericPasswordFn
	defer func() { getGenericPasswordFn = origFn }()

	targetProfile := t.TempDir()
	claudeDir := filepath.Join(targetProfile, ".claude")
	scopedService := ClaudeScopedKeychainService(claudeDir)
	destFile := filepath.Join(claudeDir, ".credentials.json")

	// 1. With only mcpOAuth in scoped keychain: should return false and NOT write file
	getGenericPasswordFn = func(service, account string) (string, error) {
		if service == scopedService {
			return `{"mcpOAuth":{"test":"token"}}`, nil
		}
		return "", fmt.Errorf("not found")
	}
	if HarvestKeychainTokenToProfile("claude", targetProfile) {
		t.Fatalf("expected HarvestKeychainTokenToProfile to return false when only mcpOAuth is in keychain")
	}
	if _, err := os.Stat(destFile); err == nil {
		t.Fatalf("expected credentials file not to be written when only mcpOAuth is in keychain")
	}

	// 2. With scoped service containing valid claudeAiOauth: should succeed and write file
	validTokenJSON := `{"claudeAiOauth":{"accessToken":"sk-ant-access-123","refreshToken":"sk-ant-refresh-123","subscriptionType":"team"},"mcpOAuth":{"plugin":"ok"}}`
	getGenericPasswordFn = func(service, account string) (string, error) {
		if service == scopedService {
			return validTokenJSON, nil
		}
		return "", fmt.Errorf("not found")
	}

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
	getGenericPasswordFn = func(service, account string) (string, error) {
		return "", fmt.Errorf("not found")
	}
	if !HarvestKeychainTokenToProfile("claude", targetProfile) {
		t.Fatalf("expected HarvestKeychainTokenToProfile to return true from existing valid file")
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

