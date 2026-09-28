package merge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollection_DispatchesByFormat(t *testing.T) {
	d := t.TempDir()
	jp := filepath.Join(d, "a.json")
	tp := filepath.Join(d, "a.toml")
	_ = os.WriteFile(jp, []byte(`{"mcpServers":{"x":{"command":"a"}}}`), 0o600)
	_ = os.WriteFile(tp, []byte("[mcp_servers.x]\ncommand = \"a\"\n"), 0o600)
	j := Collection{Agent: "claude", Name: "mcpServers", Format: JSON, Key: "mcpServers"}
	tt := Collection{Agent: "codex", Name: "mcp_servers", Format: TOML, Key: "mcp_servers"}
	for _, c := range []struct {
		c Collection
		p string
	}{{j, jp}, {tt, tp}} {
		e, ok, err := c.c.Read(c.p)
		if err != nil || !ok || !e.Has("x") {
			t.Fatalf("%s: ok=%v err=%v", c.c.ID(), ok, err)
		}
	}
	if j.ID() != "claude/mcpServers" {
		t.Fatal(j.ID())
	}
}

func TestCollection_HasKeyAndRemoveKey(t *testing.T) {
	d := t.TempDir()
	jp := filepath.Join(d, "a.json")
	_ = os.WriteFile(jp, []byte(`{"x": 1}`), 0o600)
	j := Collection{Agent: "claude", Name: "mcpServers", Format: JSON, Key: "mcpServers"}
	if ok, _ := j.HasKey(jp); ok {
		t.Fatal("key must be absent")
	}
	if _, err := j.Write(jp, NewEntries()); err != nil {
		t.Fatal(err)
	}
	if ok, _ := j.HasKey(jp); !ok {
		t.Fatal("key must exist after write")
	}
	if err := j.RemoveKey(jp); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(jp); string(b) != `{"x": 1}` {
		t.Fatalf("got %s", b)
	}
	tp := filepath.Join(d, "a.toml")
	_ = os.WriteFile(tp, []byte("model = \"x\"\n"), 0o600)
	tt := Collection{Agent: "codex", Name: "mcp_servers", Format: TOML, Key: "mcp_servers"}
	if ok, err := tt.HasKey(tp); ok || err != nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
