package main

import (
	"fmt"
	"strings"
)

// frontmatter is one file's frontmatter: top-level keys, and for a key with no
// value the one-level map of the indented lines that follow it.
type frontmatter struct {
	values map[string]string
	maps   map[string]map[string]string
}

// splitFrontmatter returns the lines between the first two `---` lines and the
// text after the second one.
func splitFrontmatter(text string) (front []string, body string, err error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	open := -1
	for i, l := range lines {
		if strings.TrimRight(l, " \t") != "---" {
			continue
		}
		if open < 0 {
			open = i
			continue
		}
		return lines[open+1 : i], strings.Join(lines[i+1:], "\n"), nil
	}
	if open < 0 {
		return nil, "", fmt.Errorf("no frontmatter")
	}
	return nil, "", fmt.Errorf("frontmatter is never closed")
}

// parseFrontmatter reads the subset the eval cases use: `key: value` at the top
// level, and `key:` followed by lines indented two spaces, each `key: value`.
func parseFrontmatter(lines []string) (frontmatter, error) {
	fm := frontmatter{values: map[string]string{}, maps: map[string]map[string]string{}}
	var cur map[string]string
	for n, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if strings.HasPrefix(l, "  ") && !strings.HasPrefix(l, "   ") {
			if cur == nil {
				return fm, fmt.Errorf("frontmatter line %d: indented line outside a map", n+1)
			}
			k, v, err := splitKeyValue(l[2:], n+1)
			if err != nil {
				return fm, err
			}
			if v == "" {
				return fm, fmt.Errorf("frontmatter line %d: key %s has no value", n+1, k)
			}
			val, err := parseValue(v, n+1)
			if err != nil {
				return fm, err
			}
			cur[k] = val
			continue
		}
		if l[0] == ' ' || l[0] == '\t' {
			return fm, fmt.Errorf("frontmatter line %d: unexpected indentation", n+1)
		}
		cur = nil
		k, v, err := splitKeyValue(l, n+1)
		if err != nil {
			return fm, err
		}
		if _, dup := fm.values[k]; dup {
			return fm, fmt.Errorf("frontmatter line %d: key %s repeated", n+1, k)
		}
		if _, dup := fm.maps[k]; dup {
			return fm, fmt.Errorf("frontmatter line %d: key %s repeated", n+1, k)
		}
		if v == "" {
			cur = map[string]string{}
			fm.maps[k] = cur
			continue
		}
		val, err := parseValue(v, n+1)
		if err != nil {
			return fm, err
		}
		fm.values[k] = val
	}
	return fm, nil
}

func splitKeyValue(l string, line int) (key, value string, err error) {
	i := strings.Index(l, ":")
	if i <= 0 {
		return "", "", fmt.Errorf("frontmatter line %d: not a key: value line", line)
	}
	key = l[:i]
	for j, r := range key {
		ok := r == '_' || r == '-' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (j > 0 && r >= '0' && r <= '9')
		if !ok {
			return "", "", fmt.Errorf("frontmatter line %d: %q is not a key", line, key)
		}
	}
	rest := l[i+1:]
	if rest != "" && rest[0] != ' ' {
		return "", "", fmt.Errorf("frontmatter line %d: no space after the colon of %s", line, key)
	}
	return key, strings.TrimSpace(rest), nil
}

// parseValue reads a plain, single-quoted (” an escaped quote) or
// double-quoted value; a plain value, a double-quoted one's text and a [...]
// value are kept as written, backslashes included.
func parseValue(v string, line int) (string, error) {
	switch v[0] {
	case '\'':
		if len(v) < 2 || v[len(v)-1] != '\'' {
			return "", fmt.Errorf("frontmatter line %d: single-quoted value is not closed", line)
		}
		inner := v[1 : len(v)-1]
		for i := 0; i < len(inner); i++ {
			if inner[i] == '\'' {
				if i+1 < len(inner) && inner[i+1] == '\'' {
					i++
					continue
				}
				return "", fmt.Errorf("frontmatter line %d: stray quote in a single-quoted value", line)
			}
		}
		return strings.ReplaceAll(inner, "''", "'"), nil
	case '"':
		if len(v) < 2 || v[len(v)-1] != '"' {
			return "", fmt.Errorf("frontmatter line %d: double-quoted value is not closed", line)
		}
		return v[1 : len(v)-1], nil
	}
	return v, nil
}
