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

func TestMixedOSDoctorAndPreflight(t *testing.T) {
	bin := buildVloop(t)
	stubPATH(t)
	root := t.TempDir()
	needMkdir(t, filepath.Join(root, ".vloop"))
	writeFile(t, filepath.Join(root, ".vloop", "config.toml"), checkConfig("go", "go version"))

	doc, code, errText := doctor(t, bin, root, nil)
	_ = code
	if errText != "" {
		t.Fatalf("doctor stderr %q", errText)
	}
	wantShell := "sh"
	if runtime.GOOS == "windows" {
		wantShell = "pwsh"
	}
	needDoctorMessage(t, doc, "checks", "warning",
		"shell is not set in .vloop/config.toml, so each machine runs the checks in its own ("+wantShell+" here): set it there")

	doc, _, _ = doctor(t, bin, root, map[string]string{"VLOOP_SHELL": "bash"})
	needDoctorMessage(t, doc, "checks", "warning",
		"shell is not set in .vloop/config.toml, so each machine runs the checks in its own (bash here): set it there")

	writeFile(t, filepath.Join(root, ".vloop", "config.toml"), "platforms = [\"windows\", \"macos\", \"macos\"]\n"+checkConfig("go", "go version"))
	doc, code, _ = doctor(t, bin, root, nil)
	needDoctorMessage(t, doc, "platforms", "problem",
		`macos, windows need shell = "pwsh" in .vloop/config.toml; it is not set`)
	needOrder(t, doc, "config", "platforms")
	if code != 1 {
		t.Fatalf("platforms doctor exit = %d, want 1", code)
	}

	writeFile(t, filepath.Join(root, ".vloop", "config.toml"), "platforms = [\"windows\", \"macos\"]\nshell = \"sh\"\n"+checkConfig("go", "go version"))
	doc, _, _ = doctor(t, bin, root, nil)
	needDoctorMessage(t, doc, "platforms", "problem",
		`macos, windows need shell = "pwsh" in .vloop/config.toml; it is "sh"`)

	writeFile(t, filepath.Join(root, ".vloop", "config.toml"), "platforms = [\"windows\", \"macos\"]\nshell = \"pwsh\"\n"+checkConfig("go", "go version"))
	doc, _, _ = doctor(t, bin, root, nil)
	host := runtime.GOOS
	if host == "darwin" {
		host = "macos"
	}
	if host == "linux" {
		needDoctorMessage(t, doc, "platforms", "warning", "this machine (linux) is not one of macos, windows")
	} else {
		needDoctorMessage(t, doc, "platforms", "pass", "macos, windows")
	}

	emptyPATH := t.TempDir()
	doc, _, _ = doctor(t, bin, root, map[string]string{"PATH": emptyPATH})
	needDoctorMessage(t, doc, "gate shell", "problem", `the configured shell "pwsh" is not on PATH`)

	writeFile(t, filepath.Join(root, ".vloop", "config.toml"), "platforms = [\"windows\", \"macos\"]\n"+checkConfig("go", "go test ./... | grep ok"))
	doc, _, _ = doctor(t, bin, root, nil)
	needDoctorMessage(t, doc, "checks", "warning",
		"go calls grep, which is not on every platform in platforms (macos, windows)")

	brief := filepath.Join(root, "docs", "briefs", "B20260101-0000-gate.loop-brief.md")
	writeFile(t, brief, "---\nname: B20260101-0000-gate.loop-brief\nstatus: ready\n---\n# Brief\n\n## Shape\n\n1 to 1 task.\n")
	code, _, errText = runVloop(t, bin, root, nil, "run", "--plan-only", "docs/briefs/B20260101-0000-gate.loop-brief.md")
	needContains(t, errText, "  ✗ platforms macos, windows need shell = \"pwsh\" in .vloop/config.toml; it is not set")
	needContains(t, errText, "  ! checks go calls grep, which is not on every platform in platforms (macos, windows)")
	if code != 1 {
		t.Fatalf("run exit = %d, want 1", code)
	}

	otherPlatforms := "linux, macos"
	switch host {
	case "linux":
		otherPlatforms = "macos, windows"
	case "macos":
		otherPlatforms = "linux, windows"
	}
	writeFile(t, filepath.Join(root, ".vloop", "config.toml"), "platforms = [\""+strings.ReplaceAll(otherPlatforms, ", ", "\", \"")+"\"]\nshell = \"pwsh\"\n"+checkConfig("go", "go version"))
	code, _, errText = runVloop(t, bin, root, nil, "run", "--plan-only", "docs/briefs/B20260101-0000-gate.loop-brief.md")
	needContains(t, errText, "  ! platforms this machine ("+host+") is not one of "+otherPlatforms)
	needContains(t, errText, "preflight failed")
	if code != 1 {
		t.Fatalf("run warning case exit = %d, want 1", code)
	}
}

func doctor(t *testing.T, bin, root string, extra map[string]string) (doctorDocument, int, string) {
	t.Helper()
	code, out, errText := runVloop(t, bin, root, extra, "doctor", "--json")
	var doc doctorDocument
	if e := json.Unmarshal([]byte(out), &doc); e != nil {
		t.Fatalf("decode doctor: %v\nstdout: %s\nstderr: %s", e, out, errText)
	}
	return doc, code, errText
}

func needDoctorMessage(t *testing.T, doc doctorDocument, name, result, message string) {
	t.Helper()
	for _, c := range doc.Checks {
		if c.Name == name && c.Result == result && c.Message == message {
			return
		}
	}
	t.Fatalf("no %s %s row %q in %#v", name, result, message, doc.Checks)
}

func needOrder(t *testing.T, doc doctorDocument, first, second string) {
	t.Helper()
	a, b := -1, -1
	for i, c := range doc.Checks {
		if c.Name == first {
			a = i
		}
		if c.Name == second {
			b = i
		}
	}
	if a < 0 || b < 0 || b != a+1 {
		t.Fatalf("%s and %s are not consecutive in order: %#v", first, second, doc.Checks)
	}
}

func checkConfig(name, run string) string {
	return "\n[[check]]\nname = \"" + name + "\"\npaths = [\"**\"]\nrun = \"" + run + "\"\n"
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
	cmd.Env = cleanEnv(extra)
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

func cleanEnv(extra map[string]string) []string {
	env := []string{}
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		if strings.HasPrefix(key, "VLOOP_") {
			continue
		}
		if _, ok := extra[key]; ok {
			continue
		}
		env = append(env, e)
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

func needMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	needMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func needWrite(t *testing.T, path, text string) {
	t.Helper()
	needMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func needContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("output does not contain %q:\n%s", want, got)
	}
}
