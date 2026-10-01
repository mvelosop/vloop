package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInterventionAddListSet(t *testing.T) {
	dir := t.TempDir()
	add := []string{"intervention", "add", "Fix it", "--phase", "run", "--kind", "halt", "--automatable", "partly", "--by", "operator"}
	code, out, _ := runDefect(t, dir, add...)
	p := strings.TrimSpace(out)
	if code != 0 || !strings.HasPrefix(p, ".vloop/interventions/I") || !strings.HasSuffix(p, "-fix-it.md") {
		t.Fatalf("add: %d %q", code, out)
	}
	id := strings.TrimSuffix(filepath.Base(p), ".md")
	runDefect(t, dir, "intervention", "add", "Plan", "--phase", "setup", "--kind", "repair", "--automatable", "yes", "--by", "both", "--brief", "B1.loop-brief")

	_, out, _ = runDefect(t, dir, "intervention", "list")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "setup  repair  yes  I") || lines[1] != "run  halt  partly  "+id {
		t.Fatalf("list %q", out)
	}
	_, out, _ = runDefect(t, dir, "intervention", "list", "--brief", "B1.loop-brief")
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("filtered %q", out)
	}
	_, out, _ = runDefect(t, dir, "--json", "intervention", "list")
	var vs []map[string]any
	if json.Unmarshal([]byte(out), &vs) != nil || len(vs) != 2 || vs[1]["id"] != id {
		t.Fatalf("json %q", out)
	}

	if code, _, _ := runDefect(t, dir, "intervention", "set", id, "kind", "ceremony"); code != 0 {
		t.Fatalf("set code %d", code)
	}
	b, _ := os.ReadFile(filepath.Join(dir, p))
	if !strings.Contains(string(b), "kind: ceremony\n") {
		t.Fatalf("set: %s", b)
	}
	if code, _, _ := runDefect(t, dir, "intervention", "set", "I0-none", "kind", "halt"); code != 1 {
		t.Fatalf("unknown id code %d", code)
	}
}

func TestInterventionAddRefusals(t *testing.T) {
	dir := t.TempDir()
	code, _, e := runDefect(t, dir, "intervention", "add", "x", "--phase", "bogus", "--kind", "halt", "--automatable", "yes", "--by", "operator")
	if code != 2 || e != "vloop: invalid value \"bogus\" for phase: want one of setup, design, run, halt, verify, close, next\n" {
		t.Fatalf("%d %q", code, e)
	}
	if code, _, _ := runDefect(t, dir, "intervention", "add", "x", "--kind", "halt", "--automatable", "yes", "--by", "operator"); code != 2 {
		t.Fatalf("missing phase: %d", code)
	}
	if _, err := os.Stat(filepath.Join(dir, ".vloop", "interventions")); err == nil {
		t.Fatal("a refused add wrote files")
	}
}

func TestExistingInterventionsValidate(t *testing.T) {
	var o, e strings.Builder
	if code := Execute(Build{}, []string{"-C", "../..", "--json", "intervention", "list"}, &o, &e); code != 0 {
		t.Fatal(e.String())
	}
	var vs []json.RawMessage
	if err := json.Unmarshal([]byte(o.String()), &vs); err != nil || len(vs) == 0 {
		t.Fatalf("%v %d", err, len(vs))
	}
	for _, v := range vs {
		f := filepath.Join(t.TempDir(), "i.json")
		os.WriteFile(f, v, 0o644)
		var so, se strings.Builder
		if code := Execute(Build{}, []string{"schema", "validate", "intervention/v1", f}, &so, &se); code != 0 {
			t.Errorf("%s: %s%s", v, so.String(), se.String())
		}
	}
}
