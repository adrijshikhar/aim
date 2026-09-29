package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aim-cli/aim/internal/agents"
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
// profileDir shares it. Pass GetBridgedPaths(realHome, ...) to check the same
// set EnsureDotfiles links. It applies the same deny, stub and linked-parent
// rules as EnsureDotfiles, so a path under a parent that already links to the
// host is covered by that parent's state, the children of a host dir that
// contains the profile are reported one by one, and it never modifies anything.
func BridgeStatus(realHome, profileDir string, paths []string) []BridgeState {
	var states []BridgeState
	forEachBridgeCandidate(realHome, profileDir, paths, func(cleanName, src, dest string) {
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
		case fi.IsDir() && isStubOrEmptyDir(dest, src, cleanName):
			st.Kind = BridgePending
		default:
			st.Kind = BridgeCopy
		}
		states = append(states, st)
	})
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

// BridgeDiagnostics summarises BridgeStatus for aim doctor and the TUI doctor
// drawer: one OK line counting the shared paths, and a WARN for each bridged
// host path the profile does not actually share (a symlink elsewhere, or a
// preserved profile-only copy that shadows the host's). It only warns;
// removing a copy is left to the user.
func BridgeDiagnostics(profileName, realHome, profileDir string, extraPaths ...string) []agents.DiagnosticResult {
	var results []agents.DiagnosticResult
	shared, pending := 0, 0
	for _, s := range BridgeStatus(realHome, profileDir, GetBridgedPaths(realHome, extraPaths...)) {
		switch s.Kind {
		case BridgeLinked:
			shared++
		case BridgePending:
			pending++
		case BridgeForeignLink:
			results = append(results, agents.DiagnosticResult{
				Category: "Bridge",
				Status:   "WARN",
				Message:  fmt.Sprintf("%s: %s links to %s, not the host's", profileName, s.Path, s.Target),
			})
		case BridgeCopy:
			results = append(results, agents.DiagnosticResult{
				Category: "Bridge",
				Status:   "WARN",
				Message: fmt.Sprintf("%s: %s is a profile-only copy; the host's is not used. Remove it to share the host's: rm -rf %s",
					profileName, s.Path, s.ProfilePath),
			})
		}
	}
	if shared+pending > 0 {
		msg := fmt.Sprintf("%d host path(s) shared", shared)
		if pending > 0 {
			msg += fmt.Sprintf(", %d linked on next launch", pending)
		}
		results = append([]agents.DiagnosticResult{{Category: "Bridge", Status: "OK", Message: msg}}, results...)
	}
	return results
}
