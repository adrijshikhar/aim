//go:build !windows

package profile

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestEnsureDotfiles_SkipsHostFifo(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(host, ".agent.fifo"), 0600); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	mustWriteFile(t, filepath.Join(host, ".gitconfig"), "[user]")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(prof, ".agent.fifo")); err == nil {
		t.Errorf("a host fifo must not be bridged")
	}
	if !isSymlink(filepath.Join(prof, ".gitconfig")) {
		t.Errorf("expected .gitconfig to be linked")
	}
}
