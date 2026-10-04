package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorkedExampleB9 plays every line of B9's worked example with the stub
// claude, one subtest per line, each in its own temporary repository, with the
// sh shell and the api/web config the example gives. The per-feature tests pin
// the details; this one pins that the example holds as the brief wrote it.

const (
	b9Config = "shell = \"sh\"\nrun.gate-scratch = [\"web/.gate/\"]\n" + twoChecks
	// b9Verify1 and b9Verify2 are T1's and T2's gates: T2's copies the planner's
	// oracle into the scratch folder and runs it.
	b9Verify1 = "test -f api/T1.out"
	b9Verify2 = "mkdir -p web/.gate && cp .vloop/state/gates/T2/oracle.sh web/.gate/oracle.sh && sh web/.gate/oracle.sh"
)

// b9Hooks makes each task's work land in the directory its check watches, and
// the planner write T2's oracle.
func b9Hooks(script string) string {
	script = strings.Replace(script, `touch "$TASK.out"`, `d=api; [ "$TASK" = T2 ] && d=web; touch "$d/$TASK.out"`, 1)
	return script + `if [ "$PHASE" = plan ]; then mkdir -p .vloop/state/gates/T2; echo 'test -f web/T2.out' > .vloop/state/gates/T2/oracle.sh; fi
`
}

func b9Plan(t *testing.T, v1, v2 string) string {
	return planJSON(t, planTask("T1", map[string]any{"verify": v1}),
		planTask("T2", map[string]any{"verify": v2, "depends_on": []string{"T1"}}))
}

// b9Repo is the example's repository: api and web checks passing on the base,
// web/.gate/ ignored and declared as gate scratch, the stub playing the plan.
func b9Repo(t *testing.T) *runRepo {
	t.Helper()
	r := checksRepo(t, b9Config)
	r.write(".gitignore", r.read(".gitignore")+"web/.gate/\n")
	r.commitAll("b9 example")
	r.scripted(b9Plan(t, b9Verify1, b9Verify2), b9Hooks(defaultScript))
	return r
}

func (r *runRepo) wantNoSession() {
	r.t.Helper()
	if n := r.sessions(); n != 0 {
		r.t.Errorf("%d session(s) started", n)
	}
	if m, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", "*", "*", "sessions", "*")); len(m) != 0 {
		r.t.Errorf("session records exist: %v", m)
	}
}

func (r *runRepo) wantScratchEmpty() {
	r.t.Helper()
	if es, _ := os.ReadDir(filepath.Join(r.dir, "web", ".gate")); len(es) != 0 {
		r.t.Errorf("web/.gate/ holds %d entries", len(es))
	}
}

func TestWorkedExampleB9(t *testing.T) {
	t.Run("the run: base logs, gate review, scoped checks, empty scratch, final pass", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		wantExit(t, r.vloop("run", runBrief), 0)
		for _, id := range []string{"T1", "T2"} {
			if _, err := os.Stat(filepath.Join(r.runFolder(), "gates", "base-"+id+".log")); err != nil {
				t.Errorf("gates/base-%s.log: %v", id, err)
			}
		}
		if got := r.gateReviewState(); got != "1 PASS" || r.gateReviews() != 1 {
			t.Errorf("gate_review is %q after %d session(s), want 1 PASS after 1", got, r.gateReviews())
		}
		its := r.iterations()
		if len(its) != 2 {
			t.Fatalf("%d iterations, want 2", len(its))
		}
		if its[0]["task"] != "T1" || iterationChecks(its[0]) != "api:0" {
			t.Errorf("iteration 1: %v checks %q, want T1 api:0", its[0]["task"], iterationChecks(its[0]))
		}
		if its[1]["task"] != "T2" || iterationChecks(its[1]) != "web:0" {
			t.Errorf("iteration 2: %v checks %q, want T2 web:0", its[1]["task"], iterationChecks(its[1]))
		}
		r.wantScratchEmpty()
		for _, n := range []string{"api", "web"} {
			if _, err := os.Stat(filepath.Join(r.runFolder(), "checks", "final-"+n+".log")); err != nil {
				t.Errorf("checks/final-%s.log: %v", n, err)
			}
		}
		r.wantClean()
	})

	t.Run("no [[check]]: refused", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.write(".vloop/config.toml", "shell = \"sh\"\n")
		r.commitAll("no check")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: no check configured — add a [[check]] to .vloop/config.toml\n")
	})

	t.Run("web/test.sh exits 1 on the base: refused, no session record", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.write("web/test.sh", "exit 1\n")
		r.commitAll("web fails")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: check web fails on the base — fix it before planning: .vloop/state/runs/", "checks/base-web.log\n")
		r.wantNoSession()
	})

	t.Run("a check fixed after its base refusal: the next run is not refused for the refusal's logs", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.write("web/test.sh", "exit 1\n")
		r.commitAll("web fails")
		wantExit(t, r.vloop("run", runBrief), 1)
		r.write("web/test.sh", "exit 0\n")
		r.git("add", "web/test.sh") // the operator commits the fix, not the refusal's logs
		r.git("commit", "-q", "-m", "web fixed")
		wantExit(t, r.vloop("run", runBrief), 0)
	})

	t.Run("web/.gate/ not in .gitignore: refused", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.write(".gitignore", strings.Replace(r.read(".gitignore"), "web/.gate/\n", "", 1))
		r.commitAll("no ignore")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: gate scratch web/.gate/ is not ignored by git — add it to .gitignore\n")
		r.wantNoSession()
	})

	t.Run("T1's gate true passes on the base: fixed in round 2, the run proceeds", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.rounds(b9Plan(t, "true", b9Verify2), b9Plan(t, b9Verify1, b9Verify2))
		r.script(b9Hooks(strings.Replace(defaultScript, `cp "$dir/plan.json" .vloop/state/state.json`,
			`n=$(cat "$dir/pn" 2>/dev/null || echo 0); n=$((n+1)); echo $n > "$dir/pn"; cp "$dir/plan.$n.json" .vloop/state/state.json`, 1)))
		wantExit(t, r.vloop("run", runBrief), 0)
		if got := r.gateReviewState(); got != "2 PASS" {
			t.Errorf("gate_review is %q, want 2 PASS", got)
		}
		r.wantTask("T1", "done", 0)
	})

	t.Run("the stub gate reviewer fails twice: exit 2", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.script(b9Hooks(defaultScript) + failingGateReview)
		res := r.vloop("run", runBrief)
		wantExit(t, res, 2)
		wantIn(t, "stderr", res.err, "vloop: the gate review failed the plan twice — see .vloop/state/runs/",
			"/reports/gate-review-2.json; amend the gates with vloop task verify, then re-run")
	})

	t.Run("work edits the oracle: restored, T2's gate runs the planner's", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.script(b9Hooks(defaultScript) + `if [ "$PHASE" = work ] && [ "$TASK" = T2 ] && [ "$ATTEMPT" = 1 ]; then
  echo 'exit 0' > .vloop/state/gates/T2/oracle.sh; rm -f web/T2.out
fi
`)
		wantExit(t, r.vloop("run", runBrief), 0)
		r.wantIterations("T1:done", "T2:gate_failed", "T2:done")
		if got := r.read(".vloop/state/gates/T2/oracle.sh"); got != "test -f web/T2.out\n" {
			t.Errorf("the oracle is %q, want the planner's", got)
		}
	})

	t.Run("operator edits the oracle by hand, then task verify records it", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		wantExit(t, r.vloop("run", "--plan-only", runBrief), 0)
		old := r.task("T2")["fixtures"]
		r.write(".vloop/state/gates/T2/oracle.sh", "test -f web/T2.out && true\n")
		res := r.vloop("run")
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: the gate fixtures of T2 changed outside vloop task verify — record the change with vloop task verify T2 --reason '<why>'")

		wantExit(t, r.vloop("task", "verify", "T2", "--reason", "seed row"), 0)
		task := r.task("T2")
		hist, _ := task["gate_history"].([]any)
		if len(hist) != 1 {
			t.Fatalf("gate_history = %v, want one entry", task["gate_history"])
		}
		h := hist[0].(map[string]any)
		if h["fixtures"] != old || h["reason"] != "seed row" || h["by"] != "operator" {
			t.Errorf("gate_history[0] = %v, want fixtures %v, reason seed row, by operator", h, old)
		}
		if task["fixtures"] == old || task["fixtures"] == "" {
			t.Errorf("fixtures %v were not re-stamped from %v", task["fixtures"], old)
		}
		res = r.vloop("task", "verify", "T2", "--reason", "again")
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: nothing to record — T2's gate and fixtures are unchanged\n")
	})

	t.Run("T2's gate writes web/leftover.txt: the gate fails, the log says so", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.scripted(b9Plan(t, b9Verify1, b9Verify2+" && echo x > web/leftover.txt"), b9Hooks(defaultScript))
		wantExit(t, r.vloop("run", "--max-attempts", "1", runBrief), 2)
		b, err := os.ReadFile(filepath.Join(r.runFolder(), "gates", "T2.log"))
		if err != nil {
			t.Fatal(err)
		}
		if want := "vloop: gate T2 changed the tree — restored: web/leftover.txt\n"; !strings.HasSuffix(string(b), want) {
			t.Errorf("gate log ends %q, want %q", b, want)
		}
		r.wantScratchEmpty()
	})

	t.Run("T1's work breaks api/test.sh: check_failed, no review", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.script(b9Hooks(defaultScript) + `if [ "$PHASE" = work ] && [ "$TASK" = T1 ]; then echo 'exit 1' > api/test.sh; fi
`)
		wantExit(t, r.vloop("run", "--max-attempts", "1", runBrief), 2)
		r.wantIterations("T1:check_failed")
		if got := iterationChecks(r.iterations()[0]); got != "api:1" {
			t.Errorf("checks = %q, want api:1", got)
		}
		if n := countArgv(r, "-p /vloop:review"); n != 0 {
			t.Errorf("%d review session(s) ran", n)
		}
	})

	t.Run("web/test.sh fails only after T2: the final pass exits 2", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		// The web check runs on the base, after T2, and in the final pass; it
		// starts failing on its third run, from nothing T2's iteration changed.
		// A check that fails in T2's own iteration is check_failed, not this.
		r.write(".gitignore", r.read(".gitignore")+".web-runs\n")
		r.write("web/test.sh", "n=$(( $(cat .web-runs 2>/dev/null || echo 0) + 1 )); echo $n > .web-runs; [ $n -lt 3 ]\n")
		r.commitAll("web breaks on its third run")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 2)
		wantIn(t, "stderr", res.err, "vloop: check web failed in the final pass — see .vloop/state/runs/", "/checks/final-web.log\n")
	})

	t.Run("a check that fails on T2's work fails every retry", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.write("web/test.sh", "test ! -f web/T2.out\n")
		r.commitAll("web breaks once T2's file exists")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 2)
		its := r.iterations()
		for _, it := range its[1:] {
			if it["task"] != "T2" || it["outcome"] != "check_failed" {
				t.Errorf("iteration %v of %v is %v %v, want T2 check_failed", it["iteration"], len(its), it["task"], it["outcome"])
			}
		}
		if strings.Contains(res.err, "final pass") {
			t.Errorf("the final pass ran; T2 went done on the work its check failed:\n%s", res.err)
		}
	})

	// "a gate that sleeps past the gate timeout: runs once" is played by
	// TestRunFlakyGate's third subtest: the shortest timeout is a minute, and the
	// package's ten-minute limit has no room for a second one.

	t.Run("a complete state/v1 plan of an earlier brief: the next brief plans", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.write(".vloop/state/state.json", `{"schema":"state/v1","run_id":"B20250101-0900-old","brief":"docs/briefs/B20250101-0900-old.loop-brief.md","status":"complete","tasks":[]}`+"\n")
		r.commitAll("the v1 plan an earlier brief completed")
		res := r.vloop("run", runBrief)
		if res.code != 0 {
			t.Fatalf("exit %d\nstdout: %s\nstderr: %s", res.code, res.out, res.err)
		}
		if r.plan()["schema"] != "state/v2" {
			t.Errorf("the new plan is %v, want state/v2", r.plan()["schema"])
		}
	})

	t.Run("an unfinished state/v1 plan of another brief: refused", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.write(".vloop/state/state.json", `{"schema":"state/v1","run_id":"B20250101-0900-old","brief":"docs/briefs/B20250101-0900-old.loop-brief.md","status":"blocked","tasks":[]}`+"\n")
		r.commitAll("a v1 plan left blocked")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: .vloop/state/state.json is a state/v1 plan — finish it with vloop 1.x or re-plan the brief\n")
	})

	t.Run("state.json is state/v1: refused", func(t *testing.T) {
		t.Parallel()
		r := b9Repo(t)
		r.write(".vloop/state/state.json", "{\"schema\":\"state/v1\"}\n")
		r.commitAll("a v1 plan")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: .vloop/state/state.json is a state/v1 plan — finish it with vloop 1.x or re-plan the brief\n")
		r.wantNoSession()
	})
}
