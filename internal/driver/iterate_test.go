package driver

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/state"
)

func TestNextReady(t *testing.T) {
	p := &state.Plan{Tasks: []state.Task{
		{ID: "T1", Status: "pending", DependsOn: []string{"T2"}},
		{ID: "T2", Status: "pending"},
		{ID: "T3", Status: "blocked"},
	}}
	if got := nextReady(p); got == nil || got.ID != "T2" {
		t.Fatalf("first ready = %v, want T2", got)
	}
	p.Tasks[1].Status = "done"
	if got := nextReady(p); got == nil || got.ID != "T1" {
		t.Fatalf("after T2 is done, first ready = %v, want T1", got)
	}
	p.Tasks[0].Status = "blocked"
	if got := nextReady(p); got != nil {
		t.Fatalf("nothing is pending, but %v is ready", got.ID)
	}
	done, blocked, pending := counts(p)
	if done != 1 || blocked != 2 || pending != 0 {
		t.Errorf("counts = %d done, %d blocked, %d pending", done, blocked, pending)
	}
}

func TestLazyLogWritesOnFlush(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.log")
	l := &lazyLog{path: path}
	l.Write([]byte("one\n"))
	if _, err := os.Stat(path); err == nil {
		t.Fatal("the log was written before the flush")
	}
	if err := l.flush(); err != nil {
		t.Fatal(err)
	}
	l.Write([]byte("two\n"))
	if err := l.flush(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != "one\ntwo\n" {
		t.Errorf("log = %q", b)
	}
}

func TestRunGateEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the gate is a POSIX shell command")
	}
	root := t.TempDir()
	dir := filepath.Join(root, "run")
	os.MkdirAll(dir, 0o755)
	out := filepath.Join(root, "seen")
	it := &Iterator{Root: root, RunDir: dir,
		Env: []string{"PATH=" + os.Getenv("PATH"), "VLOOP_ACTIVE_TASK=stale", "VLOOP_GATE_TASK=stale"},
		plan: &state.Plan{Shell: "sh", Tasks: []state.Task{
			{ID: "T1", Verify: `printf "%s %s" "$VLOOP_ACTIVE_TASK" "$VLOOP_GATE_TASK" > "` + out + `"; echo noisy; exit 3`},
		}}}
	it.r = &Runner{Root: root, RunDir: dir}
	g, err := it.runGate(4, "T2", "T1")
	if err != nil {
		t.Fatal(err)
	}
	if g.exit != 3 || g.ms < 0 {
		t.Errorf("gate = %+v, want exit 3", g)
	}
	if b, _ := os.ReadFile(out); string(b) != "T2 T1" {
		t.Errorf("the gate saw %q, want T2 T1", b)
	}
	for _, f := range []string{"T1.log", "004-T1.fail.log"} {
		if b, _ := os.ReadFile(filepath.Join(dir, "gates", f)); !strings.Contains(string(b), "noisy") {
			t.Errorf("gates/%s = %q", f, b)
		}
	}
}
