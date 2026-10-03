package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCloneDirectory_Success(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dstDir: %v", err)
	}

	// 1. Regular file with custom permissions and timestamp
	regFile := filepath.Join(srcDir, "file.txt")
	content := []byte("hello world")
	if err := os.WriteFile(regFile, content, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	modTime := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(regFile, modTime, modTime); err != nil {
		t.Fatalf("failed to set modtime: %v", err)
	}

	// 2. Executable script
	scriptFile := filepath.Join(srcDir, "run.sh")
	if err := os.WriteFile(scriptFile, []byte("#!/bin/sh\necho ok"), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	// 3. Subdirectory with nested file
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0750); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	nestedFile := filepath.Join(subDir, "nested.txt")
	if err := os.WriteFile(nestedFile, []byte("nested content"), 0600); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}
	subModTime := time.Now().Add(-5 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(subDir, subModTime, subModTime); err != nil {
		t.Fatalf("failed to set subdir modtime: %v", err)
	}

	// 4. Symlinks: relative and broken symlink
	symlinkPath := filepath.Join(srcDir, "link.txt")
	if err := os.Symlink("file.txt", symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}
	brokenSymlink := filepath.Join(srcDir, "broken_link.txt")
	if err := os.Symlink("non_existent.txt", brokenSymlink); err != nil {
		t.Fatalf("failed to create broken symlink: %v", err)
	}

	// Run cloneDirectory
	if err := cloneDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("cloneDirectory failed: %v", err)
	}

	// Verify regular file
	dstReg := filepath.Join(dstDir, "file.txt")
	gotContent, err := os.ReadFile(dstReg)
	if err != nil {
		t.Fatalf("failed to read cloned file: %v", err)
	}
	if string(gotContent) != string(content) {
		t.Errorf("expected content %q, got %q", string(content), string(gotContent))
	}
	info, err := os.Stat(dstReg)
	if err != nil {
		t.Fatalf("failed to stat cloned file: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("expected mode 0644, got %v", info.Mode().Perm())
	}
	if info.ModTime().Unix() != modTime.Unix() {
		t.Errorf("expected modtime %v, got %v", modTime.Unix(), info.ModTime().Unix())
	}

	// Verify script permissions
	dstScript := filepath.Join(dstDir, "run.sh")
	scriptInfo, err := os.Stat(dstScript)
	if err != nil {
		t.Fatalf("failed to stat cloned script: %v", err)
	}
	if scriptInfo.Mode().Perm() != 0755 {
		t.Errorf("expected mode 0755, got %v", scriptInfo.Mode().Perm())
	}

	// Verify subdirectory and nested file
	dstSub := filepath.Join(dstDir, "subdir")
	subInfo, err := os.Stat(dstSub)
	if err != nil {
		t.Fatalf("failed to stat cloned subdir: %v", err)
	}
	if !subInfo.IsDir() {
		t.Errorf("expected %s to be directory", dstSub)
	}
	dstNested := filepath.Join(dstSub, "nested.txt")
	nestedContent, err := os.ReadFile(dstNested)
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}
	if string(nestedContent) != "nested content" {
		t.Errorf("expected %q, got %q", "nested content", string(nestedContent))
	}
	nestedInfo, err := os.Stat(dstNested)
	if err != nil {
		t.Fatalf("failed to stat nested file: %v", err)
	}
	if nestedInfo.Mode().Perm() != 0600 {
		t.Errorf("expected nested mode 0600, got %v", nestedInfo.Mode().Perm())
	}

	// Verify symlinks
	dstLink := filepath.Join(dstDir, "link.txt")
	linkInfo, err := os.Lstat(dstLink)
	if err != nil {
		t.Fatalf("failed to lstat cloned symlink: %v", err)
	}
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be symlink", dstLink)
	}
	target, err := os.Readlink(dstLink)
	if err != nil {
		t.Fatalf("failed to readlink cloned symlink: %v", err)
	}
	if target != "file.txt" {
		t.Errorf("expected symlink target 'file.txt', got %q", target)
	}

	dstBrokenLink := filepath.Join(dstDir, "broken_link.txt")
	brokenInfo, err := os.Lstat(dstBrokenLink)
	if err != nil {
		t.Fatalf("failed to lstat cloned broken symlink: %v", err)
	}
	if brokenInfo.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be symlink", dstBrokenLink)
	}
	brokenTarget, err := os.Readlink(dstBrokenLink)
	if err != nil {
		t.Fatalf("failed to readlink broken symlink: %v", err)
	}
	if brokenTarget != "non_existent.txt" {
		t.Errorf("expected symlink target 'non_existent.txt', got %q", brokenTarget)
	}
}

func TestCloneDirectory_SensitiveFilesSkipped(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dstDir: %v", err)
	}

	files := []string{
		"token.json",
		"auth_tokens",
		"credentials.env",
		"USER_CREDENTIAL",
		"session.db",
		"app.DB",
		"safe_config.json",
	}

	for _, name := range files {
		path := filepath.Join(srcDir, name)
		if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	if err := cloneDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("cloneDirectory failed: %v", err)
	}

	// Safe file should exist
	if _, err := os.Stat(filepath.Join(dstDir, "safe_config.json")); err != nil {
		t.Errorf("expected safe_config.json to exist: %v", err)
	}

	// Sensitive files should be skipped
	for _, name := range files {
		if name == "safe_config.json" {
			continue
		}
		dstPath := filepath.Join(dstDir, name)
		if _, err := os.Stat(dstPath); err == nil {
			t.Errorf("sensitive file %s was not skipped during clone", name)
		}
	}
}

func TestCloneDirectory_CopyFileError(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dstDir: %v", err)
	}

	unreadable := filepath.Join(srcDir, "unreadable.txt")
	if err := os.WriteFile(unreadable, []byte("forbidden"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	// Make unreadable
	if err := os.Chmod(unreadable, 0000); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(unreadable, 0644)
	})

	// If root running tests, chmod 0000 might still allow read, so skip if readable
	f, err := os.Open(unreadable)
	if err == nil {
		_ = f.Close()
		t.Skip("skipping unreadable file test: process has root/bypass permissions")
	}

	err = cloneDirectory(srcDir, dstDir)
	if err == nil {
		t.Fatalf("expected cloneDirectory to return error for unreadable file, got nil")
	}

	expectedPrefix := "copying file from " + unreadable
	if !strings.HasPrefix(err.Error(), expectedPrefix) {
		t.Errorf("expected error message to start with %q, got %q", expectedPrefix, err.Error())
	}
}

func TestCloneDirectory_DestinationConflictError(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dstDir: %v", err)
	}

	// Create regular file in src
	srcFile := filepath.Join(srcDir, "conflict")
	if err := os.WriteFile(srcFile, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write src file: %v", err)
	}

	// Create a non-empty directory in dst with the same name.
	// os.Remove(dstPath) cannot remove a non-empty directory,
	// so cloneOrCopyFile fails when attempting to copy file to dstPath.
	dstConflict := filepath.Join(dstDir, "conflict")
	if err := os.MkdirAll(filepath.Join(dstConflict, "nested"), 0755); err != nil {
		t.Fatalf("failed to create conflicting dir: %v", err)
	}

	err := cloneDirectory(srcDir, dstDir)
	if err == nil {
		t.Fatalf("expected cloneDirectory to fail due to conflicting destination directory, got nil")
	}

	expectedPrefix := "copying file from " + srcFile
	if !strings.HasPrefix(err.Error(), expectedPrefix) {
		t.Errorf("expected error message to start with %q, got %q", expectedPrefix, err.Error())
	}
}

func TestCloneDirectory_SymlinkCreationError(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dstDir: %v", err)
	}

	// Create symlink in src
	srcLink := filepath.Join(srcDir, "symlink_conflict")
	if err := os.Symlink("target.txt", srcLink); err != nil {
		t.Fatalf("failed to create src symlink: %v", err)
	}

	// Create a non-empty directory in dst with the same name.
	// os.Remove(dstPath) cannot remove a non-empty directory,
	// so os.Symlink fails when attempting to create symlink at dstPath.
	dstConflict := filepath.Join(dstDir, "symlink_conflict")
	if err := os.MkdirAll(filepath.Join(dstConflict, "nested"), 0755); err != nil {
		t.Fatalf("failed to create conflicting dir: %v", err)
	}

	err := cloneDirectory(srcDir, dstDir)
	if err == nil {
		t.Fatalf("expected cloneDirectory to fail due to conflicting destination directory for symlink, got nil")
	}

	expectedPrefix := "creating symlink " + dstConflict
	if !strings.HasPrefix(err.Error(), expectedPrefix) {
		t.Errorf("expected error message to start with %q, got %q", expectedPrefix, err.Error())
	}
}
