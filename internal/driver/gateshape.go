package driver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mvelosop/vloop/internal/state"
)

// The plan-time gate-shape rules, ported from .loop/run.sh. A gate re-runs for
// the life of the plan, so a shape that can never pass for a correct
// implementation, or that decays after the next commit, is refused before any
// iteration is spent. None is a full shell parse, on purpose.

var (
	reSerialised = regexp.MustCompile(`JSON\.stringify\([^)]*\)\.(indexOf|includes|match|search)\(|json\.dumps\([^)]*\)\.(find|index)\(|in json\.dumps\(`)
	reSourceText = regexp.MustCompile(`(readFileSync|open)\([^)]*src/|src/[^)]*\)[[:space:]]*\.[[:space:]]*read_text\(|grep [^|]*[[:space:]]src/`)

	reInspectGrep = regexp.MustCompile(`grep[[:space:]]+(-[[:alnum:]]+[[:space:]]+)*[^[:space:]&|;-][^[:space:]&|;]*[[:space:]]+[^[:space:]&|;()"']+`)
	reInspectCat  = regexp.MustCompile(`cat[[:space:]]+[^[:space:]&|;]+`)
	reInspectTest = regexp.MustCompile(`test[[:space:]]+-f[[:space:]]+[^[:space:]&|;)]+`)
	reInspectOpen = regexp.MustCompile(`open\(\\*["'][^"']+\\*["']\)`)
	reOpenHead    = regexp.MustCompile(`^open\(\\*.`)
	reOpenTail    = regexp.MustCompile(`\\*.\)$`)

	reGitDiff   = regexp.MustCompile(`git[[:space:]]+(diff|log|rev-list)[[:space:]]+.*`)
	reGitDiffHd = regexp.MustCompile(`^git[[:space:]]+(diff|log|rev-list)[[:space:]]+`)
	reSubst     = regexp.MustCompile(`\$\([^)]*\)`)

	reHeadDiff = regexp.MustCompile(`git[[:space:]]+(diff|log|rev-list)[[:space:]]+(--[a-zA-Z-]+[[:space:]]+)*HEAD([[:space:]]|$)`)
)

// lastField is the last whitespace-separated field of s.
func lastField(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[len(f)-1]
}

// inspectedPaths are the paths a verify command reads the bytes of, as opposed
// to merely handing to a runner: grep, cat, test -f and a Python-level open.
func inspectedPaths(cmd string) []string {
	seen := map[string]bool{}
	for _, re := range []*regexp.Regexp{reInspectGrep, reInspectCat, reInspectTest} {
		for _, m := range re.FindAllString(cmd, -1) {
			if p := lastField(m); p != "" {
				seen[p] = true
			}
		}
	}
	for _, m := range reInspectOpen.FindAllString(cmd, -1) {
		m = reOpenHead.ReplaceAllString(m, "")
		m = reOpenTail.ReplaceAllString(m, "")
		if m != "" {
			seen[m] = true
		}
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// diffRef is the ref a git diff/log/rev-list in cmd is baselined against,
// other than HEAD; "" when there is none. For a range it is the left side.
func diffRef(cmd string) string {
	rest := reGitDiffHd.ReplaceAllString(reGitDiff.FindString(cmd), "")
	if rest == "" {
		return ""
	}
	ref := ""
scan:
	for _, tok := range strings.Fields(rest) {
		switch {
		case tok == "--":
			break scan
		case strings.HasPrefix(tok, "-"):
			continue
		}
		ref = tok
		break
	}
	if ref == "" {
		return ""
	}
	if strings.HasPrefix(ref, "$(") || strings.HasPrefix(ref, `"$(`) {
		ref = reSubst.FindString(rest)
	}
	ref = strings.TrimSuffix(strings.TrimPrefix(ref, `"`), `"`)
	ref = strings.TrimSuffix(strings.TrimPrefix(ref, `'`), `'`)
	if ref == "" || ref == "HEAD" {
		return ""
	}
	if i := strings.Index(ref, ".."); i >= 0 {
		return ref[:i]
	}
	return ref
}

func tracked(root, path string) bool {
	return exec.Command("git", "-C", root, "cat-file", "-e", "HEAD:"+path).Run() == nil
}

func owns(t state.Task, path string) bool {
	for _, f := range t.Files {
		if f == path {
			return true
		}
	}
	return false
}

// GateShapeProblems are the reasons the plan's gates are refused, one line
// each, each naming its task.
func GateShapeProblems(root string, p *state.Plan) []string {
	var out []string
	add := func(format string, a ...any) { out = append(out, "gate shape rejected: "+fmt.Sprintf(format, a...)) }
	for _, t := range p.Tasks {
		switch {
		case reSerialised.MatchString(t.Verify):
			add("%s  re-serialises a parsed structure and substring-matches the text; navigate to the value and assert on it", t.ID)
		case reSourceText.MatchString(t.Verify):
			add("%s  asserts on source text rather than on what the program does", t.ID)
		}
	}
	for _, t := range p.Tasks {
		for _, f := range inspectedPaths(t.Verify) {
			if tracked(root, f) && !owns(t, f) {
				add("%s  inspects %s, which it does not own; the driver reverts that file before the gate runs, so no implementation can pass; add it to the task's files or move the claim to acceptance", t.ID, f)
			}
		}
	}
	for _, t := range p.Tasks {
		if ref := diffRef(t.Verify); ref != "" {
			add("%s  diffs against %s rather than HEAD; a gate re-runs for the life of the plan and that baseline decays on the next commit", t.ID, ref)
		}
	}
	return out
}

// HeadDiffAdvisory names the tasks whose verify diffs against HEAD without
// reading VLOOP_ACTIVE_TASK or VLOOP_GATE_TASK: during another task's iteration
// the only uncommitted work in the tree is that task's, not this gate's.
func HeadDiffAdvisory(p *state.Plan) []string {
	var ids []string
	for _, t := range p.Tasks {
		if reHeadDiff.MatchString(t.Verify) &&
			!strings.Contains(t.Verify, "VLOOP_ACTIVE_TASK") && !strings.Contains(t.Verify, "VLOOP_GATE_TASK") {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

// BlindReferences are the folder references that have no index.md, README.md
// or README-*.md: a session handed one opens files until it guesses right.
func BlindReferences(root string, p *state.Plan) []string {
	var out []string
	for _, t := range p.Tasks {
		for _, r := range t.References {
			dir := filepath.Join(root, filepath.FromSlash(r.Path))
			if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
				continue
			}
			if hasEntryPoint(dir) {
				continue
			}
			out = append(out, fmt.Sprintf("%s  %s", t.ID, r.Path))
		}
	}
	return out
}

func hasEntryPoint(dir string) bool {
	for _, n := range []string{"index.md", "README.md"} {
		if fi, err := os.Stat(filepath.Join(dir, n)); err == nil && !fi.IsDir() {
			return true
		}
	}
	m, _ := filepath.Glob(filepath.Join(dir, "README-*.md"))
	return len(m) > 0
}
