package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func addRec(t *testing.T, dir, phase, kind, auto string, extra ...string) {
	t.Helper()
	args := append([]string{"intervention", "add", "s", "--phase", phase, "--kind", kind, "--automatable", auto, "--by", "both"}, extra...)
	if code, _, e := runDefect(t, dir, args...); code != 0 {
		t.Fatalf("add %v: %d %q", args, code, e)
	}
}

func TestMetricsInterventionsTables(t *testing.T) {
	dir := t.TempDir()
	two := []string{"--option", "x", "--option", "y", "--why", "w"}
	addRec(t, dir, "halt", "repair", "partly", append(two, "--recommended", "1", "--decided-option", "1")...)
	addRec(t, dir, "verify", "repair", "yes", append(two, "--recommended", "1", "--decided-option", "2")...)
	addRec(t, dir, "verify", "repair", "no", append(two, "--recommended", "1", "--decided-option", "1", "--adjusted")...)
	addRec(t, dir, "verify", "repair", "yes")
	addRec(t, dir, "design", "decision", "yes", "--option", "x", "--recommended", "1", "--why", "w", "--decided-other", "--decided", "z")
	addRec(t, dir, "run", "halt", "no")

	_, out, _ := runDefect(t, dir, "metrics", "--interventions")
	rows := strings.Split(collapse(out), "\n")
	if len(rows) != 10 {
		t.Fatalf("by kind: %d rows\n%s", len(rows), out)
	}
	want := map[int]string{
		0: "kind n recommended share other-option adjusted different no-options automatable-yes automatable-partly",
		1: "direction 0 0 n/a 0 0 0 0 0 0",
		2: "decision 1 0 0% 0 0 1 0 1 0",
		4: "halt 1 0 n/a 0 0 0 1 0 0",
		6: "repair 4 1 33% 1 1 0 1 2 1",
		9: "total 6 1 25% 1 1 1 2 3 1",
	}
	for i, w := range want {
		if rows[i] != w {
			t.Errorf("kind row %d: %q, want %q", i, rows[i], w)
		}
	}
	_, a, _ := runDefect(t, dir, "metrics", "--interventions", "--by", "kind")
	if a != out {
		t.Errorf("--by kind differs from the default")
	}

	_, out, _ = runDefect(t, dir, "metrics", "--interventions", "--by", "phase")
	rows = strings.Split(collapse(out), "\n")
	if len(rows) != 9 || rows[0] != "phase n recommended share other-option adjusted different no-options automatable-yes automatable-partly" {
		t.Fatalf("by phase:\n%s", out)
	}
	for i, w := range map[int]string{
		2: "design 1 0 0% 0 0 1 0 1 0",
		3: "run 1 0 n/a 0 0 0 1 0 0",
		4: "halt 1 1 100% 0 0 0 0 0 1",
		5: "verify 3 0 0% 1 1 0 1 2 0",
		8: "total 6 1 25% 1 1 1 2 3 1",
	} {
		if rows[i] != w {
			t.Errorf("phase row %d: %q, want %q", i, rows[i], w)
		}
	}

	_, out, _ = runDefect(t, dir, "--json", "metrics", "--interventions")
	if !strings.Contains(out, `"repair"`) {
		t.Errorf("json: %q", out)
	}
}

func TestMetricsInterventionsByTaskRefused(t *testing.T) {
	code, out, e := runDefect(t, t.TempDir(), "metrics", "--interventions", "--by", "task")
	if code != 2 || out != "" || e != "vloop: --by takes kind or phase with --interventions\n" {
		t.Errorf("%d %q %q", code, out, e)
	}
}

func TestMetricsInterventionsWorkspace(t *testing.T) {
	ws := t.TempDir()
	for _, n := range []string{"a", "b"} {
		if err := os.MkdirAll(filepath.Join(ws, n, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	two := []string{"--option", "x", "--option", "y", "--why", "w"}
	addRec(t, filepath.Join(ws, "a"), "halt", "repair", "yes", append(two, "--recommended", "1", "--decided-option", "1")...)
	addRec(t, filepath.Join(ws, "b"), "close", "ceremony", "partly", append(two, "--recommended", "1", "--decided-option", "2")...)
	addRec(t, filepath.Join(ws, "b"), "halt", "repair", "no")
	f := wsFile(t, ws, "[[repo]]\npath = \"a\"\nname = \"a\"\n[[repo]]\npath = \"b\"\nname = \"b\"\n")
	code, out, e := runCLI(t, t.TempDir(), "metrics", "--workspace", f, "--interventions")
	if code != 0 {
		t.Fatalf("%d %q", code, e)
	}
	rows := strings.Split(collapse(out), "\n")
	if rows[0] != "repo kind n recommended share other-option adjusted different no-options automatable-yes automatable-partly" {
		t.Fatalf("header %q", rows[0])
	}
	got := strings.Join(rows, "\n")
	for _, w := range []string{"\na repair 1 1 100% 0 0 0 0 1 0", "\nb ceremony 1 0 0% 1 0 0 0 0 1", "\nb repair 1 0 n/a 0 0 0 1 0 0"} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in\n%s", w, got)
		}
	}
	if last := rows[len(rows)-1]; last != "total 3 1 50% 1 0 0 1 1 1" {
		t.Errorf("total row %q", last)
	}
}
