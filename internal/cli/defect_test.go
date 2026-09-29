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
