package brief

import (
	"fmt"
	"regexp"
	"strings"
)

// Statuses are the values a brief's frontmatter status may take.
var Statuses = []string{"draft", "ready", "consumed", "abandoned"}

var fmKeyRe = regexp.MustCompile(`^([A-Za-z0-9_-]+):(?:\s+(.*)|)$`)

// frontmatter is the parsed header of a brief.
type frontmatter struct {
	Err    string // non-empty when the header is missing or unparseable
	Name   string
	Status string
	Keys   map[string]string
	// DependsOn is the names listed under depends-on, inline or as a block list.
	DependsOn []string
}

// parseFrontmatter reads the leading `---` block: `key: value` lines, comments,
// and block-list items. Anything else is unparseable.
func parseFrontmatter(text string) frontmatter {
	fm := frontmatter{Keys: map[string]string{}}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		fm.Err = "missing frontmatter"
		return fm
	}
	closed := false
	cur := ""
	var block []string
	for i, l := range lines[1:] {
		if strings.TrimSpace(l) == "---" {
			closed = true
			break
		}
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t") || strings.HasPrefix(t, "- ") || t == "-" {
			if cur == "depends-on" && strings.HasPrefix(t, "-") {
				if v := strings.Trim(stripComment(strings.TrimSpace(strings.TrimPrefix(t, "-"))), `"'`); v != "" {
					block = append(block, v)
				}
			}
			continue // continuation or block-list item of the previous key
		}
		m := fmKeyRe.FindStringSubmatch(l)
		if m == nil {
			fm.Err = fmt.Sprintf("unparseable frontmatter: line %d is not `key: value`", i+2)
			return fm
		}
		v := stripComment(m[2])
		if err := checkScalar(v); err != "" {
			fm.Err = fmt.Sprintf("unparseable frontmatter: line %d: %s", i+2, err)
			return fm
		}
		fm.Keys[m[1]] = strings.Trim(v, `"'`)
		cur = m[1]
	}
	if !closed {
		fm.Err = "unparseable frontmatter: no closing ---"
		return fm
	}
	fm.Name = fm.Keys["name"]
	fm.Status = fm.Keys["status"]
	fm.DependsOn = parseNames(strings.TrimSpace(stripComment(fm.Keys["depends-on"])), block)
	return fm
}

// stripComment drops a trailing ` #` comment that is outside quotes.
func stripComment(v string) string {
	var q rune
	for i, r := range v {
		switch {
		case q != 0:
			if r == q {
				q = 0
			}
		case r == '"' || r == '\'':
			q = r
		case r == '#' && (i == 0 || v[i-1] == ' '):
			return strings.TrimSpace(v[:i])
		}
	}
	return strings.TrimSpace(v)
}

// checkScalar rejects unbalanced quotes and flow collections.
func checkScalar(v string) string {
	if v == "" {
		return ""
	}
	switch v[0] {
	case '[':
		if !strings.HasSuffix(v, "]") {
			return "unclosed `[`"
		}
	case '{':
		if !strings.HasSuffix(v, "}") {
			return "unclosed `{`"
		}
	case '"', '\'':
		if len(v) < 2 || v[len(v)-1] != v[0] {
			return "unclosed quote"
		}
	}
	return ""
}

// frontmatterProblems lists every frontmatter problem of a brief; they apply
// whatever its status.
func frontmatterProblems(path string, fm frontmatter) []string {
	if fm.Err != "" {
		return []string{fm.Err}
	}
	var out []string
	if fm.Name == "" {
		out = append(out, "frontmatter has no name")
	} else if want := strings.TrimSuffix(pathBase(path), ".md"); fm.Name != want {
		out = append(out, fmt.Sprintf("name %s does not match filename %s", fm.Name, want))
	}
	ok := false
	for _, s := range Statuses {
		ok = ok || fm.Status == s
	}
	if !ok {
		out = append(out, fmt.Sprintf("status %q is not one of %s", fm.Status, strings.Join(Statuses, "|")))
	}
	return out
}

func pathBase(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// parseNames reads a depends-on value: an inline `[a, b]` list, a single bare
// name, or (when the value is empty) the block-list items.
func parseNames(v string, block []string) []string {
	if v == "" {
		return block
	}
	v = strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")
	var out []string
	for _, n := range strings.Split(v, ",") {
		if n = strings.Trim(strings.TrimSpace(n), `"'`); n != "" {
			out = append(out, n)
		}
	}
	return out
}
