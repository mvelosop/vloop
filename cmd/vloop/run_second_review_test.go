package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The second review: a fresh /vloop:review session of phase second-review,
// told apart by the stub from its --model. Its verdict is FAIL while
// $dir/second.fail exists, else PASS; the primary passes (the first review of
// T1 fails when $dir/primary.failfirst exists).
const secondScript = `if [ "$PHASE" = review ] && [ "$MODEL" = second-model ]; then
  echo "second $TASK visible=$(ls .vloop/tmp/verdict.json 2>/dev/null | wc -l | tr -d ' ')" >> "$dir/second.log"
fi
` + defaultScript + `if [ "$PHASE" = review ]; then
  if [ "$MODEL" = second-model ]; then
    V=PASS; [ -f "$dir/second.fail" ] && V=FAIL
    printf '{"schema":"verdict/v1","task":"%s","verdict":"%s","criteria":[],"findings":[{"summary":"second says no","kind":"bug"}],"notes":"second"}\n' "$TASK" "$V" > .vloop/tmp/verdict.json
    [ -f "$dir/second.touch" ] && echo changed >> src.txt
  elif [ -f "$dir/primary.failfirst" ] && [ "$ATTEMPT" = 1 ]; then
    printf '{"schema":"verdict/v1","task":"%s","verdict":"FAIL","criteria":[],"findings":[{"summary":"primary says no","kind":"bug"}],"notes":"primary"}\n' "$TASK" > .vloop/tmp/verdict.json
  fi
fi
`

func (r *runRepo) secondRun(mode string, flags ...string) result {
	r.t.Helper()
	r.commitFile("src.txt", "orig\n")
	for _, f := range flags {
		if err := os.WriteFile(filepath.Join(r.stub, f), nil, 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
	return r.runWith([]string{"VLOOP_RUN_SECOND_REVIEW=" + mode, "VLOOP_MODEL_SECOND_REVIEW=second-model"}, runBrief)
}

func (r *runRepo) secondSessions() []string {
	r.t.Helper()
	m, _ := filepath.Glob(filepath.Join(r.runFolder(), "sessions", "*-second-review.json"))
	return m
}

func TestRunSecondReviewRecord(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), secondScript)
	res := r.secondRun("record", "second.fail")
	wantExit(t, res, 0)
	r.wantTask("T1", "done", 0)
	log := r.runLog()
	wantIn(t, "run log", log, "\n   second review: FAIL (dissent)\n")
	wantIn(t, "run log", log, "second review: 1 dissent(s): T1 (primary PASS, second FAIL)")
	wantIn(t, "output", res.out, "second review: 1 dissent(s): T1 (primary PASS, second FAIL)")
	if got := r.read(".vloop/state/runs/" + runID + "/" + filepath.Base(r.runFolder()) + "/reports/001-second-verdict.json"); !strings.Contains(got, `"FAIL"`) {
		t.Errorf("second verdict not kept: %q", got)
	}
	if v := r.read(".vloop/state/runs/" + runID + "/" + filepath.Base(r.runFolder()) + "/reports/001-verdict.json"); !strings.Contains(v, `"primary"`) && !strings.Contains(v, `"PASS"`) {
		t.Errorf("primary verdict not kept: %q", v)
	}
	it := r.iterations()[0]
	sr, _ := it["second_review"].(map[string]any)
	if it["schema"] != "iteration/v3" || sr["verdict"] != "FAIL" || sr["dissent"] != true {
		t.Errorf("iteration record: %v", it)
	}
	s := r.secondSessions()
	if len(s) != 1 {
		t.Fatalf("%d second-review session records", len(s))
	}
	if v := r.vloop("schema", "validate", "session/v4", s[0]); v.code != 0 {
		t.Errorf("session record is not session/v4: %+v", v)
	}
	if log, _ := os.ReadFile(filepath.Join(r.stub, "second.log")); strings.Contains(string(log), "visible=1") {
		t.Errorf("the second review saw the primary's verdict: %s", log)
	}
	r.wantClean()
}

func TestRunSecondReviewOff(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), secondScript)
	res := r.secondRun("off", "second.fail")
	wantExit(t, res, 0)
	if len(r.secondSessions()) != 0 || strings.Contains(r.runLog(), "second review") || strings.Contains(res.out, "second review") {
		t.Errorf("a second review appears with the mode off")
	}
	if _, ok := r.iterations()[0]["second_review"]; ok {
		t.Errorf("iteration record has second_review")
	}
}

func TestRunSecondReviewOnPassSkipsAfterPrimaryFail(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), secondScript)
	wantExit(t, r.secondRun("on-pass", "primary.failfirst"), 0)
	r.wantIterations("T1:rejected", "T1:done")
	if n := len(r.secondSessions()); n != 1 {
		t.Errorf("%d second reviews, want 1 (after the primary PASS only)", n)
	}
	if n := r.task("T1")["notes"]; n != "" {
		t.Errorf("notes %v", n)
	}
}

func TestRunSecondReviewFeedbackJoinsNotes(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), secondScript+`if [ "$PHASE" = work ]; then echo "notes: $(jq -r '.tasks[0].notes' .vloop/state/state.json)" >> "$dir/work.log"; fi
`)
	wantExit(t, r.secondRun("feedback", "primary.failfirst", "second.fail"), 0)
	log, _ := os.ReadFile(filepath.Join(r.stub, "work.log"))
	if !strings.Contains(string(log), "primary says no; second review: second says no") {
		t.Errorf("the retry's notes: %s", log)
	}
}

func TestRunSecondReviewBothPassVeto(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), secondScript)
	wantExit(t, r.secondRun("both-pass", "second.fail"), 2)
	r.wantTask("T1", "blocked", 3)
	r.wantIterations("T1:rejected", "T1:rejected", "T1:rejected")
	if n := r.task("T1")["notes"]; n != "second review: second says no" {
		t.Errorf("notes %v", n)
	}
}

func TestRunSecondReviewChangedFileIsFailAndReverted(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), secondScript)
	wantExit(t, r.secondRun("record", "second.touch"), 0)
	r.wantTask("T1", "done", 0)
	sr, _ := r.iterations()[0]["second_review"].(map[string]any)
	if sr["verdict"] != "FAIL" {
		t.Errorf("second_review %v", sr)
	}
	if got := r.read("src.txt"); got != "orig\n" {
		t.Errorf("src.txt not reverted: %q", got)
	}
}
