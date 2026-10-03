package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func initRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for p, c := range files {
		abs := filepath.Join(d, filepath.FromSlash(p))
		os.MkdirAll(filepath.Dir(abs), 0o755)
		if err := os.WriteFile(abs, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func readFile(t *testing.T, d, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(d, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

var initBuild = Build{Version: "0.1.0", Commit: "c010"}

func TestInitWritesEverything(t *testing.T) {
	d := initRepo(t, map[string]string{"go.mod": "module a\n", ".gitignore": "x/\n"})
	out, errs, code := runPluginCLI(t, initBuild, d, "init", "--language", "es")
	if code != 0 || errs != "" {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
	for _, want := range []string{"detected stacks: go\n", "wrote .vloop/config.toml\n", "wrote .vloop/install.json\n", "-primeros-pasos.loop-brief.md\n", "updated .gitignore\n", "wrote CLAUDE.md\n", "next:\n  claude plugin marketplace add mvelosop/vloop\n  claude plugin install vloop@vloop\n  vloop doctor\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if got := readFile(t, d, ".gitignore"); got != "x/\n.vloop/tmp/\n" {
		t.Errorf(".gitignore = %q", got)
	}
	if got := readFile(t, d, ".vloop/config.toml"); !strings.Contains(got, `language = "es"`) || !strings.Contains(got, "go") {
		t.Errorf("config = %q", got)
	}
	if got := readFile(t, d, ".vloop/install.json"); !strings.Contains(got, `"version":"0.1.0"`) || !strings.Contains(got, `"upgraded":null`) {
		t.Errorf("stamp = %q", got)
	}
	cm := readFile(t, d, "CLAUDE.md")
	if !strings.HasPrefix(cm, "<!-- vloop:begin -->\n") || !strings.HasSuffix(cm, "<!-- vloop:end -->\n") || !strings.Contains(cm, "vloop 0.1.0") {
		t.Errorf("CLAUDE.md = %q", cm)
	}
}

func TestInitRefusals(t *testing.T) {
	d := t.TempDir()
	_, errs, code := runPluginCLI(t, initBuild, d, "init")
	if code != 1 || errs != "vloop: not a git repository — run git init first\n" {
		t.Errorf("no git: %d %q", code, errs)
	}
	d = initRepo(t, map[string]string{".vloop/install.json": `{"schema":"install/v1","version":"0.0.9","commit":"x","initialized":"2026-01-01T00:00:00Z","upgraded":null}`})
	for _, args := range [][]string{{"init"}, {"init", "--dry-run"}} {
		out, errs, code := runPluginCLI(t, initBuild, d, args...)
		if code != 1 || out != "" || errs != "vloop: already initialized by vloop 0.0.9 — run vloop upgrade\n" {
			t.Errorf("%v: %d %q %q", args, code, out, errs)
		}
	}
	d = initRepo(t, nil)
	for _, args := range [][]string{{"init", "--language", "fr"}, {"init", "--stacks", "cobol"}, {"init", "--stacks", "go@missing"}} {
		out, errs, code := runPluginCLI(t, initBuild, d, args...)
		if code != 2 || out != "" || !strings.Contains(errs, "invalid value") {
			t.Errorf("%v: %d %q %q", args, code, out, errs)
		}
	}
	if _, err := os.Stat(filepath.Join(d, ".vloop")); err == nil {
		t.Error("a refusal wrote .vloop")
	}
}

func TestInitKeepsConfigAndBriefs(t *testing.T) {
	d := initRepo(t, map[string]string{
		"go.mod":             "module a\n",
		".vloop/config.toml": "# mine\nlanguage = \"en\"\n",
		"docs/briefs/B20250101-0900-m.loop-brief.md": "x\n",
		".gitignore": ".vloop/tmp/\n",
	})
	out, _, code := runPluginCLI(t, initBuild, d, "init")
	if code != 0 {
		t.Fatal(code, out)
	}
	for _, want := range []string{"kept .vloop/config.toml\n", "kept docs/briefs/\n", "kept .gitignore\n", "suggest metrics.stacks: go\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("lacks %q:\n%s", want, out)
		}
	}
	if readFile(t, d, ".vloop/config.toml") != "# mine\nlanguage = \"en\"\n" {
		t.Error("config changed")
	}
}

func TestInitClaudeMarkers(t *testing.T) {
	d := initRepo(t, map[string]string{"CLAUDE.md": "# Mine\n<!-- vloop:begin -->\nold\n<!-- vloop:end -->\nbelow  \nlast"})
	out, _, _ := runPluginCLI(t, initBuild, d, "init")
	if !strings.Contains(out, "updated CLAUDE.md\n") {
		t.Error(out)
	}
	got := readFile(t, d, "CLAUDE.md")
	if !strings.HasPrefix(got, "# Mine\n<!-- vloop:begin -->\n") || !strings.HasSuffix(got, "<!-- vloop:end -->\nbelow  \nlast") || strings.Contains(got, "old") {
		t.Errorf("got %q", got)
	}
	d = initRepo(t, map[string]string{"CLAUDE.md": "# Rules\nmine", ".gitignore": "keep"})
	runPluginCLI(t, initBuild, d, "init")
	got = readFile(t, d, "CLAUDE.md")
	if !strings.HasPrefix(got, "# Rules\nmine\n\n<!-- vloop:begin -->\n") || !strings.HasSuffix(got, "<!-- vloop:end -->\n") {
		t.Errorf("appended: %q", got)
	}
	if readFile(t, d, ".gitignore") != "keep\n.vloop/tmp/\n" {
		t.Error("gitignore")
	}
}

func TestInitDryRun(t *testing.T) {
	d := initRepo(t, map[string]string{"go.mod": "module a\n"})
	out, _, code := runPluginCLI(t, initBuild, d, "init", "--dry-run")
	if code != 0 {
		t.Fatal(code)
	}
	for _, want := range []string{"would write .vloop/config.toml\n", "would write .vloop/install.json\n", "would write .gitignore\n", "would write CLAUDE.md\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("lacks %q:\n%s", want, out)
		}
	}
	ents, _ := os.ReadDir(d)
	if len(ents) != 2 {
		t.Errorf("dry run wrote: %v", ents)
	}
}

func TestInitRefusesSymlink(t *testing.T) {
	d := initRepo(t, nil)
	outside := filepath.Join(t.TempDir(), "CLAUDE.md")
	os.WriteFile(outside, []byte("outside\n"), 0o644)
	if err := os.Symlink(outside, filepath.Join(d, "CLAUDE.md")); err != nil {
		t.Skip(err)
	}
	_, _, code := runPluginCLI(t, initBuild, d, "init")
	if code != 1 || readFile(t, filepath.Dir(outside), "CLAUDE.md") != "outside\n" {
		t.Errorf("code %d or outside changed", code)
	}
	if _, err := os.Stat(filepath.Join(d, ".vloop")); err == nil {
		t.Error("wrote before refusing")
	}
}

func TestInitWritesStarterChecks(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string // entries of "check.<name>=<run> (paths: <globs>)"
	}{
		{"go", map[string]string{"go.mod": "module a\n"},
			[]string{"check.go=go test ./... && go vet ./... (paths: **)"}},
		{"scoped go", map[string]string{"svc/go.mod": "module a\n"},
			[]string{"check.go-svc=cd svc && go test ./... && go vet ./... (paths: svc/**)"}},
		{"python", map[string]string{"pyproject.toml": "x\n"},
			[]string{"check.python=pytest (paths: **)"}},
		{"node stacks share one npm check", map[string]string{"package.json": `{"dependencies":{"react":"1"}}`, "tsconfig.json": "{}"},
			[]string{"check.javascript=npm test (paths: **)"}},
		{"rust", map[string]string{"Cargo.toml": "x\n"},
			[]string{"check.rust=cargo test (paths: **)"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := initRepo(t, c.files)
			out, errs, code := runPluginCLI(t, initBuild, d, "init")
			if code != 0 || errs != "" {
				t.Fatalf("code %d, stderr %q", code, errs)
			}
			if !strings.Contains(out, "starting point") {
				t.Errorf("output does not call the checks a starting point:\n%s", out)
			}
			list, _, code := runPluginCLI(t, initBuild, d, "config", "list")
			if code != 0 {
				t.Fatalf("config list exit %d", code)
			}
			var got []string
			for _, l := range strings.Split(list, "\n") {
				if strings.HasPrefix(l, "check.") {
					got = append(got, l)
				}
			}
			if strings.Join(got, "\n") != strings.Join(c.want, "\n") {
				t.Errorf("checks = %q, want %q", got, c.want)
			}
		})
	}
}

func TestInitWithNoStackWritesNoCheck(t *testing.T) {
	d := initRepo(t, map[string]string{"README.md": "x\n"})
	out, _, code := runPluginCLI(t, initBuild, d, "init")
	if code != 0 || strings.Contains(out, "starting point") || strings.Contains(readFile(t, d, ".vloop/config.toml"), "[[check]]") {
		t.Errorf("code %d:\n%s", code, out)
	}
}
