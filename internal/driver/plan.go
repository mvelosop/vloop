package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/brief"
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

// Exit codes the driver ends a run with (R-3), beyond 0.
const (
	ExitPreflight  = 1 // a refusal before anything ran, or a plan that is not fit
	ExitRefsMoved  = 9 // a session moved git refs
	journalsDir    = ".vloop/state/journals"
	runsDir        = ".vloop/state/runs"
	planMarkdown   = ".vloop/state/plan.md"
	briefsDir      = "docs/briefs"
	briefSuffix    = ".loop-brief.md"
	runFolderStamp = "20060102-150405"
)

// Halt is a run that stops with an exit code. Its message is the one line the
// command prints after "vloop: ".
type Halt struct {
	Code int
	Err  error
}

func (h *Halt) Error() string { return h.Err.Error() }
func (h *Halt) Unwrap() error { return h.Err }

func halt(code int, format string, a ...any) *Halt {
	return &Halt{Code: code, Err: fmt.Errorf(format, a...)}
}

// Planner is the plan phase of one `vloop run`: it puts the run on its work
// branch, runs the plan session, checks what it wrote, and commits the plan.
type Planner struct {
	Root     string
	Version  string
	Brief    string // repo-relative with "/", and it exists; "" means the plan's own, else the newest
	PlanOnly bool
	Replan   bool
	Out, Err io.Writer // progress and warnings; may be nil
	Quiet    bool

	Claude string           // test seams, as Runner's
	Now    func() time.Time // the driver's clock
	Home   string
	User   string
}

// PlanResult says what the plan phase did.
type PlanResult struct {
	Plan    *state.Plan
	Planned bool   // this call ran the plan session and committed its plan
	RunDir  string // the run folder the plan session used; "" when this call did not plan
}

func (p *Planner) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

// git runs git in the repository and returns its trimmed stdout.
func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return strings.TrimSpace(string(out)), fmt.Errorf("git %s: %s", args[0], oneLine(msg))
		}
	}
	return strings.TrimSpace(string(out)), err
}

func oneLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// CurrentBranch is the branch HEAD is on; "HEAD" when detached.
func CurrentBranch(root string) string {
	b, err := git(root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return b
}

// OnDefaultBranch reports whether HEAD is on the default branch: the one
// origin/HEAD names, else main or master.
func OnDefaultBranch(root string) bool {
	cur, err := git(root, "symbolic-ref", "--short", "-q", "HEAD")
	if err != nil || cur == "" {
		return false
	}
	if def, err := git(root, "symbolic-ref", "--short", "-q", "refs/remotes/origin/HEAD"); err == nil && def != "" {
		return cur == strings.TrimPrefix(def, "origin/")
	}
	return cur == "main" || cur == "master"
}

// refsState is every ref and where HEAD points, one per line.
func refsState(root string) []string {
	head, err := git(root, "symbolic-ref", "-q", "HEAD")
	if err != nil || head == "" {
		head, _ = git(root, "rev-parse", "-q", "--verify", "HEAD")
	}
	lines := []string{"HEAD -> " + head}
	refs, _ := git(root, "for-each-ref", "--format=%(refname) %(objectname) %(symref)")
	if refs != "" {
		lines = append(lines, strings.Split(refs, "\n")...)
	}
	return lines
}

// refsDiff lists what differs between two refsState snapshots, "before:" and
// "after:" lines; empty when nothing moved.
func refsDiff(before, after []string) []string {
	count := func(ls []string) map[string]int {
		m := map[string]int{}
		for _, l := range ls {
			m[l]++
		}
		return m
	}
	b, a := count(before), count(after)
	var out []string
	for _, l := range before {
		if a[l] < b[l] {
			out = append(out, "     before: "+l)
		}
	}
	for _, l := range after {
		if b[l] < a[l] {
			out = append(out, "     after:  "+l)
		}
	}
	return out
}

// LatestBrief is the newest loop brief in docs/briefs/ (names carry their
// timestamp, so the last sorted is the newest); "" when there is none.
func LatestBrief(root string) string {
	m, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(briefsDir), "*"+briefSuffix))
	if len(m) == 0 {
		return ""
	}
	sort.Strings(m)
	return briefsDir + "/" + filepath.Base(m[len(m)-1])
}

type term struct {
	p *Planner
	r *Runner
}

// say prints a progress line and logs it.
func (t term) say(format string, a ...any) {
	line := t.r.Mask(fmt.Sprintf(format, a...))
	t.r.Logf("%s", line)
	if t.p.Out != nil && !t.p.Quiet {
		fmt.Fprintln(t.p.Out, line)
	}
}

// show prints a line on stdout without logging it.
func (t term) show(format string, a ...any) {
	if t.p.Out != nil && !t.p.Quiet {
		fmt.Fprintln(t.p.Out, t.r.Mask(fmt.Sprintf(format, a...)))
	}
}

// warn prints a warning on stderr and logs it.
func (t term) warn(format string, a ...any) {
	line := t.r.Mask(fmt.Sprintf(format, a...))
	t.r.Logf("%s", line)
	if t.p.Err != nil {
		fmt.Fprintln(t.p.Err, line)
	}
}

// Plan runs the plan phase. A refusal or a plan that is not fit is a *Halt; the
// working tree is changed only after every refusal that can be made without a
// session has been made.
func (p *Planner) Plan() (*PlanResult, error) {
	root := p.Root
	existing, err := state.Load(root)
	if err != nil && !errors.Is(err, state.ErrNoPlan) {
		return nil, halt(ExitPreflight, "%v", err)
	}
	hasPlan := err == nil
	cur := CurrentBranch(root)

	briefPath := p.Brief
	if briefPath == "" && !hasPlan {
		if briefPath = LatestBrief(root); briefPath == "" {
			return nil, halt(ExitPreflight, "no plan and no loop brief in %s/ — usage: vloop run %s/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md", briefsDir, briefsDir)
		}
	}

	// Naming a brief the plan does not hold means a fresh plan; naming the one
	// it holds is resuming it, whatever branch it was stamped on.
	resuming := hasPlan && (briefPath == "" || existing.Brief == briefPath)
	if briefPath != "" && !p.Replan && !resuming {
		journal := journalsDir + "/" + brief.RunID(briefPath) + ".md"
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(journal))); err == nil {
			return nil, halt(ExitPreflight, "refusing to plan: %s already exists — planning would re-derive work this brief already produced; re-plan deliberately with --replan", journal)
		}
	}
	if hasPlan && briefPath == "" && existing.Branch != "" && existing.Branch != cur {
		return nil, halt(ExitPreflight, "the plan %s was made on branch %s and you are on %s — name its brief to resume it, or another brief to plan afresh", existing.RunID, existing.Branch, cur)
	}

	if resuming {
		if p.PlanOnly {
			t := p.bareTerm()
			t.show("--plan-only, but %s is already planned — nothing to do", existing.RunID)
			t.show("read:   %s", planMarkdown)
			return &PlanResult{Plan: existing}, nil
		}
		if _, err := p.workBranch(existing.RunID); err != nil {
			return nil, err
		}
		return &PlanResult{Plan: existing}, nil
	}

	runID := brief.RunID(briefPath)
	if c, err := p.workBranch(runID); err != nil {
		return nil, err
	} else if c != "" {
		cur = c
	}

	if hasPlan { // another brief's plan: reset it
		for _, f := range []string{state.FilePath, planMarkdown} {
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(f))); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
	}

	runDir, err := newRunDir(root, runID, p.now())
	if err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(filepath.Join(runDir, "run.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer logFile.Close()
	r := &Runner{Root: root, Version: p.Version, RunID: runID, RunDir: runDir, Log: logFile,
		Claude: p.Claude, Now: p.Now, Home: p.Home, User: p.User}
	t := term{p, r}
	if hasPlan {
		t.say("the plan %s was for %s — you asked for %s: resetting and planning fresh", existing.RunID, existing.Brief, briefPath)
	}

	model, effort, err := ModelEffort(root, nil, PhasePlan)
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	t.say("planning from %s using %s", briefPath, model)
	before := refsState(root)
	res, err := r.Run(Spec{Phase: PhasePlan, Arg: briefPath, Model: model, Effort: effort})
	if err != nil {
		return nil, halt(ExitPreflight, "planning session failed: %v", err)
	}
	if moved := refsDiff(before, refsState(root)); len(moved) > 0 {
		t.warn("REFS MOVED — the planning session changed git refs; nothing was committed:")
		for _, l := range moved {
			t.warn("%s", l)
		}
		t.warn("restore them (git branch -m, git switch, git update-ref -d, git remote set-head), then re-run")
		return nil, halt(ExitRefsMoved, "REFS MOVED — the planning session changed git refs; nothing was committed")
	}
	if res.ExitCode != 0 {
		return nil, halt(ExitPreflight, "planning session failed (claude exited %d) — see %s", res.ExitCode, relRunDir(root, runDir))
	}

	plan, err := p.acceptPlan(t, runID, briefPath, cur)
	if err != nil {
		return nil, err
	}

	if err := writeJournal(root, plan); err != nil {
		return nil, err
	}
	rendered, err := state.Load(root)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(planMarkdown)), []byte(rendered.Markdown()), 0o644); err != nil {
		return nil, err
	}
	t.say("committing the plan")
	logFile.Sync()
	// .vloop/tmp/ is scratch: never committed, whether or not .gitignore lists it.
	if _, err := git(root, "add", "-A"); err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	if _, err := git(root, "reset", "-q", "--", ".vloop/tmp"); err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	if _, err := git(root, "commit", "-q", "-m", "[vloop] plan "+runID); err != nil {
		return nil, halt(ExitPreflight, "cannot commit the plan: %v", err)
	}
	if p.PlanOnly { // after the commit, so on the terminal only: the log is committed
		t.show("")
		t.show("═══ plan only ═══")
		t.show("plan:   %s — %d task(s)", runID, len(plan.Tasks))
		t.show("first:  %s", firstReady(plan))
		t.show("read:   %s   (verify commands are in %s)", planMarkdown, state.FilePath)
		t.show("adjust: vloop task set | note | drop")
		t.show("run it: vloop run")
	}
	return &PlanResult{Plan: rendered, Planned: true, RunDir: runDir}, nil
}

// workBranch is R-2: on the default branch the run gets a branch named for its
// run id, and the name is returned; anywhere else it runs where it is and the
// result is "".
func (p *Planner) workBranch(runID string) (string, error) {
	if !OnDefaultBranch(p.Root) {
		return "", nil
	}
	if exists := exec.Command("git", "-C", p.Root, "show-ref", "--verify", "--quiet", "refs/heads/"+runID).Run() == nil; exists {
		return "", halt(ExitPreflight, "branch %s exists — switch to it and re-run", runID)
	}
	if _, err := git(p.Root, "switch", "-q", "-c", runID); err != nil {
		return "", halt(ExitPreflight, "cannot create branch %s: %v", runID, err)
	}
	if p.Out != nil && !p.Quiet {
		fmt.Fprintf(p.Out, "created and switched to branch %s\n", runID)
	}
	return runID, nil
}

// bareTerm is a terminal without a run: progress goes to stdout only.
func (p *Planner) bareTerm() term {
	return term{p, &Runner{Home: p.Home, User: p.User}}
}

func firstReady(p *state.Plan) string {
	for _, t := range p.Tasks {
		if t.Status == "pending" && len(t.DependsOn) == 0 {
			return t.ID
		}
	}
	return "none"
}

func relRunDir(root, dir string) string {
	if rel, err := filepath.Rel(root, dir); err == nil {
		return filepath.ToSlash(rel) + "/"
	}
	return dir
}

// newRunDir creates .vloop/state/runs/<run id>/<YYYYMMDD-HHMMSS>/ with its
// sessions/ folder, suffixing -2, -3… when a run already took that second.
func newRunDir(root, runID string, now time.Time) (string, error) {
	base := filepath.Join(root, filepath.FromSlash(runsDir), runID)
	stamp := now.Format(runFolderStamp)
	dir := filepath.Join(base, stamp)
	for n := 2; ; n++ {
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			break
		}
		dir = filepath.Join(base, stamp+"-"+strconv.Itoa(n))
	}
	return dir, os.MkdirAll(filepath.Join(dir, "sessions"), 0o755)
}

var taskPointer = regexp.MustCompile(`^schema: /tasks/(\d+)`)

// acceptPlan checks the plan the session wrote and, when it is fit, stamps it
// and saves it as running. A plan that is not fit is removed — the session's
// output is not a plan — and the problems are listed.
func (p *Planner) acceptPlan(t term, runID, briefPath, branch string) (*state.Plan, error) {
	root := p.Root
	data, err := os.ReadFile(state.Path(root))
	if errors.Is(err, os.ErrNotExist) {
		return nil, halt(ExitPreflight, "planning produced no %s", state.FilePath)
	}
	if err != nil {
		return nil, err
	}
	reject := func(problems []string) error {
		for _, l := range problems {
			t.warn("  %s", l)
		}
		_ = os.Remove(state.Path(root))
		return halt(ExitPreflight, "the plan is not fit (%d problem(s)), nothing was committed — see above and %s", len(problems), relRunDir(root, t.r.RunDir)+"run.log")
	}

	if !json.Valid(data) {
		return nil, reject([]string{state.FilePath + " is not valid JSON"})
	}
	var areas []string
	if v, err := config.Get(root, "areas"); err == nil {
		areas = v.List
	}
	rep, err := state.CheckBytes(root, data, areas)
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	var plan state.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, reject(rep.Problems)
	}
	problems := make([]string, 0, len(rep.Problems))
	for _, pr := range rep.Problems {
		if m := taskPointer.FindStringSubmatch(pr); m != nil {
			if i, _ := strconv.Atoi(m[1]); i < len(plan.Tasks) {
				pr = plan.Tasks[i].ID + ": " + pr
			}
		}
		problems = append(problems, pr)
	}
	if len(rep.Problems) == 0 {
		problems = append(problems, GateShapeProblems(root, &plan)...)
	}
	if len(problems) > 0 {
		return nil, reject(problems)
	}
	for _, w := range rep.Warnings {
		t.warn("warning: %s", w)
	}
	for _, b := range BlindReferences(root, &plan) {
		t.warn("warning: folder reference with no index.md, README.md or README-*.md: %s", b)
	}
	if ids := HeadDiffAdvisory(&plan); len(ids) > 0 {
		t.warn("warning: %s diff against HEAD without reading VLOOP_ACTIVE_TASK or VLOOP_GATE_TASK — during another task's iteration the only uncommitted work in the tree is that task's, not this gate's", strings.Join(ids, " "))
	}

	// The driver stamps what it already holds, rather than trusting the session
	// to record it: which run, brief and branch the plan belongs to.
	plan.RunID, plan.Brief, plan.Branch = runID, briefPath, branch
	plan.Status = "running"
	if err := state.Save(root, &plan); err != nil {
		return nil, err
	}
	t.say("planned: %s — %d tasks", runID, len(plan.Tasks))
	return &plan, nil
}

// writeJournal opens the plan's journal, named for the run id so a resumed run
// keeps appending to it, and notes the plan in it.
func writeJournal(root string, plan *state.Plan) error {
	path := filepath.Join(root, filepath.FromSlash(journalsDir), plan.RunID+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(&b, "# Journal — %s\n\nAppend-only narrative of this plan. Rendered state lives in %s.\n", plan.RunID, planMarkdown)
	}
	fmt.Fprintf(&b, "\n## Plan — %s\n\n- **Brief:** `%s`\n- **Tasks:** %d\n", plan.RunID, plan.Brief, len(plan.Tasks))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(b.String())
	return err
}
