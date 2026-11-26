package gate_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMixedOSConfigThroughCLI(t *testing.T) {
	root := t.TempDir()
	needMkdir(t, filepath.Join(root, ".vloop"))
	bin := buildVloop(t)

	code, _, errText := vloop(t, bin, root, nil, "config", "set", "platforms", "windows,macos")
	needResult(t, code, "", errText, 0)
	code, out, errText := vloop(t, bin, root, nil, "config", "get", "platforms")
	needOutput(t, code, out, "windows,macos\n", errText)

	code, out, errText = vloop(t, bin, root, nil, "config", "set", "platforms", "macos,beos")
	needResult(t, code, "", errText, 2)
	needExact(t, "invalid platforms", strings.TrimSpace(errText),
		`vloop: invalid value "macos,beos" for platforms: want a comma-separated list of linux, macos, windows`)

	code, _, errText = vloop(t, bin, root, nil, "config", "set", "--local", "platforms", "linux")
	needResult(t, code, "", errText, 2)
	needExact(t, "local platforms", strings.TrimSpace(errText),
		`vloop: .vloop/config.local.toml: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's`)
	if _, err := os.Stat(filepath.Join(root, ".vloop", "config.local.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("refused local platforms write left a file: %v", err)
	}

	code, _, errText = vloop(t, bin, root, map[string]string{"VLOOP_PLATFORMS": "linux"}, "config", "get", "language")
	needResult(t, code, "", errText, 1)
	needExact(t, "environment platforms", strings.TrimSpace(errText),
		`vloop: VLOOP_PLATFORMS: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's`)

	code, out, errText = vloop(t, bin, root, nil, "config", "set", "platforms", "windows,macos,macos")
	needResult(t, code, "", errText, 0)
	_ = out
	code, _, errText = vloop(t, bin, root, nil, "config", "set", "--local", "shell", "pwsh")
	needResult(t, code, "", errText, 2)
	needExact(t, "local shell", strings.TrimSpace(errText),
		`vloop: .vloop/config.local.toml: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell`)
	if _, err := os.Stat(filepath.Join(root, ".vloop", "config.local.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("refused local shell write left a file: %v", err)
	}

	for _, value := range []string{"pwsh", "sh"} {
		code, _, errText = vloop(t, bin, root, map[string]string{"VLOOP_SHELL": value}, "config", "get", "shell")
		needResult(t, code, "", errText, 1)
		needExact(t, "environment shell", strings.TrimSpace(errText),
			`vloop: VLOOP_SHELL: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell`)
	}

	writeFile(t, filepath.Join(root, ".vloop", "config.toml"), "platforms = [\"windows\"]\n")
	code, _, errText = vloop(t, bin, root, nil, "config", "set", "--local", "shell", "sh")
	needResult(t, code, "", errText, 0)
	code, out, errText = vloop(t, bin, root, nil, "config", "get", "shell")
	needOutput(t, code, out, "sh\n", errText)
	code, out, errText = vloop(t, bin, root, map[string]string{"VLOOP_SHELL": "bash"}, "config", "get", "shell")
	needOutput(t, code, out, "bash\n", errText)
}

func buildVloop(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	name := "vloop"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(dir, name)
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/vloop")
	cmd.Dir = repoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build vloop: %v\n%s", err, out)
	}
	return bin
}

func vloop(t *testing.T, bin, root string, extra map[string]string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(bin, append([]string{"-C", root}, args...)...)
	cmd.Env = cleanVloopEnv(extra)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("run vloop: %v", err)
		}
		code = exit.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

func cleanVloopEnv(extra map[string]string) []string {
	env := make([]string, 0, len(os.Environ())+len(extra))
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

func needResult(t *testing.T, code int, stdout, stderr string, wantCode int) {
	t.Helper()
	if code != wantCode || stdout != "" {
		t.Fatalf("exit %d stdout %q stderr %q; want exit %d and empty stdout", code, stdout, stderr, wantCode)
	}
}

func needOutput(t *testing.T, code int, got, want, stderr string) {
	t.Helper()
	if code != 0 || got != want || stderr != "" {
		t.Fatalf("exit %d stdout %q stderr %q; want stdout %q", code, got, stderr, want)
	}
}

func needExact(t *testing.T, label, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q", label, got, want)
	}
}
