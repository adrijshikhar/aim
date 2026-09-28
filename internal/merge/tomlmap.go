package merge

import (
	"errors"
	"os"
	"reflect"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

var ErrTOMLMismatch = errors.New("spliced TOML does not decode to the intended map")

type logicalLine struct {
	start, end int      // line indices [start, end)
	header     []string // table path if this is a [header]
	arrayTable bool
	isKV       bool
}

// scanLogical groups physical lines into logical lines, joining multi-line
// arrays, inline tables and multi-line strings.
func scanLogical(lines []string) []logicalLine {
	var out []logicalLine
	for i := 0; i < len(lines); {
		t := strings.TrimSpace(lines[i])
		ll := logicalLine{start: i}
		switch {
		case strings.HasPrefix(t, "[["):
			ll.arrayTable = true
			ll.header = parseHeader(strings.TrimSuffix(strings.TrimPrefix(stripComment(t), "[["), "]]"))
			i++
		case strings.HasPrefix(t, "["):
			ll.header = parseHeader(strings.TrimSuffix(strings.TrimPrefix(stripComment(t), "["), "]"))
			i++
		case t == "" || strings.HasPrefix(t, "#"):
			i++
		default:
			ll.isKV = true
			depth, inML := 0, ""
			for {
				depth, inML = scanDepth(lines[i], depth, inML)
				i++
				if (depth <= 0 && inML == "") || i >= len(lines) {
					break
				}
			}
		}
		ll.end = i
		out = append(out, ll)
	}
	return out
}

func stripComment(s string) string {
	inQ := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQ != 0 {
			if c == inQ {
				inQ = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			inQ = c
		} else if c == '#' {
			return strings.TrimSpace(s[:i])
		}
	}
	return strings.TrimSpace(s)
}

// scanDepth tracks bracket/brace depth and multi-line string state across a line.
func scanDepth(line string, depth int, inML string) (int, string) {
	for i := 0; i < len(line); i++ {
		if inML != "" {
			if strings.HasPrefix(line[i:], inML) {
				i += len(inML) - 1
				inML = ""
			}
			continue
		}
		switch {
		case strings.HasPrefix(line[i:], `"""`):
			inML = `"""`
			i += 2
		case strings.HasPrefix(line[i:], `'''`):
			inML = `'''`
			i += 2
		case line[i] == '"' || line[i] == '\'':
			q := line[i]
			for i++; i < len(line) && line[i] != q; i++ {
				if q == '"' && line[i] == '\\' {
					i++
				}
			}
		case line[i] == '#':
			return depth, inML
		case line[i] == '[' || line[i] == '{':
			depth++
		case line[i] == ']' || line[i] == '}':
			depth--
		}
	}
	return depth, inML
}

func parseHeader(s string) []string {
	var parts []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' || c == '\'':
			j := i + 1
			for j < len(s) && s[j] != c {
				if c == '"' && s[j] == '\\' {
					j++
				}
				j++
			}
			cur.WriteString(s[i+1 : min(j, len(s))])
			i = j
		case c == '.':
			parts = append(parts, strings.TrimSpace(cur.String()))
			cur.Reset()
		case c == ' ' || c == '\t':
		default:
			cur.WriteByte(c)
		}
	}
	return append(parts, strings.TrimSpace(cur.String()))
}

// blocks returns, per entry name under key, the line spans that belong to it.
func blocks(lines []string, key string) (map[string][][2]int, map[string]bool) {
	ll := scanLogical(lines)
	spans := map[string][][2]int{}
	badEntry := map[string]bool{}
	cur := ""
	curStart, lastKV := -1, -1
	flush := func() {
		if cur != "" && curStart >= 0 {
			end := lastKV
			if end < curStart {
				end = curStart + 1
			}
			spans[cur] = append(spans[cur], [2]int{curStart, end})
		}
		cur, curStart, lastKV = "", -1, -1
	}
	inBareKey := false
	for _, l := range ll {
		switch {
		case l.header != nil:
			flush()
			inBareKey = len(l.header) == 1 && l.header[0] == key
			if len(l.header) >= 2 && l.header[0] == key {
				if l.arrayTable {
					badEntry[l.header[1]] = true
					continue
				}
				cur, curStart, lastKV = l.header[1], l.start, l.end
			}
		case l.isKV:
			if cur != "" {
				lastKV = l.end
			} else if inBareKey {
				name := strings.TrimSpace(strings.SplitN(lines[l.start], "=", 2)[0])
				name = strings.SplitN(strings.Trim(name, `"'`), ".", 2)[0]
				badEntry[name] = true
			}
		}
	}
	flush()
	return spans, badEntry
}

func decodeKey(data []byte, key string) (map[string]map[string]any, error) {
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	out := map[string]map[string]any{}
	if t, ok := doc[key].(map[string]any); ok {
		for k, v := range t {
			if m, ok := v.(map[string]any); ok {
				out[k] = m
			}
		}
	}
	return out, nil
}

func ReadTOMLKey(path, key string) (Entries, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewEntries(), false, nil
	}
	if err != nil {
		return Entries{}, true, err
	}
	vals, err := decodeKey(data, key)
	if err != nil {
		return Entries{}, true, err
	}
	lines := strings.SplitAfter(string(data), "\n")
	spans, _ := blocks(lines, key)
	names := make([]string, 0, len(vals))
	for n := range vals {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return firstLine(spans[names[i]]) < firstLine(spans[names[j]]) })
	e := NewEntries()
	for _, n := range names {
		var raw []byte
		for _, sp := range spans[n] {
			raw = append(raw, strings.Join(lines[sp[0]:sp[1]], "")...)
		}
		e.Set(n, vals[n], raw)
	}
	return e, true, nil
}

func firstLine(sp [][2]int) int {
	if len(sp) == 0 {
		return 1 << 30
	}
	return sp[0][0]
}

func WriteTOMLKey(path, key string, desired Entries, minMode os.FileMode) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	current, err := decodeKey(data, key)
	if err != nil {
		return nil, err
	}
	lines := strings.SplitAfter(string(data), "\n")
	spans, bad := blocks(lines, key)

	var skipped []string
	intended := map[string]map[string]any{}
	for n, v := range current {
		intended[n] = v
	}
	drop := map[int]bool{}       // line index → removed
	insertAt := map[int]string{} // line index → text inserted before it
	var appendText strings.Builder

	touch := func(n string) bool {
		if bad[n] {
			skipped = append(skipped, n)
			return false
		}
		return true
	}
	for n := range current {
		if desired.Has(n) {
			continue
		}
		if !touch(n) {
			continue
		}
		for _, sp := range spans[n] {
			for i := sp[0]; i < sp[1]; i++ {
				drop[i] = true
			}
			// Also drop the one blank separator line an add inserted before the
			// block, so repeated add+remove cycles never grow the file.
			if b := sp[0] - 1; b >= 0 && !drop[b] && strings.TrimSpace(lines[b]) == "" {
				drop[b] = true
			}
		}
		delete(intended, n)
	}
	for _, n := range desired.Order {
		v := desired.Values[n]
		if cur, ok := current[n]; ok && reflect.DeepEqual(normNums(cur), normNums(v)) {
			continue
		}
		if !touch(n) {
			continue
		}
		text := string(desired.Raw[n])
		if text == "" {
			b, err := toml.Marshal(map[string]any{key: map[string]any{n: v}})
			if err != nil {
				return nil, err
			}
			text = string(b)
		}
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		if sp := spans[n]; len(sp) > 0 {
			for _, s := range sp {
				for i := s[0]; i < s[1]; i++ {
					drop[i] = true
				}
			}
			insertAt[sp[0][0]] = text
		} else {
			appendText.WriteString("\n" + text)
		}
		intended[n] = v
	}

	var out strings.Builder
	regionStart := openRegionAtEOF(lines)
	for i, l := range lines {
		if i == regionStart && appendText.Len() > 0 {
			out.WriteString(appendText.String()) // "\n<block>" before the open region
			appendText.Reset()
		}
		if t, ok := insertAt[i]; ok {
			out.WriteString(t)
		}
		if !drop[i] {
			out.WriteString(l)
		}
	}
	if appendText.Len() > 0 {
		s := out.String()
		if s != "" && !strings.HasSuffix(s, "\n") {
			out.WriteString("\n")
		}
		out.WriteString(appendText.String())
	}

	check, err := decodeKey([]byte(out.String()), key)
	if err != nil || !sameKeys(check, intended) {
		return skipped, ErrTOMLMismatch
	}
	if !otherTopLevelUnchanged(data, []byte(out.String()), key) {
		return skipped, ErrTOMLMismatch
	}
	return skipped, AtomicWrite(path, []byte(out.String()), minMode)
}

func openRegionAtEOF(lines []string) int {
	open := -1
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "# >>> ") {
			open = i
		} else if strings.HasPrefix(t, "# <<< ") {
			open = -1
		}
	}
	return open
}

func normNums(v any) any {
	switch t := v.(type) {
	case map[string]any:
		o := map[string]any{}
		for k, x := range t {
			o[k] = normNums(x)
		}
		return o
	case []any:
		o := make([]any, len(t))
		for i, x := range t {
			o[i] = normNums(x)
		}
		return o
	case int64:
		return float64(t)
	}
	return v
}

func sameKeys(a, b map[string]map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !reflect.DeepEqual(normNums(v), normNums(b[k])) {
			return false
		}
	}
	return true
}

func otherTopLevelUnchanged(before, after []byte, key string) bool {
	var a, b map[string]any
	if len(before) > 0 {
		if toml.Unmarshal(before, &a) != nil {
			return false
		}
	}
	if toml.Unmarshal(after, &b) != nil {
		return false
	}
	delete(a, key)
	delete(b, key)
	if a == nil {
		a = map[string]any{}
	}
	return reflect.DeepEqual(normNums(a), normNums(b))
}
