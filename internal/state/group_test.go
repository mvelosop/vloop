package state

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func skipNoPOSIX(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the gate is a POSIX shell command")
	}
}

// alive reports whether the process in pidfile is still running, giving a
// killed one a moment to be reaped.
func alive(t *testing.T, pidfile string) bool {
	t.Helper()
	b, err := os.ReadFile(pidfile)
	if err != nil {
		t.Fatalf("the gate wrote no pid: %v", err)
	}
	pid := strings.TrimSpace(string(b))
	for i := 0; i < 40; i++ {
		if exec.Command("kill", "-0", pid).Run() != nil {
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
	return true
}

func TestGateLeavesNothingRunning(t *testing.T) {
	skipNoPOSIX(t)
	root := t.TempDir()
	pidfile := filepath.Join(root, "pid")
	var out bytes.Buffer
	start := time.Now()
	code, _, timedOut, err := RunGateWithin(root, "sh", "sleep 1000 & echo $! > pid; echo hi; exit 0", &out, &out, 0)
	if err != nil || code != 0 || timedOut {
		t.Fatalf("code %d timedOut %v err %v", code, timedOut, err)
	}
	if time.Since(start) > 10*time.Second {
		t.Errorf("the gate took %s to return", time.Since(start))
	}
	if out.String() != "hi\n" {
		t.Errorf("output %q", out.String())
	}
	if alive(t, pidfile) {
		t.Error("the background sleep is still running")
	}
}

func TestRunGateWithinTimeout(t *testing.T) {
	skipNoPOSIX(t)
	root := t.TempDir()
	pidfile := filepath.Join(root, "pid")
	var out bytes.Buffer
	start := time.Now()
	code, _, timedOut, err := RunGateWithin(root, "sh", "sleep 1000 & echo $! > pid; echo before; sleep 1000", &out, &out, time.Second)
	if err != nil || !timedOut || code == 0 {
		t.Fatalf("code %d timedOut %v err %v", code, timedOut, err)
	}
	if time.Since(start) > 15*time.Second {
		t.Errorf("the gate took %s to be killed", time.Since(start))
	}
	if !strings.Contains(out.String(), "before") {
		t.Errorf("output %q", out.String())
	}
	if alive(t, pidfile) {
		t.Error("the background sleep survived the timeout")
	}
}

func TestGateTimedOutLine(t *testing.T) {
	if got := GateTimedOutLine("T3", 15); got != "vloop: gate T3 timed out after 15 min" {
		t.Error(got)
	}
}
