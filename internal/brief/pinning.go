package brief

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GuidelinesFile is the repository's own lessons file, relative to the root.
const GuidelinesFile = ".vloop/repo-guidelines.md"

const pinningHeading = "## Pinning tests"

// pinRule is one `- when `phrase`: `name`, `name“ line: a brief whose
// contract mentions the phrase must name every test in Constraints.
type pinRule struct {
	phrase string
	names  []string
}

// pinningRules reads the rules under "## Pinning tests"; other lines and
// malformed ones are ignored. A missing file or section yields none.
func pinningRules(root string) []pinRule {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(GuidelinesFile)))
	if err != nil {
		return nil
	}
	var rules []pinRule
	in := false
	for _, l := range strings.Split(string(data), "\n") {
		l = strings.TrimRight(l, "\r ")
		switch {
		case strings.HasPrefix(l, "#"):
			in = l == pinningHeading
		case in:
			if r, ok := parsePinRule(l); ok {
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func parsePinRule(l string) (pinRule, bool) {
	rest, ok := strings.CutPrefix(l, "- when `")
	if !ok {
		return pinRule{}, false
	}
	phrase, rest, ok := strings.Cut(rest, "`")
	if !ok || phrase == "" {
		return pinRule{}, false
	}
	if rest, ok = strings.CutPrefix(rest, ": "); !ok {
		return pinRule{}, false
	}
	var names []string
	for {
		var name string
		if rest, ok = strings.CutPrefix(rest, "`"); !ok {
			return pinRule{}, false
		}
		if name, rest, ok = strings.Cut(rest, "`"); !ok || name == "" {
			return pinRule{}, false
		}
		names = append(names, name)
		if r, more := strings.CutPrefix(rest, ", "); more {
			rest = r
			continue
		}
		break
	}
	if rest != "" && !strings.HasPrefix(rest, " (") {
		return pinRule{}, false
	}
	return pinRule{phrase, names}, true
}

// sectionText is the text of each section whose level-2 heading matches re,
// running through its ### subsections to the next ## heading.
func sectionText(body string, re interface{ MatchString(string) bool }) string {
	var out []string
	in := false
	for _, l := range strings.Split(body, "\n") {
		switch {
		case strings.HasPrefix(l, "## ") || strings.HasPrefix(l, "# "):
			in = strings.HasPrefix(l, "## ") && re.MatchString(l)
		case in:
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

// pinningLines warns, per rule, when the contract mentions the phrase and
// Constraints miss some of its tests. Never a problem.
func pinningLines(root, body string, set *HeadingSet) []Line {
	rules := pinningRules(root)
	if len(rules) == 0 {
		return nil
	}
	contract := strings.ToLower(sectionText(body, heading(set.BehaviourContract)))
	constraints := sectionText(body, heading(set.Constraints))
	var out []Line
	for _, r := range rules {
		if !strings.Contains(contract, strings.ToLower(r.phrase)) {
			continue
		}
		var missing []string
		for _, n := range r.names {
			if !strings.Contains(constraints, n) {
				missing = append(missing, n)
			}
		}
		if len(missing) > 0 {
			out = append(out, Line{Warning, fmt.Sprintf("must-change list: the contract mentions %q but Constraints do not name %s (%s)",
				r.phrase, strings.Join(missing, ", "), GuidelinesFile)})
		}
	}
	return out
}
