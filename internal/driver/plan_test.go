package driver

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/state"
)

func TestInspectedPaths(t *testing.T) {
	for cmd, want := range map[string][]string{
		`grep -q TIMEOUT docs/todo.md && uv run pytest -q tests/test_thing.py`: {"docs/todo.md"},
		`cat a.txt | wc -l`:                           {"a.txt"},
		`test -f x.out && test -f y.out`:              {"x.out", "y.out"},
		`python -c "open('cfg.json')"`:                {"cfg.json"},
		`uv run pytest -q tests/test_thing.py`:        {},
		`bash scripts/check.sh`:                       {},
		`test -f a.txt && cat a.txt && test -f a.txt`: {"a.txt"},
	} {
		got := inspectedPaths(cmd)
		if len(got) == 0 && len(want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("inspectedPaths(%q) = %q, want %q", cmd, got, want)
		}
	}
}

func TestDiffRef(t *testing.T) {
	for cmd, want := range map[string]string{
		`git diff --quiet HEAD -- docs`:                          "",
		`git diff --quiet main -- a.txt`:                         "main",
		`test -z "$(git diff --name-only HEAD~1 -- docs)"`:       "HEAD~1",
		`test "$(git rev-list --count origin/main..HEAD)" -gt 0`: "origin/main",
		`git diff --quiet "$(git merge-base HEAD main)" -- docs`: "$(git merge-base HEAD main)",
		`git log --format=%H v1.0..HEAD -- docs`:                 "v1.0",
		`git status --porcelain`:                                 "",
		`test -f a.txt`:                                          "",
	} {
		if got := diffRef(cmd); got != want {
			t.Errorf("diffRef(%q) = %q, want %q", cmd, got, want)
		}
	}
}

func TestHeadDiffAdvisory(t *testing.T) {
	p := &state.Plan{Tasks: []state.Task{
		{ID: "T1", Verify: `git diff --quiet HEAD -- docs`},
		{ID: "T2", Verify: `test -n "$VLOOP_GATE_TASK" && git diff --quiet HEAD -- docs`},
		{ID: "T3", Verify: `git diff --name-only --cached HEAD`},
		{ID: "T4", Verify: `git diff --quiet main`},
	}}
	if got := HeadDiffAdvisory(p); !reflect.DeepEqual(got, []string{"T1", "T3"}) {
		t.Errorf("advisory = %q, want T1 and T3", got)
	}
}

func TestGateShapeProblems(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.name", "t")
	run("config", "user.email", "t@example.com")
	if err := os.WriteFile(filepath.Join(root, "todo.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-q", "-m", "x")

	p := &state.Plan{Tasks: []state.Task{
		{ID: "T3", Verify: `grep -q x untracked.md`},
		{ID: "T4", Verify: `git diff --quiet v1 -- .`},
		{ID: "T5", Verify: `test -f T5.out`},
	}}
	got := GateShapeProblems(root, p)
	if len(got) != 1 || !strings.Contains(got[0], "T4") || !strings.Contains(got[0], "v1") {
		t.Errorf("problems = %q", got)
	}
	for _, g := range got {
		if !strings.HasPrefix(g, "gate shape rejected: ") {
			t.Errorf("a problem without its heading: %q", g)
		}
	}
}

func TestRefsDiff(t *testing.T) {
	a := []string{"HEAD -> refs/heads/x", "refs/heads/x aaa ", "refs/heads/y bbb "}
	if got := refsDiff(a, a); len(got) != 0 {
		t.Errorf("same refs differ: %q", got)
	}
	b := []string{"HEAD -> refs/heads/y", "refs/heads/x aaa ", "refs/heads/y bbb "}
	got := refsDiff(a, b)
	if len(got) != 2 || !strings.Contains(got[0], "before") || !strings.Contains(got[1], "after") {
		t.Errorf("moved HEAD: %q", got)
	}
}

func TestNewRunDirSuffixesASecondInTheSameSecond(t *testing.T) {
	root := t.TempDir()
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	first, err := newRunDir(root, "B1", at)
	if err != nil {
		t.Fatal(err)
	}
	second, err := newRunDir(root, "B1", at)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(first) != "20260102-030405" || filepath.Base(second) != "20260102-030405-2" {
		t.Errorf("folders %s, %s", filepath.Base(first), filepath.Base(second))
	}
	if _, err := os.Stat(filepath.Join(second, "sessions")); err != nil {
		t.Error("no sessions folder")
	}
}

func TestResolveBudgetsPrecedence(t *testing.T) {
	clearEnv(t)
	for _, k := range []string{"VLOOP_RUN_MAX_ITERATIONS", "VLOOP_RUN_COST_CEILING"} {
		t.Setenv(k, "")
	}
	root := t.TempDir()
	b, err := ResolveBudgets(root, nil)
	if err != nil || b.MaxIterations != 30 || b.CostCeiling != 40 || b.MaxAttempts != 3 || b.StallLimit != 2 ||
		b.ConvergenceMax != 3 || b.ConvergenceMin != 6 {
		t.Fatalf("defaults: %+v, %v", b, err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".vloop"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".vloop", "config.toml"), []byte("[run]\nmax-iterations = 7\ncost-ceiling = 12.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if b, _ = ResolveBudgets(root, nil); b.MaxIterations != 7 || b.CostCeiling != 12.5 {
		t.Errorf("file: %+v", b)
	}
	t.Setenv("VLOOP_RUN_MAX_ITERATIONS", "9")
	if b, _ = ResolveBudgets(root, nil); b.MaxIterations != 9 || b.CostCeiling != 12.5 {
		t.Errorf("environment: %+v", b)
	}
	if b, _ = ResolveBudgets(root, map[string]string{"run.max-iterations": "0"}); b.MaxIterations != 0 {
		t.Errorf("flag: %+v", b)
	}
}

func TestNewRunDirIsNamedInUTC(t *testing.T) {
	at := time.Date(2026, 1, 2, 8, 34, 5, 0, time.FixedZone("IST", 5*3600+1800))
	dir, err := newRunDir(t.TempDir(), "B1", at)
	if err != nil {
		t.Fatal(err)
	}
	if got := filepath.Base(dir); got != "20260102-030405" {
		t.Errorf("folder %s, want the UTC stamp 20260102-030405", got)
	}
}
