package merge

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
)

var ErrUnsafeJSON = errors.New("json file has duplicate top-level keys or comments")

type member struct {
	key        string
	start, end int64 // byte span of the whole "key": value member
	valStart   int64
}

// scanTop returns the top-level members of a JSON object with byte offsets.
func scanTop(data []byte) ([]member, error) {
	if bytes.Contains(data, []byte("//")) && !json.Valid(data) {
		return nil, ErrUnsafeJSON
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, errors.New("top level is not an object")
	}
	var out []member
	seen := map[string]bool{}
	for dec.More() {
		keyStart := skipToQuote(data, dec.InputOffset())
		kt, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key := kt.(string)
		if seen[key] {
			return nil, ErrUnsafeJSON
		}
		seen[key] = true
		valStart := skipSpaceColon(data, dec.InputOffset())
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		out = append(out, member{key: key, start: keyStart, valStart: valStart, end: dec.InputOffset()})
	}
	return out, nil
}

func skipToQuote(b []byte, i int64) int64 {
	for i < int64(len(b)) && b[i] != '"' {
		i++
	}
	return i
}

func skipSpaceColon(b []byte, i int64) int64 {
	for i < int64(len(b)) && (b[i] == ' ' || b[i] == '\n' || b[i] == '\t' || b[i] == '\r' || b[i] == ':') {
		i++
	}
	return i
}

func ReadJSONKey(path, key string) (Entries, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewEntries(), false, nil
	}
	if err != nil {
		return Entries{}, true, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return NewEntries(), true, nil // an empty file is an empty object
	}
	members, err := scanTop(data)
	if err != nil {
		return Entries{}, true, err
	}
	for _, m := range members {
		if m.key == key {
			e, err := decodeObjectEntries(data[m.valStart:m.end])
			return e, true, err
		}
	}
	return NewEntries(), true, nil
}

func decodeObjectEntries(obj []byte) (Entries, error) {
	e := NewEntries()
	members, err := scanTop(obj)
	if err != nil {
		return e, err
	}
	for _, m := range members {
		raw := bytes.TrimSpace(obj[m.valStart:m.end])
		var v any
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		if err := d.Decode(&v); err != nil && err != io.EOF {
			return e, err
		}
		mv, ok := v.(map[string]any)
		if !ok {
			mv = map[string]any{"value": v}
		}
		e.Set(m.key, mv, append([]byte(nil), raw...))
	}
	return e, nil
}

// encodeValue unwraps a {"value": v} scalar. An object that is exactly
// {"value": X} with no raw bytes is written back as X; none of the merged
// collections can hold one (spec R4).
func encodeValue(name string, e Entries) []byte {
	if raw, ok := e.Raw[name]; ok {
		return raw
	}
	v := e.Values[name]
	var toEnc any = v
	if len(v) == 1 {
		if s, ok := v["value"]; ok {
			toEnc = s
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("    ", "  ")
	_ = enc.Encode(toEnc)
	return bytes.TrimRight(buf.Bytes(), "\n")
}

func renderObject(e Entries) []byte {
	var b bytes.Buffer
	b.WriteString("{")
	for i, name := range e.Order {
		if i > 0 {
			b.WriteString(",")
		}
		kb, _ := json.Marshal(name)
		b.WriteString("\n    ")
		b.Write(kb)
		b.WriteString(": ")
		b.Write(encodeValue(name, e))
	}
	if len(e.Order) > 0 {
		b.WriteString("\n  ")
	}
	b.WriteString("}")
	return b.Bytes()
}

func WriteJSONKey(path, key string, desired Entries, minMode os.FileMode) error {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		data = []byte("{}")
	}
	members, err := scanTop(data)
	if err != nil {
		return err
	}
	obj := renderObject(desired)
	for _, m := range members {
		if m.key == key {
			out := append(append(append([]byte(nil), data[:m.valStart]...), obj...), data[m.end:]...)
			return AtomicWrite(path, out, minMode)
		}
	}
	// Append as the last member, directly after the previous one, so that
	// DeleteJSONKey restores the original bytes exactly.
	kb, _ := json.Marshal(key)
	var at int
	var ins []byte
	if len(members) > 0 {
		at = int(members[len(members)-1].end)
		ins = append(ins, ",\n  "...)
	} else {
		at = bytes.IndexByte(data, '{') + 1
		ins = append(ins, "\n  "...)
	}
	ins = append(ins, kb...)
	ins = append(ins, ": "...)
	ins = append(ins, obj...)
	if len(members) == 0 {
		ins = append(ins, '\n')
	}
	out := append(append(append([]byte(nil), data[:at]...), ins...), data[at:]...)
	return AtomicWrite(path, out, minMode)
}

// JSONHasKey reports whether the file's top-level object has key.
func JSONHasKey(path, key string) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) || (err == nil && len(bytes.TrimSpace(data)) == 0) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	members, err := scanTop(data)
	if err != nil {
		return false, err
	}
	for _, m := range members {
		if m.key == key {
			return true, nil
		}
	}
	return false, nil
}

// DeleteJSONKey removes one top-level member, keeping every other byte.
func DeleteJSONKey(path, key string, minMode os.FileMode) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	members, err := scanTop(data)
	if err != nil {
		return err
	}
	for k, m := range members {
		if m.key != key {
			continue
		}
		var from, to int
		switch {
		case len(members) == 1:
			from, to = bytes.IndexByte(data, '{')+1, bytes.LastIndexByte(data, '}')
		case k > 0:
			from, to = int(members[k-1].end), int(m.end)
		default:
			from, to = int(m.start), int(members[1].start)
		}
		out := append(append([]byte(nil), data[:from]...), data[to:]...)
		return AtomicWrite(path, out, minMode)
	}
	return nil
}

// ReplaceInJSONMember replaces old with new inside the value of one top-level
// member only, keeping every other byte. It reports whether the file changed;
// an absent key, an empty old or no match is (false, nil).
func ReplaceInJSONMember(path, key, old, new string, minMode os.FileMode) (bool, error) {
	if old == "" {
		return false, nil // bytes.ReplaceAll would insert new between every byte
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	members, err := scanTop(data)
	if err != nil {
		return false, err
	}
	for _, m := range members {
		if m.key != key {
			continue
		}
		val := data[m.valStart:m.end]
		if !bytes.Contains(val, []byte(old)) {
			return false, nil
		}
		repl := bytes.ReplaceAll(val, []byte(old), []byte(new))
		out := append(append(append([]byte(nil), data[:m.valStart]...), repl...), data[m.end:]...)
		return true, AtomicWrite(path, out, minMode)
	}
	return false, nil
}
