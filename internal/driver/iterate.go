package driver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/schema"
	"github.com/mvelosop/vloop/internal/state"
)

// Exit codes of the iterating phase (R-3).
const (
	ExitBlocked       = 2 // tasks remain but none can run
	ExitStalled       = 3 // iterations in a row closed nothing and charged no attempt
	ExitMaxIter       = 4 // the iteration budget is spent
	ExitNotConverging = 5 // too many iterations per closed task
	ExitCostCeiling   = 6 // the run's sessions cost the ceiling
	ExitSessionError  = 7 // a session failed to run
	ExitRepeatBlocked = 8 // a task blocked twice with nothing changed
)

// errSessionError is a session that failed to run; the run halts, resumable.
var errSessionError = errors.New("session failed")

const tmpDir = ".vloop/tmp"

// Outcome names of an iteration (iteration/v1).
const (
	OutDone       = "done"
	OutGateFailed = "gate_failed"
	OutRejected   = "rejected"
	OutBlocked    = "blocked"
)

// Ending says how a run ended: the plan status it leaves and the exit code.
type Ending struct {
	Status string
	Code   int
	Note   string // what the operator reads first; printed with the run's end
}

// Iterator is the iterating phase of one `vloop run`: it works the plan's
// tasks one iteration at a time, each ending in exactly one commit.
type Iterator struct {
	Root     string
	Version  string
	RunDir   string // the run folder (absolute), already created
	Budgets  Budgets
	Out, Err io.Writer // progress and warnings; may be nil
	Quiet    bool

	Claude string // test seams, as Runner's
	Now    func() time.Time
	Home   string
	User   string
	Env    []string

	GateTimeout    time.Duration // test seam: replaces Budgets.GateTimeout when not zero
	SessionTimeout time.Duration // test seam: replaces Budgets.SessionTimeout when not zero

	plan       *state.Plan
	resolved   *Resolved              // model and effort defaults, read once at the start
	spent      float64                // what this run's sessions cost, summed as they finish
	noProposal []string               // tasks whose session died after changing files, with the account
	gatePasses []string               // tasks reported blocked whose own gate passes
	blockedAt  map[string]blockedMark // per task, the last iteration it ended blocked in
	r          *Runner
	log        *lazyLog
	branch     string
	snap       []byte // the metrics snapshot waiting for the next commit
	snapPath   string
}

// blockedMark is where a task last ended blocked: HEAD before that iteration's
// commit, and what the session said.
type blockedMark struct{ head, summary string }

// lazyLog holds the run log in memory and writes it to the run folder just
// before each commit, so that nothing tracked is modified while a session or a
// gate is running: a gate may switch branches, and a dirty tracked file would
// stop it.
type lazyLog struct {
	path string
	buf  bytes.Buffer
}

func (l *lazyLog) Write(p []byte) (int, error) { return l.buf.Write(p) }

func (l *lazyLog) flush() error {
	if l.buf.Len() == 0 {
		return nil
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(l.buf.Bytes())
	l.buf.Reset()
	return err
}

func (it *Iterator) now() time.Time {
	if it.Now != nil {
		return it.Now()
	}
	return time.Now()
}

func (it *Iterator) say(format string, a ...any) {
	line := it.r.Mask(fmt.Sprintf(format, a...))
	it.r.Logf("%s", line)
	if it.Out != nil && !it.Quiet {
		fmt.Fprintln(it.Out, line)
	}
}

func (it *Iterator) warn(format string, a ...any) {
	line := it.r.Mask(fmt.Sprintf(format, a...))
	it.r.Logf("%s", line)
	if it.Err != nil {
		fmt.Fprintln(it.Err, line)
	}
}

// Run works the plan under Root until it ends. A run that cannot continue
// because of a refusal is a *Halt; every other ending is the returned Ending.
func (it *Iterator) Run() (Ending, error) {
	plan, err := state.Load(it.Root)
	if err != nil {
		return Ending{}, halt(ExitPreflight, "%v", err)
	}
	if err := checkPlanRunID(plan); err != nil {
		return Ending{}, err
	}
	it.plan = plan
	it.branch = CurrentBranch(it.Root)
	if it.RunDir == "" {
		if it.RunDir, err = newRunDir(it.Root, plan.RunID, it.now()); err != nil {
			return Ending{}, err
		}
	}
	it.log = &lazyLog{path: filepath.Join(it.RunDir, "run.log")}
	it.r = &Runner{Root: it.Root, Version: it.Version, RunID: plan.RunID, RunDir: it.RunDir, Log: it.log,
		Claude: it.Claude, Now: it.Now, Home: it.Home, User: it.User, Env: it.Env,
		Timeout: it.sessionTimeout()}

	if it.resolved, err = ResolveRun(it.Root); err != nil {
		return Ending{}, halt(ExitPreflight, "%v", err)
	}
	// The one read of the records on disk: the sessions before this phase, such
	// as the plan session. Every later session is added as it finishes.
	it.spent = it.spend()

	if plan.Status != "running" || plan.Branch != it.branch {
		plan.Status, plan.Branch = "running", it.branch
		if err := it.save(); err != nil {
			return Ending{}, err
		}
	}

	runIters, stalls := 0, 0
	it.blockedAt = map[string]blockedMark{}
	end := Ending{Status: "halted", Code: ExitMaxIter}
	for {
		done, blocked, pending := counts(plan)
		if pending == 0 {
			if blocked > 0 {
				end = Ending{Status: "blocked", Code: ExitBlocked}
			} else {
				end = Ending{Status: "complete"}
			}
			break
		}
		// Budgets are checked here, between iterations and never inside one, so
		// a run always stops with the plan coherent and raising one resumes.
		if runIters >= it.Budgets.MaxIterations {
			end = Ending{Status: "halted", Code: ExitMaxIter}
			break
		}
		if spend := it.spent; spend >= it.Budgets.CostCeiling {
			it.warn("cost ceiling reached: $%.2f spent this run, ceiling $%.2f", spend, it.Budgets.CostCeiling)
			end = Ending{Status: "halted", Code: ExitCostCeiling}
			break
		}
		if runIters >= it.Budgets.ConvergenceMin &&
			(done == 0 || float64(runIters)/float64(done) > it.Budgets.ConvergenceMax) {
			it.warn("not converging: %d iteration(s) this run for %d closed task(s), over %.2f per closed task", runIters, done, it.Budgets.ConvergenceMax)
			end = Ending{Status: "halted", Code: ExitNotConverging}
			break
		}
		task := nextReady(plan)
		if task == nil {
			it.warn("%d task(s) pending but none are ready — dependencies cannot be satisfied", pending)
			end = Ending{Status: "blocked", Code: ExitBlocked}
			break
		}
		runIters++
		res, err := it.iterate(task, runIters, done, len(plan.Tasks))
		if errors.Is(err, errSessionError) {
			it.warn("a work session for %s failed to run — no attempt charged; see %s", task.ID, relRunDir(it.Root, it.RunDir))
			end = Ending{Status: "halted", Code: ExitSessionError}
			break
		}
		if err != nil {
			return Ending{}, err
		}
		if res.repeat != "" {
			it.warn("   %s", res.repeat)
			end = Ending{Status: "halted", Code: ExitRepeatBlocked, Note: res.repeat}
			break
		}
		// An iteration that closed nothing and charged no attempt made no
		// recorded progress at all.
		if nd, _, _ := counts(plan); nd <= done && res.outcome != OutGateFailed && res.outcome != OutRejected {
			stalls++
			it.warn("   no recorded progress (%d/%d)", stalls, it.Budgets.StallLimit)
			if stalls >= it.Budgets.StallLimit {
				end = Ending{Status: "stalled", Code: ExitStalled}
				break
			}
		} else {
			stalls = 0
		}
	}
	if err := it.finish(end, runIters); err != nil {
		return Ending{}, err
	}
	it.summary()
	return end, nil
}

func counts(p *state.Plan) (done, blocked, pending int) {
	for _, t := range p.Tasks {
		switch t.Status {
		case "done":
			done++
		case "blocked":
			blocked++
		default:
			pending++
		}
	}
	return
}

// nextReady is the first pending task, in plan order, whose dependencies are
// all done.
func nextReady(p *state.Plan) *state.Task {
	done := map[string]bool{}
	for _, t := range p.Tasks {
		if t.Status == "done" {
			done[t.ID] = true
		}
	}
	for i := range p.Tasks {
		t := &p.Tasks[i]
		if t.Status != "pending" {
			continue
		}
		ready := true
		for _, d := range t.DependsOn {
			ready = ready && done[d]
		}
		if ready {
			return t
		}
	}
	return nil
}

func (it *Iterator) save() error { return state.Save(it.Root, it.plan) }

// readReport reads a session's handoff file, which counts as absent unless it
// exists and passes its schema.
func (it *Iterator) readReport(name, schemaName string) ([]byte, bool) {
	data, err := readHandoff(filepath.Join(it.Root, filepath.FromSlash(tmpDir), name))
	if err != nil {
		return nil, false
	}
	if v, err := schema.Validate(schemaName, data); err != nil || len(v) > 0 {
		return nil, false
	}
	return data, true
}

// restoreInputs puts back what the session changed among the driver's inputs,
// saying so in the run log, and notes a gate refused for a changed plan. It
// reports whether anything had to be restored. keep is the session's own record.
func (it *Iterator) restoreInputs(g inputGuard, phase, keep string) bool {
	changed := g.restore(keep)
	for _, p := range changed {
		it.warn("vloop: %s session changed %s — restored", phase, p)
	}
	refused := filepath.Join(it.Root, filepath.FromSlash(GateRefusedFile))
	if _, err := os.Stat(refused); err == nil {
		it.warn("vloop: %s session ran vloop task gate against a changed plan — refused", phase)
		_ = os.Remove(refused)
	}
	return len(changed) > 0
}

// gitChanged halts when .git/config or the hooks differ from the guard.
func (it *Iterator) gitChanged(g gitGuard, phase string) error {
	what := g.changed()
	if what == "" {
		return nil
	}
	return it.haltGit("%s changed %s — nothing was committed; restore it, then re-run", phase, what)
}

// haltGit ends the run with exit 9, committing nothing.
func (it *Iterator) haltGit(format string, a ...any) error {
	err := halt(ExitRefsMoved, format, a...)
	it.warn("%v", err)
	it.plan.Status = "halted"
	_ = it.save()
	_ = it.log.flush()
	return err
}

func (it *Iterator) gitRefsMoved(before []string, task, phase string) error {
	moved := refsDiff(before, refsState(it.Root))
	if len(moved) == 0 {
		return nil
	}
	it.warn("REFS MOVED %s — the %s session changed git refs; halting before anything is committed", task, phase)
	for _, l := range moved {
		it.warn("%s", l)
	}
	// The plan says why the run stopped, but nothing is committed: HEAD may be
	// on another branch now.
	it.plan.Status = "halted"
	_ = it.save()
	_ = it.log.flush()
	return halt(ExitRefsMoved, "REFS MOVED %s — a %s session changed git refs; nothing was committed", task, phase)
}

// spend is what the records in the run folder say its sessions cost.
func (it *Iterator) spend() float64 {
	files, _ := filepath.Glob(filepath.Join(it.RunDir, "sessions", "*.json"))
	total := 0.0
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var rec struct {
			Cost float64 `json:"cost_usd"`
		}
		if json.Unmarshal(data, &rec) == nil {
			total += rec.Cost
		}
	}
	return total
}

// iterResult is what the run loop needs of an iteration beyond its commit.
type iterResult struct {
	outcome string
	repeat  string // set when the task blocked twice with nothing changed
}

func (it *Iterator) iterate(task *state.Task, runIters, done, total int) (iterResult, error) {
	root := it.Root
	plan := it.plan
	iter := plan.Iteration + 1
	started := it.now().UTC()
	id := task.ID
	it.say("── iteration %d (%d this run) · %s · %d/%d done ──", iter, runIters, id, done, total)
	it.say("   %s", task.Title)

	// A stale handoff from an earlier session is never this session's report.
	for _, f := range []string{"proposal.json", "verdict.json"} {
		_ = os.Remove(filepath.Join(root, filepath.FromSlash(tmpDir), f))
	}

	// 1. work session
	model, effort := it.resolved.For(task, PhaseWork)
	before := refsState(root)
	guard := snapshotState(root)
	gguard := snapshotGit(root)
	inputs := snapshotInputs(root, it.RunDir)
	_ = os.Remove(filepath.Join(root, filepath.FromSlash(GateRefusedFile)))
	wres, err := it.r.Run(Spec{Phase: PhaseWork, Iteration: iter, Arg: id, Model: model, Effort: effort, PlanSHA: planHash(guard.pre)})
	it.spent += wres.Cost
	if err != nil {
		return iterResult{}, err
	}
	it.restoreInputs(inputs, PhaseWork, wres.Path)
	if err := it.gitChanged(gguard, PhaseWork); err != nil {
		return iterResult{}, err
	}
	if err := it.gitRefsMoved(before, id, PhaseWork); err != nil {
		return iterResult{}, err
	}

	outcome, summary, notes := OutBlocked, "", "none"
	var proposalFiles []string
	var dispute string
	data, ok := it.readReport("proposal.json", "proposal/v1")
	if wres.TimedOut {
		it.warn("   the work session for %s timed out after %s", id, it.sessionTimeout())
		return iterResult{}, errSessionError
	}
	if !ok && (wres.ExitCode != 0 || wres.IsError) {
		// An infrastructure failure, not the task's: no attempt is charged.
		return iterResult{}, errSessionError
	}
	if !ok {
		it.warn("work session left no valid proposal")
		changed := it.sessionTreeChanges()
		if len(changed) > 0 {
			summary = fmt.Sprintf("work session left no valid proposal; before dying it changed %s — the driver committed them under this iteration", strings.Join(changed, ", "))
		} else {
			summary = "work session left no valid proposal; the working tree is unchanged"
		}
		summary = it.r.Mask(summary)
		if len(changed) > 0 {
			it.noProposal = append(it.noProposal, id+" — "+summary)
		}
	} else {
		var p struct {
			Outcome string   `json:"outcome"`
			Summary string   `json:"summary"`
			Files   []string `json:"files"`
			Notes   string   `json:"notes"`
			Dispute *struct {
				Reason   string `json:"reason"`
				Evidence string `json:"evidence"`
			} `json:"gate_dispute"`
		}
		_ = json.Unmarshal(data, &p)
		outcome, summary, notes, proposalFiles = p.Outcome, it.r.Mask(p.Summary), it.r.Mask(p.Notes), p.Files
		if p.Dispute != nil {
			dispute = it.r.Mask(fmt.Sprintf("gate disputed: %s — %s", p.Dispute.Reason, p.Dispute.Evidence))
		}
	}

	// A session does not rewrite the file its own gate runs. Restored from
	// HEAD before any gate runs, and the work is not reviewable.
	tampered := ""
	if guard.restoreIfTouched() {
		tampered = it.r.Mask(tamperNote(PhaseWork))
		it.warn("   STATE TAMPERING %s — %s was modified; restored, iteration failed", id, state.FilePath)
	}
	if moved := it.gateFilesMoved(task); len(moved) > 0 {
		note := it.restoreGateFiles(id, moved)
		if tampered != "" {
			note = tampered + "; " + note
		}
		tampered = note
	}

	// 2. gates: every done task, plus this one if it claims done or blocked
	targets := []string{}
	for _, t := range plan.Tasks {
		if t.Status == "done" {
			targets = append(targets, t.ID)
		}
	}
	if dispute == "" && (outcome == OutDone || outcome == OutBlocked) {
		targets = append(targets, id)
	}
	var own *gateResult
	failed := map[string]bool{}
	for _, gid := range targets {
		g, err := it.runGateRetry(iter, id, gid)
		if err != nil {
			return iterResult{}, err
		}
		if gid == id {
			own = g
		}
		if g.exit != 0 {
			failed[gid] = true
		}
	}
	for _, gid := range targets {
		if gid == id || !failed[gid] {
			continue
		}
		it.warn("   GATE REGRESSION %s — reverting to pending; detected during %s's iteration", gid, id)
		t := plan.Find(gid)
		t.Status, t.Attempts = "pending", t.Attempts+1
		t.Notes = fmt.Sprintf("regressed: verify failed during %s — see %s", id, it.relGate(gid))
	}

	// 3. review
	verdict := "skipped"
	var findings []string
	gatePassed := outcome == OutBlocked && dispute == "" && own != nil && own.exit == 0
	switch {
	case dispute != "":
		it.warn("   GATE DISPUTE %s — blocked for the operator; no review, no attempt charged", id)
	case tampered != "":
		outcome = OutGateFailed
		it.warn("   %s failed on a rewritten gate — the work is reverted whatever the gate said", id)
	case outcome == OutBlocked:
		it.say("   work session reported blocked")
		if gatePassed {
			it.warn("   %s", gatePassesLine(id))
			it.gatePasses = append(it.gatePasses, id)
		}
	case failed[id]:
		outcome = OutGateFailed
		it.warn("   GATE FAIL %s — review skipped, work that fails its own gate is not reviewable", id)
	default:
		model, effort := it.resolved.For(task, PhaseReview)
		before := refsState(root)
		guard := snapshotState(root)
		gguard := snapshotGit(root)
		inputs := snapshotInputs(root, it.RunDir)
		tree := snapshotTree(root)
		_ = os.Remove(filepath.Join(root, filepath.FromSlash(GateRefusedFile)))
		rres, err := it.r.Run(Spec{Phase: PhaseReview, Iteration: iter, Arg: id, Model: model, Effort: effort, PlanSHA: planHash(guard.pre)})
		it.spent += rres.Cost
		if err != nil {
			return iterResult{}, err
		}
		inputsChanged := it.restoreInputs(inputs, PhaseReview, rres.Path)
		if err := it.gitChanged(gguard, PhaseReview); err != nil {
			return iterResult{}, err
		}
		if err := it.gitRefsMoved(before, id, PhaseReview); err != nil {
			return iterResult{}, err
		}
		treeChanged := tree.revert()
		for _, p := range treeChanged {
			it.warn("vloop: the review session changed %s — reverted", p)
		}
		reviewTampered := guard.restoreIfTouched()
		if reviewTampered {
			it.warn("   STATE TAMPERING %s — review session modified %s; restored", id, state.FilePath)
		}
		if rres.TimedOut {
			it.warn("   the review session for %s timed out after %s", id, it.sessionTimeout())
			return iterResult{}, errSessionError
		}
		if data, ok := it.readReport("verdict.json", "verdict/v1"); !ok {
			it.warn("   review session left no valid verdict — treating as FAIL")
			verdict = "FAIL"
		} else {
			var v struct {
				Verdict  string `json:"verdict"`
				Findings []struct {
					Summary string `json:"summary"`
				} `json:"findings"`
			}
			_ = json.Unmarshal(data, &v)
			verdict = v.Verdict
			for _, f := range v.Findings {
				findings = append(findings, f.Summary)
			}
		}
		if verdict != "PASS" {
			outcome = OutRejected
		}
		if reviewTampered {
			outcome, verdict = OutRejected, "FAIL"
			findings = append(findings, it.r.Mask(tamperNote(PhaseReview)))
		}
		if inputsChanged {
			outcome, verdict = OutRejected, "FAIL"
			findings = append(findings, "review session changed the driver's inputs — restored by the driver; a review judges the work and changes nothing else")
		}
		if len(treeChanged) > 0 {
			outcome, verdict = OutRejected, "FAIL"
			findings = append(findings, "the review session changed files")
		}
		it.say("   review: %s", verdict)
	}

	// 4. apply: the driver makes every status transition
	attempt := task.Attempts + 1
	switch {
	case dispute != "":
		outcome = OutBlocked
		task.Status, task.Notes = "blocked", dispute
		it.say("   %s blocked — the operator resolves the disputed gate with vloop task verify", id)
	default:
		it.applyOutcome(task, outcome, summary, findings, tampered, gatePassed)
	}
	repeat := it.repeatBlocked(id, outcome, summary)
	if task.Status == "pending" && task.Attempts >= it.Budgets.MaxAttempts {
		for i := range plan.Tasks {
			if t := &plan.Tasks[i]; t.Status == "pending" && t.Attempts >= it.Budgets.MaxAttempts {
				t.Status = "blocked"
			}
		}
		it.warn("   %s hit the attempt ceiling — blocked", id)
	}
	plan.Iteration = iter
	if err := it.save(); err != nil {
		return iterResult{}, err
	}

	// 5. record
	rec := map[string]any{
		"schema": "iteration/v1", "run_id": plan.RunID, "iteration": iter, "task": id, "attempt": attempt,
		"outcome": outcome, "gate": nil,
		"started": started.Format(time.RFC3339), "ended": it.now().UTC().Format(time.RFC3339),
	}
	if own != nil {
		g := map[string]any{"exit": own.exit, "duration_ms": own.ms}
		if own.flaky {
			g["flaky"] = true
		}
		rec["gate"] = g
	}
	if err := it.appendIteration(rec); err != nil {
		return iterResult{}, err
	}
	it.copyReport("proposal.json", fmt.Sprintf("%03d-proposal.json", iter))
	it.copyReport("verdict.json", fmt.Sprintf("%03d-verdict.json", iter))
	if err := it.appendJournal(task, outcome, verdict, summary, notes, proposalFiles); err != nil {
		return iterResult{}, err
	}
	if err := it.render(); err != nil {
		return iterResult{}, err
	}

	// 6. commit: one per iteration
	if err := it.commit(fmt.Sprintf("[vloop] %s: %s", id, outcome)); err != nil {
		return iterResult{outcome: outcome, repeat: repeat}, err
	}
	it.snapshot()
	return iterResult{outcome: outcome, repeat: repeat}, nil
}

// repeatBlocked notes where a task ended blocked and says so when it is the
// second time in a row with nothing outside .vloop/state/ changed in between: a
// memoryless session given identical inputs reaches an identical conclusion, so
// a third attempt cannot carry new information. It names the first diagnosis.
// HEAD is read before this iteration's commit, so the diff from the earlier
// mark is exactly what changed between the starts of the two sessions.
func (it *Iterator) repeatBlocked(id, outcome, summary string) string {
	if outcome != OutBlocked {
		delete(it.blockedAt, id)
		return ""
	}
	head, _ := git(it.Root, "rev-parse", "HEAD")
	prev, seen := it.blockedAt[id]
	it.blockedAt[id] = blockedMark{head: head, summary: summary}
	if !seen {
		return ""
	}
	out, err := git(it.Root, "diff", "--name-only", prev.head, head)
	if err != nil {
		return ""
	}
	for _, f := range strings.Split(out, "\n") {
		if f != "" && !strings.HasPrefix(f, ".vloop/state/") {
			return ""
		}
	}
	first := prev.summary
	if first == "" {
		first = "none"
	}
	return fmt.Sprintf("%s blocked twice with nothing changed since the first attempt — halting rather than spending a third identical session. First diagnosis: %s", id, first)
}

// sessionTreeChanges is what a work session that died without a proposal left
// in the working tree, minus the driver's own bookkeeping.
func (it *Iterator) sessionTreeChanges() []string {
	out, err := git(it.Root, "status", "--porcelain", "--untracked-files=all")
	if err != nil || out == "" {
		return nil
	}
	var changed []string
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 {
			continue
		}
		p := line[3:]
		if i := strings.Index(p, " -> "); i >= 0 {
			p = p[i+4:]
		}
		p = strings.Trim(p, `"`)
		if strings.HasPrefix(p, ".vloop/state/") || strings.HasPrefix(p, ".vloop/tmp/") {
			continue
		}
		changed = append(changed, p)
	}
	return changed
}

type gateResult struct {
	exit  int
	ms    int64
	flaky bool // failed, then passed on the immediate re-run
}

func (it *Iterator) relGate(id string) string {
	return relRunDir(it.Root, it.RunDir) + "gates/" + id + ".log"
}

// gateTimeout is how long a gate may run, none when it is zero.
func (it *Iterator) gateTimeout() time.Duration {
	if it.GateTimeout != 0 {
		return it.GateTimeout
	}
	return time.Duration(it.Budgets.GateTimeout) * time.Minute
}

// sessionTimeout is how long a session may run, none when it is zero.
func (it *Iterator) sessionTimeout() time.Duration {
	if it.SessionTimeout != 0 {
		return it.SessionTimeout
	}
	return time.Duration(it.Budgets.SessionTimeout) * time.Minute
}

// runGate runs one task's verify in the plan's shell from the repo root, with
// the two ids in its environment and nowhere else. Its masked output is the
// task's gate log, and a failing one is also kept for this iteration.
func (it *Iterator) runGate(iter int, active, id string) (*gateResult, error) {
	t := it.plan.Find(id)
	base := it.Env
	if base == nil {
		base = os.Environ()
	}
	var env []string
	for _, kv := range base {
		if strings.HasPrefix(kv, "VLOOP_ACTIVE_TASK=") || strings.HasPrefix(kv, "VLOOP_GATE_TASK=") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env, "VLOOP_ACTIVE_TASK="+active, "VLOOP_GATE_TASK="+id)
	cmd, err := state.GateCommand(it.Root, it.plan.Shell, t.Verify, env)
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	refsBefore := refsState(it.Root)
	gguard := snapshotGit(it.Root)
	start := time.Now()
	timedOut, runErr := state.RunGroup(cmd, it.gateTimeout())
	d := time.Since(start)
	if err := it.gitChanged(gguard, "gate"); err != nil {
		return nil, err
	}
	if len(refsDiff(refsBefore, refsState(it.Root))) > 0 {
		return nil, it.haltGit("the gate of %s moved git refs — nothing was committed; restore them, then re-run", id)
	}
	g := &gateResult{ms: d.Milliseconds()}
	if runErr != nil {
		var ee *exec.ExitError
		if !errors.As(runErr, &ee) {
			return nil, fmt.Errorf("cannot run the gate of %s: %w", id, runErr)
		}
		if g.exit = ee.ExitCode(); g.exit <= 0 {
			g.exit = 1
		}
	}
	dir := filepath.Join(it.RunDir, "gates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if timedOut {
		if g.exit == 0 {
			g.exit = 1
		}
		if out.Len() > 0 && !bytes.HasSuffix(out.Bytes(), []byte("\n")) {
			out.WriteByte('\n')
		}
		out.WriteString(state.GateTimedOutLine(id, it.Budgets.GateTimeout) + "\n")
	}
	masked := []byte(it.r.Mask(out.String()))
	if err := os.WriteFile(filepath.Join(dir, id+".log"), masked, 0o644); err != nil {
		return nil, err
	}
	if g.exit != 0 {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%03d-%s.fail.log", iter, id)), masked, 0o644); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func (it *Iterator) appendIteration(rec map[string]any) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(it.RunDir, "iterations.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

// copyReport keeps a masked copy of a handoff file under reports/; a file that
// is not there leaves nothing.
// readHandoff reads a session's handoff file only when it is a regular file; a
// symlink is treated as missing.
func readHandoff(path string) ([]byte, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", filepath.Base(path))
	}
	return os.ReadFile(path)
}

// copyReport keeps a report in the run folder, but only one that validates.
func (it *Iterator) copyReport(name, dest string) {
	schemaName := "proposal/v1"
	if name == "verdict.json" {
		schemaName = "verdict/v1"
	}
	data, ok := it.readReport(name, schemaName)
	if !ok {
		return
	}
	dir := filepath.Join(it.RunDir, "reports")
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, dest), []byte(it.r.Mask(string(data))), 0o644)
}

func (it *Iterator) journalPath() string {
	return filepath.Join(it.Root, filepath.FromSlash(journalsDir), it.plan.RunID+".md")
}

func (it *Iterator) appendJournal(t *state.Task, outcome, verdict, summary, notes string, files []string) error {
	if summary == "" {
		summary = "none"
	}
	if notes == "" {
		notes = "none"
	}
	fl := "none"
	if len(files) > 0 {
		fl = it.r.Mask(strings.Join(files, ", "))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n## %s — %s\n\n", t.ID, t.Title)
	fmt.Fprintf(&b, "- **Outcome:** %s (review: %s)\n", outcome, verdict)
	fmt.Fprintf(&b, "- **Summary:** %s\n", summary)
	fmt.Fprintf(&b, "- **Files:** %s\n", fl)
	fmt.Fprintf(&b, "- **Notes for next iteration:** %s\n", notes)
	return appendFile(it.journalPath(), b.String())
}

func appendFile(path, s string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(s)
	return err
}

// render rewrites plan.md as `vloop status --markdown` renders the plan.
func (it *Iterator) render() error {
	p, err := state.Load(it.Root)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(it.Root, filepath.FromSlash(planMarkdown)), []byte(p.Markdown()), 0o644)
}

// commit makes one commit of everything but .vloop/tmp/, after checking HEAD
// is still on the run's branch: a gate or session that moved it must not have
// its work committed onto another branch.
func (it *Iterator) commit(subject string) error {
	if cur := CurrentBranch(it.Root); cur != it.branch {
		it.warn("HEAD is on %s, not the run's branch %s — nothing was committed", cur, it.branch)
		_ = it.log.flush()
		return halt(ExitRefsMoved, "HEAD left branch %s for %s — nothing was committed", it.branch, cur)
	}
	if err := it.log.flush(); err != nil {
		return err
	}
	if err := it.writeSnapshot(); err != nil {
		return halt(ExitPreflight, "%v", err)
	}
	if _, err := git(it.Root, "add", "-A"); err != nil {
		return halt(ExitPreflight, "%v", err)
	}
	if _, err := git(it.Root, "reset", "-q", "--", tmpDir); err != nil {
		return halt(ExitPreflight, "%v", err)
	}
	if gitCmd(it.Root, "diff", "--cached", "--quiet").Run() == nil {
		return nil
	}
	if _, err := git(it.Root, "commit", "-q", "-m", subject); err != nil {
		return halt(ExitPreflight, "cannot commit %q: %v", subject, err)
	}
	return nil
}

// finish records how the run ended: the plan status, the journal, plan.md and
// the closing commit. Nothing is logged after that commit, so the tree stays
// clean.
func (it *Iterator) finish(end Ending, runIters int) error {
	plan := it.plan
	plan.Status = end.Status
	if err := it.save(); err != nil {
		return err
	}
	done, blocked, _ := counts(plan)
	it.say("")
	it.say("═══ %s ═══", end.Status)
	it.say("run:    %s/%s  (%d iteration(s) this run)", plan.RunID, filepath.Base(it.RunDir), runIters)
	it.say("plan:   %d/%d done, %d blocked", done, len(plan.Tasks), blocked)
	switch end.Status {
	case "complete":
		it.say("plan complete. journal: %s/%s.md", journalsDir, plan.RunID)
	case "blocked":
		it.say("a human is needed. read the blocked task's notes: vloop status")
	case "stalled":
		it.say("no recorded progress %d times running — read %s", it.Budgets.StallLimit, relRunDir(it.Root, it.RunDir))
	case "halted":
		switch end.Code {
		case ExitMaxIter:
			it.say("iteration budget spent. resumable: raise --max-iterations and re-run vloop run")
		case ExitCostCeiling:
			it.say("cost ceiling reached. resumable: raise --cost-ceiling and re-run vloop run")
		case ExitNotConverging:
			it.say("iterations per closed task exceeded %g — the run is not converging.", it.Budgets.ConvergenceMax)
		case ExitSessionError:
			it.say("a claude session failed. see %s", relRunDir(it.Root, it.RunDir))
		case ExitRepeatBlocked:
			it.say("%s", end.Note)
		}
	}
	if len(it.gatePasses) > 0 {
		it.say("")
		for _, id := range it.gatePasses {
			it.say("%s", gatePassesLine(id))
		}
	}
	for _, l := range it.noProposal {
		it.say("")
		it.say("%s", l)
	}
	if err := it.render(); err != nil {
		return err
	}
	// A non-empty stderr is evidence; the empty ones only bury it.
	if m, _ := filepath.Glob(filepath.Join(it.RunDir, "*.stderr")); m != nil {
		for _, f := range m {
			if fi, err := os.Stat(f); err == nil && fi.Size() == 0 {
				_ = os.Remove(f)
			}
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n## Run ended — %s\n\n", end.Status)
	fmt.Fprintf(&b, "- **Run:** `%s` · %d iteration(s) this run\n", plan.RunID, runIters)
	fmt.Fprintf(&b, "- **Plan:** %d/%d done, %d blocked\n", done, len(plan.Tasks), blocked)
	if err := appendFile(it.journalPath(), b.String()); err != nil {
		return err
	}
	return it.commit(fmt.Sprintf("[vloop] run %s/%s: %s", plan.RunID, filepath.Base(it.RunDir), end.Status))
}

// applyOutcome is the status transition of an iteration that is not a gate
// dispute: the driver's, never the session's.
func (it *Iterator) applyOutcome(task *state.Task, outcome, summary string, findings []string, tampered string, gatePassed bool) {
	switch outcome {
	case OutDone:
		task.Status, task.Notes = "done", ""
		it.say("   %s done", task.ID)
	case OutGateFailed:
		task.Status, task.Attempts = "pending", task.Attempts+1
		task.Notes = "gate failed — see " + it.relGate(task.ID)
		if tampered != "" {
			task.Notes = tampered
		}
	case OutRejected:
		task.Status, task.Attempts = "pending", task.Attempts+1
		task.Notes = it.r.Mask(strings.Join(findings, "; "))
	default:
		task.Status, task.Attempts = "pending", task.Attempts+1
		task.Notes = summary
		if gatePassed {
			task.Notes = strings.TrimSpace(summary + " — " + gatePassesLine(task.ID))
		}
	}
}

// treeGuard is the working tree as git reports it before a review session: every
// changed or untracked path with the digest and bytes it had. It is built from
// git status, not a file-system walk.
type treeGuard struct {
	root  string
	paths map[string]treeFile
}

type treeFile struct {
	exists bool
	sum    string
	data   []byte
}

// reviewMayWrite is where a review session may write: its verdict and scratch,
// and the driver's own state, which other guards restore.
func reviewMayWrite(p string) bool {
	return strings.HasPrefix(p, tmpDir+"/") || strings.HasPrefix(p, ".vloop/state/")
}

func snapshotTree(root string) treeGuard {
	g := treeGuard{root: root, paths: map[string]treeFile{}}
	cmd := gitCmd(root, "status", "--porcelain", "-z", "--no-renames", "--untracked-files=all")
	out, err := cmd.Output()
	if err != nil {
		return g
	}
	for _, e := range strings.Split(string(out), "\x00") {
		if len(e) < 4 {
			continue
		}
		if p := e[3:]; !reviewMayWrite(p) {
			g.paths[p] = readTreeFile(root, p)
		}
	}
	return g
}

func readTreeFile(root, p string) treeFile {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
	if err != nil {
		return treeFile{}
	}
	sum := sha256.Sum256(data)
	return treeFile{exists: true, sum: hex.EncodeToString(sum[:]), data: data}
}

// revert undoes every change made since the snapshot outside what a review may
// write, and returns the paths it had to put back.
func (g treeGuard) revert() []string {
	now := snapshotTree(g.root)
	var changed []string
	for p, f := range now.paths {
		if old, ok := g.paths[p]; !ok || old.exists != f.exists || old.sum != f.sum {
			changed = append(changed, p)
		}
	}
	for p := range g.paths {
		if _, ok := now.paths[p]; !ok {
			changed = append(changed, p)
		}
	}
	sort.Strings(changed)
	for _, p := range changed {
		abs := filepath.Join(g.root, filepath.FromSlash(p))
		if old, ok := g.paths[p]; ok {
			if old.exists {
				_ = os.MkdirAll(filepath.Dir(abs), 0o755)
				_ = os.WriteFile(abs, old.data, 0o644)
			} else {
				_ = os.Remove(abs)
			}
			continue
		}
		if err := gitCmd(g.root, "cat-file", "-e", "HEAD:"+p).Run(); err == nil {
			_ = gitCmd(g.root, "checkout", "HEAD", "--", p).Run()
		} else {
			_ = os.Remove(abs)
		}
	}
	return changed
}
