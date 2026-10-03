package classify

import "testing"

type want struct{ path, cat, layer, glob string }

func run(t *testing.T, c *Classifier, cases []want) {
	t.Helper()
	for _, w := range cases {
		got := c.Classify(w.path)
		if got != (Result{w.cat, w.layer, w.glob}) {
			t.Errorf("%s: got %+v, want %s %s %s", w.path, got, w.cat, w.layer, w.glob)
		}
	}
}

func TestPresetNames(t *testing.T) {
	got := Names()
	if len(got) != 9 || got[0] != "csharp" || got[8] != "typescript" {
		t.Errorf("Names() = %v", got)
	}
	for _, n := range got {
		p, _ := Lookup(n)
		if len(p.Docs) != 1 || p.Docs[0] != "**/*.md" || len(p.Code) == 0 || len(p.Excluded) == 0 {
			t.Errorf("preset %s incomplete: %+v", n, p)
		}
	}
}

func TestPresetFixtures(t *testing.T) {
	cases := map[string][]want{
		"go":         {{"a/b.go", Code, "go", "**/*.go"}, {"a/b_test.go", Test, "go", "**/*_test.go"}, {"b.py", Other, "", ""}},
		"typescript": {{"a.ts", Code, "typescript", "**/*.ts"}, {"a.spec.ts", Test, "typescript", "**/*.spec.ts"}, {"a/node_modules/x.ts", Excluded, "typescript", "**/node_modules/**"}},
		"javascript": {{"a.cjs", Code, "javascript", "**/*.cjs"}, {"__tests__/a.js", Test, "javascript", "**/__tests__/**"}, {"a.ts", Other, "", ""}},
		"react":      {{"a.scss", Code, "react", "**/*.scss"}, {"a.stories.tsx", Test, "react", "**/*.stories.*"}, {"dist/a.tsx", Excluded, "react", "**/dist/**"}},
		"csharp":     {{"a.cshtml", Code, "csharp", "**/*.cshtml"}, {"x.Tests/a.cs", Test, "csharp", "**/*.Tests/**"}, {"bin/a.cs", Excluded, "csharp", "**/bin/**"}},
		"python":     {{"a.py", Code, "python", "**/*.py"}, {"tests/a.py", Test, "python", "**/tests/**"}, {"uv.lock", Excluded, "python", "uv.lock"}},
		"java":       {{"a.java", Code, "java", "**/*.java"}, {"m/src/test/A.java", Test, "java", "**/src/test/**"}, {"target/A.java", Excluded, "java", "**/target/**"}},
		"kotlin":     {{"a.kts", Code, "kotlin", "**/*.kts"}, {"m/src/test/A.kt", Test, "kotlin", "**/src/test/**"}, {"a.java", Other, "", ""}},
		"rust":       {{"a.rs", Code, "rust", "**/*.rs"}, {"benches/a.rs", Test, "rust", "**/benches/**"}, {"Cargo.lock", Excluded, "rust", "Cargo.lock"}},
	}
	for name, cs := range cases {
		run(t, New(Preset{}, []string{name}), append(cs, want{"docs/a.md", Docs, name, "**/*.md"}))
	}
}

func TestLayerAlways(t *testing.T) {
	c := New(Preset{Code: []string{"**"}}, []string{"go"})
	run(t, c, []want{
		{".vloop/x.go", Excluded, LayerAlways, ".vloop/**"},
		{".loop/a/b", Excluded, LayerAlways, ".loop/**"},
		{"vloop/x.go", Code, LayerRepo, "**"},
	})
}

func TestLayerRepo(t *testing.T) {
	c := New(Preset{Excluded: []string{"gen/**"}, Test: []string{"e2e/**"}, Docs: []string{"e2e/R.txt"}, Code: []string{"tpl/**"}}, []string{"go"})
	run(t, c, []want{
		{"gen/a_test.go", Excluded, LayerRepo, "gen/**"},
		{"e2e/R.txt", Test, LayerRepo, "e2e/**"},
		{"tpl/en.md", Code, LayerRepo, "tpl/**"},
		{"x.go", Code, "go", "**/*.go"},
		{"x.txt", Other, "", ""},
	})
}

func TestLayerPresetOrder(t *testing.T) {
	c := New(Preset{}, []string{"go", "python"})
	run(t, c, []want{
		{"tests/x.go", Test, "python", "**/tests/**"},
		{"vendor/tests/x.py", Excluded, "go", "vendor/**"},
		{"a.go", Code, "go", "**/*.go"},
	})
}

func TestParseStackScope(t *testing.T) {
	for _, ok := range []string{"go", "csharp@services/api", "react@web"} {
		if _, _, err := ParseStack(ok); err != nil {
			t.Errorf("ParseStack(%q): %v", ok, err)
		}
	}
	n, s, _ := ParseStack("csharp@services/api")
	if n != "csharp" || s != "services/api" {
		t.Errorf("got %q %q", n, s)
	}
	for _, bad := range []string{"", "@web", "cobol@web", "go@", "go@/a", "go@a/", "go@./a", "go@a/../b", "go@a//b", "go@a\\b", "go@."} {
		if _, _, err := ParseStack(bad); err == nil {
			t.Errorf("ParseStack(%q) accepted", bad)
		}
	}
}

func TestClassifyScope(t *testing.T) {
	c := New(Preset{}, []string{"go@services/x", "csharp@services", "python"})
	for _, tc := range []struct{ path, cat, layer, glob string }{
		{"services/x/go.sum", Excluded, "go@services/x", "go.sum"},
		{"services/x/main.go", Code, "go@services/x", "**/*.go"},
		{"services/x/a.cs", Other, "", ""},
		{"services/y/a.cs", Code, "csharp@services", "**/*.cs"},
		{"web/tests/helpers.ts", Test, "python", "**/tests/**"},
		{"a/tests/t.py", Test, "python", "**/tests/**"},
		{"services", Other, "", ""},
		{".loop/x", Excluded, LayerAlways, ".loop/**"},
	} {
		r := c.Classify(tc.path)
		if r.Category != tc.cat || r.Layer != tc.layer || (tc.glob != "" && r.Glob != tc.glob) {
			t.Errorf("%s: got %+v", tc.path, r)
		}
	}
}

func TestClassifyScopeRepoGlobsStayRepoRelative(t *testing.T) {
	c := New(Preset{Test: []string{"services/x/**"}}, []string{"go@services/x"})
	if r := c.Classify("services/x/main.go"); r.Category != Test || r.Layer != LayerRepo {
		t.Errorf("got %+v", r)
	}
}

func TestPresetsCountNestTests(t *testing.T) {
	for _, ext := range []string{"ts", "js"} {
		name := map[string]string{"ts": "typescript", "js": "javascript"}[ext]
		c := New(Preset{}, []string{name})
		for _, p := range []string{"test/links.e2e-spec." + ext, "test/jest-e2e.json", "src/links.e2e-spec." + ext} {
			if got := c.Classify(p); got.Category != Test {
				t.Errorf("%s %s: got %+v, want test", name, p, got)
			}
		}
		if got := c.Classify("src/app." + ext); got.Category != Code {
			t.Errorf("%s src/app.%s: got %+v, want code", name, ext, got)
		}
	}
}
