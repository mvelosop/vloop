package config

import (
	"bytes"
	"strings"

	"github.com/BurntSushi/toml"
)

// The config file is the operator's: Set and AddChecks edit its text in place,
// so comments, order and layout survive. Each edit is checked by decoding the
// result: when it does not hold exactly the intended keys, the caller falls
// back to encoding the whole file.

// cfgLine is one line of the file, as the editor sees it.
type cfgLine struct {
	text  string
	table string // the [table] it sits under; "" at the root, "[[" for an array table
	key   string // the full dotted key it defines, "" for anything else
	end   int    // the index of the value's last line (a multi-line array)
	cont  bool   // a continuation line of a multi-line value
}

// scanLines splits raw into lines and finds the keys each defines.
func scanLines(raw []byte) []cfgLine {
	texts := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(raw) == 0 {
		texts = nil
	}
	out := make([]cfgLine, len(texts))
	table := ""
	for i := 0; i < len(texts); i++ {
		t := strings.TrimSpace(texts[i])
		out[i] = cfgLine{text: texts[i], table: table, end: i}
		switch {
		case strings.HasPrefix(t, "[["):
			table = "[["
			out[i].table = table
		case strings.HasPrefix(t, "["):
			table = strings.TrimSpace(strings.Trim(strings.SplitN(t, "#", 2)[0], "[] \t"))
			out[i].table = table
		case t == "" || strings.HasPrefix(t, "#") || table == "[[":
		default:
			k, v, ok := strings.Cut(t, "=")
			if !ok || strings.ContainsAny(k, `"'`) {
				continue
			}
			k = strings.Join(strings.Fields(strings.ReplaceAll(k, ".", " . ")), "")
			if table != "" {
				k = table + "." + k
			}
			start := i
			out[start].key = k
			for depth := bracketDepth(v); depth > 0 && i+1 < len(texts); depth += bracketDepth(texts[i]) {
				i++
				out[i] = cfgLine{text: texts[i], table: table, end: i, cont: true}
			}
			out[start].end = i
		}
	}
	return out
}

// bracketDepth is the change in [ ] nesting over s, outside strings and
// comments.
func bracketDepth(s string) int {
	d := 0
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == '\\' && quote == '"' {
				i++
			} else if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '#':
			return d
		case c == '[':
			d++
		case c == ']':
			d--
		}
	}
	return d
}

// valueText is how the TOML encoder writes v, as the right side of "=".
func valueText(v any) (string, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(map[string]any{"k": v}); err != nil {
		return "", err
	}
	_, rhs, _ := strings.Cut(strings.TrimSpace(buf.String()), "=")
	return strings.TrimSpace(rhs), nil
}

// editKey sets name to the encoded value rhs in raw, or removes it when rhs is
// "". It reports false when it cannot place the edit.
func editKey(raw []byte, name, rhs string) ([]byte, bool) {
	lines := scanLines(raw)
	table, leaf, dotted := strings.Cut(name, ".")
	if !dotted {
		table, leaf = "", name
	}
	texts := make([]string, 0, len(lines)+3)
	for _, l := range lines {
		texts = append(texts, l.text)
	}
	// An existing definition: replace or remove its lines.
	for i, l := range lines {
		if l.key != name {
			continue
		}
		if rhs == "" {
			texts = append(texts[:i:i], texts[l.end+1:]...)
			return dropEmptyTable(texts, table), true
		}
		lhs, _, _ := strings.Cut(l.text, "=")
		texts = append(texts[:i:i], append([]string{strings.TrimRight(lhs, " \t") + " = " + rhs}, texts[l.end+1:]...)...)
		return join(texts), true
	}
	if rhs == "" {
		return raw, true
	}
	// A new key: beside the keys of its table, written the way they are.
	last, indent, header, rootLast, firstHeader := -1, "", -1, -1, -1
	for i, l := range lines {
		if l.cont {
			continue
		}
		t := strings.TrimSpace(l.text)
		if strings.HasPrefix(t, "[") && l.key == "" && firstHeader < 0 {
			firstHeader = i
		}
		if table != "" && l.table == table && strings.HasPrefix(t, "[") && l.key == "" {
			header = i
		}
		if l.key == "" {
			continue
		}
		if l.table == "" {
			rootLast = l.end
		}
		if table != "" && l.table == table {
			last, indent = l.end, l.text[:len(l.text)-len(strings.TrimLeft(l.text, " \t"))]
		}
		if table != "" && l.table == "" && strings.HasPrefix(l.key, table+".") {
			last, indent = l.end, l.text[:len(l.text)-len(strings.TrimLeft(l.text, " \t"))]
		}
	}
	insert := func(at int, add ...string) []byte {
		return join(append(texts[:at:at], append(add, texts[at:]...)...))
	}
	switch {
	case table == "":
		if rootLast >= 0 {
			return insert(rootLast+1, name+" = "+rhs), true
		}
		if firstHeader >= 0 {
			return insert(firstHeader, name+" = "+rhs, ""), true
		}
		return insert(len(texts), name+" = "+rhs), true
	case last >= 0 && lines[last].table == "":
		return insert(last+1, indent+name+" = "+rhs), true
	case last >= 0:
		return insert(last+1, indent+leaf+" = "+rhs), true
	case header >= 0:
		return insert(header+1, "  "+leaf+" = "+rhs), true
	}
	add := []string{"[" + table + "]", "  " + leaf + " = " + rhs}
	if len(texts) > 0 && strings.TrimSpace(texts[len(texts)-1]) != "" {
		add = append([]string{""}, add...)
	}
	return insert(len(texts), add...), true
}

// dropEmptyTable removes the [table] header a removal left with no keys.
func dropEmptyTable(texts []string, table string) []byte {
	if table == "" {
		return join(texts)
	}
	lines := scanLines(join(texts))
	header := -1
	for i, l := range lines {
		t := strings.TrimSpace(l.text)
		if l.table == table && l.key == "" && !l.cont && strings.HasPrefix(t, "[") {
			header = i
		}
		if l.table == table && l.key != "" {
			return join(texts)
		}
	}
	if header < 0 {
		return join(texts)
	}
	texts = append(texts[:header:header], texts[header+1:]...)
	// The blank line that set the header apart goes with it.
	if header > 0 && strings.TrimSpace(texts[header-1]) == "" && (header == len(texts) || strings.TrimSpace(texts[header]) == "") {
		texts = append(texts[:header-1:header-1], texts[header:]...)
	}
	return join(texts)
}

func join(texts []string) []byte {
	if len(texts) == 0 {
		return nil
	}
	return []byte(strings.Join(texts, "\n") + "\n")
}

// sameKeys reports whether edited decodes to exactly the keys of want.
func sameKeys(edited []byte, want map[string]any) bool {
	got := map[string]any{}
	if _, err := toml.Decode(string(edited), &got); err != nil {
		return false
	}
	var a, b bytes.Buffer
	if toml.NewEncoder(&a).Encode(got) != nil || toml.NewEncoder(&b).Encode(want) != nil {
		return false
	}
	return bytes.Equal(a.Bytes(), b.Bytes())
}
