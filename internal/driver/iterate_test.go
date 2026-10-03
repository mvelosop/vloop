package driver

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

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

func timeoutIterator(t *testing.T, verify string) (*Iterator, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the gate is a POSIX shell command")
	}
	root := t.TempDir()
	dir := filepath.Join(root, "run")
	os.MkdirAll(dir, 0o755)
	it := &Iterator{Root: root, RunDir: dir, Budgets: Budgets{GateTimeout: 15},
		GateTimeout: time.Second,
		Env:         []string{"PATH=" + os.Getenv("PATH")},
		plan:        &state.Plan{Shell: "sh", Tasks: []state.Task{{ID: "T1", Verify: verify}}}}
	it.r = &Runner{Root: root, RunDir: dir}
	return it, root
}

func TestGateTimeout(t *testing.T) {
	it, _ := timeoutIterator(t, "echo started; sleep 1000")
	start := time.Now()
	g, err := it.runGate(1, "T1", "T1")
	if err != nil {
		t.Fatal(err)
	}
	if g.exit == 0 {
		t.Error("a timed-out gate passed")
	}
	if time.Since(start) > 15*time.Second {
		t.Errorf("the gate took %s", time.Since(start))
	}
	b, _ := os.ReadFile(filepath.Join(it.RunDir, "gates", "T1.log"))
	if !strings.HasSuffix(string(b), "started\nvloop: gate T1 timed out after 15 min\n") {
		t.Errorf("log %q", b)
	}
}

func TestGateLeavesNothingRunning(t *testing.T) {
	it, root := timeoutIterator(t, "sleep 1000 & echo $! > pid; exit 0")
	it.GateTimeout = 0
	start := time.Now()
	g, err := it.runGate(1, "T1", "T1")
	if err != nil || g.exit != 0 {
		t.Fatalf("gate %+v err %v", g, err)
	}
	if time.Since(start) > 10*time.Second {
		t.Errorf("the gate took %s", time.Since(start))
	}
	b, _ := os.ReadFile(filepath.Join(root, "pid"))
	pid := strings.TrimSpace(string(b))
	for i := 0; ; i++ {
		if exec.Command("kill", "-0", pid).Run() != nil {
			return
		}
		if i == 40 {
			t.Fatal("the background sleep is still running")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestSessionTimeout(t *testing.T) {
	r, _, log := fixture(t, claudeOut)
	stub := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\nsleep 1000 &\nsleep 1000\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r.Claude, r.Timeout = stub, time.Second
	start := time.Now()
	res, err := r.Run(Spec{Phase: PhaseWork, Iteration: 1, Arg: "T1", Model: "sonnet"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.TimedOut || res.ExitCode == 0 {
		t.Errorf("result %+v", res)
	}
	if time.Since(start) > 15*time.Second {
		t.Errorf("the session took %s", time.Since(start))
	}
	if !strings.Contains(log.String(), "SESSION TIMED OUT work T1") {
		t.Errorf("log %q", log.String())
	}
}
