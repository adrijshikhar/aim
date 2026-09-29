package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
)

// A running session holds its profile's sessions lock shared; remove, rename
// and move must refuse until it exits, and succeed afterwards.
func TestLifecycle_RefusedWhileSessionRuns(t *testing.T) {
	cases := map[string]func(pm *ProfileManager, cfg *config.Config) error{
		"delete": func(pm *ProfileManager, cfg *config.Config) error { return pm.DeleteProfile("work", cfg) },
		"remove-agent": func(pm *ProfileManager, cfg *config.Config) error {
			_, err := pm.RemoveAgent("work", "codex", cfg)
			return err
		},
		"rename": func(pm *ProfileManager, cfg *config.Config) error { return pm.RenameProfile("work", "job", cfg) },
		"move": func(pm *ProfileManager, cfg *config.Config) error {
			return pm.MoveAgent("codex", "work", "other", true, cfg, nil)
		},
		"move-into": func(pm *ProfileManager, cfg *config.Config) error {
			return pm.MoveAgent("codex", "other", "work", true, cfg, nil)
		},
	}
	for name, op := range cases {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			t.Setenv("AIM_HOME", base)
			pm := NewProfileManager(base)
			pm.MergeStore = merge.Store{Dir: filepath.Join(base, "profile-merge")}
			cfg := config.NewDefaultConfig()
			for _, p := range []string{"work", "other"} {
				dir, _ := pm.EnsureProfile(p)
				cfg.AddProfileAgent(p, "codex")
				_ = os.MkdirAll(filepath.Join(dir, ".codex"), 0o700)
				_ = os.WriteFile(filepath.Join(dir, ".codex", "auth.json"), []byte("tok"), 0o600)
			}
			// A second profile whose name starts with "work." must not count.
			_ = os.MkdirAll(pm.MergeStore.Dir, 0o700)
			other, _ := merge.LockShared(pm.MergeStore.SessionsLockPath("work.x", "claude"))
			defer other.Unlock()

			l, err := merge.LockShared(pm.MergeStore.SessionsLockPath("work", "claude"))
			if err != nil {
				t.Fatal(err)
			}
			err = op(pm, cfg)
			if err == nil || !strings.Contains(err.Error(), "profile work has a running claude session; exit it first") {
				l.Unlock()
				t.Fatalf("want refusal, got %v", err)
			}
			if _, err := os.Stat(pm.ProfileDir("work")); err != nil {
				t.Fatalf("refused op touched the profile: %v", err)
			}
			l.Unlock()
			if err := op(pm, cfg); err != nil {
				t.Fatalf("after exit: %v", err)
			}
		})
	}
}

// Clone only reads the source, so a running session does not block it.
func TestClone_AllowedWhileSessionRuns(t *testing.T) {
	base := t.TempDir()
	t.Setenv("AIM_HOME", base)
	pm := NewProfileManager(base)
	pm.MergeStore = merge.Store{Dir: filepath.Join(base, "profile-merge")}
	cfg := config.NewDefaultConfig()
	_, _ = pm.EnsureProfile("work")
	cfg.AddProfileAgent("work", "claude")
	l, err := merge.LockShared(pm.MergeStore.SessionsLockPath("work", "claude"))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Unlock()
	if err := pm.CloneProfile("work", "copy", "", cfg); err != nil {
		t.Fatal(err)
	}
}

// Merge state follows the profile: rename moves it, delete removes it, so a
// renamed profile keeps its migration and recovery records.
func TestLifecycle_MergeStateFollowsProfile(t *testing.T) {
	base := t.TempDir()
	t.Setenv("AIM_HOME", base)
	pm := NewProfileManager(base)
	pm.MergeStore = merge.Store{Dir: filepath.Join(base, "profile-merge")}
	cfg := config.NewDefaultConfig()
	_, _ = pm.EnsureProfile("work")
	cfg.AddProfileAgent("work", "claude")
	st, _ := pm.MergeStore.Load("work")
	st.Active["claude/mcpServers"] = map[string]string{"jev": "sha256:x"}
	if err := pm.MergeStore.Save("work", st); err != nil {
		t.Fatal(err)
	}
	if err := pm.RenameProfile("work", "job", cfg); err != nil {
		t.Fatal(err)
	}
	got, _ := pm.MergeStore.Load("job")
	if got.Active["claude/mcpServers"]["jev"] != "sha256:x" {
		t.Fatalf("rename lost merge state: %+v", got)
	}
	if err := pm.DeleteProfile("job", cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(pm.MergeStore.Dir, "job.json")); !os.IsNotExist(err) {
		t.Fatalf("delete left merge state: %v", err)
	}
}
