package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitIn(t *testing.T, dir, date string, args ...string) {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", dir}, args...)...)
	c.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date,
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func write(t *testing.T, dir, rel, s string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

// matrixRepo builds a vloop-layout run: T1 fails its gate at 09:02, T1's gate
// is replaced at 09:03 by hist's `by`, T1 is then rejected once with a
// spec-gap finding, then done.
func matrixRepo(t *testing.T, by string) string {
	dir := t.TempDir()
	const brief = "B20260102-0900-b"
	gitIn(t, dir, "2026-01-02T08:00:00Z", "init", "-q", "-b", "main")
	write(t, dir, "docs/briefs/"+brief+".loop-brief.md", "---\nname: "+brief+".loop-brief\nstatus: ready\n---\n")
	gitIn(t, dir, "2026-01-02T08:00:00Z", "add", "-A")
	gitIn(t, dir, "2026-01-02T08:00:00Z", "commit", "-q", "-m", "init")
	plan := func(hist string) {
		write(t, dir, ".vloop/state/state.json", `{"run_id":"`+brief+`","tasks":[{"id":"T1","status":"done"`+hist+`}]}`)
	}
	plan("")
	gitIn(t, dir, "2026-01-02T09:00:00Z", "add", "-A")
	gitIn(t, dir, "2026-01-02T09:00:00Z", "commit", "-q", "-m", "[vloop] plan "+brief)
	gitIn(t, dir, "2026-01-02T09:02:00Z", "commit", "-q", "--allow-empty", "-m", "[vloop] T1: gate_failed")
	gitIn(t, dir, "2026-01-02T09:04:00Z", "commit", "-q", "--allow-empty", "-m", "[vloop] T1: rejected")
	plan(`,"gate_history":[{"verify":"x","replaced_at":"2026-01-02T09:03:00Z","reason":"r","by":"` + by + `"}]`)
	folder := ".vloop/state/runs/" + brief + "/20260102-090000/"
	write(t, dir, folder+"iterations.jsonl",
		`{"iteration":1,"task":"T1","outcome":"gate_failed"}`+"\n"+`{"iteration":2,"task":"T1","outcome":"rejected"}`+"\n")
	write(t, dir, folder+"reports/002-verdict.json", `{"task":"T1","verdict":"FAIL","findings":[{"summary":"s","kind":"spec-gap"}]}`)
	gitIn(t, dir, "2026-01-02T09:10:00Z", "add", "-A")
	gitIn(t, dir, "2026-01-02T09:10:00Z", "commit", "-q", "-m", "[vloop] run "+brief+"/20260102-090000: complete")
	return dir
}

func TestMatrixCommand(t *testing.T) {
	head := "       gate  review  operator  user\n"
	for _, c := range []struct{ by, want string }{
		{"operator", head + "brief     0       1         0     0\nplan      1       0         0     0\nwork      0       0         0     0\nenv       0       0         0     0\n"},
		{"planner", head + "brief     0       1         0     0\nplan      0       0         0     0\nwork      1       0         0     0\nenv       0       0         0     0\n"},
	} {
		dir := matrixRepo(t, c.by)
		for _, args := range [][]string{{"defect", "list", "--matrix"}, {"defect", "list", "--matrix", "--brief", "B20260102-0900-b.loop-brief"}} {
			code, out, errs := runDefect(t, dir, args...)
			if code != 0 || out != c.want {
				t.Fatalf("%s %v: %d %q %q, want %q", c.by, args, code, out, errs, c.want)
			}
		}
		if code, out, _ := runDefect(t, dir, "defect", "list"); code != 0 || out != "" {
			t.Fatalf("plain list shows derived defects: %q", out)
		}
	}
	// A recorded defect is counted alongside the derived ones.
	dir := matrixRepo(t, "planner")
	if code, _, e := runDefect(t, dir, "defect", "add", "late", "--found-by", "user", "--brief", "B20260102-0900-b.loop-brief"); code != 0 {
		t.Fatal(e)
	}
	_, out, _ := runDefect(t, dir, "defect", "list", "--matrix")
	if !strings.Contains(out, "work      1       0         0     1\n") {
		t.Fatalf("recorded defect missing: %q", out)
	}
	// A repo with no runs has an all-zero matrix.
	empty := t.TempDir()
	if code, out, _ := runDefect(t, empty, "defect", "list", "--matrix"); code != 0 || !strings.Contains(out, "env       0       0         0     0\n") {
		t.Fatalf("empty: %d %q", code, out)
	}
}
