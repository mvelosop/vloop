package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

	AwakeWarn string // why the keep-awake hold could not be taken; logged once to run.log

	SessionTimeout time.Duration // the plan session is killed after this; none when zero

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
	cmd := gitCmd(root, args...)
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

// requireCleanTree refuses when git status lists a modified, staged or
// untracked-not-ignored path, so `git add -A` cannot sweep stray files into the
// driver's commits. .vloop/tmp/ never counts; a resume also tolerates
// .vloop/state/, where the operator's task edits wait for the next iteration,
// and a fresh run its own run id's folders under .vloop/state/runs/, which a
// refusal before planning (a check failing on the base) leaves behind.
func requireCleanTree(root string, resuming bool, runID string) error {
	raw, err := gitCmd(root, "status", "--porcelain", "-z", "--untracked-files=all").Output() // untrimmed: entries start with a space
	if err != nil {
		return halt(ExitPreflight, "git status: %v", err)
	}
	out := string(raw)
	var dirty []string
	parts := strings.Split(out, "\x00")
	for i := 0; i < len(parts); i++ {
		e := parts[i]
		if len(e) < 4 {
			continue
		}
		if e[0] == 'R' || e[0] == 'C' { // the next entry is the rename's source
			i++
		}
		f := e[3:]
		if strings.HasPrefix(f, ".vloop/tmp/") || (resuming && strings.HasPrefix(f, ".vloop/state/")) ||
			(runID != "" && strings.HasPrefix(f, ".vloop/state/runs/"+runID+"/")) {
			continue
		}
		dirty = append(dirty, f)
	}
	if len(dirty) == 0 {
		return nil
	}
	sort.Strings(dirty)
	list := strings.Join(dirty, ", ")
	if len(dirty) > 10 {
		list = fmt.Sprintf("%s and %d more", strings.Join(dirty[:10], ", "), len(dirty)-10)
	}
	return halt(ExitPreflight, "the tree is not clean — commit, ignore or remove these first: %s", list)
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
		entries, err := brief.Load(root)
		if err != nil {
			return nil, halt(ExitPreflight, "%v", err)
		}
		if briefPath = brief.Newest(entries); briefPath == "" {
			return nil, halt(ExitPreflight, "no ready brief in %s/ — name one, or set status: ready on a checked brief", briefsDir)
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

	if !resuming {
		lang, err := config.Get(root, "language")
		if err != nil {
			return nil, halt(ExitPreflight, "%v", err)
		}
		if err := brief.Plannable(root, briefPath, brief.SetFor(lang.Value), p.Replan); err != nil {
			return nil, halt(ExitPreflight, "%v", err)
		}
	}
	scratch, err := config.Get(root, "run.gate-scratch")
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	for _, d := range scratch.List {
		if !GitIgnored(root, d) {
			return nil, halt(ExitPreflight, "gate scratch %s is not ignored by git — add it to .gitignore", d)
		}
	}
	checkDefs, err := config.Checks(root)
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	if len(checkDefs) == 0 {
		return nil, halt(ExitPreflight, "no check configured — add a [[check]] to .vloop/config.toml")
	}
	if err := requireCleanTree(root, resuming, brief.RunID(briefPath)); err != nil {
		return nil, err
	}

	if resuming {
		if p.PlanOnly {
			t := p.bareTerm()
			t.show("--plan-only, but %s is already planned — nothing to do", existing.RunID)
			t.show("read:   %s", planMarkdown)
			return &PlanResult{Plan: existing}, nil
		}
		if err := CheckFixtures(root, existing); err != nil {
			return nil, err
		}
		if _, err := p.workBranch(existing.RunID); err != nil {
			return nil, err
		}
		return &PlanResult{Plan: existing}, nil
	}

	runID := brief.RunID(briefPath)
	if err := p.workBranchFree(runID); err != nil {
		return nil, err
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
		Claude: p.Claude, Now: p.Now, Home: p.Home, User: p.User, Timeout: p.SessionTimeout}
	t := term{p, r}

	// Every check passes on the base before the branch is made or anything is
	// planned: one that fails would be blamed on the first task.
	shell, err := config.Get(root, "shell")
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	gt, err := config.Get(root, "run.gate-timeout")
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	gateMin, _ := strconv.Atoi(gt.Value)
	if err := t.baseChecks(planChecks(checkDefs), shell.Value, time.Duration(gateMin)*time.Minute); err != nil {
		return nil, err
	}

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
		if err := os.RemoveAll(filepath.Join(root, filepath.FromSlash(state.GatesDir))); err != nil {
			return nil, err
		}
	}

	if p.PlanOnly && p.AwakeWarn != "" {
		r.Logf("%s", AwakeWarning(p.AwakeWarn))
	}
	if hasPlan {
		t.say("the plan %s was for %s — you asked for %s: resetting and planning fresh", existing.RunID, existing.Brief, briefPath)
	}

	model, effort, err := ModelEffort(root, nil, PhasePlan)
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	t.say("planning from %s using %s", briefPath, model)
	session := func() error { return p.planSession(t, r, runDir, briefPath, model, effort) }
	if err := session(); err != nil {
		return nil, err
	}

	plan, blocked, err := p.acceptPlan(t, runID, briefPath, cur, time.Duration(gateMin)*time.Minute, session)
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
	if blocked != nil { // the plan is kept, blocked, for the operator's amendment
		return nil, blocked
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

// planSession runs one plan session and refuses what it moved or how it ended.
func (p *Planner) planSession(t term, r *Runner, runDir, briefPath, model, effort string) error {
	root := p.Root
	before := refsState(root)
	gguard := snapshotGit(root)
	res, err := r.Run(Spec{Phase: PhasePlan, Arg: briefPath, Model: model, Effort: effort})
	if err != nil {
		return halt(ExitPreflight, "planning session failed: %v", err)
	}
	if what := gguard.changed(); what != "" {
		return halt(ExitRefsMoved, "plan changed %s — nothing was committed; restore it, then re-run", what)
	}
	if moved := refsDiff(before, refsState(root)); len(moved) > 0 {
		t.warn("REFS MOVED plan — the planning session changed git refs; nothing was committed:")
		for _, l := range moved {
			t.warn("%s", l)
		}
		t.warn("restore them (git branch -m, git switch, git update-ref -d, git remote set-head), then re-run")
		return halt(ExitRefsMoved, "REFS MOVED plan — the planning session changed git refs; nothing was committed")
	}
	if res.TimedOut {
		return halt(ExitSessionError, "planning session timed out after %s — see %s", p.SessionTimeout, relRunDir(root, runDir))
	}
	if res.ExitCode != 0 {
		return halt(ExitPreflight, "planning session failed (claude exited %d) — see %s", res.ExitCode, relRunDir(root, runDir))
	}
	return nil
}

// CheckFixtures refuses a plan whose gate folder no longer matches the fixtures
// stamped on a task: the change was not recorded by vloop task verify.
func CheckFixtures(root string, plan *state.Plan) error {
	id, err := state.FixturesMismatch(root, plan)
	if err != nil {
		return halt(ExitPreflight, "%v", err)
	}
	if id != "" {
		return halt(ExitPreflight, "the gate fixtures of %s changed outside vloop task verify — record the change with vloop task verify %s --reason '<why>'", id, id)
	}
	return nil
}

// workBranch is R-2: on the default branch the run gets a branch named for its
// run id, and the name is returned; anywhere else it runs where it is and the
// result is "".
func (p *Planner) workBranch(runID string) (string, error) {
	if !OnDefaultBranch(p.Root) {
		return "", nil
	}
	if err := p.workBranchFree(runID); err != nil {
		return "", err
	}
	if _, err := git(p.Root, "switch", "-q", "-c", runID); err != nil {
		return "", halt(ExitPreflight, "cannot create branch %s: %v", runID, err)
	}
	if p.Out != nil && !p.Quiet {
		fmt.Fprintf(p.Out, "created and switched to branch %s\n", runID)
	}
	return runID, nil
}

// workBranchFree refuses when the run would branch off the default branch and
// the branch for its run id already exists.
func (p *Planner) workBranchFree(runID string) error {
	if !OnDefaultBranch(p.Root) {
		return nil
	}
	if gitCmd(p.Root, "show-ref", "--verify", "--quiet", "refs/heads/"+runID).Run() == nil {
		return halt(ExitPreflight, "branch %s exists — switch to it and re-run", runID)
	}
	return nil
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

// newRunDir creates .vloop/state/runs/<run id>/<YYYYMMDD-HHMMSS>/ (UTC) with its
// sessions/ folder, suffixing -2, -3… when a run already took that second.
func newRunDir(root, runID string, now time.Time) (string, error) {
	base := filepath.Join(root, filepath.FromSlash(runsDir), runID)
	stamp := now.UTC().Format(runFolderStamp)
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
// output is not a plan — and the problems are listed. A round is the base run
// of the gates and then the gate review; a plan that fails either goes back to
// the plan session once. The plan the gate review failed twice is saved
// blocked and returned with the halt that ends the run, for the caller to
// commit first.
func (p *Planner) acceptPlan(t term, runID, briefPath, branch string, gateTimeout time.Duration, revise func() error) (*state.Plan, *Halt, error) {
	root := p.Root
	for round := 1; ; round++ {
		plan, err := p.loadPlan(t)
		if err != nil {
			return nil, nil, err
		}
		problems, err := p.baseGates(t, plan, gateTimeout)
		if err != nil {
			return nil, nil, err
		}
		reviewed := false
		if len(problems) == 0 {
			// The review runs vloop task gate, so the plan is stamped and saved first.
			if plan, err = p.stampPlan(t, plan, runID, briefPath, branch); err != nil {
				return nil, nil, err
			}
			if problems, err = p.gateReviewRound(t, round); err != nil {
				return nil, nil, err
			}
			reviewed = true
			plan.GateReview = state.GateReview{Rounds: round, Verdict: "PASS"}
			if len(problems) > 0 {
				plan.GateReview.Verdict = "FAIL"
			}
			if err := state.Save(root, plan); err != nil {
				return nil, nil, err
			}
			if len(problems) == 0 {
				return plan, nil, nil
			}
		}
		for _, l := range problems {
			t.warn("  %s", l)
		}
		if round >= planRounds {
			if reviewed {
				plan.Status = "blocked"
				if err := state.Save(root, plan); err != nil {
					return nil, nil, err
				}
				line := gateReviewFailedLine(root, t.r.RunDir, round)
				t.r.Logf("%s", t.r.Mask(line)) // logged; the halt prints it, once
				return plan, halt(ExitBlocked, "%s", line), nil
			}
			_ = os.Remove(state.Path(root))
			_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(state.GatesDir)))
			return nil, nil, halt(ExitPreflight, "the plan is not fit (%d problem(s)) after %d rounds, nothing was committed — see above and %s", len(problems), planRounds, relRunDir(root, t.r.RunDir)+"run.log")
		}
		t.say("sending %d problem(s) back to the plan session", len(problems))
		if err := writeProblems(root, problems); err != nil {
			return nil, nil, err
		}
		err = revise()
		_ = os.Remove(filepath.Join(root, filepath.FromSlash(planProblems)))
		if err != nil {
			return nil, nil, err
		}
	}
}

// loadPlan reads and checks the plan the session wrote: a plan that fails the
// schema or the gate-shape checks is removed and halts, with no second round.
func (p *Planner) loadPlan(t term) (*state.Plan, error) {
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
		_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(state.GatesDir)))
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
	if len(rep.Problems) == 0 {
		stray, err := state.StrayGateFolders(root, &plan)
		if err != nil {
			return nil, err
		}
		for _, id := range stray {
			problems = append(problems, fmt.Sprintf("%s/%s is a gate folder for %s, which is not a task of the plan", state.GatesDir, id, id))
		}
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
	return &plan, nil
}

// stampPlan stamps an accepted plan with what the driver holds and saves it as
// running.
func (p *Planner) stampPlan(t term, plan *state.Plan, runID, briefPath, branch string) (*state.Plan, error) {
	root := p.Root

	// The driver stamps what it already holds, rather than trusting the session
	// to record it: which run, brief and branch the plan belongs to.
	plan.RunID, plan.Brief, plan.Branch = runID, briefPath, branch
	plan.Status = "running"
	// The plan's checks are the config's, whatever the session wrote.
	if defs, err := config.Checks(root); err == nil {
		plan.Checks = planChecks(defs)
	}
	if v, err := config.Get(root, "run.gate-scratch"); err == nil {
		plan.GateScratch = append([]string{}, v.List...)
	}
	// The planner may have run a gate while drafting it; acceptance leaves the
	// scratch folders empty like every other gate run.
	if err := state.EmptyScratch(root, plan.GateScratch); err != nil {
		return nil, err
	}
	if err := state.StampFixtures(root, plan); err != nil {
		return nil, err
	}
	if err := state.Save(root, plan); err != nil {
		return nil, err
	}
	t.say("planned: %s — %d tasks", runID, len(plan.Tasks))
	return plan, nil
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

// GitIgnored reports whether git ignores the repo-relative folder rel, as it
// would the files a gate leaves in it.
func GitIgnored(root, rel string) bool {
	probe := strings.TrimSuffix(rel, "/") + "/.vloop-probe"
	return gitCmd(root, "check-ignore", "-q", "--no-index", "--", probe).Run() == nil
}
