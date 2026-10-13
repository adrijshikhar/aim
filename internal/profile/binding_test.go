package profile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/profile"
)

func TestResolveDirProfile_AscendingSearch(t *testing.T) {
	tempRoot := t.TempDir()
	repoRoot := filepath.Join(tempRoot, "my-repo")
	subDir := filepath.Join(repoRoot, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// 1. No binding initially
	res, err := profile.ResolveDirProfile(subDir)
	if err != nil {
		t.Fatalf("expected nil error on empty search, got: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil resolution, got: %+v", res)
	}

	// 2. Bind repoRoot to "work"
	bindPath, err := profile.BindDirectory(repoRoot, "work")
	if err != nil {
		t.Fatalf("BindDirectory failed: %v", err)
	}
	if filepath.Base(bindPath) != ".aim-profile" {
		t.Errorf("expected .aim-profile filename, got %s", bindPath)
	}

	// 3. Resolve from deep subDir
	res, err = profile.ResolveDirProfile(subDir)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if res == nil || res.Profile != "work" {
		t.Fatalf("expected profile 'work', got: %+v", res)
	}
	if res.SourceFile != bindPath {
		t.Errorf("expected source file %s, got %s", bindPath, res.SourceFile)
	}

	// 4. Unbind
	removedPath, err := profile.UnbindDirectory(repoRoot)
	if err != nil {
		t.Fatalf("UnbindDirectory failed: %v", err)
	}
	if removedPath != bindPath {
		t.Errorf("expected removed path %s, got %s", bindPath, removedPath)
	}

	// 5. Verify resolution is now nil
	resAfter, _ := profile.ResolveDirProfile(subDir)
	if resAfter != nil {
		t.Errorf("expected nil resolution after unbind")
	}
}

func TestResolveDirProfile_WithCommentsAndWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, ".aim-profile")
	_ = os.WriteFile(filePath, []byte("  client-acme   # target workspace profile\n"), 0644)

	res, err := profile.ResolveDirProfile(tempDir)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if res == nil || res.Profile != "client-acme" {
		t.Fatalf("expected profile 'client-acme', got: %+v", res)
	}
}

func TestResolveDirProfile_GitBoundary(t *testing.T) {
	tempRoot := t.TempDir()
	parentDir := filepath.Join(tempRoot, "parent")
	repoDir := filepath.Join(parentDir, "child-repo")
	gitDir := filepath.Join(repoDir, ".git")
	subDir := filepath.Join(repoDir, "cmd")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Bind parentDir above .git boundary
	_, err := profile.BindDirectory(parentDir, "parent-profile")
	if err != nil {
		t.Fatalf("BindDirectory failed: %v", err)
	}

	// Resolving from inside child-repo should stop at .git boundary and return nil
	res, err := profile.ResolveDirProfile(subDir)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil resolution across .git boundary, got: %+v", res)
	}
}
