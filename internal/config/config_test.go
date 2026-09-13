package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Setenv("AIM_HOME", tempDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected clean default config on missing file, got error: %v", err)
	}
	if cfg.DefaultAgent != "agy" {
		t.Errorf("expected default agent 'agy', got '%s'", cfg.DefaultAgent)
	}

	cfg.DefaultProfile = "work"
	cfg.Profiles["work"] = ProfileConfig{Agents: []string{"agy"}}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	configFile := filepath.Join(tempDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatalf("expected config file to exist at %s", configFile)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if loaded.DefaultProfile != "work" {
		t.Errorf("expected default profile 'work', got '%s'", loaded.DefaultProfile)
	}
	if _, ok := loaded.Profiles["work"]; !ok {
		t.Errorf("expected profile 'work' to be present")
	}
	if !loaded.HasAgent("work", "agy") {
		t.Errorf("expected profile 'work' to have agy")
	}
}

func TestSaveConfig_Nil(t *testing.T) {
	if err := SaveConfig(nil); err == nil {
		t.Errorf("expected error when saving nil config, got nil")
	}
}

func TestConfig_BackwardCompatibility_BooleanProfiles(t *testing.T) {
	legacyJSON := `{
		"default_agent": "agy",
		"default_profile": "default",
		"profiles": {
			"bot": true,
			"work": true
		}
	}`
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	cfgFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(cfgFile, []byte(legacyJSON), 0644); err != nil {
		t.Fatalf("failed to write legacy config: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if _, ok := cfg.Profiles["bot"]; !ok {
		t.Errorf("expected profile 'bot' to be present")
	}
	if _, ok := cfg.Profiles["work"]; !ok {
		t.Errorf("expected profile 'work' to be present")
	}
}

func TestConfig_RichProfileJSON(t *testing.T) {
	richJSON := `{
		"default_agent": "agy",
		"default_profile": "work",
		"profiles": {
			"work": {
				"agents": ["agy", "claude"]
			},
			"legacy": true
		}
	}`
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	cfgFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(cfgFile, []byte(richJSON), 0644); err != nil {
		t.Fatalf("failed to write rich config: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if !cfg.HasAgent("work", "agy") {
		t.Errorf("expected work to have agy")
	}
	if !cfg.HasAgent("work", "claude") {
		t.Errorf("expected work to have claude")
	}
	if _, ok := cfg.Profiles["legacy"]; !ok {
		t.Errorf("expected profile 'legacy' to be present")
	}
}

func TestConfig_AgentHelpers(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.AddProfileAgent("work", "agy")
	cfg.AddProfileAgent("work", "claude")
	cfg.AddProfileAgent("work", "agy") // duplicate test

	if !cfg.HasAgent("work", "agy") {
		t.Errorf("expected work to have agy")
	}
	if !cfg.HasAgent("work", "claude") {
		t.Errorf("expected work to have claude")
	}
	if cfg.HasAgent("work", "gemini") {
		t.Errorf("expected work NOT to have gemini")
	}

	agents := cfg.GetProfileAgents("work")
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}

	cfg.RemoveProfileAgent("work", "agy")
	if cfg.HasAgent("work", "agy") {
		t.Errorf("expected work NOT to have agy after removal")
	}
	if !cfg.HasAgent("work", "claude") {
		t.Errorf("expected work still to have claude")
	}
}

func TestConfig_AgentHelpers_EdgeCases(t *testing.T) {
	// 1. nil *Config receiver checks
	var nilCfg *Config
	if agents := nilCfg.GetProfileAgents("work"); agents != nil {
		t.Errorf("expected nil agents on nil *Config, got %v", agents)
	}
	if nilCfg.HasAgent("work", "agy") {
		t.Errorf("expected HasAgent false on nil *Config")
	}
	nilCfg.AddProfileAgent("work", "agy")    // should not panic
	nilCfg.RemoveProfileAgent("work", "agy") // should not panic

	// 2. nil profiles map checks
	cfg := &Config{Profiles: nil}
	if agents := cfg.GetProfileAgents("nonexistent"); agents != nil {
		t.Errorf("expected nil agents for nonexistent profile on nil profiles, got %v", agents)
	}
	if cfg.HasAgent("nonexistent", "agy") {
		t.Errorf("expected HasAgent false on nil profiles")
	}
	cfg.RemoveProfileAgent("nonexistent", "agy") // should not panic

	// Adding to nil profiles initializes map
	cfg.AddProfileAgent("newprofile", "agy")
	if !cfg.HasAgent("newprofile", "agy") {
		t.Errorf("expected newprofile to have agy")
	}

	// Verify GetProfileAgents returns a copy (slice mutation does not affect internal state)
	copied := cfg.GetProfileAgents("newprofile")
	copied[0] = "mutated"
	if cfg.HasAgent("newprofile", "mutated") {
		t.Errorf("expected GetProfileAgents to return a copy, but original was modified")
	}

	// Remove non-existent agent from existing profile
	cfg.RemoveProfileAgent("newprofile", "claude")
	if !cfg.HasAgent("newprofile", "agy") {
		t.Errorf("expected newprofile to still have agy")
	}
}

func TestConfig_DeleteProfile(t *testing.T) {
	var nilCfg *Config
	nilCfg.DeleteProfile("test") // should not panic

	cfg := NewDefaultConfig()
	cfg.AddProfileAgent("test", "agy")
	if _, ok := cfg.Profiles["test"]; !ok {
		t.Fatalf("expected profile 'test' in config")
	}

	cfg.DeleteProfile("test")
	if _, ok := cfg.Profiles["test"]; ok {
		t.Errorf("expected profile 'test' to be deleted")
	}
}

func TestConfig_ProfileEnvAndArgs(t *testing.T) {
	// 1. Nil receiver safety
	var nilCfg *Config
	if env := nilCfg.GetProfileEnv("work"); env != nil {
		t.Errorf("expected nil env on nil *Config, got %v", env)
	}
	if args := nilCfg.GetProfileArgs("work"); args != nil {
		t.Errorf("expected nil args on nil *Config, got %v", args)
	}
	nilCfg.SetProfileEnv("work", map[string]string{"A": "B"}) // no panic
	nilCfg.SetProfileArgs("work", []string{"--flag"})         // no panic

	// 2. Nil profiles map initialization
	cfg := &Config{Profiles: nil}
	if env := cfg.GetProfileEnv("ghost"); env != nil {
		t.Errorf("expected nil env for non-existent profile, got %v", env)
	}
	if args := cfg.GetProfileArgs("ghost"); args != nil {
		t.Errorf("expected nil args for non-existent profile, got %v", args)
	}

	cfg.SetProfileEnv("dev", map[string]string{
		"GIT_AUTHOR_EMAIL": "dev@example.com",
		"HTTP_PROXY":       "http://127.0.0.1:8080",
	})
	cfg.SetProfileArgs("dev", []string{"--model=claude-3-7-sonnet", "--verbose"})

	env := cfg.GetProfileEnv("dev")
	if len(env) != 2 || env["GIT_AUTHOR_EMAIL"] != "dev@example.com" || env["HTTP_PROXY"] != "http://127.0.0.1:8080" {
		t.Errorf("unexpected env: %v", env)
	}

	args := cfg.GetProfileArgs("dev")
	if len(args) != 2 || args[0] != "--model=claude-3-7-sonnet" || args[1] != "--verbose" {
		t.Errorf("unexpected args: %v", args)
	}

	// 3. Defensive copying: mutating returned env or args does not modify internal config
	env["NEW_KEY"] = "leak"
	if _, ok := cfg.GetProfileEnv("dev")["NEW_KEY"]; ok {
		t.Errorf("mutating returned env leaked into config state")
	}

	args[0] = "--mutated"
	if cfg.GetProfileArgs("dev")[0] == "--mutated" {
		t.Errorf("mutating returned args leaked into config state")
	}

	// 4. Clearing env and args
	cfg.SetProfileEnv("dev", nil)
	if clearedEnv := cfg.GetProfileEnv("dev"); clearedEnv != nil {
		t.Errorf("expected nil env after setting nil, got %v", clearedEnv)
	}
	cfg.SetProfileArgs("dev", nil)
	if clearedArgs := cfg.GetProfileArgs("dev"); clearedArgs != nil {
		t.Errorf("expected nil args after setting nil, got %v", clearedArgs)
	}

	// 5. JSON serialization and deserialization
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)

	saveCfg := NewDefaultConfig()
	saveCfg.AddProfileAgent("custom", "agy")
	saveCfg.SetProfileEnv("custom", map[string]string{
		"CUSTOM_ENDPOINT": "https://api.test",
	})
	saveCfg.SetProfileArgs("custom", []string{"--no-stream"})
	if err := SaveConfig(saveCfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	loadedEnv := loaded.GetProfileEnv("custom")
	if loadedEnv["CUSTOM_ENDPOINT"] != "https://api.test" {
		t.Errorf("expected CUSTOM_ENDPOINT to be preserved across load/save, got %v", loadedEnv)
	}
	loadedArgs := loaded.GetProfileArgs("custom")
	if len(loadedArgs) != 1 || loadedArgs[0] != "--no-stream" {
		t.Errorf("expected --no-stream arg to be preserved across load/save, got %v", loadedArgs)
	}
}

func TestConfig_CustomAimHome(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	if got := BaseDir(); got != tempDir {
		t.Errorf("BaseDir() = %q, want %q", got, tempDir)
	}
	if got := ConfigDir(); got != tempDir {
		t.Errorf("ConfigDir() = %q, want %q", got, tempDir)
	}
	if got := DataDir(); got != tempDir {
		t.Errorf("DataDir() = %q, want %q", got, tempDir)
	}
	if got := StateDir(); got != tempDir {
		t.Errorf("StateDir() = %q, want %q", got, tempDir)
	}
	expectedCache := filepath.Join(tempDir, "cache")
	if got := CacheDir(); got != expectedCache {
		t.Errorf("CacheDir() = %q, want %q", got, expectedCache)
	}
	expectedConfigFile := filepath.Join(tempDir, "config.json")
	if got := ConfigFilePath(); got != expectedConfigFile {
		t.Errorf("ConfigFilePath() = %q, want %q", got, expectedConfigFile)
	}
}

func TestConfig_LegacyAimFallback(t *testing.T) {
	tempHome := t.TempDir()
	legacyAimDir := filepath.Join(tempHome, ".aim")
	if err := os.MkdirAll(legacyAimDir, 0755); err != nil {
		t.Fatalf("failed to create fake legacy .aim: %v", err)
	}

	t.Setenv("AIM_HOME", "")
	t.Setenv("AIM_REAL_HOME", tempHome)

	if got := BaseDir(); got != legacyAimDir {
		t.Errorf("BaseDir() = %q, want %q", got, legacyAimDir)
	}
	if got := ConfigDir(); got != legacyAimDir {
		t.Errorf("ConfigDir() = %q, want %q", got, legacyAimDir)
	}
	if got := DataDir(); got != legacyAimDir {
		t.Errorf("DataDir() = %q, want %q", got, legacyAimDir)
	}
	if got := StateDir(); got != legacyAimDir {
		t.Errorf("StateDir() = %q, want %q", got, legacyAimDir)
	}
	expectedCache := filepath.Join(legacyAimDir, "cache")
	if got := CacheDir(); got != expectedCache {
		t.Errorf("CacheDir() = %q, want %q", got, expectedCache)
	}
	expectedConfigFile := filepath.Join(legacyAimDir, "config.json")
	if got := ConfigFilePath(); got != expectedConfigFile {
		t.Errorf("ConfigFilePath() = %q, want %q", got, expectedConfigFile)
	}

	// Test saving and loading config in legacy directory
	cfg := NewDefaultConfig()
	cfg.DefaultProfile = "legacy-prof"
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	if _, err := os.Stat(expectedConfigFile); os.IsNotExist(err) {
		t.Fatalf("expected config at %s", expectedConfigFile)
	}
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.DefaultProfile != "legacy-prof" {
		t.Errorf("expected DefaultProfile 'legacy-prof', got %q", loaded.DefaultProfile)
	}
}

func TestConfig_DefaultXDG_WithCustomEnv(t *testing.T) {
	cleanHome := t.TempDir()
	t.Setenv("AIM_HOME", "")
	t.Setenv("AIM_REAL_HOME", cleanHome)

	cfgDir := filepath.Join(cleanHome, "custom-config")
	dataDir := filepath.Join(cleanHome, "custom-data")
	cacheDir := filepath.Join(cleanHome, "custom-cache")
	stateDir := filepath.Join(cleanHome, "custom-state")

	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("XDG_DATA_HOME", dataDir)
	t.Setenv("XDG_CACHE_HOME", cacheDir)
	t.Setenv("XDG_STATE_HOME", stateDir)
	ReloadXDG()

	expectedCfg := filepath.Join(cfgDir, "aim")
	expectedData := filepath.Join(dataDir, "aim")
	expectedCache := filepath.Join(cacheDir, "aim")
	expectedState := filepath.Join(stateDir, "aim")

	if got := ConfigDir(); got != expectedCfg {
		t.Errorf("ConfigDir() = %q, want %q", got, expectedCfg)
	}
	if got := DataDir(); got != expectedData {
		t.Errorf("DataDir() = %q, want %q", got, expectedData)
	}
	if got := CacheDir(); got != expectedCache {
		t.Errorf("CacheDir() = %q, want %q", got, expectedCache)
	}
	if got := StateDir(); got != expectedState {
		t.Errorf("StateDir() = %q, want %q", got, expectedState)
	}
	if got := BaseDir(); got != expectedData {
		t.Errorf("BaseDir() = %q, want %q", got, expectedData)
	}
	expectedConfigFile := filepath.Join(expectedCfg, "config.json")
	if got := ConfigFilePath(); got != expectedConfigFile {
		t.Errorf("ConfigFilePath() = %q, want %q", got, expectedConfigFile)
	}

	// Test saving and loading config in XDG config dir
	cfg := NewDefaultConfig()
	cfg.DefaultProfile = "xdg-prof"
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	if _, err := os.Stat(expectedConfigFile); os.IsNotExist(err) {
		t.Fatalf("expected config at %s", expectedConfigFile)
	}
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.DefaultProfile != "xdg-prof" {
		t.Errorf("expected DefaultProfile 'xdg-prof', got %q", loaded.DefaultProfile)
	}
}

func TestConfig_DefaultXDG_WithoutEnv(t *testing.T) {
	cleanHome := t.TempDir()
	t.Setenv("AIM_HOME", "")
	t.Setenv("AIM_REAL_HOME", cleanHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("XDG_STATE_HOME", "")
	ReloadXDG()

	cDir := ConfigDir()
	dDir := DataDir()
	caDir := CacheDir()
	sDir := StateDir()
	bDir := BaseDir()

	if cDir == "" || dDir == "" || caDir == "" || sDir == "" || bDir == "" {
		t.Fatalf("expected non-empty directories from XDG resolution")
	}

	// Ensure none of them default to ~/.aim when cleanHome has no .aim
	legacyAimDir := filepath.Join(cleanHome, ".aim")
	if cDir == legacyAimDir || dDir == legacyAimDir || bDir == legacyAimDir {
		t.Errorf("expected XDG resolution to not resolve to non-existent ~/.aim (%q)", legacyAimDir)
	}
}
