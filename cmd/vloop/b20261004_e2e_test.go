package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestWorkedExampleInterventionsOptions plays the interventions-options
// brief's worked example in one temporary repository with vloop init done.
// Ids, paths and HEAD's sha are computed from what the commands print.

const v1Record = `---
id: I20260101-0700-a-v1-record-with-context-and-decided
brief: B1
phase: design
kind: decision
automatable: no
by: operator
occurred: 2026-01-01
recorded: 2026-01-01T07:00:00Z
---
a v1 record with context and decided

**Trigger.** the draft had two open forks

**Done.** settled both

**Context.** the B1 draft, forks one and two

**Decided.** the operator took the narrower scope

**What would automate it.** nothing
`

var interventionPath = regexp.MustCompile(`^\.vloop/interventions/I[0-9-]+-t3-s-gate-failed-on-a-path-typo.*\.md$`)

// frontValue returns the frontmatter value of key in a record, or "" if absent.
func frontValue(t *testing.T, text, key string) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	for _, l := range lines[1:] {
		if l == "---" {
			break
		}
		if strings.HasPrefix(l, key+": ") {
			return strings.TrimPrefix(l, key+": ")
		}
	}
	return ""
}

func body(text string) string {
	parts := strings.SplitN(text, "\n---\n", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func TestWorkedExampleInterventionsOptions(t *testing.T) {
	s := newScratch(t)
	git := func(args ...string) string {
		t.Helper()
		full := append([]string{"-c", "user.name=e2e", "-c", "user.email=e2e@example.invalid", "-c", "commit.gpgsign=false"}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = s.dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("add", "-A")
	git("commit", "-q", "-m", "the worked example's head")
	sha := git("rev-parse", "HEAD")
	if r := s.run(nil, "init"); r.code != 0 {
		t.Fatalf("vloop init: %d %s%s", r.code, r.out, r.err)
	}

	add := func(extra ...string) result {
		args := []string{"intervention", "add", "T3's gate failed on a path typo", "--brief", "B1", "--phase", "halt",
			"--kind", "repair", "--automatable", "partly", "--by", "both", "--trigger", "vloop run exited 2",
			"--done", "gate replaced", "--context", "T3 in run B20260101-0900-a; see D20260101-0900-x and commit " + sha}
		return s.run(nil, append(args, extra...)...)
	}
	three := func(extra ...string) result {
		return add(append([]string{"--option", "replace the gate with vloop task verify", "--option", "reset T3 and retry",
			"--option", "abandon the brief", "--recommended", "1", "--why", "the gate, not the work, is wrong"}, extra...)...)
	}
	made := func(label string, r result, options, recommended, decided, agreement string) (rel, text string) {
		t.Helper()
		if r.code != 0 {
			t.Fatalf("%s: exit %d: %s", label, r.code, r.err)
		}
		rel = strings.TrimSpace(r.out)
		if !interventionPath.MatchString(rel) {
			t.Fatalf("%s: add printed %q", label, rel)
		}
		text = s.read(rel)
		got := strings.Join([]string{frontValue(t, text, "options"), frontValue(t, text, "recommended"),
			frontValue(t, text, "decided"), frontValue(t, text, "agreement")}, "|")
		if want := strings.Join([]string{options, recommended, decided, agreement}, "|"); got != want {
			t.Fatalf("%s: options|recommended|decided|agreement is %q, want %q", label, got, want)
		}
		return rel, text
	}

	firstRel, _ := made("first", three("--decided-option", "1", "--decided", "replaced as recommended"), "3", "1", "1", "recommended")
	firstID := strings.TrimSuffix(filepath.Base(firstRel), ".md")
	made("--decided-option 2", three("--decided-option", "2", "--decided", "reset instead"), "3", "1", "2", "other-option")
	_, adj := made("--adjusted", three("--decided-option", "1", "--adjusted", "--decided", "replaced, and more"), "3", "1", "1", "adjusted")
	if frontValue(t, adj, "adjusted") != "true" {
		t.Error("--adjusted: no adjusted: true in the frontmatter")
	}
	made("--decided-other", three("--decided-other", "--decided", "fixed by hand"), "3", "1", "other", "different")
	made("no options", add("--decided", "replaced as recommended"), "0", "0", `""`, "no-options")

	count := func() int {
		es, err := os.ReadDir(filepath.Join(s.dir, ".vloop", "interventions"))
		if err != nil {
			t.Fatal(err)
		}
		return len(es)
	}
	refuse := func(msg string, extra ...string) {
		t.Helper()
		before := count()
		r := add(extra...)
		if r.code != 2 || strings.TrimSpace(r.err) != "vloop: "+msg {
			t.Errorf("add %v gave exit %d and %q, want exit 2 and %q", extra, r.code, r.err, "vloop: "+msg)
		}
		if count() != before {
			t.Errorf("add %v: a refused add wrote a record", extra)
		}
	}
	refuse("at most three options", "--option", "a", "--option", "b", "--option", "c", "--option", "d", "--recommended", "1", "--why", "x", "--decided-option", "1")
	refuse("--recommended must name an option, 1 to 2", "--option", "a", "--option", "b", "--recommended", "3", "--why", "x", "--decided-option", "1")
	refuse("options need --decided-option or --decided-other", "--option", "a", "--recommended", "1", "--why", "x")
	refuse("--decided-option and --decided-other exclude each other", "--option", "a", "--recommended", "1", "--why", "x", "--decided-option", "1", "--decided-other")
	refuse("--recommended, --decided-option and --decided-other need options", "--decided-other")

	if r := s.run(nil, "intervention", "set", firstID, "decided", "2"); r.code != 0 {
		t.Fatalf("set decided 2: %s", r.err)
	}
	first := s.read(firstRel)
	if got := frontValue(t, first, "agreement"); got != "other-option" {
		t.Fatalf("after set decided 2 the agreement is %q, want other-option", got)
	}
	r := s.run(nil, "intervention", "set", firstID, "agreement", "recommended")
	if r.code != 2 || strings.TrimSpace(r.err) != "vloop: agreement is derived — set decided, adjusted or recommended instead" {
		t.Errorf("set agreement gave exit %d and %q", r.code, r.err)
	}
	if s.read(firstRel) != first {
		t.Error("a refused set changed the record")
	}

	s.write(firstRel, strings.Replace(first, "agreement: other-option", "agreement: different", 1))
	r = s.run(nil, "intervention", "list")
	if r.code != 1 || !strings.Contains(r.err, filepath.Base(firstRel)) {
		t.Errorf("list over a hand-edited agreement gave exit %d and %q, want exit 1 naming the file", r.code, r.err)
	}
	s.write(firstRel, first)

	v1ID := "I20260101-0700-a-v1-record-with-context-and-decided"
	v1Rel := ".vloop/interventions/" + v1ID + ".md"
	s.write(v1Rel, v1Record)
	r = s.run(nil, "intervention", "list", "--json")
	if r.code != 0 {
		t.Fatalf("list --json with a v1 record: %s", r.err)
	}
	var recs []struct {
		ID      string `json:"id"`
		Context string `json:"context"`
		Done    string `json:"done"`
		Decided struct {
			Text string `json:"text"`
		} `json:"decided"`
	}
	if err := json.Unmarshal([]byte(r.out), &recs); err != nil {
		t.Fatalf("list --json: %v", err)
	}
	found := false
	for _, rec := range recs {
		if rec.ID == v1ID {
			found = true
			if rec.Context != "the B1 draft, forks one and two" || rec.Decided.Text != "the operator took the narrower scope" || rec.Done != "settled both" {
				t.Errorf("the v1 record reads as context %q, decided %q, done %q", rec.Context, rec.Decided.Text, rec.Done)
			}
		}
	}
	if !found {
		t.Error("list --json lacks the v1 record")
	}
	bodyBefore := body(s.read(v1Rel))
	if r := s.run(nil, "intervention", "migrate"); r.code != 0 || strings.TrimSpace(r.out) != "migrated 1 record(s)" {
		t.Errorf("migrate gave exit %d and %q %q, want 'migrated 1 record(s)'", r.code, r.out, r.err)
	}
	if body(s.read(v1Rel)) != bodyBefore {
		t.Error("migrate changed the v1 record's body")
	}
	if r := s.run(nil, "intervention", "migrate"); r.code != 0 || strings.TrimSpace(r.out) != "nothing to migrate" {
		t.Errorf("a second migrate gave exit %d and %q, want 'nothing to migrate'", r.code, r.out)
	}

	r = s.run(nil, "intervention", "show", firstID)
	if r.code != 0 {
		t.Fatalf("show: %s", r.err)
	}
	links := map[string]string{}
	in := false
	for _, l := range strings.Split(r.out, "\n") {
		if strings.TrimSpace(l) == "" {
			in = false
		}
		if in {
			if f := strings.Fields(l); len(f) > 0 {
				links[f[0]] = strings.Join(f[1:], " ")
			}
		}
		if strings.TrimSuffix(strings.TrimSpace(l), ":") == "links" {
			in = true
		}
	}
	if !strings.Contains(links["D20260101-0900-x"], "(not found)") {
		t.Errorf("show's links for the defect: %q, want (not found)", links["D20260101-0900-x"])
	}
	if !strings.Contains(links[sha], "the worked example's head") {
		t.Errorf("show's links for HEAD: %q, want its subject", links[sha])
	}
	if !strings.Contains(links["B20260101-0900-a"], "(not found)") {
		t.Errorf("show's links for the run: %q, want (not found)", links["B20260101-0900-a"])
	}

	row := func(out, key string) string {
		for _, l := range strings.Split(out, "\n") {
			if f := strings.Fields(l); len(f) > 0 && f[0] == key {
				return strings.Join(f[1:], " ")
			}
		}
		return ""
	}
	r = s.run(nil, "metrics", "--interventions")
	if r.code != 0 {
		t.Fatalf("metrics --interventions: %s", r.err)
	}
	if got := row(r.out, "repair"); got != "5 0 0% 2 1 1 1 0 5" {
		t.Errorf("the repair row is %q, want %q", got, "5 0 0% 2 1 1 1 0 5")
	}
	if got := row(r.out, "total"); got != "6 0 0% 2 1 1 2 0 5" {
		t.Errorf("the total row is %q, want %q", got, "6 0 0% 2 1 1 2 0 5")
	}
	r = s.run(nil, "metrics", "--interventions", "--by", "task")
	if r.code != 2 || strings.TrimSpace(r.err) != "vloop: --by takes kind or phase with --interventions" {
		t.Errorf("--by task gave exit %d and %q", r.code, r.err)
	}

	r = s.run(nil, "metrics", "export")
	if r.code != 0 {
		t.Fatalf("metrics export: %s", r.err)
	}
	n := 0
	for _, l := range strings.Split(r.out, "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		var o map[string]any
		if err := json.Unmarshal([]byte(l), &o); err != nil {
			t.Fatalf("export line %q: %v", l, err)
		}
		if o["type"] != "intervention" {
			continue
		}
		n++
		if _, ok := o["agreement"].(string); !ok {
			t.Errorf("an exported intervention lacks a string agreement: %s", l)
		}
		if _, ok := o["options"].(float64); !ok {
			t.Errorf("an exported intervention lacks a numeric options: %s", l)
		}
	}
	if n != 6 {
		t.Errorf("%d exported interventions, want 6", n)
	}
	for _, txt := range []string{"replace the gate with vloop task verify", "reset T3 and retry", "abandon the brief",
		"the gate, not the work, is wrong", "T3 in run", "replaced as recommended", "fixed by hand", "forks one and two"} {
		if strings.Contains(r.out, txt) {
			t.Errorf("the export carries %q", txt)
		}
	}
}
