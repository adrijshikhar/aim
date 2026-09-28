package merge

import (
	"bytes"
	"strings"
	"testing"
)

func changes() []Change {
	c := Collection{Agent: "claude", Name: "mcpServers"}
	return []Change{{Collection: c, Name: "foo", Kind: Added}, {Collection: c, Name: "jev", Kind: Edited}}
}

func TestPrompter_NonInteractiveKeepsAll(t *testing.T) {
	var out bytes.Buffer
	d := Prompter(strings.NewReader(""), &out, false, "work", "claude")(changes())
	if d[0] != Keep || d[1] != Keep {
		t.Fatalf("got %v", d)
	}
	if !strings.Contains(out.String(), "2 change(s) kept in work (claude)") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestPrompter_InputEndsKeepsTheRest(t *testing.T) {
	var out bytes.Buffer
	d := Prompter(strings.NewReader("r\np\n"), &out, true, "work", "claude")(changes())
	if d[0] != Promote || d[1] != Keep {
		t.Fatalf("an unanswered change must be kept, got %v", d)
	}
	if !strings.Contains(out.String(), "1 change(s) kept in work (claude)") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestPrompter_PromoteAll(t *testing.T) {
	var out bytes.Buffer
	d := Prompter(strings.NewReader("p\n"), &out, true, "work", "claude")(changes())
	if d[0] != Promote || d[1] != Promote {
		t.Fatalf("got %v", d)
	}
	if !strings.Contains(out.String(), "+ mcpServers/foo") || !strings.Contains(out.String(), "~ mcpServers/jev") {
		t.Fatalf("listing missing:\n%s", out.String())
	}
}

func TestPrompter_ReviewEach(t *testing.T) {
	d := Prompter(strings.NewReader("r\np\nk\n"), &bytes.Buffer{}, true, "work", "claude")(changes())
	if d[0] != Promote || d[1] != Keep {
		t.Fatalf("got %v", d)
	}
}

func TestPrompter_RemovedPromoteNeedsSecondConfirm(t *testing.T) {
	c := []Change{{Collection: Collection{Agent: "claude", Name: "mcpServers"}, Name: "ctx", Kind: Removed}}
	if d := Prompter(strings.NewReader("p\nn\n"), &bytes.Buffer{}, true, "work", "claude")(c); d[0] != Keep {
		t.Fatal("removing from the host needs its own confirmation")
	}
	if d := Prompter(strings.NewReader("p\ny\n"), &bytes.Buffer{}, true, "work", "claude")(c); d[0] != Promote {
		t.Fatal("y confirms the removal")
	}
}

func TestPrompter_DefaultIsKeep(t *testing.T) {
	d := Prompter(strings.NewReader("\n"), &bytes.Buffer{}, true, "work", "claude")(changes())
	if d[0] != Keep || d[1] != Keep {
		t.Fatal("empty answer must keep")
	}
}

func TestPrompter_HostChangedIsLabelled(t *testing.T) {
	var out bytes.Buffer
	c := []Change{{Collection: Collection{Agent: "claude", Name: "mcpServers"}, Name: "jev", Kind: Edited, HostChanged: true}}
	Prompter(strings.NewReader("k\n"), &out, true, "work", "claude")(c)
	if !strings.Contains(out.String(), "host changed since start") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestLabel_PrefixesCollectionAndShowsScalarValue(t *testing.T) {
	plug := Collection{Agent: "claude", Name: "enabledPlugins"}
	cases := []struct {
		c    Change
		want string
	}{
		{Change{Collection: plug, Name: "x@m", Kind: Edited, Value: map[string]any{"value": false}},
			"  ~ enabledPlugins/x@m" + strings.Repeat(" ", 34-len("enabledPlugins/x@m")) + " edited (host item) → false"},
		{Change{Collection: plug, Name: "y@m", Kind: Added, Value: map[string]any{"value": true}},
			"  + enabledPlugins/y@m" + strings.Repeat(" ", 34-len("enabledPlugins/y@m")) + " added → true"},
		{Change{Collection: plug, Name: "z@m", Kind: Removed, Value: map[string]any{"value": true}},
			"  − enabledPlugins/z@m" + strings.Repeat(" ", 34-len("enabledPlugins/z@m")) + " removed (host item)"},
		{Change{Collection: Collection{Name: "mcpServers"}, Name: "jev", Kind: Edited, Value: map[string]any{"command": "npx"}},
			"  ~ mcpServers/jev" + strings.Repeat(" ", 34-len("mcpServers/jev")) + " edited (host item)"},
	}
	for _, tc := range cases {
		if got := label(tc.c); got != tc.want {
			t.Errorf("label =\n%q\nwant\n%q", got, tc.want)
		}
	}
	hc := label(Change{Collection: plug, Name: "x@m", Kind: Edited, Value: map[string]any{"value": false}, HostChanged: true})
	if !strings.Contains(hc, "edited (host item) → false   host changed since start") {
		t.Fatalf("the value goes before the host-changed note: %q", hc)
	}
}

// A Codex plugin is a table ([plugins."x"] enabled = false), not a scalar: its
// enabled flag is shown too, so a host-wide disable is never promoted blind.
func TestLabel_ShowsEnabledFlagOfPluginTable(t *testing.T) {
	plug := Collection{Agent: "codex", Name: "plugins"}
	pad := func(n string) string { return strings.Repeat(" ", 34-len(n)) }
	cases := []struct {
		c    Change
		want string
	}{
		{Change{Collection: plug, Name: "x@m", Kind: Edited, Value: map[string]any{"enabled": false}},
			"  ~ plugins/x@m" + pad("plugins/x@m") + " edited (host item) → enabled=false"},
		{Change{Collection: plug, Name: "y@m", Kind: Added, Value: map[string]any{"enabled": true, "source": "s"}},
			"  + plugins/y@m" + pad("plugins/y@m") + " added → enabled=true"},
		{Change{Collection: plug, Name: "z@m", Kind: Removed, Value: map[string]any{"enabled": true}},
			"  − plugins/z@m" + pad("plugins/z@m") + " removed (host item)"},
		{Change{Collection: plug, Name: "w@m", Kind: Edited, Value: map[string]any{"enabled": "yes"}},
			"  ~ plugins/w@m" + pad("plugins/w@m") + " edited (host item)"},
	}
	for _, tc := range cases {
		if got := label(tc.c); got != tc.want {
			t.Errorf("label =\n%q\nwant\n%q", got, tc.want)
		}
	}
}
