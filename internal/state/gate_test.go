package state

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestGateArgs(t *testing.T) {
	for shell, want := range map[string][]string{
		"sh":   {"-c", "x"},
		"bash": {"-c", "x"},
		"cmd":  {"/C", "x"},
	} {
		if got := GateArgs(shell, "x"); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v want %v", shell, got, want)
		}
	}
}

func TestPowerShellGateArgs(t *testing.T) {
	for _, shell := range []string{"pwsh", "powershell"} {
		got := GateArgs(shell, "recorded")
		if len(got) != 4 {
			t.Fatalf("%s: got %d arguments: %q", shell, len(got), got)
		}
		if want := []string{"-NoProfile", "-NonInteractive", "-Command"}; !reflect.DeepEqual(got[:3], want) {
			t.Errorf("%s: got prefix %q want %q", shell, got[:3], want)
		}
		if !strings.HasSuffix(got[3], "recorded") {
			t.Errorf("%s: command %q does not retain recorded suffix", shell, got[3])
		}
	}
}

func TestPowerShellGateUTF8(t *testing.T) {
	root := gateTestRepoRoot(t)
	for _, shell := range powerShellGateShells(t) {
		t.Run(shell, func(t *testing.T) {
			t.Run("native output", func(t *testing.T) {
				repo := gateTestGitRepo(t)
				verify := `if (-not (git log -1 --format=%s | Select-String -SimpleMatch 'café' -Quiet)) { exit 1 }`
				requireGate(t, repo, shell, verify, 0, nil, nil)
			})

			t.Run("own output bytes", func(t *testing.T) {
				wantEnding := []byte{'\n'}
				if runtime.GOOS == "windows" {
					wantEnding = []byte{'\r', '\n'}
				}
				want := append([]byte("caf\xc3\xa9"), wantEnding...)
				requireGate(t, root, shell, `Write-Output 'café'`, 0, want, nil)
			})

			t.Run("three BOM-free encodings", func(t *testing.T) {
				verify := `if ([Console]::OutputEncoding.CodePage -ne 65001 -or [Console]::InputEncoding.CodePage -ne 65001 -or $OutputEncoding.CodePage -ne 65001 -or [Console]::OutputEncoding.GetPreamble().Length -ne 0 -or [Console]::InputEncoding.GetPreamble().Length -ne 0 -or $OutputEncoding.GetPreamble().Length -ne 0) { exit 1 }`
				requireGate(t, root, shell, verify, 0, nil, nil)
			})

			t.Run("native stdin", func(t *testing.T) {
				ending := []byte{'\n'}
				if runtime.GOOS == "windows" {
					ending = []byte{'\r', '\n'}
				}
				input := append([]byte("caf\xc3\xa9"), ending...)
				want := gateTestGitHashObject(t, root, input)
				verify := `if (('café' | git hash-object --stdin) -ne '` + want + `') { exit 1 }`
				requireGate(t, root, shell, verify, 0, nil, nil)
			})

			t.Run("exit code and setup silence", func(t *testing.T) {
				requireGate(t, root, shell, `exit 3`, 3, nil, nil)
			})
		})
	}
}

func TestGateRefusesUnknownShell(t *testing.T) {
	var out bytes.Buffer
	_, _, err := RunGate(t.TempDir(), "python3", "touch ran", &out, &out)
	var u *UnknownShellError
	if !errors.As(err, &u) {
		t.Fatalf("got %v", err)
	}
	if want := "shell python3 is not one of sh, bash, pwsh, powershell, cmd"; err.Error() != want {
		t.Fatalf("got %q want %q", err, want)
	}
}

func TestGateCmdCommandLine(t *testing.T) {
	if got, want := GateCmdLine(`findstr "a b" f`), `/C findstr "a b" f`; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func powerShellGateShells(t *testing.T) []string {
	t.Helper()
	var shells []string
	for _, shell := range []string{"pwsh", "powershell"} {
		if _, err := exec.LookPath(shell); err == nil {
			shells = append(shells, shell)
		} else if shell == "pwsh" {
			t.Skip("pwsh is not on PATH")
		}
	}
	return shells
}

func requireGate(t *testing.T, root, shell, verify string, wantCode int, wantStdout, wantStderr []byte) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code, _, err := RunGate(root, shell, verify, &stdout, &stderr)
	if err != nil {
		t.Fatalf("RunGate: %v", err)
	}
	if code != wantCode {
		t.Errorf("code = %d, want %d", code, wantCode)
	}
	if !bytes.Equal(stdout.Bytes(), wantStdout) {
		t.Errorf("stdout bytes = % x, want % x", stdout.Bytes(), wantStdout)
	}
	if !bytes.Equal(stderr.Bytes(), wantStderr) {
		t.Errorf("stderr bytes = % x, want % x", stderr.Bytes(), wantStderr)
	}
}

func gateTestRepoRoot(t *testing.T) string {
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
			t.Fatal("repository root not found")
		}
		dir = parent
	}
}

func gateTestGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gateTestRunGit(t, root, "init", "--quiet")
	gateTestRunGit(t, root, "config", "user.name", "Gate")
	gateTestRunGit(t, root, "config", "user.email", "gate@example.invalid")
	if err := os.WriteFile(filepath.Join(root, "input.txt"), []byte("fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gateTestRunGit(t, root, "add", "input.txt")
	gateTestRunGit(t, root, "commit", "--quiet", "-m", "café")
	return root
}

func gateTestGitHashObject(t *testing.T, root string, input []byte) string {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "hash-object", "--stdin")
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git hash-object: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func gateTestRunGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"-C", root}, args...)
	out, err := exec.Command("git", cmdArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
