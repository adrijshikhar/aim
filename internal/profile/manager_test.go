package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/config"
)

func TestProfileManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-profile-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pm := NewProfileManager(tempDir)

	dir, err := pm.EnsureProfile("work")
	if err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}
	expected := filepath.Join(tempDir, "profiles", "work")
	if dir != expected {
		t.Errorf("expected dir %s, got %s", expected, dir)
	}

	profiles, err := pm.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}
	if len(profiles) != 1 || profiles[0] != "work" {
		t.Errorf("expected ['work'], got %v", profiles)
	}

	if err := pm.RemoveProfile("work"); err != nil {
		t.Fatalf("RemoveProfile failed: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected directory to be deleted")
	}
}

func TestProfileManager_EdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-profile-edge-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pm := NewProfileManager(tempDir)

	// Empty profile name error
	if _, err := pm.EnsureProfile(""); err == nil {
		t.Errorf("expected error for empty profile name, got nil")
	}

	// Remove non-existent profile error
	if err := pm.RemoveProfile("nonexistent"); err == nil {
		t.Errorf("expected error removing nonexistent profile, got nil")
	}

	// List profiles when profiles root doesn't exist
	emptyPm := NewProfileManager(filepath.Join(tempDir, "nonexistent"))
	profiles, err := emptyPm.ListProfiles()
	if err != nil {
		t.Fatalf("unexpected error listing empty profiles: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}

	// List profiles ignores files in profiles root
	_, err = pm.EnsureProfile("alpha")
	if err != nil {
		t.Fatalf("EnsureProfile alpha failed: %v", err)
	}
	regularFile := filepath.Join(pm.ProfilesRoot(), "not-a-dir.txt")
	if err := os.WriteFile(regularFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file error: %v", err)
	}

	profiles, err = pm.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}
	if len(profiles) != 1 || profiles[0] != "alpha" {
		t.Errorf("expected ['alpha'], got %v", profiles)
	}
}

func TestProfileManager_ValidationAndTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-profile-traversal-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pm := NewProfileManager(tempDir)
	_, err = pm.EnsureProfile("legit")
	if err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}

	invalidNames := []string{"", ".", "..", "a/b", `a\b`, "../escaped", "/absolute"}

	for _, name := range invalidNames {
		t.Run("EnsureProfile_"+name, func(t *testing.T) {
			if _, err := pm.EnsureProfile(name); err == nil {
				t.Errorf("EnsureProfile(%q) expected error, got nil", name)
			}
		})

		t.Run("RemoveProfile_"+name, func(t *testing.T) {
			if err := pm.RemoveProfile(name); err == nil {
				t.Errorf("RemoveProfile(%q) expected error, got nil", name)
			}
		})
	}

	// Verify that ProfilesRoot and baseDir were NOT deleted
	if _, err := os.Stat(pm.ProfilesRoot()); os.IsNotExist(err) {
		t.Fatalf("ProfilesRoot was deleted by invalid RemoveProfile call")
	}
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatalf("baseDir was deleted by invalid RemoveProfile call")
	}

	// Verify "legit" profile still exists
	if _, err := os.Stat(pm.ProfileDir("legit")); os.IsNotExist(err) {
		t.Fatalf("legit profile was deleted by invalid RemoveProfile call")
	}
}

func TestProfileManager_ListProfilesForAgent(t *testing.T) {
	baseDir := t.TempDir()
	pm := NewProfileManager(baseDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("personal")
	_, _ = pm.EnsureProfile("bot")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("work", "agy")
	cfg.AddProfileAgent("work", "claude")
	cfg.AddProfileAgent("personal", "claude")
	// "bot" has no config entry, but we mock credentials for "agy"
	agyToken := filepath.Join(pm.ProfileDir("bot"), ".gemini", "antigravity-cli", "antigravity-oauth-token")
	_ = os.MkdirAll(filepath.Dir(agyToken), 0755)
	_ = os.WriteFile(agyToken, []byte("token"), 0600)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())

	// 1. List for agy: should return "bot" (detected) and "work" (config)
	agyProfiles, err := pm.ListProfilesForAgent("agy", cfg, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agyProfiles) != 2 || agyProfiles[0] != "bot" || agyProfiles[1] != "work" {
		t.Errorf("expected [bot, work], got %v", agyProfiles)
	}

	// 2. List for claude: should return "personal" and "work"
	claudeProfiles, err := pm.ListProfilesForAgent("claude", cfg, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(claudeProfiles) != 2 || claudeProfiles[0] != "personal" || claudeProfiles[1] != "work" {
		t.Errorf("expected [personal, work], got %v", claudeProfiles)
	}

	// 3. List for gemini: should return empty
	gemProfiles, err := pm.ListProfilesForAgent("gemini", cfg, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gemProfiles) != 0 {
		t.Errorf("expected [], got %v", gemProfiles)
	}

	// 4. List with empty agent / "all": returns all 3
	allProfiles, err := pm.ListProfilesForAgent("", cfg, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(allProfiles) != 3 {
		t.Errorf("expected 3 profiles, got %v", allProfiles)
	}

	// 5. List with "all": returns all 3
	allProfiles2, err := pm.ListProfilesForAgent("all", cfg, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(allProfiles2) != 3 {
		t.Errorf("expected 3 profiles, got %v", allProfiles2)
	}

	// 6. Nil config edge case (detects via credentials only)
	nilCfgProfiles, err := pm.ListProfilesForAgent("agy", nil, reg)
	if err != nil {
		t.Fatalf("unexpected error with nil config: %v", err)
	}
	if len(nilCfgProfiles) != 1 || nilCfgProfiles[0] != "bot" {
		t.Errorf("expected [bot] with nil config, got %v", nilCfgProfiles)
	}

	// 7. Nil registry edge case (matches via config only)
	nilRegProfiles, err := pm.ListProfilesForAgent("agy", cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error with nil registry: %v", err)
	}
	if len(nilRegProfiles) != 1 || nilRegProfiles[0] != "work" {
		t.Errorf("expected [work] with nil registry, got %v", nilRegProfiles)
	}

	// 8. Nil ProfileManager receiver returns nil, nil
	var nilPM *ProfileManager
	nilPMProfiles, err := nilPM.ListProfilesForAgent("agy", cfg, reg)
	if err != nil {
		t.Fatalf("unexpected error with nil ProfileManager: %v", err)
	}
	if nilPMProfiles != nil {
		t.Errorf("expected nil profiles with nil ProfileManager, got %v", nilPMProfiles)
	}
}

func TestProfileManager_DeleteProfileAndRemoveAgent(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewProfileManager(tempDir)
	cfg := config.NewDefaultConfig()

	// 1. Create a shared profile with agy and gemini
	_, err := pm.EnsureProfile("shared")
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}
	cfg.AddProfileAgent("shared", "agy")
	cfg.AddProfileAgent("shared", "gemini")

	// 2. Remove non-existent agent returns error
	_, err = pm.RemoveAgent("shared", "claude", cfg)
	if err == nil {
		t.Errorf("expected error removing non-existent agent from shared profile")
	}

	// 3. Remove non-existent profile returns error
	_, err = pm.RemoveAgent("nonexistent", "agy", cfg)
	if err == nil {
		t.Errorf("expected error removing agent from non-existent profile")
	}

	// 4. Remove agy from shared profile (gemini remains, dir preserved)
	cleanedUp, err := pm.RemoveAgent("shared", "agy", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleanedUp {
		t.Errorf("expected cleanedUp=false since gemini remains")
	}
	if cfg.HasAgent("shared", "agy") {
		t.Errorf("expected agy to be removed from config")
	}
	if !cfg.HasAgent("shared", "gemini") {
		t.Errorf("expected gemini to remain in config")
	}
	// Directory must still exist
	if _, err := os.Stat(pm.ProfileDir("shared")); err != nil {
		t.Errorf("expected profile directory to still exist: %v", err)
	}

	// 5. Remove remaining agent gemini (now cleanedUp=true and dir deleted)
	cleanedUp, err = pm.RemoveAgent("shared", "gemini", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cleanedUp {
		t.Errorf("expected cleanedUp=true when last agent removed")
	}
	if _, err := os.Stat(pm.ProfileDir("shared")); !os.IsNotExist(err) {
		t.Errorf("expected profile directory to be deleted")
	}
	if _, ok := cfg.Profiles["shared"]; ok {
		t.Errorf("expected shared profile config to be deleted")
	}

	// 6. DeleteProfile deletes directly
	_, err = pm.EnsureProfile("direct")
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}
	cfg.AddProfileAgent("direct", "agy")

	if err := pm.DeleteProfile("direct", cfg); err != nil {
		t.Fatalf("unexpected error deleting profile: %v", err)
	}
	if _, err := os.Stat(pm.ProfileDir("direct")); !os.IsNotExist(err) {
		t.Errorf("expected direct profile directory to be deleted")
	}
	if _, ok := cfg.Profiles["direct"]; ok {
		t.Errorf("expected direct profile config to be deleted")
	}

	// 7. Delete non-existent profile returns error
	if err := pm.DeleteProfile("ghost", cfg); err == nil {
		t.Errorf("expected error deleting ghost profile")
	}
}

func TestProfileManager_CloneProfile(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewProfileManager(tempDir)
	cfg := config.NewDefaultConfig()

	// 1. Create source profile with files, dotfiles, and token
	srcDir, err := pm.EnsureProfile("src-work")
	if err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}
	cfg.AddProfileAgent("src-work", "agy")
	cfg.AddProfileAgent("src-work", "gemini")

	// Write custom settings file
	settingsFile := filepath.Join(srcDir, "settings.json")
	if err := os.WriteFile(settingsFile, []byte(`{"key": "value"}`), 0644); err != nil {
		t.Fatalf("failed to write settings file: %v", err)
	}

	// Write sensitive token file that should NOT be copied
	tokenDir := filepath.Join(srcDir, ".gemini", "antigravity-cli")
	_ = os.MkdirAll(tokenDir, 0700)
	tokenFile := filepath.Join(tokenDir, "token.json")
	if err := os.WriteFile(tokenFile, []byte(`{"access_token": "secret123"}`), 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	// 2. Clone all agents from src-work to dst-work
	if err := pm.CloneProfile("src-work", "dst-work", "", cfg); err != nil {
		t.Fatalf("CloneProfile failed: %v", err)
	}

	// Check dst-work directory exists
	dstDir := pm.ProfileDir("dst-work")
	if _, err := os.Stat(dstDir); err != nil {
		t.Fatalf("destination profile directory does not exist: %v", err)
	}

	// Check settings.json was copied
	dstSettings := filepath.Join(dstDir, "settings.json")
	if data, err := os.ReadFile(dstSettings); err != nil || string(data) != `{"key": "value"}` {
		t.Errorf("settings file was not copied correctly: err=%v, data=%s", err, string(data))
	}

	// Check token.json was NOT copied
	dstToken := filepath.Join(dstDir, ".gemini", "antigravity-cli", "token.json")
	if _, err := os.Stat(dstToken); !os.IsNotExist(err) {
		t.Errorf("expected token.json NOT to be copied to cloned profile")
	}

	// Check config has both agents
	if !cfg.HasAgent("dst-work", "agy") || !cfg.HasAgent("dst-work", "gemini") {
		t.Errorf("expected dst-work to inherit agents agy and gemini, got %v", cfg.GetProfileAgents("dst-work"))
	}

	// 3. Clone with specific agent only
	if err := pm.CloneProfile("src-work", "dst-agy-only", "agy", cfg); err != nil {
		t.Fatalf("CloneProfile with specific agent failed: %v", err)
	}
	if !cfg.HasAgent("dst-agy-only", "agy") || cfg.HasAgent("dst-agy-only", "gemini") {
		t.Errorf("expected dst-agy-only to only have agy, got %v", cfg.GetProfileAgents("dst-agy-only"))
	}

	// 4. Clone with unassociated agent returns error
	if err := pm.CloneProfile("src-work", "dst-fail", "claude", cfg); err == nil {
		t.Errorf("expected error when cloning with unassociated agent")
	}

	// 5. Clone to existing profile returns error
	if err := pm.CloneProfile("src-work", "dst-work", "", cfg); err == nil {
		t.Errorf("expected error cloning to existing profile")
	}

	// 6. Clone non-existent source profile returns error
	if err := pm.CloneProfile("ghost", "dst-ghost", "", cfg); err == nil {
		t.Errorf("expected error cloning non-existent profile")
	}

	// 7. Clone same source and destination returns error
	if err := pm.CloneProfile("src-work", "src-work", "", cfg); err == nil {
		t.Errorf("expected error cloning profile to itself")
	}

	// 8. Invalid profile names return error
	if err := pm.CloneProfile("src-work", "../bad", "", cfg); err == nil {
		t.Errorf("expected error with traversal in destination name")
	}

	// 9. Verify Env and Args are copied and isolated
	cfg.SetProfileEnv("src-work", map[string]string{
		"GIT_AUTHOR_EMAIL": "cloned@corp.internal",
	})
	cfg.SetProfileArgs("src-work", []string{"--proxy", "http://proxy:8080"})

	if err := pm.CloneProfile("src-work", "dst-with-env", "", cfg); err != nil {
		t.Fatalf("CloneProfile failed: %v", err)
	}

	clonedEnv := cfg.GetProfileEnv("dst-with-env")
	if clonedEnv["GIT_AUTHOR_EMAIL"] != "cloned@corp.internal" {
		t.Errorf("expected cloned profile to have GIT_AUTHOR_EMAIL, got %v", clonedEnv)
	}
	clonedArgs := cfg.GetProfileArgs("dst-with-env")
	if len(clonedArgs) != 2 || clonedArgs[0] != "--proxy" {
		t.Errorf("expected cloned profile to have launch args, got %v", clonedArgs)
	}

	// Mutating cloned profile config should not alter source profile config
	clonedEnv["GIT_AUTHOR_EMAIL"] = "mutated@corp.internal"
	cfg.SetProfileEnv("dst-with-env", clonedEnv)
	if cfg.GetProfileEnv("src-work")["GIT_AUTHOR_EMAIL"] == "mutated@corp.internal" {
		t.Errorf("mutating cloned profile env affected source profile")
	}
}

func TestProfileManager_RenameProfile(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewProfileManager(tempDir)
	cfg := config.NewDefaultConfig()
	cfg.DefaultProfile = "alpha"

	// 1. Setup source profile "alpha" with files, sensitive tokens, and config
	srcDir, err := pm.EnsureProfile("alpha")
	if err != nil {
		t.Fatalf("failed to ensure profile: %v", err)
	}

	tokenDir := filepath.Join(srcDir, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		t.Fatalf("failed to create token dir: %v", err)
	}
	tokenFile := filepath.Join(tokenDir, "token.json")
	if err := os.WriteFile(tokenFile, []byte(`{"access_token": "secret-token-123"}`), 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	settingsFile := filepath.Join(srcDir, "settings.json")
	if err := os.WriteFile(settingsFile, []byte(`{"theme": "dark"}`), 0644); err != nil {
		t.Fatalf("failed to write settings file: %v", err)
	}

	cfg.AddProfileAgent("alpha", "agy")
	cfg.AddProfileAgent("alpha", "gemini")
	cfg.SetProfileEnv("alpha", map[string]string{"ENV_KEY": "env_val"})
	cfg.SetProfileArgs("alpha", []string{"--custom-flag"})
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// 2. Successful rename: alpha -> beta
	if err := pm.RenameProfile("alpha", "beta", cfg); err != nil {
		t.Fatalf("RenameProfile failed: %v", err)
	}

	// Old directory should not exist
	if _, err := os.Stat(pm.ProfileDir("alpha")); !os.IsNotExist(err) {
		t.Errorf("expected old profile directory 'alpha' to no longer exist")
	}

	// New directory must exist
	dstDir := pm.ProfileDir("beta")
	if _, err := os.Stat(dstDir); err != nil {
		t.Fatalf("expected new profile directory 'beta' to exist: %v", err)
	}

	// Sensitive tokens MUST be preserved on rename
	dstToken := filepath.Join(dstDir, ".gemini", "antigravity-cli", "token.json")
	data, err := os.ReadFile(dstToken)
	if err != nil || !strings.Contains(string(data), "secret-token-123") {
		t.Errorf("expected token.json to be preserved with credentials in renamed profile: %v, data=%s", err, string(data))
	}

	// Settings file must be preserved
	dstSettings := filepath.Join(dstDir, "settings.json")
	if data, err := os.ReadFile(dstSettings); err != nil || string(data) != `{"theme": "dark"}` {
		t.Errorf("expected settings.json to be preserved: %v", err)
	}

	// Config checks
	if cfg.HasAgent("alpha", "agy") || cfg.HasAgent("alpha", "gemini") {
		t.Errorf("old profile 'alpha' should have no agents in config")
	}
	if !cfg.HasAgent("beta", "agy") || !cfg.HasAgent("beta", "gemini") {
		t.Errorf("new profile 'beta' must have agy and gemini in config")
	}
	if cfg.GetProfileEnv("beta")["ENV_KEY"] != "env_val" {
		t.Errorf("expected beta to inherit custom env")
	}
	if len(cfg.GetProfileArgs("beta")) != 1 || cfg.GetProfileArgs("beta")[0] != "--custom-flag" {
		t.Errorf("expected beta to inherit custom launch args")
	}
	if cfg.DefaultProfile != "beta" {
		t.Errorf("expected DefaultProfile to be updated to 'beta', got %q", cfg.DefaultProfile)
	}

	// 3. Error: same source and destination
	if err := pm.RenameProfile("beta", "beta", cfg); err == nil {
		t.Errorf("expected error renaming profile to itself")
	}

	// 4. Error: destination already exists
	_, _ = pm.EnsureProfile("gamma")
	if err := pm.RenameProfile("beta", "gamma", cfg); err == nil {
		t.Errorf("expected error renaming to existing profile 'gamma'")
	}

	// 5. Error: source does not exist
	if err := pm.RenameProfile("non-existent", "delta", cfg); err == nil {
		t.Errorf("expected error renaming non-existent profile")
	}

	// 6. Error: path traversal in names
	if err := pm.RenameProfile("beta", "../escaped", cfg); err == nil {
		t.Errorf("expected error with path traversal in new name")
	}
	if err := pm.RenameProfile("../escaped", "delta", cfg); err == nil {
		t.Errorf("expected error with path traversal in old name")
	}
}

func TestProfileManager_RenameProfile_RollbackOnConfigError(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewProfileManager(tempDir)
	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("orig", "agy")
	_, err := pm.EnsureProfile("orig")
	if err != nil {
		t.Fatalf("failed to create profile 'orig': %v", err)
	}

	// Create a regular file so MkdirAll fails when trying to create a directory under it
	blocker := filepath.Join(tempDir, "blocker_file")
	if err := os.WriteFile(blocker, []byte("file"), 0644); err != nil {
		t.Fatalf("failed to create blocker file: %v", err)
	}
	t.Setenv("AIM_HOME", filepath.Join(blocker, "sub"))

	err = pm.RenameProfile("orig", "dst", cfg)
	if err == nil {
		t.Fatalf("expected RenameProfile to fail when SaveConfig fails")
	}

	// Verify in-memory config was rolled back
	if !cfg.HasAgent("orig", "agy") {
		t.Errorf("expected 'orig' to still exist in in-memory config after rollback")
	}
	if cfg.HasAgent("dst", "agy") {
		t.Errorf("expected 'dst' to be removed from in-memory config after rollback")
	}

	// Verify directory on disk was rolled back
	if _, err := os.Stat(pm.ProfileDir("orig")); err != nil {
		t.Errorf("expected 'orig' directory to still exist on disk after rollback: %v", err)
	}
	if _, err := os.Stat(pm.ProfileDir("dst")); !os.IsNotExist(err) {
		t.Errorf("expected 'dst' directory to NOT exist on disk after rollback")
	}
}
