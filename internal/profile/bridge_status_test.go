package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBridgeStatus(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	elsewhere := t.TempDir()

	mustMkdir := func(p string) {
		t.Helper()
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite := func(p, s string) {
		t.Helper()
		mustMkdir(filepath.Dir(p))
		if err := os.WriteFile(p, []byte(s), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mustLink := func(target, link string) {
		t.Helper()
		mustMkdir(filepath.Dir(link))
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
	}

	// Every case has a host path; only the profile side differs.
	for _, p := range []string{".linked", ".foreign", ".dangling", ".copy", ".stub", ".absent", ".filecopy"} {
		mustWrite(filepath.Join(host, p, "creds.yml"), "host")
	}
	mustWrite(filepath.Join(host, ".hostfile"), "host")

	mustLink(filepath.Join(host, ".linked"), filepath.Join(prof, ".linked"))
	mustLink(filepath.Join(host, ".hostfile"), filepath.Join(prof, ".hostfile"))
	mustLink(elsewhere, filepath.Join(prof, ".foreign"))
	mustLink(filepath.Join(elsewhere, "gone"), filepath.Join(prof, ".dangling"))
	mustWrite(filepath.Join(prof, ".copy", "creds.yml"), "stale")
	mustMkdir(filepath.Join(prof, ".stub"))
	mustWrite(filepath.Join(prof, ".filecopy"), "not a dir")

	paths := []string{
		".linked", ".hostfile", ".foreign", ".dangling", ".copy", ".stub", ".absent", ".filecopy",
		".no-host",   // no host path: skipped
		".aim",       // denied by isAllowedBridgedPath: skipped
		"../outside", // escapes the profile: skipped
	}
	got := BridgeStatus(host, prof, paths)

	want := map[string]BridgeKind{
		".linked":   BridgeLinked,
		".hostfile": BridgeLinked,
		".foreign":  BridgeForeignLink,
		".dangling": BridgeForeignLink,
		".copy":     BridgeCopy,
		".stub":     BridgePending,
		".absent":   BridgePending,
		".filecopy": BridgeCopy,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d states, want %d: %+v", len(got), len(want), got)
	}
	for _, s := range got {
		k, ok := want[s.Path]
		if !ok {
			t.Errorf("unexpected state for %q", s.Path)
			continue
		}
		if s.Kind != k {
			t.Errorf("%s: kind = %v, want %v", s.Path, s.Kind, k)
		}
		if s.ProfilePath != filepath.Join(prof, s.Path) {
			t.Errorf("%s: ProfilePath = %q", s.Path, s.ProfilePath)
		}
	}
	for _, s := range got {
		switch s.Path {
		case ".foreign":
			if s.Target != elsewhere {
				t.Errorf(".foreign: Target = %q, want %q", s.Target, elsewhere)
			}
		case ".dangling":
			if s.Target != filepath.Join(elsewhere, "gone") {
				t.Errorf(".dangling: Target = %q", s.Target)
			}
		}
	}
}
