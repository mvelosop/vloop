package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runDefect(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	var o, e bytes.Buffer
	code := Execute(Build{}, append([]string{"-C", dir}, args...), &o, &e)
	return code, o.String(), e.String()
}

func TestDefectAddListSet(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".vloop"), 0o755)
	os.MkdirAll(filepath.Join(dir, "docs/briefs"), 0o755)
	os.WriteFile(filepath.Join(dir, "docs/briefs/B1.loop-brief.md"), []byte("---\nname: B1.loop-brief\nstatus: ready\n---\n"), 0o644)

	code, out, errs := runDefect(t, dir, "defect", "add", "x", "--found-by", "martian", "--brief", "B1.loop-brief")
	if code != ExitUsage || out != "" || errs != "vloop: invalid value \"martian\" for found-by: want one of gate, review, operator, user\n" {
		t.Fatalf("%d %q %q", code, out, errs)
	}
	if code, _, _ := runDefect(t, dir, "defect", "add", "x", "--found-by", "user", "--brief", "nope.loop-brief"); code != ExitProblems {
		t.Fatalf("unknown brief: %d", code)
	}
	if _, err := os.Stat(filepath.Join(dir, ".vloop/defects")); err == nil {
		t.Fatal("refused adds wrote .vloop/defects")
	}
	mergeBrief(t, dir, "B1.loop-brief")
	code, out, _ = runDefect(t, dir, "defect", "add", "It broke", "--found-by", "user", "--brief", "B1.loop-brief")
	path := strings.TrimSpace(out)
	if code != 0 || !strings.HasPrefix(path, ".vloop/defects/D") {
		t.Fatalf("%d %q", code, out)
	}
	id := strings.TrimSuffix(filepath.Base(path), ".md")
	code, out, _ = runDefect(t, dir, "defect", "list")
	if code != 0 || out != id+"  B1.loop-brief  -  bug  work  user  open  It broke\n" {
		t.Fatalf("%d %q", code, out)
	}
	if code, _, errs = runDefect(t, dir, "defect", "set", id, "status", "bogus"); code != ExitUsage {
		t.Fatalf("%d %q", code, errs)
	}
	if code, _, _ = runDefect(t, dir, "defect", "set", id, "status", "fixed"); code != 0 {
		t.Fatal(code)
	}
	code, out, _ = runDefect(t, dir, "defect", "list", "--json")
	if code != 0 || !strings.Contains(out, `"status":"fixed"`) || !strings.Contains(out, `"schema":"defect/v1"`) {
		t.Fatalf("%d %q", code, out)
	}
}

func TestDefectAddRefusesUnmergedBrief(t *testing.T) {
	const name = "B20260101-0900-a.loop-brief"
	want := "vloop: " + name + " is not merged: record what you found before merge as an intervention — vloop intervention add \"<summary>\" --brief " + name + " --kind verification-finding --origin <brief|plan|work|env> --phase verify --automatable partly --by operator\n"
	dir := t.TempDir()
	gitIn(t, dir, "2026-01-01T08:00:00Z", "init", "-q", "-b", "trunk")
	write(t, dir, ".vloop/config.toml", "")
	write(t, dir, "docs/briefs/"+name+".md", "---\nname: "+name+"\nstatus: ready\n---\n")
	write(t, dir, "src/u.txt", "line\n")
	gitIn(t, dir, "2026-01-01T08:00:00Z", "add", "-A")
	gitIn(t, dir, "2026-01-01T08:00:00Z", "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "u", "-m", "Vloop-Brief: "+name)

	// No main, no master: M-5's lookup error, nothing written.
	code, out, errs := runDefect(t, dir, "defect", "add", "x", "--found-by", "user", "--brief", name)
	if code != 1 || out != "" || errs != "vloop: no default branch: origin/HEAD, main and master are all missing\n" {
		t.Fatalf("no default branch: %d %q %q", code, out, errs)
	}
	gitIn(t, dir, "2026-01-01T08:00:00Z", "branch", "-m", "main")
	for _, args := range [][]string{{"--brief", name}, {"--blame", "src/u.txt:1"}} {
		code, out, errs := runDefect(t, dir, append([]string{"defect", "add", "x", "--found-by", "user"}, args...)...)
		if code != 1 || out != "" || !strings.HasSuffix(errs, want) {
			t.Fatalf("%v: %d %q %q", args, code, out, errs)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".vloop/defects")); err == nil {
		t.Fatal("a refused add wrote .vloop/defects")
	}
	mergeBrief(t, dir, name)
	if code, _, errs := runDefect(t, dir, "defect", "add", "x", "--found-by", "user", "--brief", name); code != 0 {
		t.Fatalf("merged brief: %d %q", code, errs)
	}
}
