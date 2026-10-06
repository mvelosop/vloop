package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const failingGateReview = `if [ "$PHASE" = gate-review ]; then
  printf '{"schema":"gate-verdict/v1","verdict":"FAIL","tasks":[{"task":"T1","verdict":"FAIL","findings":[{"summary":"fails for the wrong reason","kind":"wrong-reason"}]}],"notes":"none"}\n' > .vloop/tmp/gate-verdict.json
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
	if v := r.vloop("schema", "validate", "gate-verdict/v1", f); v.code != 0 {
		r.t.Errorf("%s is not gate-verdict/v1: %+v", name, v)
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
