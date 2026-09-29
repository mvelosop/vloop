package brief

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEntry(t *testing.T, root, name, status, deps string) {
	t.Helper()
	dir := filepath.Join(root, "docs", "briefs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	text := "---\nname: " + strings.TrimSuffix(name, ".md") + "\nstatus: " + status + "\ndepends-on: " + deps + "\n---\n# x\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func names(es []Entry) string {
	var n []string
	for _, e := range es {
		n = append(n, e.Name+":"+e.Ready)
	}
	return strings.Join(n, " ")
}

func TestDependsOnForms(t *testing.T) {
	for text, want := range map[string]string{
		"---\ndepends-on: []\n---\n":                           "",
		"---\ndepends-on: [a, b]\n---\n":                       "a,b",
		"---\ndepends-on: a\n---\n":                            "a",
		"---\ndepends-on:\n  - a\n  - \"b\"\nstatus: x\n---\n": "a,b",
	} {
		if got := strings.Join(parseFrontmatter(text).DependsOn, ","); got != want {
			t.Errorf("%q: got %q want %q", text, got, want)
		}
	}
}

func TestOrderAndReady(t *testing.T) {
	root := t.TempDir()
	writeEntry(t, root, "B1-d.loop-brief.md", "draft", "[]")
	writeEntry(t, root, "B2-b.loop-brief.md", "consumed", "[]")
	writeEntry(t, root, "B3-a.loop-brief.md", "ready", "[B2-b.loop-brief]")
	writeEntry(t, root, "B0-c.loop-brief.md", "draft", "[B3-a.loop-brief]")
	writeEntry(t, root, "notes.md", "ready", "[missing]")
	writeEntry(t, root, "B9-e.architect-brief.md", "ready", "[missing]")
	es, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Order(es)
	if err != nil {
		t.Fatal(err)
	}
	want := "B1-d.loop-brief:ready B2-b.loop-brief:- B3-a.loop-brief:ready B0-c.loop-brief:blocked"
	if names(got) != want {
		t.Fatalf("got %s want %s", names(got), want)
	}
}

func TestDanglingAndCycle(t *testing.T) {
	root := t.TempDir()
	writeEntry(t, root, "B1-a.loop-brief.md", "ready", "[B2-b.loop-brief]")
	writeEntry(t, root, "B2-b.loop-brief.md", "ready", "[B1-a.loop-brief]")
	es, _ := Load(root)
	if _, err := Order(es); err == nil || !strings.HasPrefix(err.Error(), "depends-on cycle: B") {
		t.Fatalf("cycle: %v", err)
	}
	p := graphProblems(es, "B1-a.loop-brief", []string{"B2-b.loop-brief"}, false)
	if len(p) != 1 || p[0] != "depends-on cycle: B1-a.loop-brief -> B2-b.loop-brief -> B1-a.loop-brief" {
		t.Fatalf("%v", p)
	}
	writeEntry(t, root, "B0-c.loop-brief.md", "ready", "[B1-a.loop-brief]")
	es, _ = Load(root)
	if _, err := Order(es); err == nil || err.Error() != "depends-on cycle: B1-a.loop-brief -> B2-b.loop-brief -> B1-a.loop-brief" {
		t.Fatalf("cycle not via the entry: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "docs", "briefs", "B0-c.loop-brief.md")); err != nil {
		t.Fatal(err)
	}
	writeEntry(t, root, "B1-a.loop-brief.md", "ready", "[nope.loop-brief]")
	es, _ = Load(root)
	if _, err := Order(es); err == nil || err.Error() != "depends-on does not resolve: nope.loop-brief" {
		t.Fatalf("dangling: %v", err)
	}
}

func TestCheckReportsDependencyProblems(t *testing.T) {
	root, path, text := scratch(t)
	text = strings.Replace(text, "depends-on: []", "depends-on: [ghost.loop-brief]", 1)
	res := Check(root, Parse(path, text), SetFor("en"))
	if !has(res.Problems(), "depends-on does not resolve: ghost.loop-brief") {
		t.Fatalf("%v", res.Lines)
	}
}
