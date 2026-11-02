package driver

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const verifyV2Report = `{"schema":"verify/v2","complete":false,"findings":[{"kind":"example","claim":"a","observed":"b","where":"x"}],"not_run":[],"checked":{"example_lines":1,"sentences":1,"surfaces":1},"notes":"none"}`

// verifyIterator is an Iterator in a fresh git repository whose verify
// session is the stub script given: it writes .vloop/tmp/verify.json when
// report is not empty, prints nothing when silent, and exits with code.
func verifyIterator(t *testing.T, report string, silent bool, code string) (*Iterator, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("stub claude is a POSIX script")
	}
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	dir := filepath.Join(root, ".vloop", "state", "runs", "B1", "20260101-000000")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nmkdir -p .vloop/tmp\n"
	if report != "" {
		script += "cat > .vloop/tmp/verify.json <<'EOF'\n" + report + "\nEOF\n"
	}
	if !silent {
		script += "cat <<'EOF'\n" + claudeOut + "EOF\n"
	}
	script += "exit " + code + "\n"
	stub := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	it := &Iterator{Root: root, RunDir: dir, Out: &out, Err: &errb, Budgets: Budgets{CostCeiling: 100},
		resolved: &Resolved{model: map[string]string{}, effort: map[string]string{}, agent: map[string]string{PhaseVerify: AgentClaude}},
		Env:      []string{"PATH=" + os.Getenv("PATH")}}
	it.r = &Runner{Root: root, Version: "1.2.3", RunID: "B1", RunDir: dir, Claude: stub, Home: t.TempDir(), User: "tester",
		Timeout: time.Minute, Env: it.Env}
	return it, &out, &errb
}

// V1 (verify/v2): the report is read as verify/v2 and kept after a session
// that exited non-zero or left no record; the failed-to-run warning follows
// the line and the stopped-early warning does not.
func TestVerifySessionKeepsReportAfterFailedSession(t *testing.T) {
	for name, c := range map[string]struct {
		silent bool
		code   string
	}{"exit non-zero": {false, "3"}, "no record": {true, "0"}} {
		t.Run(name, func(t *testing.T) {
			it, out, errb := verifyIterator(t, verifyV2Report, c.silent, c.code)
			if err := it.verifySession(); err != nil {
				t.Fatal(err)
			}
			rel := relRunDir(it.Root, it.RunDir)
			if want := "verify: 1 finding(s) — see " + rel + "reports/verify.json\n"; !strings.HasSuffix(out.String(), want) || strings.Count(out.String(), "verify: ") != 1 {
				t.Errorf("stdout %q, want %q", out.String(), want)
			}
			if want := "vloop: the verify session failed to run — verify by hand; see " + rel + "run.log\n"; errb.String() != want {
				t.Errorf("stderr %q, want %q", errb.String(), want)
			}
			kept, err := os.ReadFile(filepath.Join(it.RunDir, "reports", "verify.json"))
			if err != nil || !strings.HasSuffix(string(kept), "\n") || !strings.Contains(string(kept), `"schema":"verify/v2"`) {
				t.Errorf("kept report %q, %v", kept, err)
			}
		})
	}
}

// A verify/v1 report is not accepted from a session, failed or not.
func TestVerifySessionRejectsV1(t *testing.T) {
	v1 := `{"schema":"verify/v1","findings":[],"checked":{"example_lines":0,"sentences":0,"surfaces":0},"notes":"none"}`
	it, out, errb := verifyIterator(t, v1, false, "0")
	if err := it.verifySession(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "verify: ") || !strings.Contains(errb.String(), "left no valid .vloop/tmp/verify.json") {
		t.Errorf("stdout %q stderr %q", out.String(), errb.String())
	}
	if _, err := os.Stat(filepath.Join(it.RunDir, "reports", "verify.json")); err == nil {
		t.Error("a verify/v1 report was kept")
	}
}
