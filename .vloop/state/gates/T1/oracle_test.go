package gate

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/state"
)

func TestPowerShellGateUTF8Contract(t *testing.T) {
	root := repoRoot(t)
	for _, shell := range gateShells(t) {
		t.Run(shell, func(t *testing.T) {
			t.Run("native output", func(t *testing.T) {
				repo := gitRepo(t)
				verify := `if (-not (git log -1 --format=%s | Select-String -SimpleMatch 'café' -Quiet)) { exit 1 }`
				runGateOK(t, repo, shell, verify)
			})

			t.Run("own output bytes", func(t *testing.T) {
				_, stdout, stderr := runGate(t, root, shell, `Write-Output 'café'`)
				if len(stderr) != 0 {
					t.Fatalf("stderr = %q", stderr)
				}
				wantText := []byte("caf\xc3\xa9")
				if len(stdout) != len(wantText)+2 && len(stdout) != len(wantText)+1 {
					t.Fatalf("stdout bytes = % x", stdout)
				}
				if !bytes.Equal(stdout[:len(wantText)], wantText) {
					t.Fatalf("stdout prefix = % x", stdout)
				}
				ending := stdout[len(wantText):]
				if !bytes.Equal(ending, []byte{'\n'}) && !bytes.Equal(ending, []byte{'\r', '\n'}) {
					t.Fatalf("stdout ending = % x", ending)
				}
			})

			t.Run("three BOM-free encodings", func(t *testing.T) {
				verify := `if ([Console]::OutputEncoding.CodePage -ne 65001 -or [Console]::InputEncoding.CodePage -ne 65001 -or $OutputEncoding.CodePage -ne 65001 -or [Console]::OutputEncoding.GetPreamble().Length -ne 0 -or [Console]::InputEncoding.GetPreamble().Length -ne 0 -or $OutputEncoding.GetPreamble().Length -ne 0) { exit 1 }`
				runGateOK(t, root, shell, verify)
			})

			t.Run("native stdin", func(t *testing.T) {
				ending := []byte{'\n'}
				if runtime.GOOS == "windows" {
					ending = []byte{'\r', '\n'}
				}
				input := append([]byte("caf\xc3\xa9"), ending...)
				want := gitHashObject(t, root, input)
				verify := `if (('café' | git hash-object --stdin) -ne '` + want + `') { exit 1 }`
				runGateOK(t, root, shell, verify)
			})

			t.Run("exit code and silence", func(t *testing.T) {
				code, stdout, stderr := runGate(t, root, shell, `exit 3`)
				if code != 3 || len(stdout) != 0 || len(stderr) != 0 {
					t.Fatalf("code %d stdout %q stderr %q", code, stdout, stderr)
				}
			})
		})
	}
}

func TestUnchangedGateArguments(t *testing.T) {
	want := map[string][]string{
		"sh":   {"-c", "recorded"},
		"bash": {"-c", "recorded"},
		"cmd":  {"/C", "recorded"},
	}
	for shell, args := range want {
		if got := state.GateArgs(shell, "recorded"); !reflect.DeepEqual(got, args) {
			t.Errorf("%s arguments = %q, want %q", shell, got, args)
		}
	}
}

func gateShells(t *testing.T) []string {
	t.Helper()
	shells := []string{"pwsh"}
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Fatalf("pwsh is not on PATH: %v", err)
	}
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("powershell"); err != nil {
			t.Fatalf("powershell is not on PATH: %v", err)
		}
		shells = append(shells, "powershell")
	}
	return shells
}

func runGateOK(t *testing.T, root, shell, verify string) {
	t.Helper()
	code, stdout, stderr := runGate(t, root, shell, verify)
	if code != 0 || len(stdout) != 0 || len(stderr) != 0 {
		t.Fatalf("code %d stdout %q stderr %q", code, stdout, stderr)
	}
}

func runGate(t *testing.T, root, shell, verify string) (int, []byte, []byte) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code, _, err := state.RunGate(root, shell, verify, &stdout, &stderr)
	if err != nil {
		t.Fatalf("RunGate: %v", err)
	}
	return code, stdout.Bytes(), stderr.Bytes()
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
			t.Fatal("repository root not found")
		}
		dir = parent
	}
}

func gitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init", "--quiet")
	runGit(t, root, "config", "user.name", "Gate")
	runGit(t, root, "config", "user.email", "gate@example.invalid")
	if err := os.WriteFile(filepath.Join(root, "input.txt"), []byte("fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "input.txt")
	runGit(t, root, "commit", "--quiet", "-m", "café")
	return root
}

func gitHashObject(t *testing.T, root string, input []byte) string {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "hash-object", "--stdin")
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git hash-object: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"-C", root}, args...)
	out, err := exec.Command("git", cmdArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
