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
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "setup  repair  yes  no-options  I") || lines[1] != "run  halt  partly  no-options  "+id {
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

func TestInterventionAddOptions(t *testing.T) {
	base := []string{"intervention", "add", "x", "--phase", "run", "--kind", "halt", "--automatable", "yes", "--by", "operator"}
	opts := func(extra ...string) []string { return append(append([]string{}, base...), extra...) }
	refusals := []struct {
		msg   string
		extra []string
	}{
		{"at most three options", []string{"--option", "a", "--option", "b", "--option", "c", "--option", "d", "--recommended", "1", "--why", "w", "--decided-option", "1"}},
		{"options need --recommended and --why", []string{"--option", "a", "--why", "w", "--decided-option", "1"}},
		{"options need --recommended and --why", []string{"--option", "a", "--recommended", "1", "--decided-option", "1"}},
		{"--recommended must name an option, 1 to 2", []string{"--option", "a", "--option", "b", "--recommended", "3", "--why", "w", "--decided-option", "1"}},
		{"options need --decided-option or --decided-other", []string{"--option", "a", "--recommended", "1", "--why", "w"}},
		{"--decided-option must name an option, 1 to 1", []string{"--option", "a", "--recommended", "1", "--why", "w", "--decided-option", "2"}},
		{"--decided-option and --decided-other exclude each other", []string{"--option", "a", "--recommended", "1", "--why", "w", "--decided-option", "1", "--decided-other"}},
		{"--adjusted needs --decided-option", []string{"--option", "a", "--recommended", "1", "--why", "w", "--decided-other", "--adjusted"}},
		{"--recommended, --decided-option and --decided-other need options", []string{"--decided-other"}},
		{"--recommended, --decided-option and --decided-other need options", []string{"--recommended", "1", "--why", "w"}},
	}
	for _, r := range refusals {
		dir := t.TempDir()
		code, out, e := runDefect(t, dir, opts(r.extra...)...)
		if code != 2 || out != "" || e != "vloop: "+r.msg+"\n" {
			t.Errorf("%v: %d %q %q", r.extra, code, out, e)
		}
		if _, err := os.Stat(filepath.Join(dir, ".vloop", "interventions")); err == nil {
			t.Errorf("%v: a refused add wrote files", r.extra)
		}
	}

	cases := []struct {
		agreement string
		extra     []string
	}{
		{"recommended", []string{"--decided-option", "1"}},
		{"other-option", []string{"--decided-option", "2"}},
		{"adjusted", []string{"--decided-option", "1", "--adjusted"}},
		{"different", []string{"--decided-other"}},
	}
	for _, c := range cases {
		dir := t.TempDir()
		args := opts(append([]string{"--option", "a", "--option", "b", "--recommended", "1", "--why", "w", "--decided", "d"}, c.extra...)...)
		if code, _, e := runDefect(t, dir, args...); code != 0 {
			t.Fatalf("%s: %d %q", c.agreement, code, e)
		}
		_, out, _ := runDefect(t, dir, "--json", "intervention", "list")
		var vs []map[string]any
		if json.Unmarshal([]byte(out), &vs) != nil || len(vs) != 1 || vs[0]["agreement"] != c.agreement || len(vs[0]["options"].([]any)) != 2 {
			t.Errorf("%s: %q", c.agreement, out)
		}
	}
}
