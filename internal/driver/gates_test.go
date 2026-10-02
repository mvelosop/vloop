package driver

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mvelosop/vloop/internal/state"
)

func TestGateFilesByToken(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.name", "t")
	run("config", "user.email", "t@example.com")
	names := []string{"a.sh", "data.sh", "scripts/check.ps1", "prüfung.sh", "done.sh"}
	for _, n := range names {
		p := filepath.Join(root, filepath.FromSlash(n))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte("one\n"), 0o644)
	}
	run("add", "-A")
	run("commit", "-qm", "init")
	for _, n := range names {
		os.WriteFile(filepath.Join(root, filepath.FromSlash(n)), []byte("two\n"), 0o644)
	}

	moved := func(verify string, done ...string) map[string]bool {
		it := &Iterator{Root: root, plan: &state.Plan{}}
		for _, v := range done {
			it.plan.Tasks = append(it.plan.Tasks, state.Task{ID: "D", Status: "done", Verify: v})
		}
		got := map[string]bool{}
		for _, f := range it.gateFilesMoved(&state.Task{ID: "T9", Verify: verify}) {
			got[f] = true
		}
		return got
	}

	if got := moved("bash data.sh"); got["a.sh"] || !got["data.sh"] {
		t.Errorf("bash data.sh: %v, want data.sh only", got)
	}
	if got := moved(`pwsh -File .\scripts\check.ps1`); !got["scripts/check.ps1"] {
		t.Errorf("backslash path: %v", got)
	}
	if got := moved(`sh "prüfung.sh"`); !got["prüfung.sh"] {
		t.Errorf("non-ASCII path: %v", got)
	}
	if got := moved("true", "sh ./done.sh && true"); !got["done.sh"] || got["a.sh"] {
		t.Errorf("done task's verify: %v", got)
	}
}
