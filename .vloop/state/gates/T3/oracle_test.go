package gate_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/driver"
	"github.com/mvelosop/vloop/internal/state"
)

func TestPortableGateLint(t *testing.T) {
	root := mixedRoot(t)
	commands := []string{
		"awk", "basename", "bash", "cat", "chmod", "cmd", "cmp", "cp", "cut",
		"date", "diff", "dirname", "env", "fc", "find", "findstr", "grep", "head",
		"ln", "ls", "mkdir", "mktemp", "mv", "powershell", "printf", "readlink",
		"rm", "robocopy", "sed", "seq", "sh", "sleep", "sort", "stat", "tail",
		"tee", "test", "touch", "tr", "uniq", "wc", "which", "xargs", "xcopy",
		"zsh",
	}
	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			got := driver.GateShapeProblems(root, plan(command+" -x value"))
			want := "gate shape rejected: T1  calls " + command + ", which is not on every platform in platforms (macos, windows); use a pwsh cmdlet or the project's own toolchain"
			needProblems(t, got, want)
		})
	}

	rejected := map[string]string{
		"if (Test-Path a) { wc -l a }":             "wc",
		"go test ./x | /usr/bin/GREP.exe ok":       "grep",
		`& 'grep' -n ok a.txt`:                     "grep",
		`& "/usr/bin/GREP.exe" -n ok a.txt`:        "grep",
		"go test ./x; sed -n 1p out":               "sed",
		"go test ./x && Get-Content out | head -1": "head",
		"Get-Content out\nwc -l out":               "wc",
		"{ cp source destination }":                "cp",
		"(cmd /c echo ok)":                         "cmd",
		"Write-Output $(grep value input)":         "grep",
	}
	for command, found := range rejected {
		t.Run(found, func(t *testing.T) {
			got := driver.GateShapeProblems(root, plan(command))
			needContains(t, got, "calls "+found+",")
		})
	}

	accepted := []string{
		"Select-String -Pattern 'grep' -Path a.txt",
		"go test -run TestGrep ./...",
		`node -e "require('fs')"`,
		`Write-Output "go | grep ok"`,
		`if (Test-Path a) { Get-Content a | Measure-Object -Line }`,
		`$command = 'grep'; $value`,
		`foreach ($x in $xs) { Write-Output $x }`,
		`exit 1`,
	}
	for _, command := range accepted {
		t.Run(command, func(t *testing.T) {
			if got := driver.GateShapeProblems(root, plan(command)); len(got) != 0 {
				t.Fatalf("accepted command was rejected: %q", got)
			}
		})
	}

	script := filepath.Join(root, ".vloop", "state", "gates", "T1", "nested", "gate.ps1")
	needWrite(t, script, "$out = go test ./x | grep ok\nwc -l out.txt\n")
	got := driver.GateShapeProblems(root, plan("pwsh -NoProfile -File .vloop/state/gates/T1/nested/gate.ps1"))
	needProblems(t, got,
		"gate shape rejected: T1  .vloop/state/gates/T1/nested/gate.ps1 calls grep, which is not on every platform in platforms (macos, windows); use a pwsh cmdlet or the project's own toolchain")

	unmixed := rootWithPlatforms(t, []string{"windows"})
	if got := driver.GateShapeProblems(unmixed, plan("sh gate.sh | grep ok")); len(got) != 0 {
		t.Fatalf("single-platform plan was linted: %q", got)
	}
	undeclared := rootWithPlatforms(t, nil)
	if got := driver.GateShapeProblems(undeclared, plan("sh gate.sh | grep ok")); len(got) != 0 {
		t.Fatalf("undeclared plan was linted: %q", got)
	}
}

func TestPortableGateLintRejectsB39Plan(t *testing.T) {
	root := repoRoot(t)
	cmd := exec.Command("git", "-C", root, "show", "4c69adf0:.vloop/state/state.json")
	raw, err := cmd.Output()
	if err != nil {
		t.Fatalf("read B39 plan: %v", err)
	}
	var p state.Plan
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("decode B39 plan: %v", err)
	}
	got := driver.GateShapeProblems(root, &p)
	if len(got) != len(p.Tasks) {
		t.Fatalf("B39 problems = %d, tasks = %d:\n%s", len(got), len(p.Tasks), strings.Join(got, "\n"))
	}
	for _, task := range p.Tasks {
		found := false
		for _, problem := range got {
			if strings.Contains(problem, task.ID+"  ") && strings.Contains(problem, "calls sh,") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("B39 task %s has no sh portability refusal", task.ID)
		}
	}
}

func plan(verify string) *state.Plan {
	return &state.Plan{Tasks: []state.Task{{ID: "T1", Verify: verify}}}
}

func mixedRoot(t *testing.T) string {
	t.Helper()
	return rootWithPlatforms(t, []string{"windows", "macos"})
}

func rootWithPlatforms(t *testing.T, platforms []string) string {
	t.Helper()
	root := t.TempDir()
	needWrite(t, filepath.Join(root, ".vloop", "config.toml"), func() string {
		if len(platforms) == 0 {
			return ""
		}
		return "platforms = [\"" + strings.Join(platforms, "\", \"") + "\"]\n"
	}())
	return root
}

func needProblems(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("problems = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("problem %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func needContains(t *testing.T, got []string, part string) {
	t.Helper()
	for _, problem := range got {
		if strings.Contains(problem, part) {
			return
		}
	}
	t.Fatalf("problems %q do not contain %q", got, part)
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
