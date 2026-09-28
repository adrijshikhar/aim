package merge

import "testing"

func TestHash_IgnoresKeyOrderAndNumberForm(t *testing.T) {
	a := map[string]any{"command": "npx", "timeout": float64(20)}
	b := map[string]any{"timeout": int64(20), "command": "npx"}
	if Hash(DropEmpty, a) != Hash(DropEmpty, b) {
		t.Fatal("hash must ignore key order and int/float form")
	}
}

func TestDropEmpty_RemovesEmptyValuesRecursively(t *testing.T) {
	v := map[string]any{"command": "x", "args": []any{}, "env": map[string]any{},
		"nested": map[string]any{"a": "", "b": "keep"}}
	got := DropEmpty(v)
	if _, ok := got["args"]; ok {
		t.Fatal("empty args kept")
	}
	if _, ok := got["env"]; ok {
		t.Fatal("empty env kept")
	}
	if n := got["nested"].(map[string]any); len(n) != 1 || n["b"] != "keep" {
		t.Fatalf("nested not normalised: %v", n)
	}
}

func TestNormalise_ClaudeNoise(t *testing.T) {
	claude := func(v map[string]any) map[string]any {
		v = DropEmpty(v)
		if v["type"] == "stdio" && v["command"] != nil {
			delete(v, "type")
		}
		return v
	}
	written := map[string]any{"command": "npx", "args": []any{"-y", "jev"}}
	rewritten := map[string]any{"type": "stdio", "command": "npx", "args": []any{"-y", "jev"}, "env": map[string]any{}}
	if Hash(claude, written) != Hash(claude, rewritten) {
		t.Fatal("Claude's own normalisation must not change the hash")
	}
}

func TestEntries_SetKeepsOrderAndDelete(t *testing.T) {
	e := NewEntries()
	e.Set("b", map[string]any{"x": "1"}, []byte(`{"x":"1"}`))
	e.Set("a", map[string]any{"x": "2"}, nil)
	e.Set("b", map[string]any{"x": "3"}, nil)
	if len(e.Order) != 2 || e.Order[0] != "b" || e.Order[1] != "a" {
		t.Fatalf("order = %v", e.Order)
	}
	e.Delete("b")
	if e.Has("b") || len(e.Order) != 1 {
		t.Fatal("delete failed")
	}
}
