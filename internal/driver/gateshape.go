package driver

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mvelosop/vloop/internal/config"
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

// PlatformSpecificCommands are native commands a mixed-platform gate may not
// assume exist with the same meaning on every declared platform.
var PlatformSpecificCommands = []string{
	"awk", "basename", "bash", "cat", "chmod", "cmd", "cmp", "cp", "cut",
	"date", "diff", "dirname", "env", "fc", "find", "findstr", "grep", "head",
	"ln", "ls", "mkdir", "mktemp", "mv", "powershell", "printf", "readlink",
	"rm", "robocopy", "sed", "seq", "sh", "sleep", "sort", "stat", "tail",
	"tee", "test", "touch", "tr", "uniq", "wc", "which", "xargs", "xcopy",
	"zsh",
}

// FirstPlatformSpecificCommand returns the first listed command word in s.
func FirstPlatformSpecificCommand(s string) string {
	commandPosition := true
	afterCallOperator := false
	for i := 0; i < len(s); {
		switch s[i] {
		case ' ', '\t', '\r':
			i++
			continue
		case '&':
			afterCallOperator = (i == 0 || s[i-1] != '&') && (i+1 == len(s) || s[i+1] != '&')
			commandPosition = true
			i++
			continue
		case '\n', ';', '|', '{', '(':
			afterCallOperator = false
			commandPosition = true
			i++
			continue
		}

		start := i
		quote := byte(0)
		whollyQuoted := false
		for i < len(s) {
			c := s[i]
			if quote == 0 {
				switch c {
				case ' ', '\t', '\r', '\n', ';', '|', '&', '{', '(':
					goto token
				case '\'', '"':
					quote = c
					whollyQuoted = i == start
					i++
					continue
				}
			} else {
				if c == quote {
					quote = 0
				} else if c == '\\' && quote == '"' && i+1 < len(s) {
					i++
				}
				i++
				continue
			}
			i++
		}
	token:
		word := s[start:i]
		if commandPosition {
			commandPosition = false
			if len(word) >= 2 && whollyQuoted && afterCallOperator && (word[0] == word[len(word)-1]) {
				word = word[1 : len(word)-1]
			}
			word = strings.TrimSuffix(word, `.exe`)
			if strings.HasSuffix(word, `.EXE`) {
				word = strings.TrimSuffix(word, `.EXE`)
			}
			if word != "" && !strings.ContainsAny(word[:1], "$-@[0123456789") {
				base := filepath.Base(strings.ReplaceAll(word, `\`, `/`))
				base = strings.TrimSuffix(strings.ToLower(base), ".exe")
				for _, name := range PlatformSpecificCommands {
					if base == name {
						return name
					}
				}
			}
			afterCallOperator = false
		}
	}
	return ""
}

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
		if ref := diffRef(t.Verify); ref != "" {
			add("%s  diffs against %s rather than HEAD; a gate re-runs for the life of the plan and that baseline decays on the next commit", t.ID, ref)
		}
	}
	platforms, err := config.Get(root, "platforms")
	if err != nil || !config.MixedPlatforms(platforms.List) {
		return out
	}
	platformText := config.FormatPlatforms(platforms.List)
	for _, t := range p.Tasks {
		if command := FirstPlatformSpecificCommand(t.Verify); command != "" {
			add("%s  calls %s, which is not on every platform in platforms (%s); use a pwsh cmdlet or the project's own toolchain", t.ID, command, platformText)
		}
		gateDir := filepath.Join(root, filepath.FromSlash(state.GateFolder(t.ID)))
		_ = filepath.WalkDir(gateDir, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".ps1") {
				return nil
			}
			raw, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			if command := FirstPlatformSpecificCommand(string(raw)); command != "" {
				rel, relErr := filepath.Rel(root, path)
				if relErr != nil {
					rel = path
				}
				add("%s  %s calls %s, which is not on every platform in platforms (%s); use a pwsh cmdlet or the project's own toolchain", t.ID, filepath.ToSlash(rel), command, platformText)
			}
			return nil
		})
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
