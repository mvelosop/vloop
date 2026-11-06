package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoDir is the directory the kit's fixtures quote: its name holds a '.
// Go refuses ' in a module file path, so the committed tree spells it @REPO@.
const repoDir = "it's repo"

// fixture copies testdata/probes/NAME into a temporary directory, @DIR@
// replaced by its path and a path component @REPO@ by it's repo, and returns
// the path (symlinks resolved).
func fixture(t *testing.T, name string) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join("testdata", "probes", name)
	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		dst := filepath.Join(dir, strings.ReplaceAll(rel, "@REPO@", repoDir))
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(dst, 0o777)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		b = []byte(strings.ReplaceAll(string(b), "@DIR@", dir))
		return os.WriteFile(dst, b, info.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func stubCopilot(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stub copilot is a shell script")
	}
	p, err := filepath.Abs(filepath.Join("testdata", "copilot"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// normalised is a file the kit wrote: the directory as @DIR@ and, in the
// stub's log, without the variables that differ between two runs.
func normalised(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, strings.ReplaceAll(rel, "@REPO@", repoDir)))
	if err != nil {
		t.Fatalf("%s was not written: %v", rel, err)
	}
	s := strings.ReplaceAll(string(b), dir, "@DIR@")
	if rel == "stub.log" {
		var keep []string
		for _, l := range strings.SplitAfter(s, "\n") {
			skip := false
			for _, k := range []string{"PWD", "OLDPWD", "SHLVL", "_"} {
				if strings.HasPrefix(l, "env "+k+"=") {
					skip = true
				}
			}
			if !skip {
				keep = append(keep, l)
			}
		}
		s = strings.Join(keep, "")
	}
	return s
}

// wantFiles compares every file of the golden folder NAME with the one the kit wrote.
func wantFiles(t *testing.T, name, dir string) {
	t.Helper()
	root := filepath.Join("testdata", "golden", name)
	n := 0
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		want, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if rel == "stdout" || strings.HasSuffix(rel, ".md") {
			return nil
		}
		n++
		if got := normalised(t, dir, rel); got != string(want) {
			t.Errorf("%s differs from the golden %s:\n--- want\n%s\n--- got\n%s", rel, name, want, got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatalf("golden %s holds no file", name)
	}
}

func goldenText(t *testing.T, parts ...string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(append([]string{"testdata", "golden"}, parts...)...))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func okRun(t *testing.T, args ...string) string {
	t.Helper()
	out, serr, code := runKit(t, args...)
	if code != 0 || serr != "" {
		t.Fatalf("kit %v: exit %d: %s", args, code, serr)
	}
	return out
}

func TestProbe(t *testing.T) {
	cop := stubCopilot(t)
	dir := fixture(t, "probe")
	out := okRun(t, "probe", dir, cop)
	if want := goldenText(t, "probe", "stdout"); out != want {
		t.Errorf("stdout %q, want %q", out, want)
	}
	wantFiles(t, "probe", dir)
}

func TestSummary(t *testing.T) {
	for _, s := range []string{"sum1", "sum2"} {
		t.Run(s, func(t *testing.T) {
			dir := fixture(t, s)
			got := strings.ReplaceAll(okRun(t, "summary", dir), dir, "@DIR@")
			if want := goldenText(t, "summary", s+".md"); got != want {
				t.Errorf("summary differs from the golden:\n--- want\n%s\n--- got\n%s", want, got)
			}
		})
	}
}

func TestHookForms(t *testing.T) {
	cop := stubCopilot(t)
	dir := fixture(t, "hook")
	okRun(t, "hook", "forms", dir, cop)
	wantFiles(t, "hook-forms", dir)
}

func TestHookFailure(t *testing.T) {
	cop := stubCopilot(t)
	for _, m := range []string{"missing", "exit1", "notjson", "timeout"} {
		t.Run(m, func(t *testing.T) {
			dir := fixture(t, "hook")
			okRun(t, "hook", "failure", dir, cop, m)
			wantFiles(t, "hook-fail-"+m, dir)
		})
	}
}

func TestHookPhase(t *testing.T) {
	cop := stubCopilot(t)
	t.Setenv("VLOOP_HOOK_PHASE", "second-review")
	dir := fixture(t, "hook")
	okRun(t, "hook", "forms", dir, cop)
	wantFiles(t, "hook-forms-second-review", dir)
	b, err := os.ReadFile(filepath.Join(dir, repoDir, ".vloop", "tmp", "home", "hooks", "vloop-fence.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(b), "--phase second-review ") != 2 {
		t.Errorf("want --phase second-review in both commands:\n%s", b)
	}
}

func TestHookSummary(t *testing.T) {
	cop := stubCopilot(t)
	dir := fixture(t, "hook")
	okRun(t, "hook", "forms", dir, cop)
	for _, m := range []string{"missing", "exit1", "notjson", "timeout"} {
		okRun(t, "hook", "failure", dir, cop, m)
	}
	got := strings.ReplaceAll(okRun(t, "hook", "summary", dir), dir, "@DIR@")
	if want := goldenText(t, "hook-summary", "all.md"); got != want {
		t.Errorf("hook summary differs from the golden:\n--- want\n%s\n--- got\n%s", want, got)
	}
	wantFiles(t, "hook-summary", dir)

	part := fixture(t, "hsum")
	got = strings.ReplaceAll(okRun(t, "hook", "summary", part), part, "@DIR@")
	if want := goldenText(t, "hook-summary", "partial.md"); got != want {
		t.Errorf("partial hook summary differs from the golden:\n--- want\n%s\n--- got\n%s", want, got)
	}
}

func TestCalls(t *testing.T) {
	dir := fixture(t, "calls")
	got := strings.ReplaceAll(okRun(t, "calls", filepath.Join(dir, "out.jsonl")), dir, "@DIR@")
	if want := goldenText(t, "calls", "stdout"); got != want {
		t.Errorf("calls differ from the golden:\n--- want\n%s\n--- got\n%s", want, got)
	}
}

func TestNoReviewSession(t *testing.T) {
	cop := stubCopilot(t)
	for _, args := range [][]string{{"probe"}, {"hook", "forms"}, {"hook", "failure"}} {
		dir := fixture(t, "noreview")
		args = append(args, dir, cop)
		if args[1] == "failure" {
			args = append(args, "exit1")
		}
		out, serr, code := runKit(t, args...)
		if code != 1 || out != "" || serr != "c5: no review session in the copilot log\n" {
			t.Errorf("kit %v: exit %d, stdout %q, stderr %q", args[:2], code, out, serr)
		}
		if es, _ := os.ReadDir(dir); len(es) != 1 {
			t.Errorf("kit %v wrote files into the directory: %v", args[:2], es)
		}
	}
}

func TestProbeRefusals(t *testing.T) {
	cop := stubCopilot(t)
	dir := fixture(t, "hook")
	for _, args := range [][]string{
		{"hook", "failure", dir, cop, "bogus"},
		{"hook"}, {"hook", "bogus", dir}, {"hook", "forms", dir}, {"hook", "failure", dir, cop},
		{"probe", dir}, {"summary"}, {"calls"}, {"hook", "summary"},
	} {
		out, serr, code := runKit(t, args...)
		if code != 1 || out != "" || serr == "" || strings.Count(serr, "\n") != 1 {
			t.Errorf("kit %v: exit %d, stdout %q, stderr %q", args, code, out, serr)
		}
	}
	if exists(filepath.Join(dir, "stub.log")) {
		t.Error("a refused command started a session")
	}
}
