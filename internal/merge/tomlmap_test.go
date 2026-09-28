package merge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const codexTOML = `model = "x"

  [projects."/a"]
  trust_level = "trusted"

[mcp_servers.node_repl]
command = "/bin/node"
args = [
  "a",
  "b",
]
startup_timeout_sec = 120.0

[mcp_servers.node_repl.env]
K = "v" # trailing comment

[plugins."hevo@hevo".mcp_servers.atlassian]
url = "https://x"

[mcp_servers.goland]
url = "http://127.0.0.1:1/stream"

  # >>> caveman:native-tables
  [model_providers.caveman]
    name = "Caveman"
  # <<< caveman:native-tables
`

func write(t *testing.T, s string) string {
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(s), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestTOML_ReadIgnoresNestedPluginServers(t *testing.T) {
	e, ok, err := ReadTOMLKey(write(t, codexTOML), "mcp_servers")
	if err != nil || !ok {
		t.Fatal(err)
	}
	if strings.Join(e.Order, ",") != "goland,node_repl" && strings.Join(e.Order, ",") != "node_repl,goland" {
		t.Fatalf("order = %v", e.Order)
	}
	if e.Has("atlassian") {
		t.Fatal("plugins.*.mcp_servers must not be read as a server")
	}
	if !strings.Contains(string(e.Raw["node_repl"]), "[mcp_servers.node_repl.env]") {
		t.Fatal("block text must include sub-tables")
	}
}

func TestTOML_CavemanRegionAndMultiline(t *testing.T) {
	p := write(t, codexTOML)
	e, _, _ := ReadTOMLKey(p, "mcp_servers")
	e.Delete("goland")
	e.Set("jev", map[string]any{"command": "npx"}, []byte("[mcp_servers.jev]\ncommand = \"npx\"\n"))
	if _, err := WriteTOMLKey(p, "mcp_servers", e, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	s := string(out)
	if !strings.Contains(s, "  # >>> caveman:native-tables\n  [model_providers.caveman]\n    name = \"Caveman\"\n  # <<< caveman:native-tables") {
		t.Fatalf("caveman region changed:\n%s", s)
	}
	if !strings.Contains(s, "args = [\n  \"a\",\n  \"b\",\n]") {
		t.Fatalf("multi-line array damaged:\n%s", s)
	}
	if strings.Contains(s, "goland") {
		t.Fatal("goland not removed")
	}
	if !strings.Contains(s, "[mcp_servers.jev]") || !strings.Contains(s, "[plugins.\"hevo@hevo\".mcp_servers.atlassian]") {
		t.Fatalf("add or nested table wrong:\n%s", s)
	}
	if !strings.Contains(s, `K = "v" # trailing comment`) {
		t.Fatal("untouched block changed")
	}
}

func TestTOML_AppendBeforeOpenManagedRegionAtEOF(t *testing.T) {
	p := write(t, "a = 1\n# >>> caveman:native-tables\n[x]\ny = 1\n")
	e := NewEntries()
	e.Set("n", map[string]any{"command": "t"}, []byte("[mcp_servers.n]\ncommand = \"t\"\n"))
	if _, err := WriteTOMLKey(p, "mcp_servers", e, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	if strings.Index(string(out), "[mcp_servers.n]") > strings.Index(string(out), "# >>> caveman") {
		t.Fatalf("written inside managed region:\n%s", out)
	}
}

func TestTOML_InlineEntryRefusedPerEntry(t *testing.T) {
	p := write(t, "[mcp_servers]\ninline = { command = \"x\" }\n\n[mcp_servers.ok]\ncommand = \"y\"\n")
	e, _, _ := ReadTOMLKey(p, "mcp_servers")
	e.Delete("inline")
	e.Delete("ok")
	skipped, err := WriteTOMLKey(p, "mcp_servers", e, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 1 || skipped[0] != "inline" {
		t.Fatalf("skipped = %v", skipped)
	}
	out, _ := os.ReadFile(p)
	if strings.Contains(string(out), "[mcp_servers.ok]") || !strings.Contains(string(out), "inline = ") {
		t.Fatalf("got:\n%s", out)
	}
}

func TestTOML_KeyMissingIsEmpty(t *testing.T) {
	e, ok, err := ReadTOMLKey(write(t, "model = \"x\"\n"), "mcp_servers")
	if err != nil || !ok || len(e.Order) != 0 {
		t.Fatalf("ok=%v err=%v order=%v", ok, err, e.Order)
	}
}

func TestTOML_ReadRealConfig(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".codex", "config.toml"))
	if err != nil {
		t.Skip("no host codex config")
	}
	p := write(t, string(src))
	e, _, err := ReadTOMLKey(p, "mcp_servers")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := WriteTOMLKey(p, "mcp_servers", e, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	if string(out) != string(src) {
		t.Fatal("round trip of unchanged entries must be byte-identical")
	}
}

func TestTOML_RepeatedAddRemoveDoesNotGrow(t *testing.T) {
	for _, src := range []string{
		"model = \"x\"\n\n[mcp_servers.own]\ncommand = \"o\"\n\n  # >>> caveman:native-tables\n  [model_providers.caveman]\n    name = \"Caveman\"\n  # <<< caveman:native-tables\n",
		"a = 1\n# >>> caveman:native-tables\n[x]\ny = 1\n",
		"",
	} {
		p := write(t, src)
		for i := 0; i < 4; i++ {
			e, _, err := ReadTOMLKey(p, "mcp_servers")
			if err != nil {
				t.Fatal(err)
			}
			e.Set("jev", map[string]any{"command": "npx"}, []byte("[mcp_servers.jev]\ncommand = \"npx\"\n"))
			if _, err := WriteTOMLKey(p, "mcp_servers", e, 0o600); err != nil {
				t.Fatal(err)
			}
			e, _, _ = ReadTOMLKey(p, "mcp_servers")
			e.Delete("jev")
			if _, err := WriteTOMLKey(p, "mcp_servers", e, 0o600); err != nil {
				t.Fatal(err)
			}
			out, _ := os.ReadFile(p)
			if string(out) != src {
				t.Fatalf("cycle %d changed the file:\n--- want\n%s--- got\n%s", i+1, src, out)
			}
		}
	}
}

// An entry with no raw bytes is synthesised by the marshaller; two of them in a
// row must not define the bare [mcp_servers] super-table twice.
func TestTOML_SynthesisedAddsDoNotRepeatSuperTable(t *testing.T) {
	p := write(t, "model = \"x\"\n")
	e := NewEntries()
	for _, n := range []string{"first", "second"} {
		e.Set(n, map[string]any{"command": "true"}, nil)
		if skipped, err := WriteTOMLKey(p, "mcp_servers", e, 0o600); err != nil || len(skipped) != 0 {
			t.Fatalf("add %s: skipped=%v err=%v", n, skipped, err)
		}
	}
	got, _, err := ReadTOMLKey(p, "mcp_servers")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got.Order, ",") != "first,second" {
		t.Fatalf("entries = %v, want [first second]", got.Order)
	}
	data, _ := os.ReadFile(p)
	if strings.Contains(string(data), "[mcp_servers]\n") {
		t.Fatalf("bare super-table header written:\n%s", data)
	}
}
