package main

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

const selfUpdateIntegrationMain = `package main

import (
	"os"

	"github.com/mvelosop/vloop/internal/cli"
)

func main() {
	os.Exit(cli.Execute(cli.Build{
		Version: "0.14.0",
		Commit: "integration",
		Module: "v0.14.0",
	}, os.Args[1:], os.Stdout, os.Stderr))
}
`

const selfUpdateIntegrationGo = `package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]
	if log := os.Getenv("FAKE_GO_LOG"); log != "" {
		f, _ := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
		if f != nil {
			fmt.Fprintln(f, strings.Join(args, " "))
			_ = f.Close()
		}
	}
	switch {
	case len(args) == 4 && args[0] == "list" && args[1] == "-m" && args[2] == "-json":
		fmt.Println("{\"Path\":\"github.com/mvelosop/vloop\",\"Version\":\"v0.15.0\"}")
	case len(args) == 2 && args[0] == "env" && args[1] == "GOBIN":
		fmt.Println(os.Getenv("FAKE_GOBIN"))
	case len(args) == 2 && args[0] == "install":
		name := "vloop"
		if os.PathSeparator == '\\' {
			name += ".exe"
		}
		if err := os.WriteFile(filepath.Join(os.Getenv("FAKE_GOBIN"), name), []byte("installed"), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unexpected fake go arguments:", strings.Join(args, " "))
		os.Exit(1)
	}
}
`

const selfUpdateIntegrationAgent = `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	if log := os.Getenv("FAKE_CLAUDE_LOG"); log != "" {
		f, _ := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
		if f != nil {
			fmt.Fprintln(f, strings.Join(args, " "))
			_ = f.Close()
		}
	}
	switch {
	case len(args) == 3 && args[0] == "plugin" && args[1] == "list" && args[2] == "--json":
		fmt.Println("[{\"id\":\"vloop@vloop\",\"version\":\"0.14.0\",\"scope\":\"user\",\"enabled\":true}]")
	case len(args) == 4 && args[0] == "plugin" && args[1] == "marketplace" && args[2] == "update" && args[3] == "vloop":
	case len(args) == 5 && args[0] == "plugin" && args[1] == "update" && args[2] == "vloop@vloop" && args[3] == "--scope" && args[4] == "user":
	default:
		fmt.Fprintln(os.Stderr, "unexpected fake claude arguments:", strings.Join(args, " "))
		os.Exit(1)
	}
}
`

func TestSelfUpdateWorkedExampleIntegration(t *testing.T) {
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	home := filepath.Join(temp, "home")
	gobin := filepath.Join(home, "go", "bin")
	fake := filepath.Join(temp, "fake")
	work := filepath.Join(temp, "work")
	for _, dir := range []string{gobin, fake, work} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	exeName := "vloop"
	goName := "go"
	claudeName := "claude"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
		goName += ".exe"
		claudeName += ".exe"
	}
	vloop := filepath.Join(gobin, exeName)
	buildSelfUpdateIntegrationProgram(t, realGo, selfUpdateIntegrationMain, vloop, true)
	buildSelfUpdateIntegrationProgram(t, realGo, selfUpdateIntegrationGo, filepath.Join(fake, goName), false)
	buildSelfUpdateIntegrationProgram(t, realGo, selfUpdateIntegrationAgent, filepath.Join(fake, claudeName), false)

	goLog := filepath.Join(temp, "go.log")
	claudeLog := filepath.Join(temp, "claude.log")
	env := append(cleanEnv(home),
		"PATH="+fake,
		"FAKE_GOBIN="+gobin,
		"FAKE_GO_LOG="+goLog,
		"FAKE_CLAUDE_LOG="+claudeLog,
	)

	check := runSelfUpdateIntegration(t, vloop, work, env, "self-update", "--check")
	expect(t, check, 0,
		"current  v0.14.0\n"+
			"latest   v0.15.0 — vloop self-update installs it\n",
		"",
	)

	updated := runSelfUpdateIntegration(t, vloop, work, env, "self-update")
	expect(t, updated, 0,
		"installing vloop v0.15.0 (from v0.14.0)…\n"+
			"installed vloop v0.15.0 at ~/go/bin\n"+
			"plugin claude (user): updated\n"+
			"next: vloop upgrade in each repository that uses vloop, and vloop self-update there for a plugin installed for that repository\n",
		"",
	)

	assertSelfUpdateIntegrationLog(t, goLog, []string{
		"list -m -json github.com/mvelosop/vloop@latest",
		"list -m -json github.com/mvelosop/vloop@latest",
		"env GOBIN",
		"install github.com/mvelosop/vloop/cmd/vloop@v0.15.0",
	})
	assertSelfUpdateIntegrationLog(t, claudeLog, []string{
		"plugin list --json",
		"plugin marketplace update vloop",
		"plugin update vloop@vloop --scope user",
	})
}

func buildSelfUpdateIntegrationProgram(t *testing.T, realGo, source, output string, overlayMain bool) {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	if err := os.WriteFile(src, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"build", "-buildvcs=false"}
	if overlayMain {
		original, err := filepath.Abs("main.go")
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(map[string]any{
			"Replace": map[string]string{original: src},
		})
		if err != nil {
			t.Fatal(err)
		}
		overlay := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlay, raw, 0o644); err != nil {
			t.Fatal(err)
		}
		args = append(args, "-overlay", overlay, "-o", output, ".")
	} else {
		args = append(args, "-o", output, src)
	}
	if raw, err := exec.Command(realGo, args...).CombinedOutput(); err != nil {
		t.Fatalf("build integration program: %v\n%s", err, raw)
	}
}

func runSelfUpdateIntegration(t *testing.T, binary, dir string, env []string, args ...string) result {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatal(err)
		}
		code = exitErr.ExitCode()
	}
	return result{out: stdout.String(), err: stderr.String(), code: code}
}

func assertSelfUpdateIntegrationLog(t *testing.T, path string, want []string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("log %s:\n%s\nwant:\n%s", path, strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
