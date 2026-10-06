package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const twoChecks = "[[check]]\nname = \"api\"\npaths = [\"api/**\"]\nrun = \"sh api/test.sh\"\n" +
	"[[check]]\nname = \"web\"\npaths = [\"web/**\"]\nrun = \"sh web/test.sh\"\n"

// checksRepo is a run repository with an api and a web check, both passing.
func checksRepo(t *testing.T, config string) *runRepo {
	t.Helper()
	r := newRunRepo(t)
	r.write("api/test.sh", "exit 0\n")
	r.write("web/test.sh", "exit 0\n")
	r.write(".vloop/config.toml", config)
	r.commitAll("checks")
	return r
}

func iterationChecks(rec map[string]any) string {
	var names []string
	for _, c := range rec["checks"].([]any) {
		names = append(names, c.(map[string]any)["name"].(string)+":"+jsonNum(c.(map[string]any)["exit"]))
	}
	return strings.Join(names, ",")
}

func jsonNum(v any) string { return fmt.Sprint(int(v.(float64))) }

func TestRunChecksScopedToWhatChanged(t *testing.T) {
	t.Parallel()
	r := checksRepo(t, twoChecks)
	r.scripted(planJSON(t,
		planTask("T1", map[string]any{"verify": "test -f api/T1.out"}),
		planTask("T2", map[string]any{"verify": "test -f web/T2.out", "depends_on": []string{"T1"}})),
		strings.Replace(defaultScript, `touch "$TASK.out"`, `d=api; [ "$TASK" = T2 ] && d=web; touch "$d/$TASK.out"`, 1))
	wantExit(t, r.vloop("run", runBrief), 0)
	its := r.iterations()
	if got := iterationChecks(its[0]); got != "api:0" {
		t.Errorf("T1 ran checks %q, want api:0", got)
	}
	if got := iterationChecks(its[1]); got != "web:0" {
		t.Errorf("T2 ran checks %q, want web:0", got)
	}
	for _, f := range []string{"api", "web", "final-api", "final-web"} {
		if _, err := os.Stat(filepath.Join(r.runFolder(), "checks", f+".log")); err != nil {
			t.Errorf("checks/%s.log: %v", f, err)
		}
	}
	r.wantClean()
}

func TestRunCheckFailedEndsIterationWithoutReview(t *testing.T) {
	t.Parallel()
	r := checksRepo(t, twoChecks)
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "test -f api/T1.out"})),
		strings.Replace(defaultScript, `touch "$TASK.out"`, `touch api/"$TASK.out"; echo "exit 1" > api/test.sh`, 1))
	wantExit(t, r.vloop("run", "--max-attempts", "1", runBrief), 2)
	r.wantIterations("T1:check_failed")
	if got := iterationChecks(r.iterations()[0]); got != "api:1" {
		t.Errorf("checks = %q, want api:1", got)
	}
	if n := countArgv(r, "-p /vloop:review"); n != 0 {
		t.Errorf("%d review session(s) ran after a failed check", n)
	}
	if _, err := os.Stat(filepath.Join(r.runFolder(), "checks", "001-api.fail.log")); err != nil {
		t.Errorf("no checks/001-api.fail.log: %v", err)
	}
	if n := r.task("T1")["notes"].(string); !strings.Contains(n, "api") || !strings.Contains(n, "checks/001-api.fail.log") {
		t.Errorf("task notes %q do not name the check and its log", n)
	}
	r.wantTask("T1", "blocked", 1)
}

func countArgv(r *runRepo, prefix string) int {
	n := 0
	for _, l := range r.argv() {
		if strings.HasPrefix(l, prefix) {
			n++
		}
	}
	return n
}

func TestRunChecksNotRunAfterFailedGateAndNeverRerun(t *testing.T) {
	t.Parallel()
	r := checksRepo(t, "[[check]]\nname = \"count\"\npaths = [\"**\"]\nrun = \"echo x >> .vloop/tmp/count; test $(wc -l < .vloop/tmp/count) -ne 2\"\n")
	r.scripted(planJSON(t, planTask("T1", nil)),
		defaultScript+`if [ "$PHASE:$ATTEMPT" = work:1 ]; then rm -f T1.out; fi`+"\n")
	wantExit(t, r.vloop("run", runBrief), 0)
	var outs []string
	for _, it := range r.iterations() {
		outs = append(outs, it["outcome"].(string))
	}
	if got := strings.Join(outs, ","); got != "gate_failed,check_failed,done" {
		t.Errorf("outcomes = %s, want gate_failed,check_failed,done", got)
	}
	if got := iterationChecks(r.iterations()[0]); got != "" {
		t.Errorf("an iteration whose gate failed ran checks %q", got)
	}
	// base, the failed iteration, the done iteration and the final pass: no re-run.
	if lines := strings.Count(r.read(".vloop/tmp/count"), "\n"); lines != 4 {
		t.Errorf("the check ran %d times, want 4", lines)
	}
}

func TestRunFinalPassFailureAndResume(t *testing.T) {
	t.Parallel()
	r := checksRepo(t, twoChecks)
	r.write("web/test.sh", "test ! -f api/T1.out\n")
	r.commitAll("web breaks when api/T1.out exists")
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "test -f api/T1.out"})),
		strings.Replace(defaultScript, `touch "$TASK.out"`, `touch api/"$TASK.out"`, 1))
	res := r.vloop("run", runBrief)
	wantExit(t, res, 2)
	wantIn(t, "stderr", res.err, "vloop: check web failed in the final pass — see "+
		strings.TrimPrefix(r.runFolder(), r.dir+string(filepath.Separator))+"/checks/final-web.log\n")
	if got := iterationChecks(r.iterations()[0]); got != "api:0" {
		t.Errorf("T1 ran checks %q, want only api:0", got)
	}
	if st := r.plan()["status"]; st != "blocked" {
		t.Errorf("plan status = %v, want blocked", st)
	}
	r.write("web/test.sh", "exit 0\n")
	r.commitAll("fix web")
	wantExit(t, r.vloop("run"), 0)
	if st := r.plan()["status"]; st != "complete" {
		t.Errorf("after the resume, plan status = %v, want complete", st)
	}
}
