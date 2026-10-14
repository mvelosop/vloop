// Package gitcfg holds the pure rules of the .git/config guard: which keys an
// ignore pattern matches, and which patterns would blind the guard to keys that
// run code.
package gitcfg

import "strings"

// BuiltinIgnore is the pattern always applied: VS Code writes
// branch.<name>.vscode-merge-base when the work branch is created.
const BuiltinIgnore = "branch.*.vscode-merge-base"

// protectedSections are the sections whose keys run code or pull in more
// configuration, in the order a refusal names them.
var protectedSections = []string{"core", "filter", "include", "includeif"}

// Normalize spells pattern the way git config --list spells keys: the text
// before the first "." and after the last "." lower-cased, a subsection between
// them kept as given. A pattern with no "." is lower-cased whole.
func Normalize(pattern string) string {
	first := strings.Index(pattern, ".")
	if first < 0 {
		return strings.ToLower(pattern)
	}
	last := strings.LastIndex(pattern, ".")
	return strings.ToLower(pattern[:first]) + pattern[first:last] + strings.ToLower(pattern[last:])
}

// Match reports whether pattern, normalized by Normalize, matches key as git
// config --list prints it. A * matches any run of characters, dots included;
// everything else is literal.
func Match(pattern, key string) bool {
	pattern = Normalize(pattern)
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == key
	}
	if !strings.HasPrefix(key, parts[0]) {
		return false
	}
	key = key[len(parts[0]):]
	last := parts[len(parts)-1]
	for _, mid := range parts[1 : len(parts)-1] {
		i := strings.Index(key, mid)
		if i < 0 {
			return false
		}
		key = key[i+len(mid):]
	}
	return strings.HasSuffix(key, last)
}

// Protected reports whether pattern could ignore a key of core, filter,
// include or includeif — the text before its first * and "<section>." being
// one a prefix of the other — and names the first such section.
func Protected(pattern string) (section string, ok bool) {
	head, _, _ := strings.Cut(Normalize(pattern), "*")
	for _, s := range protectedSections {
		p := s + "."
		if strings.HasPrefix(head, p) || strings.HasPrefix(p, head) {
			return s, true
		}
	}
	return "", false
}

// Ignored reports whether any pattern matches key.
func Ignored(patterns []string, key string) bool {
	for _, p := range patterns {
		if Match(p, key) {
			return true
		}
	}
	return false
}
