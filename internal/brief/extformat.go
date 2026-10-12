package brief

import (
	"fmt"
	"strings"
)

const externalFormatsHeading = "## External formats"

// formatRule is one "- `phrase`: `path`" line: a brief whose contract
// mentions the phrase must cite the path in Binding references.
type formatRule struct {
	phrase string
	path   string
}

// externalFormatRules reads the rules under "## External formats"; other
// lines and malformed ones are ignored. A missing file or section yields none.
func externalFormatRules(root string) []formatRule {
	var rules []formatRule
	for _, l := range guidelinesSection(root, externalFormatsHeading) {
		if r, ok := parseFormatRule(l); ok {
			rules = append(rules, r)
		}
	}
	return rules
}

func parseFormatRule(l string) (formatRule, bool) {
	rest, ok := strings.CutPrefix(l, "- `")
	if !ok {
		return formatRule{}, false
	}
	phrase, rest, ok := strings.Cut(rest, "`")
	if !ok || phrase == "" {
		return formatRule{}, false
	}
	if rest, ok = strings.CutPrefix(rest, ": `"); !ok {
		return formatRule{}, false
	}
	path, rest, ok := strings.Cut(rest, "`")
	if !ok || path == "" {
		return formatRule{}, false
	}
	if rest != "" && !strings.HasPrefix(rest, " (") {
		return formatRule{}, false
	}
	return formatRule{phrase, path}, true
}

// externalFormatLines warns, per rule, when the contract mentions the phrase
// and Binding references do not cite the path. Never a problem.
func externalFormatLines(root, body string, set *HeadingSet) []Line {
	rules := externalFormatRules(root)
	if len(rules) == 0 {
		return nil
	}
	contract := strings.ToLower(sectionText(body, heading(set.BehaviourContract)))
	refs := sectionText(body, heading(set.BindingRefs))
	var out []Line
	for _, r := range rules {
		if !strings.Contains(contract, strings.ToLower(r.phrase)) || strings.Contains(refs, r.path) {
			continue
		}
		out = append(out, Line{Warning, fmt.Sprintf("external format: the contract mentions %q but Binding references do not cite %s (%s)",
			r.phrase, r.path, GuidelinesFile)})
	}
	return out
}
