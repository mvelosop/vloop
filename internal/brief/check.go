package brief

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Marker classifies a line of a check report.
type Marker string

const (
	Pass    Marker = "✓"
	Warning Marker = "!"
	Problem Marker = "✗"
)

// Line is one rule's outcome.
type Line struct {
	Marker Marker
	Text   string
}

// Result is the outcome of checking one brief.
type Result struct {
	Path    string
	Skipped bool
	Status  string
	Lines   []Line
}

func (r *Result) count(m Marker) int {
	n := 0
	for _, l := range r.Lines {
		if l.Marker == m {
			n++
		}
	}
	return n
}

// Problems and Warnings return the messages without their marker.
func (r *Result) Problems() []string { return r.texts(Problem) }
func (r *Result) Warnings() []string { return r.texts(Warning) }

func (r *Result) texts(m Marker) []string {
	out := []string{}
	for _, l := range r.Lines {
		if l.Marker == m {
			out = append(out, l.Text)
		}
	}
	return out
}

// Failed reports whether the brief has problems.
func (r *Result) Failed() bool { return r.count(Problem) > 0 }

// Summary is the closing line of a brief's report.
func (r *Result) Summary() string {
	if p := r.count(Problem); p > 0 {
		return fmt.Sprintf("%d problem(s), %d warning(s)", p, r.count(Warning))
	}
	return fmt.Sprintf("ok, %d warning(s)", r.count(Warning))
}

var (
	fenceRe   = regexp.MustCompile("(?m)^```")
	homeRe    = regexp.MustCompile(`/Users/|/home/[a-z]|[A-Za-z]:\\+Users\\`)
	symbolRe  = regexp.MustCompile("`[a-z_]+\\.(py|ts|js)::|`[a-z_]+\\.[a-z_]+\\(\\)")
	pathRe    = regexp.MustCompile("`([A-Za-z0-9_./-]+(\\.(md|puml|json|jsonl|ya?ml|sh|py|ts|tsx|js|sql|toml|csv|svg|png)|/))`")
	keyRe     = regexp.MustCompile(`\b[A-Z][A-Z0-9]{1,9}-[0-9]+\b`)
	keyIgnore = regexp.MustCompile(`^(UTF|SHA|ISO|RFC|HTTP|MD|AES|RSA|SPDX|ASCII|CVE|ES|EC|PEP|ADR|UC|SPA)-`)
	stampRe   = regexp.MustCompile(`^[A-Z]{1,3}[0-9]{8}-[0-9]{4}$`)
	trackerRe = regexp.MustCompile(`(?i)https?://[^ )]*(linear\.app|atlassian\.net|/browse/|/issues/)`)
	headingRe = regexp.MustCompile(`^#+ `)
	bulletRe  = regexp.MustCompile(`^[-*] `)
)

// heading matches a heading line naming name, case-insensitively, with any
// trailing text.
func heading(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?im)^#+ .*` + regexp.QuoteMeta(name))
}

// Check runs every rule on a brief. root is the repo root; a brief whose
// status is not "ready" is skipped.
func Check(root string, b *Brief, set *HeadingSet) *Result {
	res := &Result{Path: b.Path, Status: b.Status}
	add := func(m Marker, format string, a ...any) {
		res.Lines = append(res.Lines, Line{m, fmt.Sprintf(format, a...)})
	}
	fmProblems := frontmatterProblems(b.Path, b.fm)
	for _, p := range fmProblems {
		add(Problem, "%s", p)
	}
	if b.Status != "ready" {
		if len(fmProblems) == 0 {
			res.Skipped = true
		}
		return res
	}
	if len(fmProblems) == 0 {
		add(Pass, "frontmatter is valid")
	}
	body := b.Body

	journal := JournalDir + "/" + RunID(b.Path) + ".md"
	if st, err := os.Stat(filepath.Join(root, filepath.FromSlash(journal))); err == nil && st.Mode().IsRegular() {
		add(Problem, "already run — %s exists; this brief declares itself plannable and is not", journal)
	}

	if heading(set.WorkedExample).MatchString(body) {
		add(Pass, "has a worked example")
	} else {
		add(Problem, "no worked example section — nothing arbitrates a disagreement")
	}

	if fenceRe.MatchString(body) {
		add(Pass, "has a fenced block with concrete values")
	} else {
		add(Problem, "no fenced code block — the worked example needs exact values, not prose")
	}

	if heading(set.OutOfScope).MatchString(body) {
		if n := sectionItems(body, heading(set.OutOfScope)); n >= 2 {
			add(Pass, "out of scope: %d item(s)", n)
		} else {
			add(Problem, "out-of-scope section has %d item(s) — name what you are NOT asking for", n)
		}
	} else {
		add(Problem, "no out-of-scope section — scope creep has nothing to be measured against")
	}

	if heading(set.Constraints).MatchString(body) {
		add(Pass, "has constraints")
	} else {
		add(Warning, "no constraints section — toolchain and limits are left to guesswork")
	}

	if regexp.MustCompile(`(?i)` + set.TaskCount).MatchString(body) {
		add(Pass, "states an expected task count")
	} else {
		add(Warning, "no expected task count — decomposition has nothing to calibrate against")
	}

	if regexp.MustCompile(`(?i)` + set.ExitCodes).MatchString(body) {
		add(Pass, "pins exit or status codes")
	} else {
		add(Warning, "no exit/status codes — is the behaviour contract precise enough to gate?")
	}

	if homeRe.MatchString(body) {
		add(Problem, "contains an absolute home path")
	} else {
		add(Pass, "no absolute paths")
	}

	if n := len(symbolRe.FindAllString(body, -1)); n > 2 {
		add(Warning, "names %d internal symbols — pinning mechanics, not decisions?", n)
	}

	res.Lines = append(res.Lines, bindingRefLines(root, body, set)...)

	dead := 0
	for _, p := range citedPaths(withoutSection(body, heading(set.BindingRefs))) {
		if !resolves(root, b.Path, p) {
			add(Warning, "path does not resolve: %s", p)
			dead++
		}
	}
	if dead == 0 {
		add(Pass, "referenced docs resolve")
	}

	if keys := trackerKeys(body); len(keys) > 0 {
		add(Warning, "names issue(s) no session can open: %s -- carry what they say, not the key", strings.Join(keys, " ")+" ")
	} else if trackerRe.MatchString(body) {
		add(Warning, "links an issue tracker -- no session downstream has web access; inline what it says")
	} else {
		add(Pass, "no references a memoryless offline session cannot follow")
	}
	return res
}

// sectionItems is the largest bullet count among the sections whose heading
// matches re, each running to the next heading.
func sectionItems(body string, re *regexp.Regexp) int {
	most, count, in := 0, 0, false
	for _, l := range strings.Split(body, "\n") {
		switch {
		case re.MatchString(l):
			in, count = true, 0
		case in && headingRe.MatchString(l):
			most, in = max(most, count), false
		case in && bulletRe.MatchString(l):
			count++
		}
	}
	return max(most, count)
}

// withoutSection drops each section whose heading line matches re, up to the
// next heading.
func withoutSection(body string, re *regexp.Regexp) string {
	var out []string
	skip := false
	for _, l := range strings.Split(body, "\n") {
		if headingRe.MatchString(l) {
			skip = re.MatchString(l)
		}
		if !skip {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func citedPaths(body string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range pathRe.FindAllStringSubmatch(body, -1) {
		p := m[1]
		if strings.HasPrefix(p, "~") || strings.HasPrefix(p, "http") || strings.Contains(p, "NNN") || strings.Contains(p, "<") {
			continue
		}
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// resolves: a repo path, a path relative to the brief, or a bare name that
// exists somewhere in the repo (a worked example may name fixture files).
func resolves(root, briefPath, p string) bool {
	rel := filepath.FromSlash(p)
	if exists(filepath.Join(root, rel)) || exists(filepath.Join(root, filepath.FromSlash(path.Dir(briefPath)), rel)) {
		return true
	}
	base := path.Base(strings.TrimSuffix(p, "/"))
	found := false
	_ = filepath.WalkDir(root, func(fp string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return filepath.SkipAll
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.Name() == base {
			found = true
		}
		return nil
	})
	return found
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func trackerKeys(body string) []string {
	seen := map[string]bool{}
	var out []string
	for _, k := range keyRe.FindAllString(body, -1) {
		if keyIgnore.MatchString(k) || stampRe.MatchString(k) || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
