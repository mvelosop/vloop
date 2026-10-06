package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The plan phase of `vloop run`, as the shell loop's planning scenarios assert
// it of .loop/run.sh, translated into vloop's layout. 21 and 30 also run
// iterations and are ported with them.

const (
	runID    = "B20260101-0900-demo"
	runBrief = "docs/briefs/" + runBriefName + ".md"
)

// planTask is a valid state/v2 task; over replaces fields.
func planTask(id string, over map[string]any) map[string]any {
	t := map[string]any{"id": id, "title": "Task " + id, "goal": "Do " + id + ".", "kind": "feature",
		"fixtures": "", "references": []any{}, "depends_on": []string{},
		"acceptance": []string{id + ".out exists"}, "verify": "test -f " + id + ".out",
		"status": "pending", "attempts": 0, "notes": ""}
	for k, v := range over {
		t[k] = v
	}
	return t
}

// planJSON is a state/v2 plan for runBrief, as a planning session writes it.
func planJSON(t *testing.T, tasks ...map[string]any) string {
	t.Helper()
	b, err := json.MarshalIndent(map[string]any{"schema": "state/v2", "run_id": runID, "brief": runBrief,
		"base": "", "branch": "", "status": "planning", "iteration": 0,
		"created": "2026-01-01T09:00:00Z", "updated": "2026-01-01T09:00:00Z", "shell": "sh",
		"checks": []any{map[string]any{"name": "all", "paths": []string{"**"}, "run": "true"}}, "gate_scratch": []string{}, "gate_review": map[string]any{"rounds": 0, "verdict": ""}, "tasks": tasks}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b) + "\n"
}

// planWith makes the stub's plan session write the given plan.
func (r *runRepo) planWith(plan string) {
	r.t.Helper()
	if err := os.WriteFile(filepath.Join(r.stub, "plan.json"), []byte(plan), 0o644); err != nil {
		r.t.Fatal(err)
	}
	r.script(`if [ "$PHASE" = plan ]; then mkdir -p .vloop/state; cp "$dir/plan.json" .vloop/state/state.json; fi
`)
}

func (r *runRepo) commitAll(msg string) {
	r.t.Helper()
	r.git("add", "-A")
	r.git("commit", "-q", "-m", msg)
}

func (r *runRepo) branch() string {
	return strings.TrimSpace(r.git("rev-parse", "--abbrev-ref", "HEAD"))
}

// sessions counts the claude sessions the stub was started for.
func (r *runRepo) sessions() int {
	n := 0
	for _, l := range r.argv() {
		if strings.HasPrefix(l, "-p ") {
			n++
		}
	}
	return n
}

func (r *runRepo) has(rel string) bool {
	_, err := os.Stat(filepath.Join(r.dir, filepath.FromSlash(rel)))
	return err == nil
}

func (r *runRepo) subjects(rng string) []string {
	out := strings.TrimSpace(r.git("log", "--format=%s", rng))
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// planCommits counts the plan commits on any branch.
func (r *runRepo) planCommits() int {
	n := 0
	for _, s := range r.subjects("--all") {
		if strings.HasPrefix(s, "[vloop] plan ") {
			n++
		}
	}
	return n
}

// runDirs lists the run folders of the brief.
func (r *runRepo) runDirs() []string {
	m, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", runID, "*"))
	return m
}

// outputs is everything a run printed or logged.
func (r *runRepo) outputs(res result) string {
	all := res.out + res.err
	for _, d := range r.runDirs() {
		if b, err := os.ReadFile(filepath.Join(d, "run.log")); err == nil {
			all += string(b)
		}
	}
	return all
}

func wantExit(t *testing.T, res result, code int) {
	t.Helper()
	if res.code != code {
		t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", res.code, code, res.out, res.err)
	}
}

func wantIn(t *testing.T, what, text string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(text, s) {
			t.Errorf("%s lacks %q:\n%s", what, s, text)
		}
	}
}

func wantNotIn(t *testing.T, what, text string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if strings.Contains(text, s) {
			t.Errorf("%s has %q:\n%s", what, s, text)
		}
	}
}

// planRejected asserts a run whose plan is refused: exit 1, nothing committed,
// no plan left behind, and the session that wrote it the only one started.
func planRejected(t *testing.T, r *runRepo, res result) {
	t.Helper()
	wantExit(t, res, 1)
	if n := r.planCommits(); n != 0 {
		t.Errorf("a plan that was refused was committed (%d plan commits)", n)
	}
	if r.has(".vloop/state/state.json") {
		t.Error("a refused plan was left in .vloop/state/state.json")
	}
	if n := r.sessions(); n != 1 {
		t.Errorf("%d sessions were started, want only the plan session", n)
	}
	for _, d := range r.runDirs() {
		if r.has(strings.TrimPrefix(filepath.ToSlash(d), filepath.ToSlash(r.dir)+"/") + "/iterations.jsonl") {
			t.Error("a refused plan started iterations")
		}
	}
}

func TestRun08PlanValidation(t *testing.T) {
	t.Parallel()
	// A plan with no verify command or no acceptance is a planning failure,
	// caught before any iteration inherits it.
	for name, tc := range map[string]struct {
		task map[string]any
		want []string
	}{
		"no verify":     {planTask("T1", map[string]any{"verify": ""}), []string{"T1", "verify"}},
		"no acceptance": {planTask("T1", map[string]any{"acceptance": []string{}}), []string{"T1", "acceptance"}},
		"unknown dependency": {planTask("T1", map[string]any{"depends_on": []string{"T9"}}),
			[]string{"T1", "T9"}},
	} {
		t.Run(name, func(t *testing.T) {
			r := newRunRepo(t)
			r.planWith(planJSON(t, tc.task))
			res := r.vloop("run", "--plan-only", runBrief)
			planRejected(t, r, res)
			wantIn(t, "stderr", res.err, tc.want...)
		})
	}

	// The same plan with a verify command and acceptance is accepted.
	r := newRunRepo(t)
	r.planWith(planJSON(t, planTask("T1", nil)))
	wantExit(t, r.vloop("run", "--plan-only", runBrief), 0)
}

func TestRun25GateShape(t *testing.T) {
	t.Parallel()
	// A gate that parses a structure and then substring-matches its
	// re-serialised text has thrown away the parse; so is one that asserts on
	// source text. Refused before a single iteration is spent.
	bad := map[string]string{
		"re-serialised": `node -e "const d=require('./doc.json');if(JSON.stringify(d.paths).indexOf('url')<0)process.exit(1)"`,
		"source text":   `grep -q needle src/app.go`,
	}
	for name, verify := range bad {
		t.Run(name, func(t *testing.T) {
			r := newRunRepo(t)
			r.planWith(planJSON(t, planTask("T1", map[string]any{"verify": verify})))
			res := r.vloop("run", "--plan-only", runBrief)
			planRejected(t, r, res)
			wantIn(t, "stderr", res.err, "gate shape rejected", "T1")
			if got := r.outputs(res); !strings.Contains(got, "gate shape rejected") {
				t.Errorf("the run log does not carry the rejection:\n%s", got)
			}
		})
	}
	// Matching text that was never parsed is fine.
	r := newRunRepo(t)
	r.planWith(planJSON(t, planTask("T1", map[string]any{"verify": `test -f T1.out && grep -q ok T1.out`})))
	res := r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 0)
	wantNotIn(t, "output", r.outputs(res), "gate shape rejected")
}

func TestRun27DanglingReference(t *testing.T) {
	t.Parallel()
	// A reference that does not resolve costs a session an attempt to discover;
	// it is checked at plan time. One that does resolve is not flagged.
	r := newRunRepo(t)
	r.planWith(planJSON(t, planTask("T1", map[string]any{"references": []map[string]string{
		{"path": "docs/x.md", "why": "the error contract"},
		{"path": "docs/never-written.md", "why": "a document nobody wrote"}}})))
	res := r.vloop("run", "--plan-only", runBrief)
	planRejected(t, r, res)
	wantIn(t, "stderr", res.err, "T1", "docs/never-written.md", "does not resolve")
	wantNotIn(t, "stderr", res.err, "docs/x.md")

	// A folder reference with no way in is a warning, not a refusal.
	r = newRunRepo(t)
	r.write("docs/blind/a.md", "a\n")
	r.commitAll("a folder with no index")
	r.planWith(planJSON(t, planTask("T1", map[string]any{"references": []map[string]string{
		{"path": "docs/blind", "why": "a folder"}}})))
	res = r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "stderr", res.err, "folder reference", "T1", "docs/blind")
}

func TestRun29BriefTypo(t *testing.T) {
	t.Parallel()
	// A typo must not be able to destroy a committed plan.
	r := newRunRepo(t)
	plan := planJSON(t, planTask("T1", map[string]any{"status": "done", "attempts": 1}))
	plan = strings.Replace(plan, `"status": "planning"`, `"status": "complete"`, 1)
	r.write(".vloop/state/state.json", plan)
	r.write(".vloop/state/plan.md", "rendered\n")
	r.commitAll("a finished plan")
	head := r.git("rev-parse", "HEAD")

	untouched := func(what string) {
		t.Helper()
		if got := r.read(".vloop/state/state.json"); got != plan {
			t.Errorf("%s modified the plan", what)
		}
		if r.read(".vloop/state/plan.md") != "rendered\n" {
			t.Errorf("%s modified plan.md", what)
		}
		if r.git("rev-parse", "HEAD") != head || r.branch() != "main" || r.git("status", "--porcelain") != "" {
			t.Errorf("%s changed the repository", what)
		}
	}

	res := r.vloop("run", "docs/briefs/"+runBriefName+"-typo.md")
	wantExit(t, res, 1)
	if want := "vloop: brief not found: docs/briefs/" + runBriefName + "-typo.md\n"; res.err != want {
		t.Errorf("stderr = %q, want %q", res.err, want)
	}
	if r.sessions() != 0 || len(r.runDirs()) != 0 || r.has(".vloop/state/runs") {
		t.Error("a missing brief started a session or made a run folder")
	}
	untouched("a mistyped brief path")

	res = r.vloop("run", "--help")
	wantExit(t, res, 0)
	wantIn(t, "--help", res.out, "vloop run", "--plan-only", "--replan", "--max-iterations", "--cost-ceiling", "--max-attempts", "--stall-limit")
	untouched("--help")

	res = r.vloop("run", "--resume")
	wantExit(t, res, 2)
	wantIn(t, "stderr", res.err, "unknown flag")
	untouched("an unknown flag")
	if r.sessions() != 0 {
		t.Error("a refused invocation started a session")
	}
}

func TestRun36GateDecayingBaseline(t *testing.T) {
	t.Parallel()
	// Rule 4: a gate may not diff, log or rev-list against a ref other than
	// HEAD — that baseline decays the moment another task commits. A range that
	// merely ends in HEAD is still rejected; HEAD itself is fine.
	reject := []struct{ verify, ref string }{
		{`test -z "$(git diff --name-only 4c911d7 -- docs)"`, "4c911d7"},
		{`test -z "$(git diff --name-only HEAD~1 -- docs)"`, "HEAD~1"},
		{`test -z "$(git log --format=%H v1.0..HEAD -- docs)"`, "v1.0"},
		{`test "$(git rev-list --count origin/main..HEAD)" -gt 0`, "origin/main"},
		{`git diff --quiet "$(git merge-base HEAD main)" -- docs`, ""},
	}
	for _, c := range reject {
		r := newRunRepo(t)
		r.planWith(planJSON(t, planTask("T1", map[string]any{"verify": c.verify})))
		res := r.vloop("run", "--plan-only", runBrief)
		planRejected(t, r, res)
		wantIn(t, "stderr", res.err, "gate shape rejected")
		if !regexp.MustCompile(`T1 .*` + regexp.QuoteMeta(c.ref)).MatchString(res.err) {
			t.Errorf("%s: stderr does not name T1 and %q:\n%s", c.verify, c.ref, res.err)
		}
	}
	for _, verify := range []string{
		`test -f T1.out && test -z "$(git diff --name-only HEAD -- docs)"`,
		`test -f T1.out && git diff --quiet HEAD -- docs`,
	} {
		r := newRunRepo(t)
		r.planWith(planJSON(t, planTask("T1", map[string]any{"verify": verify})))
		res := r.vloop("run", "--plan-only", runBrief)
		wantExit(t, res, 0)
		wantNotIn(t, verify, r.outputs(res), "gate shape rejected")
	}
}

func TestRun38HeadDiffAdvisory(t *testing.T) {
	t.Parallel()
	// A gate that diffs against HEAD without reading VLOOP_ACTIVE_TASK or
	// VLOOP_GATE_TASK reads another task's in-progress edits as its own
	// regression. An advisory: it names the task and the plan is still committed.
	plan := func(verify string) *runRepo {
		r := newRunRepo(t)
		r.planWith(planJSON(t, planTask("T1", map[string]any{"verify": "test -f T1.out"}),
			planTask("T2", map[string]any{"verify": verify})))
		return r
	}
	advised := []string{
		`test -f T2.out && test -z "$(git diff --name-only HEAD -- docs)"`,
		`test -f T2.out && git diff --quiet HEAD -- docs`,
	}
	quiet := []string{
		`test -f T2.out && { [ "$VLOOP_GATE_TASK" != "$VLOOP_ACTIVE_TASK" ] || test -z "$(git diff --name-only HEAD -- docs)"; }`,
		`test -f T2.out && test -n "${VLOOP_ACTIVE_TASK:-}" && git diff --quiet HEAD -- docs`,
		`test -f T2.out && test -n "${VLOOP_GATE_TASK:-}" && git diff --quiet HEAD -- docs`,
		`test -f T2.out`,
	}
	for _, v := range advised {
		r := plan(v)
		res := r.vloop("run", "--plan-only", runBrief)
		wantExit(t, res, 0)
		if !regexp.MustCompile(`warning: T2 diff against HEAD`).MatchString(res.err) {
			t.Errorf("%s: no advisory naming T2:\n%s", v, res.err)
		}
		wantNotIn(t, v+": stderr", res.err, "T1 diff", "gate shape rejected")
		if r.planCommits() != 1 {
			t.Errorf("%s: the advisory stopped the plan from being committed", v)
		}
	}
	for _, v := range quiet {
		r := plan(v)
		res := r.vloop("run", "--plan-only", runBrief)
		wantExit(t, res, 0)
		wantNotIn(t, v+": stderr", res.err, "diff against HEAD")
	}
}

func TestRun43BriefAlreadyRun(t *testing.T) {
	t.Parallel()
	// Nothing retires a brief, so a spent one reads plannable forever. Its
	// journal's existence is the check: the run refuses before touching the
	// plan or starting a session, unless --replan.
	const fresh = "docs/briefs/B20260101-1000-fresh.loop-brief.md"
	setup := func() *runRepo {
		r := newRunRepo(t)
		r.write(fresh, strings.ReplaceAll(runBriefText(), runBriefName, "B20260101-1000-fresh.loop-brief"))
		r.write(".vloop/state/journals/"+runID+".md", "# Journal\n")
		r.commitAll("a journal and a second brief")
		r.planWith(planJSON(t, planTask("T1", nil)))
		return r
	}
	const journal = ".vloop/state/journals/" + runID + ".md"

	// vloop brief check reports an already-run brief and not a fresh one.
	r := setup()
	chk := r.vloop("brief", "check", runBrief)
	wantExit(t, chk, 1)
	wantIn(t, "brief check", chk.out+chk.err, journal)
	if chk := r.vloop("brief", "check", fresh); chk.code != 0 {
		t.Errorf("a never-run brief fails brief check: %s%s", chk.out, chk.err)
	}

	res := r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, journal, "--replan")
	if r.has(".vloop/state/state.json") || r.sessions() != 0 || r.branch() != "main" {
		t.Error("the refusal planned, started a session or moved the branch")
	}
	if r.has(".vloop/state/runs") {
		t.Error("the refusal made a run folder")
	}

	// It refuses before an existing plan for another brief is reset.
	r = setup()
	other := strings.ReplaceAll(planJSON(t, planTask("T1", nil)), runBrief, fresh)
	r.write(".vloop/state/state.json", other)
	r.commitAll("the plan of another brief")
	res = r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 1)
	if r.read(".vloop/state/state.json") != other {
		t.Error("the refusal destroyed or changed the existing plan")
	}

	// --replan overrides it.
	r = setup()
	res = r.vloop("run", "--plan-only", "--replan", runBrief)
	wantExit(t, res, 0)
	if !r.has(".vloop/state/state.json") || r.planCommits() != 1 {
		t.Error("no plan was written under --replan")
	}

	// A brief that was never run plans normally.
	r = setup()
	r.planWith(strings.ReplaceAll(strings.ReplaceAll(planJSON(t, planTask("T1", nil)), runBrief, fresh), runID, "B20260101-1000-fresh"))
	res = r.vloop("run", "--plan-only", fresh)
	wantExit(t, res, 0)
	if !r.has(".vloop/state/state.json") {
		t.Error("a never-run brief was not planned")
	}
}

func TestRunWorkBranch(t *testing.T) {
	t.Parallel()
	// R-2: on the default branch the run creates <run id> and switches to it.
	r := newRunRepo(t)
	r.planWith(planJSON(t, planTask("T1", nil)))
	main := strings.TrimSpace(r.git("rev-parse", "main"))
	res := r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 0)
	if !strings.Contains(res.out, "created and switched to branch "+runID+"\n") {
		t.Errorf("stdout lacks the creation line:\n%s", res.out)
	}
	if r.branch() != runID {
		t.Errorf("HEAD is on %s, want %s", r.branch(), runID)
	}
	if strings.TrimSpace(r.git("rev-parse", "main")) != main {
		t.Error("main moved")
	}
	if got := r.subjects("main..HEAD"); len(got) != 1 || got[0] != "[vloop] plan "+runID {
		t.Errorf("commits on the run branch: %q", got)
	}

	// The branch already exists: refused, nothing changed.
	r = newRunRepo(t)
	r.planWith(planJSON(t, planTask("T1", nil)))
	r.git("branch", runID)
	res = r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 1)
	if want := "vloop: branch " + runID + " exists — switch to it and re-run\n"; res.err != want {
		t.Errorf("stderr = %q, want %q", res.err, want)
	}
	if r.branch() != "main" || r.sessions() != 0 || r.has(".vloop/state/state.json") || r.has(".vloop/state/runs") {
		t.Error("the refusal changed something (branch, sessions, plan or run folder)")
	}

	// Any other branch: the run stays there.
	r = newRunRepo(t)
	r.planWith(planJSON(t, planTask("T1", nil)))
	r.git("checkout", "-q", "-b", "feature")
	res = r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 0)
	wantNotIn(t, "output", res.out+res.err, "created and switched")
	if r.branch() != "feature" || r.git("branch", "--list", runID) != "" {
		t.Error("on branch feature the run must stay there and create no branch")
	}
	if got := r.subjects("-1"); got[0] != "[vloop] plan "+runID {
		t.Errorf("the plan commit is not on feature: %q", got)
	}
	var plan struct{ Branch string }
	if err := json.Unmarshal([]byte(r.read(".vloop/state/state.json")), &plan); err != nil || plan.Branch != "feature" {
		t.Errorf("the plan is stamped %q, want feature", plan.Branch)
	}
}

func TestRunPlanCommit(t *testing.T) {
	t.Parallel()
	// What the plan phase leaves: the stamped plan, plan.md, the journal, the
	// plan session's record, and one commit holding them — never .vloop/tmp.
	r := newRunRepo(t)
	r.planWith(planJSON(t, planTask("T1", nil), planTask("T2", map[string]any{"depends_on": []string{"T1"}})))
	r.write(".vloop/config.toml", "[model]\nplan = \"opus\"\n\n"+runCheckConfig)
	r.commitAll("config")
	res := r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 0)

	var plan struct {
		Schema, Brief, Branch, Status string
		RunID                         string `json:"run_id"`
		Iteration                     int
		Tasks                         []struct{ Status string }
	}
	if err := json.Unmarshal([]byte(r.read(".vloop/state/state.json")), &plan); err != nil {
		t.Fatal(err)
	}
	if plan.Schema != "state/v2" || plan.RunID != runID || plan.Brief != runBrief || plan.Branch != runID ||
		plan.Status != "running" || plan.Iteration != 0 || len(plan.Tasks) != 2 {
		t.Errorf("the stamped plan: %+v", plan)
	}
	if v := r.vloop("task", "validate"); v.code != 0 {
		t.Errorf("task validate: %s%s", v.out, v.err)
	}
	if md := r.vloop("status", "--markdown"); md.out != r.read(".vloop/state/plan.md") {
		t.Error("plan.md is not what vloop status --markdown renders")
	}
	j := r.read(".vloop/state/journals/" + runID + ".md")
	wantIn(t, "the journal", j, "# Journal — "+runID, "## Plan — "+runID, runBrief)

	dirs := r.runDirs()
	if len(dirs) != 1 || !regexp.MustCompile(`^\d{8}-\d{6}$`).MatchString(filepath.Base(dirs[0])) {
		t.Fatalf("run folders: %v", dirs)
	}
	rec, err := os.ReadFile(filepath.Join(dirs[0], "sessions", "001-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Schema    string
		RunID     string `json:"run_id"`
		Iteration int
		Phase     string
		Model     string
		Task      *string
	}
	if err := json.Unmarshal(rec, &s); err != nil || s.Schema != "session/v1" || s.RunID != runID ||
		s.Iteration != 0 || s.Phase != "plan" || s.Model != "opus" || s.Task != nil {
		t.Errorf("the plan session record: %s (%v)", rec, err)
	}
	if v := r.vloop("schema", "validate", "session/v1", filepath.ToSlash(filepath.Join(".vloop/state/runs", runID, filepath.Base(dirs[0]), "sessions/001-plan.json"))); v.code != 0 {
		t.Errorf("the record does not validate: %s%s", v.out, v.err)
	}

	if got := r.subjects("main..HEAD"); len(got) != 1 || got[0] != "[vloop] plan "+runID {
		t.Fatalf("commits: %q", got)
	}
	for _, f := range []string{".vloop/state/state.json", ".vloop/state/plan.md", ".vloop/state/journals/" + runID + ".md",
		".vloop/state/runs/" + runID + "/" + filepath.Base(dirs[0]) + "/sessions/001-plan.json",
		".vloop/state/runs/" + runID + "/" + filepath.Base(dirs[0]) + "/run.log"} {
		if err := exec1(r, "cat-file", "-e", "HEAD:"+f); err != "" {
			t.Errorf("the plan commit lacks %s: %s", f, err)
		}
	}
	if tmp := r.git("ls-files", ".vloop/tmp"); tmp != "" {
		t.Errorf(".vloop/tmp was committed: %s", tmp)
	}

	// With the plan in place, --plan-only changes nothing.
	head := r.git("rev-parse", "HEAD")
	n := r.sessions()
	res = r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 0)
	if st := r.git("status", "--porcelain"); r.git("rev-parse", "HEAD") != head || r.sessions() != n || st != "" {
		t.Errorf("--plan-only on a planned run changed something: %q", st)
	}
	wantIn(t, "stdout", res.out, "already planned")

	// Without --plan-only the plan is in place but nothing runs it yet.
	if res := r.vloop("run", "--max-iterations", "x"); res.code != 2 || !strings.Contains(res.err, "--max-iterations") {
		t.Errorf("a bad budget flag: exit %d, %s", res.code, res.err)
	}
}

// exec1 runs git and returns "" on success, else its output.
func exec1(r *runRepo, args ...string) string {
	res := r.exec("git", r.env(), append([]string{"-C", r.dir}, args...)...)
	if res.code != 0 {
		return res.out + res.err
	}
	return ""
}

func TestRunPreflight(t *testing.T) {
	t.Parallel()
	// Preflight is doctor's checks, and refuses before any branch, session or
	// file change — an untrusted workspace included.
	r := newRunRepo(t)
	r.planWith(planJSON(t, planTask("T1", nil)))
	if err := os.Remove(filepath.Join(r.home, ".claude.json")); err != nil {
		t.Fatal(err)
	}
	res := r.vloop("run", "--plan-only", runBrief)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "trust", "preflight failed")
	if r.branch() != "main" || r.sessions() != 0 || r.has(".vloop/state/runs") || r.has(".vloop/state/state.json") {
		t.Error("a refused preflight changed something")
	}
}

// rounds makes the stub's plan sessions write the given plans in turn, and its
// work sessions do the default work.
func (r *runRepo) rounds(plans ...string) {
	r.t.Helper()
	for i, p := range plans {
		if err := os.WriteFile(filepath.Join(r.stub, fmt.Sprintf("plan.%d.json", i+1)), []byte(p), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
	r.script(strings.Replace(defaultScript, `cp "$dir/plan.json" .vloop/state/state.json`,
		`n=$(cat "$dir/pn" 2>/dev/null || echo 0); n=$((n+1)); echo $n > "$dir/pn"; cp "$dir/plan.$n.json" .vloop/state/state.json`, 1))
}

func (r *runRepo) planSessions() int {
	n := 0
	for _, l := range r.argv() {
		if strings.HasPrefix(l, "-p /vloop:plan ") {
			n++
		}
	}
	return n
}

func TestRunBaseGatePasses(t *testing.T) {
	t.Parallel()
	// A gate that passes before the work exists proves nothing: it goes back to
	// the plan session, once, and the revised plan runs.
	r := newRunRepo(t)
	r.rounds(planJSON(t, planTask("T1", map[string]any{"verify": "true"})),
		planJSON(t, planTask("T1", nil)))
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "output", r.outputs(res), "gate T1 passes on the base")
	if n := r.planSessions(); n != 2 {
		t.Errorf("%d plan sessions ran, want 2", n)
	}
	if !r.has(strings.TrimPrefix(filepath.ToSlash(r.runFolder()), filepath.ToSlash(r.dir)+"/") + "/gates/base-T1.log") {
		t.Error("no gates/base-T1.log in the run folder")
	}
	r.wantTask("T1", "done", 0)
}

func TestRunBaseGateChangesTree(t *testing.T) {
	t.Parallel()
	// A gate that writes the tree on the base is sent back, and what it wrote
	// does not survive.
	r := newRunRepo(t)
	r.rounds(planJSON(t, planTask("T1", map[string]any{"verify": "echo x >stray.txt; test -f T1.out"})),
		planJSON(t, planTask("T1", nil)))
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "output", r.outputs(res), "gate T1 changed the tree on the base")
	if n := r.planSessions(); n != 2 {
		t.Errorf("%d plan sessions ran, want 2", n)
	}
	if r.has("stray.txt") {
		t.Error("stray.txt, written by a gate on the base, survived")
	}
}

func TestRunBaseGateStillPassesAfterRevision(t *testing.T) {
	t.Parallel()
	// Two rounds are all there are: a second draft that still has the problem
	// halts, and nothing is committed.
	r := newRunRepo(t)
	bad := planJSON(t, planTask("T1", map[string]any{"verify": "true"}))
	r.rounds(bad, bad)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "gate T1 passes on the base")
	if n := r.planSessions(); n != 2 {
		t.Errorf("%d plan sessions ran, want 2", n)
	}
	if n := r.planCommits(); n != 0 {
		t.Errorf("%d plan commits, want 0", n)
	}
}

func TestRunSchemaFailureHasNoSecondRound(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	bad := strings.Replace(planJSON(t, planTask("T1", nil)), "state/v2", "state/v9", 1)
	r.rounds(bad, planJSON(t, planTask("T1", nil)))
	res := r.vloop("run", runBrief)
	wantExit(t, res, 1)
	if n := r.planSessions(); n != 1 {
		t.Errorf("%d plan sessions ran, want 1", n)
	}
}
