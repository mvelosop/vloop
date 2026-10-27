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
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/config"
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

// Outcome names of an iteration (iteration/v2).
const (
	OutDone       = "done"
	OutGateFailed = "gate_failed"
	OutRejected   = "rejected"
	OutBlocked    = "blocked"

	OutCheckFailed = "check_failed"
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

	AwakeWarn string // why the keep-awake hold could not be taken; logged once to run.log

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
	dissents   []dissentRecord        // second reviews that differed from the primary, this run
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

// clearTmp empties .vloop/tmp/ but for the run lock, warning of any entry it
// cannot remove.
func (it *Iterator) clearTmp() {
	ClearTmp(it.Root, func(entry string, err error) {
		it.warn("warning: cannot clear %s: %v", path.Join(tmpDir, entry), err)
	})
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
	if err := CheckFixtures(it.Root, plan); err != nil {
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
	if it.AwakeWarn != "" {
		it.r.Logf("%s", AwakeWarning(it.AwakeWarn))
	}

	// A resume takes the config's checks as they are now.
	if defs, err := config.Checks(it.Root); err == nil {
		if now := planChecks(defs); !sameChecks(plan.Checks, now) {
			it.r.Logf("the checks changed since the plan was made: were %s, now %s", checkNames(plan.Checks), checkNames(now))
			plan.Checks = now
			if err := it.save(); err != nil {
				return Ending{}, err
			}
		}
	}

	if it.resolved, err = ResolveRun(it.Root); err != nil {
		return Ending{}, halt(ExitPreflight, "%v", err)
	}
	// The one read of the records on disk: the sessions before this phase, such
	// as the plan session. Every later session is added as it finishes.
	it.spent = it.spend()

	if plan.GateReview.Verdict == "FAIL" {
		if err := it.resumeBaseGates(); err != nil {
			return Ending{}, err
		}
	}

	if plan.Status != "running" || plan.Branch != it.branch {
		plan.Status, plan.Branch = "running", it.branch
		if err := it.save(); err != nil {
			return Ending{}, err
		}
	}

	runIters, stalls := 0, 0
	doneAtStart := map[string]bool{}
	for _, t := range plan.Tasks {
		if t.Status == "done" {
			doneAtStart[t.ID] = true
		}
	}
	it.blockedAt = map[string]blockedMark{}
	end := Ending{Status: "halted", Code: ExitMaxIter}
	for {
		done, blocked := plan.Counts()
		pending := len(plan.Tasks) - done - blocked
		if pending == 0 {
			if blocked > 0 {
				end = Ending{Status: "blocked", Code: ExitBlocked}
			} else {
				it.clearTmp()
				line, err := it.finalPass()
				if err != nil {
					return Ending{}, err
				}
				if line != "" {
					it.warn("%s", line)
					end = Ending{Status: "blocked", Code: ExitBlocked, Note: line}
				} else {
					end = Ending{Status: "complete"}
					if err := it.verifySession(); err != nil {
						return Ending{}, err
					}
				}
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
		closed := closedThisRun(plan, doneAtStart)
		if runIters >= it.Budgets.ConvergenceMin &&
			(closed == 0 || float64(runIters)/float64(closed) > it.Budgets.ConvergenceMax) {
			it.warn("not converging: %d iteration(s) this run for %d closed task(s), over %.2f per closed task", runIters, closed, it.Budgets.ConvergenceMax)
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
		if nd, _ := plan.Counts(); nd <= done && res.outcome != OutGateFailed && res.outcome != OutRejected && res.outcome != OutCheckFailed {
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

// closedThisRun counts the tasks that became done during this run. A task done
// at run start that a gate regression reverted is forgotten from doneAtStart,
// so its redo counts as closed.
func closedThisRun(p *state.Plan, doneAtStart map[string]bool) int {
	closed := 0
	for _, t := range p.Tasks {
		switch {
		case t.Status != "done":
			delete(doneAtStart, t.ID)
		case !doneAtStart[t.ID]:
			closed++
		}
	}
	return closed
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
// reports whether anything had to be restored, and fails when something could
// not be. keep is the session's own record.
func (it *Iterator) restoreInputs(g inputGuard, phase, keep string) (bool, error) {
	changed, failed := g.restore(keep)
	for _, p := range changed {
		it.warn("vloop: %s session changed %s — restored", phase, p)
	}
	if len(failed) > 0 {
		return len(changed) > 0, halt(ExitPreflight, "%s session changed inputs the driver could not restore: %s", phase, strings.Join(failed, "; "))
	}
	refused := filepath.Join(it.Root, filepath.FromSlash(GateRefusedFile))
	if _, err := os.Stat(refused); err == nil {
		it.warn("vloop: %s session ran vloop task gate against a changed plan — refused", phase)
		_ = os.Remove(refused)
	}
	return len(changed) > 0, nil
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
func (it *Iterator) spend() float64 { return spendIn(it.RunDir) }

// sessionDone prints the line of a session that has just been added to
// it.spent, and the warnings for the thresholds it crossed.
func (it *Iterator) sessionDone(s Spec, res Result) {
	it.say("%s", sessionLine(s, res.Wall, res.Cost, it.spent, it.Budgets.CostCeiling))
	for _, w := range costWarnings(it.spent-res.Cost, it.spent, it.Budgets.CostCeiling) {
		it.warn("%s", w)
	}
}

// spendIn is what the sessions recorded in a run folder cost.
func spendIn(runDir string) float64 {
	files, _ := filepath.Glob(filepath.Join(runDir, "sessions", "*.json"))
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

// workReport is what a work session left, as the driver reads it: the
// proposed outcome and the text that goes in the journal, or the reason the
// session's work cannot be judged (a dispute or tampering with the plan).
type workReport struct {
	outcome, summary, notes string
	files                   []string
	dispute, tampered       string
}

func (it *Iterator) iterate(task *state.Task, runIters, done, total int) (iterResult, error) {
	plan := it.plan
	iter := plan.Iteration + 1
	started := it.now().UTC()
	id := task.ID
	it.say("── iteration %d (%d this run) · %s · %d/%d done ──", iter, runIters, id, done, total)
	it.say("   %s", task.Title)

	// No earlier session's scratch, and no stale handoff, reaches this session.
	it.clearTmp()

	// 1. work session
	w, err := it.workSession(task, iter)
	if err != nil {
		return iterResult{}, err
	}
	outcome, summary := w.outcome, w.summary

	// 2. gates: every done task, plus this one if it claims done or blocked
	own, failed, err := it.runGates(iter, id, w.dispute == "" && (outcome == OutDone || outcome == OutBlocked))
	if err != nil {
		return iterResult{}, err
	}

	// 3. checks: only when every gate that ran passed
	var checkRuns []checkRun
	failedCheck := ""
	if w.dispute == "" && w.tampered == "" && outcome == OutDone && len(failed) == 0 {
		checkRuns, failedCheck, err = it.checksStep(id, iter)
		if err != nil {
			return iterResult{}, err
		}
	}

	// 4. review
	verdict := "skipped"
	var findings []string
	var second *secondReview
	gatePassed := outcome == OutBlocked && w.dispute == "" && own != nil && own.exit == 0
	switch {
	case w.dispute != "":
		it.warn("   GATE DISPUTE %s — blocked for the operator; no review, no attempt charged", id)
	case w.tampered != "":
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
	case failedCheck != "":
		outcome = OutCheckFailed
		it.warn("   CHECK FAIL %s — %s; review skipped", id, failedCheck)
	default:
		outcome, verdict, findings, err = it.reviewSession(task, iter, outcome, checkRuns)
		if err != nil {
			return iterResult{}, err
		}
		it.copyReport("verdict.json", fmt.Sprintf("%03d-verdict.json", iter))
		mode := it.secondReviewMode()
		if second, err = it.secondReviewStep(task, iter, mode, verdict); err != nil {
			return iterResult{}, err
		}
		if second != nil {
			switch {
			case mode == "feedback" && verdict != "PASS":
				findings = append(findings, second.findings...)
			case mode == "both-pass" && second.verdict != "PASS":
				outcome = OutRejected
				findings = second.findings
				if len(findings) == 0 {
					findings = []string{"second review: the second review failed the task"}
				}
				verdict += ", second review: " + second.verdict
			}
		}
	}

	// 5. apply: the driver makes every status transition
	attempt := task.Attempts + 1
	switch {
	case w.dispute != "":
		outcome = OutBlocked
		task.Status, task.Notes = "blocked", w.dispute
		it.say("   %s blocked — the operator resolves the disputed gate with vloop task verify", id)
	default:
		it.applyOutcome(task, outcome, summary, findings, w.tampered, gatePassed, failedCheck)
	}
	repeat := it.repeatBlocked(id, outcome, summary)
	it.blockAtCeiling(task)
	plan.Iteration = iter
	if err := it.save(); err != nil {
		return iterResult{}, err
	}

	// 6. record
	rec := map[string]any{
		"schema": "iteration/v2", "run_id": plan.RunID, "iteration": iter, "task": id, "attempt": attempt,
		"outcome": outcome, "gate": gateRecord(own), "checks": checksRecord(checkRuns),
		"started": started.Format(time.RFC3339), "ended": it.now().UTC().Format(time.RFC3339),
	}
	if second != nil {
		rec["schema"] = "iteration/v3"
		rec["second_review"] = map[string]any{"verdict": second.verdict, "dissent": second.dissent}
	}
	if err := it.record(rec, task, iter, outcome, verdict, w); err != nil {
		return iterResult{}, err
	}

	// 7. commit: one per iteration
	if err := it.commit(fmt.Sprintf("[vloop] %s: %s", id, outcome)); err != nil {
		return iterResult{outcome: outcome, repeat: repeat}, err
	}
	it.snapshot()
	return iterResult{outcome: outcome, repeat: repeat}, nil
}

// checksStep runs the checks that match what the task changed. It returns the
// runs and a description of the first failing one, or "".
func (it *Iterator) checksStep(id string, iter int) ([]checkRun, string, error) {
	runs, err := it.runChecks(checksFor(it.plan.Checks, changedPaths(it.Root, id)), "", iter)
	if err != nil {
		return nil, "", err
	}
	failed := ""
	for _, c := range runs {
		if c.exit != 0 {
			failed = fmt.Sprintf("check %s failed — see %s", c.name, c.log)
		}
	}
	return runs, failed, nil
}

// blockAtCeiling blocks every pending task that has used its attempts, saying
// so for the task just iterated.
func (it *Iterator) blockAtCeiling(task *state.Task) {
	if task.Status != "pending" || task.Attempts < it.Budgets.MaxAttempts {
		return
	}
	for i := range it.plan.Tasks {
		if t := &it.plan.Tasks[i]; t.Status == "pending" && t.Attempts >= it.Budgets.MaxAttempts {
			t.Status = "blocked"
		}
	}
	it.warn("   %s hit the attempt ceiling — blocked", task.ID)
}

// gateRecord is the iteration record's view of the task's own gate run, nil
// when it did not run.
func gateRecord(own *gateResult) map[string]any {
	if own == nil {
		return nil
	}
	g := map[string]any{"exit": own.exit, "duration_ms": own.ms}
	if own.flaky {
		g["flaky"] = true
	}
	return g
}

// record writes the iteration's record, copies the session reports, and
// appends the journal entry and the rendered plan.
func (it *Iterator) record(rec map[string]any, task *state.Task, iter int, outcome, verdict string, w workReport) error {
	if err := it.appendIteration(rec); err != nil {
		return err
	}
	it.copyReport("proposal.json", fmt.Sprintf("%03d-proposal.json", iter))
	if err := it.appendJournal(task, outcome, verdict, w.summary, w.notes, w.files); err != nil {
		return err
	}
	return it.render()
}

// workSession runs the task's work session, restores what it changed among
// the driver's inputs and reads its proposal. A session that could not run, or
// timed out, is errSessionError and charges no attempt.
func (it *Iterator) workSession(task *state.Task, iter int) (workReport, error) {
	root := it.Root
	id := task.ID
	model, effort := it.resolved.For(task, PhaseWork)
	before := refsState(root)
	guard := snapshotState(root)
	gguard := snapshotGit(root)
	inputs := snapshotInputs(root, it.RunDir)
	_ = os.Remove(filepath.Join(root, filepath.FromSlash(GateRefusedFile)))
	wres, err := it.r.Run(Spec{Agent: it.resolved.Agent(PhaseWork), Phase: PhaseWork, Iteration: iter, Arg: id, Model: model, Effort: effort, PlanSHA: guard.hash()})
	it.spent += wres.Cost
	if err == nil {
		it.sessionDone(Spec{Phase: PhaseWork, Arg: id}, wres)
	}
	if err != nil {
		return workReport{}, err
	}
	if _, err := it.restoreInputs(inputs, PhaseWork, wres.Path); err != nil {
		return workReport{}, err
	}
	if err := it.gitChanged(gguard, PhaseWork); err != nil {
		return workReport{}, err
	}
	if err := it.gitRefsMoved(before, id, PhaseWork); err != nil {
		return workReport{}, err
	}

	w := workReport{outcome: OutBlocked, notes: "none"}
	data, ok := it.readReport("proposal.json", "proposal/v1")
	if wres.TimedOut {
		it.warn("   the work session for %s timed out after %s", id, it.sessionTimeout())
		return workReport{}, errSessionError
	}
	if !ok && (wres.ExitCode != 0 || wres.IsError) {
		// An infrastructure failure, not the task's: no attempt is charged.
		return workReport{}, errSessionError
	}
	if !ok {
		it.warn("work session left no valid proposal")
		changed := it.sessionTreeChanges()
		if len(changed) > 0 {
			w.summary = fmt.Sprintf("work session left no valid proposal; before dying it changed %s — the driver committed them under this iteration", strings.Join(changed, ", "))
		} else {
			w.summary = "work session left no valid proposal; the working tree is unchanged"
		}
		w.summary = it.r.Mask(w.summary)
		if len(changed) > 0 {
			it.noProposal = append(it.noProposal, id+" — "+w.summary)
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
		w.outcome, w.summary, w.notes, w.files = p.Outcome, it.r.Mask(p.Summary), it.r.Mask(p.Notes), p.Files
		if p.Dispute != nil {
			w.dispute = it.r.Mask(fmt.Sprintf("gate disputed: %s — %s", p.Dispute.Reason, p.Dispute.Evidence))
		}
	}

	// A session does not touch the plan. Restored from HEAD before any gate
	// runs, and the work is not reviewable.
	if touched := guard.restoreTouched(); len(touched) > 0 {
		w.tampered = it.r.Mask(tamperNote(PhaseWork))
		it.warn("   STATE TAMPERING %s — %s was modified; restored, iteration failed", id, strings.Join(touched, ", "))
	}
	return w, nil
}

// runGates runs the gate of every done task, plus the task's own when
// withOwn, and reverts to pending any other task whose gate now fails. It
// returns the task's own result (nil when not run) and the ids that failed.
func (it *Iterator) runGates(iter int, id string, withOwn bool) (*gateResult, map[string]bool, error) {
	plan := it.plan
	targets := []string{}
	for _, t := range plan.Tasks {
		if t.Status == "done" {
			targets = append(targets, t.ID)
		}
	}
	if withOwn {
		targets = append(targets, id)
	}
	var own *gateResult
	failed := map[string]bool{}
	for _, gid := range targets {
		g, err := it.runGateRetry(iter, id, gid)
		if err != nil {
			return nil, nil, err
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
	return own, failed, nil
}

// reviewResult is what a review session of either kind came to: its verdict
// and findings after the guards, or that it timed out.
type reviewResult struct {
	verdict  string
	findings []string
	timedOut bool
}

// reviewRun runs one review session of the phase (review or second-review) and
// applies the guards to it: inputs restored, a moved ref or git setting halts,
// changed files reverted, tampering with the plan restored. A session that
// left no valid verdict, changed a file or tampered is FAIL, with the finding.
// An error is a session that could not run or a halt.
func (it *Iterator) reviewRun(task *state.Task, iter int, phase string) (reviewResult, error) {
	root := it.Root
	id := task.ID
	model, effort := it.resolved.For(task, phase)
	before := refsState(root)
	guard := snapshotState(root)
	gguard := snapshotGit(root)
	inputs := snapshotInputs(root, it.RunDir)
	tree := snapshotTree(root)
	_ = os.Remove(filepath.Join(root, filepath.FromSlash(GateRefusedFile)))
	rres, err := it.r.Run(Spec{Agent: it.resolved.Agent(phase), Phase: phase, Iteration: iter, Arg: id, Model: model, Effort: effort, PlanSHA: guard.hash()})
	it.spent += rres.Cost
	if err == nil {
		it.sessionDone(Spec{Phase: phase, Arg: id}, rres)
	}
	if err != nil {
		return reviewResult{}, err
	}
	inputsChanged, err := it.restoreInputs(inputs, phase, rres.Path)
	if err != nil {
		return reviewResult{}, err
	}
	if err := it.gitChanged(gguard, phase); err != nil {
		return reviewResult{}, err
	}
	if err := it.gitRefsMoved(before, id, phase); err != nil {
		return reviewResult{}, err
	}
	treeChanged, treeFailed := tree.revert()
	if len(treeFailed) > 0 {
		return reviewResult{}, halt(ExitPreflight, "the %s session changed files the driver could not revert: %s", phase, strings.Join(treeFailed, "; "))
	}
	for _, p := range treeChanged {
		it.warn("vloop: the %s session changed %s — reverted", phase, p)
	}
	touched := guard.restoreTouched()
	if len(touched) > 0 {
		it.warn("   STATE TAMPERING %s — %s session modified %s; restored", id, phase, strings.Join(touched, ", "))
	}
	if rres.TimedOut {
		return reviewResult{timedOut: true}, nil
	}
	res := reviewResult{}
	if data, ok := it.readReport("verdict.json", "verdict/v1"); !ok {
		it.warn("   %s session left no valid verdict — treating as FAIL", phase)
		res.verdict = "FAIL"
		if phase == PhaseSecondReview {
			res.findings = append(res.findings, "the second review left no valid verdict")
		}
	} else {
		var v struct {
			Verdict  string `json:"verdict"`
			Findings []struct {
				Summary string `json:"summary"`
			} `json:"findings"`
		}
		_ = json.Unmarshal(data, &v)
		res.verdict = v.Verdict
		for _, f := range v.Findings {
			res.findings = append(res.findings, f.Summary)
		}
	}
	if len(touched) > 0 {
		res.verdict = "FAIL"
		res.findings = append(res.findings, it.r.Mask(tamperNote(phase)))
	}
	if inputsChanged {
		res.verdict = "FAIL"
		res.findings = append(res.findings, phase+" session changed the driver's inputs — restored by the driver; a review judges the work and changes nothing else")
	}
	if len(treeChanged) > 0 {
		res.verdict = "FAIL"
		res.findings = append(res.findings, "the "+phase+" session changed files")
	}
	return res, nil
}

// reviewSession runs the review of a task whose gates and checks passed and
// returns the outcome it leads to, the verdict and the findings.
func (it *Iterator) reviewSession(task *state.Task, iter int, outcome string, checkRuns []checkRun) (string, string, []string, error) {
	if err := it.writeChecksFile(checkRuns); err != nil {
		return outcome, "skipped", nil, err
	}
	res, err := it.reviewRun(task, iter, PhaseReview)
	if err != nil {
		return outcome, "skipped", nil, err
	}
	if res.timedOut {
		it.warn("   the review session for %s timed out after %s", task.ID, it.sessionTimeout())
		return outcome, "skipped", nil, errSessionError
	}
	if res.verdict != "PASS" {
		outcome = OutRejected
	}
	it.say("   review: %s", res.verdict)
	return outcome, res.verdict, res.findings, nil
}

// secondReviewMode is run.second-review: off, record, feedback, on-pass or
// both-pass.
func (it *Iterator) secondReviewMode() string {
	if v, err := config.Get(it.Root, "run.second-review"); err == nil {
		return v.Value
	}
	return "off"
}

// secondReview is what a second review came to: its verdict, its findings
// (each prefixed "second review: ") and whether it differs from the primary's.
type secondReview struct {
	verdict  string
	findings []string
	dissent  bool
}

// dissentRecord is one dissent of the run, for the list at its end.
type dissentRecord struct{ task, primary, second string }

// secondReviewStep runs the second review that follows a primary review that
// ran, when run.second-review asks for one after this primary verdict. The
// primary's verdict is removed from .vloop/tmp/ first, so the second session
// cannot see it; the caller has already kept it. It returns nil when none ran; a
// session error or timeout is a warning, except in both-pass, where it ends
// the iteration as the primary's would.
func (it *Iterator) secondReviewStep(task *state.Task, iter int, mode, primary string) (*secondReview, error) {
	switch {
	case mode == "off", mode == "":
		return nil, nil
	case (mode == "on-pass" || mode == "both-pass") && primary != "PASS":
		return nil, nil
	}
	_ = os.Remove(filepath.Join(it.Root, filepath.FromSlash(tmpDir), "verdict.json"))
	res, err := it.reviewRun(task, iter, PhaseSecondReview)
	if err == nil && res.timedOut {
		it.warn("   the second review session for %s timed out after %s", task.ID, it.sessionTimeout())
		err = errSessionError
	}
	if err != nil {
		if halted(err) || mode == "both-pass" {
			return nil, err
		}
		it.warn("   the second review for %s did not run: %v", task.ID, err)
		return nil, nil
	}
	it.copyReport("verdict.json", fmt.Sprintf("%03d-second-verdict.json", iter))
	_ = os.Remove(filepath.Join(it.Root, filepath.FromSlash(tmpDir), "verdict.json"))
	sr := &secondReview{verdict: res.verdict, dissent: res.verdict != primary}
	for _, f := range res.findings {
		sr.findings = append(sr.findings, "second review: "+it.r.Mask(f))
	}
	line := "   second review: " + sr.verdict
	if sr.dissent {
		line += " (dissent)"
		it.dissents = append(it.dissents, dissentRecord{task.ID, primary, sr.verdict})
	}
	it.say("%s", line)
	return sr, nil
}

// halted reports whether err ends the run with a halt rather than being a
// session that failed to run.
func halted(err error) bool {
	var h *Halt
	return errors.As(err, &h)
}

// verifySession runs the one verification session of a complete run, when
// run.verify allows it, and keeps its findings report. It is a report, not a
// gate: nothing it finds changes the run's ending. Only a ref or a git
// setting it moved halts the run, as for any session.
func (it *Iterator) verifySession() error {
	root := it.Root
	if on, err := config.Get(root, "run.verify"); err == nil && on.Value == "false" {
		return nil
	}
	if it.spent >= it.Budgets.CostCeiling {
		it.warn("vloop: cost ceiling reached — verify skipped")
		return nil
	}
	it.clearTmp()
	model, effort := it.resolved.For(nil, PhaseVerify)
	before := refsState(root)
	guard := snapshotState(root)
	gguard := snapshotGit(root)
	inputs := snapshotInputs(root, it.RunDir)
	tree := snapshotTree(root)
	res, err := it.r.Run(Spec{Agent: it.resolved.Agent(PhaseVerify), Phase: PhaseVerify, Model: model, Effort: effort, PlanSHA: guard.hash()})
	it.spent += res.Cost
	if err != nil {
		it.warn("vloop: the verify session failed to run — verify by hand; see %s", relRunDir(root, it.RunDir)+"run.log")
		return nil
	}
	it.sessionDone(Spec{Phase: PhaseVerify}, res)
	if _, err := it.restoreInputs(inputs, PhaseVerify, res.Path); err != nil {
		return err
	}
	if err := it.gitChanged(gguard, PhaseVerify); err != nil {
		return err
	}
	if err := it.gitRefsMoved(before, "verify", PhaseVerify); err != nil {
		return err
	}
	changed, failed := tree.revert()
	if len(failed) > 0 {
		return halt(ExitPreflight, "the verify session changed files the driver could not revert: %s", strings.Join(failed, "; "))
	}
	for _, p := range changed {
		it.warn("vloop: the verify session changed %s — reverted", p)
	}
	guard.restoreTouched()
	if res.TimedOut || res.ExitCode != 0 || !res.Recorded {
		it.warn("vloop: the verify session failed to run — verify by hand; see %s", relRunDir(root, it.RunDir)+"run.log")
		return nil
	}
	data, ok := it.readReport("verify.json", "verify/v1")
	if !ok {
		it.warn("vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand")
		return nil
	}
	var v struct {
		Findings []json.RawMessage `json:"findings"`
	}
	_ = json.Unmarshal(data, &v)
	dir := filepath.Join(it.RunDir, "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "verify.json"), []byte(it.r.Mask(string(data))+"\n"), 0o644); err != nil {
		return err
	}
	if len(v.Findings) == 0 {
		it.say("verify: no findings")
	} else {
		it.say("verify: %d finding(s) — see %sreports/verify.json", len(v.Findings), relRunDir(root, it.RunDir))
	}
	return nil
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
	// timedOut: the gate hit its time limit; it is not re-run
	timedOut bool
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
	env := dropEnv(base, "VLOOP_ACTIVE_TASK", "VLOOP_GATE_TASK")
	env = append(env, "VLOOP_ACTIVE_TASK="+active, "VLOOP_GATE_TASK="+id)
	cmd, err := state.GateCommand(it.Root, it.plan.Shell, t.Verify, env)
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	refsBefore := refsState(it.Root)
	gguard := snapshotGit(it.Root)
	tree := snapshotTreeIgnoring(it.Root, gateMayWrite(it.plan.GateScratch))
	start := time.Now()
	timedOut, runErr := state.RunGroup(cmd, it.gateTimeout())
	d := time.Since(start)
	restored, revertFailed := tree.revert()
	if len(revertFailed) > 0 {
		return nil, halt(ExitPreflight, "gate %s changed the tree and the driver could not revert it: %s", id, strings.Join(revertFailed, "; "))
	}
	if err := state.EmptyScratch(it.Root, it.plan.GateScratch); err != nil {
		it.warn("   gate scratch of %s not emptied: %v", id, err)
	}
	if err := it.gitChanged(gguard, "gate"); err != nil {
		return nil, err
	}
	if len(refsDiff(refsBefore, refsState(it.Root))) > 0 {
		return nil, it.haltGit("the gate of %s moved git refs — nothing was committed; restore them, then re-run", id)
	}
	g := &gateResult{ms: d.Milliseconds(), timedOut: timedOut}
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
	if len(restored) > 0 {
		if g.exit == 0 {
			g.exit = 1
		}
		if out.Len() > 0 && !bytes.HasSuffix(out.Bytes(), []byte("\n")) {
			out.WriteByte('\n')
		}
		out.WriteString(state.GateChangedTreeLine(id, restored) + "\n")
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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		it.warn("vloop: could not keep %s in the run folder: %v", name, err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, dest), []byte(it.r.Mask(string(data))), 0o644); err != nil {
		it.warn("vloop: could not keep %s in the run folder: %v", name, err)
	}
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
	done, blocked := plan.Counts()
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
	if len(it.dissents) > 0 {
		var parts []string
		for _, d := range it.dissents {
			parts = append(parts, fmt.Sprintf("%s (primary %s, second %s)", d.task, d.primary, d.second))
		}
		it.say("")
		it.say("second review: %d dissent(s): %s", len(it.dissents), strings.Join(parts, ", "))
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
func (it *Iterator) applyOutcome(task *state.Task, outcome, summary string, findings []string, tampered string, gatePassed bool, failedCheck string) {
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
	case OutCheckFailed:
		task.Status, task.Attempts = "pending", task.Attempts+1
		task.Notes = failedCheck
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
	root   string
	paths  map[string]treeFile
	ignore func(string) bool
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

// gateMayWrite is where a gate may write without being failed: the run's
// scratch space and its own gate scratch folders.
func gateMayWrite(scratch []string) func(string) bool {
	return func(p string) bool {
		if strings.HasPrefix(p, tmpDir+"/") {
			return true
		}
		for _, d := range scratch {
			if strings.HasPrefix(p, strings.TrimSuffix(d, "/")+"/") {
				return true
			}
		}
		return false
	}
}

func snapshotTree(root string) treeGuard {
	return snapshotTreeIgnoring(root, reviewMayWrite)
}

// snapshotTreeIgnoring is snapshotTree leaving out the paths ignore names.
func snapshotTreeIgnoring(root string, ignore func(string) bool) treeGuard {
	g := treeGuard{root: root, paths: map[string]treeFile{}, ignore: ignore}
	cmd := gitCmd(root, "status", "--porcelain", "-z", "--no-renames", "--untracked-files=all")
	out, err := cmd.Output()
	if err != nil {
		return g
	}
	for _, e := range strings.Split(string(out), "\x00") {
		if len(e) < 4 {
			continue
		}
		if p := e[3:]; !ignore(p) {
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
func (g treeGuard) revert() (changed, failed []string) {
	now := snapshotTreeIgnoring(g.root, g.ignore)
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
	var reverted []string
	for _, p := range changed {
		abs := filepath.Join(g.root, filepath.FromSlash(p))
		var err error
		if old, ok := g.paths[p]; ok {
			if old.exists {
				err = os.MkdirAll(filepath.Dir(abs), 0o755)
				if err == nil {
					err = os.WriteFile(abs, old.data, 0o644)
				}
			} else if err = os.Remove(abs); os.IsNotExist(err) {
				err = nil
			}
		} else if gitCmd(g.root, "cat-file", "-e", "HEAD:"+p).Run() == nil {
			err = gitCmd(g.root, "checkout", "HEAD", "--", p).Run()
		} else if err = os.Remove(abs); os.IsNotExist(err) {
			err = nil
		}
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", p, err))
			continue
		}
		reverted = append(reverted, p)
	}
	return reverted, failed
}
