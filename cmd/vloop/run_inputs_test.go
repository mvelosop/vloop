package main

import (
	"strings"
	"testing"
)

// The driver keeps its inputs in memory and puts back what a session changed
// among them; a review session that changes one fails.

func TestRunCostRecordsZeroed(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", nil), planTask("T2", map[string]any{"depends_on": []string{"T1"}}),
		planTask("T3", map[string]any{"depends_on": []string{"T2"}})), defaultScript+`if [ "$PHASE" = work ] && [ "$TASK" != T1 ]; then
  for f in .vloop/state/runs/*/*/sessions/*.json; do sed 's/"cost_usd": [0-9.]*/"cost_usd": 0/' "$f" > "$f.z" && mv "$f.z" "$f"; done
fi
`)
	// plan, work and review of T1 cost 0.30; T2 ends at 0.50 with the earlier records zeroed, so T3 must not start.
	res := r.runWith(nil, "--cost-ceiling", "0.35", runBrief)
	wantExit(t, res, 6)
}

func TestRunSessionInputsRestored(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	orig := r.read(".vloop/config.toml")
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then printf '\n[model]\nreview = "haiku"\n' >> .vloop/config.toml; fi
`)
	wantExit(t, r.runWith(nil, runBrief), 0)
	if got := r.read(".vloop/config.toml"); got != orig {
		t.Errorf("config not restored: %q", got)
	}
	wantIn(t, "run log", r.runLog(), "vloop: work session changed .vloop/config.toml — restored")
	if !strings.Contains(strings.Join(r.argv(), "\n"), "/vloop:review T1 --model sonnet") {
		t.Errorf("the review did not run with the configured model: %v", r.argv())
	}
}

func TestRunReviewInputsFail(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	orig := r.read(".vloop/config.toml")
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = review ] && [ "$ATTEMPT" = 1 ]; then printf '\n# edited\n' >> .vloop/config.toml; fi
`)
	wantExit(t, r.runWith(nil, runBrief), 0)
	wantIn(t, "run log", r.runLog(), "vloop: review session changed .vloop/config.toml — restored")
	r.wantIterations("T1:rejected", "T1:done")
	if got := r.read(".vloop/config.toml"); got != orig {
		t.Errorf("config not restored: %q", got)
	}
}
