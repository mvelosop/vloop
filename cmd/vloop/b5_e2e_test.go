package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	b5Mu   sync.Mutex
	b5Bins = map[string]string{}
)

// b5Binary builds vloop stamped at the given version, once per version.
func b5Binary(t *testing.T, version string) string {
	t.Helper()
	b5Mu.Lock()
	defer b5Mu.Unlock()
	if p, ok := b5Bins[version]; ok {
		return p
	}
	p := filepath.Join(filepath.Dir(binPath), "vloop-"+version)
	ld := "-X main.version=" + version + " -X main.commit=b5test"
	if out, err := exec.Command("go", "build", "-ldflags", ld, "-o", p, ".").CombinedOutput(); err != nil {
		t.Fatalf("build vloop %s: %v\n%s", version, err, out)
	}
	b5Bins[version] = p
	return p
}

// b5Repo is a fixture repository with a fake home and a stub claude first on PATH.
type b5Repo struct {
	t      *testing.T
	dir    string
	home   string
	stub   string
	plugin string // the version the stub's `plugin list --json` reports
}

func b5Git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(cleanEnv(t.TempDir()), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// b5New makes a git repository (with an identity unless noIdentity) under a
// fake home that marks it trusted, and a stub claude reporting vloop@vloop 0.1.0.
func b5New(t *testing.T, noGit, noIdentity bool) *b5Repo {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stub claude is a shell script")
	}
	r := &b5Repo{t: t, dir: t.TempDir(), home: t.TempDir(), stub: t.TempDir(), plugin: "0.1.0"}
	if !noGit {
		b5Git(t, r.dir, "init", "-q", "-b", "main")
		if !noIdentity {
			b5Git(t, r.dir, "config", "user.name", "gate")
			b5Git(t, r.dir, "config", "user.email", "gate@example.com")
		}
	}
	real, err := filepath.EvalSymlinks(r.dir)
	if err != nil {
		t.Fatal(err)
	}
	trust := map[string]any{}
	for _, k := range []string{r.dir, real} {
		trust[k] = map[string]any{"hasTrustDialogAccepted": true}
	}
	raw, _ := json.Marshal(map[string]any{"projects": trust})
	if err := os.WriteFile(filepath.Join(r.home, ".claude.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	r.setPlugin("0.1.0")
	return r
}

func (r *b5Repo) setPlugin(version string) {
	r.t.Helper()
	script := "#!/bin/sh\ncase \"$1\" in\n" +
		"  --version) echo '2.0.0 (Claude Code)';;\n" +
		"  plugin) echo '[{\"id\":\"vloop@vloop\",\"version\":\"" + version + "\",\"enabled\":true}]';;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(r.stub, "claude"), []byte(script), 0o755); err != nil {
		r.t.Fatal(err)
	}
}

func (r *b5Repo) write(rel, content string) {
	r.t.Helper()
	p := filepath.Join(r.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *b5Repo) read(rel string) string {
	r.t.Helper()
	b, err := os.ReadFile(filepath.Join(r.dir, filepath.FromSlash(rel)))
	if err != nil {
		r.t.Fatal(err)
	}
	return string(b)
}

func (r *b5Repo) exists(rel string) bool {
	_, err := os.Stat(filepath.Join(r.dir, filepath.FromSlash(rel)))
	return err == nil
}

// run plays one command with the binary built at version, in the fixture.
func (r *b5Repo) run(version string, args ...string) result {
	r.t.Helper()
	cmd := exec.Command(b5Binary(r.t, version), args...)
	cmd.Dir = r.dir
	cmd.Env = append(cleanEnv(r.home), "PATH="+r.stub+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	code := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			r.t.Fatalf("run %v: %v", args, err)
		}
		code = ee.ExitCode()
	}
	return result{o.String(), e.String(), code}
}

var b5Spaces = regexp.MustCompile(`[ \t]+`)

// b5Lines splits output into lines with runs of spaces collapsed.
func b5Lines(s string) []string {
	var out []string
	for _, l := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		out = append(out, strings.TrimSpace(b5Spaces.ReplaceAllString(l, " ")))
	}
	return out
}

// b5Expect checks the exit code, stderr and stdout lines (spaces collapsed).
func b5Expect(t *testing.T, r result, code int, out []string, errOut string) {
	t.Helper()
	got := b5Lines(r.out)
	if r.out == "" {
		got = nil
	}
	if r.code != code || r.err != errOut || strings.Join(got, "\n") != strings.Join(out, "\n") {
		t.Fatalf("got code=%d out=%q err=%q\nwant code=%d out=%q err=%q", r.code, got, r.err, code, out, errOut)
	}
	if strings.Contains(r.out+r.err, "\x1b") {
		t.Fatalf("colour escapes in output: %q %q", r.out, r.err)
	}
}

// b5Has checks that a stdout line (spaces collapsed) is present.
func b5Has(t *testing.T, r result, line string) {
	t.Helper()
	for _, l := range b5Lines(r.out) {
		if l == line {
			return
		}
	}
	t.Fatalf("no stdout line %q in %q", line, r.out)
}

func b5Code(t *testing.T, r result, code int) {
	t.Helper()
	if r.code != code {
		t.Fatalf("exit %d, want %d; out=%q err=%q", r.code, code, r.out, r.err)
	}
}

func TestWorkedExampleB5Session(t *testing.T) {
	r := b5New(t, false, false)
	r.write("go.mod", "module example.com/shop\n\ngo 1.22\n")
	r.write("package.json", `{"dependencies":{"react":"^18.0.0"}}`+"\n")
	r.write(".gitignore", "node_modules/\n")

	init := r.run("0.1.0", "init", "--language", "es")
	b5Code(t, init, 0)
	if init.err != "" {
		t.Fatalf("init stderr: %q", init.err)
	}
	lines := b5Lines(init.out)
	if len(lines) != 10 {
		t.Fatalf("init printed %d lines: %q", len(lines), init.out)
	}
	brief := regexp.MustCompile(`^wrote docs/briefs/B\d{8}-\d{4}-primeros-pasos\.loop-brief\.md$`)
	if !brief.MatchString(lines[3]) {
		t.Fatalf("starter brief line: %q", lines[3])
	}
	want := []string{"detected stacks: go, javascript, react", "wrote .vloop/config.toml", "wrote .vloop/install.json", lines[3],
		"updated .gitignore", "wrote CLAUDE.md", "next:", "claude plugin marketplace add mvelosop/vloop",
		"claude plugin install vloop@vloop", "vloop doctor"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("init printed %q, want %q", lines, want)
	}

	b5Expect(t, r.run("0.1.0", "config", "get", "metrics.stacks"), 0, []string{"go,javascript,react"}, "")
	b5Expect(t, r.run("0.1.0", "init"), 1, nil, "vloop: already initialized by vloop 0.1.0 — run vloop upgrade\n")

	b5Expect(t, r.run("0.1.0", "doctor"), 0, []string{"✓ git", "✓ install", "✓ config", "✓ claude", "✓ trust", "✓ gate shell",
		"- plan", "- branch", "✓ plugin", "- self-hosting", "doctor: 0 problem(s), 0 warning(s)"}, "")

	b5Expect(t, r.run("0.1.1", "upgrade"), 0, []string{"updated CLAUDE.md", "updated .vloop/install.json"}, "")
	b5Expect(t, r.run("0.2.0", "upgrade"), 1, []string{"would update CLAUDE.md"},
		"vloop: 0.1.1 → 0.2.0 may break this repository's setup — review the changes above, then run vloop upgrade --yes\n")
	b5Code(t, r.run("0.2.0", "upgrade", "--yes"), 0)
	b5Expect(t, r.run("0.1.1", "upgrade"), 1, nil,
		"vloop: this repository was set up by vloop 0.2.0, newer than this binary (0.1.1) — upgrade the binary, not the repository\n")

	b5Expect(t, r.run("0.2.0", "plugin", "path"), 0, []string{".vloop/tmp/plugin/0.2.0"}, "")
	b5Expect(t, r.run("0.2.0", "version", "--check-plugin", ".vloop/tmp/plugin/0.2.0"), 0, nil, "")
	b5Expect(t, r.run("0.1.1", "version", "--check-plugin", ".vloop/tmp/plugin/0.2.0"), 0,
		[]string{"vloop: the vloop plugin is 0.2.0 but the vloop binary is 0.1.1 — update the one that is behind"}, "")
}

func TestWorkedExampleB5Monorepo(t *testing.T) {
	r := b5New(t, false, false)
	r.write("go.mod", "module example.com/shop\n")
	r.write("services/api/Api.sln", "")
	r.write("services/api/Api/Api.csproj", "")
	r.write("services/api/Api.Tests/Api.Tests.csproj", "")
	r.write("web/package.json", `{"dependencies":{"react":"^18.0.0"}}`+"\n")
	r.write("tools/scripts/requirements.txt", "")

	init := r.run("0.2.0", "init")
	b5Code(t, init, 0)
	b5Has(t, init, "detected stacks: go, csharp@services/api, javascript@web, python@tools/scripts, react@web")

	b5Expect(t, r.run("0.2.0", "metrics", "classify", "services/api/Api.Tests/CalcTests.cs", "services/api/obj/Api.cs",
		"web/src/App.test.tsx", "web/tests/helpers.ts", "tools/bin/run.sh", "go.sum"), 0, []string{
		"services/api/Api.Tests/CalcTests.cs test (csharp@services/api: **/*.Tests/**)",
		"services/api/obj/Api.cs excluded (csharp@services/api: **/obj/**)",
		"web/src/App.test.tsx test (react@web: **/*.test.tsx)",
		"web/tests/helpers.ts other (none)",
		"tools/bin/run.sh other (none)",
		"go.sum excluded (go: go.sum)"}, "")

	b5Expect(t, r.run("0.2.0", "config", "set", "metrics.stacks", "go,csharp@services/missing"), 2, nil,
		"vloop: invalid value \"csharp@services/missing\" for metrics.stacks: want <stack> or <stack>@<existing directory>\n")
}

func TestWorkedExampleB5PlantedFailures(t *testing.T) {
	t.Run("no .git", func(t *testing.T) {
		r := b5New(t, true, false)
		b5Expect(t, r.run("0.2.0", "init"), 1, nil, "vloop: not a git repository — run git init first\n")
		if r.exists(".vloop") || r.exists("CLAUDE.md") {
			t.Fatal("init wrote into a directory without .git")
		}
	})

	t.Run("operator text around the markers", func(t *testing.T) {
		r := b5New(t, false, false)
		r.write("go.mod", "module example.com/shop\n")
		b5Code(t, r.run("0.1.0", "init"), 0)
		md := r.read("CLAUDE.md")
		const begin, end = "<!-- vloop:begin -->", "<!-- vloop:end -->"
		i, j := strings.Index(md, begin), strings.Index(md, end)
		if i < 0 || j < i {
			t.Fatalf("CLAUDE.md has no markers: %q", md)
		}
		above, below := "# Shop\n\nThe operator's own notes.\n\n", "\n\n## Mine\n\nKeep this, byte for byte.\n  trailing  \n"
		r.write("CLAUDE.md", above+md[i:j+len(end)]+below)
		b5Code(t, r.run("0.2.0", "upgrade", "--yes"), 0)
		got := r.read("CLAUDE.md")
		if !strings.HasPrefix(got, above+begin) || !strings.HasSuffix(got, end+below) {
			t.Fatalf("text outside the markers changed: %q", got)
		}
		if !strings.Contains(got, "Written by vloop 0.2.0.") {
			t.Fatalf("the marked section was not refreshed: %q", got)
		}
	})

	t.Run("existing config kept", func(t *testing.T) {
		r := b5New(t, false, false)
		r.write("go.mod", "module example.com/shop\n")
		r.write(".vloop/config.toml", "language = \"en\"\n")
		init := r.run("0.2.0", "init")
		b5Code(t, init, 0)
		b5Has(t, init, "kept .vloop/config.toml")
		b5Has(t, init, "wrote .vloop/install.json")
		if got := r.read(".vloop/config.toml"); got != "language = \"en\"\n" {
			t.Fatalf("config changed: %q", got)
		}
	})

	t.Run("docs/briefs already holds a brief", func(t *testing.T) {
		r := b5New(t, false, false)
		const old = "docs/briefs/B20200101-0000-old.loop-brief.md"
		r.write(old, "---\nname: B20200101-0000-old.loop-brief\n---\n")
		init := r.run("0.2.0", "init")
		b5Code(t, init, 0)
		b5Has(t, init, "kept docs/briefs/")
		ents, err := os.ReadDir(filepath.Join(r.dir, "docs", "briefs"))
		if err != nil || len(ents) != 1 {
			t.Fatalf("docs/briefs holds %v (%v), want only the existing brief", ents, err)
		}
	})

	t.Run("plugin version warning", func(t *testing.T) {
		r := b5New(t, false, false)
		b5Code(t, r.run("0.2.0", "init"), 0)
		doc := r.run("0.2.0", "doctor")
		b5Code(t, doc, 0)
		b5Has(t, doc, "! plugin the vloop plugin is 0.1.0 but this binary is 0.2.0 — update the one that is behind")
		b5Has(t, doc, "doctor: 0 problem(s), 1 warning(s)")
	})

	t.Run("no user.email", func(t *testing.T) {
		r := b5New(t, false, true)
		b5Code(t, r.run("0.2.0", "init"), 0)
		doc := r.run("0.2.0", "doctor")
		b5Code(t, doc, 1)
		b5Has(t, doc, "✗ git user.name and user.email not set")
		var summary string
		for _, l := range b5Lines(doc.out) {
			if strings.HasPrefix(l, "doctor: 1 problem(s)") {
				summary = l
			}
		}
		if summary == "" {
			t.Fatalf("no 1 problem(s) summary in %q", doc.out)
		}
	})

	t.Run("plugin.json is not JSON", func(t *testing.T) {
		r := b5New(t, false, false)
		r.write("bad/.claude-plugin/plugin.json", "this is not json\n")
		dir := filepath.Join(r.dir, "bad")
		b5Expect(t, r.run("0.2.0", "version", "--check-plugin", dir), 0,
			[]string{"vloop: cannot read the vloop plugin's version in " + dir}, "")
	})

	t.Run("stray file removed by plugin path", func(t *testing.T) {
		r := b5New(t, false, false)
		b5Expect(t, r.run("0.2.0", "plugin", "path"), 0, []string{".vloop/tmp/plugin/0.2.0"}, "")
		r.write(".vloop/tmp/plugin/0.2.0/stray.txt", "not part of the plugin\n")
		b5Expect(t, r.run("0.2.0", "plugin", "path"), 0, []string{".vloop/tmp/plugin/0.2.0"}, "")
		if r.exists(".vloop/tmp/plugin/0.2.0/stray.txt") {
			t.Fatal("the stray file survived plugin path")
		}
	})
}
