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
	for _, want := range []string{"detected stacks: go\n", "wrote .vloop/config.toml\n", "wrote .vloop/install.json\n", "wrote .vloop/install.json\nwrote .vloop/repo-guidelines.md\n", "-primeros-pasos.loop-brief.md\n", "updated .gitignore\n", "wrote CLAUDE.md\n", "next:\n  vloop doctor\n  get /vloop:operate for your Claude Code sessions in this repository, once:\n    claude plugin marketplace add mvelosop/vloop\n    claude plugin install vloop@vloop --scope project\n  or, for one session: claude --plugin-dir \"$(vloop plugin path)\"\n  for Copilot CLI: trust this folder in copilot (the plugin is declared in .claude/settings.json), or once:\n    copilot plugin marketplace add mvelosop/vloop\n    copilot plugin install vloop@vloop\n  or, for one session: copilot --plugin-dir \"$(vloop plugin path)\"\n  then ask Claude Code for what you want (\"run the brief docs/briefs/<name>\"); /vloop:operate drives vloop\n"} {
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
	if !strings.HasPrefix(got, "# Mine\n<!-- vloop:begin -->\n") || !strings.HasSuffix(got, "<!-- vloop:end -->\nbelow  \nlast") || strings.Contains(got, "\nold\n") {
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
	for _, want := range []string{"would write .vloop/config.toml\n", "would write .vloop/install.json\n", "would write .vloop/repo-guidelines.md\n", "would write .gitignore\n", "would write CLAUDE.md\n"} {
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

func TestInitGuidelinesTemplate(t *testing.T) {
	for _, lang := range []string{"en", "es"} {
		d := initRepo(t, map[string]string{"go.mod": "module a\n"})
		if _, _, code := runPluginCLI(t, initBuild, d, "init", "--language", lang); code != 0 {
			t.Fatalf("%s: code %d", lang, code)
		}
		got := readFile(t, d, ".vloop/repo-guidelines.md")
		// K2 retired the two sections: the title and one sentence only.
		if n := len(strings.Split(strings.TrimSpace(got), "\n\n")); n != 2 || strings.Contains(got, "## ") || strings.Contains(got, "<!--") {
			t.Errorf("%s: not the title and one sentence:\n%s", lang, got)
		}
		if lang == "es" && !strings.Contains(got, "lecciones") {
			t.Errorf("es prose is not Spanish:\n%s", got)
		}
	}
}

func TestInitKeepsExistingGuidelines(t *testing.T) {
	const mine = "# Mine\n\n- keep me (D20261011-1455)\n"
	d := initRepo(t, map[string]string{"go.mod": "module a\n", ".vloop/repo-guidelines.md": mine})
	out, _, code := runPluginCLI(t, initBuild, d, "init")
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	if got := readFile(t, d, ".vloop/repo-guidelines.md"); got != mine {
		t.Errorf("file changed: %q", got)
	}
	if strings.Contains(out, "repo-guidelines.md") {
		t.Errorf("init printed a line for the kept file:\n%s", out)
	}
}

const (
	settingsFresh = "{\n  \"extraKnownMarketplaces\": {\n    \"vloop\": {\n      \"source\": {\n        \"source\": \"github\",\n        \"repo\": \"mvelosop/vloop\"\n      }\n    }\n  },\n  \"enabledPlugins\": {\n    \"vloop@vloop\": true\n  }\n}\n"
	settingsRef   = ".claude/settings.json"
)

func TestInitSettingsNoFileWritesTheDeclaration(t *testing.T) {
	d := initRepo(t, nil)
	out, errs, code := runPluginCLI(t, initBuild, d, "init", "--language", "en")
	if code != 0 || errs != "" || !strings.Contains(out, "wrote CLAUDE.md\nwrote .claude/settings.json\n") {
		t.Fatalf("code %d out %q err %q", code, out, errs)
	}
	if got := readFile(t, d, settingsRef); got != settingsFresh {
		t.Errorf("settings = %q", got)
	}
}

func TestInitSettingsDryRunWritesNothing(t *testing.T) {
	d := initRepo(t, nil)
	out, _, code := runPluginCLI(t, initBuild, d, "init", "--dry-run")
	if code != 0 || !strings.Contains(out, "would write .claude/settings.json\n") {
		t.Fatalf("code %d out %q", code, out)
	}
	if _, err := os.Stat(filepath.Join(d, ".claude")); err == nil {
		t.Error(".claude/ was created on --dry-run")
	}
}

func TestInitSettingsMergesAfterTheKeysItKeeps(t *testing.T) {
	d := initRepo(t, map[string]string{settingsRef: `{"permissions":{"allow":["Bash(a && b)"]},"enabledPlugins":{"other@x":true}}`})
	out, _, code := runPluginCLI(t, initBuild, d, "init")
	if code != 0 || !strings.Contains(out, "updated .claude/settings.json\n") {
		t.Fatalf("code %d out %q", code, out)
	}
	want := "{\n  \"permissions\": {\n    \"allow\": [\n      \"Bash(a && b)\"\n    ]\n  },\n  \"enabledPlugins\": {\n    \"other@x\": true,\n    \"vloop@vloop\": true\n  },\n  \"extraKnownMarketplaces\": {\n    \"vloop\": {\n      \"source\": {\n        \"source\": \"github\",\n        \"repo\": \"mvelosop/vloop\"\n      }\n    }\n  }\n}\n"
	if got := readFile(t, d, settingsRef); got != want {
		t.Errorf("settings = %q", got)
	}
}

func TestInitSettingsKeepsAFileWithBothEntries(t *testing.T) {
	in := `{"extraKnownMarketplaces":{"vloop":{"source":{"source":"directory","path":"../vloop"}}},"enabledPlugins":{"vloop@vloop":false}}`
	for _, args := range [][]string{{"init", "--dry-run"}, {"init"}} {
		d := initRepo(t, map[string]string{settingsRef: in})
		out, _, code := runPluginCLI(t, initBuild, d, args...)
		verb := map[bool]string{true: "would keep", false: "kept"}[len(args) > 1]
		if code != 0 || !strings.Contains(out, verb+" .claude/settings.json\n") {
			t.Errorf("%v: code %d out %q", args, code, out)
		}
		if got := readFile(t, d, settingsRef); got != in {
			t.Errorf("%v: file rewritten: %q", args, got)
		}
	}
}

func TestInitSettingsKeepsOptOutAndOwnMarketplace(t *testing.T) {
	d := initRepo(t, map[string]string{settingsRef: `{"extraKnownMarketplaces":{"vloop":{"source":{"source":"directory","path":"../vloop"}}},"enabledPlugins":{"vloop@vloop":false}}`})
	// only the missing half is added: drop the plugin entry
	os.WriteFile(filepath.Join(d, settingsRef), []byte(`{"enabledPlugins":{"vloop@vloop":false}}`), 0o644)
	if _, _, code := runPluginCLI(t, initBuild, d, "init"); code != 0 {
		t.Fatal(code)
	}
	got := readFile(t, d, settingsRef)
	if !strings.Contains(got, `"vloop@vloop": false`) || strings.Contains(got, "true") {
		t.Errorf("opt-out not kept: %s", got)
	}
}

func TestInitSettingsLayoutKeepsNumbersAndText(t *testing.T) {
	d := initRepo(t, map[string]string{settingsRef: `{"a":[],"b":{},"n":1.50,"big":12345678901234567890,"s":"a<b>&c"}`})
	if _, _, code := runPluginCLI(t, initBuild, d, "init"); code != 0 {
		t.Fatal(code)
	}
	got := readFile(t, d, settingsRef)
	for _, want := range []string{"\"a\": [],\n", "\"b\": {},\n", "\"n\": 1.50,\n", "\"big\": 12345678901234567890,\n", "\"s\": \"a<b>&c\",\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("settings lacks %q:\n%s", want, got)
		}
	}
	if !strings.HasSuffix(got, "}\n") {
		t.Errorf("no final newline: %q", got)
	}
}

func TestInitSettingsRefusesWhatIsNotAnObject(t *testing.T) {
	const notObj = "vloop: .claude/settings.json is not a JSON object — fix it, then re-run\n"
	cases := []struct{ in, want string }{
		{"[1]", notObj}, {"{", notObj}, {`"x"`, notObj},
		{`{"enabledPlugins":null}`, "vloop: .claude/settings.json's enabledPlugins is not a JSON object — fix it, then re-run\n"},
		{`{"extraKnownMarketplaces":[]}`, "vloop: .claude/settings.json's extraKnownMarketplaces is not a JSON object — fix it, then re-run\n"},
	}
	for _, c := range cases {
		for _, dry := range []bool{false, true} {
			d := initRepo(t, map[string]string{settingsRef: c.in})
			args := []string{"init"}
			if dry {
				args = append(args, "--dry-run")
			}
			out, errs, code := runPluginCLI(t, initBuild, d, args...)
			if code != 1 || errs != c.want || strings.Contains(out, "would ") || strings.Contains(out, "wrote ") {
				t.Errorf("%s dry=%v: code %d out %q err %q", c.in, dry, code, out, errs)
			}
			if _, err := os.Stat(filepath.Join(d, ".vloop")); err == nil {
				t.Errorf("%s dry=%v: files were written", c.in, dry)
			}
			if readFile(t, d, settingsRef) != c.in {
				t.Errorf("%s: file changed", c.in)
			}
		}
	}
}

func TestInitSettingsRefusesSymlink(t *testing.T) {
	d := initRepo(t, nil)
	target := filepath.Join(t.TempDir(), "t.json")
	os.WriteFile(target, []byte("{}"), 0o644)
	os.MkdirAll(filepath.Join(d, ".claude"), 0o755)
	if err := os.Symlink(target, filepath.Join(d, settingsRef)); err != nil {
		t.Skip(err)
	}
	_, errs, code := runPluginCLI(t, initBuild, d, "init")
	if code != 1 || !strings.Contains(errs, "symlink") {
		t.Errorf("code %d err %q", code, errs)
	}
	if b, _ := os.ReadFile(target); string(b) != "{}" {
		t.Error("wrote through the symlink")
	}
	if _, err := os.Stat(filepath.Join(d, ".vloop")); err == nil {
		t.Error("files were written")
	}
}
