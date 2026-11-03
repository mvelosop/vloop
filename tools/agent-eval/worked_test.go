package main

import (
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// workedEnv installs the fake copilot, claude and vloop from testdata/worked
// first on PATH and returns the fixture cases, the state dir and the vloop path.
func workedEnv(t *testing.T) (cases, st, vloop string) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash is needed for the fixture scaffolds")
	}
	src, _ := filepath.Abs("testdata/worked")
	root := t.TempDir()
	bin, st := filepath.Join(root, "bin"), filepath.Join(root, "st")
	for _, d := range []string{bin, st} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []string{"copilot", "claude", "vloop"} {
		b, err := os.ReadFile(filepath.Join(src, "fake-"+n+".sh"))
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(bin, n), strings.ReplaceAll(string(b), "@ST@", st))
		if err := os.Chmod(filepath.Join(bin, n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []string{"copilot-session.jsonl", "claude-stream.jsonl"} {
		b, err := os.ReadFile(filepath.Join(src, n))
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(st, n), string(b))
	}
	cases = filepath.Join(root, "cases")
	if out, err := exec.Command("cp", "-R", filepath.Join(src, "cases"), cases).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v %s", err, out)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return cases, st, filepath.Join(bin, "vloop")
}

func gradeMap(t *testing.T, r report) string {
	t.Helper()
	var parts []string
	for _, g := range r.Cases[0].Runs[0].Graders {
		parts = append(parts, g.Name+"="+strconv.FormatFloat(g.Score, 'g', -1, 64))
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

func TestWorkedExample(t *testing.T) {
	cases, st, vloop := workedEnv(t)
	out := t.TempDir()
	o := func(n string) string { return filepath.Join(out, n) }
	base := []string{"--agent", "copilot", "--model", "m", "--vloop", vloop, "--cases", cases}

	// o1: all four graders pass.
	code, _, errs := invoke(append(base, "--case", "fixture-review", "--out", o("o1"))...)
	if code != 0 {
		t.Fatalf("o1: exit %d: %s", code, errs)
	}
	r1 := readReport(t, o("o1"))
	if c := r1.Cases[0]; c.Status != "scored" || c.Mean == nil || *c.Mean != 1 {
		t.Fatalf("o1: %+v", c)
	}
	if got := gradeMap(t, r1); got != "a=1 b=1 c=1 d=1" {
		t.Errorf("o1 graders: %s", got)
	}
	trace1, err := os.ReadFile(filepath.Join(o("o1"), "fixture-review", "run-1", "trace.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(trace1), `"tool":"Bash"`) || !strings.Contains(string(trace1), "vloop schema validate") {
		t.Errorf("o1 trace has no Bash tool event running vloop schema validate")
	}

	// o2: the verdict is PASS, so grader b scores 0.
	write(t, filepath.Join(st, "mode"), "pass")
	code, stdout, errs := invoke(append(base, "--case", "fixture-review", "--out", o("o2"))...)
	os.Remove(filepath.Join(st, "mode"))
	if code != 0 {
		t.Fatalf("o2: exit %d: %s", code, errs)
	}
	r2 := readReport(t, o("o2"))
	if got := gradeMap(t, r2); got != "a=1 b=0 c=1 d=1" {
		t.Errorf("o2 graders: %s", got)
	}
	if m := r2.Cases[0].Mean; m == nil || math.Abs(*m-4.0/6) > 1e-9 {
		t.Errorf("o2 mean: %v", m)
	}
	var row []string
	for _, l := range strings.Split(stdout, "\n") {
		if f := strings.Fields(l); len(f) > 0 && f[0] == "fixture-review" {
			row = f
		}
	}
	if len(row) < 4 || row[3] != "0.67" || row[len(row)-1] != "b" {
		t.Errorf("o2 table row: %v\n%s", row, stdout)
	}

	// o3: the scaffold exits 3.
	td3 := filepath.Join(t.TempDir(), "td3")
	if b, err := exec.Command("cp", "-R", cases, td3).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v %s", err, b)
	}
	write(t, filepath.Join(td3, "fixture-review", "scaffold.sh"), "#!/usr/bin/env bash\necho \"cannot build\" >&2\nexit 3\n")
	args3 := append([]string(nil), base...)
	args3[len(args3)-1] = td3
	code, _, _ = invoke(append(args3, "--case", "fixture-review", "--out", o("o3"))...)
	if code != 1 {
		t.Errorf("o3: exit %d, want 1", code)
	}
	if run := readReport(t, o("o3")).Cases[0].Runs[0]; run.Status != "error" || !strings.HasPrefix(run.Reason, "scaffold exited 3:") {
		t.Errorf("o3: %+v", run)
	}

	// o4: a case without a phase is skipped.
	code, _, errs = invoke(append(base, "--case", "fixture-operate", "--out", o("o4"))...)
	if code != 0 {
		t.Fatalf("o4: exit %d: %s", code, errs)
	}
	if c := readReport(t, o("o4")).Cases[0]; c.Status != "skipped" || c.Reason != "no phase" {
		t.Errorf("o4: %+v", c)
	}

	// --list over the fixture cases.
	code, stdout, errs = invoke("--list", "--cases", cases)
	var rows []string
	for i, l := range strings.Split(stdout, "\n") {
		if f := strings.Fields(l); i > 0 && len(f) >= 3 {
			rows = append(rows, strings.Join(f[:3], " "))
		}
	}
	if code != 0 || errs != "" || strings.Join(rows, ";") != "fixture-operate - 0;fixture-review review 4" {
		t.Errorf("--list: code %d, %q, %q", code, stdout, errs)
	}

	// --list over plugin/evals: a header and one line per case.
	code, stdout, errs = invoke("--list", "--cases", filepath.Join("..", "..", "plugin", "evals"))
	entries, _ := os.ReadDir(filepath.Join("..", "..", "plugin", "evals"))
	n := 0
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), "_") {
			n++
		}
	}
	lines := 0
	for _, l := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(l) != "" {
			lines++
		}
	}
	if code != 0 || errs != "" || lines != n+1 {
		t.Errorf("--list plugin/evals: code %d, %d lines for %d cases, stderr %q", code, lines, n, errs)
	}

	// An unknown agent and an --out that is not empty.
	code, _, errs = invoke("--agent", "nope", "--model", "m", "--out", o("o5"))
	if code != 2 || errs != "agent-eval: unknown agent nope\n" {
		t.Errorf("--agent nope: %d %q", code, errs)
	}
	code, _, errs = invoke("--agent", "claude", "--model", "m", "--cases", cases, "--out", o("o1"))
	if code != 2 || errs != "agent-eval: "+o("o1")+" is not empty\n" {
		t.Errorf("--out not empty: %d %q", code, errs)
	}

	// o6: the same case on Claude gives o1's trace and mean 1.
	code, _, errs = invoke("--agent", "claude", "--model", "m", "--cases", cases, "--case", "fixture-review", "--out", o("o6"))
	if code != 0 {
		t.Fatalf("o6: exit %d: %s", code, errs)
	}
	trace6, err := os.ReadFile(filepath.Join(o("o6"), "fixture-review", "run-1", "trace.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if string(trace6) != string(trace1) {
		t.Errorf("o6: trace differs from o1's\n%s\n---\n%s", trace6, trace1)
	}
	if m := readReport(t, o("o6")).Cases[0].Mean; m == nil || *m != 1 {
		t.Errorf("o6 mean: %v", m)
	}
}
