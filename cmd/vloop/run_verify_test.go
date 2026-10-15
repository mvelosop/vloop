package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const verifyTwoFindings = `case "$PHASE" in
  verify) mkdir -p .vloop/tmp
    printf '{"schema":"verify/v1","findings":[{"kind":"example","claim":"a","observed":"b","where":"x"},{"kind":"surface","claim":"c","observed":"d","where":"y"}],"checked":{"example_lines":1,"sentences":1,"surfaces":1},"notes":"none"}\n' > .vloop/tmp/verify.json;;
esac
`

// TestRunVerifyKeepsReport: a complete run ends with one verify session whose
// report is kept and announced and whose record is session/v2; the findings
// change nothing about the ending (V1).
func TestRunVerifyKeepsReport(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+verifyTwoFindings)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	r.wantStatus("complete")
	wantIn(t, "stdout", res.out, "verify: 2 finding(s) — see ")
	if _, err := os.Stat(filepath.Join(r.runFolder(), "reports", "verify.json")); err != nil {
		t.Error("no reports/verify.json")
	}
	if got := strings.Join(r.sessionFiles(), " "); !strings.HasSuffix(got, "-verify.json") {
		t.Errorf("session files %q, want a verify record last", got)
	}
	r.wantClean()
}

// TestRunVerifyOff: run.verify = false runs no verify session and prints nothing of it.
func TestRunVerifyOff(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript)
	wantExit(t, r.vloop("config", "set", "run.verify", "false"), 0)
	r.git("add", "-A")
	r.git("commit", "-q", "-m", "verify off")
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantNotIn(t, "stdout", res.out, "verify: ")
	if strings.Contains(strings.Join(r.argv(), "\n"), "/vloop:verify") {
		t.Error("a verify session ran with run.verify = false")
	}
}

const verifyEmptyReport = `case "$PHASE" in
  verify) mkdir -p .vloop/tmp
    printf '{"schema":"verify/v1","findings":[],"checked":{"example_lines":1,"sentences":1,"surfaces":1},"notes":"none"}\n' > .vloop/tmp/verify.json;;
esac
`

func (r *runRepo) verifySessions() int {
	r.t.Helper()
	n := 0
	for _, a := range r.argv() {
		if strings.Contains(a, "-p /vloop:verify") {
			n++
		}
	}
	return n
}

// V1: an empty findings list is announced as "verify: no findings" and kept.
func TestRunVerifyNoFindings(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+verifyEmptyReport)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "stdout", res.out, "verify: no findings")
	if _, err := os.Stat(filepath.Join(r.runFolder(), "reports", "verify.json")); err != nil {
		t.Error("no reports/verify.json")
	}
}

// V1: a missing report and an invalid one both warn and end the run complete.
func TestRunVerifyMissingAndInvalidReport(t *testing.T) {
	t.Parallel()
	const warn = "vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand"
	for name, body := range map[string]string{
		"missing": `case "$PHASE" in verify) rm -f .vloop/tmp/verify.json;; esac
`,
		"invalid": `case "$PHASE" in verify) mkdir -p .vloop/tmp; echo '{"schema":"verify/v1","findings":[{"kind":"bug"}]}' > .vloop/tmp/verify.json;; esac
`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := newRunRepo(t)
			r.scripted(oneTask(t), defaultScript+body)
			res := r.vloop("run", runBrief)
			wantExit(t, res, 0)
			r.wantStatus("complete")
			wantIn(t, "output", res.out+res.err, warn)
			wantNotIn(t, "output", res.out, "verify: ")
			if _, err := os.Stat(filepath.Join(r.runFolder(), "reports", "verify.json")); err == nil {
				t.Error("a report was kept")
			}
		})
	}
}

// V1: a verify session that exits non-zero warns and the run still completes.
func TestRunVerifySessionFails(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+`[ "$PHASE" = verify ] && STUB_EXIT=3
`)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	r.wantStatus("complete")
	wantIn(t, "output", res.out+res.err, "vloop: the verify session failed to run — verify by hand; see ")
}

// V1: with the ceiling already reached no verify session starts.
func TestRunVerifyCostCeiling(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), "STUB_COST=0.25\n"+defaultScript)
	res := r.runWith(nil, "--cost-ceiling", "1", runBrief)
	wantExit(t, res, 0)
	r.wantStatus("complete")
	wantIn(t, "output", res.out+res.err, "vloop: cost ceiling reached — verify skipped")
	if n := r.verifySessions(); n != 0 {
		t.Errorf("%d verify sessions past the ceiling", n)
	}
}

// V1: what a verify session changes outside .vloop/tmp/ is reverted and named.
func TestRunVerifyRevertsChanges(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+verifyEmptyReport+`[ "$PHASE" = verify ] && { echo changed >> docs/x.md; : > stray.txt; }
`)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	for _, p := range []string{"docs/x.md", "stray.txt"} {
		wantIn(t, "output", res.out+res.err, "vloop: the verify session changed "+p+" — reverted")
	}
	if r.has("stray.txt") {
		t.Error("stray.txt was kept")
	}
	if got := r.read("docs/x.md"); got != "x\n" {
		t.Errorf("docs/x.md = %q, want it reverted", got)
	}
	if _, err := os.Stat(filepath.Join(r.runFolder(), "reports", "verify.json")); err != nil {
		t.Error("the report was not kept")
	}
	r.wantClean()
}

// V1: a verify session that moves a ref halts the run with exit 9.
func TestRunVerifyMovesRef(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+verifyEmptyReport+`[ "$PHASE" = verify ] && git tag planted-by-verify
`)
	wantExit(t, r.vloop("run", runBrief), 9)
}

// V1: vloop status names the phase while the verify session runs, and the
// record is session/v2 with no task.
func TestRunVerifyStatusAndRecord(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+verifyEmptyReport+`[ "$PHASE" = verify ] && '`+binPath+`' status > "$dir/status.txt" 2>&1
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	b, err := os.ReadFile(filepath.Join(r.stub, "status.txt"))
	if err != nil {
		t.Fatal(err)
	}
	wantIn(t, "status", string(b), "running: verify for ")
	var rec struct {
		Schema    string `json:"schema"`
		Phase     string `json:"phase"`
		Iteration int    `json:"iteration"`
	}
	files := r.sessionFiles()
	raw, err := os.ReadFile(filepath.Join(r.runFolder(), "sessions", files[len(files)-1]))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatal(err)
	}
	if rec.Schema != "session/v2" || rec.Phase != "verify" || rec.Iteration != 0 || strings.Contains(string(raw), `"task"`) {
		t.Errorf("record = %s", raw)
	}
}

// V1: a resume with every task done repeats the final pass, then verifies.
func TestRunVerifyOnResume(t *testing.T) {
	t.Parallel()
	r := checksRepo(t, twoChecks)
	r.write("web/test.sh", "test ! -f api/T1.out\n")
	r.commitAll("web breaks when api/T1.out exists")
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "test -f api/T1.out"})),
		strings.Replace(defaultScript, `touch "$TASK.out"`, `touch api/"$TASK.out"`, 1)+verifyEmptyReport)
	wantExit(t, r.vloop("run", runBrief), 2)
	if n := r.verifySessions(); n != 0 {
		t.Fatalf("%d verify sessions after a failed final pass", n)
	}
	r.write("web/test.sh", "exit 0\n")
	r.commitAll("fix web")
	res := r.vloop("run")
	wantExit(t, res, 0)
	wantIn(t, "stdout", res.out, "verify: no findings")
	if n := r.verifySessions(); n != 1 {
		t.Errorf("%d verify sessions after the resume, want 1", n)
	}
}
