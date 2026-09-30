package metrics

import (
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/runs"
)

type repo struct {
	t    *testing.T
	root string
	n    int
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	r := &repo{t: t, root: t.TempDir()}
	r.git("init", "-q", "-b", "main", ".")
	r.git("config", "user.name", "t")
	r.git("config", "user.email", "t@example.com")
	r.git("config", "commit.gpgsign", "false")
	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.root}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_DATE=2026-01-01T09:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T09:00:00Z")
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (r *repo) write(rel, body string) {
	r.t.Helper()
	p := filepath.Join(r.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *repo) commit(subject string) string {
	r.git("add", "-A")
	r.git("commit", "-q", "--allow-empty", "-m", subject)
	return r.git("rev-parse", "HEAD")
}

// diff commits the current tree and counts the change against the previous commit.
func (r *repo) diff(subject string) Lines {
	r.t.Helper()
	sha := r.commit(subject)
	l, err := DiffLines(r.root, sha+"^", sha, classify.New(classify.Preset{}, []string{"go"}))
	if err != nil {
		r.t.Fatal(err)
	}
	return l
}

func nonBlank(n int) string { return strings.Repeat("x := 1\n", n) }

func TestLinesBlankAndWhitespaceOnly(t *testing.T) {
	r := newRepo(t)
	r.write("keep.txt", "seed\n")
	r.commit("seed")
	r.write("a.go", nonBlank(3)+"\n\n   \n\t\n"+nonBlank(2))
	if l := r.diff("add"); l.Added.Code != 5 {
		t.Errorf("blank lines counted: %+v", l)
	}
	// Control: a file without blank lines counts every line.
	r.write("b.go", nonBlank(4))
	if l := r.diff("add b"); l.Added.Code != 4 {
		t.Errorf("plain lines: %+v", l)
	}
}

func TestLinesCategoriesAndDeletions(t *testing.T) {
	r := newRepo(t)
	r.write("a.go", nonBlank(6)+"\n")
	r.commit("seed")
	r.write("a.go", nonBlank(4)+"  \n"+"y := 2\nz := 3\nw := 4\nv := 5\n")
	r.write("a_test.go", nonBlank(2))
	r.write("README.md", "# t\n")
	r.write("notes.txt", "n\n")
	l := r.diff("change")
	want := Lines{Added: Counts{Code: 4, Test: 2, Docs: 1, Other: 1}, Deleted: Counts{Code: 6 - 4}}
	if l != want {
		t.Errorf("got %+v want %+v", l, want)
	}
}

func TestLinesBinaryNotCounted(t *testing.T) {
	r := newRepo(t)
	r.write("seed.txt", "s\n")
	r.commit("seed")
	r.write("blob.go", "a\x00b\nc\nd\n")
	if l := r.diff("binary"); l != (Lines{}) {
		t.Errorf("binary file counted: %+v", l)
	}
	r.write("text.go", "a\nb\nc\n")
	if l := r.diff("text"); l.Added.Code != 3 {
		t.Errorf("text file: %+v", l)
	}
}

func TestLinesRenameCountsContentChangeOnly(t *testing.T) {
	r := newRepo(t)
	body := nonBlank(20)
	r.write("old.go", body)
	r.commit("seed")
	if err := os.Rename(filepath.Join(r.root, "old.go"), filepath.Join(r.root, "new.go")); err != nil {
		t.Fatal(err)
	}
	if l := r.diff("pure rename"); l != (Lines{}) {
		t.Errorf("pure rename counted: %+v", l)
	}
	r.write("new.go", body+"z := 3\n")
	r.commit("edit")
	if err := os.Rename(filepath.Join(r.root, "new.go"), filepath.Join(r.root, "newer.go")); err != nil {
		t.Fatal(err)
	}
	r.write("newer.go", body+"z := 4\n")
	if l := r.diff("rename and edit"); l.Added.Code != 1 || l.Deleted.Code != 1 {
		t.Errorf("rename with edit: %+v", l)
	}
}

func TestLinesExcludedPathsNeverCounted(t *testing.T) {
	r := newRepo(t)
	r.write("seed.txt", "s\n")
	r.commit("seed")
	r.write(".loop/state/state.json", nonBlank(5))
	r.write(".vloop/config.toml", nonBlank(5))
	r.write("go.sum", nonBlank(5))
	if l := r.diff("excluded"); l != (Lines{}) {
		t.Errorf("excluded paths counted: %+v", l)
	}
	r.write("sub/.loop/x.go", nonBlank(2)) // not the top-level .loop/: counts
	if l := r.diff("nested"); l.Added.Code != 2 {
		t.Errorf("nested .loop path: %+v", l)
	}
}

// worked builds the brief's worked-example history.
func worked(t *testing.T) (*repo, *runs.Owned) {
	r := newRepo(t)
	r.write("README.md", "seed\n")
	r.commit("seed")
	r.write(".loop/state/state.json", `{"run_id":"B1","tasks":[]}`)
	r.commit("[loop] plan B1")
	r.write("main.go", nonBlank(10)+"\n\n")
	r.write("main_test.go", nonBlank(6))
	r.write("README.md", "seed\na\nb\nc\n")
	r.commit("[loop] T1: done")
	r.write("util.go", nonBlank(4))
	r.commit("[loop] T2: gate_fail")
	r.write("util.go", nonBlank(2)+"a\nb\nc\n")
	r.commit("[loop] T2: done")
	r.commit("[loop] run B1/20260101-090000: complete")
	o, err := runs.OwnedCommits(r.root, "B1")
	if err != nil || o == nil {
		t.Fatalf("owned: %v %v", o, err)
	}
	return r, o
}

func TestReworkWorkedExample(t *testing.T) {
	r, o := worked(t)
	s, err := Measure(r.root, o, classify.New(classify.Preset{}, []string{"go"}))
	if err != nil {
		t.Fatal(err)
	}
	if d := s.Delivered.Added; d != (Counts{Code: 15, Test: 6, Docs: 3}) {
		t.Errorf("delivered %+v", d)
	}
	if s.Churn.Added != (Counts{Code: 17, Test: 6, Docs: 3}) {
		t.Errorf("churn %+v", s.Churn.Added)
	}
	if len(s.ByTask) != 2 || s.ByTask[0].Task != "T1" || s.ByTask[0].Lines.Added.Code != 10 ||
		s.ByTask[1].Task != "T2" || s.ByTask[1].Commits != 2 || s.ByTask[1].Lines.Added.Code != 7 {
		t.Errorf("by task %+v", s.ByTask)
	}
	if s.Delivered.Deleted.Code != 0 || s.Churn.Deleted.Code != 2 {
		t.Errorf("deletions: delivered %+v churn %+v", s.Delivered.Deleted, s.Churn.Deleted)
	}
	if rw := s.Rework(); rw == nil || math.Abs(*rw-17.0/15) > 1e-9 {
		t.Errorf("rework %v", rw)
	}
	if tc := s.TestCode(); tc == nil || math.Abs(*tc-0.4) > 1e-9 {
		t.Errorf("test:code %v", tc)
	}
}

func TestReworkZeroDeliveredCodeIsUnknown(t *testing.T) {
	var s Size
	s.Delivered.Added.Test = 5
	s.Churn.Added.Code = 3
	if s.Rework() != nil || s.TestCode() != nil {
		t.Errorf("zero denominator must be nil: %v %v", s.Rework(), s.TestCode())
	}
	s.Delivered.Added.Code = 10
	if s.Rework() == nil || s.TestCode() == nil {
		t.Error("non-zero denominator must give a ratio")
	}
}
