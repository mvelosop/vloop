package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/schema"
)

const fakeSelfUpdateGo = `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	log := os.Getenv("FAKE_GO_LOG")
	dir, _ := os.Getwd()
	_ = os.WriteFile(log, []byte(strings.Join(args, " ")+"\n"+dir+"\n"), 0o666)
	if len(args) > 1 && args[0] == "list" && args[1] == "-m" {
		if os.Getenv("FAKE_GO_LIST_FAIL") != "" {
			fmt.Fprintln(os.Stderr, "first error")
			os.Exit(1)
		}
		fmt.Printf("{\"Path\":\"github.com/mvelosop/vloop\",\"Version\":%q}\n", os.Getenv("FAKE_GO_LATEST"))
		return
	}
	if len(args) > 1 && args[0] == "env" {
		if os.Getenv("FAKE_GO_ENV_FAIL") != "" {
			fmt.Fprintln(os.Stderr, "env detail")
			fmt.Fprintln(os.Stderr, os.Getenv("FAKE_GO_ENV_FAIL"))
			os.Exit(1)
		}
		if args[1] == "GOBIN" {
			fmt.Println(os.Getenv("FAKE_GOBIN"))
			return
		}
		if args[1] == "GOPATH" {
			fmt.Println(os.Getenv("FAKE_GOPATH"))
			return
		}
	}
	os.Exit(9)
}
`

func TestSelfUpdateCheckAndJSON(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)

	for _, tc := range []struct {
		name   string
		args   []string
		output string
	}{
		{"latest newer", []string{"self-update", "--check"}, "current  v0.14.0\nlatest   v0.15.0 — vloop self-update installs it\n"},
		{"explicit equal", []string{"self-update", "--check", "--version", "v0.14.0"}, "current  v0.14.0\ntarget   v0.14.0 — already running\n"},
		{"explicit older", []string{"self-update", "--check", "--version", "v0.13.0"}, "current  v0.14.0\ntarget   v0.13.0 — vloop self-update --version v0.13.0 installs it\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := h.run(tc.args...)
			if code != 0 || out != tc.output || errOut != "" {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, out, errOut)
			}
		})
	}

	code, out, errOut := h.run("--json", "self-update", "--check")
	if code != 0 || errOut != "" {
		t.Fatalf("json: code=%d stderr=%q stdout=%q", code, errOut, out)
	}
	var doc selfUpdateDocument
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Schema != "self-update/v1" || doc.Current != "v0.14.0" || doc.Target != "v0.15.0" || !doc.Newer || doc.Installed || doc.Dir != "" || len(doc.Plugins) != 0 {
		t.Fatalf("unexpected document: %+v", doc)
	}
	violations, err := schema.Validate("self-update/v1", []byte(out))
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("schema violations: %+v", violations)
	}
}

func TestSelfUpdateRefusals(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)

	code, out, errOut := h.run("self-update", "--version", "0.15")
	assertSelfUpdateResult(t, code, out, errOut, 2, "", "vloop: --version 0.15 is not a version like v0.15.0\n")
	code, out, errOut = h.run("--json", "self-update", "--version", "0.15")
	assertSelfUpdateResult(t, code, out, errOut, 2, "", "vloop: --version 0.15 is not a version like v0.15.0\n")

	h.build.Module = ""
	code, out, errOut = h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 1, "", "vloop: this vloop was built from a checkout, not installed from a release — build it again from the checkout, or go install github.com/mvelosop/vloop/cmd/vloop@latest\n")
	h.build.Module = "v0.14.0"

	t.Setenv("FAKE_GO_LIST_FAIL", "1")
	code, out, errOut = h.run("self-update", "--check")
	assertSelfUpdateResult(t, code, out, errOut, 1, "", "vloop: cannot find the latest release: first error\n")
	t.Setenv("FAKE_GO_LIST_FAIL", "")

	t.Setenv("FAKE_GO_LIST_FAIL", "must not be called")
	code, out, errOut = h.run("self-update", "--check", "--version", "v0.14.0")
	assertSelfUpdateResult(t, code, out, errOut, 0, "current  v0.14.0\ntarget   v0.14.0 — already running\n", "")
	t.Setenv("FAKE_GO_LIST_FAIL", "")

	t.Setenv("PATH", t.TempDir())
	code, out, errOut = h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 1, "", "vloop: self-update needs go on PATH — install the release by hand: go install github.com/mvelosop/vloop/cmd/vloop@latest\n")
}

func TestSelfUpdateJSONNonCheckBranches(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)

	h.build.Module = "v0.15.0-rc.4"
	t.Setenv("FAKE_GO_LATEST", "v0.14.0")
	code, out, errOut := h.run("--json", "self-update")
	assertSelfUpdateResult(t, code, out, errOut, 0,
		"{\"schema\":\"self-update/v1\",\"current\":\"v0.15.0-rc.4\",\"target\":\"v0.14.0\",\"newer\":false,\"installed\":false,\"dir\":\"\",\"plugins\":[]}\n",
		"",
	)
	assertSelfUpdateDocumentValid(t, out)

	h.build.Module = "v0.14.0"
	t.Setenv("FAKE_GO_LATEST", "v0.15.0")
	code, out, errOut = h.run("--json", "self-update")
	assertSelfUpdateResult(t, code, out, errOut, 0,
		fmt.Sprintf("{\"schema\":\"self-update/v1\",\"current\":\"v0.14.0\",\"target\":\"v0.15.0\",\"newer\":true,\"installed\":false,\"dir\":%q,\"plugins\":[]}\n",
			displaySelfUpdatePath(h.gobin)),
		"",
	)
	assertSelfUpdateDocumentValid(t, out)
}

func TestSelfUpdateInstallLocationChecksOnlyWhenNeeded(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)

	t.Setenv("FAKE_GO_ENV_FAIL", "must not be called")
	code, out, errOut := h.run("self-update", "--check")
	if code != 0 || out != "current  v0.14.0\nlatest   v0.15.0 — vloop self-update installs it\n" || errOut != "" {
		t.Fatalf("check called go env: code=%d stdout=%q stderr=%q", code, out, errOut)
	}

	h.build.Module = "v0.15.0-rc.4"
	t.Setenv("FAKE_GO_LATEST", "v0.14.0")
	code, out, errOut = h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 0, "vloop v0.15.0-rc.4 is up to date (latest release v0.14.0)\n", "")

	h.build.Module = "v0.14.0"
	t.Setenv("FAKE_GO_LATEST", "v0.15.0")
	t.Setenv("FAKE_GO_ENV_FAIL", "")
	other := filepath.Join(h.temp, "other", "vloop")
	if runtime.GOOS == "windows" {
		other += ".exe"
	}
	if err := os.MkdirAll(filepath.Dir(other), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte("running"), 0o755); err != nil {
		t.Fatal(err)
	}
	selfUpdateExecutable = func() (string, error) { return other, nil }
	code, out, errOut = h.run("self-update")
	wantErr := fmt.Sprintf(
		"vloop: this vloop runs from %s, but go install writes to %s — update it the way you installed it, or put %s first on PATH\n",
		displaySelfUpdatePath(other), displaySelfUpdatePath(h.gobin), displaySelfUpdatePath(h.gobin),
	)
	assertSelfUpdateResult(t, code, out, errOut, 1, "", wantErr)

	selfUpdateExecutable = func() (string, error) { return filepath.Join(h.gobin, filepath.Base(h.executable)), nil }
	t.Setenv("FAKE_GO_ENV_FAIL", "go env broke")
	code, out, errOut = h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 1, "", "vloop: cannot find where go install writes: go env broke\n")
	t.Setenv("FAKE_GO_ENV_FAIL", "")
	t.Setenv("FAKE_GOBIN", "")
	t.Setenv("FAKE_GOPATH", "")
	code, out, errOut = h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 1, "", "vloop: cannot find where go install writes: GOBIN and GOPATH are empty\n")

	gopath := filepath.Join(h.temp, "gopath")
	expected := filepath.Join(gopath, "bin")
	selfUpdateExecutable = func() (string, error) { return filepath.Join(expected, filepath.Base(h.executable)), nil }
	t.Setenv("FAKE_GOPATH", gopath)
	if err := os.MkdirAll(expected, 0o755); err != nil {
		t.Fatal(err)
	}
	installExe := filepath.Join(expected, filepath.Base(h.executable))
	if err := os.WriteFile(installExe, []byte("running"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, out, errOut = h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 0, "", "")
}

func TestSelfUpdateRunsLatestLookupOutsideModule(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)
	code, _, errOut := h.run("self-update", "--check")
	if code != 0 {
		t.Fatal(errOut)
	}
	args, err := os.ReadFile(h.log)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(args)), "\n")
	if len(lines) != 2 || lines[0] != "list -m -json github.com/mvelosop/vloop@latest" {
		t.Fatalf("go invocation: %q", args)
	}
	if _, err := os.Stat(filepath.Join(lines[1], "go.mod")); !os.IsNotExist(err) {
		t.Fatalf("latest lookup ran inside module: %q", lines[1])
	}
}

type selfUpdateHarness struct {
	temp, gobin, log, executable string
	build                        Build
	oldExecutable                func() (string, error)
}

func newSelfUpdateHarness(t *testing.T) *selfUpdateHarness {
	t.Helper()
	temp := t.TempDir()
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(temp, "fakego.go")
	if err := os.WriteFile(src, []byte(fakeSelfUpdateGo), 0o666); err != nil {
		t.Fatal(err)
	}
	fakeDir := filepath.Join(temp, "fake")
	gobin := filepath.Join(temp, "gobin")
	if err := os.MkdirAll(fakeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(gobin, 0o755); err != nil {
		t.Fatal(err)
	}
	goName := "go"
	if runtime.GOOS == "windows" {
		goName += ".exe"
	}
	fakeGo := filepath.Join(fakeDir, goName)
	build := exec.Command(realGo, "build", "-o", fakeGo, src)
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build fake go: %v: %s", err, raw)
	}
	exeName := "vloop"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}
	executable := filepath.Join(gobin, exeName)
	if err := os.WriteFile(executable, []byte("running"), 0o755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(temp, "go.log")
	t.Setenv("PATH", fakeDir)
	t.Setenv("FAKE_GOBIN", gobin)
	t.Setenv("FAKE_GOPATH", "")
	t.Setenv("FAKE_GO_LATEST", "v0.15.0")
	t.Setenv("FAKE_GO_LIST_FAIL", "")
	t.Setenv("FAKE_GO_ENV_FAIL", "")
	t.Setenv("FAKE_GO_LOG", log)
	oldExecutable := selfUpdateExecutable
	selfUpdateExecutable = func() (string, error) { return executable, nil }
	return &selfUpdateHarness{
		temp: temp, gobin: gobin, log: log, executable: executable,
		build:         Build{Version: "0.14.0", Module: "v0.14.0"},
		oldExecutable: oldExecutable,
	}
}

func (h *selfUpdateHarness) restoreExecutable() { selfUpdateExecutable = h.oldExecutable }

func (h *selfUpdateHarness) run(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := Execute(h.build, args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func assertSelfUpdateDocumentValid(t *testing.T, out string) {
	t.Helper()
	violations, err := schema.Validate("self-update/v1", []byte(out))
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("schema violations: %+v", violations)
	}
}

func assertSelfUpdateResult(t *testing.T, code int, out, errOut string, wantCode int, wantOut, wantErr string) {
	t.Helper()
	if code != wantCode || out != wantOut || errOut != wantErr {
		t.Fatalf("code=%d stdout=%q stderr=%q; want %d %q %q", code, out, errOut, wantCode, wantOut, wantErr)
	}
}
