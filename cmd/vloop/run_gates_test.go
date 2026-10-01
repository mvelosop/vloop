package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Gates across iterations, as the shell loop's scenarios 03, 33, 39 and 41
// assert them of .loop/run.sh, translated into vloop's layout, plus the flaky
// gate and the gate dispute the shell driver never had.

func (r *runRepo) runLog() string {
	r.t.Helper()
	b, err := os.ReadFile(filepath.Join(r.runFolder(), "run.log"))
	if err != nil {
		r.t.Fatal(err)
	}
	return string(b)
}

func (r *runRepo) reviewed(id string) int {
	return strings.Count(strings.Join(r.argv(), "\n"), "/vloop:review "+id)
}

// commitFile puts a file into HEAD on main before the run, as a planner's
// artefact would be.
func (r *runRepo) commitFile(rel, content string) {
	r.t.Helper()
	r.write(rel, content)
	r.git("add", "-A")
	r.git("commit", "-q", "-m", "pre-existing "+rel)
}

// clobberOnce makes T2's work session delete T1's output, once.
const clobberOnce = `if [ "$PHASE" = work ] && [ "$TASK" = T2 ] && [ ! -f "$dir/broke" ]; then : > "$dir/broke"; rm -f T1.out; fi
`

// TestRun03GateRegression: a later task breaks an earlier one; re-running every
// done task's gate is what sees it.
func TestRun03GateRegression(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+clobberOnce)
	wantExit(t, r.vloop("run", runBrief), 0)
	wantIn(t, "run log", r.runLog(), "GATE REGRESSION T1")
	r.wantTask("T1", "done", 1)
	r.wantTask("T2", "done", 0)
	its := r.iterations()
	if len(its) < 3 || its[2]["task"] != "T1" {
		t.Errorf("the third iteration is not T1's: %v", its)
	}
	r.wantClean()
}

// TestRun41RegressionNamesBoth: the regression line names the reverted task
// and the task whose iteration caused it.
func TestRun41RegressionNamesBoth(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+clobberOnce)
	wantExit(t, r.vloop("run", runBrief), 0)
	r.wantTask("T1", "done", 1)
	r.wantTask("T2", "done", 0)
	var line string
	for _, l := range strings.Split(r.runLog(), "\n") {
		if strings.Contains(l, "GATE REGRESSION T1") {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatal("no GATE REGRESSION T1 line in the run log")
	}
	if !strings.Contains(strings.Replace(line, "GATE REGRESSION T1", "", 1), "T2") {
		t.Errorf("the regression line does not name T2: %s", line)
	}
}

// gateRewriteRepo is a repository whose single task T1 has the given files and
// verify, with the stub's work arm after the default's.
func gateRewriteRepo(t *testing.T, files []string, verify, work string) *runRepo {
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", map[string]any{"files": files, "verify": verify})), defaultScript+work)
	return r
}

func TestRun33GateRewrite(t *testing.T) {
	t.Run("modified, pre-existing, not assigned: the goalpost moved", func(t *testing.T) {
		r := newRunRepo(t)
		r.commitFile("probe.txt", "exit 0\n")
		r.scripted(planJSON(t, planTask("T1", map[string]any{"files": []string{"thing.txt"}, "verify": "sh probe.txt"})),
			defaultScript+`if [ "$PHASE" = work ]; then touch thing.txt; echo "exit 1" > probe.txt; fi
`)
		wantExit(t, r.vloop("run", runBrief), 2)
		wantIn(t, "run log", r.runLog(), "GATE REWRITE T1")
		if got := r.read("probe.txt"); got != "exit 0\n" {
			t.Errorf("probe.txt reads %q — a weakened gate survived the iteration", got)
		}
		its := r.iterations()
		if len(its) == 0 || its[0]["outcome"] != "gate_failed" {
			t.Errorf("iteration 1 = %v, want gate_failed", its)
		}
		if n := r.reviewed("T1"); n != 0 {
			t.Errorf("%d review(s): work is not reviewable when its gate was edited", n)
		}
		wantNotIn(t, "run log", r.runLog(), "review: PASS")
		// The stub offends every attempt, so the task ends blocked on the ceiling.
		r.wantTask("T1", "blocked", 3)
		if v := r.task("T1")["verify"]; v != "sh probe.txt" {
			t.Errorf("the driver changed verify to %v", v)
		}
		r.wantClean()
	})

	t.Run("created by the task that ships it", func(t *testing.T) {
		r := gateRewriteRepo(t, []string{"made.txt"}, "grep -q OK made.txt", `if [ "$PHASE" = work ]; then echo OK > made.txt; fi
`)
		wantExit(t, r.vloop("run", runBrief), 0)
		wantNotIn(t, "run log", r.runLog(), "GATE REWRITE")
		r.wantTask("T1", "done", 0)
	})

	t.Run("modified, pre-existing, but assigned", func(t *testing.T) {
		r := newRunRepo(t)
		r.commitFile("owned.txt", "OLD\n")
		r.scripted(planJSON(t, planTask("T1", map[string]any{"files": []string{"owned.txt"}, "verify": "grep -q NEW owned.txt"})),
			defaultScript+`if [ "$PHASE" = work ]; then echo NEW > owned.txt; fi
`)
		wantExit(t, r.vloop("run", runBrief), 0)
		wantNotIn(t, "run log", r.runLog(), "GATE REWRITE")
		r.wantTask("T1", "done", 0)
	})

	t.Run("pre-existing only because this task's own failed attempt committed it", func(t *testing.T) {
		r := gateRewriteRepo(t, []string{"other.txt"}, "grep -q RETRY owned_by_me.txt", `if [ "$PHASE" = work ]; then
  if [ -f owned_by_me.txt ]; then echo RETRY > owned_by_me.txt; else echo FIRST > owned_by_me.txt; fi
fi
if [ "$PHASE" = review ] && ! grep -q RETRY owned_by_me.txt 2>/dev/null; then
  printf '{"schema":"verdict/v1","task":"T1","verdict":"FAIL","criteria":[],"findings":[{"summary":"not yet","kind":"bug"}],"notes":"n"}\n' > .vloop/tmp/verdict.json
fi
`)
		// The first attempt writes FIRST, which the gate rejects; the retry's
		// edit of its own committed file is not someone else's gate.
		wantExit(t, r.vloop("run", runBrief), 0)
		wantNotIn(t, "run log", r.runLog(), "GATE REWRITE")
		r.wantTask("T1", "done", 1)
		if got := r.read("owned_by_me.txt"); got != "RETRY\n" {
			t.Errorf("owned_by_me.txt reads %q — the retry was falsely reverted", got)
		}
	})
}

// planOnlyScript plays the plan and leaves work and review to what follows.
const planOnlyScript = `if [ "$PHASE" = plan ]; then mkdir -p .vloop/state; cp "$dir/plan.json" .vloop/state/state.json; fi
`

const blockedProposal = `if [ "$PHASE" = work ]; then
  [ "$MAKE" = 1 ] && touch T1.out
  printf '{"schema":"proposal/v1","task":"T1","outcome":"blocked","summary":"stub summary","files":[],"verified":"none","notes":"stub summary"}\n' > .vloop/tmp/proposal.json
fi
`

var gateAndPass = regexp.MustCompile(`(?i)gate.*pass|pass.*gate`)

// says is the contract of the shell scenario: "gate" and "pass" together,
// naming T1 — not exact wording.
func says(text string) bool {
	for _, l := range strings.Split(text, "\n") {
		if gateAndPass.MatchString(l) && strings.Contains(l, "T1") {
			return true
		}
	}
	return false
}

func TestRun39BlockedGatePasses(t *testing.T) {
	blk := func(t *testing.T, make string) (*runRepo, result) {
		r := newRunRepo(t)
		r.scripted(planJSON(t, planTask("T1", nil)), planOnlyScript+strings.Replace(blockedProposal, `"$MAKE"`, `"`+make+`"`, 1))
		return r, r.vloop("run", runBrief, "--max-iterations", "1")
	}
	afterIteration := func(s string) string {
		if i := strings.Index(s, "── iteration"); i >= 0 {
			return s[i:]
		}
		return s
	}

	t.Run("blocked, and the deliverable satisfies the gate", func(t *testing.T) {
		r, res := blk(t, "1")
		wantExit(t, res, 4)
		r.wantIterations("T1:blocked")
		r.wantTask("T1", "pending", 1)
		wantNotIn(t, "run log", r.runLog(), "review: PASS")
		if n := r.reviewed("T1"); n != 0 {
			t.Errorf("%d review(s) of a blocked task", n)
		}
		if notes := r.task("T1")["notes"].(string); !says("T1 " + notes) {
			t.Errorf("notes do not say the gate passes: %s", notes)
		}
		iterLog := afterIteration(r.runLog())
		if i := strings.Index(iterLog, "═══"); i >= 0 {
			iterLog = iterLog[:i]
		}
		if !says(iterLog) {
			t.Errorf("the run log's iteration does not say it:\n%s", iterLog)
		}
		end := res.out[strings.Index(res.out, "═══"):]
		if !says(end) {
			t.Errorf("the run-end report does not say it:\n%s", end)
		}
	})

	t.Run("blocked, and the gate fails: nothing to report", func(t *testing.T) {
		r, res := blk(t, "0")
		wantExit(t, res, 4)
		r.wantIterations("T1:blocked")
		r.wantTask("T1", "pending", 1)
		wantNotIn(t, "run log", r.runLog(), "GATE REGRESSION", "GATE FAIL")
		if notes := r.task("T1")["notes"].(string); says("T1 " + notes) {
			t.Errorf("notes claim a passing gate: %s", notes)
		}
		if says(afterIteration(r.runLog())) {
			t.Errorf("the log claims a passing gate:\n%s", r.runLog())
		}
		if says(res.out) {
			t.Errorf("stdout claims a passing gate:\n%s", res.out)
		}
	})
}

func TestRunFlakyGate(t *testing.T) {
	t.Run("a gate that fails once, then passes on the immediate re-run", func(t *testing.T) {
		r := newRunRepo(t)
		seen := filepath.Join(r.stub, "flaky.seen")
		verify := `test -f T1.out && { [ -f "` + seen + `" ] || { : > "` + seen + `"; exit 1; }; }`
		r.scripted(twoTasks(t), defaultScript)
		r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": verify}), planTask("T2", map[string]any{"depends_on": []string{"T1"}})), defaultScript)
		wantExit(t, r.vloop("run", runBrief), 0)
		wantIn(t, "run log", r.runLog(), "FLAKY GATE T1")
		r.wantIterations("T1:done", "T2:done")
		its := r.iterations()
		g := its[0]["gate"].(map[string]any)
		if g["exit"] != float64(0) || g["flaky"] != true || g["duration_ms"] == nil {
			t.Errorf("the flaky iteration's gate = %v", g)
		}
		if g2 := its[1]["gate"].(map[string]any); g2["flaky"] != nil || len(g2) != 2 {
			t.Errorf("a gate that passed first time = %v, want exit and duration_ms only", g2)
		}
		r.wantTask("T1", "done", 0)
		if n := strings.Count(strings.Join(r.argv(), "\n"), "/vloop:work T1"); n != 1 {
			t.Errorf("%d work sessions for T1, want 1", n)
		}
	})

	t.Run("a gate that fails twice is not flaky", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(planJSON(t, planTask("T1", nil)), defaultScript+`if [ "$PHASE" = work ] && [ ! -f "$dir/once" ]; then : > "$dir/once"; rm -f T1.out; fi
`)
		wantExit(t, r.vloop("run", runBrief), 0)
		r.wantIterations("T1:gate_failed", "T1:done")
		if g := r.iterations()[0]["gate"].(map[string]any); g["flaky"] != nil || g["exit"] == float64(0) {
			t.Errorf("a twice-failing gate = %v", g)
		}
		r.wantTask("T1", "done", 1)
		wantNotIn(t, "run log", r.runLog(), "FLAKY GATE")
	})
}

func TestRunGateDispute(t *testing.T) {
	const dispute = `if [ "$PHASE" = work ] && [ "$TASK" = T1 ]; then
  printf '{"schema":"proposal/v1","task":"T1","outcome":"blocked","summary":"s","files":[],"verified":"v","notes":"n","gate_dispute":{"reason":"the gate reads the wrong file","evidence":"a.txt is written to out/a.txt"}}\n' > .vloop/tmp/proposal.json
fi
`
	two := func(t *testing.T) string {
		return planJSON(t, planTask("T1", nil), planTask("T2", nil))
	}

	t.Run("a disputed gate blocks its task at once", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(two(t), defaultScript+dispute)
		wantExit(t, r.vloop("run", runBrief), 2)
		p := r.plan()
		if p["status"] != "blocked" {
			t.Errorf("plan status %v, want blocked", p["status"])
		}
		r.wantTask("T1", "blocked", 0)
		r.wantTask("T2", "done", 0)
		if got, want := r.task("T1")["notes"], "gate disputed: the gate reads the wrong file — a.txt is written to out/a.txt"; got != want {
			t.Errorf("notes = %q, want %q", got, want)
		}
		if v := r.task("T1")["verify"]; v != "test -f T1.out" {
			t.Errorf("the driver changed verify to %v", v)
		}
		if n := r.reviewed("T1"); n != 0 {
			t.Errorf("a disputed task was reviewed %d time(s)", n)
		}
		if n := strings.Count(strings.Join(r.argv(), "\n"), "/vloop:work T1"); n != 1 {
			t.Errorf("a disputed task was worked %d times", n)
		}
		r.wantClean()
	})

	t.Run("a blocked proposal without gate_dispute charges an attempt", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(two(t), defaultScript+strings.Replace(dispute, `,"gate_dispute":{"reason":"the gate reads the wrong file","evidence":"a.txt is written to out/a.txt"}`, "", 1)+`if [ "$PHASE" = work ]; then : > "m-$ATTEMPT"; fi
`)
		wantExit(t, r.vloop("run", "--stall-limit", "99", runBrief), 2)
		r.wantTask("T1", "blocked", 3)
		wantNotIn(t, "task notes", r.task("T1")["notes"].(string), "gate disputed")
	})
}
