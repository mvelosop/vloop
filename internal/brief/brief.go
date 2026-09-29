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
