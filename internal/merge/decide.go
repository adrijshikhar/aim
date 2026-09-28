package merge

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

var kindMark = map[ChangeKind]string{Added: "+", Edited: "~", Removed: "−"}

// label names the change with its collection (Claude and Codex merge more
// than one) and the new enablement of a plugin, so a host-wide disable is
// never promoted blind: the value of a scalar entry (Claude's enabledPlugins)
// or the enabled flag of a table (Codex's [plugins."x"]).
func label(c Change) string {
	what := string(c.Kind)
	if c.Kind != Added {
		what += " (host item)"
	}
	if c.Kind != Removed {
		if v, ok := c.Value["value"]; ok && len(c.Value) == 1 {
			what += fmt.Sprintf(" → %v", v)
		} else if on, ok := c.Value["enabled"].(bool); ok {
			what += fmt.Sprintf(" → enabled=%v", on)
		}
	}
	if c.HostChanged {
		what += "   host changed since start — promote is refused"
	}
	return fmt.Sprintf("  %s %-34s %s", kindMark[c.Kind], c.Collection.Name+"/"+c.Name, what)
}

// Prompter returns the exit decision function (spec §5 End step 2). Without a
// terminal, or when input ends before an answer, the remaining changes are kept.
func Prompter(in io.Reader, out io.Writer, interactive bool, profile, agent string) func([]Change) []Decision {
	r := bufio.NewReader(in)
	ask := func(q string) (string, bool) {
		fmt.Fprint(out, q)
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			return "", false
		}
		return strings.ToLower(strings.TrimSpace(line)), true
	}
	// keepFrom keeps every change from index i on and says so.
	keepFrom := func(d []Decision, i int) []Decision {
		for j := i; j < len(d); j++ {
			d[j] = Keep
		}
		fmt.Fprintf(out, "%d change(s) kept in %s (%s)\n", len(d)-i, profile, agent)
		return d
	}
	return func(cs []Change) []Decision {
		d := make([]Decision, len(cs))
		if !interactive {
			return keepFrom(d, 0)
		}
		fmt.Fprintf(out, "Session in %s (%s) changed:\n", profile, agent)
		for _, c := range cs {
			fmt.Fprintln(out, label(c))
		}
		// promote asks separately before removing an item from the host.
		promote := func(i int) bool {
			if cs[i].Kind != Removed {
				d[i] = Promote
				return true
			}
			a, ok := ask(fmt.Sprintf("Remove %s from the host (every profile)? [y/N] ", cs[i].Name))
			if a == "y" || a == "yes" {
				d[i] = Promote
			}
			return ok
		}
		a, ok := ask("[p] promote all to host   [k] keep all in " + profile + "   [r] review each   (default: k) ")
		if !ok {
			return keepFrom(d, 0)
		}
		switch a {
		case "p":
			for i := range cs {
				if !promote(i) {
					return keepFrom(d, i)
				}
			}
		case "r":
			for i, c := range cs {
				a, ok := ask(label(c) + "  [p]romote / [k]eep? ")
				if !ok {
					return keepFrom(d, i)
				}
				if a == "p" && !promote(i) {
					return keepFrom(d, i)
				}
			}
		}
		return d
	}
}
