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

// Brief is a brief file read from disk.
type Brief struct {
	Path   string // repo-root-relative, slash-separated
	Body   string // the whole file
	Status string // frontmatter status; empty when absent
	fm     frontmatter
}

// RunID is the name minus the extension and ".loop-brief": it names the
// run's journal.
func RunID(path string) string {
	n := filepath.Base(filepath.FromSlash(path))
	return strings.TrimSuffix(strings.TrimSuffix(n, ".md"), ".loop-brief")
}

// Parse reads the frontmatter out of a brief's text.
func Parse(path, text string) *Brief {
	fm := parseFrontmatter(text)
	return &Brief{Path: path, Body: text, Status: fm.Status, fm: fm}
}
