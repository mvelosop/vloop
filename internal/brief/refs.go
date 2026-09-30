package brief

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// refEntryRe matches a binding-reference list item: a backticked path, then an
// optional ` — ` or ` - ` separator and reason.
var refEntryRe = regexp.MustCompile("^\\s*[-*] +`([^`]+)`(.*)$")

var reasonSepRe = regexp.MustCompile(`^\s+(—|-)\s+(\S.*)$`)

// sectionLines returns the lines of every section whose heading matches re, up
// to the next heading.
func sectionLines(body string, re *regexp.Regexp) []string {
	var out []string
	in := false
	for _, l := range strings.Split(body, "\n") {
		if headingRe.MatchString(l) {
			in = re.MatchString(l)
			continue
		}
		if in {
			out = append(out, l)
		}
	}
	return out
}

type refEntry struct {
	path string
	rest string // everything after the path, continuation lines joined by a space
}

// refEntries groups section lines into list items: an entry line plus the
// non-blank lines indented under it, up to a blank line, the next item or a
// line that is not indented.
func refEntries(lines []string) []refEntry {
	var out []refEntry
	open := false
	for _, l := range lines {
		if m := refEntryRe.FindStringSubmatch(l); m != nil {
			out = append(out, refEntry{path: m[1], rest: m[2]})
			open = true
			continue
		}
		if open && strings.TrimSpace(l) != "" && (l[0] == ' ' || l[0] == '\t') {
			out[len(out)-1].rest += " " + strings.TrimSpace(l)
			continue
		}
		open = false
	}
	return out
}

// bindingRefLines runs the binding-references rules, returning problem lines
// and a warning when the section is absent. dead lists paths already reported
// so they are not reported again as generic unresolved paths.
func bindingRefLines(root, body string, set *HeadingSet) []Line {
	re := heading(set.BindingRefs)
	if !re.MatchString(body) {
		return []Line{{Warning, "no binding references — nothing binds a task to a document"}}
	}
	var out []Line
	bad := 0
	for _, e := range refEntries(sectionLines(body, re)) {
		p := e.path
		if !exists(filepath.Join(root, filepath.FromSlash(p))) {
			out = append(out, Line{Problem, "binding reference does not resolve: " + p})
			bad++
			continue
		}
		if !reasonSepRe.MatchString(e.rest) {
			out = append(out, Line{Problem, "binding reference has no reason: " + p})
			bad++
			continue
		}
		if strings.HasSuffix(p, "/") && !hasEntryPoint(filepath.Join(root, filepath.FromSlash(p))) {
			out = append(out, Line{Problem, "binding reference directory has no entry point: " + p})
			bad++
		}
	}
	if bad == 0 {
		out = append(out, Line{Pass, "binding references resolve and say why"})
	}
	return out
}

func hasEntryPoint(dir string) bool {
	for _, n := range []string{"README.md", "index.md"} {
		if st, err := os.Stat(filepath.Join(dir, n)); err == nil && st.Mode().IsRegular() {
			return true
		}
	}
	m, _ := filepath.Glob(filepath.Join(dir, "README-*.md"))
	return len(m) > 0
}
