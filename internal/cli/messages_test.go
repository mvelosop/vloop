package cli

import (
	"encoding/json"
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

func TestBrokenPlanMessages(t *testing.T) {
	t.Parallel()
	const p = ".vloop/state/state.json"
	cases := []struct {
		name, plan, want string
	}{
		{"first line", "{bad", "vloop: " + p + " is not valid JSON (line 1, column 2)\n"},
		{"later line", "{\n  \"schema\": \"state/v2\",\n  \"run_id\": bad\n}\n", "vloop: " + p + " is not valid JSON (line 3, column 13)\n"},
		{"schema", `{"version":"x"}`, "vloop: " + p + " is not a valid plan — vloop task validate lists the problems\n"},
	}
	for _, c := range cases {
		root := statusRepo(t, c.plan)
		code, out, errOut := run(t, "-C", root, "status")
		if code != 1 || out != "" || errOut != c.want {
			t.Errorf("%s: %d %q %q", c.name, code, out, errOut)
		}
		code, out, _ = run(t, "-C", root, "status", "--json")
		var v struct{ Error string }
		if code != 1 || json.Unmarshal([]byte(out), &v) != nil || v.Error != strings.TrimSuffix(strings.TrimPrefix(c.want, "vloop: "), "\n") {
			t.Errorf("%s --json: %d %q", c.name, code, out)
		}
	}
	root := statusRepo(t, "{\n  \"a\": bad}")
	for _, args := range [][]string{{"task", "validate"}, {"schema", "validate", "state/v2", p}} {
		code, out, errOut := run(t, append([]string{"-C", root}, args...)...)
		want := p + ": not valid JSON (line 2, column 8)"
		if code != 1 || !strings.Contains(out+errOut, want) || strings.Contains(out+errOut, ": :") {
			t.Errorf("%v: %d %q %q", args, code, out, errOut)
		}
	}
}

func TestPluginPathOutsideVloopRepo(t *testing.T) {
	d := scratchRepo(t)
	code, out, e := run(t, "-C", d, "plugin", "path")
	if code != 1 || out != "" || e != "vloop: not in a vloop repository — run vloop init first\n" {
		t.Fatalf("outside: %d %q %q", code, out, e)
	}
	if _, err := os.Stat(filepath.Join(d, ".vloop")); err == nil {
		t.Fatal("plugin path created .vloop")
	}
}

func TestInterventionShowSchemaOptionsBrief(t *testing.T) {
	d := scratchRepo(t)
	code, out, e := run(t, "-C", d, "intervention", "add", "no brief", "--phase", "run", "--kind", "repair", "--automatable", "no", "--by", "operator")
	if code != 0 {
		t.Fatalf("add: %d %q", code, e)
	}
	id := strings.TrimSuffix(filepath.Base(strings.TrimSpace(out)), ".md")
	code, out, e = run(t, "-C", d, "intervention", "show", id)
	if code != 0 || !strings.Contains(out, "\nschema: intervention/v2\n") || !strings.Contains(out, "\noptions: 0\n") || strings.Contains(out, "brief:") {
		t.Fatalf("show: %d %q %q", code, out, e)
	}
}
