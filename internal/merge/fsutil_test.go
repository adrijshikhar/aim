package merge

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAtomicWrite_ModeNeverLooser(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.json")
	if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AtomicWrite(p, []byte(`{"a":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	fi, _ := os.Stat(p)
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %v, want 0600", fi.Mode().Perm())
	}
}

func TestLocks_SharedThenUpgrade(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s.lock")
	a, err := LockShared(p)
	if err != nil {
		t.Fatal(err)
	}
	b, err := LockShared(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := TryUpgrade(p); ok {
		t.Fatal("upgrade must fail while two shared holders exist")
	}
	_ = b.Unlock()
	_ = a.Unlock()
	l, ok, err := TryUpgrade(p)
	if err != nil || !ok {
		t.Fatalf("upgrade after release: ok=%v err=%v", ok, err)
	}
	_ = l.Unlock()
}

func TestLockExclusive_Timeout(t *testing.T) {
	p := filepath.Join(t.TempDir(), "m.lock")
	l, _ := LockExclusive(p, time.Second)
	defer l.Unlock()
	if _, err := LockExclusive(p, 50*time.Millisecond); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestLock_UnlockTwiceIsSafe(t *testing.T) {
	l, err := LockExclusive(filepath.Join(t.TempDir(), "d.lock"), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Unlock(); err != nil {
		t.Fatal(err)
	}
	if err := l.Unlock(); err != nil {
		t.Fatalf("second Unlock must be a no-op, got %v", err)
	}
}
