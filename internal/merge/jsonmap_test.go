package merge

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const claudeJSON = `{
  "oauthAccount": {"email": "x"},
  "projects": {"/a": {"allowedTools": []}},
  "mcpServers": {
    "mine": {"command": "a"}
  },
  "big": 12345678901234567890
}`

func TestJSON_RewritesOnlyKeyAndPreservesBytes(t *testing.T) {
	p := filepath.Join(t.TempDir(), ".claude.json")
	_ = os.WriteFile(p, []byte(claudeJSON), 0o600)
	e, ok, err := ReadJSONKey(p, "mcpServers")
	if err != nil || !ok || !e.Has("mine") {
		t.Fatalf("read: ok=%v err=%v entries=%v", ok, err, e.Order)
	}
	e.Set("jev", map[string]any{"command": "npx", "url": "a<b&c"}, nil)
	if err := WriteJSONKey(p, "mcpServers", e, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	s := string(out)
	for _, want := range []string{`"oauthAccount": {"email": "x"}`, `"projects": {"/a": {"allowedTools": []}}`,
		`"big": 12345678901234567890`, `"mine": {"command": "a"}`, `"a<b&c"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in\n%s", want, s)
		}
	}
	e2, _, _ := ReadJSONKey(p, "mcpServers")
	if len(e2.Order) != 2 || e2.Order[0] != "mine" || e2.Order[1] != "jev" {
		t.Fatalf("order = %v", e2.Order)
	}
}

func TestJSON_KeyAbsentIsAppended(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s.json")
	_ = os.WriteFile(p, []byte(`{"a": 1}`), 0o600)
	e := NewEntries()
	e.Set("x@y", map[string]any{}, []byte("true"))
	if err := WriteJSONKey(p, "enabledPlugins", e, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	if !strings.HasPrefix(string(out), `{"a": 1,`) || !strings.Contains(string(out), `"enabledPlugins"`) {
		t.Fatalf("got %s", out)
	}
}

func TestJSON_DuplicateTopLevelKeyRefused(t *testing.T) {
	p := filepath.Join(t.TempDir(), "d.json")
	_ = os.WriteFile(p, []byte(`{"mcpServers": {}, "mcpServers": {}}`), 0o600)
	if _, _, err := ReadJSONKey(p, "mcpServers"); err != ErrUnsafeJSON {
		t.Fatalf("err = %v, want ErrUnsafeJSON", err)
	}
}

func TestJSON_MissingFile(t *testing.T) {
	_, ok, err := ReadJSONKey(filepath.Join(t.TempDir(), "none.json"), "mcpServers")
	if ok || err != nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestJSON_EmptyFileIsEmptyObject(t *testing.T) {
	p := filepath.Join(t.TempDir(), "mcp_config.json")
	_ = os.WriteFile(p, nil, 0o600)
	e, ok, err := ReadJSONKey(p, "mcpServers")
	if err != nil || !ok || len(e.Order) != 0 {
		t.Fatalf("ok=%v err=%v order=%v", ok, err, e.Order)
	}
	e.Set("a", map[string]any{"command": "x"}, nil)
	if err := WriteJSONKey(p, "mcpServers", e, 0o600); err != nil {
		t.Fatal(err)
	}
	got, _, err := ReadJSONKey(p, "mcpServers")
	if err != nil || !got.Has("a") {
		t.Fatalf("write to empty file: err=%v order=%v", err, got.Order)
	}
}

func TestJSON_AppendThenDeleteKeyRestoresBytes(t *testing.T) {
	for _, src := range []string{`{"a": 1}`, "{\n  \"a\": 1\n}\n", `{}`} {
		p := filepath.Join(t.TempDir(), "s.json")
		_ = os.WriteFile(p, []byte(src), 0o600)
		e := NewEntries()
		e.Set("x@y", map[string]any{"value": true}, []byte("true"))
		if err := WriteJSONKey(p, "enabledPlugins", e, 0o600); err != nil {
			t.Fatal(err)
		}
		if ok, err := JSONHasKey(p, "enabledPlugins"); !ok || err != nil {
			t.Fatalf("key not added: ok=%v err=%v", ok, err)
		}
		if err := DeleteJSONKey(p, "enabledPlugins", 0o600); err != nil {
			t.Fatal(err)
		}
		out, _ := os.ReadFile(p)
		if string(out) != src {
			t.Fatalf("got %q, want %q", out, src)
		}
	}
}

// Escapes, braces and the key's own name inside strings must not confuse the
// member scan: only the key's value is rewritten.
func TestJSON_StringsThatLookLikeStructurePreserveBytes(t *testing.T) {
	head := "{\n  \"note\": \"a \\\"q\\\" \\\\ {x} \\\"mcpServers\\\": {\\\"y\\\": 1} \\\\\",\n  \"mcpServers\": "
	mine := `{"command": "a", "args": ["\"mcpServers\": {", "\\", "{", "}", "\\\""]}`
	tail := ",\n  \"tail\": \"}{\\\\\\\"mcpServers\\\"\"\n}\n"
	p := filepath.Join(t.TempDir(), ".claude.json")
	_ = os.WriteFile(p, []byte(head+"{\n    \"mine\": "+mine+"\n  }"+tail), 0o600)
	e, _, err := ReadJSONKey(p, "mcpServers")
	if err != nil || strings.Join(e.Order, ",") != "mine" {
		t.Fatalf("read: order=%v err=%v", e.Order, err)
	}
	if args, _ := e.Values["mine"]["args"].([]any); len(args) != 5 || args[0] != `"mcpServers": {` || args[4] != `\"` {
		t.Fatalf("args = %#v", e.Values["mine"]["args"])
	}
	e.Set("jev", map[string]any{"command": "npx"}, nil)
	if err := WriteJSONKey(p, "mcpServers", e, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	s := string(out)
	if !strings.HasPrefix(s, head) || !strings.HasSuffix(s, tail) {
		t.Fatalf("bytes outside the key changed:\n%s", s)
	}
	if !strings.Contains(s[len(head):len(s)-len(tail)], `"mine": `+mine) {
		t.Fatalf("untouched entry rewritten:\n%s", s)
	}
	got, _, err := ReadJSONKey(p, "mcpServers")
	if err != nil || strings.Join(got.Order, ",") != "mine,jev" {
		t.Fatalf("after write: order=%v err=%v", got.Order, err)
	}
}

func TestReplaceInJSONMember_OnlyInsideTheMember(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.json")
	src := `{
  "hooks": {"Stop": [{"command": "/h/.claude/hooks/stop.sh"}]},
  "extraKnownMarketplaces": {"m": {"source": {"source": "directory", "path": "/h/.claude/mkt"}}},
  "statusLine": {"command": "/h/.claude/status.sh"}
}`
	_ = os.WriteFile(p, []byte(src), 0o644)
	changed, err := ReplaceInJSONMember(p, "hooks", "/h/.claude", "/p/.claude", 0o600)
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	out, _ := os.ReadFile(p)
	want := strings.Replace(src, "/h/.claude/hooks", "/p/.claude/hooks", 1)
	if string(out) != want {
		t.Fatalf("only the hooks value may change:\n%s", out)
	}
	if fi, _ := os.Stat(p); fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %v, want the 0600 floor", fi.Mode().Perm())
	}
}

func TestReplaceInJSONMember_NoOp(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.json")
	src := `{"hooks": {"Stop": []}, "statusLine": {"command": "/h/.claude/s.sh"}}`
	_ = os.WriteFile(p, []byte(src), 0o644)
	for _, tc := range []struct{ key, old string }{{"hooks", "/h/.claude"}, {"missing", "/h/.claude"}, {"statusLine", ""}} {
		changed, err := ReplaceInJSONMember(p, tc.key, tc.old, "/p/.claude", 0o600)
		if err != nil || changed {
			t.Fatalf("%s/%q: changed=%v err=%v", tc.key, tc.old, changed, err)
		}
	}
	out, _ := os.ReadFile(p)
	if string(out) != src {
		t.Fatalf("a no-op must leave the file byte-identical:\n%s", out)
	}
	if fi, _ := os.Stat(p); fi.Mode().Perm() != 0o644 {
		t.Fatalf("a no-op must not rewrite the file: mode %v", fi.Mode().Perm())
	}
	if _, err := ReplaceInJSONMember(filepath.Join(t.TempDir(), "none.json"), "hooks", "a", "b", 0o600); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file: err = %v", err)
	}
}
