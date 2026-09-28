package merge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_RoundTripModesAndLifecycle(t *testing.T) {
	s := Store{Dir: filepath.Join(t.TempDir(), "profile-merge")}
	st, err := s.Load("work")
	if err != nil || st.Version != 1 {
		t.Fatalf("new state: %v %v", st, err)
	}
	st.Active["claude/mcpServers"] = map[string]string{"jev": "sha256:x"}
	st.AddedKey["claude/mcpServers"] = true
	if err := s.Save("work", st); err != nil {
		t.Fatal(err)
	}
	fi, _ := os.Stat(filepath.Join(s.Dir, "work.json"))
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", fi.Mode().Perm())
	}
	di, _ := os.Stat(s.Dir)
	if di.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode %v", di.Mode().Perm())
	}
	got, _ := s.Load("work")
	if got.Active["claude/mcpServers"]["jev"] != "sha256:x" || !got.AddedKey["claude/mcpServers"] {
		t.Fatalf("round trip lost data: %+v", got)
	}
	if s.SessionsLockPath("work", "claude") == s.SessionsLockPath("work", "codex") {
		t.Fatal("sessions lock must be per profile and agent")
	}
	if err := s.Rename("work", "job"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "job.json")); err != nil {
		t.Fatal("rename failed")
	}
	_ = os.WriteFile(s.SessionsLockPath("job", "claude"), nil, 0o600)
	if err := s.Remove("job"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "job.json")); !os.IsNotExist(err) {
		t.Fatal("state not removed")
	}
	if _, err := os.Stat(s.SessionsLockPath("job", "claude")); err != nil {
		t.Fatal("lock files must never be deleted")
	}
}
