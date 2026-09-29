package brief

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func refsRun(t *testing.T, lang, fixture string, mutate func(root, text string) string) *Result {
	t.Helper()
	root, path, _ := scratch(t)
	data, err := os.ReadFile(filepath.Join("testdata", "refs", fixture+".loop-brief.md"))
	if err != nil {
		t.Fatal(err)
	}
	return Check(root, parseFit(path, mutate(root, string(data))), SetFor(lang))
}

func keep(_, s string) string { return s }

func TestBindingRefsPass(t *testing.T) {
	for lang, f := range map[string]string{"en": "en", "es": "es"} {
		if res := refsRun(t, lang, f, keep); res.Failed() || len(res.Warnings()) != 0 {
			t.Errorf("%s: %v", lang, res.Lines)
		}
	}
	dash := func(_, s string) string { return strings.Replace(s, " — ", " - ", 1) }
	if res := refsRun(t, "en", "en", dash); res.Failed() {
		t.Errorf("hyphen separator: %v", res.Lines)
	}
}

func TestBindingRefsFire(t *testing.T) {
	entry := map[string]string{"en": "- `docs/x.md` — the error contract", "es": "- `docs/x.md` — el contrato de errores"}
	for lang := range entry {
		sub := func(with string) func(string, string) string {
			return func(_, s string) string { return strings.Replace(s, entry[lang], with, 1) }
		}
		cases := []struct {
			name, want string
			warn       bool
			mutate     func(string, string) string
		}{
			{"dead", "binding reference does not resolve: docs/gone.md", false, sub("- `docs/gone.md` — why")},
			{"no reason", "binding reference has no reason: docs/x.md", false, sub("- `docs/x.md`")},
			{"empty reason", "binding reference has no reason: docs/x.md", false, sub("- `docs/x.md` —")},
			{"no entry point", "binding reference directory has no entry point: docs/empty/", false, func(root, s string) string {
				_ = os.MkdirAll(filepath.Join(root, "docs", "empty"), 0o755)
				_ = os.WriteFile(filepath.Join(root, "docs", "empty", "a.md"), []byte("x"), 0o644)
				return sub("- `docs/empty/` — bundle")(root, s)
			}},
		}
		for _, c := range cases {
			res := refsRun(t, lang, lang, c.mutate)
			if !has(res.Problems(), c.want) {
				t.Errorf("%s/%s: want %q got %v", lang, c.name, c.want, res.Lines)
			}
			if c.name == "dead" && len(res.Warnings()) != 0 {
				t.Errorf("%s: a dead reference is not also a warning: %v", lang, res.Lines)
			}
		}
		none := func(_, s string) string {
			s = strings.Replace(s, entry[lang], "", 1)
			return strings.Replace(s, "## Binding references", "## Other", 1)
		}
		if lang == "es" {
			none = func(_, s string) string {
				s = strings.Replace(s, entry[lang], "", 1)
				return strings.Replace(s, "## Referencias vinculantes", "## Otro", 1)
			}
		}
		res := refsRun(t, lang, lang, none)
		if !has(res.Warnings(), "no binding references — nothing binds a task to a document") || res.Failed() {
			t.Errorf("%s: missing section: %v", lang, res.Lines)
		}
	}
}

func TestBindingRefsEntryPoints(t *testing.T) {
	for _, n := range []string{"README.md", "index.md", "README-bundle.md"} {
		res := refsRun(t, "en", "en", func(root, s string) string {
			_ = os.MkdirAll(filepath.Join(root, "docs", "d"), 0o755)
			_ = os.WriteFile(filepath.Join(root, "docs", "d", n), []byte("x"), 0o644)
			return strings.Replace(s, "- `docs/x.md` — the error contract", "- `docs/d/` — bundle", 1)
		})
		if res.Failed() {
			t.Errorf("%s: %v", n, res.Lines)
		}
	}
}
