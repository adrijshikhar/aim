package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
)

func setupConfigTestEnv(t *testing.T) (string, *agents.Registry, *profile.ProfileManager) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	t.Setenv("AIM_AUTO_UPDATE", "")

	reg := defaultRegistry()
	pm := profile.NewProfileManager(tmpDir)

	return tmpDir, reg, pm
}

func TestCLI_Config_Path(t *testing.T) {
	tmpDir, reg, pm := setupConfigTestEnv(t)

	var buf bytes.Buffer
	cmd := newConfigCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"path"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config path failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "config.json")
	if strings.TrimSpace(buf.String()) != expected {
		t.Errorf("expected path %q, got %q", expected, strings.TrimSpace(buf.String()))
	}
}

func TestCLI_Config_List(t *testing.T) {
	_, reg, pm := setupConfigTestEnv(t)

	// Plain text table output
	var buf bytes.Buffer
	cmd := newConfigCmd(reg, pm)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config list failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "=== AIM Configuration ===") {
		t.Errorf("expected banner in config list, got: %s", out)
	}
	if !strings.Contains(out, "default_agent") || !strings.Contains(out, "auto_update") || !strings.Contains(out, "telemetry") {
		t.Errorf("expected key names in config table, got: %s", out)
	}

	// JSON output
	var jsonBuf bytes.Buffer
	cmdJSON := newConfigCmd(reg, pm)
	cmdJSON.SetOut(&jsonBuf)
	cmdJSON.SetArgs([]string{"list", "--json"})

	if err := cmdJSON.Execute(); err != nil {
		t.Fatalf("config list --json failed: %v", err)
	}

	var parsed config.Config
	if err := json.Unmarshal(jsonBuf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse config list JSON output: %v\nOutput: %s", err, jsonBuf.String())
	}
	if parsed.DefaultAgent != "agy" {
		t.Errorf("expected default agent agy in parsed JSON, got %q", parsed.DefaultAgent)
	}
}

func TestCLI_Config_GetSet(t *testing.T) {
	tmpDir, reg, pm := setupConfigTestEnv(t)

	// Initial get
	var getBuf bytes.Buffer
	cmd := newConfigCmd(reg, pm)
	cmd.SetOut(&getBuf)
	cmd.SetArgs([]string{"get", "default_agent"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config get failed: %v", err)
	}
	if strings.TrimSpace(getBuf.String()) != "agy" {
		t.Errorf("expected default_agent 'agy', got %q", strings.TrimSpace(getBuf.String()))
	}

	// Set telemetry off (exact command from opt-out notice)
	var setBuf bytes.Buffer
	cmdSet := newConfigCmd(reg, pm)
	cmdSet.SetOut(&setBuf)
	cmdSet.SetArgs([]string{"set", "telemetry", "off"})
	if err := cmdSet.Execute(); err != nil {
		t.Fatalf("config set telemetry off failed: %v", err)
	}
	if !strings.Contains(setBuf.String(), "Set telemetry = off") {
		t.Errorf("expected success message in set output, got: %s", setBuf.String())
	}

	// Verify on-disk config file
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if !cfg.TelemetryDisabled {
		t.Errorf("expected TelemetryDisabled true after set telemetry off")
	}

	// Get telemetry
	var getTelBuf bytes.Buffer
	cmdGetTel := newConfigCmd(reg, pm)
	cmdGetTel.SetOut(&getTelBuf)
	cmdGetTel.SetArgs([]string{"get", "telemetry"})
	if err := cmdGetTel.Execute(); err != nil {
		t.Fatalf("config get telemetry failed: %v", err)
	}
	if strings.TrimSpace(getTelBuf.String()) != "off" {
		t.Errorf("expected telemetry 'off', got %q", strings.TrimSpace(getTelBuf.String()))
	}

	// Get with --json
	var getJSONBuf bytes.Buffer
	cmdGetJSON := newConfigCmd(reg, pm)
	cmdGetJSON.SetOut(&getJSONBuf)
	cmdGetJSON.SetArgs([]string{"get", "telemetry", "--json"})
	if err := cmdGetJSON.Execute(); err != nil {
		t.Fatalf("config get --json failed: %v", err)
	}
	var res map[string]string
	if err := json.Unmarshal(getJSONBuf.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse get JSON: %v", err)
	}
	if res["value"] != "off" {
		t.Errorf("expected value 'off', got %q", res["value"])
	}

	// Set auto_update notify
	cmdSetUpdate := newConfigCmd(reg, pm)
	cmdSetUpdate.SetArgs([]string{"set", "auto_update", "notify"})
	if err := cmdSetUpdate.Execute(); err != nil {
		t.Fatalf("config set auto_update notify failed: %v", err)
	}

	// Set default_agent claude
	cmdSetAgent := newConfigCmd(reg, pm)
	cmdSetAgent.SetArgs([]string{"set", "default_agent", "claude"})
	if err := cmdSetAgent.Execute(); err != nil {
		t.Fatalf("config set default_agent claude failed: %v", err)
	}

	// Set invalid agent
	cmdSetBadAgent := newConfigCmd(reg, pm)
	cmdSetBadAgent.SetArgs([]string{"set", "default_agent", "nonexistent-agent"})
	if err := cmdSetBadAgent.Execute(); err == nil {
		t.Errorf("expected error setting invalid agent")
	}

	// Set invalid key
	cmdSetBadKey := newConfigCmd(reg, pm)
	cmdSetBadKey.SetArgs([]string{"set", "fake_key", "val"})
	if err := cmdSetBadKey.Execute(); err == nil {
		t.Errorf("expected error setting unknown key")
	}

	// Verify file permissions (0600)
	cfgPath := filepath.Join(tmpDir, "config.json")
	fi, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if perm := fi.Mode().Perm(); perm != 0600 {
		t.Errorf("expected config file permissions 0600, got %04o", perm)
	}

	// Test normalized display for numeric auto_update (0 -> off)
	var setNormBuf bytes.Buffer
	cmdSetNorm := newConfigCmd(reg, pm)
	cmdSetNorm.SetOut(&setNormBuf)
	cmdSetNorm.SetArgs([]string{"set", "auto_update", "0"})
	if err := cmdSetNorm.Execute(); err != nil {
		t.Fatalf("config set auto_update 0 failed: %v", err)
	}
	if !strings.Contains(setNormBuf.String(), "Set auto_update = off") {
		t.Errorf("expected normalized value 'off' in output, got: %s", setNormBuf.String())
	}

	// Test advisory note for nonexistent profile
	var setProfBuf bytes.Buffer
	cmdSetProf := newConfigCmd(reg, pm)
	cmdSetProf.SetOut(&setProfBuf)
	cmdSetProf.SetArgs([]string{"set", "default_profile", "fresh_profile"})
	if err := cmdSetProf.Execute(); err != nil {
		t.Fatalf("config set default_profile failed: %v", err)
	}
	if !strings.Contains(setProfBuf.String(), "Profile \"fresh_profile\" does not exist yet") {
		t.Errorf("expected advisory note about nonexistent profile, got: %s", setProfBuf.String())
	}
}

func TestCLI_Config_Completions(t *testing.T) {
	_, reg, pm := setupConfigTestEnv(t)
	cmd := newConfigCmd(reg, pm)

	// Test get completions
	getCmd, _, err := cmd.Find([]string{"get"})
	if err != nil {
		t.Fatalf("failed to find get command: %v", err)
	}
	comps, _ := getCmd.ValidArgsFunction(getCmd, []string{}, "")
	if len(comps) == 0 {
		t.Errorf("expected non-empty key completions for config get")
	}

	// Test set completions for keys
	setCmd, _, err := cmd.Find([]string{"set"})
	if err != nil {
		t.Fatalf("failed to find set command: %v", err)
	}
	comps, _ = setCmd.ValidArgsFunction(setCmd, []string{}, "")
	if len(comps) == 0 {
		t.Errorf("expected non-empty key completions for config set")
	}

	// Test set completions for auto_update values
	valComps, _ := setCmd.ValidArgsFunction(setCmd, []string{"auto_update"}, "")
	if len(valComps) != 4 {
		t.Errorf("expected 4 auto_update completions, got: %v", valComps)
	}

	// Test set completions for statusline_global values
	slComps, _ := setCmd.ValidArgsFunction(setCmd, []string{"statusline_global"}, "")
	if len(slComps) != 4 {
		t.Errorf("expected 4 statusline_global completions, got: %v", slComps)
	}
}

func TestCLI_Config_StatuslineGlobal(t *testing.T) {
	_, reg, pm := setupConfigTestEnv(t)

	// Default get
	var getBuf bytes.Buffer
	cmdGet := newConfigCmd(reg, pm)
	cmdGet.SetOut(&getBuf)
	cmdGet.SetArgs([]string{"get", "statusline_global"})
	if err := cmdGet.Execute(); err != nil {
		t.Fatalf("config get statusline_global failed: %v", err)
	}
	if strings.TrimSpace(getBuf.String()) != "true" {
		t.Errorf("expected default statusline_global 'true', got %q", strings.TrimSpace(getBuf.String()))
	}

	// Set false
	var setBuf bytes.Buffer
	cmdSet := newConfigCmd(reg, pm)
	cmdSet.SetOut(&setBuf)
	cmdSet.SetArgs([]string{"set", "statusline_global", "off"})
	if err := cmdSet.Execute(); err != nil {
		t.Fatalf("config set statusline_global off failed: %v", err)
	}
	if !strings.Contains(setBuf.String(), "Set statusline_global = false") {
		t.Errorf("expected success message, got: %s", setBuf.String())
	}

	// Verify loaded config
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if cfg.StatuslineGlobal == nil || *cfg.StatuslineGlobal != false {
		t.Errorf("expected StatuslineGlobal to be false in loaded config")
	}
}
