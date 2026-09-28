// Package merge merges host agent config collections into profile configs for
// the duration of an aim session.
package merge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Entries is an ordered map of named collection items. Values holds the decoded
// item; Raw holds the item's original bytes (JSON value or TOML block text) when
// it was read from a file, so unchanged items are written back byte-for-byte.
type Entries struct {
	Order  []string
	Values map[string]map[string]any
	Raw    map[string][]byte
}

func NewEntries() Entries {
	return Entries{Values: map[string]map[string]any{}, Raw: map[string][]byte{}}
}

func (e *Entries) Set(name string, value map[string]any, raw []byte) {
	if _, ok := e.Values[name]; !ok {
		e.Order = append(e.Order, name)
	}
	e.Values[name] = value
	if raw != nil {
		e.Raw[name] = raw
	} else {
		delete(e.Raw, name)
	}
}

func (e *Entries) Delete(name string) {
	if _, ok := e.Values[name]; !ok {
		return
	}
	delete(e.Values, name)
	delete(e.Raw, name)
	for i, n := range e.Order {
		if n == name {
			e.Order = append(e.Order[:i], e.Order[i+1:]...)
			break
		}
	}
}

func (e Entries) Has(name string) bool { _, ok := e.Values[name]; return ok }

// Normaliser maps an item to the form used for equality. Files keep real values.
type Normaliser func(map[string]any) map[string]any

// DropEmpty removes keys whose value is an empty string, array or object, recursively.
func DropEmpty(v map[string]any) map[string]any {
	out := make(map[string]any, len(v))
	for k, val := range v {
		switch t := val.(type) {
		case string:
			if t == "" {
				continue
			}
		case []any:
			if len(t) == 0 {
				continue
			}
		case map[string]any:
			n := DropEmpty(t)
			if len(n) == 0 {
				continue
			}
			val = n
		}
		out[k] = val
	}
	return out
}

// Hash returns "sha256:<hex>" of the normalised value as canonical JSON:
// sorted keys, no whitespace, numbers in shortest decimal form.
func Hash(n Normaliser, v map[string]any) string {
	if n == nil {
		n = DropEmpty
	}
	var b strings.Builder
	canon(&b, n(copyMap(v)))
	sum := sha256.Sum256([]byte(b.String()))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func copyMap(v map[string]any) map[string]any {
	out := make(map[string]any, len(v))
	for k, val := range v {
		out[k] = val
	}
	return out
}

func canon(b *strings.Builder, v any) {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			kb, _ := json.Marshal(k)
			b.Write(kb)
			b.WriteByte(':')
			canon(b, t[k])
		}
		b.WriteByte('}')
	case []any:
		b.WriteByte('[')
		for i, x := range t {
			if i > 0 {
				b.WriteByte(',')
			}
			canon(b, x)
		}
		b.WriteByte(']')
	case json.Number:
		f, _ := t.Float64()
		b.WriteString(num(f))
	case float64:
		b.WriteString(num(t))
	case float32:
		b.WriteString(num(float64(t)))
	case int:
		b.WriteString(num(float64(t)))
	case int64:
		b.WriteString(num(float64(t)))
	default:
		vb, _ := json.Marshal(t)
		b.Write(vb)
	}
}

func num(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}
