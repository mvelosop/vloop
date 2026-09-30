package brief

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// scratch is a temporary repo holding docs/x.md and the passing fixture.
func scratch(t *testing.T) (root, path, text string) {
	t.Helper()
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs", "briefs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "x.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("testdata", "en", "pass.loop-brief.md"))
	if err != nil {
		t.Fatal(err)
	}
	return root, "docs/briefs/B20260101-0900-a.loop-brief.md", string(data)
}

// parseFit parses text after pointing the frontmatter name at path's file, as a
// real brief's must be.
func parseFit(path, text string) *Brief {
	name := strings.TrimSuffix(filepath.Base(path), ".md")
	return Parse(path, regexp.MustCompile(`(?m)^name: .*$`).ReplaceAllString(text, "name: "+name))
}

func run(root, path, text string) *Result {
	return Check(root, parseFit(path, text), SetFor("en"))
}

func has(lines []string, prefix string) bool {
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			return true
		}
	}
	return false
}

func TestPassingFixtureIsClean(t *testing.T) {
	root, path, text := scratch(t)
	res := run(root, path, text)
	if res.Failed() || len(res.Warnings()) != 0 || res.Summary() != "ok, 0 warning(s)" {
		t.Fatalf("%+v", res.Lines)
	}
	if n := res.count(Pass); n < 8 {
		t.Fatalf("want at least 8 passes, got %d", n)
	}
}

func TestRules(t *testing.T) {
	cut := func(s string) func(string) string {
		return func(text string) string { return strings.Replace(text, s, "", 1) }
	}
	add := func(s string) func(string) string {
		return func(text string) string { return text + "\n" + s + "\n" }
	}
	repl := func(a, b string) func(string) string {
		return func(text string) string { return strings.Replace(text, a, b, 1) }
	}
	cases := []struct {
		name   string
		mutate func(string) string
		marker Marker
		want   string
	}{
		{"no worked example", repl("## Worked example — the happy path", "## Other"), Problem, "no worked example section — nothing arbitrates a disagreement"},
		{"no fence", cut("```\n$ a\nok\n```"), Problem, "no fenced code block"},
		{"no out of scope", repl("## Out of scope", "## Elsewhere"), Problem, "no out-of-scope section"},
		{"one item", cut("- two\n"), Problem, "out-of-scope section has 1 item(s)"},
		{"unix home", add("Notes in /Users/someone/notes."), Problem, "contains an absolute home path"},
		{"linux home", add("Notes in /home/someone/notes."), Problem, "contains an absolute home path"},
		{"windows home", add(`Notes in C:\Users\someone\notes.`), Problem, "contains an absolute home path"},
		{"no constraints", repl("## Constraints", "## Limits"), Warning, "no constraints section"},
		{"no task count", cut("6 to 9 tasks."), Warning, "no expected task count"},
		{"no exit codes", cut("An unknown id exits 1 with a message."), Warning, "no exit/status codes"},
		{"symbols", add("Uses `a.b()`, `c.d()` and `e.f()`."), Warning, "names 3 internal symbols"},
		{"dead path", add("See `docs/missing.md`."), Warning, "path does not resolve: docs/missing.md"},
		{"tracker key", add("Tracked as PROJ-123."), Warning, "names issue(s) no session can open: PROJ-123"},
		{"tracker url", add("See https://linear.app/acme/issue/abc."), Warning, "links an issue tracker"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root, path, text := scratch(t)
			res := run(root, path, c.mutate(text))
			var got []string
			for _, l := range res.Lines {
				if l.Marker == c.marker {
					got = append(got, l.Text)
				}
			}
			if !has(got, c.want) {
				t.Fatalf("want %v starting %q, got %v", c.marker, c.want, res.Lines)
			}
			if c.marker == Problem && !res.Failed() {
				t.Fatal("must fail")
			}
			if c.marker == Warning && res.Failed() {
				t.Fatal("a warning must not fail")
			}
		})
	}
}

func TestPassesForEachRule(t *testing.T) {
	root, path, text := scratch(t)
	res := run(root, path, text)
	for _, want := range []string{
		"has a worked example", "has a fenced block", "out of scope: 2 item(s)", "has constraints",
		"states an expected task count", "pins exit or status codes", "no absolute paths",
		"referenced docs resolve", "no references a memoryless",
	} {
		found := false
		for _, l := range res.Lines {
			found = found || (l.Marker == Pass && strings.HasPrefix(l.Text, want))
		}
		if !found {
			t.Errorf("no pass line %q in %v", want, res.Lines)
		}
	}
}

func TestAlreadyRun(t *testing.T) {
	root, path, text := scratch(t)
	j := filepath.Join(root, ".vloop", "state", "journals")
	if err := os.MkdirAll(j, 0o755); err != nil {
		t.Fatal(err)
	}
	// the shell loop's journal directory is not vloop's
	old := filepath.Join(root, ".loop", "state", "journals")
	_ = os.MkdirAll(old, 0o755)
	_ = os.WriteFile(filepath.Join(old, "B20260101-0900-a.md"), nil, 0o644)
	if res := run(root, path, text); res.Failed() {
		t.Fatalf("a .loop journal must not count: %v", res.Lines)
	}
	_ = os.WriteFile(filepath.Join(j, "B20260101-0900-a.md"), nil, 0o644)
	res := run(root, path, text)
	if !has(res.Problems(), "already run — .vloop/state/journals/B20260101-0900-a.md exists") {
		t.Fatalf("%v", res.Lines)
	}
}

func TestPathResolution(t *testing.T) {
	root, path, text := scratch(t)
	_ = os.WriteFile(filepath.Join(root, "docs", "briefs", "beside.md"), nil, 0o644)
	_ = os.MkdirAll(filepath.Join(root, "deep", "er"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "deep", "er", "fixture.json"), nil, 0o644)
	text += "\nSee `beside.md` and `fixture.json` and `~/a.md` and `docs/<n>.md`.\n"
	if res := run(root, path, text); len(res.Warnings()) != 0 {
		t.Fatalf("%v", res.Lines)
	}
}

func TestBindingSectionPathsAreNotScannedHere(t *testing.T) {
	root, path, text := scratch(t)
	text = strings.Replace(text, "`docs/x.md`", "`docs/gone.md`", 1)
	if res := run(root, path, text); len(res.Warnings()) != 0 {
		t.Fatalf("%v", res.Lines)
	}
}

func TestSkipUnlessReady(t *testing.T) {
	root, path, text := scratch(t)
	for _, st := range []string{"draft", "consumed", "abandoned"} {
		res := run(root, path, strings.Replace(text, "status: ready", "status: "+st, 1))
		if !res.Skipped || res.Status != st {
			t.Errorf("%s: %+v", st, res)
		}
	}
}

func TestParseStatus(t *testing.T) {
	if b := Parse("p", "---\nstatus: ready   # draft | ready\n---\n"); b.Status != "ready" {
		t.Errorf("%q", b.Status)
	}
	if b := Parse("p", "status: ready\n"); b.Status != "" {
		t.Errorf("outside frontmatter: %q", b.Status)
	}
	if got := RunID("docs/briefs/B1-x.loop-brief.md"); got != "B1-x" {
		t.Errorf("%q", got)
	}
}
