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
	if !strings.Contains(out.String(), "+ foo") || !strings.Contains(out.String(), "~ jev") {
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
