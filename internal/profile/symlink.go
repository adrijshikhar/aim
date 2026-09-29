package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aim-cli/aim/internal/logger"
)

// Every top-level dot entry of the host home is bridged into a profile unless
// it is denied here. Two kinds are denied:
//
//   - aim and agent state that each profile keeps for itself (credentials,
//     settings, session history). Linking any of these would merge accounts.
//   - host-only noise: Finder and shell bookkeeping that is either meaningless
//     inside a profile or written on every prompt.
var isolatedDotfiles = map[string]bool{
	".aim":         true,
	".claude":      true,
	".claude.json": true,
	".codex":       true,
	".gemini":      true,
}

var hostOnlyDotfiles = map[string]bool{
	".Trash":              true,
	".DS_Store":           true,
	".CFUserTextEncoding": true,
	".localized":          true,
	".zsh_sessions":       true,
	".bash_sessions":      true,
}

// isDeniedHostDotfile reports whether a top-level host dot entry stays out of
// profiles. .claude.json.* covers Claude's backup and atomic-write temp files.
func isDeniedHostDotfile(name string) bool {
	return isolatedDotfiles[name] ||
		strings.HasPrefix(name, ".claude.json.") ||
		hostOnlyDotfiles[name] ||
		strings.HasSuffix(name, "_history") ||
		strings.HasPrefix(name, ".zcompdump")
}

// defaultNonDotPaths are host paths outside the dot scan that are shared too:
// GOPATH's default module and build cache, and on macOS the keychains (so gh
// and git-credential-osxkeychain work) and the per-user cache dir, which would
// otherwise be duplicated per profile. Library itself stays a real dir in the
// profile; Application Support and Preferences hold per-app state and are not
// shared.
var defaultNonDotPaths = func() []string {
	paths := []string{"go"}
	if runtime.GOOS == "darwin" {
		paths = append(paths, filepath.Join("Library", "Keychains"), filepath.Join("Library", "Caches"))
	}
	return paths
}()

// nestedBridgedPaths are bridged below a top-level entry. The Claude extension
// dirs live under the per-profile .claude. The rest are fallbacks for a profile
// that overrides the parent (.config, .cargo, .local, .pip) with a real dir of
// its own; whenever that parent links to the host they are skipped, since they
// already resolve into it. A parent that contains the profile (.local on the
// XDG layout) is not an override: its children are bridged one by one instead
// (see forEachBridgeCandidate).
var nestedBridgedPaths = []string{
	filepath.Join(".claude", "plugins"),
	filepath.Join(".claude", "skills"),
	filepath.Join(".claude", "rules"),
	filepath.Join(".claude", "commands"),
	filepath.Join(".claude", "hooks"),

	filepath.Join(".config", "git"),
	filepath.Join(".config", "gh"),
	filepath.Join(".config", "glab"),
	filepath.Join(".config", "fish"),
	filepath.Join(".config", "gcloud"),
	filepath.Join(".config", "cxstatusline"),
	filepath.Join(".config", "ccstatusline"),
	filepath.Join(".local", "share", "omf"),
	filepath.Join(".local", "share", "fish"),
	filepath.Join(".pip", "pip.conf"),
	filepath.Join(".cargo", "config.toml"),
	filepath.Join(".cargo", "credentials.toml"),
	filepath.Join(".cargo", "env.fish"),
	filepath.Join(".cargo", "env"),
}

// GetBridgedPaths returns the paths to bridge from realHome, relative to it:
// every top-level dot entry not denied by isDeniedHostDotfile, then the default
// non-dot and nested paths, then extraPaths (custom_bridged_paths), without
// duplicates. Top-level entries come first so a parent is linked before any
// nested default under it is considered. The set is recomputed on every call,
// so a host dotdir created after a profile is picked up on the next launch.
func GetBridgedPaths(realHome string, extraPaths ...string) []string {
	var paths []string
	if entries, err := os.ReadDir(realHome); err == nil {
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, ".") && !isDeniedHostDotfile(name) {
				paths = append(paths, name)
			}
		}
	}
	paths = append(paths, defaultNonDotPaths...)
	paths = append(paths, nestedBridgedPaths...)
	paths = append(paths, extraPaths...)

	seen := make(map[string]bool, len(paths))
	out := paths[:0]
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// isAllowedBridgedPath validates that the path does not escape the profile
// directory and does not bridge internal AIM or agent credential stores.
func isAllowedBridgedPath(clean string) bool {
	if !filepath.IsLocal(clean) {
		return false
	}
	// Deny AIM internal management state
	if clean == ".aim" || strings.HasPrefix(clean, ".aim"+string(filepath.Separator)) {
		return false
	}
	// Deny agent token storage directories that AIM isolates per-profile
	if clean == ".gemini" || strings.HasPrefix(clean, ".gemini"+string(filepath.Separator)) {
		return false
	}
	if clean == ".claude" || clean == ".claude.json" || strings.HasPrefix(clean, ".claude.json.") {
		return false
	}
	if strings.HasPrefix(clean, ".claude"+string(filepath.Separator)) {
		// Allow specific non-credential Claude extensions to be bridged
		rel := strings.TrimPrefix(clean, ".claude"+string(filepath.Separator))
		top := strings.Split(rel, string(filepath.Separator))[0]
		switch top {
		case "plugins", "skills", "rules", "commands", "hooks":
			return true
		default:
			return false
		}
	}
	if clean == ".codex" || strings.HasPrefix(clean, ".codex"+string(filepath.Separator)) {
		return false
	}
	return true
}

// bridgeCandidate resolves one bridged path for profileDir and reports whether
// it may be acted on: it must be allowed, exist on the host as a regular file,
// dir or symlink (not a socket, fifo or device), not contain the profile, and
// not sit under a profile parent that is a link or a file. That last rule keeps
// EnsureDotfiles from creating, replacing or removing anything inside a dir
// that is itself a link to the host, which would rewrite the real home.
//
// A host dir that contains the profile is never acted on, but descend reports
// whether its children should be tried instead: true while it strictly
// contains the aim data dir (the parent of the profiles root), so .local and
// .local/share are descended on the XDG layout and the data dir itself, with
// its config and the other profiles, is not.
func bridgeCandidate(realHome, profileDir, name string) (cleanName, src, dest string, ok, descend bool) {
	cleanName = filepath.Clean(filepath.FromSlash(name))
	if !isAllowedBridgedPath(cleanName) {
		logger.Debug("[symlink] Skipping disallowed path: %s", name)
		return "", "", "", false, false
	}
	src = filepath.Join(realHome, cleanName)
	fi, err := os.Lstat(src)
	if err != nil {
		return "", "", "", false, false
	}
	if fi.Mode()&(os.ModeSocket|os.ModeNamedPipe|os.ModeDevice|os.ModeCharDevice|os.ModeIrregular) != 0 {
		logger.Debug("[symlink] Skipping special host file: %s", src)
		return "", "", "", false, false
	}
	if isWithin(profileDir, src) {
		logger.Debug("[symlink] Skipping %s: it contains the profile", src)
		dataDir := filepath.Dir(filepath.Dir(profileDir))
		descend = fi.IsDir() && isWithin(dataDir, src) && filepath.Clean(dataDir) != filepath.Clean(src)
		return cleanName, src, "", false, descend
	}
	if underLinkOrFile(profileDir, cleanName) {
		return "", "", "", false, false
	}
	return cleanName, src, filepath.Join(profileDir, cleanName), true, false
}

// isWithin reports whether path is dir or lies under it.
func isWithin(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && filepath.IsLocal(rel)
}

// forEachBridgeCandidate calls fn for every path in paths that bridgeCandidate
// accepts, descending into host dirs that contain the profile, once per path.
// Each candidate is evaluated just before fn runs, so a parent that fn has
// linked already rules out the paths below it.
func forEachBridgeCandidate(realHome, profileDir string, paths []string, fn func(cleanName, src, dest string)) {
	seen := make(map[string]bool)
	var visit func(name string)
	visit = func(name string) {
		cleanName, src, dest, ok, descend := bridgeCandidate(realHome, profileDir, name)
		switch {
		case ok && !seen[cleanName]:
			seen[cleanName] = true
			fn(cleanName, src, dest)
		case descend:
			entries, err := os.ReadDir(src)
			if err != nil {
				logger.Debug("[symlink] Failed to read %s: %v", src, err)
				return
			}
			for _, e := range entries {
				visit(filepath.Join(cleanName, e.Name()))
			}
		}
	}
	for _, name := range paths {
		visit(name)
	}
}

// underLinkOrFile reports whether any parent of cleanName inside profileDir is
// a symlink or a non-directory. An absent parent is fine: MkdirAll creates a
// real dir for it.
func underLinkOrFile(profileDir, cleanName string) bool {
	dir := filepath.Dir(cleanName)
	if dir == "." {
		return false
	}
	cur := profileDir
	for _, part := range strings.Split(dir, string(filepath.Separator)) {
		cur = filepath.Join(cur, part)
		fi, err := os.Lstat(cur)
		if err != nil {
			return false
		}
		if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
			return true
		}
	}
	return false
}

// EnsureDotfiles links the host's dotfiles, keychains and caches (see
// GetBridgedPaths) into an isolated profile directory so developers keep their
// global developer configs without losing agent isolation. A non-empty real
// entry in the profile is an override and is kept; an empty or stub dir is
// replaced with a link. Nothing else in the profile is ever removed.
func EnsureDotfiles(realHome, profileDir string, extraPaths ...string) error {
	if _, err := os.Stat(profileDir); err != nil {
		return fmt.Errorf("profile directory does not exist: %w", err)
	}

	logger.Debug("[symlink] Ensuring dotfiles for profile at %s (host: %s)", profileDir, realHome)
	var errs []error

	forEachBridgeCandidate(realHome, profileDir, GetBridgedPaths(realHome, extraPaths...), func(cleanName, src, dest string) {
		if fi, err := os.Lstat(dest); err == nil {
			if fi.Mode()&os.ModeSymlink != 0 {
				return
			}
			if !fi.IsDir() || !isStubOrEmptyDir(dest, src, cleanName) {
				return
			}
			if err := os.RemoveAll(dest); err != nil {
				logger.Debug("[symlink] Failed to remove stub dir %s: %v", dest, err)
				errs = append(errs, err)
				return
			}
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			errs = append(errs, err)
			return
		}
		if err := os.Symlink(src, dest); err != nil {
			logger.Debug("[symlink] Failed to bridge %s -> %s: %v", src, dest, err)
			errs = append(errs, err)
		} else {
			logger.Debug("[symlink] Bridged: %s -> %s", cleanName, src)
		}
	})

	return errors.Join(errs...)
}

// isStubOrEmptyDir reports whether the profile dir at path carries nothing of
// its own, so replacing it with a link to src loses nothing: it is empty, a
// scaffold of links into src, or a known agent-generated stub.
func isStubOrEmptyDir(path, src, cleanName string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	if len(entries) == 0 || isHostScaffold(path, src) {
		return true
	}
	if cleanName == filepath.Join(".claude", "plugins") {
		pluginFile := filepath.Join(path, "installed_plugins.json")
		data, err := os.ReadFile(pluginFile)
		if err != nil {
			return false
		}
		var manifest struct {
			Plugins map[string]any `json:"plugins"`
		}
		if err := json.Unmarshal(data, &manifest); err == nil {
			return len(manifest.Plugins) == 0
		}
	}
	if cleanName == ".config/ccstatusline" || cleanName == ".config/cxstatusline" {
		if len(entries) == 1 && entries[0].Name() == "settings.json" {
			return true
		}
	}
	return false
}

// isHostScaffold reports whether dir holds nothing but links to the matching
// entries of hostDir, and dirs of such links that the host has too. That is
// what the old nested allow-list left in a profile's .config, .cargo or .local;
// it has no content of its own. RemoveAll on it removes only the links, never
// their targets. An empty dir is a scaffold, as isStubOrEmptyDir treats it.
func isHostScaffold(dir, hostDir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		h := filepath.Join(hostDir, e.Name())
		switch {
		case e.Type()&os.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil || filepath.Clean(target) != h {
				return false
			}
		case e.IsDir():
			// A subdir the host lacks is the profile's own structure, even empty.
			if fi, err := os.Stat(h); err != nil || !fi.IsDir() || !isHostScaffold(p, h) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
