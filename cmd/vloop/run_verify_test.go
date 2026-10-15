package main

import (
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
