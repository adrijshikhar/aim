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
