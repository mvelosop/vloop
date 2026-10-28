package fence

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	vloop "github.com/mvelosop/vloop"
	"github.com/mvelosop/vloop/internal/driver"
)

// The phases a judge serves; the two review phases share the review fence.
const (
	PhaseReview       = "review"
	PhaseSecondReview = "second-review"
	PhaseWork         = "work"
)

// maxDepth bounds how many sh -c layers are looked through.
const maxDepth = 8

// Judge decides tool calls of one phase.
type Judge struct {
	phase  string
	root   string
	home   string
	fence  []string // Bash(…) deny rule bodies of the phase's embedded fence
	stems  []string // the adapter's deny-list stems
	skills map[string]bool
}

// Decision is a judge's answer.
type Decision struct {
	Allow  bool
	Reason string
}

func allow() Decision { return Decision{Allow: true} }

func deny(format string, a ...any) Decision {
	return Decision{Reason: fmt.Sprintf(format, a...)}
}

// New reads the phase's embedded fence, the adapter's deny-list stems and the
// embedded plugin's skill names. Relative paths resolve against root; home is
// what a leading ~ expands to ("" leaves it unexpanded).
func New(phase, root, home string) (*Judge, error) {
	if phase != PhaseReview && phase != PhaseSecondReview && phase != PhaseWork {
		return nil, fmt.Errorf("unknown phase %q: want review, second-review or work", phase)
	}
	fenceName := PhaseReview
	if phase == PhaseWork {
		fenceName = PhaseWork
	}
	raw, err := vloop.Fence(fenceName)
	if err != nil {
		return nil, err
	}
	var f struct {
		Permissions struct {
			Deny []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, err
	}
	j := &Judge{phase: phase, root: filepath.Clean(root), home: home, stems: driver.CopilotDenyStems(), skills: map[string]bool{}}
	if phase == PhaseWork { // the work session may remove files: the review list without rm:*
		var stems []string
		for _, st := range j.stems {
			if st != "rm:*" {
				stems = append(stems, st)
			}
		}
		j.stems = stems
	}
	for _, r := range f.Permissions.Deny {
		if b, ok := strings.CutPrefix(r, "Bash("); ok && strings.HasSuffix(b, ")") {
			j.fence = append(j.fence, strings.TrimSuffix(b, ")"))
		}
	}
	ents, err := fs.ReadDir(vloop.Plugin, "plugin/skills")
	if err != nil {
		return nil, err
	}
	for _, e := range ents {
		if e.IsDir() {
			j.skills[e.Name()] = true
		}
	}
	return j, nil
}

// Decide judges one call: the tool's name, the path it names (a file's, or a
// skill's name) and, for a shell tool, its command string, or for apply_patch,
// the patch.
func (j *Judge) Decide(tool, path, cmd string) Decision {
	switch strings.ToLower(tool) {
	case "bash", "powershell":
		if r := j.shell(cmd, 0); r != "" {
			return Decision{Reason: r}
		}
		return allow()
	case "create", "edit":
		if path == "" {
			return deny("%s names no path", tool)
		}
		return j.write(path)
	case "apply_patch":
		paths := PatchPaths(cmd)
		if len(paths) == 0 {
			return deny("the patch names no file")
		}
		for _, p := range paths {
			if d := j.write(p); !d.Allow {
				return d
			}
		}
		return allow()
	case "view", "read_file", "grep", "rg", "grep_search", "glob", "file_search":
		if path != "" && j.sensitive(path) {
			return deny("read of %s is not allowed: it is under ~/.claude, ~/.copilot or ~/.config/gh", path)
		}
		return allow()
	case "skill":
		if !j.skills[path] {
			return deny("skill %s is not vloop's — use /vloop:%s", path, j.skillPhase())
		}
		return allow()
	}
	return deny("tool %s is not allowed in a %s session", tool, j.phase)
}

// PatchPaths are the files an apply_patch patch adds, updates, deletes or
// moves to.
func PatchPaths(patch string) []string {
	var out []string
	for _, line := range strings.Split(patch, "\n") {
		line = strings.TrimRight(line, "\r")
		for _, p := range []string{"*** Add File: ", "*** Update File: ", "*** Delete File: ", "*** Move to: "} {
			if rest, ok := strings.CutPrefix(line, p); ok {
				out = append(out, strings.TrimSpace(rest))
			}
		}
	}
	return out
}

func (j *Judge) expand(p string) string {
	if j.home != "" && (p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`)) {
		p = filepath.Join(j.home, p[1:])
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(j.root, p)
	}
	return filepath.Clean(p)
}

func under(dir, p string) bool {
	rel, err := filepath.Rel(dir, p)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// skillPhase is the plugin skill a session of this phase runs.
func (j *Judge) skillPhase() string {
	if j.phase == PhaseWork {
		return PhaseWork
	}
	return PhaseReview
}

func (j *Judge) write(p string) Decision {
	if p == "/dev/null" {
		return allow()
	}
	r := j.expand(p)
	if j.phase == PhaseWork {
		return j.workWrite(p, r)
	}
	if rel, err := filepath.Rel(filepath.Join(j.root, ".vloop", "tmp"), r); err == nil && rel != "." &&
		rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return allow()
	}
	return deny("write to %s is not allowed in a %s session: only .vloop/tmp/ may be written", p, j.phase)
}

// workWrite decides a write of a work session (S-4): anywhere under the root
// but .git/ and .vloop/ outside .vloop/tmp/.
func (j *Judge) workWrite(p, r string) Decision {
	if !under(j.root, r) {
		return deny("write to %s is not allowed in a work session: it is outside the repository", p)
	}
	if !under(filepath.Join(j.root, ".vloop", "tmp"), r) || r == filepath.Join(j.root, ".vloop", "tmp") {
		if r == filepath.Join(j.root, ".git") || under(filepath.Join(j.root, ".git"), r) ||
			r == filepath.Join(j.root, ".vloop") || under(filepath.Join(j.root, ".vloop"), r) {
			return deny("write to %s is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/", p)
		}
	}
	return allow()
}

// sensitive reports whether a read of p is refused: p is under the user's
// global Claude, Copilot or GitHub CLI directory.
func (j *Judge) sensitive(p string) bool {
	for _, d := range []string{".claude", ".copilot", filepath.Join(".config", "gh")} {
		slash := filepath.ToSlash(d)
		if p == "~/"+slash || strings.HasPrefix(p, "~/"+slash+"/") {
			return true
		}
		if j.home != "" && under(filepath.Join(j.home, d), j.expand(p)) {
			return true
		}
	}
	return false
}

// shell judges a command string: "" when every simple command in it is
// allowed, else why not.
func (j *Judge) shell(s string, depth int) string {
	if depth > maxDepth {
		return "shell wrappers nest too deep to judge"
	}
	cmds, err := split(s)
	if err != nil {
		return err.Error()
	}
	for _, c := range cmds {
		for _, w := range c.writes {
			if d := j.write(w); !d.Allow {
				return d.Reason
			}
		}
		for _, r := range c.reads {
			if j.sensitive(r) {
				return fmt.Sprintf("read of %s is not allowed: it is under ~/.claude, ~/.copilot or ~/.config/gh", r)
			}
		}
		if r := j.words(c.words, depth); r != "" {
			return r
		}
	}
	return ""
}

// keywords are shell syntax that starts a command the splitter does not see.
var keywords = map[string]bool{
	"!": true, "{": true, "}": true, "if": true, "then": true, "else": true, "elif": true, "fi": true,
	"for": true, "while": true, "until": true, "do": true, "done": true, "case": true, "esac": true,
	"select": true, "function": true, "coproc": true, "[[": true,
}

var assignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

func baseName(w string) string {
	b := w
	if i := strings.LastIndexAny(w, `/\`); i >= 0 && i+1 < len(w) {
		b = w[i+1:]
	}
	if len(b) > 4 && strings.EqualFold(b[len(b)-4:], ".exe") {
		b = b[:len(b)-4]
	}
	return b
}

// words judges one simple command's words, looking through wrappers.
func (j *Judge) words(ws []string, depth int) string {
	for {
		for len(ws) > 0 && assignment.MatchString(ws[0]) {
			ws = ws[1:]
		}
		if len(ws) == 0 {
			return ""
		}
		w := ws[0]
		if strings.HasPrefix(w, "$") {
			return fmt.Sprintf("command word %s starts with $ and cannot be judged", w)
		}
		if w == "." {
			return ". (source) is not allowed"
		}
		name, rest := baseName(w), ws[1:]
		if keywords[name] {
			return fmt.Sprintf("shell syntax %s cannot be judged", name)
		}
		var err string
		switch name {
		case "eval", "source":
			return name + " is not allowed"
		case "env":
			rest, err = skipEnv(rest)
		case "command", "time":
			rest = skipDashes(rest)
		case "exec":
			rest = skipExec(rest)
		case "nohup":
		case "nice":
			rest = skipNice(rest)
		case "timeout":
			rest = skipTimeout(rest)
		case "stdbuf":
			rest = skipStdbuf(rest)
		case "xargs":
			rest = skipXargs(rest)
		case "sh", "bash", "zsh", "dash", "ksh":
			str, e := shellString(rest)
			if e != "" {
				return fmt.Sprintf("%s: %s", name, e)
			}
			return j.shell(str, depth+1)
		default:
			return j.match(append([]string{name}, rest...))
		}
		if err != "" {
			return err
		}
		ws = rest
	}
}

func (j *Judge) fenceLabel() string {
	if j.phase == PhaseWork {
		return PhaseWork
	}
	return PhaseReview
}

// match denies a simple command that a deny rule of the review fence or one of
// the adapter's stems covers.
func (j *Judge) match(ws []string) string {
	text := strings.Join(ws, " ")
	for _, p := range j.fence {
		if patternMatch(p, text) {
			return fmt.Sprintf("%s is denied by the %s fence (Bash(%s))", text, j.fenceLabel(), p)
		}
	}
	for _, p := range j.stems {
		if patternMatch(p, text) {
			return fmt.Sprintf("%s is denied by the Copilot deny list (%s)", text, p)
		}
	}
	return ""
}

// patternMatch is the fence's matching: `pre:*` is the prefix alone or
// followed by a space, `*` elsewhere any run of characters.
func patternMatch(p, cmd string) bool {
	if pre, ok := strings.CutSuffix(p, ":*"); ok {
		return cmd == pre || strings.HasPrefix(cmd, pre+" ")
	}
	if !strings.Contains(p, "*") {
		return cmd == p
	}
	parts := strings.Split(p, "*")
	for i, q := range parts {
		parts[i] = regexp.QuoteMeta(q)
	}
	return regexp.MustCompile("^" + strings.Join(parts, ".*") + "$").MatchString(cmd)
}

func isOption(w string) bool { return len(w) > 1 && w[0] == '-' }

func skipDashes(rest []string) []string {
	for len(rest) > 0 && isOption(rest[0]) {
		rest = rest[1:]
	}
	return rest
}

func skipEnv(rest []string) ([]string, string) {
	for len(rest) > 0 {
		w := rest[0]
		switch {
		case w == "-" || w == "-i" || w == "--ignore-environment" || w == "--" || assignment.MatchString(w):
			rest = rest[1:]
		case w == "-u" || w == "--unset":
			if len(rest) < 2 {
				return nil, "env -u has no name"
			}
			rest = rest[2:]
		case strings.HasPrefix(w, "-u") || strings.HasPrefix(w, "--unset="):
			rest = rest[1:]
		case isOption(w):
			return nil, "env option " + w + " cannot be judged"
		default:
			return rest, ""
		}
	}
	return rest, ""
}

func skipExec(rest []string) []string {
	for len(rest) > 0 && isOption(rest[0]) {
		if rest[0] == "-a" && len(rest) > 1 {
			rest = rest[1:]
		}
		rest = rest[1:]
	}
	return rest
}

func skipNice(rest []string) []string {
	for len(rest) > 0 && isOption(rest[0]) {
		w := rest[0]
		rest = rest[1:]
		if (w == "-n" || w == "--adjustment") && len(rest) > 0 {
			rest = rest[1:]
		}
	}
	return rest
}

func skipTimeout(rest []string) []string {
	for len(rest) > 0 && isOption(rest[0]) {
		w := rest[0]
		rest = rest[1:]
		if (w == "-s" || w == "-k" || w == "--signal" || w == "--kill-after") && len(rest) > 0 {
			rest = rest[1:]
		}
	}
	if len(rest) > 0 { // the duration
		rest = rest[1:]
	}
	return rest
}

func skipStdbuf(rest []string) []string {
	for len(rest) > 0 && isOption(rest[0]) {
		w := rest[0]
		rest = rest[1:]
		if (w == "-i" || w == "-o" || w == "-e") && len(rest) > 0 {
			rest = rest[1:]
		}
	}
	return rest
}

var xargsLongArgs = map[string]bool{
	"--max-args": true, "--max-procs": true, "--delimiter": true, "--arg-file": true,
	"--eof": true, "--max-lines": true, "--max-chars": true,
}

func skipXargs(rest []string) []string {
	for len(rest) > 0 && isOption(rest[0]) {
		o := rest[0]
		rest = rest[1:]
		if o == "--" {
			break
		}
		if strings.HasPrefix(o, "--") {
			if xargsLongArgs[o] && len(rest) > 0 {
				rest = rest[1:]
			}
			continue
		}
		for k := 1; k < len(o); k++ {
			if strings.IndexByte("adEILnPs", o[k]) >= 0 {
				if k == len(o)-1 && len(rest) > 0 {
					rest = rest[1:]
				}
				break
			}
			if strings.IndexByte("eil", o[k]) >= 0 {
				break
			}
		}
	}
	return rest
}

// shellString finds the -c string of a shell invocation's arguments.
func shellString(rest []string) (string, string) {
	for i := 0; i < len(rest); i++ {
		o := rest[i]
		switch {
		case o == "-o" || o == "+o":
			i++
		case o != "" && o[0] == '-' && !strings.HasPrefix(o, "--") && strings.Contains(o, "c"):
			if i+1 >= len(rest) {
				return "", "-c has no command string"
			}
			if strings.Contains(rest[i+1], "$") {
				return "", "the -c string holds $ and cannot be judged"
			}
			return rest[i+1], ""
		case isOption(o) || (len(o) > 1 && o[0] == '+'):
		default:
			return "", "runs a script or stdin, not a literal -c string"
		}
	}
	return "", "runs a script or stdin, not a literal -c string"
}
