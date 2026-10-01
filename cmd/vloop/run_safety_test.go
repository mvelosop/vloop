package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Safety and records, as the shell loop's scenarios 10, 13, 14, 15, 18, 20, 22,
// 23 and 24 assert them of .loop/run.sh, translated into vloop's layout, plus
// the silent-session record.

// untrust makes the fake home stop trusting the repository.
func (r *runRepo) untrust() {
	r.t.Helper()
	if err := os.WriteFile(filepath.Join(r.home, ".claude.json"), []byte(`{"projects":{}}`), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// wantNothingRan asserts a refused preflight changed nothing: no session, no
// plan, no branch, no commit.
func (r *runRepo) wantNothingRan(head string) {
	r.t.Helper()
	if n := r.sessions(); n != 0 {
		r.t.Errorf("%d session(s) ran despite the refused preflight", n)
	}
	if r.has(".vloop/state/state.json") {
		r.t.Error("a plan was written despite the refused preflight")
	}
	if r.has(".vloop/state/runs") {
		if m, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", "*", "*", "sessions", "*")); len(m) > 0 {
			r.t.Errorf("session records exist: %v", m)
		}
	}
	if b := r.branch(); b != "main" {
		r.t.Errorf("on branch %s, want main", b)
	}
	if got := strings.TrimSpace(r.git("rev-parse", "HEAD")); got != head {
		r.t.Error("a commit was made despite the refused preflight")
	}
}

func (r *runRepo) head() string { return strings.TrimSpace(r.git("rev-parse", "HEAD")) }

func (r *runRepo) sessionFiles() []string {
	r.t.Helper()
	m, _ := filepath.Glob(filepath.Join(r.runFolder(), "sessions", "*.json"))
	var names []string
	for _, f := range m {
		names = append(names, filepath.Base(f))
	}
	return names
}

// TestRun10Containment: nothing is written outside the repository and nothing
// persisted names the machine, even when a session reports an absolute path.
func TestRun10Containment(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$PHASE" = work ]; then
  printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"wrote %s/secret/file.py","files":["%s/secret/file.py"],"verified":"ok","notes":"see %s/notes.txt"}\n' "$TASK" "$HOME" "$HOME" "$HOME" > .vloop/tmp/proposal.json
fi
`)
	wantExit(t, r.vloop("run", runBrief), 0)

	tracked := strings.Split(strings.TrimSpace(r.git("ls-files", "-z")), "\x00")
	for _, f := range tracked {
		if f == "" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(r.dir, f))
		if err != nil {
			continue
		}
		for _, bad := range []string{r.dir, r.home, filepath.Base(r.home)} {
			if strings.Contains(string(b), bad) {
				t.Errorf("tracked file %s holds %q", f, bad)
			}
		}
	}
	wantIn(t, "journal", r.journal(), "~/secret/file.py")
	wantNotIn(t, "journal", r.journal(), r.home)
	if out := r.git("grep", "-l", "~/notes.txt"); !strings.Contains(out, ".vloop/") {
		t.Errorf("the masked home path reached no committed record: %q", out)
	}
	ents, _ := os.ReadDir(r.home)
	for _, e := range ents {
		if e.Name() != ".claude.json" {
			t.Errorf("the run wrote %s into the home directory", e.Name())
		}
	}
	r.wantClean()
}

// TestRun13EmptyRunSignals: a run with no iterations closes nothing, records
// nothing, and still prints its end-of-run report.
func TestRun13EmptyRunSignals(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	res := r.vloop("run", "--max-iterations", "0", runBrief)
	wantExit(t, res, 4)
	r.wantIterations()
	wantIn(t, "output", res.out, "═══ halted ═══", "(0 iteration(s) this run)", "0/2 done, 0 blocked", "iteration budget spent")
	wantIn(t, "run log", r.runLog(), "0 iteration(s) this run", "0/2 done, 0 blocked")
	if n := r.sessions(); n != 1 {
		t.Errorf("%d sessions, want only the plan session", n)
	}
	if m := r.vloop("metrics", runID+".loop-brief"); m.code != 0 {
		t.Errorf("vloop metrics on an empty run: %+v", m)
	}
}

// TestRun14RenderedViews: plan.md tracks state and is what vloop status
// --markdown renders; the journal opens with the plan and closes with the end.
func TestRun14RenderedViews(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)

	p := r.read(".vloop/state/plan.md")
	wantIn(t, "plan.md", p, "Do NOT edit", "2/2 done", "- [x] **T1**", "- [x] **T2**", "<details><summary>verify command")
	j := r.journal()
	wantIn(t, "journal", j, "## Plan — "+runID, "## T1 —", "## T2 —", "## Run ended — complete")
	if !strings.Contains(j, "## Plan — "+runID) || strings.Index(j, "## Plan —") > strings.Index(j, "## T1 —") ||
		strings.Index(j, "## T2 —") > strings.Index(j, "## Run ended —") {
		t.Errorf("the journal is out of order:\n%s", j)
	}
	if n := strings.Count(j, "## T1 —") + strings.Count(j, "## T2 —"); n != 2 {
		t.Errorf("%d iteration entries, want one per iteration", n)
	}
	for i := 0; i < 2; i++ {
		md := r.vloop("status", "--markdown")
		wantExit(t, md, 0)
		if md.out != p {
			t.Errorf("render %d differs from plan.md", i+1)
		}
	}
}

// TestRun15TelemetryContract: every record validates against its schema and
// session files sort in run order.
func TestRun15TelemetryContract(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)

	want := "001-plan.json 002-work.json 003-review.json 004-work.json 005-review.json"
	if got := strings.Join(r.sessionFiles(), " "); got != want {
		t.Fatalf("session files %q, want %q", got, want)
	}
	pairs := []string{}
	for i, n := range r.sessionFiles() {
		f := filepath.Join(r.runFolder(), "sessions", n)
		if v := r.vloop("schema", "validate", "session/v1", f); v.code != 0 {
			t.Errorf("%s is not session/v1: %+v", n, v)
		}
		var rec map[string]any
		sessionRecord(t, f, &rec)
		if rec["phase"] == "plan" {
			if i != 0 || rec["iteration"] != float64(0) {
				t.Errorf("the plan session is %v at position %d, want iteration 0 first", rec["iteration"], i)
			}
			continue
		}
		pairs = append(pairs, rec["phase"].(string)+":"+strconv.Itoa(int(rec["iteration"].(float64))))
		for _, k := range []string{"cost_usd", "turns", "duration_ms", "is_error", "permission_denials"} {
			if _, ok := rec[k]; !ok {
				t.Errorf("%s lost %s", n, k)
			}
		}
	}
	if got := strings.Join(pairs, " "); got != "work:1 review:1 work:2 review:2" {
		t.Errorf("work/review pairing %q", got)
	}
	r.wantIterations("T1:done", "T2:done")
	if n := len(r.iterations()); n != 2 {
		t.Errorf("%d iteration records", n)
	}
}

// TestRun18PreflightUntrusted: an untrusted workspace is refused before any
// session, branch, plan or commit.
func TestRun18PreflightUntrusted(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	head := r.head()
	r.untrust()
	res := r.vloop("run", runBrief)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "trust", "preflight failed")
	r.wantNothingRan(head)
}

// TestRun20ParallelSafeLayout: runs live under the run id, the journal is one
// per plan, and a resumed run appends to it in a folder of its own.
func TestRun20ParallelSafeLayout(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", "--max-iterations", "1", runBrief), 4)

	dirs, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", "*"))
	if len(dirs) != 1 || filepath.Base(dirs[0]) != runID {
		t.Errorf("run folders %v, want .vloop/state/runs/%s", dirs, runID)
	}
	stamp := filepath.Base(r.runFolder())
	if len(stamp) < 15 || stamp[8] != '-' {
		t.Errorf("run folder %q is not YYYYMMDD-HHMMSS", stamp)
	}
	journals, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "journals", "*.md"))
	if len(journals) != 1 || filepath.Base(journals[0]) != runID+".md" {
		t.Errorf("journals %v, want the one for %s", journals, runID)
	}
	before := strings.Count(r.journal(), "\n## ")

	wantExit(t, r.vloop("run"), 0)
	journals, _ = filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "journals", "*.md"))
	if len(journals) != 1 {
		t.Errorf("a resumed run started a new journal: %v", journals)
	}
	if after := strings.Count(r.journal(), "\n## "); after <= before {
		t.Errorf("the resumed run did not append (%d -> %d sections)", before, after)
	}
	wantIn(t, "journal", r.journal(), "## T1 —", "## T2 —")
	if dirs, _ = filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", runID, "*")); len(dirs) != 2 {
		t.Errorf("%d run folders, want one per run: %v", len(dirs), dirs)
	}
}

// TestRun22RunLock: a live lock stops a second run, a dead one is cleared, and
// the lock is released at the end and never committed.
func TestRun22RunLock(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	head := r.head()

	r.write(".vloop/tmp/.running", `{"pid":"`+strconv.Itoa(os.Getpid())+`","branch":"other-branch","started":"2026-08-18T00:00:00Z","run":"x"}`+"\n")
	res := r.vloop("run", runBrief)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "a loop is already running in this working tree", "git worktree add", "other-branch")
	r.wantNothingRan(head)
	if !r.has(".vloop/tmp/.running") {
		t.Error("the live lock was deleted by the run it blocked")
	}

	r.write(".vloop/tmp/.running", `{"pid":"2147483646","branch":"crashed-run","started":"2026-08-18T00:00:00Z","run":"x"}`+"\n")
	res = r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "stderr", res.err, "clearing a stale lock")
	if r.has(".vloop/tmp/.running") {
		t.Error("the lock survived a completed run")
	}
	r.wantClean()
	if out := r.git("check-ignore", ".vloop/tmp/.running"); !strings.Contains(out, ".running") {
		t.Error(".vloop/tmp/.running is not ignored")
	}
}

// TestRun23GitIdentity: no git identity is refused before spending; a
// pre-commit hook is warned about and not blocked on.
func TestRun23GitIdentity(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	head := r.head()
	r.git("config", "--unset", "user.email")
	r.git("config", "--unset", "user.name")
	res := r.vloop("run", runBrief)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "user.name and user.email not set", "preflight failed")
	r.wantNothingRan(head)

	r.git("config", "user.email", "fixture@test")
	r.git("config", "user.name", "fixture")
	r.write(".git/hooks/pre-commit", "#!/bin/sh\nexit 0\n")
	if err := os.Chmod(filepath.Join(r.dir, ".git", "hooks", "pre-commit"), 0o755); err != nil {
		t.Fatal(err)
	}
	res = r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "stderr", res.err, "a pre-commit hook is active")
}

// TestRun24StateTampering: a session that edits the plan has it restored, and
// the iteration fails whatever the gate said.
func TestRun24StateTampering(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$PHASE" = work ] && [ "$TASK" = T1 ] && [ "$ATTEMPT" = 1 ]; then
  sed 's/"verify": "test -f T1.out"/"verify": "true"/' .vloop/state/state.json > st.tmp && mv st.tmp .vloop/state/state.json
fi
if [ "$PHASE" = work ] && [ "$TASK" = T2 ] && [ "$ATTEMPT" = 1 ]; then
  sed 's/"notes": ""/"notes": "watch out for the fixture"/' .vloop/state/state.json > st.tmp && mv st.tmp .vloop/state/state.json
fi
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	wantIn(t, "run log", r.runLog(), "STATE TAMPERING T1", "STATE TAMPERING T2")
	if n := strings.Count(r.runLog(), "STATE TAMPERING"); n != 2 {
		t.Errorf("%d STATE TAMPERING lines, want 2 (verify and notes)", n)
	}
	r.wantIterations("T1:gate_failed", "T1:done", "T2:gate_failed", "T2:done")
	if v := r.task("T1")["verify"]; v != "test -f T1.out" {
		t.Errorf("T1's authored verify did not survive: %v", v)
	}
	if n := r.task("T2")["notes"]; n != "" {
		t.Errorf("T2's forged note survived: %v", n)
	}
	r.wantTask("T1", "done", 1)
	r.wantTask("T2", "done", 1)
	r.wantStatus("complete")
	r.wantClean()
}

// TestRunSessionRecordMissing: a session that prints nothing leaves no record
// and a log line, a normal one leaves its record, and metrics reports the gap.
func TestRunSessionRecordMissing(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$PHASE" = review ] && [ "$TASK" = T1 ] && [ "$ATTEMPT" = 1 ]; then
  printf '{"schema":"verdict/v1","task":"T1","verdict":"FAIL","criteria":[],"findings":["no"],"notes":"n"}\n' > .vloop/tmp/verdict.json
  STUB_SILENT=1
fi
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	wantIn(t, "run log", r.runLog(), "SESSION RECORD MISSING review T1 (iteration 1)")
	wantNotIn(t, "run log", r.runLog(), "SESSION RECORD MISSING plan", "SESSION RECORD MISSING work")
	reviews := 0
	for _, n := range r.sessionFiles() {
		var rec map[string]any
		sessionRecord(t, filepath.Join(r.runFolder(), "sessions", n), &rec)
		if rec["phase"] == "review" {
			reviews++
			if rec["iteration"] == float64(1) {
				t.Errorf("%s: a record exists for the review that printed nothing", n)
			}
		}
	}
	if reviews != 2 {
		t.Errorf("%d review records, want the 2 sessions that printed", reviews)
	}
	m := r.vloop("metrics", runID+".loop-brief")
	wantExit(t, m, 0)
	wantIn(t, "metrics", m.out, "session record(s) missing")
}

func sessionRecord(t *testing.T, path string, into *map[string]any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, into); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}
