//go:build darwin

package driver

import (
	"os/exec"
	"testing"
	"time"
)

func TestCaffeinateExitsWithWatchedProcess(t *testing.T) {
	watched := exec.Command("sleep", "60")
	if err := watched.Start(); err != nil {
		t.Fatal(err)
	}
	argv, err := KeepAwakeCommand("darwin", watched.Process.Pid, nil)
	if err != nil {
		t.Fatal(err)
	}
	caf := exec.Command(argv[0], argv[1:]...)
	if err := caf.Start(); err != nil {
		t.Skipf("caffeinate unavailable: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- caf.Wait() }()
	select {
	case <-done:
		t.Fatal("caffeinate exited while the watched process lived")
	case <-time.After(500 * time.Millisecond):
	}
	watched.Process.Kill()
	watched.Wait()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		caf.Process.Kill()
		t.Fatal("caffeinate outlived the watched process")
	}
}
