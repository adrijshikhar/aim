package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
)

type mockAdapter struct {
	name       string
	binaryPath string
	args       []string
	exitErr    error
	hasCreds   bool
}

func (m *mockAdapter) Name() string        { return m.name }
func (m *mockAdapter) DisplayName() string { return m.name }
func (m *mockAdapter) Aliases() []string   { return []string{"m"} }
func (m *mockAdapter) BinaryName() string  { return filepath.Base(m.binaryPath) }
func (m *mockAdapter) HasCredentials(profileDir string) bool {
	return m.hasCreds
}
func (m *mockAdapter) Login(ctx context.Context, profileName, profileDir string) error {
	return m.exitErr
}
func (m *mockAdapter) PrepareEnv(profileName, profileDir string) (agents.LaunchEnv, error) {
	if m.exitErr != nil {
		return agents.LaunchEnv{}, m.exitErr
	}
	return agents.LaunchEnv{
		BinaryPath: m.binaryPath,
		Args:       m.args,
		WorkingDir: profileDir,
		Env: map[string]string{
			"TEST_PROFILE": profileName,
		},
	}, nil
}
func (m *mockAdapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	return []agents.DiagnosticResult{
		{Category: "Auth", Status: "OK", Message: "Authenticated"},
	}
}
func (m *mockAdapter) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	return &usage.Report{
		Agent:     m.name,
		Profile:   profileName,
		Status:    usage.StatusOK,
		FetchedAt: time.Now(),
		Windows: []usage.LimitWindow{
			{Name: "Five Hour Limit Remaining", RemainingPct: 82, ResetsIn: 2 * time.Hour},
			{Name: "Weekly Limit Remaining", RemainingPct: 90, ResetsIn: 5 * 24 * time.Hour},
		},
	}, nil
}

func TestCLIDispatchList(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("personal")

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("work", "agy")
	cfg.AddProfileAgent("personal", "agy")
	_ = config.SaveConfig(cfg)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	err = runList(reg, pm, "agy")
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
}

func TestCLIDispatchListEmpty(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()
	err = runList(reg, pm, "agy")
	if err != nil {
		t.Fatalf("runList empty failed: %v", err)
	}
}

func TestCLIRunRemove(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	_, err = pm.EnsureProfile("to-delete")
	if err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("to-delete", "agy")
	_ = config.SaveConfig(cfg)

	err = runRemove(nil, pm, []string{"to-delete"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Verify profile directory is removed
	if _, err := os.Stat(pm.ProfileDir("to-delete")); !os.IsNotExist(err) {
		t.Fatalf("profile directory still exists")
	}

	// Verify profile is deleted from config
	cfgReload, _ := config.LoadConfig()
	if _, exists := cfgReload.Profiles["to-delete"]; exists {
		t.Fatalf("profile still exists in config")
	}

	// Removing non-existent profile should return error
	errNonExistent := runRemove(nil, pm, []string{"non-existent"})
	if errNonExistent == nil {
		t.Fatalf("expected error for non-existent profile, got nil")
	}
}

func TestCLIRunMv(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	srcDir, err := pm.EnsureProfile("rs")
	if err != nil {
		t.Fatalf("EnsureProfile rs failed: %v", err)
	}

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("rs", "agy")
	cfg.AddProfileAgent("rs", "codex")
	_ = config.SaveConfig(cfg)

	// Mock codex credentials in rs
	codexDir := filepath.Join(srcDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)
	_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte("secret-token"), 0600)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{name: "codex"})
	reg.Register(&mockAdapter{name: "agy"})

	// Execute aim mv codex rs work
	code := executeMv(reg, pm, "codex", "rs", "work", false)
	if code != 0 {
		t.Fatalf("executeMv failed with exit code: %d", code)
	}

	// Verify rs still has agy but not codex
	cfgReload, _ := config.LoadConfig()
	if !cfgReload.HasAgent("rs", "agy") {
		t.Errorf("expected rs to still have agy")
	}
	if cfgReload.HasAgent("rs", "codex") {
		t.Errorf("expected rs to no longer have codex")
	}

	// Verify work has codex and credentials
	if !cfgReload.HasAgent("work", "codex") {
		t.Errorf("expected work to have codex")
	}
	workAuth, err := os.ReadFile(filepath.Join(pm.ProfileDir("work"), ".codex", "auth.json"))
	if err != nil || string(workAuth) != "secret-token" {
		t.Errorf("expected moved codex credentials in work profile, got %s (err: %v)", string(workAuth), err)
	}
}

func TestCLIRunDoctor(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("testprof")

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{name: "mock"})

	// Known agent
	runDoctor(reg, pm, "mock")

	// Unknown agent
	runDoctor(reg, pm, "unknown")
}

func TestCLIExecuteRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()

	// Unknown agent should return 1
	code := executeRun(reg, pm, "non-existent", "test", nil)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}

	// Successful run with mock
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})
	code = executeRun(reg, pm, "mock", "work", nil)
	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}

	// Exit code forwarding with non-zero exit
	reg.Register(&mockAdapter{
		name:       "failer",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 42"},
	})
	code = executeRun(reg, pm, "failer", "work", nil)
	if code != 42 {
		t.Fatalf("expected child exit code 42, got %d", code)
	}
}

func TestCLIExecuteLogin(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()

	// 1. Unknown agent should fail with exit code 1
	if code := executeLogin(reg, pm, "unknown", "work"); code != 1 {
		t.Fatalf("expected code 1 for unknown agent, got %d", code)
	}

	reg.Register(&mockAdapter{name: "mock"})

	// 2. Creating a brand new profile via login
	newProfName := "brand_new_profile"
	newProfDir := pm.ProfileDir(newProfName)
	if _, err := os.Stat(newProfDir); !os.IsNotExist(err) {
		t.Fatalf("expected new profile dir to not exist before login")
	}

	if code := executeLogin(reg, pm, "mock", newProfName); code != 0 {
		t.Fatalf("expected login code 0 when creating new profile, got %d", code)
	}

	if fi, err := os.Stat(newProfDir); err != nil || !fi.IsDir() {
		t.Fatalf("expected profile dir %s to be created as directory", newProfDir)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if !cfg.HasAgent(newProfName, "mock") {
		t.Fatalf("expected new profile %q to have agent 'mock' in config after login", newProfName)
	}

	// 3. Logging in an existing profile where creds are missing
	existingProfName := "existing_no_creds"
	existingProfDir, err := pm.EnsureProfile(existingProfName)
	if err != nil {
		t.Fatalf("failed to pre-create existing profile: %v", err)
	}
	if fi, err := os.Stat(existingProfDir); err != nil || !fi.IsDir() {
		t.Fatalf("expected existing profile dir to exist")
	}

	if code := executeLogin(reg, pm, "mock", existingProfName); code != 0 {
		t.Fatalf("expected login code 0 for existing profile, got %d", code)
	}

	cfgReload, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if !cfgReload.HasAgent(existingProfName, "mock") {
		t.Fatalf("expected existing profile %q to have agent 'mock' tagged in config", existingProfName)
	}

	// 4. Login with agent alias "m" tags canonical name "mock"
	if code := executeLogin(reg, pm, "m", "alias_prof"); code != 0 {
		t.Fatalf("expected login code 0 with alias 'm', got %d", code)
	}
	cfgAlias, _ := config.LoadConfig()
	if !cfgAlias.HasAgent("alias_prof", "mock") {
		t.Fatalf("expected canonical agent 'mock' to be tagged for alias 'm'")
	}
}

func TestCLIExecuteShell(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()

	// Unknown agent
	if code := executeShell(reg, pm, "unknown", "work"); code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestCLICorruptedConfigRecovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{name: "mock"})

	// Corrupt the config.json file
	configPath := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(configPath, []byte("NOT_VALID_JSON{{{"), 0644); err != nil {
		t.Fatalf("failed to write corrupted config: %v", err)
	}

	// executeLogin should not panic on corrupted config and should recover
	code := executeLogin(reg, pm, "mock", "profile1")
	if code != 0 {
		t.Fatalf("executeLogin expected 0, got %d", code)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed after login recovery: %v", err)
	}
	if !cfg.HasAgent("profile1", "mock") {
		t.Fatalf("expected profile1 to have agent 'mock' in config")
	}

	// Corrupt config.json again
	if err := os.WriteFile(configPath, []byte("CORRUPTED_AGAIN{{{"), 0644); err != nil {
		t.Fatalf("failed to write corrupted config: %v", err)
	}

	// runRemove should not panic on corrupted config and should recover
	err = runRemove(nil, pm, []string{"profile1"})
	if err != nil {
		t.Fatalf("runRemove expected nil error, got %v", err)
	}
	cfg, err = config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed after remove recovery: %v", err)
	}
	if _, exists := cfg.Profiles["profile1"]; exists {
		t.Fatalf("expected profile1 to be deleted from config")
	}
}

func TestCLIDispatch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)
	t.Setenv("AIM_HOME", tempDir)

	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})

	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{
			name:     "version flag -v",
			args:     []string{"-v"},
			wantCode: 0,
		},
		{
			name:     "version flag --version",
			args:     []string{"--version"},
			wantCode: 0,
		},
		{
			name:     "version command",
			args:     []string{"version"},
			wantCode: 0,
		},
		{
			name:     "help flag -h",
			args:     []string{"-h"},
			wantCode: 0,
		},
		{
			name:     "help flag --help",
			args:     []string{"--help"},
			wantCode: 0,
		},
		{
			name:     "help command",
			args:     []string{"help"},
			wantCode: 0,
		},
		{
			name:     "invalid command",
			args:     []string{"invalid-cmd"},
			wantCode: 1,
		},
		{
			name:     "incomplete run command",
			args:     []string{"run"},
			wantCode: 1,
		},
		{
			name:     "incomplete run command with agent only",
			args:     []string{"run", "mock"},
			wantCode: 1,
		},
		{
			name:     "incomplete shell command",
			args:     []string{"shell"},
			wantCode: 1,
		},
		{
			name:     "incomplete shell command with agent only",
			args:     []string{"shell", "mock"},
			wantCode: 1,
		},
		{
			name:     "incomplete login command",
			args:     []string{"login"},
			wantCode: 1,
		},
		{
			name:     "incomplete login command with agent only",
			args:     []string{"login", "mock"},
			wantCode: 1,
		},
		{
			name:     "incomplete remove command",
			args:     []string{"remove"},
			wantCode: 1,
		},
		{
			name:     "rm command rejected as unknown",
			args:     []string{"rm", "testprof"},
			wantCode: 1,
		},
		{
			name:     "standard list",
			args:     []string{"list"},
			wantCode: 0,
		},
		{
			name:     "standard doctor",
			args:     []string{"doctor"},
			wantCode: 0,
		},
		{
			name:     "standard login with mock",
			args:     []string{"login", "mock", "testprof"},
			wantCode: 0,
		},
		{
			name:     "standard run with mock",
			args:     []string{"run", "mock", "testprof"},
			wantCode: 0,
		},
		{
			name:     "standard run with mock and extra args",
			args:     []string{"run", "mock", "testprof", "--", "foo"},
			wantCode: 0,
		},
		{
			name:     "standard remove existing profile",
			args:     []string{"remove", "testprof"},
			wantCode: 0,
		},
		{
			name:     "shorthand alias agy list rejected",
			args:     []string{"agy", "list"},
			wantCode: 1,
		},
		{
			name:     "shorthand agent name only rejected",
			args:     []string{"agy"},
			wantCode: 1,
		},
		{
			name:     "shorthand alias mock list rejected",
			args:     []string{"mock", "list"},
			wantCode: 1,
		},
		{
			name:     "shorthand alias mock doctor rejected",
			args:     []string{"mock", "doctor"},
			wantCode: 1,
		},
		{
			name:     "shorthand alias mock login rejected",
			args:     []string{"mock", "login", "shortprof"},
			wantCode: 1,
		},
		{
			name:     "shorthand alias mock run rejected",
			args:     []string{"mock", "run", "shortprof"},
			wantCode: 1,
		},
		{
			name:     "shorthand alias mock run with extra args rejected",
			args:     []string{"mock", "run", "shortprof", "--", "arg1"},
			wantCode: 1,
		},
		{
			name:     "shorthand alias mock remove rejected",
			args:     []string{"mock", "remove", "shortprof"},
			wantCode: 1,
		},
		{
			name:     "completion without shell",
			args:     []string{"completion"},
			wantCode: 1,
		},
		{
			name:     "completion with valid shell",
			args:     []string{"completion", "zsh"},
			wantCode: 0,
		},
		{
			name:     "completion with invalid shell",
			args:     []string{"completion", "invalid"},
			wantCode: 1,
		},
		{
			name:     "hidden complete root",
			args:     []string{"__complete"},
			wantCode: 0,
		},
		{
			name:     "hidden complete run",
			args:     []string{"__complete", "run"},
			wantCode: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotCode := dispatch(tc.args, reg, pm)
			if gotCode != tc.wantCode {
				t.Errorf("dispatch(%v) = %d, want %d", tc.args, gotCode, tc.wantCode)
			}
		})
	}
}

func TestCLIDispatchUI(t *testing.T) {
	orig := tuiRunner
	defer func() { tuiRunner = orig }()

	calledCount := 0
	tuiRunner = func(reg *agents.Registry, pm *profile.ProfileManager) int {
		calledCount++
		return 0
	}

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(t.TempDir())

	// Test dispatch with no args (bare `aim` directly launches TUI)
	code := dispatch([]string{}, reg, pm)
	if code != 0 {
		t.Errorf("dispatch([]) expected code 0, got %d", code)
	}
	if calledCount != 1 {
		t.Errorf("expected tuiRunner to be called once after bare aim dispatch, got %d", calledCount)
	}

	// Test that "ui" subcommand has been removed
	codeUI := dispatch([]string{"ui"}, reg, pm)
	if codeUI == 0 {
		t.Errorf("expected dispatch([\"ui\"]) to fail since ui command is removed, got 0")
	}
}

func TestRunTUINilProfileManager(t *testing.T) {
	// Verify that runTUI with nil ProfileManager does not panic
	orig := tuiRunner
	defer func() { tuiRunner = orig }()

	called := false
	tuiRunner = func(reg *agents.Registry, pm *profile.ProfileManager) int {
		cfg, _ := config.LoadConfig()
		_ = tui.NewModel(reg, pm, cfg)
		called = true
		return 0
	}

	reg := agents.NewRegistry()
	code := runTUI(reg, nil)
	if code != 0 || !called {
		t.Errorf("expected code 0 and called=true, got code %d", code)
	}
}

func TestCLI_ExecuteRun_ContinueFlag(t *testing.T) {
	tempBase := t.TempDir()
	origBaseDir := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origBaseDir) }()
	_ = os.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	_, _ = pm.EnsureProfile("work")

	reg := agents.NewRegistry()
	outFile := filepath.Join(tempBase, "args.txt")
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", `echo "$@" > "` + outFile + `"`, "--"},
	})

	code := executeRun(reg, pm, "mock", "work", []string{"--continue"})
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	content, _ := os.ReadFile(outFile)
	if strings.TrimSpace(string(content)) != "--continue" {
		t.Errorf("expected '--continue', got %q", string(content))
	}
}

func TestCLI_AgentAwareListAndDoctor(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	reg := agents.DefaultRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tmpDir)

	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("other")

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("work", "agy")
	_ = config.SaveConfig(cfg)

	// Dispatch aim list agy: should succeed
	code := dispatch([]string{"list", "agy"}, reg, pm)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}

	// Dispatch aim doctor agy: should succeed
	code = dispatch([]string{"doctor", "agy"}, reg, pm)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

func TestCLI_AgentAwareListFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	reg.Register(&mockAdapter{name: "mock"})
	pm := profile.NewProfileManager(tmpDir)

	_, _ = pm.EnsureProfile("agy-prof")
	_, _ = pm.EnsureProfile("mock-prof")
	_, _ = pm.EnsureProfile("multi-prof")
	_, _ = pm.EnsureProfile("empty-prof")

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("agy-prof", "agy")
	cfg.AddProfileAgent("mock-prof", "mock")
	cfg.AddProfileAgent("multi-prof", "agy")
	cfg.AddProfileAgent("multi-prof", "mock")
	_ = config.SaveConfig(cfg)

	// List with specific agent filter "agy"
	if err := runList(reg, pm, "agy"); err != nil {
		t.Fatalf("runList for agy failed: %v", err)
	}

	// List with specific agent filter "mock"
	if err := runList(reg, pm, "mock"); err != nil {
		t.Fatalf("runList for mock failed: %v", err)
	}

	// List for agent with no profiles
	if err := runList(reg, pm, "other"); err != nil {
		t.Fatalf("runList for other failed: %v", err)
	}

	// List all profiles (agentName == "")
	if err := runList(reg, pm, ""); err != nil {
		t.Fatalf("runList without agent failed: %v", err)
	}
}

func TestCLI_AgentAwareDoctor(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	reg := agents.NewRegistry()
	mock := &mockAdapter{name: "mock"}
	reg.Register(mock)
	pm := profile.NewProfileManager(tmpDir)

	_, _ = pm.EnsureProfile("prof1")
	_, _ = pm.EnsureProfile("prof2")

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("prof1", "mock")
	_ = config.SaveConfig(cfg)

	// Doctor for agent with configured profiles
	runDoctor(reg, pm, "mock")

	// Doctor for agent with no profiles
	runDoctor(reg, pm, "mock-none")

	// Doctor for unknown agent
	runDoctor(reg, pm, "unknown")

	// Bare doctor: runs diagnostics across all registered adapters
	runDoctor(reg, pm, "")

	// Bare doctor via dispatch
	code := dispatch([]string{"doctor"}, reg, pm)
	if code != 0 {
		t.Errorf("expected dispatch([\"doctor\"]) exit code 0, got %d", code)
	}
}

func TestCLI_ExecuteRemove_AgentFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	reg.Register(&mockAdapter{name: "mock"})

	_, err := pm.EnsureProfile("shared")
	if err != nil {
		t.Fatalf("failed to ensure profile: %v", err)
	}

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("shared", "agy")
	cfg.AddProfileAgent("shared", "mock")
	_ = config.SaveConfig(cfg)

	// 1. Invalid usage with no arguments
	if code := executeRemove(reg, pm, []string{}); code != 1 {
		t.Errorf("expected exit code 1 for empty args, got %d", code)
	}

	// 2. Removing an unassociated agent should fail and NOT delete the directory
	if code := executeRemove(reg, pm, []string{"unassociated-agent", "shared"}); code != 1 {
		t.Fatalf("expected exit code 1 when removing unassociated agent, got %d", code)
	}
	if _, err := os.Stat(pm.ProfileDir("shared")); os.IsNotExist(err) {
		t.Fatalf("profile directory was wrongly deleted when removing unassociated agent")
	}

	// 3. Removing non-existent profile returns 1
	if code := executeRemove(reg, pm, []string{"non-existent-profile"}); code != 1 {
		t.Fatalf("expected exit code 1 when removing non-existent profile, got %d", code)
	}
	if code := executeRemove(reg, pm, []string{"mock", "non-existent-profile"}); code != 1 {
		t.Fatalf("expected exit code 1 when removing agent from non-existent profile, got %d", code)
	}

	// 4. Remove one agent from multi-agent profile
	if code := executeRemove(reg, pm, []string{"mock", "shared"}); code != 0 {
		t.Fatalf("expected exit code 0 when removing agent, got %d", code)
	}

	// Directory should still exist
	if _, err := os.Stat(pm.ProfileDir("shared")); os.IsNotExist(err) {
		t.Fatalf("profile directory was removed prematurely while 'agy' still attached")
	}

	cfgReload, _ := config.LoadConfig()
	if cfgReload.HasAgent("shared", "mock") {
		t.Errorf("expected 'mock' to be removed from config")
	}
	if !cfgReload.HasAgent("shared", "agy") {
		t.Errorf("expected 'agy' to remain in config")
	}

	// 5. Remove remaining agent: directory should now be cleaned up
	if code := executeRemove(reg, pm, []string{"agy", "shared"}); code != 0 {
		t.Fatalf("expected exit code 0 when removing final agent, got %d", code)
	}

	if _, err := os.Stat(pm.ProfileDir("shared")); !os.IsNotExist(err) {
		t.Fatalf("expected profile directory to be deleted after removing all agents")
	}

	cfgReload2, _ := config.LoadConfig()
	if _, exists := cfgReload2.Profiles["shared"]; exists {
		t.Fatalf("expected profile 'shared' to be deleted from config")
	}

	// 6. Remove via profile only
	_, _ = pm.EnsureProfile("direct-del")
	cfgReload2.AddProfileAgent("direct-del", "agy")
	_ = config.SaveConfig(cfgReload2)

	if code := executeRemove(reg, pm, []string{"direct-del"}); code != 0 {
		t.Fatalf("expected exit code 0 for direct remove, got %d", code)
	}
	if _, err := os.Stat(pm.ProfileDir("direct-del")); !os.IsNotExist(err) {
		t.Fatalf("expected profile directory to be deleted")
	}
}

func TestCLI_ExecuteRemove_AliasResolution(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)

	_, err := pm.EnsureProfile("alias-prof")
	if err != nil {
		t.Fatalf("failed to ensure profile: %v", err)
	}

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("alias-prof", "agy")
	_ = config.SaveConfig(cfg)

	// Removing with alias "antigravity" resolves to canonical "agy"
	if code := executeRemove(reg, pm, []string{"antigravity", "alias-prof"}); code != 0 {
		t.Fatalf("expected exit code 0 when removing with alias, got %d", code)
	}

	cfgReload, _ := config.LoadConfig()
	if cfgReload.HasAgent("alias-prof", "agy") {
		t.Errorf("expected 'agy' to be removed from profile 'alias-prof' via alias 'antigravity'")
	}

	if _, err := os.Stat(pm.ProfileDir("alias-prof")); !os.IsNotExist(err) {
		t.Fatalf("expected profile directory to be deleted after removing sole agent via alias")
	}

	// Also verify via dispatch: aim remove antigravity <prof>
	_, err = pm.EnsureProfile("dispatch-alias-prof")
	if err != nil {
		t.Fatalf("failed to ensure profile: %v", err)
	}
	cfgReload.AddProfileAgent("dispatch-alias-prof", "agy")
	_ = config.SaveConfig(cfgReload)

	code := dispatch([]string{"remove", "antigravity", "dispatch-alias-prof"}, reg, pm)
	if code != 0 {
		t.Fatalf("expected dispatch exit code 0 for alias remove, got %d", code)
	}
	cfgReload2, _ := config.LoadConfig()
	if cfgReload2.HasAgent("dispatch-alias-prof", "agy") {
		t.Errorf("expected 'agy' to be removed via dispatch alias remove")
	}
}

func TestCLI_ExecuteClone(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aim-cli-clone-test-*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Setenv("AIM_HOME", tempDir)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	reg.Register(&mockAdapter{name: "mock"})

	pm := profile.NewProfileManager(tempDir)
	srcDir, _ := pm.EnsureProfile("orig")

	// Write mock settings and sensitive token
	_ = os.WriteFile(filepath.Join(srcDir, "custom.conf"), []byte("env=prod"), 0644)
	tokenDir := filepath.Join(srcDir, ".gemini", "antigravity-cli")
	_ = os.MkdirAll(tokenDir, 0700)
	_ = os.WriteFile(filepath.Join(tokenDir, "token.json"), []byte(`{"secret": true}`), 0600)

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("orig", "agy")
	cfg.AddProfileAgent("orig", "mock")
	_ = config.SaveConfig(cfg)

	// 1. Invalid usage with < 2 args
	if code := executeClone(reg, pm, []string{}); code != 1 {
		t.Errorf("expected exit code 1 for empty args, got %d", code)
	}
	if code := executeClone(reg, pm, []string{"orig"}); code != 1 {
		t.Errorf("expected exit code 1 for 1 arg, got %d", code)
	}

	// 2. Clone all agents: aim clone orig cloned-all
	if code := executeClone(reg, pm, []string{"orig", "cloned-all"}); code != 0 {
		t.Fatalf("expected exit code 0 for 2-arg clone, got %d", code)
	}
	dstAllDir := pm.ProfileDir("cloned-all")
	if _, err := os.Stat(dstAllDir); err != nil {
		t.Fatalf("expected cloned-all directory to exist: %v", err)
	}
	// Non-sensitive file copied
	if data, err := os.ReadFile(filepath.Join(dstAllDir, "custom.conf")); err != nil || string(data) != "env=prod" {
		t.Errorf("expected custom.conf to be copied, got %s (err: %v)", string(data), err)
	}
	// Token NOT copied
	if _, err := os.Stat(filepath.Join(dstAllDir, ".gemini", "antigravity-cli", "token.json")); !os.IsNotExist(err) {
		t.Errorf("expected token.json NOT to be copied")
	}
	// Config inherited both agents
	cfgReload, _ := config.LoadConfig()
	if !cfgReload.HasAgent("cloned-all", "agy") || !cfgReload.HasAgent("cloned-all", "mock") {
		t.Errorf("expected cloned-all to have agy and mock, got %v", cfgReload.GetProfileAgents("cloned-all"))
	}

	// 3. Clone with specific agent: aim clone mock orig cloned-mock
	if code := executeClone(reg, pm, []string{"mock", "orig", "cloned-mock"}); code != 0 {
		t.Fatalf("expected exit code 0 for agent-specific clone, got %d", code)
	}
	cfgReload2, _ := config.LoadConfig()
	if !cfgReload2.HasAgent("cloned-mock", "mock") || cfgReload2.HasAgent("cloned-mock", "agy") {
		t.Errorf("expected cloned-mock to only have mock, got %v", cfgReload2.GetProfileAgents("cloned-mock"))
	}

	// 4. Clone with agent alias: aim clone antigravity orig cloned-agy-alias
	code := dispatch([]string{"clone", "antigravity", "orig", "cloned-agy-alias"}, reg, pm)
	if code != 0 {
		t.Fatalf("expected dispatch exit code 0 for alias clone, got %d", code)
	}
	cfgReload3, _ := config.LoadConfig()
	if !cfgReload3.HasAgent("cloned-agy-alias", "agy") {
		t.Errorf("expected cloned-agy-alias to have agy")
	}

	// 5. Clone unassociated agent
	if code := executeClone(reg, pm, []string{"unassociated", "orig", "fail-dst"}); code != 1 {
		t.Errorf("expected exit code 1 for unassociated agent clone, got %d", code)
	}

	// 6. Clone non-existent profile
	if code := executeClone(reg, pm, []string{"ghost", "fail-dst"}); code != 1 {
		t.Errorf("expected exit code 1 for non-existent source clone, got %d", code)
	}
}

func TestCLI_RenameCommandNotRegistered(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("AIM_HOME", tempDir)
	pm := profile.NewProfileManager(tempDir)
	reg := agents.NewRegistry()

	// Profile renaming is exclusively an interactive TUI action ([m] / [R])
	code := dispatch([]string{"rename", "a", "b"}, reg, pm)
	if code == 0 {
		t.Errorf("expected dispatch to fail for unregistered 'rename' command, got %d", code)
	}
}

func TestCLI_ExecuteRun_UpdatesAgentTag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
		args:       []string{"-c", "exit 0"},
	})

	code := executeRun(reg, pm, "mock", "runprof", nil)
	if code != 0 {
		t.Fatalf("expected run code 0, got %d", code)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if !cfg.HasAgent("runprof", "mock") {
		t.Errorf("expected 'mock' to be tagged on 'runprof' after executeRun")
	}
}

func TestCLI_ExecuteRun_WithConfigOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	reg := agents.NewRegistry()

	reg.Register(&mockAdapter{
		name:       "mock",
		binaryPath: "/bin/sh",
	})

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("custom_prof", "mock")
	cfg.SetProfileEnv("custom_prof", map[string]string{
		"AIM_CUSTOM_VAR": "hello_aim_override",
	})
	cfg.SetProfileArgs("custom_prof", []string{"-c"})
	_ = config.SaveConfig(cfg)

	code := executeRun(reg, pm, "mock", "custom_prof", []string{`[ "$AIM_CUSTOM_VAR" = "hello_aim_override" ] && exit 0 || exit 1`})
	if code != 0 {
		t.Fatalf("expected run code 0 when custom env and args are injected, got %d", code)
	}
}

func TestCLI_Doctor_ConfigOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	_, _ = pm.EnsureProfile("overridden")
	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{name: "mock", binaryPath: "/bin/sh"})

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("overridden", "mock")
	cfg.SetProfileEnv("overridden", map[string]string{"K": "V"})
	cfg.SetProfileArgs("overridden", []string{"--arg1"})
	_ = config.SaveConfig(cfg)

	stdout, _ := captureOutput(t, func() {
		runDoctor(reg, pm, "mock")
	})

	if !strings.Contains(stdout, "1 custom env var(s) configured") {
		t.Errorf("expected doctor output to mention 1 custom env var(s), got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "1 custom launch arg(s) configured") {
		t.Errorf("expected doctor output to mention 1 custom launch arg(s), got:\n%s", stdout)
	}
}

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	fn()

	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	out := <-outC
	errStr := <-errC
	_ = rOut.Close()
	_ = rErr.Close()
	return out, errStr
}

func TestCLIDispatch_Completion(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	reg := agents.NewRegistry()

	// 1. Missing shell argument -> prints usage to stdout and returns 1
	var code int
	out, _ := captureOutput(t, func() {
		code = dispatch([]string{"completion"}, reg, pm)
	})
	if code != 1 {
		t.Errorf("dispatch([\"completion\"]) = %d, want 1", code)
	}
	if !strings.Contains(out, "Usage: aim completion <zsh|bash|fish>") {
		t.Errorf("expected usage output, got %q", out)
	}

	// 2. Unsupported shell -> prints error to stderr and returns 1
	_, errOut := captureOutput(t, func() {
		code = dispatch([]string{"completion", "unsupported"}, reg, pm)
	})
	if code != 1 {
		t.Errorf("dispatch([\"completion\", \"unsupported\"]) = %d, want 1", code)
	}
	if !strings.Contains(errOut, "unsupported shell") {
		t.Errorf("expected error output for unsupported shell, got %q", errOut)
	}

	// 3. Supported shells -> prints script to stdout and returns 0
	shellSignatures := map[string]string{
		"zsh":  "#compdef aim",
		"bash": "complete -o default",
		"fish": "complete -c aim",
	}
	for sh, sig := range shellSignatures {
		out, _ := captureOutput(t, func() {
			code = dispatch([]string{"completion", sh}, reg, pm)
		})
		if code != 0 {
			t.Errorf("dispatch([\"completion\", %q]) = %d, want 0", sh, code)
		}
		if !strings.Contains(out, sig) {
			t.Errorf("expected script for %q to contain %q, got: %s", sh, sig, out)
		}
	}
}

func TestCLIDispatch_HiddenComplete(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{name: "mock"})

	_, _ = pm.EnsureProfile("work")
	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("work", "mock")
	_ = config.SaveConfig(cfg)

	// 1. Root completions -> returns 0 and outputs subcommands only
	var code int
	out, _ := captureOutput(t, func() {
		code = dispatch([]string{"__complete"}, reg, pm)
	})
	if code != 0 {
		t.Errorf("dispatch([\"__complete\"]) = %d, want 0", code)
	}
	if !strings.Contains(out, "run\t") {
		t.Errorf("expected run subcommand in completions, got %q", out)
	}
	if !strings.Contains(out, "completion\t") {
		t.Errorf("expected completion subcommand in completions, got %q", out)
	}
	if strings.Contains(out, "mock\t") {
		t.Errorf("did not expect mock agent in root completions, got %q", out)
	}

	// 2. Subcommand agent completions
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"__complete", "run", ""}, reg, pm)
	})
	if code != 0 {
		t.Errorf("dispatch([\"__complete\", \"run\", \"\"]) = %d, want 0", code)
	}
	if !strings.Contains(out, "mock") {
		t.Errorf("expected mock agent in agent completions, got %q", out)
	}

	// 3. Subcommand profile completions
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"__complete", "run", "mock", ""}, reg, pm)
	})
	if code != 0 {
		t.Errorf("dispatch([\"__complete\", \"run\", \"mock\", \"\"]) = %d, want 0", code)
	}
	if !strings.Contains(out, "work") {
		t.Errorf("expected 'work' profile completion, got %q", out)
	}
}

func TestCLIUsageCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("personal")

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("work", "mock")
	cfg.AddProfileAgent("personal", "mock")
	_ = config.SaveConfig(cfg)

	reg := agents.NewRegistry()
	reg.Register(&mockAdapter{
		name: "mock",
	})

	// 1. Basic aim usage tabular output
	var code int
	out, _ := captureOutput(t, func() {
		code = dispatch([]string{"usage"}, reg, pm)
	})
	if code != 0 {
		t.Fatalf("dispatch usage returned %d, want 0", code)
	}
	for _, header := range []string{"AGENT", "PROFILE", "STATUS", "5H LIMIT", "5H RESET", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected header %q in usage output:\n%s", header, out)
		}
	}
	if !strings.Contains(out, "mock") || !strings.Contains(out, "work") {
		t.Errorf("expected mock and work in usage output:\n%s", out)
	}

	// 2. aim usage --json output
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"usage", "--json"}, reg, pm)
	})
	if code != 0 {
		t.Fatalf("dispatch usage --json returned %d, want 0", code)
	}
	var reports []usage.Report
	if err := json.Unmarshal([]byte(out), &reports); err != nil {
		t.Fatalf("failed to parse json output: %v\nOutput: %s", err, out)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports in json, got %d", len(reports))
	}
	if reports[0].Agent != "mock" {
		t.Errorf("expected agent 'mock', got %s", reports[0].Agent)
	}

	// 3. aim usage with agent filter
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"usage", "mock"}, reg, pm)
	})
	if code != 0 {
		t.Fatalf("dispatch usage mock returned %d, want 0", code)
	}
	if !strings.Contains(out, "work") {
		t.Errorf("expected work in filtered usage output:\n%s", out)
	}

	// 4. aim usage with agent and profile filter
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"usage", "mock", "work"}, reg, pm)
	})
	if code != 0 {
		t.Fatalf("dispatch usage mock work returned %d, want 0", code)
	}
	if !strings.Contains(out, "work") || strings.Contains(out, "personal") {
		t.Errorf("expected only work in filtered output:\n%s", out)
	}

	// 5. aim usage with unknown agent returns error code 1
	out, errOut := captureOutput(t, func() {
		code = dispatch([]string{"usage", "unknown"}, reg, pm)
	})
	if code != 1 {
		t.Errorf("dispatch usage unknown returned %d, want 1", code)
	}
	if !strings.Contains(errOut, "unknown agent") && !strings.Contains(out, "unknown agent") {
		t.Errorf("expected unknown agent error, got out=%q, errOut=%q", out, errOut)
	}

	// 6. aim usage with no profiles configured
	emptyPM := profile.NewProfileManager(t.TempDir())
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"usage"}, reg, emptyPM)
	})
	if code != 0 {
		t.Errorf("dispatch usage with no profiles returned %d, want 0", code)
	}
	if !strings.Contains(out, "No profiles configured") {
		t.Errorf("expected 'No profiles configured' message, got %s", out)
	}

	// 6b. aim usage --json with no profiles configured returns "[]"
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"usage", "--json"}, reg, emptyPM)
	})
	if code != 0 {
		t.Errorf("dispatch usage --json with no profiles returned %d, want 0", code)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("expected '[]' for usage --json with no profiles, got %q", out)
	}

	// 6c. aim usage --json with agent having no profiles returns "[]"
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"usage", "--json", "mock"}, reg, emptyPM)
	})
	if code != 0 {
		t.Errorf("dispatch usage --json mock with no profiles returned %d, want 0", code)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("expected '[]' for usage --json with no profiles for agent, got %q", out)
	}

	// 6d. aim usage --json with nil ProfileManager returns "[]"
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"usage", "--json"}, reg, nil)
	})
	if code != 0 {
		t.Errorf("dispatch usage --json with nil pm returned %d, want 0", code)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("expected '[]' for usage --json with nil pm, got %q", out)
	}

	// 7. aim usage with --refresh / -r
	out, _ = captureOutput(t, func() {
		code = dispatch([]string{"usage", "-r"}, reg, pm)
	})
	if code != 0 {
		t.Fatalf("dispatch usage -r returned %d, want 0", code)
	}
}

func TestCLIListWithUsageBadges(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIM_HOME", tmpDir)
	pm := profile.NewProfileManager(tmpDir)
	_, _ = pm.EnsureProfile("work")

	cfg, _ := config.LoadConfig()
	cfg.AddProfileAgent("work", "agy")
	_ = config.SaveConfig(cfg)

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())

	// Populate cache with report having 5h and weekly windows
	cache := usage.NewCacheStore(tmpDir, 10*time.Minute)
	rep := usage.Report{
		Agent:     "agy",
		Profile:   "work",
		Status:    usage.StatusOK,
		FetchedAt: time.Now(),
		Windows: []usage.LimitWindow{
			{Name: "Five Hour Limit", RemainingPct: 82, ResetsAt: time.Now().Add(2*time.Hour + 15*time.Minute + 5*time.Second)},
			{Name: "Weekly Limit", RemainingPct: 90, ResetsAt: time.Now().Add(5*24*time.Hour + 14*time.Hour + 5*time.Second)},
		},
	}
	_ = cache.Put(rep)

	// Test aim list (all profiles) displays badge
	out, _ := captureOutput(t, func() {
		_ = runList(reg, pm, "")
	})
	if !strings.Contains(out, "(5h: 82% [2h 15m], wk: 90% [5d 14h])") {
		t.Errorf("expected multi-window badge in aim list output, got:\n%s", out)
	}

	// Test aim list agy (agent-filtered) displays badge
	out, _ = captureOutput(t, func() {
		_ = runList(reg, pm, "agy")
	})
	if !strings.Contains(out, "(5h: 82% [2h 15m], wk: 90% [5d 14h])") {
		t.Errorf("expected multi-window badge in aim list agy output, got:\n%s", out)
	}

	// Single window badge
	_, _ = pm.EnsureProfile("single-win")
	cfg.AddProfileAgent("single-win", "agy")
	_ = config.SaveConfig(cfg)
	repSingle := usage.Report{
		Agent:     "agy",
		Profile:   "single-win",
		Status:    usage.StatusOK,
		FetchedAt: time.Now(),
		Windows: []usage.LimitWindow{
			{Name: "Weekly Limit", RemainingPct: 75, ResetsAt: time.Now().Add(3*24*time.Hour + 8*time.Hour + 5*time.Second)},
		},
	}
	_ = cache.Put(repSingle)

	out, _ = captureOutput(t, func() {
		_ = runList(reg, pm, "agy")
	})
	if !strings.Contains(out, "(wk: 75% [3d 8h])") {
		t.Errorf("expected single-window badge '(wk: 75%% [3d 8h])' in aim list agy, got:\n%s", out)
	}

	// Non-operational: no credentials
	_, _ = pm.EnsureProfile("unauthed")
	cfg.AddProfileAgent("unauthed", "agy")
	_ = config.SaveConfig(cfg)
	repUnauth := usage.Report{
		Agent:     "agy",
		Profile:   "unauthed",
		Status:    usage.StatusUnknown,
		FetchedAt: time.Now(),
		Error:     "no credentials",
	}
	_ = cache.Put(repUnauth)

	out, _ = captureOutput(t, func() {
		_ = runList(reg, pm, "agy")
	})
	if !strings.Contains(out, "unauthed") || !strings.Contains(out, "[no credentials]") {
		t.Errorf("expected '[no credentials]' badge in aim list agy, got:\n%s", out)
	}

	// Non-operational: offline
	_, _ = pm.EnsureProfile("offline-prof")
	cfg.AddProfileAgent("offline-prof", "agy")
	_ = config.SaveConfig(cfg)
	repOffline := usage.Report{
		Agent:     "agy",
		Profile:   "offline-prof",
		Status:    usage.StatusUnknown,
		Summary:   "Offline",
		FetchedAt: time.Now(),
		Error:     "connection refused",
	}
	_ = cache.Put(repOffline)

	out, _ = captureOutput(t, func() {
		_ = runList(reg, pm, "agy")
	})
	if !strings.Contains(out, "offline-prof") || !strings.Contains(out, "[offline]") {
		t.Errorf("expected '[offline]' badge in aim list agy, got:\n%s", out)
	}
}

func TestCLIPrewarmCommand(t *testing.T) {
	tempBase := t.TempDir()
	origBaseDir := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origBaseDir) }()
	_ = os.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	_, _ = pm.EnsureProfile("prewarm-prof")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("prewarm-prof", "mock")
	_ = config.SaveConfig(cfg)

	reg := agents.NewRegistry()
	mockAd := &mockAdapter{
		name: "mock",
	}
	reg.Register(mockAd)

	// Run prewarm via dispatch
	code := dispatch([]string{"__prewarm", "mock"}, reg, pm)
	if code != 0 {
		t.Fatalf("dispatch([__prewarm mock]) failed with exit code %d", code)
	}

	// Verify that report was saved into cache
	cache := usage.NewCacheStore(tempBase, usage.DefaultTTL)
	rep, found := cache.Get("mock", "prewarm-prof")
	if !found {
		t.Fatalf("expected prewarm-prof to be cached after __prewarm")
	}
	if rep.Status != usage.StatusOK {
		t.Errorf("expected StatusOK, got %s", rep.Status)
	}
}

func TestTriggerPrewarmAsync(t *testing.T) {
	tempBase := t.TempDir()

	// Should not panic even if executable cannot be spawned or throttled
	triggerPrewarmAsync(tempBase, "mock")
	// Second call should be throttled by lock
	triggerPrewarmAsync(tempBase, "mock")
}

func TestDebug_DefaultOff(t *testing.T) {
	logger.Reset()
	t.Setenv("AIM_DEBUG", "")

	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	code := dispatch([]string{"whoami"}, reg, pm)
	if code != 0 {
		t.Fatalf("dispatch whoami failed: %d", code)
	}

	if logger.IsDebug() {
		t.Errorf("expected debug to be off by default")
	}
}

func TestDebug_FlagEnablesDebug(t *testing.T) {
	logger.Reset()
	t.Setenv("AIM_DEBUG", "")

	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	code := dispatch([]string{"--debug", "whoami"}, reg, pm)
	if code != 0 {
		t.Fatalf("dispatch --debug whoami failed: %d", code)
	}

	if !logger.IsDebug() {
		t.Errorf("expected debug to be enabled via --debug flag")
	}
}

func TestDebug_EnvVarEnablesDebug(t *testing.T) {
	logger.Reset()
	t.Setenv("AIM_DEBUG", "1")

	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	code := dispatch([]string{"whoami"}, reg, pm)
	if code != 0 {
		t.Fatalf("dispatch whoami failed: %d", code)
	}

	if !logger.IsDebug() {
		t.Errorf("expected debug to be enabled via AIM_DEBUG=1")
	}
}

func TestDebug_ConfigEnablesDebug(t *testing.T) {
	logger.Reset()
	t.Setenv("AIM_DEBUG", "")

	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	cfg := config.NewDefaultConfig()
	cfg.Debug = true
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	pm := profile.NewProfileManager(tempBase)
	reg := agents.NewRegistry()

	code := dispatch([]string{"whoami"}, reg, pm)
	if code != 0 {
		t.Fatalf("dispatch whoami failed: %d", code)
	}

	if !logger.IsDebug() {
		t.Errorf("expected debug to be enabled via config.Debug=true")
	}
}

func TestCLI_Usage_AccountColumn(t *testing.T) {
	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	_, _ = pm.EnsureProfile("userprof")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("userprof", "mock")
	_ = config.SaveConfig(cfg)

	cache := usage.NewCacheStore(tempBase, usage.DefaultTTL)
	_ = cache.Put(usage.Report{
		Agent:        "mock",
		Profile:      "userprof",
		Status:       usage.StatusOK,
		AccountEmail: "user@example.com",
		Windows: []usage.LimitWindow{
			{Name: "Five Hour", RemainingPct: 80},
		},
		FetchedAt: time.Now(),
	})

	reg := agents.NewRegistry()
	mockAd := &mockAdapter{name: "mock"}
	reg.Register(mockAd)

	out, _ := captureOutput(t, func() {
		_ = executeUsage(reg, pm, []string{"mock", "userprof"}, usageOptions{})
	})

	if !strings.Contains(out, "ACCOUNT") {
		t.Errorf("expected table header to contain 'ACCOUNT', got:\n%s", out)
	}
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("expected table body to contain 'user@example.com', got:\n%s", out)
	}
}

func TestUsageTableColumns(t *testing.T) {
	tests := []struct {
		name                           string
		hasAccount, hasCategory, has5h bool
		wantHeaders, wantRow           []string
	}{
		{"weekly", false, false, false, []string{"AGENT", "PROFILE", "STATUS", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"}, []string{"agent", "profile", "status", "weekly", "weekly reset", "checked"}},
		{"weekly account", true, false, false, []string{"AGENT", "PROFILE", "ACCOUNT", "STATUS", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"}, []string{"agent", "profile", "account", "status", "weekly", "weekly reset", "checked"}},
		{"weekly category", false, true, false, []string{"AGENT", "PROFILE", "MODEL", "STATUS", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"}, []string{"agent", "profile", "category", "status", "weekly", "weekly reset", "checked"}},
		{"weekly account category", true, true, false, []string{"AGENT", "PROFILE", "ACCOUNT", "MODEL", "STATUS", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"}, []string{"agent", "profile", "account", "category", "status", "weekly", "weekly reset", "checked"}},
		{"hourly", false, false, true, []string{"AGENT", "PROFILE", "STATUS", "5H LIMIT", "5H RESET", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"}, []string{"agent", "profile", "status", "primary", "primary reset", "weekly", "weekly reset", "checked"}},
		{"hourly account", true, false, true, []string{"AGENT", "PROFILE", "ACCOUNT", "STATUS", "5H LIMIT", "5H RESET", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"}, []string{"agent", "profile", "account", "status", "primary", "primary reset", "weekly", "weekly reset", "checked"}},
		{"hourly category", false, true, true, []string{"AGENT", "PROFILE", "MODEL", "STATUS", "5H LIMIT", "5H RESET", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"}, []string{"agent", "profile", "category", "status", "primary", "primary reset", "weekly", "weekly reset", "checked"}},
		{"hourly account category", true, true, true, []string{"AGENT", "PROFILE", "ACCOUNT", "MODEL", "STATUS", "5H LIMIT", "5H RESET", "WEEKLY LIMIT", "WEEKLY RESET", "CHECKED"}, []string{"agent", "profile", "account", "category", "status", "primary", "primary reset", "weekly", "weekly reset", "checked"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columns := usageTableColumns{hasAccount: tt.hasAccount, hasCategory: tt.hasCategory, has5h: tt.has5h}
			if got := columns.headers(); !reflect.DeepEqual(got, tt.wantHeaders) {
				t.Errorf("headers = %v; want %v", got, tt.wantHeaders)
			}
			got := columns.row(usageRowValues{agent: "agent", profile: "profile", account: "account", category: "category", status: "status", primary: "primary", primaryReset: "primary reset", weekly: "weekly", weeklyReset: "weekly reset", checked: "checked"})
			if !reflect.DeepEqual(got, tt.wantRow) {
				t.Errorf("row = %v; want %v", got, tt.wantRow)
			}
		})
	}
}

func TestCLI_List_MultiCategoryBadge(t *testing.T) {
	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	_, _ = pm.EnsureProfile("multiprof")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("multiprof", "agy")
	_ = config.SaveConfig(cfg)

	cache := usage.NewCacheStore(tempBase, usage.DefaultTTL)
	_ = cache.Put(usage.Report{
		Agent:   "agy",
		Profile: "multiprof",
		Status:  usage.StatusOK,
		Windows: []usage.LimitWindow{
			{Category: "Gemini", Name: "5h", RemainingPct: 90},
			{Category: "Claude", Name: "5h", RemainingPct: 85},
		},
		FetchedAt: time.Now(),
	})

	reg := agents.DefaultRegistry()
	out, _ := captureOutput(t, func() {
		_ = runList(reg, pm, "agy")
	})

	if !strings.Contains(out, "gemini:") || !strings.Contains(out, "claude:") {
		t.Errorf("expected multi-category list badge with gemini and claude, got:\n%s", out)
	}
}

func TestCLI_CodexIntegration(t *testing.T) {
	tempBase := t.TempDir()
	t.Setenv("AIM_HOME", tempBase)

	pm := profile.NewProfileManager(tempBase)
	pDir, _ := pm.EnsureProfile("work")

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("work", "codex")
	_ = config.SaveConfig(cfg)

	// Mock auth.json in work profile
	codexDir := filepath.Join(pDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)
	_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(`{"tokens":{"access_token":"mock-token"}}`), 0600)

	reg := agents.NewRegistry()
	reg.Register(codex.NewAdapter())

	// 1. Test 'aim list codex'
	outList, _ := captureOutput(t, func() {
		code := dispatch([]string{"list", "codex"}, reg, pm)
		if code != 0 {
			t.Errorf("expected exit code 0 for 'list codex', got %d", code)
		}
	})
	if !strings.Contains(outList, "work") {
		t.Errorf("expected 'work' in list output, got:\n%s", outList)
	}

	// 2. Test 'aim doctor codex'
	outDoc, _ := captureOutput(t, func() {
		code := dispatch([]string{"doctor", "codex"}, reg, pm)
		if code != 0 {
			t.Errorf("expected exit code 0 for 'doctor codex', got %d", code)
		}
	})
	if !strings.Contains(outDoc, "Codex") {
		t.Errorf("expected doctor output to mention Codex, got:\n%s", outDoc)
	}

	// 3. Test 'aim clone codex work cloned-work'
	outClone, _ := captureOutput(t, func() {
		code := dispatch([]string{"clone", "codex", "work", "cloned-work"}, reg, pm)
		if code != 0 {
			t.Errorf("expected exit code 0 for 'clone codex', got %d", code)
		}
	})
	if !strings.Contains(outClone, "Cloned") {
		t.Errorf("expected clone success output, got:\n%s", outClone)
	}
	cfgReloaded, _ := config.LoadConfig()
	if !cfgReloaded.HasAgent("cloned-work", "codex") {
		t.Errorf("expected cloned profile to have codex agent")
	}
}
