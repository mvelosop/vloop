package runs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const briefPath = "docs/briefs/B1-x.loop-brief.md"

func put(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunID(t *testing.T) {
	for _, in := range []string{briefPath, "B1-x.loop-brief", "B1-x", "./" + briefPath} {
		if got := RunID(in); got != "B1-x" {
			t.Errorf("RunID(%q) = %q", in, got)
		}
	}
	if got := BriefPath(t.TempDir(), "B1-x"); got != briefPath {
		t.Errorf("BriefPath(name) = %q", got)
	}
}

func shellFixture(t *testing.T) string {
	root := t.TempDir()
	base := ".loop/state/runs/B1-x/20260101-100000/"
	put(t, root, base+"loop.log", "\x1b[36m[loop]\x1b[0m planning from "+briefPath+" using opus\n")
	put(t, root, base+"iterations.jsonl",
		`{"iteration":1,"task":"T1","outcome":"done","attempts":0}`+"\n"+
			`{"iteration":2,"task":"T2","outcome":"review_fail","attempts":0}`+"\n")
	put(t, root, base+"sessions/001-plan.json",
		`{"total_cost_usd":1.5,"duration_ms":1000,"permission_denials":[],"phase":"plan","iteration":0,`+
			`"modelUsage":{"claude-opus":{"inputTokens":1,"outputTokens":2,"cacheReadInputTokens":3,"cacheCreationInputTokens":4,"costUSD":1.5}}}`)
	put(t, root, base+"sessions/002-work.json",
		`{"total_cost_usd":0.5,"duration_ms":2000,"permission_denials":[{"x":1}],"phase":"work","iteration":1,"modelUsage":{}}`)
	put(t, root, base+"sessions/003-review.json",
		`{"total_cost_usd":0.25,"duration_ms":500,"permission_denials":[],"phase":"review","iteration":2}`)
	put(t, root, base+"sessions/004-work.json.raw", `not json`)
	put(t, root, base+"reports/002-verdict.json", `{"task":"T2","verdict":"FAIL","findings":["a bug","a gap"]}`)

	// a folder that names another brief, and one with no log at all
	put(t, root, ".loop/state/runs/B2-y/20260102-100000/loop.log", "planning from docs/briefs/B2-y.loop-brief.md\n")
	put(t, root, ".loop/state/runs/B2-y/20260102-100000/sessions/001-plan.json", `{"total_cost_usd":9,"phase":"plan"}`)
	put(t, root, ".loop/state/runs/B1-x/20260103-100000/iterations.jsonl", `{"iteration":1,"task":"T1","outcome":"done"}`+"\n")

	// a resuming folder under another branch's name, and a plan session
	// that lives in yet another branch's folder
	put(t, root, ".loop/state/runs/other-branch/20260104-100000/loop.log", "\x1b[36m[loop]\x1b[0m resuming B1-x — 1/2 done\n")
	put(t, root, ".loop/state/runs/other-branch/20260104-100000/iterations.jsonl", `{"iteration":1,"task":"T2","outcome":"done"}`+"\n")
	put(t, root, ".loop/state/runs/other-branch/20260104-100000/sessions/001-work.json", `{"total_cost_usd":2,"phase":"work","iteration":1}`)
	put(t, root, ".loop/state/runs/other-branch/20260105-100000/loop.log", "resuming B1-xyz\n")
	return root
}

func TestShellLoopFolders(t *testing.T) {
	root := shellFixture(t)
	m, err := Load(root, briefPath)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, f := range m.Folders {
		paths = append(paths, f.Path)
	}
	want := ".loop/state/runs/B1-x/20260101-100000 .loop/state/runs/other-branch/20260104-100000"
	if strings.Join(paths, " ") != want {
		t.Fatalf("folders = %v, want %s", paths, want)
	}
	// the brief may also be given by name
	if m2, _ := Load(root, "B1-x"); len(m2.Folders) != 2 {
		t.Errorf("by name: %d folders", len(m2.Folders))
	}
}

func TestShellLoopSessionsIterationsVerdicts(t *testing.T) {
	root := shellFixture(t)
	m, err := Load(root, briefPath)
	if err != nil {
		t.Fatal(err)
	}
	f := m.Folders[0]
	if len(f.Sessions) != 3 {
		t.Fatalf("sessions = %d (the .raw file must be skipped)", len(f.Sessions))
	}
	plan, work, review := f.Sessions[0], f.Sessions[1], f.Sessions[2]
	if plan.Phase != "plan" || plan.Task != "" || plan.CostUSD != 1.5 || plan.DurationMS != 1000 {
		t.Errorf("plan = %+v", plan)
	}
	if tk := plan.Models["claude-opus"]; tk != (Tokens{1, 2, 3, 4, 1.5}) {
		t.Errorf("tokens = %+v", tk)
	}
	if work.Task != "T1" || work.Denials != 1 || work.Seq != 2 {
		t.Errorf("work = %+v", work)
	}
	if review.Task != "T2" || review.Phase != "review" || review.Iteration != 2 {
		t.Errorf("review = %+v", review)
	}
	if len(f.Iterations) != 2 || f.Iterations[1].Canonical() != "rejected" || f.Iterations[1].GateMS != nil {
		t.Errorf("iterations = %+v", f.Iterations)
	}
	if len(f.Verdicts) != 1 {
		t.Fatalf("verdicts = %+v", f.Verdicts)
	}
	v := f.Verdicts[0]
	if v.Iteration != 2 || v.Task != "T2" || v.Verdict != "FAIL" || len(v.Findings) != 2 || v.Findings[0] != (Finding{Summary: "a bug"}) {
		t.Errorf("verdict = %+v", v)
	}
}

func TestVloopLayoutFolders(t *testing.T) {
	root := t.TempDir()
	base := ".vloop/state/runs/B1-x/20260101-100000/"
	put(t, root, base+"iterations.jsonl",
		`{"schema":"iteration/v1","run_id":"B1-x","iteration":1,"task":"T1","attempt":1,"outcome":"gate_failed","gate":{"exit":1,"duration_ms":1500},"started":"2026-01-01T10:00:00Z","ended":"2026-01-01T10:05:00Z"}`+"\n"+
			`{"schema":"iteration/v1","run_id":"B1-x","iteration":2,"task":"T1","attempt":2,"outcome":"session_error","gate":null,"started":"2026-01-01T10:06:00Z","ended":"2026-01-01T10:07:00Z"}`+"\n")
	put(t, root, base+"sessions/001-work.json",
		`{"schema":"session/v1","run_id":"B1-x","iteration":1,"phase":"work","task":"T1","model":"sonnet","effort":"high",`+
			`"models_used":{"claude-sonnet":{"input_tokens":5,"output_tokens":6,"cache_read_input_tokens":7,"cache_creation_input_tokens":8,"cost_usd":0.75}},`+
			`"started":"2026-01-01T10:00:00Z","duration_ms":3000,"cost_usd":0.75,"turns":3,"is_error":false,"permission_denials":[]}`)
	put(t, root, base+"sessions/002-review.json",
		`{"schema":"session/v1","run_id":"B1-x","iteration":2,"phase":"review","task":"T1","model":"sonnet","effort":null,"models_used":{},"started":"2026-01-01T10:06:00Z","duration_ms":1,"cost_usd":0.1,"turns":1,"is_error":true,"permission_denials":[]}`)
	put(t, root, base+"reports/002-verdict.json",
		`{"schema":"verdict/v1","task":"T1","verdict":"FAIL","criteria":[],"findings":[{"summary":"s","kind":"spec-gap"}],"notes":""}`)
	// another run id's folder is not this brief's
	put(t, root, ".vloop/state/runs/B9-z/20260101-100000/iterations.jsonl", `{"iteration":1,"task":"T1","outcome":"done"}`+"\n")

	m, err := Load(root, "B1-x")
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Folders) != 1 || m.Folders[0].Layout != Vloop || m.Folders[0].Path != ".vloop/state/runs/B1-x/20260101-100000" {
		t.Fatalf("folders = %+v", m.Folders)
	}
	f := m.Folders[0]
	if g := f.Iterations[0].GateMS; g == nil || *g != 1500 || f.Iterations[1].GateMS != nil {
		t.Errorf("gate durations = %+v", f.Iterations)
	}
	if f.Iterations[0].Attempt != 1 || f.Iterations[0].Canonical() != "gate_failed" || !f.Iterations[0].Ended.After(f.Iterations[0].Started) {
		t.Errorf("iteration = %+v", f.Iterations[0])
	}
	w, r := f.Sessions[0], f.Sessions[1]
	if w.CostUSD != 0.75 || w.Effort != "high" || w.Model != "sonnet" || w.Task != "T1" || w.Models["claude-sonnet"] != (Tokens{5, 6, 7, 8, 0.75}) {
		t.Errorf("work = %+v", w)
	}
	if !r.IsError || r.Effort != "" {
		t.Errorf("review = %+v", r)
	}
	if v := f.Verdicts[0]; len(v.Findings) != 1 || v.Findings[0] != (Finding{"s", "spec-gap"}) {
		t.Errorf("verdict = %+v", v)
	}
}

// ---- owned commits ----

type repo struct {
	t    *testing.T
	root string
	tick int
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.root}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func newRepo(t *testing.T) *repo {
	r := &repo{t: t, root: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.name", "gate")
	r.git("config", "user.email", "gate@example.com")
	r.git("config", "commit.gpgsign", "false")
	return r
}

func (r *repo) commit(subject string) string {
	r.tick++
	d := time.Date(2026, 1, 1, 10, r.tick, 0, 0, time.UTC).Format(time.RFC3339)
	cmd := exec.Command("git", "-C", r.root, "commit", "-q", "--allow-empty", "-m", subject)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_DATE="+d, "GIT_COMMITTER_DATE="+d)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("commit %q: %v\n%s", subject, err, out)
	}
	return r.git("rev-parse", "HEAD")
}

func (r *repo) plan(l Layout, subject, body string) string {
	put(r.t, r.root, l.StatePath(), body)
	r.git("add", "-A")
	return r.commit(subject)
}

func subjects(cs []Commit) string {
	var s []string
	for _, c := range cs {
		s = append(s, c.Subject)
	}
	return strings.Join(s, "|")
}

func TestOwnedCommitsLatestPlanOnly(t *testing.T) {
	r := newRepo(t)
	r.commit("[loop] plan B1-x")
	r.commit("[loop] T1: done")
	r.commit("[loop] run B1-x/20260101-100000: stalled")
	replan := r.plan(ShellLoop, "[loop] plan B1-x", `{"run_id":"B1-x","tasks":[{"id":"T1","status":"pending"}]}`)
	r.commit("[loop] T1: gate_fail")
	r.commit("[loop] T1: done")
	r.commit("[loop] T2: done")
	run := r.commit("[loop] run B1-x/20260102-100000: complete")
	// commits after the last run commit, and another brief's, are not owned
	r.commit("[loop] T3: done")
	r.commit("[loop] plan B2-y")

	o, err := OwnedCommits(r.root, "B1-x")
	if err != nil || o == nil {
		t.Fatalf("owned = %v, %v", o, err)
	}
	if o.Plan.SHA != replan || o.Layout != ShellLoop {
		t.Errorf("plan = %s (%s), want the re-plan %s", o.Plan.SHA, o.Layout.Name, replan)
	}
	if o.Run == nil || o.Run.SHA != run || o.Run.Outcome != "complete" {
		t.Errorf("run = %+v", o.Run)
	}
	if got := subjects(o.Tasks); got != "[loop] T1: gate_fail|[loop] T1: done|[loop] T2: done" {
		t.Errorf("tasks = %s", got)
	}
	if o.Tasks[0].Task != "T1" || o.Tasks[0].Outcome != "gate_fail" {
		t.Errorf("task commit = %+v", o.Tasks[0])
	}
	if none, err := OwnedCommits(r.root, "B9-none"); err != nil || none != nil {
		t.Errorf("unknown run id: %v, %v", none, err)
	}
}

func TestOwnedCommitsAnyRefAndNoRunCommit(t *testing.T) {
	r := newRepo(t)
	r.commit("base")
	r.git("switch", "-q", "-c", "work")
	r.commit("[vloop] plan B1-x")
	r.commit("[vloop] T1: done")
	r.commit("[loop] T9: done") // the other driver's commit does not belong
	r.git("switch", "-q", "main")

	o, err := OwnedCommits(r.root, "B1-x")
	if err != nil || o == nil {
		t.Fatalf("owned = %v, %v", o, err)
	}
	if o.Layout != Vloop || o.Run != nil || subjects(o.Tasks) != "[vloop] T1: done" {
		t.Errorf("owned = %+v", o)
	}
}

func TestOwnedCommitsPlanReadFromGit(t *testing.T) {
	r := newRepo(t)
	sha := r.plan(ShellLoop, "[loop] plan B1-x", `{"run_id":"B1-x","brief":"`+briefPath+`","tasks":[{"id":"T1","title":"t","status":"pending","attempts":0},{"id":"T2","area":"go","status":"pending"}]}`)
	// the working tree moves on; the commit must not
	put(t, r.root, ShellLoop.StatePath(), `{"run_id":"B1-x","tasks":[]}`)

	p, err := PlanAt(r.root, ShellLoop, sha)
	if err != nil {
		t.Fatal(err)
	}
	if p.RunID != "B1-x" || len(p.Tasks) != 2 || p.Tasks[1].Area != "go" {
		t.Errorf("plan = %+v", p)
	}
	if _, err := PlanAt(r.root, Vloop, sha); err == nil {
		t.Error("no .vloop plan at that commit: want an error")
	}
	m, err := Read(r.root, briefPath)
	if err != nil || m.Owned == nil || m.Owned.Plan.SHA != sha {
		t.Errorf("Read = %+v, %v", m, err)
	}
	if _, err := os.Stat(filepath.Join(r.root, ".vloop")); !os.IsNotExist(err) {
		t.Error("reading must not create .vloop")
	}
}

func TestVloopV2ChecksAndGateReviewSession(t *testing.T) {
	root := t.TempDir()
	base := ".vloop/state/runs/B2-x/20260101-100000/"
	put(t, root, base+"iterations.jsonl",
		`{"schema":"iteration/v2","run_id":"B2-x","iteration":1,"task":"T1","attempt":1,"outcome":"check_failed","gate":null,"checks":[{"name":"a","exit":0,"duration_ms":400},{"name":"b","exit":1,"duration_ms":600}],"started":"2026-01-01T10:00:00Z","ended":"2026-01-01T10:05:00Z"}`+"\n"+
			`{"schema":"iteration/v2","run_id":"B2-x","iteration":2,"task":"T1","attempt":2,"outcome":"done","gate":null,"checks":[],"started":"2026-01-01T10:06:00Z","ended":"2026-01-01T10:07:00Z"}`+"\n")
	put(t, root, base+"sessions/001-gate-review.json",
		`{"schema":"session/v2","run_id":"B2-x","iteration":0,"phase":"gate-review","model":"sonnet","effort":null,"models_used":{},"started":"2026-01-01T10:00:00Z","duration_ms":2000,"cost_usd":0.3,"turns":1,"is_error":false,"permission_denials":[]}`)
	m, err := Load(root, "B2-x")
	if err != nil {
		t.Fatal(err)
	}
	f := m.Folders[0]
	if c := f.Iterations[0].ChecksMS; c == nil || *c != 1000 || f.Iterations[1].ChecksMS != nil {
		t.Errorf("checks durations = %+v", f.Iterations)
	}
	if len(f.Sessions) != 1 || f.Sessions[0].Phase != "gate-review" || f.Sessions[0].DurationMS != 2000 {
		t.Errorf("sessions = %+v", f.Sessions)
	}
}

func TestFoldersAreOrderedByTheirRecordsNotTheirNames(t *testing.T) {
	root := t.TempDir()
	// A DST fall-back: the later run's local name sorts first.
	base := ".vloop/state/runs/B1-x/"
	put(t, root, base+"20261025-011000/iterations.jsonl", `{"iteration":1,"task":"T1","started":"2026-10-25T01:10:00Z"}`+"\n")
	put(t, root, base+"20261025-015000/iterations.jsonl", `{"iteration":1,"task":"T1","started":"2026-10-25T00:50:00Z"}`+"\n")
	put(t, root, base+"20261026-000000/sessions/001-plan.json", `{"phase":"plan"}`)
	m, err := Load(root, briefPath)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, f := range m.Folders {
		got = append(got, filepath.Base(f.Path))
	}
	want := "20261025-015000 20261025-011000 20261026-000000"
	if strings.Join(got, " ") != want {
		t.Errorf("folders %v, want %s", got, want)
	}
}
