package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/state"
)

func gateRepo(t *testing.T, shell, verify string) string {
	t.Helper()
	root := amendRepo(t, "")
	p, err := state.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	p.Shell, p.Tasks[1].Verify = shell, verify
	if err := state.Save(root, p); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestTaskGateShPassFail(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is not on PATH")
	}
	root := gateRepo(t, "sh", "echo hello; exit 3")
	before := planBytes(t, root)
	code, out, _ := run(t, "-C", root, "task", "gate", "T2")
	if code != 1 || !regexp.MustCompile(`^hello\ngate T2: fail \(exit 3, [^)]+\)\n$`).MatchString(out) {
		t.Fatalf("code %d out %q", code, out)
	}
	if planBytes(t, root) != before {
		t.Fatal("task gate wrote the plan")
	}
	setVerify(t, root, "pwd >pwd.txt")
	code, out, _ = run(t, "-C", root, "task", "gate", "T2")
	if code != 0 || !regexp.MustCompile(`^gate T2: pass \([^)]+\)\n$`).MatchString(out) {
		t.Fatalf("code %d out %q", code, out)
	}
	got, _ := os.ReadFile(filepath.Join(root, "pwd.txt"))
	want, _ := filepath.EvalSymlinks(root)
	have, _ := filepath.EvalSymlinks(strings.TrimSpace(string(got)))
	if have != want {
		t.Fatalf("gate ran in %q, want repo root %q", have, want)
	}
}

func setVerify(t *testing.T, root, verify string) {
	t.Helper()
	p, _ := state.Load(root)
	p.Tasks[1].Verify = verify
	if err := state.Save(root, p); err != nil {
		t.Fatal(err)
	}
}

func TestTaskGateJSON(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is not on PATH")
	}
	root := gateRepo(t, "sh", "echo noisy; exit 2")
	code, out, errOut := run(t, "-C", root, "task", "gate", "T2", "--json")
	if code != 1 || errOut != "noisy\n" {
		t.Fatalf("code %d err %q", code, errOut)
	}
	if !regexp.MustCompile(`^\{"task":"T2","passed":false,"exit":2,"duration_ms":\d+\}\n$`).MatchString(out) {
		t.Fatalf("out %q", out)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatal(err)
	}
}

func TestTaskGateShellMissing(t *testing.T) {
	root := gateRepo(t, "pwsh", "true")
	before := planBytes(t, root)
	t.Setenv("PATH", t.TempDir())
	code, out, errOut := run(t, "-C", root, "task", "gate", "T2")
	if code != 1 || out != "" || errOut != "vloop: this plan's gates are pwsh commands and pwsh is not on PATH\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	if planBytes(t, root) != before {
		t.Fatal("task gate wrote the plan")
	}
}

// TestTaskGateShellInvocation uses a PATH test double for each shell and
// checks the arguments it was given.
func TestTaskGateShellInvocation(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is not on PATH, so the shell test doubles cannot run")
	}
	for shell, want := range map[string]string{
		"bash":       "-c|CMD",
		"pwsh":       "-NoProfile|-NonInteractive|-Command|CMD",
		"powershell": "-NoProfile|-NonInteractive|-Command|CMD",
		"cmd":        "/C|CMD",
	} {
		bin, log := t.TempDir(), filepath.Join(t.TempDir(), "args")
		script := "#!/bin/sh\nfor a in \"$@\"; do printf '%s|' \"$a\"; done >" + log + "\n"
		if err := os.WriteFile(filepath.Join(bin, shell), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		root := gateRepo(t, shell, "CMD")
		t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
		if code, out, errOut := run(t, "-C", root, "task", "gate", "T2"); code != 0 {
			t.Fatalf("%s: code %d out %q err %q", shell, code, out, errOut)
		}
		got, _ := os.ReadFile(log)
		if string(got) != want+"|" {
			t.Errorf("%s: got %q, want %q", shell, got, want+"|")
		}
	}
}

func TestTaskGateUnknownTask(t *testing.T) {
	root := gateRepo(t, "sh", "true")
	if code, out, errOut := run(t, "-C", root, "task", "gate", "T9"); code != 1 || out != "" || errOut != "vloop: no task T9\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestTaskVerifyReplaces(t *testing.T) {
	root := amendRepo(t, "")
	p, _ := state.Load(root)
	old := p.Tasks[1].Verify
	if code, out, errOut := run(t, "-C", root, "task", "verify", "T2", "test -f x", "--reason", "wrong file"); code != 0 || out != "" || errOut != "" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	got, err := state.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	h := got.Tasks[1].GateHistory
	if got.Tasks[1].Verify != "test -f x" || len(h) != 1 || h[0].Verify != old || h[0].Reason != "wrong file" || h[0].By != "operator" ||
		!regexp.MustCompile(`^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$`).MatchString(h[0].ReplacedAt) {
		t.Fatalf("got %+v", got.Tasks[1])
	}
	// Nothing else changed: put the old plan's task back and compare.
	got.Tasks[1].Verify, got.Tasks[1].GateHistory = old, nil
	p.Updated = got.Updated
	a, _ := state.Marshal(p)
	b, _ := state.Marshal(got)
	if string(a) != string(b) {
		t.Fatal("more than the verify command and gate_history changed")
	}
}

func TestTaskVerifyRefusals(t *testing.T) {
	root := amendRepo(t, "")
	refuse(t, root, 2, `vloop: required flag(s) "reason" not set`, "task", "verify", "T2", "true")
	p, _ := state.Load(root)
	refuse(t, root, 1, "vloop: T2 already has that verify command\n", "task", "verify", "T2", p.Tasks[1].Verify, "--reason", "r")
	refuse(t, root, 1, "vloop: no task T9\n", "task", "verify", "T9", "x", "--reason", "r")
	if code, out, _ := run(t, "-C", root, "task", "verify", "T2", "", "--reason", "r"); code == 0 || out != "" {
		t.Fatalf("empty command: code %d out %q", code, out)
	}
	if planBytes(t, root) != mustMarshal(t, p) {
		t.Fatal("the plan was modified")
	}
}

func mustMarshal(t *testing.T, p *state.Plan) string {
	t.Helper()
	b, err := state.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
