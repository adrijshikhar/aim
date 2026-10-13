package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

func TestBindCmd_BasicWorkflow(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempHome)

	testDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	pm := profile.NewProfileManager(tempHome)
	reg := agents.NewRegistry()

	// 1. Resolve initially (empty)
	out, errOut := captureOutput(t, func() {
		code := dispatch([]string{"resolve"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "No profile binding found") {
		t.Errorf("expected no binding output, got: %s (err: %s)", out, errOut)
	}

	// 2. Bind current directory to "work"
	out, errOut = captureOutput(t, func() {
		code := dispatch([]string{"bind", "work"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "Bound directory to profile 'work'") {
		t.Errorf("expected bind confirmation, got: %s (err: %s)", out, errOut)
	}

	// Verify file content
	content, err := os.ReadFile(filepath.Join(testDir, ".aim-profile"))
	if err != nil {
		t.Fatalf("failed to read .aim-profile: %v", err)
	}
	if strings.TrimSpace(string(content)) != "work" {
		t.Errorf("expected 'work' in .aim-profile, got: %q", string(content))
	}

	// 3. Resolve bound directory
	out, _ = captureOutput(t, func() {
		code := dispatch([]string{"resolve"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "Profile:     work") {
		t.Errorf("expected 'Profile:     work', got: %s", out)
	}

	// 4. Resolve with --json
	out, _ = captureOutput(t, func() {
		code := dispatch([]string{"resolve", "--json"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})
	var jsonRes struct {
		Profile    string `json:"profile"`
		SourceFile string `json:"source_file"`
		Directory  string `json:"directory"`
	}
	if err := json.Unmarshal([]byte(out), &jsonRes); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v, raw: %s", err, out)
	}
	if jsonRes.Profile != "work" {
		t.Errorf("expected profile 'work', got %q", jsonRes.Profile)
	}

	// 5. Unbind directory
	out, _ = captureOutput(t, func() {
		code := dispatch([]string{"unbind"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "Removed directory profile binding") {
		t.Errorf("expected unbind confirmation, got: %s", out)
	}

	// 6. Verify file is gone
	if _, err := os.Stat(filepath.Join(testDir, ".aim-profile")); !os.IsNotExist(err) {
		t.Errorf("expected .aim-profile to be removed")
	}
}

func TestBindCmd_RunWithBoundProfile(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempHome)

	testDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	pm := profile.NewProfileManager(tempHome)
	_, _ = pm.EnsureProfile("work")

	reg := agents.NewRegistry()
	mock := &mockAdapter{name: "mock-agent", binaryPath: "/bin/echo"}
	reg.Register(mock)

	// Bind testDir to "work"
	_ = os.WriteFile(filepath.Join(testDir, ".aim-profile"), []byte("work\n"), 0644)

	// Run without specifying profile
	_, errOut := captureOutput(t, func() {
		code := dispatch([]string{"run", "mock-agent", "-v"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})

	if !strings.Contains(errOut, "[aim] Using repository-bound profile 'work' from .aim-profile") {
		t.Errorf("expected notice about bound profile on stderr, got: %s", errOut)
	}
}

func TestBindCmd_AscendingSubdirResolution(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempHome)

	repoDir := t.TempDir()
	subDir := filepath.Join(repoDir, "src", "pkg", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	_ = os.WriteFile(filepath.Join(repoDir, ".aim-profile"), []byte("bound-client\n"), 0644)

	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	pm := profile.NewProfileManager(tempHome)
	reg := agents.NewRegistry()

	out, _ := captureOutput(t, func() {
		code := dispatch([]string{"resolve"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "Profile:     bound-client") {
		t.Errorf("expected ascending resolution to find 'bound-client', got: %s", out)
	}
}

func TestBindCmd_PathTraversalRejected(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempHome)

	testDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	pm := profile.NewProfileManager(tempHome)
	reg := agents.NewRegistry()

	_, errOut := captureOutput(t, func() {
		code := dispatch([]string{"bind", "../../evil"}, reg, pm)
		if code == 0 {
			t.Fatalf("expected non-zero exit code for path traversal profile name")
		}
	})

	if !strings.Contains(errOut, "path traversal not allowed") {
		t.Errorf("expected error about path traversal, got: %s", errOut)
	}

	// Verify .aim-profile was NOT written
	if _, err := os.Stat(filepath.Join(testDir, ".aim-profile")); !os.IsNotExist(err) {
		t.Errorf(".aim-profile must not be created when validation fails")
	}
}

func TestBindCmd_InvalidProfileInFileSurfacedInRun(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempHome)

	testDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	pm := profile.NewProfileManager(tempHome)
	reg := agents.NewRegistry()
	mock := &mockAdapter{name: "mock-agent", binaryPath: "/bin/echo"}
	reg.Register(mock)

	// Write malicious .aim-profile
	_ = os.WriteFile(filepath.Join(testDir, ".aim-profile"), []byte("../../evil\n"), 0644)

	_, errOut := captureOutput(t, func() {
		code := dispatch([]string{"run", "mock-agent"}, reg, pm)
		if code == 0 {
			t.Fatalf("expected non-zero exit code when resolving malicious .aim-profile")
		}
	})

	if !strings.Contains(errOut, "resolving repository profile") || !strings.Contains(errOut, "path traversal not allowed") {
		t.Errorf("expected error about path traversal when running, got: %s", errOut)
	}
}

func TestBindCmd_ProfileNotExistsWarning(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("AIM_HOME")
	defer func() { _ = os.Setenv("AIM_HOME", origHome) }()
	_ = os.Setenv("AIM_HOME", tempHome)

	testDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	pm := profile.NewProfileManager(tempHome)
	reg := agents.NewRegistry()

	_, errOut := captureOutput(t, func() {
		code := dispatch([]string{"bind", "non-existent-profile"}, reg, pm)
		if code != 0 {
			t.Fatalf("expected code 0, got %d", code)
		}
	})

	if !strings.Contains(errOut, "Warning: profile 'non-existent-profile' does not currently exist") {
		t.Errorf("expected warning on stderr, got: %s", errOut)
	}
}
