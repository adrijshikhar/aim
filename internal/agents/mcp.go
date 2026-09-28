package agents

import (
	"slices"
	"strings"

	"github.com/aim-cli/aim/internal/merge"
)

// MCPProvider is implemented by adapters whose config holds an MCP server map
// that aim merges per session (spec §4).
type MCPProvider interface {
	MCPCollections(profileDir, realHome string) []merge.Collection
	// IsBackground reports whether the native args start a session that
	// outlives the child aim waits on (spec R2): merge, run, no exit step.
	IsBackground(args []string) bool
}

// BackgroundMatch is true when a token before the first "--" is one of flags,
// or the very first token is one of firstWords.
func BackgroundMatch(args, flags, firstWords []string) bool {
	for _, a := range args {
		if a == "--" {
			break
		}
		if slices.Contains(flags, a) {
			return true
		}
	}
	return len(args) > 0 && !strings.HasPrefix(args[0], "-") && slices.Contains(firstWords, args[0])
}
