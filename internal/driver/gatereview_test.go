package driver

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// gateRound runs one gate review round against a stub session that writes the
// verdict, and returns the problems, the check-gap count and the run folder.
func gateRound(t *testing.T, verdict string) ([]string, int, string) {
	t.Helper()
	r, _, _ := fixture(t, claudeOut)
	root := r.Root
	if err := os.RemoveAll(filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if err := os.MkdirAll(filepath.Join(root, tmpDir), 0o755); err != nil {
		t.Fatal(err)
	}
	vf := filepath.Join(root, "verdict.json")
	if err := os.WriteFile(vf, []byte(verdict), 0o644); err != nil {
		t.Fatal(err)
	}
	stub := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\ncp '" + vf + "' '" + filepath.Join(root, gateVerdictFile) + "'\ncat <<'EOF'\n" + claudeOut + "EOF\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	r.Claude = stub
	p := &Planner{Root: root}
	problems, gaps, err := p.gateReviewRound(term{p, r}, 2)
	if err != nil {
		t.Fatal(err)
	}
	return problems, gaps, r.RunDir
}

const gapVerdict = `{"schema":"gate-verdict/v2","verdict":"FAIL","tasks":[{"task":"T1","verdict":"FAIL","findings":[` +
	`{"kind":"check-gap","summary":"pnpm smoke: no check runs it"},{"kind":"wrong-reason","summary":"fails on a typo"}]}],"notes":"n"}`

func TestGateReviewRoundCountsCheckGaps(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub claude is a POSIX script")
	}
	problems, gaps, dir := gateRound(t, gapVerdict)
	if gaps != 1 || len(problems) != 2 {
		t.Fatalf("gaps %d problems %q, want 1 gap among 2 problems", gaps, problems)
	}
	if problems[0] != "gate T1 (check-gap): pnpm smoke: no check runs it" {
		t.Errorf("problem %q", problems[0])
	}
	if _, err := os.Stat(filepath.Join(dir, "reports", "gate-review-2.json")); err != nil {
		t.Errorf("the report was not kept: %v", err)
	}
}

func TestGateReviewRoundUncoveredIsNoCheckGap(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub claude is a POSIX script")
	}
	v := strings.Replace(gapVerdict, `{"kind":"check-gap","summary":"pnpm smoke: no check runs it"},`, `{"kind":"uncovered","summary":"T1.out"},`, 1)
	problems, gaps, _ := gateRound(t, v)
	if gaps != 0 || len(problems) != 2 {
		t.Fatalf("gaps %d problems %q, want 0 gaps among 2 problems", gaps, problems)
	}
}

func TestGateReviewRoundRejectsV1(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub claude is a POSIX script")
	}
	problems, gaps, _ := gateRound(t, strings.Replace(gapVerdict, "gate-verdict/v2", "gate-verdict/v1", 1))
	if gaps != 0 || len(problems) != 1 || !strings.Contains(problems[0], "gate-verdict/v2") {
		t.Fatalf("gaps %d problems %q", gaps, problems)
	}
}

func TestCheckGapLine(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".vloop", "state", "runs", "B1", "20260101-000000")
	want := "the gate review found 2 check gap(s): the plan needs a check .vloop/config.toml does not have — add it, commit it, then re-run (see .vloop/state/runs/B1/20260101-000000/reports/gate-review-2.json)"
	if got := checkGapLine(root, dir, 2, 2); got != want {
		t.Errorf("line %q, want %q", got, want)
	}
}
