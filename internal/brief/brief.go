// Package brief parses loop briefs and checks them against the rules of the
// configured language's heading set.
package brief

import (
	"path/filepath"
	"strings"
)

// Suffix is what makes a file a loop brief.
const Suffix = ".loop-brief.md"

// JournalDir holds one journal per run, named for the run id.
const JournalDir = ".vloop/state/journals"

// HeadingSet is the language-specific part of the rules: the section headings
// and the phrasings of the task-count and exit-code rules. Patterns are
// regular expressions matched case-insensitively against the whole brief.
type HeadingSet struct {
	BindingRefs   string
	WorkedExample string
	OutOfScope    string
	Constraints   string
	TaskCount     string
	ExitCodes     string
}

// English is the heading set for language "en".
var English = &HeadingSet{
	BindingRefs:   "Binding references",
	WorkedExample: "Worked example",
	OutOfScope:    "Out of scope",
	Constraints:   "Constraints",
	TaskCount:     `[0-9]+ (to|–|-) ?[0-9]* ?tasks|[a-z]+ to [a-z]+ tasks|[0-9]+ tasks`,
	ExitCodes:     `exit (code|[0-9])|\b(200|201|302|400|404|409|500)\b|exits? [0-9]`,
}

// sets maps a config language to its heading set.
var sets = map[string]*HeadingSet{"en": English}

// SetFor returns the heading set of a language.
func SetFor(language string) *HeadingSet {
	if s, ok := sets[language]; ok {
		return s
	}
	return English
}

// Brief is a brief file read from disk.
type Brief struct {
	Path   string // repo-root-relative, slash-separated
	Body   string // the whole file
	Status string // frontmatter status; empty when absent
}

// RunID is the name minus the extension and ".loop-brief": it names the
// run's journal.
func RunID(path string) string {
	n := filepath.Base(filepath.FromSlash(path))
	return strings.TrimSuffix(strings.TrimSuffix(n, ".md"), ".loop-brief")
}

// Parse reads the frontmatter status out of a brief's text.
func Parse(path, text string) *Brief {
	b := &Brief{Path: path, Body: text}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return b
	}
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "---" {
			break
		}
		k, v, ok := strings.Cut(l, ":")
		if !ok || strings.TrimSpace(k) != "status" {
			continue
		}
		if i := strings.Index(v, " #"); i >= 0 {
			v = v[:i]
		}
		b.Status = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return b
}
