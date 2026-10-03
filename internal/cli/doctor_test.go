package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type doctorEnv struct {
	t       *testing.T
	repo    string
	home    string
	stub    string
	build   Build
	plugins string
}

// newDoctorEnv is a committed git repository with an identity, a stub claude
// first on PATH, and a fake home that trusts the repository.
func newDoctorEnv(t *testing.T) *doctorEnv {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stub claude is a shell script")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	e := &doctorEnv{t: t, build: Build{Version: "0.1.0", Commit: "c010"}}
	e.repo, _ = filepath.EvalSymlinks(t.TempDir())
	e.home = t.TempDir()
	e.stub = t.TempDir()
	e.plugins = `[{"id":"vloop@vloop","version":"0.1.0","scope":"user","enabled":true}]`
	t.Setenv("HOME", e.home)
	t.Setenv("USERPROFILE", e.home)
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "VLOOP_SHELL", "VLOOP_LANGUAGE", "VLOOP_METRICS_STACKS"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
	e.git("init", "-q", "-b", "main")
	e.git("config", "user.name", "t")
	e.git("config", "user.email", "t@example.com")
	e.git("commit", "-q", "--allow-empty", "-m", "base")
	e.claude(true)
	e.trust(true)
	e.run("init") // stamp at 0.1.0
	return e
}

func (e *doctorEnv) git(args ...string) {
	e.t.Helper()
	if out, err := exec.Command("git", append([]string{"-C", e.repo}, args...)...).CombinedOutput(); err != nil {
		e.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// claude installs the stub (or removes it) and points PATH at the stub plus
// the directories of git and sh only.
func (e *doctorEnv) claude(present bool) {
	e.t.Helper()
	bin := e.t.TempDir()
	for _, x := range []string{"git", "sh", "dirname", "cat"} {
		p, err := exec.LookPath(x)
		if err != nil {
			e.t.Skip(x + " not on PATH")
		}
		if err := os.Symlink(p, filepath.Join(bin, x)); err != nil {
			e.t.Fatal(err)
		}
	}
	path := bin
	if present {
		script := "#!/bin/sh\nif [ \"$*\" = \"--version\" ]; then echo 2.1.0; exit 0; fi\n" +
			"if [ \"$*\" = \"plugin list --json\" ]; then echo '" + e.plugins + "'; exit 0; fi\nexit 3\n"
		if err := os.WriteFile(filepath.Join(e.stub, "claude"), []byte(script), 0o755); err != nil {
			e.t.Fatal(err)
		}
		path = e.stub + string(os.PathListSeparator) + bin
	}
	e.t.Setenv("PATH", path)
}

func (e *doctorEnv) trust(v bool) {
	e.t.Helper()
	b, _ := json.Marshal(map[string]any{"projects": map[string]any{e.repo: map[string]any{"hasTrustDialogAccepted": v}}})
	if err := os.WriteFile(filepath.Join(e.home, ".claude.json"), b, 0o644); err != nil {
		e.t.Fatal(err)
	}
}

func (e *doctorEnv) write(rel, content string) {
	e.t.Helper()
	p := filepath.Join(e.repo, filepath.FromSlash(rel))
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		e.t.Fatal(err)
	}
}

func (e *doctorEnv) run(args ...string) (string, int) {
	e.t.Helper()
	out, _, code := runPluginCLI(e.t, e.build, e.repo, args...)
	return out, code
}

// check runs doctor --json and returns the named check (there must be one) and
// the exit code.
func (e *doctorEnv) check(name string) (doctorCheck, int) {
	e.t.Helper()
	out, code := e.run("--json", "doctor")
	var doc struct {
		OK     bool          `json:"ok"`
		Checks []doctorCheck `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		e.t.Fatalf("doctor --json: %v\n%s", err, out)
	}
	if doc.OK != (code == 0) {
		e.t.Errorf("ok = %v with exit %d", doc.OK, code)
	}
	var found []doctorCheck
	for _, c := range doc.Checks {
		if c.Name == name {
			found = append(found, c)
		}
	}
	if len(found) != 1 {
		e.t.Fatalf("checks named %q: %v\n%s", name, found, out)
	}
	return found[0], code
}

func (e *doctorEnv) want(name, result string) doctorCheck {
	e.t.Helper()
	c, _ := e.check(name)
	if c.Result != result {
		e.t.Errorf("%s = %s %q, want %s", name, c.Result, c.Message, result)
	}
	return c
}

func (e *doctorEnv) plan(status, shell string) {
	e.t.Helper()
	e.write("docs/briefs/B20260101-0900-a.loop-brief.md", "x\n")
	doc, _ := json.Marshal(map[string]any{
		"schema": "state/v2", "run_id": "B20260101-0900-a", "brief": "docs/briefs/B20260101-0900-a.loop-brief.md",
		"base": strings.Repeat("0123456789", 4), "branch": "B20260101-0900-a", "status": status, "iteration": 0,
		"created": "2026-01-01T09:00:00Z", "updated": "2026-01-01T09:00:00Z", "shell": shell,
		"checks": []any{}, "gate_scratch": []string{}, "gate_review": map[string]any{"rounds": 0, "verdict": ""},
		"tasks": []any{map[string]any{"id": "T1", "title": "t", "goal": "g", "kind": "feature", "fixtures": "",
			"references": []any{}, "depends_on": []string{}, "acceptance": []string{"a"}, "verify": "true",
			"status": "pending", "attempts": 0, "notes": ""}},
	})
	e.write(".vloop/state/state.json", string(doc))
}

func TestDoctorWorkedExampleText(t *testing.T) {
	e := newDoctorEnv(t)
	out, code := e.run("doctor")
	want := "✓ git\n✓ install\n✓ config\n✓ claude\n✓ trust\n✓ gate shell\n- plan\n- branch\n✓ plugin\n- self-hosting\ndoctor: 0 problem(s), 0 warning(s)\n"
	if code != 0 || out != want {
		t.Errorf("exit %d, output:\n%s\nwant:\n%s", code, out, want)
	}
}

func TestDoctorGitInstallConfigClaude(t *testing.T) {
	e := newDoctorEnv(t)
	e.want("git", resPass)
	e.git("config", "--unset", "user.email")
	e.want("git", resProblem)
	e.git("config", "user.email", "t@example.com")
	e.git("config", "--unset", "user.name")
	e.want("git", resProblem)
	e.git("config", "user.name", "t")

	e.want("install", resPass)
	e.build.Version = "0.2.0"
	if c := e.want("install", resWarning); !strings.Contains(c.Message, "run vloop upgrade") {
		t.Errorf("older stamp message %q", c.Message)
	}
	e.build.Version = "0.0.9"
	e.want("install", resProblem)
	e.build.Version = "0.1.0"
	os.Remove(filepath.Join(e.repo, ".vloop", "install.json"))
	e.want("install", resWarning)

	e.want("config", resPass)
	e.write(".vloop/config.toml", "language = \n")
	e.want("config", resProblem)
	e.write(".vloop/config.toml", "language = \"fr\"\n")
	e.want("config", resProblem)
	e.write(".vloop/config.toml", "")
	e.want("config", resPass)

	e.want("claude", resPass)
	e.claude(false)
	e.want("claude", resProblem)
}

func TestDoctorNotAGitRepository(t *testing.T) {
	e := newDoctorEnv(t)
	if err := os.RemoveAll(filepath.Join(e.repo, ".git")); err != nil {
		t.Fatal(err)
	}
	e.want("git", resProblem)
	e.want("self-hosting", resNA)
}

func TestDoctorTrust(t *testing.T) {
	e := newDoctorEnv(t)
	e.want("trust", resPass)
	e.trust(false)
	if c := e.want("trust", resWarning); !strings.Contains(c.Message, "run claude here once and accept the trust dialog") {
		t.Errorf("message %q", c.Message)
	}
	os.Remove(filepath.Join(e.home, ".claude.json"))
	e.want("trust", resWarning)
	if err := os.WriteFile(filepath.Join(e.home, ".claude.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.want("trust", resWarning)
}

func TestDoctorGateShellPlanAndBranch(t *testing.T) {
	e := newDoctorEnv(t)
	e.want("gate shell", resPass)
	e.want("plan", resNA)
	e.want("branch", resNA)
	e.write(".vloop/config.toml", "shell = \"pwsh\"\n")
	e.want("gate shell", resWarning)
	e.write(".vloop/config.toml", "")

	e.git("checkout", "-q", "-b", "B20260101-0900-a")
	e.plan("running", "sh")
	e.want("plan", resPass)
	e.want("branch", resPass)
	e.want("gate shell", resPass)

	e.plan("running", "nonexistent-shell")
	e.want("gate shell", resProblem)

	e.plan("running", "sh")
	e.write(".vloop/state/state.json", strings.Replace(readFile(t, e.repo, ".vloop/state/state.json"), `"depends_on":[]`, `"depends_on":["T9"]`, 1))
	e.want("plan", resProblem)

	e.plan("running", "sh")
	e.git("checkout", "-q", "main")
	e.want("branch", resProblem)
	e.plan("planning", "sh")
	e.want("branch", resProblem)
	e.plan("complete", "sh")
	e.want("branch", resPass)
}

func TestDoctorPlugin(t *testing.T) {
	e := newDoctorEnv(t)
	e.want("plugin", resPass)
	for plugins, want := range map[string]string{
		`[]`: resWarning,
		`[{"id":"vloop@vloop","version":"0.1.0","enabled":false}]`:                                                      resWarning,
		`[{"id":"vloop@vloop","version":"0.0.9","enabled":true}]`:                                                       resWarning,
		`[{"id":"other@x","version":"0.1.0","enabled":true},{"id":"vloop@elsewhere","version":"0.1.0","enabled":true}]`: resPass,
	} {
		e.plugins = plugins
		e.claude(true)
		e.want("plugin", want)
	}
	e.claude(false)
	e.want("plugin", resNA)
}

func TestDoctorStacks(t *testing.T) {
	e := newDoctorEnv(t)
	out, _ := e.run("--json", "doctor")
	if strings.Contains(out, `"stacks"`) {
		t.Errorf("a stacks check without a scoped entry:\n%s", out)
	}
	e.write("services/api/x", "x")
	e.write(".vloop/config.toml", "[metrics]\nstacks = [\"go\", \"csharp@services/api\"]\n")
	e.want("stacks", resPass)
	if err := os.RemoveAll(filepath.Join(e.repo, "services")); err != nil {
		t.Fatal(err)
	}
	if c := e.want("stacks", resWarning); c.Message != "csharp@services/api: no such directory" {
		t.Errorf("message %q", c.Message)
	}
	out, code := e.run("doctor")
	if code != 0 || !strings.Contains(out, "! stacks csharp@services/api: no such directory\n") {
		t.Errorf("exit %d:\n%s", code, out)
	}
}

func TestDoctorSelfHosting(t *testing.T) {
	e := newDoctorEnv(t)
	e.want("self-hosting", resNA)
	e.write("go.mod", "module github.com/mvelosop/vloop\n\ngo 1.22\n")
	e.want("self-hosting", resPass)
	head, err := exec.Command("git", "-C", e.repo, "rev-parse", "--short=7", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	e.build.Commit = strings.TrimSpace(string(head))
	e.want("self-hosting", resWarning)
}

func TestDoctorSelfHostingUnstamped(t *testing.T) {
	e := newDoctorEnv(t)
	e.write("go.mod", "module github.com/mvelosop/vloop\n\ngo 1.22\n")
	for _, c := range []string{"", "unknown"} {
		e.build.Commit = c
		e.want("self-hosting", resWarning)
	}
}

func TestDoctorWritesNothing(t *testing.T) {
	e := newDoctorEnv(t)
	snap := func() string {
		var s strings.Builder
		for _, root := range []string{e.repo, e.home} {
			filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
				if err != nil || strings.Contains(p, string(filepath.Separator)+".git"+string(filepath.Separator)) {
					return nil
				}
				if !fi.IsDir() {
					b, _ := os.ReadFile(p)
					s.WriteString(p + "\x00" + string(b) + "\x00")
				}
				return nil
			})
		}
		return s.String()
	}
	before := snap()
	e.run("doctor")
	e.run("--json", "doctor")
	if snap() != before {
		t.Error("doctor changed files")
	}
}
