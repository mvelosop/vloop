package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The iterating phase of `vloop run`, as the shell loop's scenarios 01, 02, 09,
// 16, 17, 21, 30, 37 and 40 assert it of .loop/run.sh, translated into vloop's
// layout, plus the checks the shell loop never made (HEAD, gate timing).

const defaultScript = `case "$PHASE" in
  plan) mkdir -p .vloop/state; cp "$dir/plan.json" .vloop/state/state.json;;
  work) touch "$TASK.out"; mkdir -p .vloop/tmp
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$TASK" "$TASK" > .vloop/tmp/proposal.json;;
  review) mkdir -p .vloop/tmp
    printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$TASK" > .vloop/tmp/verdict.json;;
esac
`

// scripted makes the stub play the plan, then script's own shell after the
// default's phase dispatch: the default for any phase script leaves alone.
func (r *runRepo) scripted(plan string, script string) {
	r.t.Helper()
	if err := os.WriteFile(filepath.Join(r.stub, "plan.json"), []byte(plan), 0o644); err != nil {
		r.t.Fatal(err)
	}
	r.script(script)
}

func (r *runRepo) runFolder() string {
	r.t.Helper()
	m, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", runID, "*"))
	if len(m) == 0 {
		r.t.Fatal("no run folder")
	}
	return m[len(m)-1]
}

func (r *runRepo) iterations() []map[string]any {
	r.t.Helper()
	var out []map[string]any
	m, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", runID, "*", "iterations.jsonl"))
	for _, f := range m {
		fh, err := os.Open(f)
		if err != nil {
			r.t.Fatal(err)
		}
		sc := bufio.NewScanner(fh)
		for sc.Scan() {
			var rec map[string]any
			if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
				r.t.Fatalf("iteration record %q: %v", sc.Text(), err)
			}
			rec["_line"] = sc.Text()
			out = append(out, rec)
		}
		fh.Close()
	}
	return out
}

func (r *runRepo) wantIterations(outcomes ...string) {
	r.t.Helper()
	its := r.iterations()
	if len(its) != len(outcomes) {
		r.t.Fatalf("%d iteration records, want %d: %v", len(its), len(outcomes), its)
	}
	for i, o := range outcomes {
		want := strings.SplitN(o, ":", 2) // task:outcome
		if its[i]["task"] != want[0] || its[i]["outcome"] != want[1] {
			r.t.Errorf("iteration %d is %v %v, want %s", i+1, its[i]["task"], its[i]["outcome"], o)
		}
		if int(its[i]["iteration"].(float64)) != i+1 {
			r.t.Errorf("iteration %d numbered %v", i+1, its[i]["iteration"])
		}
		f := filepath.Join(r.t.TempDir(), "it.json")
		os.WriteFile(f, []byte(its[i]["_line"].(string)), 0o644)
		if res := r.vloop("schema", "validate", "iteration/v2", f); res.code != 0 {
			r.t.Errorf("iteration record %d is not iteration/v2: %+v", i+1, res)
		}
	}
}

func (r *runRepo) plan() map[string]any {
	r.t.Helper()
	var p map[string]any
	if err := json.Unmarshal([]byte(r.read(".vloop/state/state.json")), &p); err != nil {
		r.t.Fatal(err)
	}
	return p
}

func (r *runRepo) task(id string) map[string]any {
	r.t.Helper()
	for _, t := range r.plan()["tasks"].([]any) {
		if m := t.(map[string]any); m["id"] == id {
			return m
		}
	}
	r.t.Fatalf("no task %s", id)
	return nil
}

func (r *runRepo) wantTask(id, status string, attempts int) {
	r.t.Helper()
	t := r.task(id)
	if t["status"] != status || int(t["attempts"].(float64)) != attempts {
		r.t.Errorf("%s is %v with %v attempt(s), want %s with %d", id, t["status"], t["attempts"], status, attempts)
	}
}

func (r *runRepo) journal() string { return r.read(".vloop/state/journals/" + runID + ".md") }

func (r *runRepo) wantClean() {
	r.t.Helper()
	if s := r.git("status", "--porcelain", "--untracked-files=all"); s != "" {
		r.t.Errorf("the working tree is not clean after the run:\n%s", s)
	}
	if s := r.git("ls-files", ".vloop/tmp"); s != "" {
		r.t.Errorf(".vloop/tmp/ was committed:\n%s", s)
	}
}

func twoTasks(t *testing.T) string {
	return planJSON(t, planTask("T1", nil), planTask("T2", map[string]any{"depends_on": []string{"T1"}}))
}

func TestRun01HappyPath(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	base := strings.TrimSpace(r.git("rev-parse", "main"))
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	all := r.outputs(res)
	wantIn(t, "output", all, "created and switched to branch "+runID, "plan complete")

	if got := strings.TrimSpace(r.git("rev-parse", "main")); got != base {
		t.Error("main moved")
	}
	p := r.plan()
	if p["status"] != "complete" || p["iteration"] != float64(2) || p["branch"] != runID {
		t.Errorf("plan: status %v iteration %v branch %v", p["status"], p["iteration"], p["branch"])
	}
	r.wantTask("T1", "done", 0)
	r.wantTask("T2", "done", 0)
	r.wantIterations("T1:done", "T2:done")
	its := r.iterations()
	for i, it := range its {
		g, _ := it["gate"].(map[string]any)
		if g == nil || g["exit"] != float64(0) || g["duration_ms"] == nil {
			t.Errorf("iteration %d gate = %v, want exit 0 with a duration", i+1, it["gate"])
		}
		if it["schema"] != "iteration/v2" || iterationChecks(it) != "all:0" {
			t.Errorf("iteration %d is %v with checks %q, want iteration/v2 with all:0", i+1, it["schema"], iterationChecks(it))
		}
		if it["run_id"] != runID || it["attempt"] != float64(1) {
			t.Errorf("iteration %d: run_id %v attempt %v", i+1, it["run_id"], it["attempt"])
		}
	}

	subs := r.subjects("main..HEAD")
	if len(subs) != 4 || subs[3] != "[vloop] plan "+runID || subs[2] != "[vloop] T1: done" || subs[1] != "[vloop] T2: done" ||
		!strings.HasPrefix(subs[0], "[vloop] run "+runID+"/") || !strings.HasSuffix(subs[0], ": complete") {
		t.Errorf("commit subjects: %q", subs)
	}
	dir := r.runFolder()
	if !strings.Contains(subs[0], "/"+filepath.Base(dir)+": complete") {
		t.Errorf("closing commit %q does not name the run folder %s", subs[0], filepath.Base(dir))
	}
	for _, f := range []string{"gates/T1.log", "gates/T2.log", "reports/001-verdict.json", "reports/002-verdict.json", "run.log", "iterations.jsonl"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f))); err != nil {
			t.Errorf("run folder lacks %s", f)
		}
	}
	for _, n := range []string{"001-verdict", "002-verdict"} {
		if v := r.vloop("schema", "validate", "verdict/v1", filepath.Join(dir, "reports", n+".json")); v.code != 0 {
			t.Errorf("%s is not verdict/v1: %+v", n, v)
		}
	}
	if r.git("ls-files", ".vloop/state/journals/"+runID+".md") == "" {
		t.Error("the journal is not committed")
	}
	if md := r.vloop("status", "--markdown"); md.out != r.read(".vloop/state/plan.md") {
		t.Error("plan.md is not what vloop status --markdown renders")
	}
	wantIn(t, "journal", r.journal(), "## T1 — Task T1", "## T2 — Task T2", "- **Outcome:** done (review: PASS)", "## Run ended — complete")
	if n := r.sessions(); n != 6 {
		t.Errorf("%d sessions, want 6 (plan, the gate review, then work and review twice)", n)
	}
	wantIn(t, "argv", strings.Join(r.argv(), "\n"), "-p /vloop:work T1 ", "-p /vloop:review T2 ")
	r.wantClean()
}

func TestRun02ReviewFail(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", nil)), defaultScript+`if [ "$PHASE" = review ] && [ ! -f "$dir/reviewed" ]; then
  : > "$dir/reviewed"
  printf '{"schema":"verdict/v1","task":"T1","verdict":"FAIL","criteria":[],"findings":[{"summary":"not really done","kind":"bug"}],"notes":"n"}\n' > .vloop/tmp/verdict.json
fi
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	r.wantIterations("T1:rejected", "T1:done")
	r.wantTask("T1", "done", 1)
	subs := strings.Join(r.subjects("main..HEAD"), "\n")
	wantIn(t, "subjects", subs, "[vloop] T1: rejected", "[vloop] T1: done")
	// the rejection reverted the task, charged an attempt and kept the finding
	rejected := strings.TrimSpace(r.git("log", "--format=%H", "--grep=^\\[vloop\\] T1: rejected$"))
	var snap struct {
		Tasks []struct{ Status, Notes string } `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(r.git("show", rejected+":.vloop/state/state.json")), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Tasks[0].Status != "pending" || snap.Tasks[0].Notes != "not really done" {
		t.Errorf("after the rejection the task is %+v", snap.Tasks[0])
	}
	if n := len(r.argv()); n < 1 {
		t.Error("no sessions")
	}
	r.wantClean()
}

func TestRun09DependencyOrder(t *testing.T) {
	t.Parallel()
	// T1 is listed first but depends on T2: file order and ready order disagree.
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", map[string]any{"depends_on": []string{"T2"}}), planTask("T2", nil)), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)
	r.wantIterations("T2:done", "T1:done")
}

func TestRun16ReviewFailsClosed(t *testing.T) {
	t.Parallel()
	for name, verdict := range map[string]string{
		"no verdict":                   `true`,
		"a verdict failing its schema": `printf '{"verdict":"PASS"}\n' > .vloop/tmp/verdict.json`,
	} {
		t.Run(name, func(t *testing.T) {
			r := newRunRepo(t)
			r.scripted(planJSON(t, planTask("T1", nil)), defaultScript+`if [ "$PHASE" = review ]; then rm -f .vloop/tmp/verdict.json; `+verdict+`; fi
`)
			res := r.vloop("run", "--max-attempts", "2", "--stall-limit", "99", runBrief)
			wantExit(t, res, 2)
			r.wantTask("T1", "blocked", 2)
			r.wantIterations("T1:rejected", "T1:rejected")
			all := r.outputs(res)
			wantIn(t, "output", all, "no valid verdict")
			wantNotIn(t, "output", all, "T1 done")
			if p := r.plan(); p["status"] != "blocked" {
				t.Errorf("plan status %v, want blocked", p["status"])
			}
			if s := r.subjects("-1")[0]; !strings.HasPrefix(s, "[vloop] run "+runID+"/") || !strings.HasSuffix(s, ": blocked") {
				t.Errorf("closing commit %q", s)
			}
			r.wantClean()
		})
	}
}

func TestRun17StaleHandoff(t *testing.T) {
	t.Parallel()
	// Only T1 reports. T2's session writes nothing, leaving T1's proposal on
	// disk as the most recent one: it must not be taken for T2's.
	r := newRunRepo(t)
	r.scripted(twoTasks(t), strings.Replace(defaultScript,
		`  work) touch "$TASK.out"; mkdir -p .vloop/tmp
    printf`, `  work) touch "$TASK.out"; mkdir -p .vloop/tmp
    [ "$TASK" = T1 ] && printf`, 1))
	res := r.vloop("run", "--max-attempts", "2", "--stall-limit", "99", runBrief)
	wantExit(t, res, 2)
	r.wantTask("T1", "done", 0)
	r.wantIterations("T1:done", "T2:blocked", "T2:blocked")
	wantIn(t, "output", r.outputs(res), "no valid proposal")
	j := r.journal()
	if n := strings.Count(j, "made T1"); n != 1 {
		t.Errorf("T1's report is in the journal %d times, want once (in its own entry):\n%s", n, j)
	}
	if !strings.Contains(j, "made T1") {
		t.Error("T1's report is missing from its entry")
	}
}

func TestRun21ForeignState(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)
	if p := r.plan(); p["branch"] != r.branch() {
		t.Errorf("the plan is stamped with branch %v, want %s", p["branch"], r.branch())
	}

	// A different brief is a different plan, so reset.
	const other = "B20260101-1000-other"
	r.write("docs/briefs/"+other+".loop-brief.md", strings.Replace(runBriefText(), runBriefName, other+".loop-brief", 1))
	r.commitAll("another brief")
	// The gates wait for a flag the work writes: T1.out and T2.out are already
	// committed, so a gate on them alone would pass on the base.
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "test -f other.flag"}),
		planTask("T2", map[string]any{"verify": "test -f other.flag", "depends_on": []string{"T1"}})),
		defaultScript+"if [ \"$PHASE\" = work ]; then touch other.flag; fi\n")
	res := r.vloop("run", "docs/briefs/"+other+".loop-brief.md")
	wantExit(t, res, 0)
	wantIn(t, "output", res.out+res.err, "resetting and planning fresh")
	otherDirs, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", other, "*", "sessions", "*-plan.json"))
	if len(otherDirs) != 1 {
		t.Error("no plan session ran: the old plan was resumed instead of reset")
	}
	if p := r.plan(); p["brief"] != "docs/briefs/"+other+".loop-brief.md" {
		t.Errorf("the plan's brief is %v, want the one asked for", p["brief"])
	}

	// Same brief, branch renamed: resume, never destroy.
	stamp := func() {
		var p map[string]any
		json.Unmarshal([]byte(r.read(".vloop/state/state.json")), &p)
		p["branch"] = "the-old-branch-name"
		b, _ := json.MarshalIndent(p, "", "  ")
		r.write(".vloop/state/state.json", string(b)+"\n")
	}
	stamp()
	before := r.plan()["run_id"]
	plans := len(otherDirs)
	res = r.vloop("run", "docs/briefs/"+other+".loop-brief.md")
	wantExit(t, res, 0)
	wantNotIn(t, "output", res.out+res.err, "resetting and planning fresh")
	if r.plan()["run_id"] != before {
		t.Error("renaming the branch destroyed the plan")
	}
	otherDirs, _ = filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", other, "*", "sessions", "*-plan.json"))
	if len(otherDirs) != plans {
		t.Error("it re-planned instead of resuming")
	}

	// No brief, and the stamp disagrees: refuse, do not guess.
	stamp()
	state := r.read(".vloop/state/state.json")
	res = r.vloop("run")
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "was made on branch the-old-branch-name")
	if r.read(".vloop/state/state.json") != state {
		t.Error("refusing changed the state")
	}
}

func TestRun30PlanOnly(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(planJSON(t,
		planTask("T1", nil),
		planTask("T2", map[string]any{"depends_on": []string{"T1"}})), defaultScript)

	res := r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "output", res.out, "plan only", "2 task(s)", "T1")
	if n := len(r.iterations()); n != 0 {
		t.Errorf("%d iterations ran, want 0", n)
	}
	if recs, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", runID, "*", "sessions", "*.json")); len(recs) != 2 {
		t.Errorf("%d session records, want two (the plan and the gate review)", len(recs))
	}
	for _, id := range []string{"T1", "T2"} {
		r.wantTask(id, "pending", 0)
	}
	if !r.has(".vloop/state/plan.md") {
		t.Error("plan.md was not rendered")
	}
	if s := r.subjects("-1"); len(s) != 1 || s[0] != "[vloop] plan "+runID {
		t.Errorf("last commit %q, want the plan's", s)
	}

	// The plan it left is resumable, which is the point.
	wantExit(t, r.vloop("run"), 0)
	if r.plan()["status"] != "complete" {
		t.Errorf("the resumed run ended %v, want complete", r.plan()["status"])
	}
	r.wantTask("T1", "done", 0)
	r.wantTask("T2", "done", 0)

	// --plan-only on an already-planned run does nothing, cheaply.
	before := r.read(".vloop/state/state.json")
	res = r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "output", res.out, "already planned")
	if r.read(".vloop/state/state.json") != before {
		t.Error("--plan-only modified an existing plan")
	}
}

func TestRun37GateTaskEnv(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	envLog := filepath.Join(r.stub, "env.log")
	rec := func(id string) map[string]any {
		return planTask(id, map[string]any{"verify": `test -f ` + id + `.out && printf "%s %s\n" "${VLOOP_ACTIVE_TASK-unset}" "${VLOOP_GATE_TASK-unset}" >> "` + envLog + `"`})
	}
	t2 := rec("T2")
	t2["depends_on"] = []string{"T1"}
	r.scripted(planJSON(t, rec("T1"), t2), defaultScript+`printf "%s %s %s\n" "$PHASE" "${VLOOP_ACTIVE_TASK-unset}" "${VLOOP_GATE_TASK-unset}" >> "`+envLog+`"
`)
	// A stale value in the driver's own environment must reach no session.
	res := r.exec(binPath, r.env("VLOOP_ACTIVE_TASK=stale", "VLOOP_GATE_TASK=stale"), "run", runBrief)
	wantExit(t, res, 0)
	r.wantTask("T1", "done", 0)
	r.wantTask("T2", "done", 0)
	b, err := os.ReadFile(envLog)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	count := func(l string) int {
		n := 0
		for _, x := range lines {
			if x == l {
				n++
			}
		}
		return n
	}
	for _, l := range []string{"T1 T1", "T2 T1", "T2 T2"} {
		if count(l) == 0 {
			t.Errorf("no gate saw active/gate = %s; saw %q", l, lines)
		}
	}
	if n := count("work unset unset"); n != 2 {
		t.Errorf("%d work sessions saw neither variable, want 2: %q", n, lines)
	}
	if n := count("review unset unset"); n != 2 {
		t.Errorf("%d review sessions saw neither variable, want 2: %q", n, lines)
	}
	if n := count("plan unset unset"); n != 1 {
		t.Errorf("the plan session saw a gate variable: %q", lines)
	}
}

func TestRun40NoProposalTree(t *testing.T) {
	t.Parallel()
	run := func(t *testing.T, edit bool) (r *runRepo, notes, journal, out string) {
		r = newRunRepo(t)
		script := `case "$PHASE" in
  plan) mkdir -p .vloop/state; cp "$dir/plan.json" .vloop/state/state.json;;
  review) mkdir -p .vloop/tmp; printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$TASK" > .vloop/tmp/verdict.json;;
`
		if edit {
			script += `  work) echo a > thing_one.txt; echo b > thing_two.txt;;
`
		}
		script += "esac\n"
		r.scripted(planJSON(t, planTask("T1", nil)), script)
		res := r.vloop("run", "--max-iterations", "1", runBrief)
		wantExit(t, res, 4)
		r.wantIterations("T1:blocked")
		r.wantTask("T1", "pending", 1)
		if p := r.plan(); p["status"] != "halted" {
			t.Errorf("plan status %v, want halted", p["status"])
		}
		end := res.out[strings.Index(res.out, "═══"):]
		return r, r.task("T1")["notes"].(string), r.journal(), end
	}

	t.Run("the session edited two files and died", func(t *testing.T) {
		r, notes, journal, end := run(t, true)
		for _, f := range []string{"thing_one.txt", "thing_two.txt"} {
			wantIn(t, "notes", notes, f)
			wantIn(t, "journal", journal, f)
			wantIn(t, "run-end report", end, f)
			if r.git("ls-files", "--error-unmatch", f) == "" {
				t.Errorf("%s was not committed", f)
			}
		}
		wantNotIn(t, "notes", notes, ".vloop/")
		wantIn(t, "subjects", strings.Join(r.subjects("main..HEAD"), "\n"), "[vloop] T1: blocked")
	})
	t.Run("the session did nothing and died", func(t *testing.T) {
		_, notes, _, _ := run(t, false)
		wantIn(t, "notes", notes, "no valid proposal")
		wantNotIn(t, "notes", notes, "thing_", ".vloop/")
	})
}

func TestRunMaxIterations(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", "--max-iterations", "1", runBrief), 4)
	r.wantIterations("T1:done")
	if p := r.plan(); p["status"] != "halted" {
		t.Errorf("plan status %v, want halted", p["status"])
	}
	if s := r.subjects("-1")[0]; !strings.HasSuffix(s, ": halted") {
		t.Errorf("closing commit %q", s)
	}
	r.wantClean()
	// Raising the budget resumes.
	wantExit(t, r.vloop("run", runBrief), 0)
	r.wantTask("T2", "done", 0)
}

func TestRunGateFailure(t *testing.T) {
	t.Parallel()
	// A gate that fails once, then passes: no review for the failed iteration,
	// one attempt charged, and the retry is not another task's.
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", nil)), defaultScript+`if [ "$PHASE" = work ] && [ ! -f "$dir/once" ]; then : > "$dir/once"; rm -f T1.out; fi
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	r.wantIterations("T1:gate_failed", "T1:done")
	if g := r.iterations()[0]["gate"].(map[string]any); g["exit"] == float64(0) {
		t.Errorf("the failed iteration's gate = %v", g)
	}
	r.wantTask("T1", "done", 1)
	if n := strings.Count(strings.Join(r.argv(), "\n"), "/vloop:review T1"); n != 1 {
		t.Errorf("%d reviews, want 1: none for the iteration whose gate failed", n)
	}
}

func TestRunHeadCheck(t *testing.T) {
	t.Parallel()
	// A gate that moves HEAD off the run's branch: nothing may be committed.
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "test -f T1.out && git checkout -q main"})), defaultScript)
	base := strings.TrimSpace(r.git("rev-parse", "main"))
	res := r.vloop("run", runBrief)
	wantExit(t, res, 9)
	for _, s := range r.subjects("--all") {
		if strings.HasPrefix(s, "[vloop] T1") || strings.HasPrefix(s, "[vloop] run ") {
			t.Errorf("a commit was made after HEAD left the run branch: %q", s)
		}
	}
	if got := strings.TrimSpace(r.git("rev-parse", "main")); got != base {
		t.Error("main received a commit")
	}
	wantIn(t, "stderr", res.err, "nothing was committed")
}
