package gate_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type doctorDocument struct {
	OK     bool `json:"ok"`
	Checks []struct {
		Name    string `json:"name"`
		Result  string `json:"result"`
		Message string `json:"message"`
	} `json:"checks"`
}

func TestWorkedMixedOSExample(t *testing.T) {
	root := t.TempDir()
	bin := buildVloop(t)
	stubPATH(t)
	needWrite(t, filepath.Join(root, ".vloop", "config.toml"), checkConfig("go version"))

	doc, _, _ := doctor(t, bin, root, nil)
	resolved := "sh"
	if runtime.GOOS == "windows" {
		resolved = "pwsh"
	}
	needRow(t, doc, "checks", "warning",
		"shell is not set in .vloop/config.toml, so each machine runs the checks in its own ("+resolved+" here): set it there")
	doc, _, _ = doctor(t, bin, root, map[string]string{"VLOOP_SHELL": "bash"})
	needRow(t, doc, "checks", "warning",
		"shell is not set in .vloop/config.toml, so each machine runs the checks in its own (bash here): set it there")

	code, out, stderr := runVloop(t, bin, root, nil, "config", "set", "platforms", "windows,macos")
	needExit(t, code, 0, out, stderr)
	code, out, stderr = runVloop(t, bin, root, nil, "config", "get", "platforms")
	needOutput(t, code, out, "windows,macos\n", stderr)

	doc, code, _ = doctor(t, bin, root, nil)
	needRow(t, doc, "platforms", "problem",
		`macos, windows need shell = "pwsh" in .vloop/config.toml; it is not set`)
	if code != 1 {
		t.Fatalf("doctor exit = %d, want 1", code)
	}

	code, out, stderr = runVloop(t, bin, root, nil, "config", "set", "shell", "pwsh")
	needExit(t, code, 0, out, stderr)
	doc, _, _ = doctor(t, bin, root, nil)
	host := runtime.GOOS
	if host == "darwin" {
		host = "macos"
	}
	if host == "linux" {
		needRow(t, doc, "platforms", "warning", "this machine (linux) is not one of macos, windows")
	} else {
		needRow(t, doc, "platforms", "pass", "macos, windows")
	}

	code, out, stderr = runVloop(t, bin, root, map[string]string{"VLOOP_SHELL": "pwsh"}, "config", "get", "shell")
	needExactFailure(t, code, 1, out, stderr,
		`vloop: VLOOP_SHELL: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell`)

	code, out, stderr = runVloop(t, bin, root, nil, "config", "set", "--local", "shell", "pwsh")
	needExactFailure(t, code, 2, out, stderr,
		`vloop: .vloop/config.local.toml: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell`)
	code, out, stderr = runVloop(t, bin, root, nil, "config", "set", "--local", "platforms", "linux")
	needExactFailure(t, code, 2, out, stderr,
		`vloop: .vloop/config.local.toml: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's`)
	code, out, stderr = runVloop(t, bin, root, nil, "config", "set", "platforms", "macos,beos")
	needExactFailure(t, code, 2, out, stderr,
		`vloop: invalid value "macos,beos" for platforms: want a comma-separated list of linux, macos, windows`)

	needWrite(t, filepath.Join(root, "docs", "briefs", "B20260101-0000-gate.loop-brief.md"),
		"---\nname: B20260101-0000-gate.loop-brief\nstatus: ready\n---\n# Brief\n\n## Shape\n\n1 to 1 task.\n")
	needWrite(t, filepath.Join(root, ".vloop", "config.toml"),
		"platforms = [\"windows\", \"macos\"]\n"+checkConfig("go test ./... | grep ok"))
	code, out, stderr = runVloop(t, bin, root, nil, "run", "--plan-only", "docs/briefs/B20260101-0000-gate.loop-brief.md")
	needContains(t, stderr, "  ✗ platforms macos, windows need shell = \"pwsh\" in .vloop/config.toml; it is not set")
	needContains(t, stderr, "  ! checks go calls grep, which is not on every platform in platforms (macos, windows)")
	needContains(t, stderr, "preflight failed")
	if code != 1 {
		t.Fatalf("run exit = %d, want 1; stdout %q stderr %q", code, out, stderr)
	}
}

func doctor(t *testing.T, bin, root string, extra map[string]string) (doctorDocument, int, string) {
	t.Helper()
	code, out, stderr := runVloop(t, bin, root, extra, "doctor", "--json")
	var doc doctorDocument
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("decode doctor: %v\nstdout: %s\nstderr: %s", err, out, stderr)
	}
	return doc, code, stderr
}

func needRow(t *testing.T, doc doctorDocument, name, result, message string) {
	t.Helper()
	for _, row := range doc.Checks {
		if row.Name == name && row.Result == result && row.Message == message {
			return
		}
	}
	t.Fatalf("missing %s/%s %q in %#v", name, result, message, doc.Checks)
}

func checkConfig(run string) string {
	return "[[check]]\nname = \"go\"\npaths = [\"**\"]\nrun = \"" + run + "\"\n"
}

func buildVloop(t *testing.T) string {
	t.Helper()
	name := "vloop"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/vloop")
	cmd.Dir = repoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build vloop: %v\n%s", err, out)
	}
	return bin
}

func stubPATH(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	needWrite(t, src, "package main\nimport (\"fmt\"; \"os\")\nfunc main() { if len(os.Args) > 1 && os.Args[1] == \"plugin\" { fmt.Print(\"[]\") } }\n")
	name := "claude"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, name), src)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build claude stub: %v\n%s", err, out)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func runVloop(t *testing.T, bin, root string, extra map[string]string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(bin, append([]string{"-C", root}, args...)...)
	cmd.Env = cleanVloopEnv(extra)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

func cleanVloopEnv(extra map[string]string) []string {
	env := []string{}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "VLOOP_") {
			env = append(env, e)
		}
	}
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	return env
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func needWrite(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func needExit(t *testing.T, code, want int, out, stderr string) {
	t.Helper()
	if code != want || out != "" || stderr != "" {
		t.Fatalf("exit %d stdout %q stderr %q, want %d with no output", code, out, stderr, want)
	}
}

func needOutput(t *testing.T, code int, got, want, stderr string) {
	t.Helper()
	if code != 0 || got != want || stderr != "" {
		t.Fatalf("exit %d stdout %q stderr %q, want stdout %q", code, got, stderr, want)
	}
}

func needExactFailure(t *testing.T, code, wantCode int, stdout, stderr, message string) {
	t.Helper()
	if code != wantCode || stdout != "" || strings.TrimSpace(stderr) != message {
		t.Fatalf("exit %d stdout %q stderr %q, want exit %d and %q", code, stdout, stderr, wantCode, message)
	}
}

func needContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("output does not contain %q:\n%s", want, got)
	}
}
