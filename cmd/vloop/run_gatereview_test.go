package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const failingGateReview = `if [ "$PHASE" = gate-review ]; then
  printf '{"schema":"gate-verdict/v2","verdict":"FAIL","tasks":[{"task":"T1","verdict":"FAIL","findings":[{"summary":"fails for the wrong reason","kind":"wrong-reason"}]}],"notes":"none"}\n' > .vloop/tmp/gate-verdict.json
fi
`

func (r *runRepo) gateReviews() int {
	return strings.Count(strings.Join(r.argv(), "\n"), "-p /vloop:gate-review")
}

func (r *runRepo) gateReviewState() string {
	r.t.Helper()
	raw, err := os.ReadFile(filepath.Join(r.dir, ".vloop", "state", "state.json"))
	if err != nil {
		r.t.Fatal(err)
	}
	var p struct {
		GateReview struct {
			Rounds  int    `json:"rounds"`
			Verdict string `json:"verdict"`
		} `json:"gate_review"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		r.t.Fatal(err)
	}
	return strings.Join([]string{string(rune('0' + p.GateReview.Rounds)), p.GateReview.Verdict}, " ")
}

func (r *runRepo) wantReport(name string) {
	r.t.Helper()
	f := filepath.Join(r.runFolder(), "reports", name)
	if v := r.vloop("schema", "validate", "gate-verdict/v2", f); v.code != 0 {
		r.t.Errorf("%s is not gate-verdict/v2: %+v", name, v)
	}
}

func TestRunGateReviewPasses(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)
	if n := r.gateReviews(); n != 1 {
		t.Errorf("%d gate review sessions, want 1", n)
	}
	if got := r.gateReviewState(); got != "1 PASS" {
		t.Errorf("gate_review is %q, want 1 PASS", got)
	}
	r.wantReport("gate-review-1.json")
	if n := r.planSessions(); n != 1 {
		t.Errorf("%d plan sessions, want 1", n)
	}
}

func TestRunGateReviewFailThenPass(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$ATTEMPT" = 1 ]; then
`+failingGateReview+`fi
`)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "output", r.outputs(res), "gate T1 (wrong-reason): fails for the wrong reason")
	if r.planSessions() != 2 || r.gateReviews() != 2 {
		t.Errorf("%d plan and %d gate review sessions, want 2 and 2", r.planSessions(), r.gateReviews())
	}
	if got := r.gateReviewState(); got != "2 PASS" {
		t.Errorf("gate_review is %q, want 2 PASS", got)
	}
	r.wantReport("gate-review-1.json")
	r.wantReport("gate-review-2.json")
}

func TestRunGateReviewFailsTwiceThenResumes(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+failingGateReview)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 2)
	wantIn(t, "stderr", res.err, "vloop: the gate review failed the plan twice — see .vloop/state/runs/"+runID+"/")
	if n := strings.Count(res.err, "the gate review failed the plan twice"); n != 1 {
		t.Errorf("the refusal is printed %d times, want once (one line starting with vloop: ):\n%s", n, res.err)
	}
	wantIn(t, "stderr", res.err, "/reports/gate-review-2.json; amend the gates with vloop task verify, then re-run")
	r.wantStatus("blocked")
	if r.planSessions() != 2 || r.gateReviews() != 2 || strings.Contains(strings.Join(r.argv(), "\n"), "/vloop:work") {
		t.Errorf("%d plan, %d gate review sessions or a work session ran, want 2, 2 and none", r.planSessions(), r.gateReviews())
	}
	if got := r.gateReviewState(); got != "2 FAIL" {
		t.Errorf("gate_review is %q, want 2 FAIL", got)
	}
	r.wantReport("gate-review-1.json")
	r.wantReport("gate-review-2.json")

	// Resuming repeats the base run of the gates and not the review.
	r.script(defaultScript)
	wantExit(t, r.vloop("run"), 0)
	if n := r.gateReviews(); n != 2 {
		t.Errorf("%d gate review sessions after the resume, want still 2", n)
	}
	if !r.has(".vloop/state/runs/" + runID + "/" + filepath.Base(r.runFolder()) + "/gates/base-T1.log") {
		t.Error("the resume did not run the gates on the base")
	}
	r.wantTask("T1", "done", 0)
}

func TestRunGateReviewMissingVerdictFails(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$PHASE" = gate-review ] && [ "$ATTEMPT" = 1 ]; then rm -f .vloop/tmp/gate-verdict.json; fi
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	if r.planSessions() != 2 {
		t.Errorf("%d plan sessions, want 2: a missing verdict is a FAIL", r.planSessions())
	}
	if got := r.gateReviewState(); got != "2 PASS" {
		t.Errorf("gate_review is %q, want 2 PASS", got)
	}
	r.wantReport("gate-review-1.json")
}

const gapFinding = `{"summary":"pnpm smoke: no check runs it","kind":"check-gap"}`
const wrongFinding = `{"summary":"fails for the wrong reason","kind":"wrong-reason"}`

// verdictScript makes the gate-review session of the given attempt write a FAIL
// verdict holding the findings.
func verdictScript(attempt int, findings ...string) string {
	return fmt.Sprintf(`if [ "$PHASE" = gate-review ] && [ "$ATTEMPT" = %d ]; then
  printf '{"schema":"gate-verdict/v2","verdict":"FAIL","tasks":[{"task":"T1","verdict":"FAIL","findings":[%s]}],"notes":"none"}\n' > .vloop/tmp/gate-verdict.json
fi
`, attempt, strings.Join(findings, ","))
}

// wantGapHalt checks the run ended with the check-gap halt of the round and kept nothing.
func wantGapHalt(t *testing.T, r *runRepo, res result, round, plans int) {
	t.Helper()
	wantExit(t, res, 2)
	wantIn(t, "stderr", res.err, "vloop: the gate review found 1 check gap(s): the plan needs a check .vloop/config.toml does not have — add it, commit it, then re-run (see .vloop/state/runs/"+runID+"/")
	wantIn(t, "stderr", res.err, fmt.Sprintf("/reports/gate-review-%d.json)", round))
	wantIn(t, "output", r.outputs(res), "gate T1 (check-gap): pnpm smoke: no check runs it")
	if strings.Contains(r.outputs(res), "the gate review failed the plan twice") {
		t.Error("the run says the gate review failed the plan twice")
	}
	r.wantReport(fmt.Sprintf("gate-review-%d.json", round))
	if r.planSessions() != plans || r.gateReviews() != round || strings.Contains(strings.Join(r.argv(), "\n"), "/vloop:work") {
		t.Errorf("%d plan, %d gate review sessions or a work session ran, want %d, %d and none", r.planSessions(), r.gateReviews(), plans, round)
	}
	for _, f := range []string{".vloop/state/state.json", ".vloop/state/plan.md", ".vloop/state/gates"} {
		if r.has(f) {
			t.Errorf("%s was kept by the halt", f)
		}
	}
	if n := r.planCommits(); n != 0 {
		t.Errorf("%d plan commit(s), want none", n)
	}
}

func TestRunGateReviewCheckGapHalts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		script   string
		round    int
		findings string
	}{
		{"round 1 alone", verdictScript(1, gapFinding), 1, ""},
		{"round 1 beside another finding", verdictScript(1, gapFinding, wrongFinding), 1, wrongFinding},
		{"round 2 after a sent-back finding", verdictScript(1, wrongFinding) + verdictScript(2, gapFinding), 2, wrongFinding},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := newRunRepo(t)
			r.scripted(twoTasks(t), defaultScript+c.script)
			res := r.vloop("run", runBrief)
			wantGapHalt(t, r, res, c.round, c.round)
			if c.findings != "" {
				wantIn(t, "output", r.outputs(res), "gate T1 (wrong-reason): fails for the wrong reason")
			}
			if c.round == 1 && strings.Contains(r.outputs(res), "sending") {
				t.Error("the check-gap was sent back to the plan session")
			}

			// The re-run, with the check added, plans afresh.
			r.script(defaultScript)
			before := r.planSessions()
			wantExit(t, r.vloop("run", runBrief), 0)
			if r.planSessions() != before+1 {
				t.Errorf("the re-run ran %d plan session(s), want 1", r.planSessions()-before)
			}
		})
	}
}

func TestRunGateReviewUncoveredIsSentBack(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+verdictScript(1, `{"summary":"T1.out holds what the operator meant","kind":"uncovered"}`))
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "output", r.outputs(res), "gate T1 (uncovered): T1.out holds what the operator meant")
	wantIn(t, "output", r.outputs(res), "sending 1 problem(s) back to the plan session")
	if r.planSessions() != 2 || r.gateReviews() != 2 {
		t.Errorf("%d plan and %d gate review sessions, want 2 and 2", r.planSessions(), r.gateReviews())
	}
}
