package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
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
	// openTTY opens the controlling terminal for the prompt when stdin is
	// piped. Replaced in tests.
	openTTY = func() (io.ReadWriteCloser, error) { return os.OpenFile("/dev/tty", os.O_RDWR, 0) }
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

// withSessionMerge merges the host's MCP servers and plugin enablement into the
// profile around run (spec §5): Start before, Diff + decide + Finish after. A
// background launch (IsBackground) merges and skips the exit step.
// profiles.<p>.args precede extraArgs in the launched command, so both checks
// see them separately: a background arg there counts too, and profile flags
// such as `--model o3` do not hide a first word like `app-server` in
// extraArgs. A non-session invocation (the adapter's version or help
// spelling) runs without Start at all, so a crashed or background session is
// recovered by the next real session.
func withSessionMerge(adapter agents.AgentAdapter, store merge.Store, profileName, pDir string, cfg *config.Config, extraArgs []string, run func() int) int {
	mp, ok := adapter.(agents.MCPProvider)
	if !ok {
		return run()
	}
	profileArgs := cfg.GetProfileArgs(profileName)
	if !mp.IsSession(profileArgs, extraArgs) {
		return run()
	}
	eng := &merge.Engine{Store: store, Out: os.Stderr, Now: nowFunc}
	cols := mp.MCPCollections(pDir, config.RealHomeDir())
	// Every collection goes to Start, switched-off groups included: recovery
	// must still strip what an earlier background launch merged (spec §4).
	enabled := func(c merge.Collection) bool { return cfg.GroupEnabled(profileName, c.Group) }
	opt := merge.StartOptions{Enabled: enabled, Background: mp.IsBackground(profileArgs, extraArgs)}
	sess, err := eng.Start(profileName, adapter.Name(), cols, opt)
	if err != nil {
		fmtWarn("aim: host merge: %v", err)
	}
	var sigs chan os.Signal
	if sess != nil {
		// The runner forwards SIGINT/SIGTERM/SIGHUP to the agent; until Finish
		// they must not kill aim, or the host's items stay in the profile until
		// the next launch recovers them.
		sigs = make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		defer signal.Stop(sigs)
	}
	code := run()
	if sess != nil {
		changes, err := sess.Diff()
		if err != nil {
			fmtWarn("aim: host diff: %v", err)
		}
		if err := sess.Finish(changes, exitDecider(sigs, profileName, adapter.Name())); err != nil {
			fmtWarn("aim: host finish: %v", err)
		}
	}
	return code
}

// exitDecider prompts on stdin when it is a terminal, else on the controlling
// terminal; without either every change is kept. SIGINT, SIGTERM or SIGHUP
// at the prompt keeps every change.
func exitDecider(sigs chan os.Signal, profile, agent string) func([]merge.Change) []merge.Decision {
	return func(cs []merge.Change) []merge.Decision {
		in, out, interactive := sessionPromptIn, io.Writer(os.Stderr), stdinIsTerminal()
		if !interactive {
			if tty, err := openTTY(); err == nil {
				defer tty.Close()
				in, out, interactive = tty, tty, true
			}
		}
		decide := merge.Prompter(in, out, interactive, profile, agent)
		if !interactive {
			return decide(cs)
		}
		for len(sigs) > 0 { // the Ctrl+C that ended the agent is not an answer
			<-sigs
		}
		res := make(chan []merge.Decision, 1)
		go func() { res <- decide(cs) }()
		select {
		case d := <-res:
			return d
		case <-sigs:
			fmt.Fprintf(out, "\n%d change(s) kept in %s (%s)\n", len(cs), profile, agent)
			return make([]merge.Decision, len(cs)) // Keep is the zero Decision
		}
	}
}
