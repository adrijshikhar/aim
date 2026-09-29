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
	// profileArgs are profiles.<p>.args, which precede args when launched.
	IsBackground(profileArgs, args []string) bool
	// IsSession is false for an invocation that prints version or help and
	// exits: there is nothing to merge. Each CLI spells these differently.
	IsSession(profileArgs, args []string) bool
}

// ArgsMatch is true when a token before the first "--" of profileArgs then
// args (their launch order) is one of flags, or the very first token of
// either is one of firstWords. Checking each first token keeps profile flags
// such as `--model opus` from hiding the CLI's subcommand, while a profile's
// own `["app-server"]` still counts. A "--" in profileArgs makes all of args
// the agent's.
func ArgsMatch(profileArgs, args, flags, firstWords []string) bool {
	if firstWord(profileArgs, firstWords) {
		return true
	}
	if slices.Contains(profileArgs, "--") {
		return hasFlag(profileArgs, flags)
	}
	return hasFlag(profileArgs, flags) || hasFlag(args, flags) || firstWord(args, firstWords)
}

// hasFlag is true when a token before the first "--" is one of flags.
func hasFlag(args, flags []string) bool {
	for _, a := range args {
		if a == "--" {
			return false
		}
		if slices.Contains(flags, a) {
			return true
		}
	}
	return false
}

func firstWord(args, words []string) bool {
	return len(args) > 0 && !strings.HasPrefix(args[0], "-") && slices.Contains(words, args[0])
}
