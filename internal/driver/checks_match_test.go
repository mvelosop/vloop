package driver

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/mvelosop/vloop/internal/state"
)

func TestMatchGlob(t *testing.T) {
	for _, c := range []struct {
		glob, path string
		want       bool
	}{
		{"**", "a.go", true},
		{"**", "a/b/c.go", true},
		{"api/**", "api/x.go", true},
		{"api/**", "api/a/b/x.go", true},
		{"api/**", "web/x.go", false},
		{"api/**", "apix/x.go", false},
		{"**/*.go", "a/b/c.go", true},
		{"**/*.go", "c.go", true},
		{"**/*.go", "a/c.txt", false},
		{"*.go", "a/c.go", false},
		{"go.mod", "go.mod", true},
	} {
		if got := matchGlob(c.glob, c.path); got != c.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", c.glob, c.path, got, c.want)
		}
	}
}

func TestChecksForKeepsConfigOrder(t *testing.T) {
	checks := []state.PlanCheck{
		{Name: "api", Paths: []string{"api/**"}},
		{Name: "web", Paths: []string{"web/**"}},
		{Name: "docs", Paths: []string{"**/*.md"}},
	}
	var names []string
	for _, c := range checksFor(checks, []string{"web/a.ts", "README.md"}) {
		names = append(names, c.Name)
	}
	if !slices.Equal(names, []string{"web", "docs"}) {
		t.Errorf("checksFor = %v, want [web docs]", names)
	}
	if got := checksFor(checks, nil); len(got) != 0 {
		t.Errorf("no changed path selected %v", got)
	}
}

// A failed attempt is committed, so a retry that changes nothing new must
// still see what this task's earlier attempts changed — or its check is
// skipped and the task goes done on the work the check failed.
func TestChangedPathsIncludeTheTasksEarlierAttempts(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	commit := func(file, subject string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(subject+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		run("add", "-A")
		run("commit", "-q", "-m", subject)
	}
	run("init", "-q")
	run("config", "user.name", "t")
	run("config", "user.email", "t@example.com")
	commit("README.md", "base")
	commit(".vloop/state/state.json", "[vloop] plan B1")
	commit("api/a.go", "[vloop] T1: done")
	commit("web/b.ts", "[vloop] T2: check_failed")
	commit("web/c.ts", "[vloop] T2: rejected")
	commit("notes.md", "Operator: reset T2")

	got := changedPaths(root, "T2")
	for _, want := range []string{"web/b.ts", "web/c.ts", "notes.md"} {
		if !slices.Contains(got, want) {
			t.Errorf("changedPaths(T2) = %v, missing %s", got, want)
		}
	}
	for _, not := range []string{"api/a.go", "README.md", ".vloop/state/state.json"} {
		if slices.Contains(got, not) {
			t.Errorf("changedPaths(T2) = %v, includes %s", got, not)
		}
	}
	if got := changedPaths(root, "T3"); slices.Contains(got, "web/b.ts") {
		t.Errorf("changedPaths(T3) = %v, includes T2's work", got)
	}
}
