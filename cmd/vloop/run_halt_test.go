package main

import (
	"strings"
	"testing"
)

// Halting and budgets, as the shell loop's scenarios 04, 05, 06, 07, 11, 19, 42
// and 44 assert them of .loop/run.sh, translated into vloop's layout.

// reviewFails makes every review session say FAIL.
const reviewFails = `if [ "$PHASE" = review ]; then
  printf '{"schema":"verdict/v1","task":"%s","verdict":"FAIL","criteria":[],"findings":["never good enough"],"notes":"none"}\n' "$TASK" > .vloop/tmp/verdict.json
fi
`

// blockedWork makes every work session report blocked, saying "diagnosis <n>",
// after running before.
func blockedWork(before string) string {
	return `if [ "$PHASE" = work ]; then
  rm -f "$TASK.out"
` + before + `  printf '{"schema":"proposal/v1","task":"%s","outcome":"blocked","summary":"diagnosis %s","files":[],"verified":"none","notes":"diagnosis %s"}\n' "$TASK" "$ATTEMPT" "$ATTEMPT" > .vloop/tmp/proposal.json
fi
`
}

func (r *runRepo) runWith(env []string, args ...string) result {
	r.t.Helper()
	return r.exec(binPath, r.env(env...), append([]string{"run"}, args...)...)
}

func (r *runRepo) wantStatus(status string) {
	r.t.Helper()
	if got := r.plan()["status"]; got != status {
		r.t.Errorf("plan status %v, want %s", got, status)
	}
}

func oneTask(t *testing.T) string { return planJSON(t, planTask("T1", nil)) }

// TestRun04AttemptCeiling: a task that keeps failing review is blocked, not
// retried forever.
func TestRun04AttemptCeiling(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+reviewFails)
	res := r.runWith([]string{"VLOOP_RUN_MAX_ATTEMPTS=2", "VLOOP_RUN_CONVERGENCE_MIN=99"}, runBrief)
	wantExit(t, res, 2)
	r.wantStatus("blocked")
	r.wantTask("T1", "blocked", 2)
	wantIn(t, "run log", r.runLog(), "attempt ceiling")
	r.wantClean()
}

// TestRun05MaxIterationsResumable: budgets are per run and checked between
// iterations, so raising one and re-running just works with no state edit.
func TestRun05MaxIterationsResumable(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	res := r.runWith([]string{"VLOOP_RUN_MAX_ITERATIONS=1"}, runBrief)
	wantExit(t, res, 4)
	r.wantStatus("halted")
	r.wantTask("T1", "done", 0)
	r.wantTask("T2", "pending", 0)
	wantNotIn(t, "stderr", res.err, "vloop: ", "line ")
	if s := r.subjects("-1")[0]; s != "[vloop] run "+runID+"/"+r.runFolderName()+": halted" {
		t.Errorf("closing commit %q", s)
	}
	r.wantClean()

	res = r.vloop("run")
	wantExit(t, res, 0)
	wantNotIn(t, "stderr", res.err, "vloop: ", "line ")
	r.wantStatus("complete")
	r.wantTask("T2", "done", 0)
	if n := strings.Count(strings.Join(r.argv(), "\n"), "/vloop:plan"); n != 1 {
		t.Errorf("%d plan sessions, the resume must not plan again", n)
	}
	if n := len(r.runDirs()); n != 2 {
		t.Errorf("%d run folders, want 2: the resume has its own", n)
	}
}

func (r *runRepo) runFolderName() string {
	r.t.Helper()
	f := r.runFolder()
	return f[strings.LastIndex(f, "/")+1:]
}

// TestRun06CostCeilingResumable: the same promise for the cost ceiling.
func TestRun06CostCeilingResumable(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), "STUB_COST=1.00\n"+defaultScript)
	wantExit(t, r.runWith(nil, "--cost-ceiling", "2", runBrief), 6)
	r.wantStatus("halted")
	r.wantTask("T2", "pending", 0)
	wantIn(t, "run log", r.runLog(), "cost ceiling")
	r.wantClean()

	wantExit(t, r.runWith(nil, "--cost-ceiling", "99"), 0)
	r.wantStatus("complete")
	r.wantTask("T2", "done", 0)
}

// TestRun07ConvergenceHalt: every gate is green and every review thorough, and
// nothing ever closes.
func TestRun07ConvergenceHalt(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+reviewFails)
	wantExit(t, r.runWith([]string{"VLOOP_RUN_CONVERGENCE_MIN=2", "VLOOP_RUN_MAX_ATTEMPTS=99"}, runBrief), 5)
	r.wantStatus("halted")
	wantIn(t, "run log", r.runLog(), "not converging")
	r.wantClean()
}

// TestRun11Stall: a work session that keeps reporting blocked makes no recorded
// progress. Each attempt touches a fresh marker so the repeat-blocked halt never
// fires: something changes every time.
func TestRun11Stall(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+blockedWork(`  : > "attempt-$ATTEMPT.marker"
`))
	wantExit(t, r.runWith([]string{"VLOOP_RUN_MAX_ATTEMPTS=99", "VLOOP_RUN_CONVERGENCE_MIN=99"}, runBrief), 3)
	r.wantStatus("stalled")
	wantIn(t, "run log", r.runLog(), "no recorded progress")
	r.wantClean()
}

// TestRun19SessionError: a claude session that dies is an infrastructure
// failure, not the task's: no attempt is charged.
func TestRun19SessionError(t *testing.T) {
	t.Parallel()
	for name, script := range map[string]string{
		"exits non-zero": `if [ "$PHASE" = work ]; then STUB_EXIT=9; rm -f .vloop/tmp/proposal.json; else :; fi
`,
		"reports is_error": `if [ "$PHASE" = work ]; then STUB_ERROR=true; fi
`,
	} {
		t.Run(name, func(t *testing.T) {
			r := newRunRepo(t)
			body := `case "$PHASE" in
  plan) mkdir -p .vloop/state; cp "$dir/plan.json" .vloop/state/state.json;;
esac
` + script
			r.scripted(oneTask(t), body)
			res := r.runWith(nil, runBrief)
			wantExit(t, res, 7)
			r.wantStatus("halted")
			r.wantTask("T1", "pending", 0)
			if n := r.reviewed("T1"); n != 0 {
				t.Errorf("%d review(s) after a work session that failed", n)
			}
			wantIn(t, "run log", r.runLog(), "failed to run")
		})
	}
}

// TestRun42RepeatBlockedHalts: a task that blocks twice with nothing changed
// between the attempts halts and surfaces the FIRST diagnosis; when the first
// session changed something, the retry is allowed and the rule is re-evaluated
// on the next pair.
func TestRun42RepeatBlockedHalts(t *testing.T) {
	t.Parallel()
	plan := func(t *testing.T) string {
		return planJSON(t, planTask("T1", nil))
	}
	env := []string{"VLOOP_RUN_STALL_LIMIT=9", "VLOOP_RUN_MAX_ATTEMPTS=5", "VLOOP_RUN_CONVERGENCE_MIN=99"}
	endOf := func(res result) string {
		i := strings.Index(res.out, "═══")
		if i < 0 {
			t.Fatalf("no run end in stdout: %s", res.out)
		}
		return res.out[i:]
	}

	t.Run("blocked twice on identical inputs", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(plan(t), defaultScript+blockedWork(""))
		res := r.runWith(env, runBrief)
		wantExit(t, res, 8)
		r.wantIterations("T1:blocked", "T1:blocked")
		r.wantTask("T1", "pending", 2)
		r.wantStatus("halted")
		wantIn(t, "run end", endOf(res), "diagnosis 1")
		wantNotIn(t, "run end", endOf(res), "diagnosis 2")
	})

	t.Run("the first attempt changes something", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(plan(t), defaultScript+blockedWork(`  [ "$ATTEMPT" = 1 ] && echo changed > thing.txt
`))
		res := r.runWith(env, runBrief)
		wantExit(t, res, 8)
		r.wantIterations("T1:blocked", "T1:blocked", "T1:blocked")
		r.wantTask("T1", "pending", 3)
		wantIn(t, "run end", endOf(res), "diagnosis 2")
		wantNotIn(t, "run end", endOf(res), "diagnosis 1", "diagnosis 3")
	})
}

// TestRun44RefsMovedHalts: a session must not move git refs. The driver
// snapshots every ref and HEAD before each session and compares after; any
// difference halts with exit 9 and commits nothing.
func TestRun44RefsMovedHalts(t *testing.T) {
	t.Parallel()
	loopCommits := func(r *runRepo, prefix string) int {
		n := 0
		for _, s := range r.subjects("--all") {
			if strings.HasPrefix(s, prefix) {
				n++
			}
		}
		return n
	}
	moving := func(t *testing.T, phase, cmd string) (*runRepo, result) {
		r := newRunRepo(t)
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = `+phase+` ]; then `+cmd+`; fi
`)
		return r, r.runWith(nil, runBrief)
	}

	t.Run("the planning session renames the branch", func(t *testing.T) {
		r, res := moving(t, "plan", `git branch -m "$(git symbolic-ref --short HEAD)" renamed-by-session`)
		wantExit(t, res, 9)
		all := r.outputs(res)
		wantIn(t, "output", all, "REFS MOVED plan", "refs/heads/renamed-by-session")
		if n := r.planCommits(); n != 0 {
			t.Errorf("the plan was committed onto moved refs (%d)", n)
		}
	})

	t.Run("the work session plants a remote ref and repoints origin/HEAD", func(t *testing.T) {
		r, res := moving(t, "work", `git update-ref refs/remotes/origin/trunk HEAD && git symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/trunk`)
		wantExit(t, res, 9)
		wantIn(t, "output", r.outputs(res), "REFS MOVED T1", "refs/remotes/origin/trunk")
		if n := loopCommits(r, "[vloop] T1"); n != 0 {
			t.Errorf("the iteration was committed after refs moved (%d)", n)
		}
		if n := loopCommits(r, "[vloop] run "); n != 0 {
			t.Errorf("the run's closing commit was made after refs moved (%d)", n)
		}
	})

	t.Run("the review session checks out a new branch", func(t *testing.T) {
		r, res := moving(t, "review", `git checkout -q -b work`)
		wantExit(t, res, 9)
		wantIn(t, "output", r.outputs(res), "REFS MOVED T1", "refs/heads/work")
		if n := loopCommits(r, "[vloop] T1"); n != 0 {
			t.Errorf("the iteration was committed onto the new branch (%d)", n)
		}
		if n := loopCommits(r, "[vloop] run "); n != 0 {
			t.Errorf("the run's closing commit was made after refs moved (%d)", n)
		}
		if w, b := strings.TrimSpace(r.git("rev-list", "--count", "work")), strings.TrimSpace(r.git("rev-list", "--count", runID)); w != b {
			t.Errorf("commits landed on the branch the session created: work has %s, %s has %s", w, runID, b)
		}
	})

	t.Run("control: a session that leaves refs alone runs to completion", func(t *testing.T) {
		r, res := moving(t, "none", `true`)
		wantExit(t, res, 0)
		wantNotIn(t, "output", r.outputs(res), "REFS MOVED")
	})
}

// TestRunBudgetPrecedence: a flag beats the environment, which beats the file,
// which beats the default.
func TestRunBudgetPrecedence(t *testing.T) {
	t.Parallel()
	fresh := func(t *testing.T, fileValue string) *runRepo {
		r := newRunRepo(t)
		r.scripted(twoTasks(t), defaultScript)
		if fileValue != "" {
			if res := r.vloop("config", "set", "run.max-iterations", fileValue); res.code != 0 {
				t.Fatalf("config set: %+v", res)
			}
			r.git("add", "-A")
			r.git("commit", "-q", "-m", "harness: budget")
		}
		return r
	}
	t.Run("the file beats the default", func(t *testing.T) {
		wantExit(t, fresh(t, "1").runWith(nil, runBrief), 4)
	})
	t.Run("the environment beats the file", func(t *testing.T) {
		wantExit(t, fresh(t, "1").runWith([]string{"VLOOP_RUN_MAX_ITERATIONS=5"}, runBrief), 0)
	})
	t.Run("the flag beats the environment", func(t *testing.T) {
		wantExit(t, fresh(t, "").runWith([]string{"VLOOP_RUN_MAX_ITERATIONS=5"}, "--max-iterations", "1", runBrief), 4)
	})
	t.Run("the default allows the whole plan", func(t *testing.T) {
		wantExit(t, fresh(t, "").runWith(nil, runBrief), 0)
	})
}
