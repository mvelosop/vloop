package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoSuchBriefAndNoSuchDir(t *testing.T) {
	t.Parallel()
	root := statusRepo(t, "")
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	code, out, errOut := run(t, "-C", root, "metrics", "nope")
	if code != 1 || out != "" || errOut != "vloop: no brief nope — vloop brief list shows the briefs\n" {
		t.Errorf("no brief: %d %q %q", code, out, errOut)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs", "briefs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "briefs", "b.loop-brief.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut = run(t, "-C", root, "metrics", "b")
	if code != 1 || errOut != "vloop: no runs for b\n" {
		t.Errorf("no runs: %d %q", code, errOut)
	}
	missing := filepath.Join(root, "no-such-dir")
	for _, args := range [][]string{{"status"}, {"version"}, {"task", "list"}, {"init"}} {
		code, out, errOut = run(t, append([]string{"-C", missing}, args...)...)
		if code != 1 || out != "" || errOut != "vloop: -C "+missing+": no such directory\n" {
			t.Errorf("%v: %d %q %q", args, code, out, errOut)
		}
	}
	if _, err := os.Stat(missing); err == nil {
		t.Error("-C created its directory")
	}
}

func TestNotAGitRepository(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	code, out, errOut := run(t, "-C", dir, "metrics")
	if code != 1 || out != "" || !strings.HasSuffix(errOut, " is not a git repository\n") || strings.Contains(errOut, "exit status") {
		t.Errorf("%d %q %q", code, out, errOut)
	}
}
