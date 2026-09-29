package brief

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func esRun(t *testing.T, root, lang, file string) *Result {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "es", file+".loop-brief.md"))
	if err != nil {
		t.Fatal(err)
	}
	return Check(root, parseFit("docs/briefs/B20260101-0900-a.loop-brief.md", string(data)), SetFor(lang))
}

func esRoot(t *testing.T) string {
	root, _, _ := scratch(t)
	return root
}

func TestSpanishFixtures(t *testing.T) {
	root := esRoot(t)
	for _, f := range []string{"pass", "single-task-count", "exit-sale-con"} {
		res := esRun(t, root, "es", f)
		if res.Failed() || len(res.Warnings()) != 0 {
			t.Errorf("%s: %v", f, res.Lines)
		}
	}
	cases := []struct {
		file   string
		marker Marker
		want   string
	}{
		{"no-worked-example", Problem, "no worked example section"},
		{"no-fence", Problem, "no fenced code block"},
		{"no-out-of-scope", Problem, "no out-of-scope section"},
		{"one-item", Problem, "out-of-scope section has 1 item(s)"},
		{"home-path", Problem, "contains an absolute home path"},
		{"no-constraints", Warning, "no constraints section"},
		{"no-task-count", Warning, "no expected task count"},
		{"no-exit-codes", Warning, "no exit/status codes"},
	}
	for _, c := range cases {
		res := esRun(t, root, "es", c.file)
		got := res.Problems()
		if c.marker == Warning {
			got = res.Warnings()
		}
		if !has(got, c.want) {
			t.Errorf("%s: want %q, got %v", c.file, c.want, res.Lines)
		}
	}
}

func TestCrossLanguage(t *testing.T) {
	root := esRoot(t)
	// English headings in an es repo.
	res := esRun(t, root, "es", "english-headings")
	if !has(res.Problems(), "no out-of-scope section") {
		t.Errorf("%v", res.Lines)
	}
	data, _ := os.ReadFile(filepath.Join("testdata", "en", "pass.loop-brief.md"))
	res = Check(root, parseFit("docs/briefs/B20260101-0900-a.loop-brief.md", string(data)), SetFor("es"))
	if !has(res.Problems(), "no worked example section") || !has(res.Problems(), "no out-of-scope section") ||
		!has(res.Warnings(), "no constraints section") {
		t.Errorf("English brief in es: %v", res.Lines)
	}
	// Spanish headings in an en repo.
	res = esRun(t, root, "en", "pass")
	if !has(res.Problems(), "no worked example section") || !has(res.Problems(), "no out-of-scope section") {
		t.Errorf("Spanish brief in en: %v", res.Lines)
	}
}

func TestSpanishMatchingIsAccentSensitive(t *testing.T) {
	re := heading(Spanish.WorkedExample)
	for _, l := range []string{"## Ejemplo trabajado", "## EJEMPLO TRABAJADO — x", "### ejemplo trabajado"} {
		if !re.MatchString(l) {
			t.Errorf("%q must match", l)
		}
	}
	if !heading("Qué es").MatchString("## QUÉ ES") || heading("Qué es").MatchString("## Que es") {
		t.Error("accents must matter, case must not")
	}
}

func TestMessagesStayEnglish(t *testing.T) {
	res := esRun(t, esRoot(t), "es", "no-task-count")
	for _, l := range res.Lines {
		if strings.ContainsAny(l.Text, "áéíóúñ") && !strings.Contains(l.Text, "—") {
			t.Errorf("translated message: %q", l.Text)
		}
	}
	if s := res.Summary(); !strings.Contains(s, "warning(s)") {
		t.Errorf("%q", s)
	}
}
