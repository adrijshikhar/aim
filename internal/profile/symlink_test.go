package profile

import (
	"os"
	"path/filepath"
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
