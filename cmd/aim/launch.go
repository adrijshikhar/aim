package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/mattn/go-isatty"
)

var (
	nowFunc = time.Now
	// sessionPromptIn feeds the exit decision prompt; stdinIsTerminal decides
	// whether it may be shown. Both are replaced in tests.
	sessionPromptIn io.Reader = os.Stdin
	stdinIsTerminal           = func() bool { return isatty.IsTerminal(os.Stdin.Fd()) }
)

func fmtWarn(f string, a ...any) { fmt.Fprintf(os.Stderr, f+"\n", a...) }

// splitRunFlags removes aim's -y/--yes/--create, but only before the first
// "--": everything after it belongs to the agent (E4).
func splitRunFlags(args []string) (bool, []string) {
	auto := false
	var rest []string
	for i, a := range args {
		if a == "--" {
			rest = append(rest, args[i:]...)
			break
		}
		if a == "-y" || a == "--yes" || a == "--create" {
			auto = true
			continue
		}
		rest = append(rest, a)
	}
	return auto, rest
}

// withSessionMerge merges the host's MCP servers into the profile around run
// (spec §5): Start before, Diff + decide + Finish after. A background launch
// (IsBackground) merges and skips the exit step. profiles.<p>.args precede
// extraArgs in the launched command, so a background arg there counts too.
func withSessionMerge(adapter agents.AgentAdapter, store merge.Store, profileName, pDir string, cfg *config.Config, extraArgs []string, run func() int) int {
	mp, ok := adapter.(agents.MCPProvider)
	if !ok {
		return run()
	}
	eng := &merge.Engine{Store: store, Out: os.Stderr, Now: nowFunc}
	cols := mp.MCPCollections(pDir, config.RealHomeDir())
	native := append(cfg.GetProfileArgs(profileName), extraArgs...) // a copy: extraArgs is not aliased
	opt := merge.StartOptions{Enabled: cfg.MCPGlobalEnabled(profileName), Background: mp.IsBackground(native)}
	sess, err := eng.Start(profileName, adapter.Name(), cols, opt)
	if err != nil {
		fmtWarn("aim: MCP merge: %v", err)
	}
	code := run()
	if sess != nil {
		changes, err := sess.Diff()
		if err != nil {
			fmtWarn("aim: MCP diff: %v", err)
		}
		decide := merge.Prompter(sessionPromptIn, os.Stderr, stdinIsTerminal(), profileName, adapter.Name())
		if err := sess.Finish(changes, decide); err != nil {
			fmtWarn("aim: MCP finish: %v", err)
		}
	}
	return code
}
