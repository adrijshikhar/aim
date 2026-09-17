package profile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureDotfiles(t *testing.T) {
	fakeHome, err := os.MkdirTemp("", "aim-fake-home-*")
	if err != nil {
		t.Fatalf("fake home error: %v", err)
	}
	defer os.RemoveAll(fakeHome)

	profileDir, err := os.MkdirTemp("", "aim-test-profile-*")
	if err != nil {
		t.Fatalf("profile dir error: %v", err)
	}
	defer os.RemoveAll(profileDir)

	gitconfig := filepath.Join(fakeHome, ".gitconfig")
	if err := os.WriteFile(gitconfig, []byte("[user]\nname = Test\n"), 0644); err != nil {
		t.Fatalf("write gitconfig error: %v", err)
	}

	sshDir := filepath.Join(fakeHome, ".ssh")
	if err := os.Mkdir(sshDir, 0700); err != nil {
		t.Fatalf("mkdir ssh error: %v", err)
	}

	if err := EnsureDotfiles(fakeHome, profileDir); err != nil {
		t.Fatalf("EnsureDotfiles failed: %v", err)
	}

	targetGitconfig := filepath.Join(profileDir, ".gitconfig")
	fi, err := os.Lstat(targetGitconfig)
	if err != nil {
		t.Fatalf("symlinked gitconfig does not exist: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .gitconfig to be a symlink")
	}

	targetSSH := filepath.Join(profileDir, ".ssh")
	fiSSH, err := os.Lstat(targetSSH)
	if err != nil {
		t.Fatalf("symlinked .ssh does not exist: %v", err)
	}
	if fiSSH.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .ssh to be a symlink")
	}
}

func TestEnsureDotfiles_NonDestructive(t *testing.T) {
	fakeHome, err := os.MkdirTemp("", "aim-fake-home-nd-*")
	if err != nil {
		t.Fatalf("fake home error: %v", err)
	}
	defer os.RemoveAll(fakeHome)

	profileDir, err := os.MkdirTemp("", "aim-test-profile-nd-*")
	if err != nil {
		t.Fatalf("profile dir error: %v", err)
	}
	defer os.RemoveAll(profileDir)

	// Create existing file in profileDir
	customGitconfig := filepath.Join(profileDir, ".gitconfig")
	if err := os.WriteFile(customGitconfig, []byte("[user]\nname = Custom\n"), 0644); err != nil {
		t.Fatalf("write custom gitconfig error: %v", err)
	}

	// Create .gitconfig in fakeHome
	if err := os.WriteFile(filepath.Join(fakeHome, ".gitconfig"), []byte("[user]\nname = Home\n"), 0644); err != nil {
		t.Fatalf("write home gitconfig error: %v", err)
	}

	// Running EnsureDotfiles should not overwrite customGitconfig
	if err := EnsureDotfiles(fakeHome, profileDir); err != nil {
		t.Fatalf("EnsureDotfiles failed: %v", err)
	}

	fi, err := os.Lstat(customGitconfig)
	if err != nil {
		t.Fatalf("custom gitconfig stat failed: %v", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected custom .gitconfig NOT to be overwritten by symlink")
	}

	content, err := os.ReadFile(customGitconfig)
	if err != nil {
		t.Fatalf("read custom gitconfig error: %v", err)
	}
	if string(content) != "[user]\nname = Custom\n" {
		t.Errorf("content was overwritten: %s", string(content))
	}
}

func TestEnsureDotfiles_ErrorPropagation(t *testing.T) {
	fakeHome, err := os.MkdirTemp("", "aim-fake-home-err-*")
	if err != nil {
		t.Fatalf("fake home error: %v", err)
	}
	defer os.RemoveAll(fakeHome)

	gitconfig := filepath.Join(fakeHome, ".gitconfig")
	if err := os.WriteFile(gitconfig, []byte("test"), 0644); err != nil {
		t.Fatalf("write gitconfig error: %v", err)
	}

	// Point profileDir to a non-existent path where parent directory does not exist
	invalidProfileDir := filepath.Join(fakeHome, "nonexistent", "subdir")
	err = EnsureDotfiles(fakeHome, invalidProfileDir)
	if err == nil {
		t.Errorf("expected EnsureDotfiles to return error for non-existent destination directory, got nil")
	}
}

func TestEnsureDotfiles_ComprehensiveDeveloperTools(t *testing.T) {
	fakeHome := t.TempDir()
	profileDir := t.TempDir()

	// Setup host developer tools & configs
	ghDir := filepath.Join(fakeHome, ".config", "gh")
	if err := os.MkdirAll(ghDir, 0755); err != nil {
		t.Fatalf("mkdir gh failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ghDir, "hosts.yml"), []byte("hosts"), 0644); err != nil {
		t.Fatalf("write hosts failed: %v", err)
	}

	npmrc := filepath.Join(fakeHome, ".npmrc")
	if err := os.WriteFile(npmrc, []byte("//registry.npmjs.org/:_authToken=secret"), 0600); err != nil {
		t.Fatalf("write npmrc failed: %v", err)
	}

	dockerDir := filepath.Join(fakeHome, ".docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		t.Fatalf("mkdir docker failed: %v", err)
	}

	cargoDir := filepath.Join(fakeHome, ".cargo")
	if err := os.MkdirAll(cargoDir, 0755); err != nil {
		t.Fatalf("mkdir cargo failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cargoDir, "config.toml"), []byte("[build]"), 0644); err != nil {
		t.Fatalf("write cargo config failed: %v", err)
	}

	// AI agent skills
	agentsDir := filepath.Join(fakeHome, ".agents", "skills")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("mkdir .agents/skills failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "test.md"), []byte("# Skill"), 0644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}

	// Terminal & Statusline tooling (cxstatusline)
	cxDir := filepath.Join(fakeHome, ".config", "cxstatusline")
	if err := os.MkdirAll(cxDir, 0755); err != nil {
		t.Fatalf("mkdir cxstatusline failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cxDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("write settings failed: %v", err)
	}

	// Run EnsureDotfiles with an extra custom path
	customFile := filepath.Join(fakeHome, ".custom-dev-token")
	if err := os.WriteFile(customFile, []byte("token"), 0600); err != nil {
		t.Fatalf("write custom token failed: %v", err)
	}

	if err := EnsureDotfiles(fakeHome, profileDir, ".custom-dev-token"); err != nil {
		t.Fatalf("EnsureDotfiles failed: %v", err)
	}

	// Verify .config/gh symlink
	targetGH := filepath.Join(profileDir, ".config", "gh")
	fi, err := os.Lstat(targetGH)
	if err != nil {
		t.Fatalf("expected .config/gh to exist: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .config/gh to be a symlink")
	}

	// Verify .npmrc symlink
	targetNpm := filepath.Join(profileDir, ".npmrc")
	fiNpm, err := os.Lstat(targetNpm)
	if err != nil {
		t.Fatalf("expected .npmrc to exist: %v", err)
	}
	if fiNpm.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .npmrc to be a symlink")
	}

	// Verify .docker symlink
	targetDocker := filepath.Join(profileDir, ".docker")
	fiDocker, err := os.Lstat(targetDocker)
	if err != nil {
		t.Fatalf("expected .docker to exist: %v", err)
	}
	if fiDocker.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .docker to be a symlink")
	}

	// Verify .cargo/config.toml symlink
	targetCargo := filepath.Join(profileDir, ".cargo", "config.toml")
	fiCargo, err := os.Lstat(targetCargo)
	if err != nil {
		t.Fatalf("expected .cargo/config.toml to exist: %v", err)
	}
	if fiCargo.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .cargo/config.toml to be a symlink")
	}

	// Verify .agents symlink
	targetAgents := filepath.Join(profileDir, ".agents")
	fiAgents, err := os.Lstat(targetAgents)
	if err != nil {
		t.Fatalf("expected .agents to exist: %v", err)
	}
	if fiAgents.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .agents to be a symlink")
	}

	// Verify .config/cxstatusline symlink
	targetCX := filepath.Join(profileDir, ".config", "cxstatusline")
	fiCX, err := os.Lstat(targetCX)
	if err != nil {
		t.Fatalf("expected .config/cxstatusline to exist: %v", err)
	}
	if fiCX.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .config/cxstatusline to be a symlink")
	}

	// Verify custom path symlink
	targetCustom := filepath.Join(profileDir, ".custom-dev-token")
	fiCustom, err := os.Lstat(targetCustom)
	if err != nil {
		t.Fatalf("expected .custom-dev-token to exist: %v", err)
	}
	if fiCustom.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected .custom-dev-token to be a symlink")
	}
}

func TestEnsureDotfiles_DarwinKeychains(t *testing.T) {
	fakeHome := t.TempDir()
	profileDir := t.TempDir()

	keychainsDir := filepath.Join(fakeHome, "Library", "Keychains")
	if err := os.MkdirAll(keychainsDir, 0755); err != nil {
		t.Fatalf("mkdir keychains failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(keychainsDir, "login.keychain-db"), []byte("mock-db"), 0600); err != nil {
		t.Fatalf("write login keychain failed: %v", err)
	}

	if err := EnsureDotfiles(fakeHome, profileDir); err != nil {
		t.Fatalf("EnsureDotfiles failed: %v", err)
	}

	targetKeychains := filepath.Join(profileDir, "Library", "Keychains")
	fi, err := os.Lstat(targetKeychains)
	if runtime.GOOS == "darwin" {
		if err != nil {
			t.Fatalf("expected Library/Keychains to exist on darwin: %v", err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected Library/Keychains to be a symlink")
		}
	} else {
		// On non-darwin, Library/Keychains should not be linked by default
		if err == nil {
			t.Errorf("expected Library/Keychains not to be bridged on non-darwin OS")
		}
	}
}

func TestProfileManager_EnsureAllProfilesDotfiles(t *testing.T) {
	baseDir := t.TempDir()
	fakeHome := t.TempDir()
	t.Setenv("AIM_REAL_HOME", fakeHome)

	// Create host dotfile
	if err := os.WriteFile(filepath.Join(fakeHome, ".gitconfig"), []byte("[user]"), 0644); err != nil {
		t.Fatalf("write gitconfig failed: %v", err)
	}

	pm := NewProfileManager(baseDir)
	if _, err := pm.EnsureProfile("prof1"); err != nil {
		t.Fatalf("EnsureProfile prof1 failed: %v", err)
	}
	if _, err := pm.EnsureProfile("prof2"); err != nil {
		t.Fatalf("EnsureProfile prof2 failed: %v", err)
	}

	// Now add another file on host after profiles were created
	ghDir := filepath.Join(fakeHome, ".config", "gh")
	if err := os.MkdirAll(ghDir, 0755); err != nil {
		t.Fatalf("mkdir gh failed: %v", err)
	}

	// Run EnsureAllProfilesDotfiles
	if err := pm.EnsureAllProfilesDotfiles(); err != nil {
		t.Fatalf("EnsureAllProfilesDotfiles failed: %v", err)
	}

	for _, p := range []string{"prof1", "prof2"} {
		destGH := filepath.Join(pm.ProfileDir(p), ".config", "gh")
		fi, err := os.Lstat(destGH)
		if err != nil {
			t.Fatalf("expected %s in %s to exist: %v", destGH, p, err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected %s to be symlink", destGH)
		}
	}
}

func TestEnsureDotfiles_SecurityBoundaries(t *testing.T) {
	fakeHome := t.TempDir()
	profileDir := t.TempDir()

	// Create dangerous paths on host
	aimDir := filepath.Join(fakeHome, ".aim")
	if err := os.MkdirAll(aimDir, 0755); err != nil {
		t.Fatalf("mkdir .aim failed: %v", err)
	}

	geminiDir := filepath.Join(fakeHome, ".gemini")
	if err := os.MkdirAll(geminiDir, 0755); err != nil {
		t.Fatalf("mkdir .gemini failed: %v", err)
	}

	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("mkdir .claude failed: %v", err)
	}

	outsideDir := filepath.Join(fakeHome, "outside")
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("mkdir outside failed: %v", err)
	}

	dangerousExtraPaths := []string{
		"../outside",
		"../../outside",
		"/etc/passwd",
		".aim",
		".aim/config.json",
		".gemini",
		".gemini/antigravity-cli",
		".claude",
		".claude.json",
		".codex",
	}

	if err := EnsureDotfiles(fakeHome, profileDir, dangerousExtraPaths...); err != nil {
		t.Fatalf("EnsureDotfiles failed: %v", err)
	}

	// Verify NONE of the dangerous paths were linked
	for _, forbidden := range []string{".aim", ".gemini", ".claude", "outside"} {
		target := filepath.Join(profileDir, forbidden)
		if _, err := os.Lstat(target); err == nil {
			t.Errorf("SECURITY VIOLATION: forbidden path %s was symlinked in profile", forbidden)
		}
	}
}
