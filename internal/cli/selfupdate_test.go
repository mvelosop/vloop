package cli

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"path/filepath"
	"runtime"
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
	if len(args) > 1 && args[0] == "install" {
		_ = os.WriteFile(log, []byte(strings.Join(args, " ")+"\n"+dir+"\n"), 0o666)
		name := "vloop"
		if runtime.GOOS == "windows" || os.Getenv("FAKE_GO_INSTALL_WINDOWS") != "" {
			name += ".exe"
		}
		installDir := os.Getenv("FAKE_GOBIN")
		if installDir == "" {
			installDir = filepath.Join(os.Getenv("FAKE_GOPATH"), "bin")
		}
		exe := filepath.Join(installDir, name)
		if os.Getenv("FAKE_GO_INSTALL_PARTIAL") != "" {
			_ = os.WriteFile(exe, []byte("partial"), 0o755)
		}
		if os.Getenv("FAKE_GO_INSTALL_FAIL") != "" {
			fmt.Fprintln(os.Stderr, "install detail")
			fmt.Fprintln(os.Stderr, os.Getenv("FAKE_GO_INSTALL_FAIL"))
			os.Exit(1)
		}
		if err := os.WriteFile(exe, []byte("installed"), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	os.Exit(9)
}
`

const fakeSelfUpdateAgent = `package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	agent := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	args := os.Args[1:]
	if log := os.Getenv("FAKE_AGENT_LOG"); log != "" {
		f, _ := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
		if f != nil {
			dir, _ := os.Getwd()
			fmt.Fprintln(f, agent+"|"+filepath.ToSlash(dir)+"|"+strings.Join(args, " "))
			_ = f.Close()
		}
	}
	key := strings.ToUpper(agent)
	if len(args) >= 3 && args[0] == "plugin" && args[1] == "list" && args[2] == "--json" {
		if fail := os.Getenv("FAKE_" + key + "_LIST_FAIL"); fail != "" {
			fmt.Fprintln(os.Stderr, fail)
			os.Exit(1)
		}
		fmt.Println(os.Getenv("FAKE_" + key + "_LIST"))
		return
	}
	if len(args) >= 3 && args[0] == "plugin" && args[1] == "marketplace" && args[2] == "update" {
		if fail := os.Getenv("FAKE_" + key + "_MARKETPLACE_FAIL"); fail != "" {
			fmt.Fprintln(os.Stderr, "marketplace detail")
			fmt.Fprintln(os.Stderr, fail)
			os.Exit(1)
		}
		return
	}
	if len(args) >= 2 && args[0] == "plugin" && args[1] == "update" {
		if fail := os.Getenv("FAKE_" + key + "_UPDATE_FAIL"); fail != "" {
			fmt.Fprintln(os.Stderr, "update detail")
			fmt.Fprintln(os.Stderr, fail)
			os.Exit(1)
		}
		return
	}
	fmt.Fprintln(os.Stderr, "unexpected fake "+agent+" arguments:", strings.Join(args, " "))
	os.Exit(1)
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
	// Go stamps a build from a checkout with a pseudo-version, +dirty when the
	// tree was modified: still a checkout build, not a release.
	for _, m := range []string{"v0.15.0-rc.4.0.20261126155805-37d1d135e2c6", "v0.15.0-rc.4.0.20261126155805-37d1d135e2c6+dirty", "v0.14.0+dirty"} {
		h.build.Module = m
		code, out, errOut = h.run("self-update")
		assertSelfUpdateResult(t, code, out, errOut, 1, "", "vloop: this vloop was built from a checkout, not installed from a release — build it again from the checkout, or go install github.com/mvelosop/vloop/cmd/vloop@latest\n")
	}
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
		fmt.Sprintf("{\"schema\":\"self-update/v1\",\"current\":\"v0.14.0\",\"target\":\"v0.15.0\",\"newer\":true,\"installed\":true,\"dir\":%q,\"plugins\":[]}\n",
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
	assertSelfUpdateResult(t, code, out, errOut, 0,
		"installing vloop v0.15.0 (from v0.14.0)…\n"+
			"installed vloop v0.15.0 at "+displaySelfUpdatePath(expected)+"\n"+
			"next: vloop upgrade in each repository that uses vloop, and vloop self-update there for a plugin installed for that repository\n",
		"",
	)
}

func TestSelfUpdateInstalls(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)

	code, out, errOut := h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 0,
		"installing vloop v0.15.0 (from v0.14.0)…\n"+
			"installed vloop v0.15.0 at "+displaySelfUpdatePath(h.gobin)+"\n"+
			"next: vloop upgrade in each repository that uses vloop, and vloop self-update there for a plugin installed for that repository\n",
		"",
	)
	log, err := os.ReadFile(h.log)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Split(string(log), "\n")[0]; got != "install github.com/mvelosop/vloop/cmd/vloop@v0.15.0" {
		t.Fatalf("go invocation: %q", got)
	}
	installed, err := os.ReadFile(h.executable)
	if err != nil {
		t.Fatal(err)
	}
	if string(installed) != "installed" {
		t.Fatalf("installed executable: %q", installed)
	}

	code, out, errOut = h.run("--json", "self-update", "--version", "v0.13.0")
	assertSelfUpdateResult(t, code, out, errOut, 0,
		fmt.Sprintf("{\"schema\":\"self-update/v1\",\"current\":\"v0.14.0\",\"target\":\"v0.13.0\",\"newer\":false,\"installed\":true,\"dir\":%q,\"plugins\":[]}\n",
			displaySelfUpdatePath(h.gobin)),
		"",
	)
	assertSelfUpdateDocumentValid(t, out)
}

func TestSelfUpdateInstallFailure(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)
	t.Setenv("FAKE_GO_INSTALL_FAIL", "no matching versions")

	code, out, errOut := h.run("self-update")
	wantOut := "installing vloop v0.15.0 (from v0.14.0)…\n"
	assertSelfUpdateResult(t, code, out, errOut, 1, wantOut, "vloop: go install failed: no matching versions\n")
	raw, err := os.ReadFile(h.executable)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "running" {
		t.Fatalf("running executable changed: %q", raw)
	}
}

func TestSelfUpdateWindowsMoveAsideFailures(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)
	oldGOOS := selfUpdateGOOS
	oldRemove := selfUpdateRemove
	oldRename := selfUpdateRename
	selfUpdateGOOS = "windows"
	t.Cleanup(func() {
		selfUpdateGOOS = oldGOOS
		selfUpdateRemove = oldRemove
		selfUpdateRename = oldRename
	})
	useWindowsSelfUpdateExecutable(t, h)
	exe := h.executable
	old := exe + ".old"

	if err := os.Mkdir(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "locked"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := h.run("self-update")
	if code != 1 || !strings.HasPrefix(errOut, "vloop: cannot remove "+displaySelfUpdatePath(old)+" (") ||
		!strings.HasSuffix(errOut, ") — delete it and run vloop self-update again\n") {
		t.Fatalf("old removal: code=%d stderr=%q", code, errOut)
	}
	if log, err := os.ReadFile(h.log); err == nil && strings.Contains(string(log), "install ") {
		t.Fatal("install ran after stale old removal failure")
	}
	if err := os.RemoveAll(old); err != nil {
		t.Fatal(err)
	}

	selfUpdateRename = func(oldPath, newPath string) error {
		if oldPath == exe {
			return errors.New("move denied")
		}
		return oldRename(oldPath, newPath)
	}
	code, _, errOut = h.run("self-update")
	if want := fmt.Sprintf("vloop: cannot move %s aside (move denied) — nothing was installed\n", displaySelfUpdatePath(exe)); errOut != want || code != 1 {
		t.Fatalf("move aside: code=%d stderr=%q", code, errOut)
	}
	if log, err := os.ReadFile(h.log); err == nil && strings.Contains(string(log), "install ") {
		t.Fatal("install ran after move-aside failure")
	}
}

func TestSelfUpdateWindowsRestoresAfterInstallFailure(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)
	oldGOOS := selfUpdateGOOS
	oldRename := selfUpdateRename
	selfUpdateGOOS = "windows"
	t.Cleanup(func() {
		selfUpdateGOOS = oldGOOS
		selfUpdateRename = oldRename
	})
	useWindowsSelfUpdateExecutable(t, h)
	t.Setenv("FAKE_GO_INSTALL_WINDOWS", "1")
	t.Setenv("FAKE_GO_INSTALL_FAIL", "no matching versions")
	t.Setenv("FAKE_GO_INSTALL_PARTIAL", "1")

	original := h.executable
	old := original + ".old"
	code, out, errOut := h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 1,
		"installing vloop v0.15.0 (from v0.14.0)…\n",
		"vloop: go install failed: no matching versions\n",
	)
	raw, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "running" {
		t.Fatalf("running executable was not restored: %q", raw)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("old executable still exists after restore: %v", err)
	}

	t.Setenv("FAKE_GO_INSTALL_PARTIAL", "")
	t.Setenv("FAKE_GO_INSTALL_WINDOWS", "1")
	t.Setenv("FAKE_GO_INSTALL_FAIL", "no matching versions")
	original = h.executable
	old = original + ".old"
	selfUpdateRename = func(oldPath, newPath string) error {
		if oldPath == old {
			return errors.New("restore denied")
		}
		return oldRename(oldPath, newPath)
	}
	code, out, errOut = h.run("self-update")
	want := fmt.Sprintf(
		"vloop: go install failed: no matching versions; and %s could not be moved back to %s (restore denied) — rename it back by hand\n",
		displaySelfUpdatePath(old), displaySelfUpdatePath(original),
	)
	assertSelfUpdateResult(t, code, out, errOut, 1,
		"installing vloop v0.15.0 (from v0.14.0)…\n",
		want,
	)
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

func TestSelfUpdatePluginsReuseDoctorDiscovery(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)
	h.addAgents(t, "claude", "copilot")
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FAKE_GO_LATEST", "v0.14.0")
	t.Setenv("FAKE_CLAUDE_LIST", fmt.Sprintf(
		`[{"id":"other@x","version":"1","scope":"user"},{"id":"vloop@vloop","version":"0.14.0","scope":""},{"id":"vloop@vloop","version":"0.14.0","scope":"project","projectPath":%q},{"id":"vloop@vloop","version":"","scope":"user"},{"id":"vloop@vloop","version":"0.14.0","scope":"local","projectPath":%q}]`,
		filepath.ToSlash(root), filepath.ToSlash(filepath.Join(t.TempDir(), "other")),
	))
	t.Setenv("FAKE_COPILOT_LIST", `[{"name":"other"},{"name":"vloop"}]`)

	code, out, errOut := h.run("self-update")
	want := "vloop v0.14.0 is up to date (latest release v0.14.0)\n" +
		"plugin claude (user): updated\n" +
		"plugin claude (project): updated\n" +
		"plugin copilot: updated\n" +
		"next: vloop upgrade in each repository that uses vloop, and vloop self-update there for a plugin installed for that repository\n"
	assertSelfUpdateResult(t, code, out, errOut, 0, want, "")

	raw, err := os.ReadFile(h.agentLog)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	prefix := filepath.ToSlash(root)
	wantCommands := []string{
		"claude|" + prefix + "|plugin list --json",
		"claude|" + prefix + "|plugin marketplace update vloop",
		"claude|" + prefix + "|plugin update vloop@vloop --scope user",
		"claude|" + prefix + "|plugin update vloop@vloop --scope project",
		"copilot|" + prefix + "|plugin list --json",
		"copilot|" + prefix + "|plugin update vloop@vloop",
	}
	if strings.Join(lines, "\n") != strings.Join(wantCommands, "\n") {
		t.Fatalf("agent order or working directory:\n%s", strings.Join(lines, "\n"))
	}
}

func TestSelfUpdatePluginFailuresAreStructuredAndNonFatal(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)
	h.addAgents(t, "claude", "copilot")
	t.Setenv("FAKE_GO_LATEST", "v0.14.0")
	t.Setenv("FAKE_CLAUDE_LIST", `[{"id":"vloop@vloop","version":"0.14.0","scope":"user"}]`)
	t.Setenv("FAKE_CLAUDE_UPDATE_FAIL", "network down")
	t.Setenv("FAKE_COPILOT_LIST", `not-json`)

	code, out, errOut := h.run("--json", "self-update")
	assertSelfUpdateResult(t, code, "", errOut, 0, "", "")
	var doc selfUpdateDocument
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	assertSelfUpdateDocumentValid(t, out)
	wantPlugins := []selfUpdatePlugin{
		{Agent: "claude", Scope: "user", Result: "failed", Message: "network down"},
		{Agent: "copilot", Result: "failed", Message: "invalid character 'o' in literal null (expecting 'u')"},
	}
	if len(doc.Plugins) != len(wantPlugins) {
		t.Fatalf("plugins: %+v", doc.Plugins)
	}
	for i := range wantPlugins {
		if doc.Plugins[i].Agent != wantPlugins[i].Agent || doc.Plugins[i].Scope != wantPlugins[i].Scope ||
			doc.Plugins[i].Result != wantPlugins[i].Result || doc.Plugins[i].Message != wantPlugins[i].Message {
			t.Fatalf("plugin %d: %+v; want %+v", i, doc.Plugins[i], wantPlugins[i])
		}
	}
}

func TestSelfUpdatePluginMarketplaceFailureAndCheckIsolation(t *testing.T) {
	h := newSelfUpdateHarness(t)
	t.Cleanup(h.restoreExecutable)
	h.addAgents(t, "claude")
	t.Setenv("FAKE_GO_LATEST", "v0.14.0")
	t.Setenv("FAKE_CLAUDE_LIST", `[{"id":"vloop@vloop","version":"0.14.0","scope":"local","projectPath":"different"}]`)

	code, out, errOut := h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 0,
		"vloop v0.14.0 is up to date (latest release v0.14.0)\nplugin claude: not installed\n", "")
	if err := os.Remove(h.agentLog); err != nil {
		t.Fatal(err)
	}

	t.Setenv("FAKE_CLAUDE_LIST", `[{"id":"vloop@vloop","version":"0.14.0","scope":"user"}]`)
	t.Setenv("FAKE_CLAUDE_MARKETPLACE_FAIL", "market unavailable")
	code, out, errOut = h.run("self-update")
	assertSelfUpdateResult(t, code, out, errOut, 0,
		"vloop v0.14.0 is up to date (latest release v0.14.0)\n"+
			"! plugin claude: claude plugin marketplace update vloop failed: market unavailable — run it by hand\n",
		"",
	)
	raw, err := os.ReadFile(h.agentLog)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "plugin update vloop@vloop") {
		t.Fatalf("plugin update ran after marketplace failure:\n%s", raw)
	}
	before := len(raw)

	code, _, errOut = h.run("self-update", "--check")
	if code != 0 || errOut != "" {
		t.Fatalf("check: code=%d stderr=%q", code, errOut)
	}
	after, err := os.ReadFile(h.agentLog)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != before {
		t.Fatal("--check inspected an agent plugin")
	}
}

type selfUpdateHarness struct {
	temp, gobin, log, executable string
	agentLog, realGo             string
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
	t.Setenv("FAKE_GO_INSTALL_FAIL", "")
	t.Setenv("FAKE_GO_INSTALL_PARTIAL", "")
	t.Setenv("FAKE_GO_INSTALL_WINDOWS", "")
	t.Setenv("FAKE_GO_LOG", log)
	agentLog := filepath.Join(temp, "agent.log")
	t.Setenv("FAKE_AGENT_LOG", agentLog)
	oldExecutable := selfUpdateExecutable
	selfUpdateExecutable = func() (string, error) { return executable, nil }
	return &selfUpdateHarness{
		temp: temp, gobin: gobin, log: log, executable: executable,
		agentLog: agentLog, realGo: realGo,
		build:         Build{Version: "0.14.0", Module: "v0.14.0"},
		oldExecutable: oldExecutable,
	}
}

func (h *selfUpdateHarness) addAgents(t *testing.T, agents ...string) {
	t.Helper()
	src := filepath.Join(h.temp, "fakeagent.go")
	if err := os.WriteFile(src, []byte(fakeSelfUpdateAgent), 0o666); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(h.temp, "fake", "agent")
	if runtime.GOOS == "windows" {
		fake += ".exe"
	}
	build := exec.Command(h.realGo, "build", "-o", fake, src)
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build fake agent: %v: %s", err, raw)
	}
	for _, agent := range agents {
		name := agent
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		dst := filepath.Join(h.temp, "fake", name)
		raw, err := os.ReadFile(fake)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, raw, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func useWindowsSelfUpdateExecutable(t *testing.T, h *selfUpdateHarness) {
	t.Helper()
	t.Setenv("FAKE_GO_INSTALL_WINDOWS", "1")
	if runtime.GOOS == "windows" {
		return
	}
	exe := h.executable + ".exe"
	if err := os.Rename(h.executable, exe); err != nil {
		t.Fatal(err)
	}
	h.executable = exe
	selfUpdateExecutable = func() (string, error) { return exe, nil }
}

func (h *selfUpdateHarness) restoreExecutable() {
	selfUpdateExecutable = h.oldExecutable
}

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
