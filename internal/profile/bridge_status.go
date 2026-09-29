package profile

import (
	"os"
	"path/filepath"
)

// BridgeKind classifies how a profile's copy of a bridged path relates to the host's.
type BridgeKind int

const (
	// BridgeLinked: the profile path is a symlink that resolves to the host path.
	BridgeLinked BridgeKind = iota
	// BridgePending: the profile path is absent or an empty/stub dir, which
	// EnsureDotfiles replaces with a link on the next launch.
	BridgePending
	// BridgeForeignLink: the profile path is a symlink to somewhere else, or dangling.
	BridgeForeignLink
	// BridgeCopy: the profile path is a regular file or non-stub dir that
	// EnsureDotfiles preserves as a profile override, so the host's is never used.
	BridgeCopy
)

// BridgeState is the observed state of one bridged path inside a profile.
type BridgeState struct {
	Path        string // bridged path relative to home, e.g. ".hevo"
	ProfilePath string // absolute path inside the profile
	Kind        BridgeKind
	Target      string // raw symlink target, for BridgeForeignLink
}

// BridgeStatus reports, for each bridged path that exists on the host, whether
// profileDir shares it. It applies the same allow-list and stub rules as
// EnsureDotfiles and never modifies anything.
func BridgeStatus(realHome, profileDir string, paths []string) []BridgeState {
	var states []BridgeState
	for _, name := range paths {
		cleanName := filepath.Clean(filepath.FromSlash(name))
		if !isAllowedBridgedPath(cleanName) {
			continue
		}
		src := filepath.Join(realHome, cleanName)
		if _, err := os.Lstat(src); err != nil {
			continue
		}
		dest := filepath.Join(profileDir, cleanName)
		st := BridgeState{Path: cleanName, ProfilePath: dest}
		fi, err := os.Lstat(dest)
		switch {
		case err != nil:
			st.Kind = BridgePending
		case fi.Mode()&os.ModeSymlink != 0:
			st.Kind = BridgeForeignLink
			st.Target, _ = os.Readlink(dest)
			if sameFile(dest, src) {
				st.Kind = BridgeLinked
			}
		case fi.IsDir() && isStubOrEmptyDir(dest, cleanName):
			st.Kind = BridgePending
		default:
			st.Kind = BridgeCopy
		}
		states = append(states, st)
	}
	return states
}

// sameFile reports whether a and b resolve to the same path; a dangling link never matches.
func sameFile(a, b string) bool {
	ra, err := filepath.EvalSymlinks(a)
	if err != nil {
		return false
	}
	rb, err := filepath.EvalSymlinks(b)
	return err == nil && ra == rb
}
