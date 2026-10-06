package cli

import (
	"os/exec"
	"reflect"

	"encoding/json"
	"github.com/mvelosop/vloop/internal/intervention"
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
		var head struct {
			Schema string `json:"schema"`
		}
		if err := json.Unmarshal(v, &head); err != nil {
			t.Fatal(err)
		}
		var so, se strings.Builder
		if code := Execute(Build{}, []string{"schema", "validate", head.Schema, f}, &so, &se); code != 0 {
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

func TestInterventionSetChoice(t *testing.T) {
	dir := t.TempDir()
	_, out, _ := runDefect(t, dir, "intervention", "add", "Pick", "--phase", "run", "--kind", "decision", "--automatable", "no", "--by", "both",
		"--option", "a", "--option", "b", "--recommended", "1", "--why", "w", "--decided-option", "1")
	p := strings.TrimSpace(out)
	id := strings.TrimSuffix(filepath.Base(p), ".md")
	read := func() string { b, _ := os.ReadFile(filepath.Join(dir, p)); return string(b) }
	set := func(f, v string) int { c, _, _ := runDefect(t, dir, "intervention", "set", id, f, v); return c }
	for _, c := range []struct{ f, v, want string }{
		{"decided", "2", "agreement: other-option\n"},
		{"adjusted", "true", "agreement: adjusted\n"},
		{"adjusted", "false", "agreement: other-option\n"},
		{"recommended", "2", "agreement: recommended\n"},
		{"decided", "other", "agreement: different\n"},
	} {
		if code := set(c.f, c.v); code != 0 || !strings.Contains(read(), c.want) {
			t.Fatalf("set %s %s: %d\n%s", c.f, c.v, code, read())
		}
	}
	if strings.Contains(read(), "adjusted:") {
		t.Fatalf("adjusted lingers:\n%s", read())
	}
	before := read()
	for _, c := range [][2]string{{"decided", "3"}, {"recommended", "3"}, {"adjusted", "maybe"}} {
		if code := set(c[0], c[1]); code == 0 || read() != before {
			t.Fatalf("set %s %s accepted or changed the file", c[0], c[1])
		}
	}
	for f, msg := range map[string]string{"agreement": "vloop: agreement is derived — set decided, adjusted or recommended instead\n", "options": "vloop: options are recorded with the intervention, not set\n"} {
		code, _, e := runDefect(t, dir, "intervention", "set", id, f, "1")
		if code != 2 || e != msg || read() != before {
			t.Fatalf("set %s: %d %q", f, code, e)
		}
	}
}

func TestInterventionMigrate(t *testing.T) {
	dir := t.TempDir()
	idir := filepath.Join(dir, ".vloop", "interventions")
	os.MkdirAll(idir, 0o755)
	v1 := "---\nid: I20260101-0900-old\nbrief: \"\"\nphase: run\nkind: halt\nautomatable: yes\nby: operator\noccurred: 2026-01-01\nrecorded: 2026-01-01T09:00:00Z\nbackfilled: true\n---\nold  \n\n**Context.** keep   spacing\n"
	want := strings.Replace(v1, "by: operator\n", "by: operator\nschema: intervention/v2\noptions: 0\nrecommended: 0\ndecided: \"\"\nagreement: no-options\n", 1)
	path := filepath.Join(idir, "I20260101-0900-old.md")
	os.WriteFile(path, []byte(v1), 0o644)
	read := func() string { b, _ := os.ReadFile(path); return string(b) }

	code, out, _ := runDefect(t, dir, "intervention", "migrate", "--dry-run")
	if code != 0 || strings.TrimSpace(out) != ".vloop/interventions/I20260101-0900-old.md" || read() != v1 {
		t.Fatalf("dry run: %d %q", code, out)
	}
	code, out, _ = runDefect(t, dir, "intervention", "migrate")
	if code != 0 || out != "migrated 1 record(s)\n" || read() != want {
		t.Fatalf("migrate: %d %q\n%s", code, out, read())
	}
	code, out, _ = runDefect(t, dir, "intervention", "migrate")
	if code != 0 || out != "nothing to migrate\n" || read() != want {
		t.Fatalf("again: %d %q", code, out)
	}
}

func TestInterventionShowLinks(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", dir}, args...)...)
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid", "GIT_CONFIG_GLOBAL=/dev/null")
		b, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, b)
		}
		return strings.TrimSpace(string(b))
	}
	git("init", "-q", "-b", "main")
	git("commit", "-q", "--allow-empty", "-m", "the subject for show")
	sha := git("rev-parse", "--short=12", "HEAD")
	os.MkdirAll(filepath.Join(dir, ".vloop/state/runs/B1-run/20260101-090000"), 0o755)
	os.MkdirAll(filepath.Join(dir, ".vloop/state/runs/B20260101-0900-a/20260101-090000"), 0o755)
	plan := func(title string) {
		os.WriteFile(filepath.Join(dir, ".vloop/state/state.json"), []byte(`{"run_id":"B20260101-0900-a","tasks":[{"id":"T3","title":"`+title+`"}]}`), 0o644)
	}
	plan("Old title")
	git("add", "-A")
	git("commit", "-q", "-m", "[vloop] plan B20260101-0900-a")
	plan("Latest title")
	git("add", "-A")
	git("commit", "-q", "-m", "[vloop] plan B20260101-0900-a")
	plan("Drifted")
	git("add", "-A")
	git("commit", "-q", "-m", "a later change")

	os.MkdirAll(filepath.Join(dir, "docs/briefs"), 0o755)
	os.WriteFile(filepath.Join(dir, "docs/briefs/B20260101-0900-a.loop-brief.md"), []byte("---\nname: B20260101-0900-a.loop-brief\nstatus: ready\n---\n"), 0o644)
	_, out, _ := runDefect(t, dir, "defect", "add", "the defect", "--found-by", "user", "--brief", "B20260101-0900-a.loop-brief")
	did := strings.TrimSuffix(filepath.Base(strings.TrimSpace(out)), ".md")
	_, out, _ = runDefect(t, dir, "intervention", "add", "the other one", "--phase", "run", "--kind", "repair", "--automatable", "no", "--by", "operator")
	oid := strings.TrimSuffix(filepath.Base(strings.TrimSpace(out)), ".md")
	ctx := "T3 in B20260101-0900-a; " + sha + " then " + did + " and " + oid + " and T3 again, not D20990101-0000-nope."
	_, out, _ = runDefect(t, dir, "intervention", "add", "Show me", "--brief", "B20260101-0900-a.loop-brief", "--phase", "halt", "--kind", "repair", "--automatable", "no", "--by", "both", "--context", ctx+" Note T9 also.")
	id := strings.TrimSuffix(filepath.Base(strings.TrimSpace(out)), ".md")

	want := []intervention.Link{
		{Ref: "T3", Kind: "task", Title: "Latest title"},
		{Ref: "B20260101-0900-a", Kind: "run", Title: ".vloop/state/runs/B20260101-0900-a/20260101-090000"},
		{Ref: sha, Kind: "commit", Title: "the subject for show"},
		{Ref: did, Kind: "defect", Title: "the defect"},
		{Ref: oid, Kind: "intervention", Title: "the other one"},
		{Ref: "D20990101-0000-nope", Kind: "defect", Title: ""},
		{Ref: "T9", Kind: "task", Title: ""},
	}
	code, out, _ := runDefect(t, dir, "--json", "intervention", "show", id)
	var got struct {
		ID, Summary, Agreement string
		Links                  []intervention.Link
	}
	if code != 0 || json.Unmarshal([]byte(out), &got) != nil || got.ID != id || got.Summary != "Show me" || got.Agreement != "no-options" {
		t.Fatalf("json %d %q", code, out)
	}
	if !reflect.DeepEqual(got.Links, want) {
		t.Fatalf("links %+v\nwant %+v", got.Links, want)
	}

	code, out, _ = runDefect(t, dir, "intervention", "show", id)
	i := strings.Index(out, "links:\n")
	if code != 0 || i < 0 || !strings.HasPrefix(out, id+"\nShow me\n") || !strings.Contains(out, "\nContext\n") {
		t.Fatalf("text %d %q", code, out)
	}
	lines := strings.Split(strings.TrimSpace(out[i+len("links:\n"):]), "\n")
	if len(lines) != len(want) {
		t.Fatalf("links text %q", lines)
	}
	for n, l := range lines {
		title := want[n].Title
		if title == "" {
			title = "(not found)"
		}
		if strings.TrimSpace(l) != want[n].Ref+"  "+title {
			t.Fatalf("link %d %q", n, l)
		}
	}

	code, out, errs := runDefect(t, dir, "intervention", "show", "I20990101-0000-none")
	if code != ExitProblems || out != "" || errs != "vloop: no intervention I20990101-0000-none\n" {
		t.Fatalf("unknown %d %q %q", code, out, errs)
	}
	code, out, _ = runDefect(t, dir, "--json", "intervention", "show", "I20990101-0000-none")
	var e struct{ Error string }
	if code != ExitProblems || json.Unmarshal([]byte(out), &e) != nil || e.Error != "no intervention I20990101-0000-none" {
		t.Fatalf("unknown json %d %q", code, out)
	}
}
