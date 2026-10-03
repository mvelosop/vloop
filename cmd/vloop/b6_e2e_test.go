package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// B6's close: the brief's worked example on the stub claude, each planted
// failure in its own copy of the fixture, and `vloop run` on a throwaway clone
// of this repository.

// b6Repo is the harness repository configured as the worked example: the
// review on opus and work at high effort, committed.
func b6Repo(t *testing.T) *runRepo {
	t.Helper()
	r := newRunRepo(t)
	for _, kv := range [][2]string{{"model.review", "opus"}, {"effort.work", "high"}} {
		if res := r.vloop("config", "set", kv[0], kv[1]); res.code != 0 {
			t.Fatalf("config set %s: %+v", kv[0], res)
		}
	}
	r.commitAll("fixture config")
	return r
}

// b6Lines is the stub's argument lines for sessions whose prompt starts with prefix.
func b6Lines(r *runRepo, prefix string) []string {
	var out []string
	for _, l := range r.argv() {
		if strings.HasPrefix(l, "-p "+prefix) {
			out = append(out, l)
		}
	}
	return out
}

func b6Squash(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestWorkedExampleB6(t *testing.T) {
	r := b6Repo(t)
	r.scripted(twoTasks(t), defaultScript)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	out := b6Squash(strings.ReplaceAll(res.out, "\n", " | "))
	wantIn(t, "stdout", out, "created and switched to branch "+runID, runID+" ready · not merged",
		"tasks 2 planned", "· 2 done · 0 blocked · first-pass 2/2")

	folder := filepath.Base(r.runFolder())
	want := []string{"[vloop] run " + runID + "/" + folder + ": complete", "[vloop] T2: done", "[vloop] T1: done", "[vloop] plan " + runID}
	if got := r.subjects("main..HEAD"); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("commit subjects = %q, want %q", got, want)
	}

	st := r.vloop("status")
	wantExit(t, st, 0)
	if first := strings.SplitN(st.out, "\n", 2)[0]; first != runID+" — complete · 2/2 done · 0 blocked · iteration 2" {
		t.Errorf("vloop status first line = %q", first)
	}
	r.wantClean()
}

func TestWorkedExampleB6Sessions(t *testing.T) {
	r := b6Repo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)

	fence := regexp.MustCompile(`--settings \.vloop/tmp/fence/([^/ ]+)/work\.json`)
	plan := b6Lines(r, "/vloop:plan "+runBrief)
	if len(plan) != 1 || !strings.Contains(plan[0], "--model opus") || strings.Contains(plan[0], "--effort") {
		t.Errorf("plan session = %q, want one with --model opus and no --effort", plan)
	}
	work := b6Lines(r, "/vloop:work T1")
	if len(work) != 1 {
		t.Fatalf("work sessions = %q", work)
	}
	if !strings.Contains(work[0], "--model sonnet") || !strings.Contains(work[0], "--effort high") {
		t.Errorf("work session = %q, want --model sonnet --effort high", work[0])
	}
	m := fence.FindStringSubmatch(work[0])
	if m == nil {
		t.Fatalf("work session has no --settings .vloop/tmp/fence/<v>/work.json: %q", work[0])
	}
	if !strings.Contains(work[0], "--plugin-dir .vloop/tmp/plugin/"+m[1]) {
		t.Errorf("work session = %q, want --plugin-dir .vloop/tmp/plugin/%s", work[0], m[1])
	}
	review := b6Lines(r, "/vloop:review T1")
	if len(review) != 1 || !strings.Contains(review[0], "--model opus") || strings.Contains(review[0], "--effort") {
		t.Errorf("review session = %q, want one with --model opus and no --effort", review)
	}

	sess := filepath.Join(r.runFolder(), "sessions", "002-work.json")
	if res := r.vloop("schema", "validate", "session/v1", sess); res.code != 0 {
		t.Errorf("002-work.json is not session/v1: %+v", res)
	}
	g, ok := r.iterations()[0]["gate"].(map[string]any)
	if !ok || len(g) != 2 || g["exit"] != float64(0) {
		t.Fatalf("the first iteration's gate = %v, want {exit:0, duration_ms:<n>}", r.iterations()[0]["gate"])
	}
	if _, ok := g["duration_ms"].(float64); !ok {
		t.Errorf("gate duration_ms = %v, want a number", g["duration_ms"])
	}

	res := r.vloop("metrics", runBriefName)
	wantExit(t, res, 0)
	if !regexp.MustCompile(`(?m)^\s*time\s.*gates [0-9][0-9.,]*( min)? · wall`).MatchString(res.out) {
		t.Errorf("vloop metrics has no gate minutes on the time line:\n%s", res.out)
	}
}

func TestWorkedExampleB6BranchExists(t *testing.T) {
	r := b6Repo(t)
	r.scripted(twoTasks(t), defaultScript)
	r.git("branch", runID)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 1)
	if want := "vloop: branch " + runID + " exists — switch to it and re-run\n"; res.err != want {
		t.Errorf("stderr = %q, want %q", res.err, want)
	}
}

func TestWorkedExampleB6GateDispute(t *testing.T) {
	r := b6Repo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$PHASE" = work ] && [ "$TASK" = T2 ]; then
  printf '{"schema":"proposal/v1","task":"T2","outcome":"blocked","summary":"s","files":[],"verified":"v","notes":"n","gate_dispute":{"reason":"r","evidence":"e"}}\n' > .vloop/tmp/proposal.json
fi
`)
	wantExit(t, r.vloop("run", runBrief), 2)
	r.wantTask("T2", "blocked", 0)
	if n := r.task("T2")["notes"]; n != "gate disputed: r — e" {
		t.Errorf("T2 notes = %q", n)
	}
	if n := len(b6Lines(r, "/vloop:review T2")); n != 0 {
		t.Errorf("T2 was reviewed %d time(s)", n)
	}
}

func TestWorkedExampleB6FlakyGate(t *testing.T) {
	r := b6Repo(t)
	seen := filepath.Join(r.stub, "flaky.seen")
	verify := `test -f T1.out && { [ -f "` + seen + `" ] || { : > "` + seen + `"; exit 1; }; }`
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": verify}), planTask("T2", map[string]any{"depends_on": []string{"T1"}})), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)
	wantIn(t, "run log", r.runLog(), "FLAKY GATE T1")
	if g := r.iterations()[0]["gate"].(map[string]any); g["flaky"] != true {
		t.Errorf("iteration 1 gate = %v, want flaky: true", g)
	}
	r.wantTask("T1", "done", 0)

	res := r.vloop("defect", "list", "--matrix", "--json", "--brief", runBriefName)
	wantExit(t, res, 0)
	var matrix [][]int
	if err := json.Unmarshal([]byte(res.out), &matrix); err != nil || len(matrix) < 4 || len(matrix[3]) < 1 || matrix[3][0] != 1 {
		t.Errorf("defect matrix = %q (%v), want one env defect at [3][0]", res.out, err)
	}
	wantExit(t, r.vloop("metrics", runBriefName), 0)
}

func TestWorkedExampleB6SilentReview(t *testing.T) {
	r := b6Repo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$PHASE" = review ] && [ "$TASK" = T1 ]; then STUB_SILENT=1; fi
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	wantIn(t, "run log", r.runLog(), "SESSION RECORD MISSING review T1 (iteration 1)")
	for _, n := range r.sessionFiles() {
		var rec map[string]any
		sessionRecord(t, filepath.Join(r.runFolder(), "sessions", n), &rec)
		if rec["phase"] == "review" && rec["iteration"] == float64(1) {
			t.Errorf("%s: a record exists for the review that printed nothing", n)
		}
	}
	res := r.vloop("metrics", runBriefName)
	wantExit(t, res, 0)
	if !regexp.MustCompile(`(?m)^\s*records\s`).MatchString(res.out) {
		t.Errorf("vloop metrics has no records line:\n%s", res.out)
	}
}

func TestWorkedExampleB6RefsMoved(t *testing.T) {
	r := b6Repo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$PHASE" = work ] && [ "$TASK" = T1 ]; then git checkout -q -b other; fi
`)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 9)
	wantIn(t, "output", r.outputs(res), "REFS MOVED T1")
	for _, s := range r.subjects("--all") {
		if strings.HasPrefix(s, "[vloop] T1") || strings.HasPrefix(s, "[vloop] run ") {
			t.Errorf("committed after refs moved: %q", s)
		}
	}
}

func TestWorkedExampleB6MaxIterations(t *testing.T) {
	r := b6Repo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", "--max-iterations", "1", runBrief), 4)
	if s := r.plan()["status"]; s != "halted" {
		t.Errorf("plan status = %v, want halted", s)
	}
	wantExit(t, r.vloop("run"), 0)
	if s := r.plan()["status"]; s != "complete" {
		t.Errorf("plan status after the re-run = %v, want complete", s)
	}
}

// TestWorkedExampleB6RealData runs vloop on a throwaway clone of this
// repository, which has the shell loop's .loop/, B4's snapshots and no
// .vloop/tmp/ ignore line, and checks this repository is untouched.
func TestWorkedExampleB6RealData(t *testing.T) {
	src := newRunRepo(t) // for its stub, home and helpers; its own repository is unused
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	status := func() string { return runGit(t, root, "status", "--porcelain", "--untracked-files=all") }
	before := status()

	clone := filepath.Join(t.TempDir(), "clone")
	runGit(t, root, "clone", "-q", root, clone)
	runGit(t, clone, "config", "user.name", "gate")
	runGit(t, clone, "config", "user.email", "gate@example.com")
	runGit(t, clone, "config", "commit.gpgsign", "false")
	real, _ := filepath.EvalSymlinks(clone)
	trust := map[string]any{}
	for _, k := range []string{clone, real} {
		trust[k] = map[string]any{"hasTrustDialogAccepted": true}
	}
	raw, _ := json.Marshal(map[string]any{"projects": trust})
	if err := os.WriteFile(filepath.Join(src.home, ".claude.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}

	r := &runRepo{t: t, dir: clone, home: src.home, stub: src.stub}
	// The clone carries this repository's live plan, which may be a state/v1
	// one that vloop run refuses; the run below plans a fresh brief.
	if err := os.Remove(filepath.Join(clone, ".vloop", "state", "state.json")); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	r.write(".vloop/config.toml", r.read(".vloop/config.toml")+"\n"+runCheckConfig)
	r.write("docs/x.md", "x\n")
	r.write("docs/briefs/"+runBriefName+".md", runBriefText())
	r.commitAll("fixture brief")
	r.scripted(oneTask(t), defaultScript)

	wantExit(t, r.vloop("run", runBrief), 0)
	res := r.vloop("metrics", runBriefName)
	wantExit(t, res, 0)
	if !regexp.MustCompile(`(?m)^\s*tasks\s+1 planned.* · 1 done · 0 blocked`).MatchString(res.out) {
		t.Errorf("vloop metrics did not read the run:\n%s", res.out)
	}
	if s := r.git("ls-files", ".vloop/tmp"); s != "" {
		t.Errorf(".vloop/tmp/ was committed in the clone:\n%s", s)
	}
	if after := status(); after != before {
		t.Errorf("the real-data check changed this repository:\n%s", after)
	}
}
