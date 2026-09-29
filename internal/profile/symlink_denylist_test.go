package profile

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// Every test here passes an explicit t.TempDir() host home to EnsureDotfiles,
// GetBridgedPaths or BridgeStatus, so none of them can reach the real home.

func mustMkdirAll(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0755); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, p, s string) {
	t.Helper()
	mustMkdirAll(t, filepath.Dir(p))
	if err := os.WriteFile(p, []byte(s), 0644); err != nil {
		t.Fatal(err)
	}
}

func isSymlink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

// assertSharedWithHost checks that rel inside the profile resolves to the
// host's rel, whether it is linked itself or sits under a linked parent.
func assertSharedWithHost(t *testing.T, profileDir, hostHome, rel string) {
	t.Helper()
	got, err := filepath.EvalSymlinks(filepath.Join(profileDir, rel))
	if err != nil {
		t.Errorf("%s: not reachable in the profile: %v", rel, err)
		return
	}
	want, err := filepath.EvalSymlinks(filepath.Join(hostHome, rel))
	if err != nil {
		t.Fatalf("%s: host path missing: %v", rel, err)
	}
	if got != want {
		t.Errorf("%s: resolves to %s, want the host's %s", rel, got, want)
	}
}

func TestEnsureDotfiles_BridgesEveryHostDotfile(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()

	mustWriteFile(t, filepath.Join(host, ".gitconfig"), "[user]")
	mustWriteFile(t, filepath.Join(host, ".newtool", "cfg"), "x")
	mustWriteFile(t, filepath.Join(host, ".mcp-auth", "token"), "x")
	mustWriteFile(t, filepath.Join(host, ".config", "gh", "hosts.yml"), "x")
	mustWriteFile(t, filepath.Join(host, "Documents", "notes"), "x")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}

	for _, rel := range []string{".gitconfig", ".newtool", ".mcp-auth", ".config"} {
		if !isSymlink(filepath.Join(prof, rel)) {
			t.Errorf("expected %s to be linked to the host", rel)
		}
		assertSharedWithHost(t, prof, host, rel)
	}
	assertSharedWithHost(t, prof, host, filepath.Join(".config", "gh", "hosts.yml"))
	if _, err := os.Lstat(filepath.Join(prof, "Documents")); err == nil {
		t.Errorf("a non-dot host entry must not be bridged")
	}
}

func TestEnsureDotfiles_NeverLinksIsolatedAgentState(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()

	for _, dir := range []string{".aim", ".claude", ".codex", ".gemini"} {
		mustWriteFile(t, filepath.Join(host, dir, "state"), "host")
	}
	for _, f := range []string{".claude.json", ".claude.json.backup", ".claude.json.tmp.123"} {
		mustWriteFile(t, filepath.Join(host, f), "{}")
	}

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}

	for _, name := range []string{".aim", ".claude", ".codex", ".gemini", ".claude.json", ".claude.json.backup", ".claude.json.tmp.123"} {
		if _, err := os.Lstat(filepath.Join(prof, name)); err == nil {
			t.Errorf("ISOLATION VIOLATION: %s was bridged into the profile", name)
		}
	}
	for _, name := range []string{".aim", ".claude", ".codex", ".gemini", ".claude.json", ".claude.json.backup"} {
		if slices.Contains(GetBridgedPaths(host), name) {
			t.Errorf("GetBridgedPaths must not list %s", name)
		}
	}
}

func TestEnsureDotfiles_SkipsHostOnlyNoise(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()

	noise := []string{
		".Trash", ".DS_Store", ".CFUserTextEncoding", ".localized",
		".zsh_history", ".bash_history", ".python_history", ".node_repl_history",
		".zsh_sessions", ".bash_sessions", ".zcompdump", ".zcompdump-host-5.9",
	}
	for _, n := range noise {
		mustWriteFile(t, filepath.Join(host, n), "x")
	}

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	for _, n := range noise {
		if _, err := os.Lstat(filepath.Join(prof, n)); err == nil {
			t.Errorf("host-only %s must not be bridged", n)
		}
	}
}

func TestEnsureDotfiles_LinksHostDotdirCreatedLater(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	mustWriteFile(t, filepath.Join(host, ".gitconfig"), "[user]")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(prof, ".newtool")); err == nil {
		t.Fatalf(".newtool must not exist before the host has it")
	}

	mustWriteFile(t, filepath.Join(host, ".newtool", "cfg"), "x")
	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if !isSymlink(filepath.Join(prof, ".newtool")) {
		t.Errorf("expected a host dotdir created after the profile to be linked on the next run")
	}
}

func TestEnsureDotfiles_KeepsProfileOverrideDir(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	mustWriteFile(t, filepath.Join(host, ".newtool", "cfg"), "host")
	mustWriteFile(t, filepath.Join(prof, ".newtool", "cfg"), "profile")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if isSymlink(filepath.Join(prof, ".newtool")) {
		t.Fatalf("a non-empty profile dir is an override and must be kept")
	}
	data, err := os.ReadFile(filepath.Join(prof, ".newtool", "cfg"))
	if err != nil || string(data) != "profile" {
		t.Errorf("override content changed: %q, %v", data, err)
	}
}

// A nested default such as .config/ccstatusline must never be acted on through
// a profile .config that links to the host: the stub check would see the host's
// own empty dir, remove it and link it to itself, rewriting the real home.
func TestEnsureDotfiles_NeverWritesThroughALinkedParent(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	hostStub := filepath.Join(host, ".config", "ccstatusline")
	mustMkdirAll(t, hostStub)
	mustWriteFile(t, filepath.Join(host, ".config", "gh", "hosts.yml"), "x")
	if err := os.Symlink(filepath.Join(host, ".config"), filepath.Join(prof, ".config")); err != nil {
		t.Fatal(err)
	}

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}

	fi, err := os.Lstat(hostStub)
	if err != nil {
		t.Fatalf("host %s was removed: %v", hostStub, err)
	}
	if !fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("host %s was replaced (mode %v)", hostStub, fi.Mode())
	}
	if isSymlink(filepath.Join(host, ".config", "gh")) {
		t.Errorf("host .config/gh was replaced with a link")
	}
}

// A profile that overrides .config with its own real dir still gets the nested
// fallbacks linked inside it.
func TestEnsureDotfiles_NestedFallbackUnderOverriddenParent(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	mustWriteFile(t, filepath.Join(host, ".config", "gh", "hosts.yml"), "x")
	mustWriteFile(t, filepath.Join(host, ".config", "other", "cfg"), "x")
	mustWriteFile(t, filepath.Join(prof, ".config", "mine", "cfg"), "profile")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if isSymlink(filepath.Join(prof, ".config")) {
		t.Fatalf("the overriding .config must be kept")
	}
	if !isSymlink(filepath.Join(prof, ".config", "gh")) {
		t.Errorf("expected the nested .config/gh fallback to be linked")
	}
	if _, err := os.Lstat(filepath.Join(prof, ".config", "other")); err == nil {
		t.Errorf("only nested defaults are linked inside an override")
	}
}

// Profiles made under the old allow-list hold a real .config containing only
// links into the host's .config. That scaffold carries nothing of its own, so
// it is replaced with one link and the profile picks up every host config dir.
func TestEnsureDotfiles_ReplacesLegacyLinkScaffold(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	mustWriteFile(t, filepath.Join(host, ".config", "gh", "hosts.yml"), "x")
	mustWriteFile(t, filepath.Join(host, ".local", "share", "fish", "hist"), "x")
	mustWriteFile(t, filepath.Join(host, ".config", "newtool", "cfg"), "x")
	mustMkdirAll(t, filepath.Join(prof, ".config"))
	mustMkdirAll(t, filepath.Join(prof, ".local", "share"))
	if err := os.Symlink(filepath.Join(host, ".config", "gh"), filepath.Join(prof, ".config", "gh")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(host, ".local", "share", "fish"), filepath.Join(prof, ".local", "share", "fish")); err != nil {
		t.Fatal(err)
	}
	// A scaffold that also holds a real file is an override and stays.
	mustWriteFile(t, filepath.Join(host, ".cargo", "config.toml"), "x")
	mustWriteFile(t, filepath.Join(prof, ".cargo", "own.toml"), "profile")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	for _, rel := range []string{".config", ".local"} {
		if !isSymlink(filepath.Join(prof, rel)) {
			t.Errorf("expected legacy scaffold %s to be replaced with a link", rel)
		}
	}
	assertSharedWithHost(t, prof, host, filepath.Join(".config", "newtool", "cfg"))
	if _, err := os.Stat(filepath.Join(host, ".config", "gh", "hosts.yml")); err != nil {
		t.Errorf("replacing the scaffold touched the host: %v", err)
	}
	if isSymlink(filepath.Join(prof, ".cargo")) {
		t.Errorf("a .cargo holding a real file must be kept")
	}
	if !isSymlink(filepath.Join(prof, ".cargo", "config.toml")) {
		t.Errorf("expected the nested .cargo/config.toml fallback inside the override")
	}
}

// A profile living inside a host dotdir (a custom AIM_HOME under ~/.local, say)
// must not get a link to its own ancestor.
func TestEnsureDotfiles_SkipsHostEntryContainingTheProfile(t *testing.T) {
	host := t.TempDir()
	prof := filepath.Join(host, ".local", "aim", "profiles", "p")
	mustMkdirAll(t, prof)
	mustWriteFile(t, filepath.Join(host, ".gitconfig"), "[user]")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(prof, ".local")); err == nil {
		t.Errorf("the host dir containing the profile must not be linked into it")
	}
	if !isSymlink(filepath.Join(prof, ".gitconfig")) {
		t.Errorf("expected .gitconfig to be linked")
	}
}

func TestGetBridgedPaths_ScansHostAndKeepsExtras(t *testing.T) {
	host := t.TempDir()
	mustWriteFile(t, filepath.Join(host, ".gitconfig"), "x")
	mustWriteFile(t, filepath.Join(host, ".newtool", "cfg"), "x")
	mustWriteFile(t, filepath.Join(host, ".zsh_history"), "x")
	mustWriteFile(t, filepath.Join(host, ".codex", "auth.json"), "x")

	extra := filepath.Join("Library", "Application Support", "Tool")
	got := GetBridgedPaths(host, extra, ".newtool")

	for _, want := range []string{".gitconfig", ".newtool", filepath.Join(".claude", "plugins"), extra} {
		if !slices.Contains(got, want) {
			t.Errorf("expected %s in %v", want, got)
		}
	}
	for _, deny := range []string{".zsh_history", ".codex"} {
		if slices.Contains(got, deny) {
			t.Errorf("did not expect %s in %v", deny, got)
		}
	}
	seen := map[string]bool{}
	for _, p := range got {
		if seen[p] {
			t.Errorf("duplicate entry %s", p)
		}
		seen[p] = true
	}
	// Top-level entries come first so a parent is linked before its nested defaults.
	if slices.Index(got, ".newtool") > slices.Index(got, filepath.Join(".claude", "plugins")) {
		t.Errorf("expected scanned top-level entries before nested defaults: %v", got)
	}
}

func TestBridgeStatus_ListsScannedCopies(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	mustWriteFile(t, filepath.Join(host, ".newtool", "cfg"), "host")
	mustWriteFile(t, filepath.Join(prof, ".newtool", "cfg"), "stale")
	mustWriteFile(t, filepath.Join(host, ".codex", "auth.json"), "host")
	mustWriteFile(t, filepath.Join(prof, ".codex", "auth.json"), "profile")
	mustWriteFile(t, filepath.Join(host, ".config", "gh", "hosts.yml"), "x")
	if err := os.Symlink(filepath.Join(host, ".config"), filepath.Join(prof, ".config")); err != nil {
		t.Fatal(err)
	}

	got := map[string]BridgeKind{}
	for _, s := range BridgeStatus(host, prof, GetBridgedPaths(host)) {
		got[s.Path] = s.Kind
	}
	if k, ok := got[".newtool"]; !ok || k != BridgeCopy {
		t.Errorf(".newtool: got %v (present %v), want BridgeCopy", k, ok)
	}
	if _, ok := got[".codex"]; ok {
		t.Errorf(".codex is per-profile state and must not be reported")
	}
	if k := got[".config"]; k != BridgeLinked {
		t.Errorf(".config: got %v, want BridgeLinked", k)
	}
	if _, ok := got[filepath.Join(".config", "gh")]; ok {
		t.Errorf("a nested path under a linked parent must not be reported separately")
	}
}

func TestEnsureDotfiles_SharesGoPathWhenPresent(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(prof, "go")); err == nil {
		t.Fatalf("go must not be created when the host has none")
	}

	mustWriteFile(t, filepath.Join(host, "go", "pkg", "mod", "cache", "x"), "x")
	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if !isSymlink(filepath.Join(prof, "go")) {
		t.Errorf("expected the host's go dir to be linked")
	}
}

func TestEnsureDotfiles_LibraryCachesAndKeychainsOnly(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	mustWriteFile(t, filepath.Join(host, "Library", "Caches", "go-build", "x"), "x")
	mustWriteFile(t, filepath.Join(host, "Library", "Keychains", "login.keychain-db"), "x")
	mustWriteFile(t, filepath.Join(host, "Library", "Application Support", "App", "state"), "x")
	mustWriteFile(t, filepath.Join(host, "Library", "Preferences", "app.plist"), "x")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}

	lib := filepath.Join(prof, "Library")
	if isSymlink(lib) {
		t.Fatalf("Library itself must stay a real dir in the profile")
	}
	for _, rel := range []string{"Application Support", "Preferences"} {
		if _, err := os.Lstat(filepath.Join(lib, rel)); err == nil {
			t.Errorf("Library/%s holds per-app state and must not be bridged", rel)
		}
	}
	for _, rel := range []string{"Caches", "Keychains"} {
		linked := isSymlink(filepath.Join(lib, rel))
		if runtime.GOOS == "darwin" && !linked {
			t.Errorf("expected Library/%s to be linked on darwin", rel)
		}
		if runtime.GOOS != "darwin" && linked {
			t.Errorf("Library/%s must not be linked off darwin", rel)
		}
	}
}

func TestEnsureDotfiles_LibraryCachesOverrideKeptEmptyReplaced(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Library/Caches is bridged on darwin only")
	}
	host := t.TempDir()
	kept := t.TempDir()
	empty := t.TempDir()
	mustWriteFile(t, filepath.Join(host, "Library", "Caches", "pip", "x"), "host")
	mustWriteFile(t, filepath.Join(kept, "Library", "Caches", "ms-playwright", "x"), "profile")
	mustMkdirAll(t, filepath.Join(empty, "Library", "Caches"))

	for _, prof := range []string{kept, empty} {
		if err := EnsureDotfiles(host, prof); err != nil {
			t.Fatalf("EnsureDotfiles: %v", err)
		}
	}
	if isSymlink(filepath.Join(kept, "Library", "Caches")) {
		t.Errorf("a non-empty profile Library/Caches is an override and must be kept")
	}
	if _, err := os.Stat(filepath.Join(kept, "Library", "Caches", "ms-playwright", "x")); err != nil {
		t.Errorf("override content lost: %v", err)
	}
	if !isSymlink(filepath.Join(empty, "Library", "Caches")) {
		t.Errorf("an empty profile Library/Caches must be replaced with a link")
	}
}

// On the XDG layout profiles live at ~/.local/share/aim/profiles/<p>, so .local
// contains the profile and cannot be linked. Its children are bridged instead,
// descending only through the entries that still contain the profile, and the
// aim data dir itself (config, other profiles) is never linked.
func TestEnsureDotfiles_DescendsIntoHostDirContainingTheProfile(t *testing.T) {
	host := t.TempDir()
	prof := filepath.Join(host, ".local", "share", "aim", "profiles", "p")
	mustMkdirAll(t, prof)
	mustWriteFile(t, filepath.Join(host, ".local", "bin", "tool"), "#!/bin/sh")
	mustWriteFile(t, filepath.Join(host, ".local", "share", "omf", "init.fish"), "x")
	mustWriteFile(t, filepath.Join(host, ".local", "share", "claude", "versions", "1"), "x")
	mustWriteFile(t, filepath.Join(host, ".local", "state", "gh", "state.yml"), "x")
	mustWriteFile(t, filepath.Join(host, ".local", "share", "aim", "config.json"), "{}")
	mustMkdirAll(t, filepath.Join(host, ".local", "share", "aim", "profiles", "q", ".claude"))

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}

	for _, rel := range []string{
		filepath.Join(".local", "bin"),
		filepath.Join(".local", "state"),
		filepath.Join(".local", "share", "omf"),
		filepath.Join(".local", "share", "claude"),
	} {
		if !isSymlink(filepath.Join(prof, rel)) {
			t.Errorf("expected %s to be linked to the host", rel)
		}
		assertSharedWithHost(t, prof, host, rel)
	}
	for _, rel := range []string{".local", filepath.Join(".local", "share")} {
		if isSymlink(filepath.Join(prof, rel)) {
			t.Errorf("%s contains the profile and must not be linked", rel)
		}
	}
	if _, err := os.Lstat(filepath.Join(prof, ".local", "share", "aim")); err == nil {
		t.Errorf("the aim data dir must not be bridged into the profile")
	}

	got := map[string]BridgeKind{}
	for _, s := range BridgeStatus(host, prof, GetBridgedPaths(host)) {
		if _, dup := got[s.Path]; dup {
			t.Errorf("BridgeStatus reported %s twice", s.Path)
		}
		got[s.Path] = s.Kind
	}
	for _, rel := range []string{
		filepath.Join(".local", "bin"),
		filepath.Join(".local", "share", "omf"),
		filepath.Join(".local", "share", "claude"),
	} {
		if k, ok := got[rel]; !ok || k != BridgeLinked {
			t.Errorf("BridgeStatus %s: got %v (present %v), want BridgeLinked", rel, k, ok)
		}
	}
	for p := range got {
		if p == ".local" || strings.HasPrefix(p, filepath.Join(".local", "share", "aim")) {
			t.Errorf("BridgeStatus must not report %s", p)
		}
	}

	// No link inside the profile may lead back to the profile or an ancestor.
	_ = filepath.WalkDir(prof, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.Type()&os.ModeSymlink == 0 {
			return nil
		}
		target, _ := filepath.EvalSymlinks(p)
		if rel, err := filepath.Rel(target, prof); err == nil && filepath.IsLocal(rel) {
			t.Errorf("%s -> %s leads back to the profile", p, target)
		}
		return nil
	})
}

// An empty subdir is part of a scaffold only where the host has that dir too;
// otherwise it is profile structure of its own and the dir is kept.
func TestEnsureDotfiles_KeepsDirWithEmptySubdirTheHostLacks(t *testing.T) {
	host := t.TempDir()
	prof := t.TempDir()
	mustWriteFile(t, filepath.Join(host, ".config", "gh", "hosts.yml"), "x")
	mustMkdirAll(t, filepath.Join(prof, ".config", "mine"))
	if err := os.Symlink(filepath.Join(host, ".config", "gh"), filepath.Join(prof, ".config", "gh")); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(host, ".config2", "cfg"), "x")
	mustMkdirAll(t, filepath.Join(prof, ".config2", "a", "b"))

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	for _, rel := range []string{filepath.Join(".config", "mine"), filepath.Join(".config2", "a", "b")} {
		fi, err := os.Lstat(filepath.Join(prof, rel))
		if err != nil || !fi.IsDir() {
			t.Errorf("%s must be kept as the profile's own dir (err %v)", rel, err)
		}
	}
	for _, rel := range []string{".config", ".config2"} {
		if isSymlink(filepath.Join(prof, rel)) {
			t.Errorf("%s holds a dir the host lacks and must not be replaced", rel)
		}
	}
}

// The deny-list is checked against where a host entry resolves, not only its
// name: a host link into aim or agent state is never bridged.
func TestEnsureDotfiles_HostLinkIntoIsolatedStateNotLinked(t *testing.T) {
	host := t.TempDir()
	prof := filepath.Join(host, ".aim", "profiles", "office")
	mustMkdirAll(t, prof)
	mustWriteFile(t, filepath.Join(host, ".aim", "config.json"), "{}")
	mustWriteFile(t, filepath.Join(host, ".aim", "profiles", "work", ".claude", ".credentials.json"), "secret")
	mustWriteFile(t, filepath.Join(host, "dotfiles", "tool", "cfg"), "x")
	for link, target := range map[string]string{
		".claude-work": filepath.Join(host, ".aim", "profiles", "work", ".claude"),
		".aimlink":     filepath.Join(host, ".aim"),
		".toollink":    filepath.Join(host, "dotfiles", "tool"),
	} {
		if err := os.Symlink(target, filepath.Join(host, link)); err != nil {
			t.Fatal(err)
		}
	}
	other := t.TempDir()

	for _, p := range []string{prof, other} {
		if err := EnsureDotfiles(host, p); err != nil {
			t.Fatalf("EnsureDotfiles(%s): %v", p, err)
		}
		for _, rel := range []string{".claude-work", ".aimlink"} {
			if _, err := os.Lstat(filepath.Join(p, rel)); err == nil {
				t.Errorf("%s: %s resolves into aim/agent state and must not be linked", p, rel)
			}
		}
		if !isSymlink(filepath.Join(p, ".toollink")) {
			t.Errorf("%s: a host link to ordinary config must still be linked", p)
		}
		for _, s := range BridgeStatus(host, p, GetBridgedPaths(host)) {
			if s.Path == ".claude-work" || s.Path == ".aimlink" {
				t.Errorf("%s: BridgeStatus must not report %s", p, s.Path)
			}
		}
	}
}

// A profiles root reached through a symlink into a host dotdir is still inside
// that dotdir: containment compares resolved paths, so no cycle is created.
func TestEnsureDotfiles_SymlinkedProfilesRootNoCycle(t *testing.T) {
	host := t.TempDir()
	mustMkdirAll(t, filepath.Join(host, ".local", "share", "aim", "profiles", "p"))
	mustWriteFile(t, filepath.Join(host, ".local", "bin", "tool"), "#!/bin/sh")
	if err := os.Symlink(filepath.Join(host, ".local", "share", "aim"), filepath.Join(host, "aimhome")); err != nil {
		t.Fatal(err)
	}
	prof := filepath.Join(host, "aimhome", "profiles", "p")

	if err := EnsureDotfiles(host, prof); err != nil {
		t.Fatalf("EnsureDotfiles: %v", err)
	}
	if isSymlink(filepath.Join(prof, ".local")) {
		t.Fatalf(".local contains the profile through aimhome and must not be linked")
	}
	if !isSymlink(filepath.Join(prof, ".local", "bin")) {
		t.Errorf("expected .local/bin to be linked")
	}
	if _, err := os.Lstat(filepath.Join(prof, ".local", "share", "aim")); err == nil {
		t.Errorf("the aim data dir must not be bridged into the profile")
	}

	// A profile dir that is the real home itself (via a link) is left alone.
	mustMkdirAll(t, filepath.Join(host, ".emptydir"))
	selfLink := filepath.Join(t.TempDir(), "self")
	if err := os.Symlink(host, selfLink); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDotfiles(host, selfLink); err != nil {
		t.Fatalf("EnsureDotfiles(self): %v", err)
	}
	if fi, err := os.Lstat(filepath.Join(host, ".emptydir")); err != nil || !fi.IsDir() {
		t.Errorf("a profile that is the real home must not have its dirs replaced (err %v)", err)
	}
}
